package events

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOrderCreatedEvent_RoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Second) // truncate to avoid sub-second diff in JSON
	orig := OrderCreatedEvent{
		OrderID:   "order-001",
		UserID:    "user-abc",
		ProductID: "prod-xyz",
		Quantity:  3,
		Amount:    149.97,
		CreatedAt: now,
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded OrderCreatedEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.OrderID != orig.OrderID {
		t.Errorf("OrderID: got %q, want %q", decoded.OrderID, orig.OrderID)
	}
	if decoded.Quantity != orig.Quantity {
		t.Errorf("Quantity: got %d, want %d", decoded.Quantity, orig.Quantity)
	}
	if decoded.Amount != orig.Amount {
		t.Errorf("Amount: got %f, want %f", decoded.Amount, orig.Amount)
	}
}

func TestPaymentProcessedEvent_RoundTrip(t *testing.T) {
	orig := PaymentProcessedEvent{
		PaymentID:   "pay-001",
		OrderID:     "order-001",
		UserID:      "user-abc",
		Amount:      299.99,
		Status:      "success",
		ProcessedAt: time.Now().Truncate(time.Second),
	}

	data, _ := json.Marshal(orig)
	var decoded PaymentProcessedEvent
	json.Unmarshal(data, &decoded)

	if decoded.Status != "success" {
		t.Errorf("Status: got %q, want success", decoded.Status)
	}
	if decoded.Amount != orig.Amount {
		t.Errorf("Amount mismatch")
	}
}

func TestPaymentFailedEvent_RoundTrip(t *testing.T) {
	orig := PaymentFailedEvent{
		PaymentID: "pay-002",
		OrderID:   "order-002",
		UserID:    "user-def",
		Reason:    "card declined",
		FailedAt:  time.Now().Truncate(time.Second),
	}

	data, _ := json.Marshal(orig)
	var decoded PaymentFailedEvent
	json.Unmarshal(data, &decoded)

	if decoded.Reason != "card declined" {
		t.Errorf("Reason: got %q, want %q", decoded.Reason, "card declined")
	}
}

func TestTopicConstants(t *testing.T) {
	topics := []string{TopicOrderCreated, TopicPaymentProcessed, TopicPaymentFailed}
	for _, topic := range topics {
		if topic == "" {
			t.Error("topic constant must not be empty")
		}
	}
	// Ensure unique topic names
	seen := make(map[string]bool)
	for _, topic := range topics {
		if seen[topic] {
			t.Errorf("duplicate topic: %q", topic)
		}
		seen[topic] = true
	}
}
