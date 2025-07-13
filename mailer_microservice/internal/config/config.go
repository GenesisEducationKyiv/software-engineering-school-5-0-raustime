package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppBaseURL   string
	Port         string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	Environment  string
}

// Load завантажує конфігурацію з змінних оточення.
func Load() *Config {
	smtpPortStr := getEnv("SMTP_PORT", "587")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid SMTP_PORT: %s (defaulting to 587)\n", smtpPortStr)
		smtpPort = 587
	}

	return &Config{
		AppBaseURL:   getEnv("APP_BASE_URL", "http://localhost:8089"),
		Port:         getEnv("PORT", "8089"),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     smtpPort,
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		Environment:  strings.ToLower(getEnv("ENVIRONMENT", "development")),
	}
}

func LoadTestConfig() *Config {
	cfg := Load()
	cfg.Environment = "test"
	return cfg
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

// Validate перевіряє чи всі обов'язкові конфігурації встановлені.
func (c *Config) Validate() error {
	var errors []string

	if c.SMTPHost != "" && c.SMTPUser == "" {
		errors = append(errors, "SMTP_USER is required when SMTP_HOST is set")
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
