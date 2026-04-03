package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecommerce/microservices/shared/config"
	"github.com/ecommerce/microservices/shared/events"
	appkafka "github.com/ecommerce/microservices/shared/kafka"
)

// ─── Notifier ────────────────────────────────────────────────────────────────

// Notifier simulates sending email/SMS. In production, plug in SendGrid, Twilio, etc.
type Notifier struct{}

func (n *Notifier) SendPaymentSuccess(evt events.PaymentProcessedEvent) {
	msg := fmt.Sprintf(
		"📧 [EMAIL] To: user-%s | Subject: Payment Confirmed!\n"+
			"   Your order %s has been paid (%.2f). Payment ID: %s",
		evt.UserID, evt.OrderID, evt.Amount, evt.PaymentID,
	)
	log.Println(msg)
}

func (n *Notifier) SendPaymentFailed(evt events.PaymentFailedEvent) {
	msg := fmt.Sprintf(
		"📧 [EMAIL] To: user-%s | Subject: Payment Failed\n"+
			"   Your order %s payment failed. Reason: %s. Please retry.",
		evt.UserID, evt.OrderID, evt.Reason,
	)
	log.Println(msg)
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	cfg := config.Load()
	notifier := &Notifier{}

	ctx, cancel := context.WithCancel(context.Background())

	// Consumer for payment.processed
	go func() {
		c := appkafka.NewConsumer(cfg.KafkaBrokers, events.TopicPaymentProcessed, "notification-service-paid")
		defer c.Close()
		c.Consume(ctx, func(ctx context.Context, key string, payload []byte) error {
			var evt events.PaymentProcessedEvent
			if err := appkafka.Unmarshal(payload, &evt); err != nil {
				log.Printf("⚠  notification-service: bad payment.processed payload: %v", err)
				return nil
			}
			notifier.SendPaymentSuccess(evt)
			return nil
		})
	}()

	// Consumer for payment.failed
	go func() {
		c := appkafka.NewConsumer(cfg.KafkaBrokers, events.TopicPaymentFailed, "notification-service-failed")
		defer c.Close()
		c.Consume(ctx, func(ctx context.Context, key string, payload []byte) error {
			var evt events.PaymentFailedEvent
			if err := appkafka.Unmarshal(payload, &evt); err != nil {
				log.Printf("⚠  notification-service: bad payment.failed payload: %v", err)
				return nil
			}
			notifier.SendPaymentFailed(evt)
			return nil
		})
	}()

	log.Printf("🚀 notification-service: listening on topics [%s, %s]",
		events.TopicPaymentProcessed, events.TopicPaymentFailed)

	// Health check: print a heartbeat every 30s so you can see it's alive
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
				log.Printf("💓 notification-service: heartbeat @ %s", time.Now().Format(time.RFC3339))
			}
		}
	}()

	hostname, _ := os.Hostname()
	log.Printf("ℹ  notification-service running on host: %s", hostname)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 notification-service: shutting down...")
	cancel()
}
