package config

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	GRPCPort    string
	DatabaseURL string
	RedisAddr   string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8084"),
		GRPCPort:    getEnv("GRPC_PORT", "9084"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
