package server

import (
	"net/http"

	"weather_microservice/internal/client"
	"weather_microservice/internal/config"
	"weather_microservice/internal/server/handlers"
	"weather_microservice/internal/server/middleware"
	"weather_microservice/internal/weather_service"
)

type Router struct {
	mux                 *http.ServeMux
	weatherHandler      handlers.WeatherHandler
	subscriptionHandler handlers.SubscriptionHandler
}

func NewRouter(cfg *config.Config, weatherService weather_service.WeatherService) http.Handler {
	subscriptionClient := client.NewSubscriptionClient(cfg.SubscriptionServiceURL)

	router := &Router{
		mux:                 http.NewServeMux(),
		weatherHandler:      handlers.NewWeatherHandler(weatherService),
		subscriptionHandler: handlers.NewSubscriptionHandler(subscriptionClient),
	}

	router.setupRoutes()

	return middleware.Chain(
		router.mux,
		middleware.Recovery(), // ловить panic найглибше
		middleware.Trace(),    // додає trace_id у ctx
		middleware.Logging(),  // логування запитів
		middleware.CORS(),     // заголовки після логіки
	)
}

func (r *Router) setupRoutes() {
	r.mux.HandleFunc("GET /api/weather", r.weatherHandler.GetWeather)

	r.mux.HandleFunc("POST /api/subscribe", r.subscriptionHandler.Subscribe)
	r.mux.HandleFunc("GET /api/confirm/{token}", r.subscriptionHandler.Confirm)
	r.mux.HandleFunc("GET /api/unsubscribe/{token}", r.subscriptionHandler.Unsubscribe)
}
