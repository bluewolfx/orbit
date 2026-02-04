package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL      string
	RedisURL         string
	ManagerGRPCPort  string
	ManagerHTTPPort  string
	WorkerPoolSize   int
	PrometheusPort   string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://orbit:orbit@localhost:5432/orbit?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", "localhost:6379"),
		ManagerGRPCPort: getEnv("MANAGER_GRPC_PORT", "50051"),
		ManagerHTTPPort: getEnv("MANAGER_HTTP_PORT", "8080"),
		WorkerPoolSize:  getEnvAsInt("WORKER_POOL_SIZE", 10),
		PrometheusPort:  getEnv("PROMETHEUS_PORT", "9090"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	return nil
}
