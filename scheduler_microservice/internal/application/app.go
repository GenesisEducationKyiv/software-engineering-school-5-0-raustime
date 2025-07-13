package application

import (
	"net/http"
	"scheduler_microservice/internal/clients"
	"scheduler_microservice/internal/scheduler"
)

type App struct {
	scheduler *scheduler.Scheduler
}

func NewApp() *App {
	httpClient := http.DefaultClient

	subClient := clients.NewSubscriptionClient(httpClient, "http://subscription-service:8090")
	mailerClient := clients.NewMailerClient(httpClient, "http://mailer-service:8091")
	weatherClient := clients.NewWeatherHttpClient("http://weather-service:8080")

	s := scheduler.NewScheduler(subClient, mailerClient, weatherClient)
	return &App{scheduler: s}
}

func (a *App) Run() {
	a.scheduler.Start()
}

func (a *App) Shutdown() {
	a.scheduler.Stop()
}
