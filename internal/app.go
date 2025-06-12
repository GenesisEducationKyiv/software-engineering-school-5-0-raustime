package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weatherapi/internal/config"
	"weatherapi/internal/db/migration"
	"weatherapi/internal/jobs"
	"weatherapi/internal/mailer"
	"weatherapi/internal/server"
	"weatherapi/internal/services"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type App struct {
	config              *config.Config
	db                  *bun.DB
	httpServer          *http.Server
	weatherService      services.WeatherService
	subscriptionService services.SubscriptionService
	mailerService       services.MailerService
	jobScheduler        *jobs.Scheduler
}

// New створює новий екземпляр додатку
func New() (*App, error) {
	// Load .env file (ignore error for Docker)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	migrationRunner := migration.NewRunner(db, "migrations")
	if err := migrationRunner.RunMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize services
	emailSender := &mailer.SMTPSender{}
	mailerService := services.NewMailerService(emailSender)
	weatherService := services.NewWeatherService()
	subscriptionService := services.NewSubscriptionService(db)
	
	// Initialize job scheduler
	jobScheduler := jobs.NewScheduler(db, mailerService, weatherService, subscriptionService)

	return &App{
		config:              cfg,
		db:                  db,
		weatherService:      weatherService,
		subscriptionService: subscriptionService,
		mailerService:       mailerService,
		jobScheduler:        jobScheduler,
	}, nil
}
// Run starts the application with graceful shutdown
func (a *App) Run() error {
	// Start background jobs
	a.jobScheduler.Start()

	// Setup HTTP server
	router := server.NewRouter(a.weatherService, a.subscriptionService, a.mailerService)
	a.httpServer = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", a.config.Port),
		Handler: router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s (environment: %s)", a.httpServer.Addr, a.config.Environment)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	return a.waitForShutdown()
}

// waitForShutdown waits for interrupt signal and performs graceful shutdown
func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.Close(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}

// Close gracefully shuts down the application
func (a *App) Close(ctx context.Context) error {
	var err error

	// Stop job scheduler
	if a.jobScheduler != nil {
		a.jobScheduler.Stop()
	}

	// Shutdown HTTP server
	if a.httpServer != nil {
		if shutdownErr := a.httpServer.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("http server shutdown error: %w", shutdownErr)
		}
	}

	// Close database connection
	if a.db != nil {
		if dbErr := a.db.Close(); dbErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; db close error: %v", err, dbErr)
			} else {
				err = fmt.Errorf("db close error: %w", dbErr)
			}
		}
	}

	return err
}

// initDatabase initializes the database connection
func initDatabase(cfg *config.Config) (*bun.DB, error) {
	sqldb, err := sql.Open("pg", cfg.GetDatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}

	dbconn := bun.NewDB(sqldb, pgdialect.New())

	// Setup debug mode for Bun
	if cfg.IsBunDebugEnabled() {
		dbconn.AddQueryHook(bundebug.NewQueryHook(
			bundebug.WithVerbose(true),
		))
	}

	if err := dbconn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}

	return dbconn, nil
}

// GetDB returns the database connection 
func (a *App) GetDB() *bun.DB {
	return a.db
}

// GetConfig returns the application configuration 
func (a *App) GetConfig() *config.Config {
	return a.config
}