package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// HandlerFunc is called for each Kafka message received.
type HandlerFunc func(ctx context.Context, key string, payload []byte) error

// Consumer wraps a kafka.Reader for structured event consumption.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer creates a Kafka consumer subscribed to a single topic.
func NewConsumer(brokers []string, topic, groupID string) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 1,
		MaxBytes: 10e6, // 10 MB
	})
	return &Consumer{reader: r}
}

// Consume starts consuming messages and calls handler for each one.
// It blocks until ctx is cancelled.
func (c *Consumer) Consume(ctx context.Context, handler HandlerFunc) {
	log.Printf("📥 Kafka consumer listening on topic: %s", c.reader.Config().Topic)
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return // graceful shutdown
			}
			log.Printf("⚠  kafka fetch error: %v", err)
			continue
		}

		if err := handler(ctx, string(msg.Key), msg.Value); err != nil {
			log.Printf("⚠  handler error for topic %s: %v", msg.Topic, err)
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("⚠  kafka commit error: %v", err)
		}
	}
}

// Unmarshal is a helper to decode a raw Kafka payload into v.
func Unmarshal(payload []byte, v interface{}) error {
	return json.Unmarshal(payload, v)
}

// Close closes the reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}
