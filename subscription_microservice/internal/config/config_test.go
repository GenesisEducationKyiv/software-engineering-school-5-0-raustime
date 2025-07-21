package config_test

import (
	"os"
	"testing"
	"log"

	"github.com/stretchr/testify/require"

	"subscription_microservice/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg := config.Load()
	require.Equal(t, "8090", cfg.GrpcPort)
	require.Equal(t, "8091", cfg.HttpPort)
	require.Equal(t, "http://localhost:8089", cfg.MailerGRPCAddr)
	require.Equal(t, "", cfg.DatabaseURL)
	require.Equal(t, "", cfg.DatabaseTestURL)
	require.Equal(t, "development", cfg.Environment)
	require.False(t, cfg.IsBunDebugEnabled())
}

func TestLoad_WithEnv(t *testing.T) {
	mustSetEnv("GRPC_PORT", "9000")
	mustSetEnv("HTTP_PORT", "9001")
	mustSetEnv("MAILER_GRPC_URL", "https://mailer.example.com")
	mustSetEnv("DB_URL", "postgres://user:pass@localhost:5432/db")
	mustSetEnv("ENVIRONMENT", "production")
	mustSetEnv("BUNDEBUG", "true")

	cfg := config.Load()
	require.Equal(t, "9000", cfg.GrpcPort)
	require.Equal(t, "9001", cfg.HttpPort)
	require.Equal(t, "https://mailer.example.com", cfg.MailerGRPCAddr)
	require.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DatabaseURL)
	require.Equal(t, "production", cfg.Environment)
	require.True(t, cfg.IsBunDebugEnabled())
	require.True(t, cfg.IsProduction())
	require.False(t, cfg.IsDevelopment())
	require.False(t, cfg.IsTest())
}

func mustSetEnv(key, value string) {
	if err := os.Setenv(key, value); err != nil {
		log.Fatalf("failed to set %s: %v", key, err)
	}
}

func TestValidate_DBRequired(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL: "",
	}
	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "DB_URL is required")
}

func TestGetDatabaseURL_PrefersTestInTestMode(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL:     "postgres://main",
		DatabaseTestURL: "postgres://test",
		Environment:     "test",
	}
	require.Equal(t, "postgres://test", cfg.GetDatabaseURL())

	cfg.Environment = "production"
	require.Equal(t, "postgres://main", cfg.GetDatabaseURL())
}
