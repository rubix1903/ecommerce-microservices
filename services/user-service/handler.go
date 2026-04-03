package main

import (
	"context"
	"errors"
	"time"

	userpb "github.com/ecommerce/microservices/proto/user"
	"github.com/ecommerce/microservices/shared/middleware"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// UserHandler implements userpb.UserServiceServer.
type UserHandler struct {
	userpb.UnimplementedUserServiceServer
	db        *gorm.DB
	jwtSecret string
}

func NewUserHandler(db *gorm.DB, jwtSecret string) *UserHandler {
	return &UserHandler{db: db, jwtSecret: jwtSecret}
}

// Register creates a new user with a bcrypt-hashed password.
func (h *UserHandler) Register(_ context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name, email, and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := &User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hash),
	}

	if err := h.db.Create(user).Error; err != nil {
		return nil, status.Errorf(codes.AlreadyExists, "email already registered: %v", err)
	}

	return &userpb.RegisterResponse{ID: user.ID, Message: "user registered successfully"}, nil
}

// Login validates credentials and returns a signed JWT.
func (h *UserHandler) Login(_ context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	var user User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "invalid credentials")
		}
		return nil, status.Error(codes.Internal, "database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to sign token")
	}

	return &userpb.LoginResponse{Token: signed, UserID: user.ID}, nil
}

// GetUser fetches a user's public profile by ID.
func (h *UserHandler) GetUser(_ context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	if req.ID == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	var user User
	if err := h.db.First(&user, "id = ?", req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "database error")
	}

	return &userpb.GetUserResponse{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}
