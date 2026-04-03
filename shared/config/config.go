// Package config loads service configuration from environment variables.
package config

import (
	"os"
	"strings"
)

// Config holds common configuration for all services.
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Kafka
	KafkaBrokers []string

	// JWT
	JWTSecret string

	// gRPC service addresses
	UserServiceAddr    string
	ProductServiceAddr string
	OrderServiceAddr   string
	PaymentServiceAddr string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	brokers := getEnv("KAFKA_BROKERS", "kafka:9092")
	return &Config{
		DBHost:     getEnv("DB_HOST", "postgres"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "ecommerce"),

		KafkaBrokers: strings.Split(brokers, ","),

		JWTSecret: getEnv("JWT_SECRET", "super-secret-change-in-production"),

		UserServiceAddr:    getEnv("USER_SERVICE_ADDR", "user-service:50051"),
		ProductServiceAddr: getEnv("PRODUCT_SERVICE_ADDR", "product-service:50052"),
		OrderServiceAddr:   getEnv("ORDER_SERVICE_ADDR", "order-service:50053"),
		PaymentServiceAddr: getEnv("PAYMENT_SERVICE_ADDR", "payment-service:50054"),
	}
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" port=" + c.DBPort +
		" sslmode=disable TimeZone=UTC"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
