package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"weatherapi/internal/config"
	"weatherapi/internal/db/migration"
	"weatherapi/internal/jobs"
	"weatherapi/internal/mailer"

	"weatherapi/internal/server"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

type Container struct {
	Config              *config.Config
	DB                  *bun.DB
	WeatherService      weather_service.IWeatherService
	MailerService       mailer_service.IMailerService
	SubscriptionService subscription_service.ISubscriptionService
	JobScheduler        jobs.IJobScheduler
	Router              http.Handler
}

// BuildContainer створює всі залежності і повертає контейнер
func BuildContainer() (*Container, error) {
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("config load failed: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Init DB
	db, err := initDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("database init failed: %w", err)
	}

	// Run migrations
	mr := migration.NewRunner(db, "migrations")
	if err := mr.RunMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}

	// Init Weather API adapter

	weatherService := weather_service.NewWeatherService()

	// Init Mailer
	emailSender := &mailer.SMTPSender{}
	mailerService := mailer_service.NewMailerService(emailSender, cfg.AppBaseURL)

	// Init SubscriptionService
	subscriptionService := subscription_service.NewSubscriptionService(db)

	// Init JobScheduler
	jobScheduler := jobs.NewScheduler(subscriptionService, mailerService, weatherService)

	// Init HTTP router
	router := server.NewRouter(weatherService, subscriptionService, mailerService)

	return &Container{
		Config:              cfg,
		DB:                  db,
		WeatherService:      weatherService,
		MailerService:       mailerService,
		SubscriptionService: subscriptionService,
		JobScheduler:        jobScheduler,
		Router:              router,
	}, nil
}

// initDatabase sets up Bun with PostgreSQL
func initDatabase(cfg *config.Config) (*bun.DB, error) {
	sqlDB, err := sql.Open("pg", cfg.GetDatabaseURL())
	if err != nil {
		return nil, err
	}

	db := bun.NewDB(sqlDB, pgdialect.New())

	if cfg.IsBunDebugEnabled() {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
