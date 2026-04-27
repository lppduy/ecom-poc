package config

import (
	"os"
	"strings"
)

type Config struct {
	Port          string
	ESAddress     string
	CatalogBaseURL string
}

func Load() Config {
	return Config{
		Port:           getEnv("PORT", "8085"),
		ESAddress:      getEnv("ES_ADDRESS", "http://localhost:9200"),
		CatalogBaseURL: getEnv("CATALOG_BASE_URL", "http://localhost:8081"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
