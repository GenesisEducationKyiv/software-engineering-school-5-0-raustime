package application

import (
	"log"
	"net/http"

	"scheduler_microservice/internal/broker"
	"scheduler_microservice/internal/config"
	"scheduler_microservice/internal/scheduler"
	"scheduler_microservice/internal/clients"
)

type App struct {
	scheduler  *scheduler.Scheduler
	config     *config.Config
	natsClient *broker.NATSClient
}

func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	httpClient := http.DefaultClient

	subClient := clients.NewSubscriptionClient(httpClient, cfg.SubscriptionURL)
	weatherClient := clients.NewWeatherHttpClient(cfg.WeatherServiceURL)

	// 🔄 Замість mailerClient — підключення до NATS
	natsClient, err := broker.NewNATSClient(cfg.NATSUrl)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}

	// 🆕 передаємо NATS publisher замість mailer
	s := scheduler.NewScheduler(subClient, natsClient, weatherClient)

	return &App{
		scheduler:  s,
		config:     cfg,
		natsClient: natsClient,
	}
}

func (a *App) Run() {
	a.scheduler.Start()
}

func (a *App) Shutdown() {
	a.scheduler.Stop()
	if a.natsClient != nil {
		a.natsClient.Close()
	}
}

func (a *App) GetConfig() *config.Config {
	return a.config
}
