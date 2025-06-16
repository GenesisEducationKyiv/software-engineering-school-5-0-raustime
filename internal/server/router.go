package server

import (
	"net/http"

	"weatherapi/internal/server/handlers"
	"weatherapi/internal/server/middleware"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"
)

type Router struct {
	mux                 *http.ServeMux
	weatherHandler      handlers.WeatherHandler
	subscriptionHandler handlers.SubscriptionHandler
}

func NewRouter(
	weatherService weather_service.WeatherService,
	subscriptionService subscription_service.SubscriptionService,
	mailerService mailer_service.MailerService,
) http.Handler {
	router := &Router{
		mux:                 http.NewServeMux(),
		weatherHandler:      handlers.NewWeatherHandler(weatherService),
		subscriptionHandler: handlers.NewSubscriptionHandler(subscriptionService),
	}

	router.setupRoutes()

	return middleware.Chain(
		router.mux,
		middleware.CORS(),
		middleware.Logging(),
		middleware.Recovery(),
	)
}

func (r *Router) setupRoutes() {
	// Weather routes - GET метод явно вказаний
	r.mux.HandleFunc("GET /api/weather", r.weatherHandler.GetWeather)

	// Subscription routes з явним указанням методів
	r.mux.HandleFunc("POST /api/subscribe", r.subscriptionHandler.Subscribe)
	r.mux.HandleFunc("GET /api/confirm/{token}", r.subscriptionHandler.Confirm)
	r.mux.HandleFunc("GET /api/unsubscribe/{token}", r.subscriptionHandler.Unsubscribe)
}
