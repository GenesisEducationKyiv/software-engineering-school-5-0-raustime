package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	GrpcPort        string
	HttpPort        string
	MailerGRPCAddr  string
	DatabaseURL     string
	DatabaseTestURL string
	Environment     string
	BunDebugMode    string `env:"BUNDEBUG"`
}

// Load завантажує конфігурацію з змінних оточення.
func Load() *Config {

	return &Config{
		GrpcPort:        getEnv("GRPC_PORT", "8090"),
		HttpPort:        getEnv("HTTP_PORT", "8091"),
		MailerGRPCAddr:  getEnv("MAILER_GRPC_URL", "http://localhost:8089"),
		DatabaseURL:     getEnv("DB_URL", ""),
		DatabaseTestURL: getEnv("TEST_DB_URL", ""),
		Environment:     strings.ToLower(getEnv("ENVIRONMENT", "development")),
		BunDebugMode:    getEnv("BUNDEBUG", "0"),
	}

}

// IsProduction перевіряє чи додаток працює в продакшен режимі.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment перевіряє чи додаток працює в режимі розробки.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsTest перевіряє чи додаток працює в тестовому режимі.
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

// GetDatabaseURL повертає URL бази даних в залежності від середовища.
func (c *Config) GetDatabaseURL() string {
	if c.IsTest() && c.DatabaseTestURL != "" {
		return c.DatabaseTestURL
	}
	return c.DatabaseURL
}

// IsBunDebugEnabled перевіряє чи включений debug режим для Bun ORM.
func (c *Config) IsBunDebugEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(c.BunDebugMode)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// Validate перевіряє чи всі обов'язкові конфігурації встановлені.
func (c *Config) Validate() error {
	var errors []string

	if c.DatabaseURL == "" {
		errors = append(errors, "DB_URL is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// getEnv отримує значення змінної оточення або повертає значення за замовчуванням.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
