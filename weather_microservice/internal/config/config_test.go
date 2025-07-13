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
				OpenWeatherKey: "abc123",
			},
			wantErr: false,
		},
		{
			name: "missing openweather key",
			config: &Config{
				Port:        "8080",
				Environment: "development",
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

func TestConfig_DefaultValues(t *testing.T) {
	clearEnvVars()
	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %v", cfg.Port)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected default environment development, got %v", cfg.Environment)
	}

	if cfg.Cache.Expiration != 10*60*1e9 {
		t.Errorf("expected default expiration 10 minutes, got %v", cfg.Cache.Expiration)
	}
}

func TestConfig_EnvironmentVariableOverrides(t *testing.T) {
	clearEnvVars()

	envs := map[string]string{
		"PORT":                     "9090",
		"ENVIRONMENT":              "production",
		"OPENWEATHER_API_KEY":      "abc",
		"CACHE_ENABLED":            "true",
		"CACHE_EXPIRATION_MINUTES": "15",
	}

	for k, v := range envs {
		_ = os.Setenv(k, v)
	}

	cfg := Load()
	if cfg == nil {
		t.Fatalf("Load() error: cfg is nil")
	}

	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %v", cfg.Port)
	}

	if cfg.Environment != "production" {
		t.Errorf("unexpected environment: %v", cfg.Environment)
	}

	if !cfg.Cache.Enabled {
		t.Errorf("expected cache to be enabled")
	}

	if cfg.Cache.Expiration.Minutes() != 15 {
		t.Errorf("expected cache expiration 15 minutes, got %v", cfg.Cache.Expiration)
	}
}

func clearEnvVars() {
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("ENVIRONMENT")
	_ = os.Unsetenv("OPENWEATHER_API_KEY")
	_ = os.Unsetenv("CACHE_ENABLED")
	_ = os.Unsetenv("CACHE_EXPIRATION_MINUTES")
}
