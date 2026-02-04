package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.DatabaseURL == "" {
		t.Error("DatabaseURL should not be empty")
	}

	if cfg.RedisURL == "" {
		t.Error("RedisURL should not be empty")
	}

	if cfg.ManagerGRPCPort == "" {
		t.Error("ManagerGRPCPort should not be empty")
	}

	if cfg.ManagerHTTPPort == "" {
		t.Error("ManagerHTTPPort should not be empty")
	}

	if cfg.WorkerPoolSize <= 0 {
		t.Error("WorkerPoolSize should be positive")
	}
}

func TestLoadWithEnv(t *testing.T) {
	os.Setenv("WORKER_POOL_SIZE", "20")
	defer os.Unsetenv("WORKER_POOL_SIZE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.WorkerPoolSize != 20 {
		t.Errorf("Expected WorkerPoolSize to be 20, got %d", cfg.WorkerPoolSize)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		expectError bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				DatabaseURL: "postgres://test",
				RedisURL:    "localhost:6379",
			},
			expectError: false,
		},
		{
			name: "missing database URL",
			cfg: &Config{
				DatabaseURL: "",
				RedisURL:    "localhost:6379",
			},
			expectError: true,
		},
		{
			name: "missing redis URL",
			cfg: &Config{
				DatabaseURL: "postgres://test",
				RedisURL:    "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
