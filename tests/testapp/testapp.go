package testapp

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"weatherapi/internal/adapters"
	"weatherapi/internal/config"
	"weatherapi/internal/db/migration"
	"weatherapi/internal/db/repositories"
	"weatherapi/internal/server"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"

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
	SubscriptionService *subscription_service.SubscriptionService
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

	// Skip migrations if SKIP_MIGRATIONS is set
	if os.Getenv("SKIP_MIGRATIONS") == "true" {
		log.Println("⚠️ Skipping migrations as requested")
	} else {
		// Run test migrations from resolved absolute path
		migrationsDir := resolveMigrationsPath()
		log.Printf("📁 Using migrations from: %s", migrationsDir)

		mr := migration.NewRunner(db, migrationsDir)
		if err := mr.RunMigrations(context.Background()); err != nil {
			log.Fatalf("❌ Failed to run test migrations: %v", err)
		}
	}

	subscriptionRepo := repositories.NewSubscriptionRepo(db)

	// App dependencies
	api := adapters.OpenWeatherAdapter{}
	weatherService := weather_service.NewWeatherService(api)

	mockSender := mailer_service.NewMockSender()
	mailerService := mailer_service.NewMailerService(mockSender, cfg.AppBaseURL)

	subscriptionService := subscription_service.New(subscriptionRepo, mailerService)
	router := server.NewRouter(weatherService, subscriptionService, mailerService)

	return &TestContainer{
		Config:              cfg,
		DB:                  db,
		WeatherService:      weatherService,
		MailerService:       mailerService,
		SubscriptionService: subscriptionService,
		Router:              router,
	}
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

// resolveMigrationsPath detects the full path to the `migrations` directory
func resolveMigrationsPath() string {
	log.Printf("🔍 Starting migrations path resolution...")

	// First, try environment variable (useful for Docker)
	if migrationsPath := os.Getenv("MIGRATIONS_PATH"); migrationsPath != "" {
		log.Printf("🔍 Trying env MIGRATIONS_PATH: %s", migrationsPath)
		if _, err := os.Stat(migrationsPath); err == nil {
			log.Printf("📁 Using migrations path from env: %s", migrationsPath)
			return migrationsPath
		} else {
			log.Printf("❌ Env path not found: %v", err)
		}
	}

	// Get current working directory
	workDir, err := os.Getwd()
	if err != nil {
		log.Printf("❌ Failed to get working directory: %v", err)
	} else {
		log.Printf("🔍 Current working directory: %s", workDir)
	}

	// Get runtime path info
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("❌ Unable to get current filename to resolve migrations path")
	}
	log.Printf("🔍 Current file: %s", filename)
	baseDir := filepath.Dir(filename)
	log.Printf("🔍 Base directory: %s", baseDir)

	// Try different possible paths with detailed logging
	possiblePaths := []string{
		"./migrations",                       // Relative to working directory
		"migrations",                         // Simple relative path
		filepath.Join(workDir, "migrations"), // Working dir + migrations
		filepath.Join(baseDir, "..", "..", "migrations"), // From tests/testapp -> project root
		filepath.Join(baseDir, "..", "migrations"),       // From tests -> project root
		filepath.Join(baseDir, "migrations"),             // Same directory as testapp
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

				// List contents to verify it's the right directory
				files, err := os.ReadDir(absPath)
				if err == nil {
					log.Printf("📁 Directory contains %d items:", len(files))
					for j, file := range files {
						if j < 5 { // Only show first 5 files
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

	// As a last resort, let's see what's in the current directory
	log.Printf("🔍 Listing current directory contents:")
	if files, err := os.ReadDir("."); err == nil {
		for _, file := range files {
			log.Printf("   - %s (dir: %t)", file.Name(), file.IsDir())
		}
	}

	log.Fatal("❌ Migrations directory not found in any expected location")
	return ""
}
