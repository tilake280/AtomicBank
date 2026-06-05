package events

import (
	"log"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// TransferEvent represents the payload sent to the event bus
type TransferEvent struct {
	TransactionID uuid.UUID
	FromAccountID uuid.UUID
	ToAccountID   uuid.UUID
	Amount        decimal.Decimal
}

// EventBus holds our Go channel (simulating a Kafka topic)
type EventBus struct {
	TransferStream chan TransferEvent
}

// NewEventBus initializes the channel with a buffer so it doesn't block
func NewEventBus() *EventBus {
	return &EventBus{
		TransferStream: make(chan TransferEvent, 100),
	}
}

// StartNotificationWorker runs in the background listening for events
func (b *EventBus) StartNotificationWorker() {
	go func() {
		for event := range b.TransferStream {
			log.Printf("🔔 [PUSH NOTIFICATION] $%.2f successfully transferred from %s to %s (Txn: %s)\n",
				event.Amount.InexactFloat64(),
				event.FromAccountID.String()[:8],
				event.ToAccountID.String()[:8],
				event.TransactionID.String()[:8],
			)
		}
	}()
}
