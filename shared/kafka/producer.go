// Package kafka provides simple Kafka producer and consumer helpers.
package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer wraps a kafka.Writer for structured event publishing.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer connected to the given brokers.
func NewProducer(brokers []string) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		// Allow auto-creation of topics in development
		AllowAutoTopicCreation: true,
	}
	return &Producer{writer: w}
}

// Publish serialises msg as JSON and sends it to the given topic.
func (p *Producer) Publish(ctx context.Context, topic string, key string, msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	})
}

// Close flushes and closes the writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

// EnsureTopics pre-creates Kafka topics (useful in development).
func EnsureTopics(brokers []string, topics []string) {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		log.Printf("⚠  kafka: could not dial broker to create topics: %v", err)
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("⚠  kafka: could not get controller: %v", err)
		return
	}

	controllerConn, err := kafka.Dial("tcp", controller.Host+":"+string(rune('0'+controller.Port/1000%10))+"...")
	if err != nil {
		// Best-effort; topics are also auto-created on first publish
		log.Printf("ℹ  kafka: topics will be auto-created on first message")
		return
	}
	defer controllerConn.Close()

	topicConfigs := make([]kafka.TopicConfig, len(topics))
	for i, t := range topics {
		topicConfigs[i] = kafka.TopicConfig{
			Topic:             t,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
	}
	if err := controllerConn.CreateTopics(topicConfigs...); err != nil {
		log.Printf("ℹ  kafka topic creation: %v (may already exist)", err)
	}
}
