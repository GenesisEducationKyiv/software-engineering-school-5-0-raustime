package config

import (
	"os"
	"testing"
)

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
			name: "missing required DB_URL",
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
		"DB_URL":              "postgres://test:test@localhost/testdb",
		"ENVIRONMENT":         "production",
		"OPENWEATHER_API_KEY": "abc",
		"BUNDEBUG":            "1",
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
	_ = os.Unsetenv("DB_URL")
	_ = os.Unsetenv("ENVIRONMENT")
	_ = os.Unsetenv("BUNDEBUG")
	_ = os.Unsetenv("OPENWEATHER_API_KEY")
}
