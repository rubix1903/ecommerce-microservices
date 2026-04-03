// Package events defines the Kafka event payloads shared across all services.
package events

import "time"

// Topic names - centralised to avoid typos across services
const (
	TopicOrderCreated      = "order.created"
	TopicPaymentProcessed  = "payment.processed"
	TopicPaymentFailed     = "payment.failed"
)

// OrderCreatedEvent is published to Kafka when a new order is placed.
type OrderCreatedEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	ProductID string    `json:"product_id"`
	Quantity  int32     `json:"quantity"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}

// PaymentProcessedEvent is published when a payment succeeds.
type PaymentProcessedEvent struct {
	PaymentID   string    `json:"payment_id"`
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

// PaymentFailedEvent is published when a payment fails.
type PaymentFailedEvent struct {
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}
