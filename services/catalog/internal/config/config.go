package config

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ecom:ecom@localhost:5432/ecom?sslmode=disable"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
