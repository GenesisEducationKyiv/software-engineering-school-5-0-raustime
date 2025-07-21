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
			name: "valid config with SMTP",
			config: &Config{
				Port:        "8080",
				Environment: "development",
				SMTPHost:    "smtp.example.com",
				SMTPUser:    "user",
			},
			wantErr: false,
		},
		{
			name: "missing SMTP_USER",
			config: &Config{
				Port:        "8080",
				Environment: "development",
				SMTPHost:    "smtp.example.com",
			},
			wantErr: true,
		},
		{
			name: "SMTP empty",
			config: &Config{
				Port:        "8080",
				Environment: "development",
			},
			wantErr: false,
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

func TestConfig_LoadDefaults(t *testing.T) {
	clearEnvVars()
	cfg := Load()

	if cfg == nil {
		t.Fatal("Load() returned nil")
	}

	if cfg.Port != "8089" {
		t.Errorf("expected default port 8089, got %v", cfg.Port)
	}

	if cfg.Environment != "development" {
		t.Errorf("expected default environment 'development', got %v", cfg.Environment)
	}
}

func TestConfig_EnvOverrides(t *testing.T) {
	clearEnvVars()
	_ = os.Setenv("PORT", "9999")
	_ = os.Setenv("ENVIRONMENT", "test")
	_ = os.Setenv("SMTP_HOST", "smtp.test.com")
	_ = os.Setenv("SMTP_USER", "tester")
	_ = os.Setenv("SMTP_PASSWORD", "pass")
	_ = os.Setenv("SMTP_PORT", "2525")

	cfg := Load()

	if cfg.Port != "9999" {
		t.Errorf("expected port 9999, got %v", cfg.Port)
	}

	if cfg.Environment != "test" {
		t.Errorf("expected environment 'test', got %v", cfg.Environment)
	}

	if cfg.SMTPHost != "smtp.test.com" {
		t.Errorf("unexpected SMTPHost: %v", cfg.SMTPHost)
	}

	if cfg.SMTPUser != "tester" {
		t.Errorf("unexpected SMTPUser: %v", cfg.SMTPUser)
	}

	if cfg.SMTPPassword != "pass" {
		t.Errorf("unexpected SMTPPassword: %v", cfg.SMTPPassword)
	}

	if cfg.SMTPPort != 2525 {
		t.Errorf("expected SMTPPort 2525, got %v", cfg.SMTPPort)
	}
}

func clearEnvVars() {
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("ENVIRONMENT")
	_ = os.Unsetenv("SMTP_HOST")
	_ = os.Unsetenv("SMTP_USER")
	_ = os.Unsetenv("SMTP_PASSWORD")
	_ = os.Unsetenv("SMTP_PORT")
}
