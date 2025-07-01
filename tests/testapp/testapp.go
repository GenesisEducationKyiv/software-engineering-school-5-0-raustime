package testapp

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"weatherapi/internal/adapters"
	"weatherapi/internal/cache"
	"weatherapi/internal/config"
	"weatherapi/internal/db/migration"
	"weatherapi/internal/db/repositories"
	"weatherapi/internal/logging"
	"weatherapi/internal/server"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"
	"weatherapi/internal/services/weather_service/chain"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type TestContainer struct {
	Config              *config.Config
	DB                  *bun.DB
	WeatherService      weather_service.WeatherServiceProvider
	MailerService       mailer_service.MailerService
	SubscriptionService subscription_service.SubscriptionService
	Router              http.Handler
}

func Initialize() *TestContainer {
	_ = godotenv.Load(".env.test")

	cfg, err := config.LoadTestConfig()

	if err != nil {
		log.Fatalf("❌ Failed to load config for test: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Invalid test config: %v", err)
	}

	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to init test DB: %v", err)
	}

	// Skip migrations if SKIP_MIGRATIONS is set.
	if os.Getenv("SKIP_MIGRATIONS") == "true" {
		log.Println("⚠️ Skipping migrations as requested")
	} else {
		// Run test migrations from resolved absolute path.
		migrationsDir := resolveMigrationsPath()
		log.Printf("📁 Using migrations from: %s", migrationsDir)

		mr := migration.NewRunner(db, migrationsDir)
		if err := mr.RunMigrations(context.Background()); err != nil {
			log.Fatalf("❌ Failed to run test migrations: %v", err)
		}
	}

	subscriptionRepo := repositories.NewSubscriptionRepo(db)

	// Setup weather service with chain of responsibility.
	weatherService, err := setupWeatherService(cfg)

	if err != nil {
		log.Fatalf("❌ Failed to setup weather service: %v", err)
	}

	mockSender := mailer_service.NewMockSender()
	mailerService := mailer_service.NewMailerService(mockSender, cfg.AppBaseURL)

	subscriptionService := subscription_service.New(subscriptionRepo, mailerService)
	router := server.NewRouter(weatherService.(weather_service.WeatherService), subscriptionService, mailerService)

	return &TestContainer{
		Config:              cfg,
		DB:                  db,
		WeatherService:      weatherService,
		MailerService:       mailerService,
		SubscriptionService: subscriptionService,
		Router:              router,
	}
}

// setupWeatherService creates a weather service with chain of responsibility.
func setupWeatherService(cfg *config.Config) (weather_service.WeatherServiceProvider, error) {
	// Create logger for tests
	logger := logging.NewFileWeatherLogger("test_weather_providers.log")

	// Create adapters with config.
	openWeatherAdapter, err := adapters.NewOpenWeatherAdapter(cfg.OpenWeatherKey)
	if err != nil {
		var emptyProvider weather_service.WeatherServiceProvider
		return emptyProvider, fmt.Errorf("failed to create adapter: %w", err)
	}
	weatherAPIAdapter, err := adapters.NewWeatherAPIAdapter(cfg.WeatherKey)
	if err != nil {
		var emptyProvider weather_service.WeatherServiceProvider
		return emptyProvider, fmt.Errorf("failed to create adapter: %w", err)
	}
	// Create handlers for the chain with logger.
	openWeatherHandler := chain.NewBaseWeatherHandler(&openWeatherAdapter, "openweathermap.org")
	weatherAPIHandler := chain.NewBaseWeatherHandler(&weatherAPIAdapter, "weatherapi.com")

	// Set up the chain: OpenWeather -> WeatherAPI.
	// In tests, we might want to use a simpler chain or mock.
	openWeatherHandler.SetNext(weatherAPIHandler)

	// Create and configure the chain
	weatherChain := chain.NewWeatherChain(logger)
	weatherChain.SetFirstHandler(openWeatherHandler)

	cache := cache.NoopWeatherCache{} // Use a no-op cache.
	cacheDuration := 5 * 60           // 5 minutes in seconds.

	return weather_service.NewWeatherService(weatherChain, cache, time.Duration(cacheDuration)*time.Second), nil
}

// Alternative setup for tests that need more control.
func setupWeatherServiceForTests(cfg *config.Config, useOnlyPrimary bool) (weather_service.WeatherServiceProvider, error) {
	// Create logger for tests.
	logger := logging.NewFileWeatherLogger("test_weather_providers.log")

	// Create adapters with config.
	openWeatherAdapter, err := adapters.NewOpenWeatherAdapter(cfg.OpenWeatherKey)
	if err != nil {
		var emptyProvider weather_service.WeatherServiceProvider
		return emptyProvider, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Create handler for the chain with logger.
	openWeatherHandler := chain.NewBaseWeatherHandler(&openWeatherAdapter, "openweathermap.org")

	if !useOnlyPrimary {
		// Add secondary provider for full chain.
		weatherAPIAdapter, err := adapters.NewWeatherAPIAdapter(cfg.WeatherKey)
		if err != nil {
			var emptyProvider weather_service.WeatherServiceProvider
			return emptyProvider, fmt.Errorf("failed to create adapter: %w", err)
		}
		weatherAPIHandler := chain.NewBaseWeatherHandler(&weatherAPIAdapter, "weatherapi.com")
		openWeatherHandler.SetNext(weatherAPIHandler)
	}

	// Create and configure the chain.
	weatherChain := chain.NewWeatherChain(logger)
	weatherChain.SetFirstHandler(openWeatherHandler)

	// Create weather service with the chain.
	cache := cache.NoopWeatherCache{} // Use a no-op cache.
	cacheDuration := 5 * 60           // 5 minutes in seconds.
	return weather_service.NewWeatherService(weatherChain, cache, time.Duration(cacheDuration)*time.Second), nil
}

// InitializeWithSingleProvider creates a test container with only one weather provider.
// Useful for tests that need to control which provider is used.
func InitializeWithSingleProvider() *TestContainer {
	_ = godotenv.Load(".env.test")

	cfg, err := config.LoadTestConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config for test: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Invalid test config: %v", err)
	}

	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to init test DB: %v", err)
	}

	// Skip migrations if SKIP_MIGRATIONS is set.
	if os.Getenv("SKIP_MIGRATIONS") == "true" {
		log.Println("⚠️ Skipping migrations as requested")
	} else {
		migrationsDir := resolveMigrationsPath()
		log.Printf("📁 Using migrations from: %s", migrationsDir)

		mr := migration.NewRunner(db, migrationsDir)
		if err := mr.RunMigrations(context.Background()); err != nil {
			log.Fatalf("❌ Failed to run test migrations: %v", err)
		}
	}

	subscriptionRepo := repositories.NewSubscriptionRepo(db)

	// Setup weather service with single provider for testing.
	weatherService, err := setupWeatherServiceForTests(cfg, true)
	if err != nil {
		log.Fatalf("❌ Failed to setup weather service: %v", err)
	}

	mockSender := mailer_service.NewMockSender()
	mailerService := mailer_service.NewMailerService(mockSender, cfg.AppBaseURL)

	subscriptionService := subscription_service.New(subscriptionRepo, mailerService)
	router := server.NewRouter(weatherService.(weather_service.WeatherService), subscriptionService, mailerService)

	return &TestContainer{
		Config:              cfg,
		DB:                  db,
		WeatherService:      weatherService,
		MailerService:       mailerService,
		SubscriptionService: subscriptionService,
		Router:              router,
	}
}

// InitializeWithMockLogger creates a test container with mock logger for testing.
func InitializeWithMockLogger() *TestContainer {
	_ = godotenv.Load(".env.test")

	cfg, err := config.LoadTestConfig()
	if err != nil {
		log.Fatalf("❌ Failed to load config for test: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("❌ Invalid test config: %v", err)
	}

	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to init test DB: %v", err)
	}

	// Skip migrations if SKIP_MIGRATIONS is set.
	if os.Getenv("SKIP_MIGRATIONS") == "true" {
		log.Println("⚠️ Skipping migrations as requested")
	} else {
		migrationsDir := resolveMigrationsPath()
		log.Printf("📁 Using migrations from: %s", migrationsDir)

		mr := migration.NewRunner(db, migrationsDir)
		if err := mr.RunMigrations(context.Background()); err != nil {
			log.Fatalf("❌ Failed to run test migrations: %v", err)
		}
	}

	subscriptionRepo := repositories.NewSubscriptionRepo(db)

	// Setup weather service with mock logger.
	weatherService, err := setupWeatherServiceWithMockLogger(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to setup weather service: %v", err)
	}

	mockSender := mailer_service.NewMockSender()
	mailerService := mailer_service.NewMailerService(mockSender, cfg.AppBaseURL)

	subscriptionService := subscription_service.New(subscriptionRepo, mailerService)
	router := server.NewRouter(weatherService.(weather_service.WeatherService), subscriptionService, mailerService)

	return &TestContainer{
		Config:              cfg,
		DB:                  db,
		WeatherService:      weatherService,
		MailerService:       mailerService,
		SubscriptionService: subscriptionService,
		Router:              router,
	}
}

// setupWeatherServiceWithMockLogger creates weather service with mock logger for testing.
func setupWeatherServiceWithMockLogger(cfg *config.Config) (weather_service.WeatherServiceProvider, error) {
	// Create mock logger for testing.
	mockLogger := logging.NewMockLogger()

	// Create adapters with config.
	openWeatherAdapter, err := adapters.NewOpenWeatherAdapter(cfg.OpenWeatherKey)
	if err != nil {
		var emptyProvider weather_service.WeatherServiceProvider
		return emptyProvider, fmt.Errorf("failed to create OpenWeather adapter: %w", err)
	}
	weatherAPIAdapter, err := adapters.NewWeatherAPIAdapter(cfg.WeatherKey)
	if err != nil {
		var emptyProvider weather_service.WeatherServiceProvider
		return emptyProvider, fmt.Errorf("failed to create WeatherAPI adapter: %w", err)
	}
	// Create handlers for the chain with mock logger.
	openWeatherHandler := chain.NewBaseWeatherHandler(&openWeatherAdapter, "openweathermap.org")
	weatherAPIHandler := chain.NewBaseWeatherHandler(&weatherAPIAdapter, "weatherapi.com")

	// Set up the chain: OpenWeather -> WeatherAPI.
	openWeatherHandler.SetNext(weatherAPIHandler)

	// Create and configure the chain.
	weatherChain := chain.NewWeatherChain(mockLogger)
	weatherChain.SetFirstHandler(openWeatherHandler)

	// Create weather service with the chain.
	cache := cache.NoopWeatherCache{} // Use a no-op cache for tests.
	cacheDuration := 5 * 60           // 5 minutes in seconds.
	return weather_service.NewWeatherService(weatherChain, cache, time.Duration(cacheDuration)*time.Second), nil
}

func initDatabase(cfg *config.Config) (*bun.DB, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(cfg.GetDatabaseURL())))
	log.Println("💡 DB_URL =", cfg.GetDatabaseURL())
	log.Println(cfg.Environment)
	db := bun.NewDB(sqlDB, pgdialect.New())

	if cfg.IsBunDebugEnabled() {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// resolveMigrationsPath detects the full path to the `migrations` directory.
func resolveMigrationsPath() string {
	log.Printf("🔍 Starting migrations path resolution...")

	// First, try environment variable (useful for Docker).
	if migrationsPath := os.Getenv("MIGRATIONS_PATH"); migrationsPath != "" {
		log.Printf("🔍 Trying env MIGRATIONS_PATH: %s", migrationsPath)
		if _, err := os.Stat(migrationsPath); err == nil {
			log.Printf("📁 Using migrations path from env: %s", migrationsPath)
			return migrationsPath
		} else {
			log.Printf("❌ Env path not found: %v", err)
		}
	}

	// Get current working directory.
	workDir, err := os.Getwd()
	if err != nil {
		log.Printf("❌ Failed to get working directory: %v", err)
	} else {
		log.Printf("🔍 Current working directory: %s", workDir)
	}

	// Get runtime path info.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("❌ Unable to get current filename to resolve migrations path")
	}
	log.Printf("🔍 Current file: %s", filename)
	baseDir := filepath.Dir(filename)
	log.Printf("🔍 Base directory: %s", baseDir)

	// Try different possible paths with detailed logging.
	possiblePaths := []string{
		"./migrations",                       // Relative to working directory.
		"migrations",                         // Simple relative path.
		filepath.Join(workDir, "migrations"), // Working dir + migrations.
		filepath.Join(baseDir, "..", "..", "migrations"), // From tests/testapp -> project root.
		filepath.Join(baseDir, "..", "migrations"),       // From tests -> project root.
		filepath.Join(baseDir, "migrations"),             // Same directory as testapp.
	}

	for i, path := range possiblePaths {
		log.Printf("🔍 Trying path %d: %s", i+1, path)

		absPath, err := filepath.Abs(path)
		if err != nil {
			log.Printf("❌ Failed to get absolute path for %s: %v", path, err)
			continue
		}
		log.Printf("🔍 Absolute path: %s", absPath)

		if stat, err := os.Stat(absPath); err == nil {
			if stat.IsDir() {
				log.Printf("📁 Found migrations directory: %s", absPath)

				// List contents to verify it's the right directory.
				files, err := os.ReadDir(absPath)
				if err == nil {
					log.Printf("📁 Directory contains %d items:", len(files))
					for j, file := range files {
						if j < 5 { // Only show first 5 files.
							log.Printf("   - %s", file.Name())
						}
					}
				}

				return absPath
			} else {
				log.Printf("❌ Path exists but is not a directory: %s", absPath)
			}
		} else {
			log.Printf("❌ Path not found: %s (%v)", absPath, err)
		}
	}

	// As a last resort, let's see what's in the current directory.
	log.Printf("🔍 Listing current directory contents:")
	if files, err := os.ReadDir("."); err == nil {
		for _, file := range files {
			log.Printf("   - %s (dir: %t)", file.Name(), file.IsDir())
		}
	}

	log.Fatal("❌ Migrations directory not found in any expected location")
	return ""
}
