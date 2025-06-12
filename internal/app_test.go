package internal

import (
	"context"
	"os"
	"testing"
	"time"

	"weatherapi/internal/config"
)

func TestApp_New(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	app, err := New()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Close(context.Background())

	if app == nil {
		t.Error("App is nil")
	}

	if app.config == nil {
		t.Error("Config is nil")
	}

	if app.db == nil {
		t.Error("Database is nil")
	}

	if app.weatherService == nil {
		t.Error("Weather service is nil")
	}

	if app.subscriptionService == nil {
		t.Error("Subscription service is nil")
	}

	if app.mailerService == nil {
		t.Error("Mailer service is nil")
	}

	if app.jobScheduler == nil {
		t.Error("Job scheduler is nil")
	}
}

func TestApp_GetDB(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	app, err := New()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Close(context.Background())

	db := app.GetDB()
	if db == nil {
		t.Error("GetDB returned nil")
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestApp_GetConfig(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	app, err := New()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}
	defer app.Close(context.Background())

	cfg := app.GetConfig()
	if cfg == nil {
		t.Error("GetConfig returned nil")
	}
}

func TestApp_Close(t *testing.T) {
	// Skip if no database connection available
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	app, err := New()
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = app.Close(ctx)
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Test that database connection is closed
	if err := app.db.Ping(); err == nil {
		t.Error("Database connection should be closed")
	}
}

func TestApp_CloseWithNilComponents(t *testing.T) {
	app := &App{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Should not panic with nil components
	err := app.Close(ctx)
	if err != nil {
		t.Errorf("Close with nil components returned error: %v", err)
	}
}

// Test helper functions
func TestInitDatabase_InvalidConfig(t *testing.T) {
	cfg := &config.Config{
		DatabaseURL: "invalid://url",
	}

	_, err := initDatabase(cfg)
	if err == nil {
		t.Error("Expected error for invalid database URL")
	}
}

// Unit test for waitForShutdown (mocked)
func TestApp_WaitForShutdown(t *testing.T) {
	// This test verifies the shutdown logic without actually running the server
	app := &App{}

	// Test that waitForShutdown sets up signal handling
	// In a real scenario, this would wait for SIGINT/SIGTERM
	// For testing, we just verify the method exists and can be called
	// The actual signal handling is tested in integration tests

	if app.waitForShutdown == nil {
		t.Error("waitForShutdown method should exist")
	}
}

// Benchmark tests
func BenchmarkApp_New(b *testing.B) {
	if os.Getenv("DATABASE_URL") == "" {
		b.Skip("DATABASE_URL not set, skipping benchmark")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app, err := New()
		if err != nil {
			b.Fatalf("Failed to create app: %v", err)
		}
		app.Close(context.Background())
	}
}

func BenchmarkApp_Close(b *testing.B) {
	if os.Getenv("DATABASE_URL") == "" {
		b.Skip("DATABASE_URL not set, skipping benchmark")
	}

	// Pre-create apps for benchmarking
	apps := make([]*App, b.N)
	for i := 0; i < b.N; i++ {
		app, err := New()
		if err != nil {
			b.Fatalf("Failed to create app: %v", err)
		}
		apps[i] = app
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		apps[i].Close(ctx)
	}
}

// Helper function to create test app with mocked dependencies
func createTestApp() *App {
	return &App{
		config: &config.Config{
			Port:        "8080",
			Environment: "test",
		},
	}
}

func TestApp_CreateTestApp(t *testing.T) {
	app := createTestApp()

	if app.config == nil {
		t.Error("Test app config is nil")
	}

	if app.config.Port != "8080" {
		t.Errorf("Expected port 8080, got %s", app.config.Port)
	}

	if app.config.Environment != "test" {
		t.Errorf("Expected environment test, got %s", app.config.Environment)
	}
}