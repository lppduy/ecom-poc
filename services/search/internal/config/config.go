package config

import (
	"os"
	"strings"
)

type Config struct {
	Port             string
	ESAddress        string
	CatalogGRPCAddr  string
}

func Load() Config {
	return Config{
		Port:            getEnv("PORT", "8085"),
		ESAddress:       getEnv("ES_ADDRESS", "http://localhost:9200"),
		CatalogGRPCAddr: getEnv("CATALOG_GRPC_ADDR", "localhost:9081"),
	}
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}
