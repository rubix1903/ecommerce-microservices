package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/ecommerce/microservices/shared/codec"
	"github.com/ecommerce/microservices/shared/config"
	"github.com/ecommerce/microservices/shared/middleware"
	userpb "github.com/ecommerce/microservices/proto/user"
	productpb "github.com/ecommerce/microservices/proto/product"
	orderpb "github.com/ecommerce/microservices/proto/order"
	paymentpb "github.com/ecommerce/microservices/proto/payment"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("❌ api-gateway: failed to dial %s: %v", addr, err)
	}
	return conn
}

func main() {
	cfg := config.Load()

	// ── gRPC clients ─────────────────────────────────────────────────────────
	userConn := mustDial(cfg.UserServiceAddr)
	productConn := mustDial(cfg.ProductServiceAddr)
	orderConn := mustDial(cfg.OrderServiceAddr)
	paymentConn := mustDial(cfg.PaymentServiceAddr)
	defer userConn.Close()
	defer productConn.Close()
	defer orderConn.Close()
	defer paymentConn.Close()

	userClient := userpb.NewUserServiceClient(userConn)
	productClient := productpb.NewProductServiceClient(productConn)
	orderClient := orderpb.NewOrderServiceClient(orderConn)
	paymentClient := paymentpb.NewPaymentServiceClient(paymentConn)

	// ── Gin router ───────────────────────────────────────────────────────────
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"}) })

	// ── Auth routes (no JWT required) ────────────────────────────────────────
	auth := r.Group("/api/v1/auth")
	{
		h := NewUserHTTPHandler(userClient)
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}

	// ── Protected routes (JWT required) ──────────────────────────────────────
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Users
		uh := NewUserHTTPHandler(userClient)
		protected.GET("/users/:id", uh.GetUser)

		// Products
		ph := NewProductHTTPHandler(productClient)
		protected.POST("/products", ph.CreateProduct)
		protected.GET("/products", ph.ListProducts)
		protected.GET("/products/:id", ph.GetProduct)

		// Orders
		oh := NewOrderHTTPHandler(orderClient)
		protected.POST("/orders", oh.CreateOrder)
		protected.GET("/orders", oh.ListOrders)
		protected.GET("/orders/:id", oh.GetOrder)

		// Payments
		pyh := NewPaymentHTTPHandler(paymentClient)
		protected.GET("/payments/:id", pyh.GetPayment)
	}

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 api-gateway: HTTP listening on :%s", port)
	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("❌ api-gateway: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 api-gateway: shutting down...")
}
