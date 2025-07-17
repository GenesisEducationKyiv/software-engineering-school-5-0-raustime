package config

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{OpenWeatherKey: "abc123"}
		if err := cfg.Validate(); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("missing OpenWeatherKey", func(t *testing.T) {
		cfg := &Config{}
		if err := cfg.Validate(); err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestLoad_DefaultValues(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("OPENWEATHER_API_KEY", "abc123") // to pass validation
	t.Setenv("CACHE_ENABLED", "")
	t.Setenv("CACHE_EXPIRATION_MINUTES", "")

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %v", cfg.Port)
	}
	if cfg.Environment != "development" {
		t.Errorf("expected default environment development, got %v", cfg.Environment)
	}
	if cfg.Cache.Expiration != 10*time.Minute {
		t.Errorf("expected expiration 10m, got %v", cfg.Cache.Expiration)
	}
	if cfg.Cache.Enabled {
		t.Errorf("expected cache to be disabled by default")
	}
}

func TestLoad_WithOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("GRPC_PORT", "9091")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("OPENWEATHER_API_KEY", "abc123")
	t.Setenv("CACHE_ENABLED", "true")
	t.Setenv("CACHE_EXPIRATION_MINUTES", "15")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("REDIS_POOL_SIZE", "20")
	t.Setenv("REDIS_TIMEOUT_SECONDS", "3")
	t.Setenv("REDIS_ADDR", "redis:6379")
	t.Setenv("REDIS_PASSWORD", "secret")

	cfg := Load()

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %v", cfg.Port)
	}
	if cfg.GRPCPort != "9091" {
		t.Errorf("expected port 9091, got %v", cfg.Port)
	}
	if cfg.Environment != "production" {
		t.Errorf("expected production env, got %v", cfg.Environment)
	}
	if !cfg.Cache.Enabled {
		t.Errorf("expected cache to be enabled")
	}
	if cfg.Cache.Expiration != 15*time.Minute {
		t.Errorf("expected expiration 15m, got %v", cfg.Cache.Expiration)
	}
	if cfg.Cache.Redis.DB != 2 {
		t.Errorf("expected Redis DB 2, got %v", cfg.Cache.Redis.DB)
	}
	if cfg.Cache.Redis.PoolSize != 20 {
		t.Errorf("expected Redis pool 20, got %v", cfg.Cache.Redis.PoolSize)
	}
	if cfg.Cache.Redis.Timeout != 3*time.Second {
		t.Errorf("expected Redis timeout 3s, got %v", cfg.Cache.Redis.Timeout)
	}
	if cfg.Cache.Redis.Addr != "redis:6379" {
		t.Errorf("expected Redis addr redis:6379, got %v", cfg.Cache.Redis.Addr)
	}
	if cfg.Cache.Redis.Password != "secret" {
		t.Errorf("expected Redis password secret, got %v", cfg.Cache.Redis.Password)
	}
}
