package application

import (
	"log"
	"net/http"
	"scheduler_microservice/internal/clients"
	"scheduler_microservice/internal/config"
	"scheduler_microservice/internal/scheduler"
)

type App struct {
	scheduler *scheduler.Scheduler
	config    *config.Config
}

func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	httpClient := http.DefaultClient

	subClient := clients.NewSubscriptionClient(httpClient, cfg.SubscriptionURL)
	mailerClient := clients.NewMailerClient(cfg.MailerServiceURL)
	weatherClient := clients.NewWeatherHttpClient(cfg.WeatherServiceURL)

	s := scheduler.NewScheduler(subClient, mailerClient, weatherClient)
	return &App{
		scheduler: s,
		config:    cfg,
	}
}

func (a *App) Run() {
	a.scheduler.Start()
}

func (a *App) Shutdown() {
	a.scheduler.Stop()
}

func (a *App) GetConfig() *config.Config {
	return a.config
}
