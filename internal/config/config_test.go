package config

import (
	"os"
	"testing"
)

func TestConfig_Load(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		expectErr bool
	}{
		{
			name: "valid config",
			envVars: map[string]string{
				"PORT":                "8080",
				"ENVIRONMENT":         "development",
				"DATABASE_URL":        "postgres://user:pass@localhost/db",
				"OPENWEATHER_API_KEY": "abc123",
			},
			expectErr: false,
		},
		{
			name: "missing required DATABASE_URL",
			envVars: map[string]string{
				"PORT":                "8080",
				"ENVIRONMENT":         "development",
				"OPENWEATHER_API_KEY": "abc123",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			clearEnvVars()
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			cfg, err := Load()
			if tt.expectErr {
				if err != nil {
					return // expected error
				}
				// Validate separately
				if err := cfg.Validate(); err == nil {
					t.Error("expected validation error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:           "8080",
				Environment:    "development",
				DatabaseURL:    "postgres://user:pass@localhost/db",
				OpenWeatherKey: "abc123",
			},
			wantErr: false,
		},
		{
			name: "missing database URL",
			config: &Config{
				Port:           "8080",
				Environment:    "development",
				OpenWeatherKey: "abc123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
				BunDebugMode: "1",
			},
			expected: true,
		},
		{
			name: "debug disabled",
			config: &Config{
				BunDebugMode: "0",
			},
			expected: false,
		},
		{
			name: "case insensitive true",
			config: &Config{
				BunDebugMode: "TRUE",
			},
			expected: true,
		},
		{
			name: "alternative yes",
			config: &Config{
				BunDebugMode: "yes",
			},
			expected: true,
		},
		{
			name: "empty string",
			config: &Config{
				BunDebugMode: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsBunDebugEnabled()
			if result != tt.expected {
				t.Errorf("IsBunDebugEnabled() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	clearEnvVars()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %v", cfg.Port)
	}

	if cfg.BunDebugMode != "0" {
		t.Errorf("expected default BunDebug 0, got %v", cfg.BunDebugMode)
	}
}

func TestConfig_EnvironmentVariableOverrides(t *testing.T) {
	clearEnvVars()

	envs := map[string]string{
		"PORT":                "9090",
		"DATABASE_URL":        "postgres://test:test@localhost/testdb",
		"ENVIRONMENT":         "production",
		"OPENWEATHER_API_KEY": "abc",
		"BUN_DEBUG":           "1",
	}

	for k, v := range envs {
		_ = os.Setenv(k, v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %v", cfg.Port)
	}

	if cfg.DatabaseURL != "postgres://test:test@localhost/testdb" {
		t.Errorf("unexpected db url: %v", cfg.DatabaseURL)
	}

	if cfg.Environment != "production" {
		t.Errorf("unexpected environment: %v", cfg.Environment)
	}

	if cfg.BunDebugMode != "1" {
		t.Errorf("expected BunDebug 1, got %v", cfg.BunDebugMode)
	}
}

func clearEnvVars() {
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("DATABASE_URL")
	_ = os.Unsetenv("ENVIRONMENT")
	_ = os.Unsetenv("BUN_DEBUG")
	_ = os.Unsetenv("OPENWEATHER_API_KEY")
}
