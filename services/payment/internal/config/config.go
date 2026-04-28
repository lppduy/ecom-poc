package config

import "os"

type Config struct {
	Port         string
	DatabaseURL  string
	KafkaBrokers string
	JWTSecret    string
}

func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8087"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		JWTSecret:    getEnv("JWT_SECRET", "supersecret-change-in-prod"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
