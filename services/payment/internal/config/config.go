package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	OrderBaseURL string
	JWTSecret   string
}

func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8086"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
		OrderBaseURL: getEnv("ORDER_BASE_URL", "http://localhost:8083"),
		JWTSecret:    getEnv("JWT_SECRET", "supersecret-change-in-prod"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
