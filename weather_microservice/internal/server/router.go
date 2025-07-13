// weather_microservice/internal/server/router.go
package server

import (
	"net/http"

	"weather_microservice/internal/client"
	"weather_microservice/internal/server/handlers"
	"weather_microservice/internal/server/middleware"
	"weather_microservice/internal/weather_service"
)

type Router struct {
	mux                 *http.ServeMux
	weatherHandler      handlers.WeatherHandler
	subscriptionHandler handlers.SubscriptionHandler
}

func NewRouter(weatherService weather_service.WeatherService) http.Handler {
	subscriptionClient := client.NewSubscriptionClient("http://subscription_microservice:8081")

	router := &Router{
		mux:                 http.NewServeMux(),
		weatherHandler:      handlers.NewWeatherHandler(weatherService),
		subscriptionHandler: handlers.NewSubscriptionHandler(subscriptionClient),
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
	// Weather route
	r.mux.HandleFunc("GET /api/weather", r.weatherHandler.GetWeather)

	// Subscription routes
	r.mux.HandleFunc("POST /api/subscribe", r.subscriptionHandler.Subscribe)
	r.mux.HandleFunc("GET /api/confirm/{token}", r.subscriptionHandler.Confirm)
	r.mux.HandleFunc("GET /api/unsubscribe/{token}", r.subscriptionHandler.Unsubscribe)
}
