package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ecommerce/microservices/shared/codec"
	"github.com/ecommerce/microservices/shared/config"
	"github.com/ecommerce/microservices/shared/events"
	appkafka "github.com/ecommerce/microservices/shared/kafka"
	paymentpb "github.com/ecommerce/microservices/proto/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ─── Model ───────────────────────────────────────────────────────────────────

type Payment struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID     string    `gorm:"uniqueIndex;not null"`
	UserID      string    `gorm:"not null;index"`
	Amount      float64   `gorm:"not null"`
	Status      string    `gorm:"not null"` // success|failed|pending
	Gateway     string    `gorm:"default:'mock'"`
	ProcessedAt time.Time
	CreatedAt   time.Time
}

// ─── gRPC Handler ─────────────────────────────────────────────────────────────

type PaymentHandler struct {
	paymentpb.UnimplementedPaymentServiceServer
	db            *gorm.DB
	kafkaProducer *appkafka.Producer
}

func NewPaymentHandler(db *gorm.DB, kp *appkafka.Producer) *PaymentHandler {
	return &PaymentHandler{db: db, kafkaProducer: kp}
}

// ProcessPayment handles a direct gRPC payment call (e.g. from api-gateway).
func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	if req.OrderID == "" || req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id and amount > 0 required")
	}
	return h.doProcessPayment(ctx, req.OrderID, req.UserID, req.Amount)
}

func (h *PaymentHandler) GetPayment(_ context.Context, req *paymentpb.GetPaymentRequest) (*paymentpb.Payment, error) {
	var p Payment
	if err := h.db.Where("id = ?", req.PaymentID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "payment not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}
	return toProtoPayment(&p), nil
}

// doProcessPayment contains the shared payment logic used by both gRPC and Kafka paths.
func (h *PaymentHandler) doProcessPayment(ctx context.Context, orderID, userID string, amount float64) (*paymentpb.ProcessPaymentResponse, error) {
	// Idempotency: return existing payment if already processed
	var existing Payment
	if err := h.db.Where("order_id = ?", orderID).First(&existing).Error; err == nil {
		return &paymentpb.ProcessPaymentResponse{
			PaymentID: existing.ID, Status: existing.Status,
			Message: "payment already processed", Amount: existing.Amount,
		}, nil
	}

	// Simulate a payment gateway call (95% success rate)
	paymentStatus := "success"
	if rand.Float32() < 0.05 {
		paymentStatus = "failed"
	}

	payment := &Payment{
		OrderID:     orderID,
		UserID:      userID,
		Amount:      amount,
		Status:      paymentStatus,
		Gateway:     "mock-gateway",
		ProcessedAt: time.Now(),
	}
	if err := h.db.Create(payment).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to persist payment: %v", err)
	}

	// Publish result to Kafka (notification-service and order-service consume this)
	if paymentStatus == "success" {
		evt := events.PaymentProcessedEvent{
			PaymentID: payment.ID, OrderID: orderID, UserID: userID,
			Amount: amount, Status: "success", ProcessedAt: payment.ProcessedAt,
		}
		if err := h.kafkaProducer.Publish(ctx, events.TopicPaymentProcessed, payment.ID, evt); err != nil {
			log.Printf("⚠  payment-service: failed to publish payment.processed: %v", err)
		}
	} else {
		evt := events.PaymentFailedEvent{
			PaymentID: payment.ID, OrderID: orderID, UserID: userID,
			Reason: "card declined (simulated)", FailedAt: payment.ProcessedAt,
		}
		if err := h.kafkaProducer.Publish(ctx, events.TopicPaymentFailed, payment.ID, evt); err != nil {
			log.Printf("⚠  payment-service: failed to publish payment.failed: %v", err)
		}
	}

	log.Printf("💳 payment %s → %s (order: %s, amount: %.2f)", payment.ID, paymentStatus, orderID, amount)
	return &paymentpb.ProcessPaymentResponse{
		PaymentID: payment.ID, Status: paymentStatus,
		Message: "payment " + paymentStatus, Amount: amount,
	}, nil
}

// ─── Kafka Consumer ───────────────────────────────────────────────────────────

// startKafkaConsumer listens for order.created events and triggers payment processing.
func startKafkaConsumer(ctx context.Context, handler *PaymentHandler, cfg *config.Config) {
	consumer := appkafka.NewConsumer(cfg.KafkaBrokers, events.TopicOrderCreated, "payment-service-group")
	defer consumer.Close()

	consumer.Consume(ctx, func(ctx context.Context, key string, payload []byte) error {
		var evt events.OrderCreatedEvent
		if err := appkafka.Unmarshal(payload, &evt); err != nil {
			log.Printf("⚠  payment-service: failed to unmarshal order.created: %v", err)
			return nil // don't retry bad messages
		}
		log.Printf("📨 payment-service received order.created for order %s", evt.OrderID)
		_, err := handler.doProcessPayment(ctx, evt.OrderID, evt.UserID, evt.Amount)
		return err
	})
}

func toProtoPayment(p *Payment) *paymentpb.Payment {
	return &paymentpb.Payment{
		ID: p.ID, OrderID: p.OrderID, UserID: p.UserID,
		Amount: p.Amount, Status: p.Status, Gateway: p.Gateway, ProcessedAt: p.ProcessedAt,
	}
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ payment-service: DB connect failed: %v", err)
	}
	db.AutoMigrate(&Payment{})
	log.Println("✅ payment-service: database connected")

	producer := appkafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	handler := NewPaymentHandler(db, producer)

	// Start Kafka consumer in background
	ctx, cancel := context.WithCancel(context.Background())
	go startKafkaConsumer(ctx, handler, cfg)

	// Also expose a gRPC endpoint for direct payment calls
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50054"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ payment-service: listen failed: %v", err)
	}

	srv := grpc.NewServer()
	paymentpb.RegisterPaymentServiceServer(srv, handler)
	reflection.Register(srv)

	log.Printf("🚀 payment-service: gRPC on :%s + Kafka consumer on %s", port, events.TopicOrderCreated)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("❌ payment-service: serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 payment-service: shutting down...")
	cancel()
	srv.GracefulStop()
}
