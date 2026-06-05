package ledger

import (
	"atomicbank/events"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrInsufficientFunds = errors.New("insufficient funds")

func TransferFunds(ctx context.Context, db *sql.DB, bus *events.EventBus, fromAccountID, toAccountID uuid.UUID, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("transfer amount must be strictly positive")
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var currentBalance decimal.Decimal
	err = tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount), 0) 
		FROM ledger_entries 
		WHERE account_id = $1 FOR UPDATE`, fromAccountID).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to read balance: %w", err)
	}

	if currentBalance.LessThan(amount) {
		return ErrInsufficientFunds
	}

	transactionID := uuid.New()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ledger_entries (id, transaction_id, account_id, amount, direction)
		VALUES ($1, $2, $3, $4, 'DEBIT')`,
		uuid.New(), transactionID, fromAccountID, amount.Neg())
	if err != nil {
		return fmt.Errorf("failed to create debit entry: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ledger_entries (id, transaction_id, account_id, amount, direction)
		VALUES ($1, $2, $3, $4, 'CREDIT')`,
		uuid.New(), transactionID, toAccountID, amount)
	if err != nil {
		return fmt.Errorf("failed to create credit entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Publish to asynchronous event system
	bus.TransferStream <- events.TransferEvent{
		TransactionID: transactionID,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
	}

	return nil
}
