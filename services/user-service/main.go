package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	userpb "github.com/ecommerce/microservices/proto/user"
	"github.com/ecommerce/microservices/shared/codec"
	"github.com/ecommerce/microservices/shared/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	codec.Register() // must be first — overrides gRPC's built-in proto codec

	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ user-service: failed to connect to DB: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("❌ user-service: automigrate failed: %v", err)
	}
	log.Println("✅ user-service: database connected")

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
	reflection.Register(srv)

	log.Printf("🚀 user-service: gRPC listening on :%s", port)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("❌ user-service: serve error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 user-service: shutting down...")
	srv.GracefulStop()
}
