package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/ecommerce/microservices/shared/codec" // register JSON gRPC codec
	"github.com/ecommerce/microservices/shared/config"
	userpb "github.com/ecommerce/microservices/proto/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	// ── Database ─────────────────────────────────────────────────────────────
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ user-service: failed to connect to DB: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("❌ user-service: automigrate failed: %v", err)
	}
	log.Println("✅ user-service: database connected")

	// ── gRPC server ───────────────────────────────────────────────────────────
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("❌ user-service: failed to listen: %v", err)
	}

	srv := grpc.NewServer()
	userpb.RegisterUserServiceServer(srv, NewUserHandler(db, cfg.JWTSecret))
	reflection.Register(srv) // enables grpcurl introspection

	log.Printf("🚀 user-service: gRPC listening on :%s", port)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("❌ user-service: serve error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 user-service: shutting down...")
	srv.GracefulStop()
}
