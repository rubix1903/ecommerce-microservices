package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/ecommerce/microservices/shared/codec"
	"github.com/ecommerce/microservices/shared/config"
	"github.com/ecommerce/microservices/shared/events"
	appkafka "github.com/ecommerce/microservices/shared/kafka"
	orderpb "github.com/ecommerce/microservices/proto/order"
	productpb "github.com/ecommerce/microservices/proto/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ─── Model ───────────────────────────────────────────────────────────────────

type Order struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string    `gorm:"not null;index"`
	ProductID string    `gorm:"not null"`
	Quantity  int32     `gorm:"not null"`
	Amount    float64   `gorm:"not null"`
	Status    string    `gorm:"not null;default:'pending'"` // pending|paid|cancelled
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ─── Handler ─────────────────────────────────────────────────────────────────

type OrderHandler struct {
	orderpb.UnimplementedOrderServiceServer
	db             *gorm.DB
	productClient  productpb.ProductServiceClient
	kafkaProducer  *appkafka.Producer
}

func NewOrderHandler(db *gorm.DB, pc productpb.ProductServiceClient, kp *appkafka.Producer) *OrderHandler {
	return &OrderHandler{db: db, productClient: pc, kafkaProducer: kp}
}

// CreateOrder is the core business logic:
// 1. Fetch product & price from product-service via gRPC
// 2. Deduct stock atomically
// 3. Persist order in DB
// 4. Publish order.created to Kafka → triggers async payment processing
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	if req.UserID == "" || req.ProductID == "" || req.Quantity <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, product_id, and quantity > 0 are required")
	}

	// Step 1 – verify product exists and get price
	product, err := h.productClient.GetProduct(ctx, &productpb.GetProductRequest{ID: req.ProductID})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "product not found: %v", err)
	}

	// Step 2 – deduct stock (transactional in product-service)
	deduct, err := h.productClient.DeductStock(ctx, &productpb.DeductStockRequest{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		return nil, err // already a gRPC status error
	}
	if !deduct.Success {
		return nil, status.Error(codes.FailedPrecondition, "stock deduction failed")
	}

	// Step 3 – persist the order
	totalAmount := product.Price * float64(req.Quantity)
	order := &Order{
		UserID:    req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Amount:    totalAmount,
		Status:    "pending",
	}
	if err := h.db.Create(order).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to persist order: %v", err)
	}

	// Step 4 – publish OrderCreated event to Kafka (async; payment-service consumes)
	event := events.OrderCreatedEvent{
		OrderID:   order.ID,
		UserID:    order.UserID,
		ProductID: order.ProductID,
		Quantity:  order.Quantity,
		Amount:    order.Amount,
		CreatedAt: order.CreatedAt,
	}
	if err := h.kafkaProducer.Publish(ctx, events.TopicOrderCreated, order.ID, event); err != nil {
		// Non-fatal: log and continue. In production, use an outbox pattern.
		log.Printf("⚠  order-service: failed to publish order.created for %s: %v", order.ID, err)
	}

	log.Printf("✅ order created: %s (%.2f)", order.ID, order.Amount)
	return &orderpb.CreateOrderResponse{
		OrderID: order.ID,
		Amount:  order.Amount,
		Status:  order.Status,
		Message: "order created – payment processing initiated",
	}, nil
}

func (h *OrderHandler) GetOrder(_ context.Context, req *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	var o Order
	if err := h.db.Where("id = ? AND user_id = ?", req.OrderID, req.UserID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "order not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}
	return toProtoOrder(&o), nil
}

func (h *OrderHandler) ListOrders(_ context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	page, limit := req.Page, req.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var orders []Order
	var total int64
	h.db.Model(&Order{}).Where("user_id = ?", req.UserID).Count(&total)
	if err := h.db.Where("user_id = ?", req.UserID).
		Offset(int((page-1)*limit)).Limit(int(limit)).
		Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to list orders")
	}

	out := make([]*orderpb.Order, len(orders))
	for i, o := range orders {
		out[i] = toProtoOrder(&o)
	}
	return &orderpb.ListOrdersResponse{Orders: out, Total: total}, nil
}

func (h *OrderHandler) UpdateOrderStatus(_ context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	if err := h.db.Model(&Order{}).Where("id = ?", req.OrderID).Update("status", req.Status).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update order: %v", err)
	}
	return &orderpb.UpdateOrderStatusResponse{Success: true}, nil
}

func toProtoOrder(o *Order) *orderpb.Order {
	return &orderpb.Order{
		ID: o.ID, UserID: o.UserID, ProductID: o.ProductID,
		Quantity: o.Quantity, Amount: o.Amount, Status: o.Status, CreatedAt: o.CreatedAt,
	}
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ order-service: DB connect failed: %v", err)
	}
	db.AutoMigrate(&Order{})
	log.Println("✅ order-service: database connected")

	// gRPC client → product-service
	productConn, err := grpc.NewClient(cfg.ProductServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ order-service: failed to dial product-service: %v", err)
	}
	defer productConn.Close()
	productClient := productpb.NewProductServiceClient(productConn)

	// Kafka producer
	producer := appkafka.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50053"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ order-service: listen failed: %v", err)
	}

	srv := grpc.NewServer()
	orderpb.RegisterOrderServiceServer(srv, NewOrderHandler(db, productClient, producer))
	reflection.Register(srv)

	log.Printf("🚀 order-service: gRPC listening on :%s", port)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("❌ order-service: serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 order-service: shutting down...")
	srv.GracefulStop()
}
