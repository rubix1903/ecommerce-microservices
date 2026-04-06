package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	orderpb "github.com/ecommerce/microservices/proto/order"
	paymentpb "github.com/ecommerce/microservices/proto/payment"
	productpb "github.com/ecommerce/microservices/proto/product"
	userpb "github.com/ecommerce/microservices/proto/user"
	"github.com/ecommerce/microservices/shared/codec"
	"github.com/ecommerce/microservices/shared/config"
	"github.com/ecommerce/microservices/shared/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func mustDial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, //nolint:staticcheck
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.JSONCodec{})),
	)
	if err != nil {
		log.Fatalf("❌ api-gateway: failed to dial %s: %v", addr, err)
	}
	return conn
}

func main() {
	codec.Register() // must be first

	cfg := config.Load()

	userConn := mustDial(cfg.UserServiceAddr)
	productConn := mustDial(cfg.ProductServiceAddr)
	orderConn := mustDial(cfg.OrderServiceAddr)
	paymentConn := mustDial(cfg.PaymentServiceAddr)
	defer userConn.Close()
	defer productConn.Close()
	defer orderConn.Close()
	defer paymentConn.Close()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}))

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	auth := r.Group("/api/v1/auth")
	{
		h := NewUserHTTPHandler(userpb.NewUserServiceClient(userConn))
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}

	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		uh := NewUserHTTPHandler(userpb.NewUserServiceClient(userConn))
		protected.GET("/users/:id", uh.GetUser)

		ph := NewProductHTTPHandler(productpb.NewProductServiceClient(productConn))
		protected.POST("/products", ph.CreateProduct)
		protected.GET("/products", ph.ListProducts)
		protected.GET("/products/:id", ph.GetProduct)

		oh := NewOrderHTTPHandler(orderpb.NewOrderServiceClient(orderConn))
		protected.POST("/orders", oh.CreateOrder)
		protected.GET("/orders", oh.ListOrders)
		protected.GET("/orders/:id", oh.GetOrder)

		pyh := NewPaymentHTTPHandler(paymentpb.NewPaymentServiceClient(paymentConn))
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
