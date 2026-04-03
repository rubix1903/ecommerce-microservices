package main

import (
	"context"
	"testing"

	productpb "github.com/ecommerce/microservices/proto/product"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.AutoMigrate(&Product{})
	return db
}

func createTestProduct(t *testing.T, h *ProductHandler, name string, price float64, stock int32) string {
	t.Helper()
	resp, err := h.CreateProduct(context.Background(), &productpb.CreateProductRequest{
		Name: name, Price: price, Stock: stock,
	})
	if err != nil {
		t.Fatalf("createTestProduct: %v", err)
	}
	return resp.ID
}

func TestCreateProduct_Success(t *testing.T) {
	h := NewProductHandler(openTestDB(t))

	resp, err := h.CreateProduct(context.Background(), &productpb.CreateProductRequest{
		Name: "MacBook Pro", Description: "M3 chip", Price: 2499.99, Stock: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty product ID")
	}
}

func TestCreateProduct_InvalidArgs(t *testing.T) {
	h := NewProductHandler(openTestDB(t))

	// Missing name
	if _, err := h.CreateProduct(context.Background(), &productpb.CreateProductRequest{Price: 10}); err == nil {
		t.Error("expected error for missing name")
	}
	// Zero price
	if _, err := h.CreateProduct(context.Background(), &productpb.CreateProductRequest{Name: "X", Price: 0}); err == nil {
		t.Error("expected error for zero price")
	}
}

func TestGetProduct_Success(t *testing.T) {
	h := NewProductHandler(openTestDB(t))
	id := createTestProduct(t, h, "iPhone 15", 999.99, 100)

	resp, err := h.GetProduct(context.Background(), &productpb.GetProductRequest{ID: id})
	if err != nil {
		t.Fatalf("GetProduct: %v", err)
	}
	if resp.Name != "iPhone 15" {
		t.Errorf("got %q, want %q", resp.Name, "iPhone 15")
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	h := NewProductHandler(openTestDB(t))

	_, err := h.GetProduct(context.Background(), &productpb.GetProductRequest{ID: "does-not-exist"})
	if err == nil {
		t.Error("expected not-found error")
	}
}

func TestListProducts_Pagination(t *testing.T) {
	h := NewProductHandler(openTestDB(t))
	for i := 0; i < 5; i++ {
		createTestProduct(t, h, "Product", 10.0, 10)
	}

	resp, err := h.ListProducts(context.Background(), &productpb.ListProductsRequest{Page: 1, Limit: 3})
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if len(resp.Products) != 3 {
		t.Errorf("got %d products, want 3", len(resp.Products))
	}
	if resp.Total != 5 {
		t.Errorf("got total %d, want 5", resp.Total)
	}
}

func TestDeductStock_Success(t *testing.T) {
	h := NewProductHandler(openTestDB(t))
	id := createTestProduct(t, h, "Widget", 5.00, 20)

	resp, err := h.DeductStock(context.Background(), &productpb.DeductStockRequest{
		ProductID: id, Quantity: 7,
	})
	if err != nil {
		t.Fatalf("DeductStock: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.RemainingStock != 13 {
		t.Errorf("remaining stock = %d, want 13", resp.RemainingStock)
	}
}

func TestDeductStock_InsufficientStock(t *testing.T) {
	h := NewProductHandler(openTestDB(t))
	id := createTestProduct(t, h, "Rare Item", 99.99, 3)

	_, err := h.DeductStock(context.Background(), &productpb.DeductStockRequest{
		ProductID: id, Quantity: 10,
	})
	if err == nil {
		t.Error("expected error for insufficient stock, got nil")
	}
}

func TestDeductStock_ExactQuantity(t *testing.T) {
	h := NewProductHandler(openTestDB(t))
	id := createTestProduct(t, h, "Last One", 50.00, 1)

	resp, err := h.DeductStock(context.Background(), &productpb.DeductStockRequest{
		ProductID: id, Quantity: 1,
	})
	if err != nil {
		t.Fatalf("DeductStock: %v", err)
	}
	if resp.RemainingStock != 0 {
		t.Errorf("remaining = %d, want 0", resp.RemainingStock)
	}
}
