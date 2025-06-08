package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"weatherapi/internal/api"
	"weatherapi/internal/config"
	"weatherapi/internal/db/migration"
	"weatherapi/internal/jobs"
	"weatherapi/internal/mailer"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// App represents the main application
type App struct {
	config *config.Config
	db     *bun.DB
	router *gin.Engine
	mailer mailer.EmailSender
}

// New creates a new application instance
func New() (*App, error) {
	app := &App{}

	// Load .env file (ignore error for Docker)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	app.config = cfg

	// Validate required configurations
	if err := app.validateConfig(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Initialize database
	dbconn, err := app.initDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	app.db = dbconn

	// Run migrations
	migrationRunner := migration.NewRunner(app.db, "migrations")
	if err := migrationRunner.RunMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize mailer
	app.mailer = &mailer.SMTPSender{}

	// Setup router
	app.router = api.SetupRouter(app.db, app.mailer)

	// Setup CORS
	app.setupCORS()

	// Setup trusted proxies
	if err := app.router.SetTrustedProxies([]string{}); err != nil {
		return nil, fmt.Errorf("failed to set trusted proxies: %w", err)
	}

	return app, nil
}

// Run starts the application
func (a *App) Run() error {
	// Start background jobs
	jobs.StartWeatherNotificationLoop(a.db, a.mailer)

	// Start server
	addr := fmt.Sprintf("0.0.0.0:%s", a.config.Port)
	log.Printf("Starting server on %s (environment: %s)", addr, a.config.Environment)

	if err := a.router.Run(addr); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// Close gracefully shuts down the application
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// validateConfig validates the application configuration
func (a *App) validateConfig() error {
	return a.config.Validate()
}

// initDatabase initializes the database connection
func (a *App) initDatabase() (*bun.DB, error) {
	sqldb, err := sql.Open("pg", a.config.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	dbconn := bun.NewDB(sqldb, pgdialect.New())

	// Setup debug mode for Bun
	if a.config.IsBunDebugEnabled() {
		dbconn.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
		))
	}

	if err := dbconn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return dbconn, nil
}

// setupCORS configures CORS for the router
func (a *App) setupCORS() {
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// In production, restrict origins
	if a.config.Environment == "production" {
		corsConfig.AllowOrigins = []string{
			"https://yourdomain.com",
			"https://www.yourdomain.com",
		}
	} else {
		corsConfig.AllowOrigins = []string{"*"}
	}

	a.router.Use(cors.New(corsConfig))
}

// GetDB returns the database connection (useful for testing)
func (a *App) GetDB() *bun.DB {
	return a.db
}

// GetRouter returns the gin router (useful for testing)
func (a *App) GetRouter() *gin.Engine {
	return a.router
}

// GetConfig returns the application configuration (useful for testing)
func (a *App) GetConfig() *config.Config {
	return a.config
}
