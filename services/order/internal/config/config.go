package config

import (
	"os"
	"strings"
)

type Config struct {
	Port             string
	DatabaseURL      string
	CartBaseURL      string
	InventoryBaseURL string
	KafkaBrokers     string
	JWTSecret        string
	RedisAddr        string
}

func Load() Config {
	return Config{
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
		CartBaseURL:      getEnv("CART_BASE_URL", "http://localhost:8082"),
		InventoryBaseURL: getEnv("INVENTORY_BASE_URL", "http://localhost:8084"),
		KafkaBrokers:     getEnv("KAFKA_BROKERS", "localhost:9092"),
		JWTSecret:        getEnv("JWT_SECRET", "supersecret-change-in-prod"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
