package ledger

import (
	"atomicbank/events"
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

func TestTransferFunds_ConcurrentDoubleSpend(t *testing.T) {
	// 1. Connect to the local Docker database
	connStr := "postgresql://admin:secret@localhost:5432/atomicbank?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to db: %v", err)
	}
	defer db.Close()

	// 2. Setup dummy accounts and an Event Bus
	bus := events.NewEventBus()

	// We will use our seeded accounts for this test
	fromAccountID, _ := uuid.Parse("11111111-1111-1111-1111-111111111111")
	toAccountID, _ := uuid.Parse("22222222-2222-2222-2222-222222222222")
	amountToSteal := decimal.NewFromFloat(50.00)

	// 3. Inject exactly $50 for this test via a manual ledger entry
	_, err = db.Exec(`
		INSERT INTO ledger_entries (id, transaction_id, account_id, amount, direction)
		VALUES ($1, $2, $3, $4, 'CREDIT')`,
		uuid.New(), uuid.New(), fromAccountID, amountToSteal)
	if err != nil {
		t.Fatalf("Failed to seed test money: %v", err)
	}

	// 4. The Attack: Fire 100 concurrent transfer requests
	concurrentRequests := 100
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	successCount := 0
	failCount := 0
	var mu sync.Mutex // To safely count results across goroutines

	t.Logf("Firing %d concurrent requests to steal the same $50...", concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			defer wg.Done()

			// Context with a short timeout to prevent hanging tests
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := TransferFunds(ctx, db, bus, fromAccountID, toAccountID, amountToSteal)

			mu.Lock()
			if err == nil {
				successCount++
			} else {
				failCount++
			}
			mu.Unlock()
		}()
	}

	// 5. Wait for all 100 requests to finish fighting
	wg.Wait()

	// 6. The Proof: Assert that ONLY ONE transaction succeeded
	t.Logf("Successes: %d | Failures: %d", successCount, failCount)

	if successCount != 1 {
		t.Errorf("CRITICAL FAILURE: Expected exactly 1 success, got %d. Double spend occurred!", successCount)
	}
	if failCount != concurrentRequests-1 {
		t.Errorf("Expected %d failures, got %d", concurrentRequests-1, failCount)
	}
}
