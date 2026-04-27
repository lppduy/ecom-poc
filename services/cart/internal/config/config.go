package config

import (
	"os"
	"strings"
)

type Config struct {
	Port          string
	DatabaseURL   string
	DefaultUserID string
	RedisAddr     string
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8082"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
		DefaultUserID: getEnv("DEFAULT_USER_ID", "u_001"),
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
