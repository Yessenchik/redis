package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	OrderHTTPPort   string
	OrderGRPCPort   string
	PaymentGRPCAddr string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	RedisAddr        string
	OrderCacheTTL    time.Duration
}

func Load() *Config {
	_ = godotenv.Load("../.env")

	cfg := &Config{
		OrderHTTPPort:    getEnv("ORDER_HTTP_PORT", "8080"),
		OrderGRPCPort:    getEnv("ORDER_GRPC_PORT", "50052"),
		PaymentGRPCAddr:  getEnv("PAYMENT_GRPC_ADDR", "localhost:50051"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       getEnv("POSTGRES_DB", "ap2_db"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
	}
	ttl, err := time.ParseDuration(getEnv("ORDER_CACHE_TTL", "5m"))
	if err != nil {
		ttl = 5 * time.Minute
	}
	cfg.OrderCacheTTL = ttl

	log.Printf("order-service config loaded")
	return cfg
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
