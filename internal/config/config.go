package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppBaseURL      string
	Port            string
	DatabaseURL     string
	DatabaseTestURL string
	OpenWeatherKey  string
	SMTPHost        string
	SMTPPort        int
	SMTPUser        string
	SMTPPassword    string
	Environment     string
	BunDebugMode    string
}

// Load завантажує конфігурацію з змінних оточення
func Load() (*Config, error) {
	smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	cfg := &Config{
		AppBaseURL:      getEnv("APP_BASE_URL", "http://localhost:8080"),
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DB_URL", ""),
		DatabaseTestURL: getEnv("TEST_DB_URL", ""),
		OpenWeatherKey:  getEnv("OPENWEATHER_API_KEY", ""),
		SMTPHost:        getEnv("SMTP_HOST", ""),
		SMTPPort:        smtpPort,
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPassword:    getEnv("SMTP_PASSWORD", ""),
		Environment:     strings.ToLower(getEnv("ENVIRONMENT", "development")),
		BunDebugMode:    getEnv("BUNDEBUG", "0"),
	}

	return cfg, nil
}

// IsProduction перевіряє чи додаток працює в продакшен режимі
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment перевіряє чи додаток працює в режимі розробки
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsTest перевіряє чи додаток працює в тестовому режимі
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

// GetDatabaseURL повертає URL бази даних в залежності від середовища
func (c *Config) GetDatabaseURL() string {
	if c.IsTest() && c.DatabaseTestURL != "" {
		return c.DatabaseTestURL
	}
	return c.DatabaseURL
}

// IsBunDebugEnabled перевіряє чи включений debug режим для Bun ORM
func (c *Config) IsBunDebugEnabled() bool {
	return strings.ToLower(c.BunDebugMode) == "1"
}

// Validate перевіряє чи всі обов'язкові конфігурації встановлені
func (c *Config) Validate() error {
	var errors []string

	if c.DatabaseURL == "" {
		errors = append(errors, "DB_URL is required")
	}

	if c.OpenWeatherKey == "" {
		errors = append(errors, "OPENWEATHER_API_KEY is required")
	}

	if c.SMTPHost != "" && c.SMTPUser == "" {
		errors = append(errors, "SMTP_USER is required when SMTP_HOST is set")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// getEnv отримує значення змінної оточення або повертає значення за замовчуванням
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
