package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"weatherapi/internal/config"
	"weatherapi/internal/db/migration"
	"weatherapi/internal/jobs"

	"weatherapi/internal/adapters"
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
	WeatherService      weather_service.WeatherService
	MailerService       mailer_service.MailerService
	SubscriptionService subscription_service.SubscriptionService
	JobScheduler        jobs.Scheduler
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
	api := adapters.OpenWeatherAdapter{}
	weatherService := weather_service.NewWeatherService(api)

	// Init Mailer

	emailSender := mailer_service.NewSMTPSender(
		cfg.SMTPUser,
		cfg.SMTPPassword,
		cfg.SMTPHost,
		strconv.Itoa(cfg.SMTPPort),
	)
	mailerService := mailer_service.NewMailerService(emailSender, cfg.AppBaseURL)

	subscriptionService := subscription_service.NewSubscriptionService(db, mailerService)
	jobScheduler := jobs.NewScheduler(subscriptionService, mailerService, weatherService)
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
