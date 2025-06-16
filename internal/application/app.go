package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weatherapi/internal/config"
	"weatherapi/internal/di"
	"weatherapi/internal/jobs"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"

	"github.com/uptrace/bun"
)

type App struct {
	config              *config.Config
	db                  *bun.DB
	httpServer          *http.Server
	weatherService      weather_service.WeatherService
	subscriptionService subscription_service.SubscriptionService
	mailerService       mailer_service.MailerService
	jobScheduler        jobs.Scheduler
}

// New створює новий екземпляр додатку
func New() (*App, error) {
	container, err := di.BuildContainer()
	if err != nil {
		return nil, fmt.Errorf("failed to build container: %w", err)
	}

	app := &App{
		config:              container.Config,
		db:                  container.DB,
		weatherService:      container.WeatherService,
		subscriptionService: container.SubscriptionService,
		mailerService:       container.MailerService,
		jobScheduler:        container.JobScheduler,
		httpServer: &http.Server{
			Addr:         ":" + container.Config.Port,
			Handler:      http.HandlerFunc(container.Router.ServeHTTP),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	return app, nil
}

// Run запускає додаток з graceful shutdown
func (a *App) Run() error {
	a.jobScheduler.Start()

	go func() {
		log.Printf("Starting server on %s (env: %s)", a.httpServer.Addr, a.config.Environment)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server stopped with error: %v", err) // краще, ніж log.Fatal
		}
	}()

	return a.waitForShutdown()
}

// waitForShutdown очікує сигнал завершення і завершує роботу
func (a *App) waitForShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.Close(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	log.Println("Server exited gracefully")
	return nil
}

// Close — м'яке завершення роботи сервісів, БД, HTTP
func (a *App) Close(ctx context.Context) error {
	var err error

	a.jobScheduler.Stop()

	if a.httpServer != nil {
		if shutdownErr := a.httpServer.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("http shutdown error: %w", shutdownErr)
		}
	}

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

// GetDB повертає з'єднання з базою
func (a *App) GetDB() *bun.DB {
	return a.db
}

// GetConfig повертає конфігурацію
func (a *App) GetConfig() *config.Config {
	return a.config
}
