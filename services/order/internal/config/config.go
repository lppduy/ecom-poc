package config

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
	CartBaseURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
		CartBaseURL: getEnv("CART_BASE_URL", "http://localhost:8082"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
