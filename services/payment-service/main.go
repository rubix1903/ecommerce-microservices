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

	paymentpb "github.com/ecommerce/microservices/proto/payment"
	"github.com/ecommerce/microservices/shared/codec"
	"github.com/ecommerce/microservices/shared/config"
	"github.com/ecommerce/microservices/shared/events"
	appkafka "github.com/ecommerce/microservices/shared/kafka"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Payment struct {
	ID          string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrderID     string  `gorm:"uniqueIndex;not null"`
	UserID      string  `gorm:"not null;index"`
	Amount      float64 `gorm:"not null"`
	Status      string  `gorm:"not null"`
	Gateway     string  `gorm:"default:'mock'"`
	ProcessedAt time.Time
	CreatedAt   time.Time
}

type PaymentHandler struct {
	paymentpb.UnimplementedPaymentServiceServer
	db            *gorm.DB
	kafkaProducer *appkafka.Producer
}

func NewPaymentHandler(db *gorm.DB, kp *appkafka.Producer) *PaymentHandler {
	return &PaymentHandler{db: db, kafkaProducer: kp}
}

func (h *PaymentHandler) ProcessPayment(ctx context.Context, req *paymentpb.ProcessPaymentRequest) (*paymentpb.ProcessPaymentResponse, error) {
	if req.OrderID == "" || req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "order_id and amount > 0 required")
	}
	return h.doProcess(ctx, req.OrderID, req.UserID, req.Amount)
}

func (h *PaymentHandler) GetPayment(_ context.Context, req *paymentpb.GetPaymentRequest) (*paymentpb.Payment, error) {
	var p Payment
	if err := h.db.Where("id = ?", req.PaymentID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "payment not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}
	return &paymentpb.Payment{ID: p.ID, OrderID: p.OrderID, UserID: p.UserID, Amount: p.Amount, Status: p.Status, Gateway: p.Gateway, ProcessedAt: p.ProcessedAt}, nil
}

func (h *PaymentHandler) doProcess(ctx context.Context, orderID, userID string, amount float64) (*paymentpb.ProcessPaymentResponse, error) {
	var existing Payment
	if err := h.db.Where("order_id = ?", orderID).First(&existing).Error; err == nil {
		return &paymentpb.ProcessPaymentResponse{PaymentID: existing.ID, Status: existing.Status, Amount: existing.Amount, Message: "already processed"}, nil
	}
	payStatus := "success"
	if rand.Float32() < 0.05 {
		payStatus = "failed"
	}
	p := &Payment{OrderID: orderID, UserID: userID, Amount: amount, Status: payStatus, Gateway: "mock", ProcessedAt: time.Now()}
	if err := h.db.Create(p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to persist payment: %v", err)
	}
	if payStatus == "success" {
		h.kafkaProducer.Publish(ctx, events.TopicPaymentProcessed, p.ID, events.PaymentProcessedEvent{
			PaymentID: p.ID, OrderID: orderID, UserID: userID, Amount: amount, Status: "success", ProcessedAt: p.ProcessedAt,
		})
	} else {
		h.kafkaProducer.Publish(ctx, events.TopicPaymentFailed, p.ID, events.PaymentFailedEvent{
			PaymentID: p.ID, OrderID: orderID, UserID: userID, Reason: "card declined (simulated)", FailedAt: p.ProcessedAt,
		})
	}
	log.Printf("💳 payment %s → %s (order: %s)", p.ID, payStatus, orderID)
	return &paymentpb.ProcessPaymentResponse{PaymentID: p.ID, Status: payStatus, Amount: amount, Message: "payment " + payStatus}, nil
}

func startKafkaConsumer(ctx context.Context, h *PaymentHandler, cfg *config.Config) {
	c := appkafka.NewConsumer(cfg.KafkaBrokers, events.TopicOrderCreated, "payment-service-group")
	defer c.Close()
	c.Consume(ctx, func(ctx context.Context, key string, payload []byte) error {
		var evt events.OrderCreatedEvent
		if err := appkafka.Unmarshal(payload, &evt); err != nil {
			return nil
		}
		log.Printf("📨 payment-service: processing order %s", evt.OrderID)
		_, err := h.doProcess(ctx, evt.OrderID, evt.UserID, evt.Amount)
		return err
	})
}

func main() {
	codec.Register() // must be first

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
	ctx, cancel := context.WithCancel(context.Background())
	go startKafkaConsumer(ctx, handler, cfg)

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

	log.Printf("🚀 payment-service: gRPC on :%s + Kafka consumer", port)
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
