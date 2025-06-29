package di

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	"weatherapi/internal/adapters"
	"weatherapi/internal/config"
	"weatherapi/internal/db/migration"
	"weatherapi/internal/db/repositories"
	"weatherapi/internal/jobs"
	"weatherapi/internal/logging"
	"weatherapi/internal/server"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"
	"weatherapi/internal/services/weather_service/chain"
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
func BuildContainer() (Container, error) {
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return Container{}, fmt.Errorf("config load failed: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Container{}, fmt.Errorf("config validation failed: %w", err)
	}

	// Init DB
	db, err := initDatabase(cfg)
	if err != nil {
		return Container{}, fmt.Errorf("database init failed: %w", err)
	}

	// Run migrations
	mr := migration.NewRunner(db, "migrations")
	if err := mr.RunMigrations(context.Background()); err != nil {
		return Container{}, fmt.Errorf("migrations failed: %w", err)
	}

	subscriptionRepo := repositories.NewSubscriptionRepo(db)

	// Init Weather API adapters with config
	openWeatherAdapter, err := adapters.NewOpenWeatherAdapter(cfg.OpenWeatherKey)
	if err != nil {
		return Container{}, fmt.Errorf("failed to create adapter: %w", err)
	}

	weatherAPIAdapter, err := adapters.NewWeatherAPIAdapter(cfg.WeatherKey)
	if err != nil {
		return Container{}, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Create logger
	logger := logging.NewFileWeatherLogger("weather_providers.log")

	openWeatherHandler := chain.NewBaseWeatherHandler(&openWeatherAdapter, "openweathermap.org")
	weatherAPIHandler := chain.NewBaseWeatherHandler(&weatherAPIAdapter, "weatherapi.com")

	// Set up the chain: OpenWeather -> WeatherAPI
	openWeatherHandler.SetNext(weatherAPIHandler)

	weatherChain := chain.NewWeatherChain(logger)
	weatherChain.SetFirstHandler(openWeatherHandler)

	weatherService := weather_service.NewWeatherService(weatherChain)
	mailerService := mailer_service.NewMailerService(mailer_service.NewSMTPSender(cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost, strconv.Itoa(cfg.SMTPPort)), cfg.AppBaseURL)

	subscriptionService := subscription_service.New(subscriptionRepo, mailerService)
	jobScheduler := jobs.NewScheduler(subscriptionService, mailerService, weatherService)
	router := server.NewRouter(weatherService, subscriptionService, mailerService)

	return Container{
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
