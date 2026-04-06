package main

import (
	"context"
	"testing"

	userpb "github.com/ecommerce/microservices/proto/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// openTestDB creates an in-memory SQLite database for testing.
// I used SQLite here to avoid needing a running Postgres in CI.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
	return db
}

func TestRegister_Success(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	resp, err := h.Register(context.Background(), &userpb.RegisterRequest{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestRegister_MissingFields(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	_, err := h.Register(context.Background(), &userpb.RegisterRequest{
		Email: "alice@example.com",
		// Name and Password missing
	})
	if err == nil {
		t.Fatal("expected error for missing fields, got nil")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	req := &userpb.RegisterRequest{Name: "Alice", Email: "alice@example.com", Password: "pass"}
	if _, err := h.Register(context.Background(), req); err != nil {
		t.Fatalf("first register failed: %v", err)
	}
	_, err := h.Register(context.Background(), req)
	if err == nil {
		t.Fatal("expected error on duplicate email, got nil")
	}
}

func TestLogin_Success(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	if _, err := h.Register(context.Background(), &userpb.RegisterRequest{
		Name: "Bob", Email: "bob@example.com", Password: "mypassword",
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	resp, err := h.Login(context.Background(), &userpb.LoginRequest{
		Email: "bob@example.com", Password: "mypassword",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty JWT token")
	}
	if resp.UserID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	h.Register(context.Background(), &userpb.RegisterRequest{
		Name: "Charlie", Email: "charlie@example.com", Password: "correct",
	})

	_, err := h.Login(context.Background(), &userpb.LoginRequest{
		Email: "charlie@example.com", Password: "wrong",
	})
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	_, err := h.Login(context.Background(), &userpb.LoginRequest{
		Email: "nobody@example.com", Password: "pass",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}
}

func TestGetUser_Success(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	reg, _ := h.Register(context.Background(), &userpb.RegisterRequest{
		Name: "Dave", Email: "dave@example.com", Password: "pass",
	})

	resp, err := h.GetUser(context.Background(), &userpb.GetUserRequest{ID: reg.ID})
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if resp.Email != "dave@example.com" {
		t.Errorf("got email %q, want %q", resp.Email, "dave@example.com")
	}
}

func TestGetUser_NotFound(t *testing.T) {
	db := openTestDB(t)
	h := NewUserHandler(db, "test-secret")

	_, err := h.GetUser(context.Background(), &userpb.GetUserRequest{ID: "non-existent-id"})
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
}
