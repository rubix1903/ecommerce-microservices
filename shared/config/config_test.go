package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Ensure env is clean for this test
	for _, k := range []string{"DB_HOST", "DB_PORT", "KAFKA_BROKERS", "JWT_SECRET"} {
		os.Unsetenv(k)
	}

	cfg := Load()

	if cfg.DBHost != "postgres" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "postgres")
	}
	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %q, want %q", cfg.DBPort, "5432")
	}
	if len(cfg.KafkaBrokers) != 1 || cfg.KafkaBrokers[0] != "kafka:9092" {
		t.Errorf("KafkaBrokers = %v, want [kafka:9092]", cfg.KafkaBrokers)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("DB_HOST", "myhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("KAFKA_BROKERS", "broker1:9092,broker2:9092")
	os.Setenv("JWT_SECRET", "mysecret")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("KAFKA_BROKERS")
		os.Unsetenv("JWT_SECRET")
	}()

	cfg := Load()

	if cfg.DBHost != "myhost" {
		t.Errorf("DBHost = %q, want %q", cfg.DBHost, "myhost")
	}
	if cfg.DBPort != "5433" {
		t.Errorf("DBPort = %q, want %q", cfg.DBPort, "5433")
	}
	if len(cfg.KafkaBrokers) != 2 {
		t.Errorf("expected 2 brokers, got %d", len(cfg.KafkaBrokers))
	}
	if cfg.JWTSecret != "mysecret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "mysecret")
	}
}

func TestDSN_Format(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "admin")
	os.Setenv("DB_PASSWORD", "pass")
	os.Setenv("DB_NAME", "shop")
	os.Setenv("DB_PORT", "5432")
	defer func() {
		for _, k := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_PORT"} {
			os.Unsetenv(k)
		}
	}()

	cfg := Load()
	dsn := cfg.DSN()

	for _, part := range []string{"host=localhost", "user=admin", "password=pass", "dbname=shop", "port=5432"} {
		found := false
		for i := 0; i <= len(dsn)-len(part); i++ {
			if dsn[i:i+len(part)] == part {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DSN %q does not contain %q", dsn, part)
		}
	}
}
