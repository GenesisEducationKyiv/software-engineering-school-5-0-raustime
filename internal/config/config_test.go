package config

import (
	"os"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"PORT":         os.Getenv("PORT"),
		"ENVIRONMENT":  os.Getenv("ENVIRONMENT"),
		"DATABASE_URL": os.Getenv("DATABASE_URL"),
		"BUN_DEBUG":    os.Getenv("BUN_DEBUG"),
	}

	// Clean up after test
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name      string
		envVars   map[string]string
		expectErr bool
	}{
		{
			name: "valid config with all env vars",
			envVars: map[string]string{
				"PORT":         "8080",
				"ENVIRONMENT":  "development",
				"DATABASE_URL": "postgres://user:pass@localhost/db",
				"BUN_DEBUG":    "true",
			},
			expectErr: false,
		},
		{
			name: "valid config with defaults",
			envVars: map[string]string{
				"DATABASE_URL": "postgres://user:pass@localhost/db",
			},
			expectErr: false,
		},
		{
			name: "missing required DATABASE_URL",
			envVars: map[string]string{
				"PORT":        "8080",
				"ENVIRONMENT": "development",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			for key := range originalVars {
				os.Unsetenv(key)
			}

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg, err := Load()

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Validate loaded config
			if cfg.DatabaseURL == "" {
				t.Error("DatabaseURL should not be empty")
			}

			// Check defaults
			if cfg.Port == "" {
				t.Error("Port should have default value")
			}

			if cfg.Environment == "" {
				t.Error("Environment should have default value")
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:         "8080",
				Environment:  "development",
				DatabaseURL:  "postgres://user:pass@localhost/db",
				BunDebugMode: "false",
			},
			expectErr: false,
		},
		{
			name: "missing port",
			config: &Config{
				Environment: "development",
				DatabaseURL: "postgres://user:pass@localhost/db",
			},
			expectErr: true,
		},
		{
			name: "missing environment",
			config: &Config{
				Port:        "8080",
				DatabaseURL: "postgres://user:pass@localhost/db",
			},
			expectErr: true,
		},
		{
			name: "missing database URL",
			config: &Config{
				Port:        "8080",
				Environment: "development",
			},
			expectErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Port:        "invalid",
				Environment: "development",
				DatabaseURL: "postgres://user:pass@localhost/db",
			},
			expectErr: true,
		},
		{
			name: "port out of range",
			config: &Config{
				Port:        "99999",
				Environment: "development",
				DatabaseURL: "postgres://user:pass@localhost/db",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectErr && err == nil {
				t.Error("expected validation error but got none")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestConfig_GetDatabaseURL(t *testing.T) {
	config := &Config{
		DatabaseURL: "postgres://user:pass@localhost/testdb",
	}

	url := config.GetDatabaseURL()
	if url != config.DatabaseURL {
		t.Errorf("expected %s, got %s", config.DatabaseURL, url)
	}
}

func TestConfig_IsBunDebugEnabled(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name: "debug enabled",
			config: &Config{
				BunDebugMode: "true",
			},
			expected: true,
		},
		{
			name: "debug disabled",
			config: &Config{
				BunDebugMode: "false",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsBunDebugEnabled()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	// Clear environment variables
	envVars := []string{"PORT", "ENVIRONMENT", "BUN_DEBUG"}
	originalValues := make(map[string]string)

	for _, envVar := range envVars {
		originalValues[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	// Clean up after test
	defer func() {
		for envVar, originalValue := range originalValues {
			if originalValue == "" {
				os.Unsetenv(envVar)
			} else {
				os.Setenv(envVar, originalValue)
			}
		}
	}()

	// Set required DATABASE_URL
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost/db")
	defer os.Unsetenv("DATABASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check default values
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected default environment development, got %s", cfg.Environment)
	}

	if cfg.BunDebugMode != "false" {
		t.Errorf("expected default BunDebug false, got %v", cfg.BunDebugMode)
	}
}

func TestConfig_EnvironmentVariableOverrides(t *testing.T) {
	// Set environment variables
	testVars := map[string]string{
		"PORT":         "3000",
		"ENVIRONMENT":  "production",
		"DATABASE_URL": "postgres://prod:pass@prod-host/proddb",
		"BUN_DEBUG":    "true",
	}

	for key, value := range testVars {
		os.Setenv(key, value)
	}

	// Clean up after test
	defer func() {
		for key := range testVars {
			os.Unsetenv(key)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify environment variables override defaults
	if cfg.Port != "3000" {
		t.Errorf("expected port 3000, got %s", cfg.Port)
	}

	if cfg.Environment != "production" {
		t.Errorf("expected environment production, got %s", cfg.Environment)
	}

	if cfg.DatabaseURL != "postgres://prod:pass@prod-host/proddb" {
		t.Errorf("expected production database URL, got %s", cfg.DatabaseURL)
	}

	if cfg.BunDebugMode != "true" {
		t.Errorf("expected BunDebug true, got %v", cfg.BunDebugMode)
	}
}
