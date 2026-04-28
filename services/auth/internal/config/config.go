package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	RedisAddr   string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8087"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "supersecret-change-in-prod"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
