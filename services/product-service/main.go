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
	productpb "github.com/ecommerce/microservices/proto/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ─── Model ───────────────────────────────────────────────────────────────────

type Product struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `gorm:"not null"`
	Description string
	Price       float64 `gorm:"not null"`
	Stock       int32   `gorm:"not null;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// ─── Handler ─────────────────────────────────────────────────────────────────

type ProductHandler struct {
	productpb.UnimplementedProductServiceServer
	db *gorm.DB
}

func NewProductHandler(db *gorm.DB) *ProductHandler { return &ProductHandler{db: db} }

func (h *ProductHandler) CreateProduct(_ context.Context, req *productpb.CreateProductRequest) (*productpb.CreateProductResponse, error) {
	if req.Name == "" || req.Price <= 0 {
		return nil, status.Error(codes.InvalidArgument, "name and a positive price are required")
	}
	p := &Product{Name: req.Name, Description: req.Description, Price: req.Price, Stock: req.Stock}
	if err := h.db.Create(p).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}
	return &productpb.CreateProductResponse{ID: p.ID, Message: "product created"}, nil
}

func (h *ProductHandler) GetProduct(_ context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {
	var p Product
	if err := h.db.First(&p, "id = ?", req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}
	return toProtoProduct(&p), nil
}

func (h *ProductHandler) ListProducts(_ context.Context, req *productpb.ListProductsRequest) (*productpb.ListProductsResponse, error) {
	page, limit := req.Page, req.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var products []Product
	var total int64
	h.db.Model(&Product{}).Count(&total)
	if err := h.db.Offset(int(offset)).Limit(int(limit)).Find(&products).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to list products")
	}

	out := make([]*productpb.Product, len(products))
	for i, p := range products {
		out[i] = toProtoProduct(&p)
	}
	return &productpb.ListProductsResponse{Products: out, Total: total}, nil
}

// DeductStock atomically reduces the stock for an ordered product.
// Uses a DB transaction to prevent race conditions with concurrent orders.
func (h *ProductHandler) DeductStock(_ context.Context, req *productpb.DeductStockRequest) (*productpb.DeductStockResponse, error) {
	var remaining int32
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var p Product
		// SELECT ... FOR UPDATE prevents concurrent over-deductions
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&p, "id = ?", req.ProductID).Error; err != nil {
			return err
		}
		if p.Stock < req.Quantity {
			return status.Errorf(codes.FailedPrecondition, "insufficient stock: have %d, need %d", p.Stock, req.Quantity)
		}
		p.Stock -= req.Quantity
		remaining = p.Stock
		return tx.Save(&p).Error
	})
	if err != nil {
		if s, ok := status.FromError(err); ok {
			return nil, s.Err()
		}
		return nil, status.Errorf(codes.Internal, "stock deduction failed: %v", err)
	}
	return &productpb.DeductStockResponse{Success: true, RemainingStock: remaining}, nil
}

func toProtoProduct(p *Product) *productpb.Product {
	return &productpb.Product{ID: p.ID, Name: p.Name, Description: p.Description, Price: p.Price, Stock: p.Stock}
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ product-service: DB connect failed: %v", err)
	}
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatalf("❌ product-service: automigrate failed: %v", err)
	}
	log.Println("✅ product-service: database connected")

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50052"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ product-service: listen failed: %v", err)
	}

	srv := grpc.NewServer()
	productpb.RegisterProductServiceServer(srv, NewProductHandler(db))
	reflection.Register(srv)

	log.Printf("🚀 product-service: gRPC listening on :%s", port)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("❌ product-service: serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 product-service: shutting down...")
	srv.GracefulStop()
}
