package server

import (
	"net/http"

	"weatherapi/internal/server/handlers"
	"weatherapi/internal/server/middleware"
	"weatherapi/internal/services/mailer_service"
	"weatherapi/internal/services/subscription_service"
	"weatherapi/internal/services/weather_service"
)

// Router represents HTTP router following Single Responsibility Principle
type Router struct {
	mux                 *http.ServeMux
	weatherHandler      *handlers.WeatherHandler
	subscriptionHandler *handlers.SubscriptionHandler
}

// NewRouter creates a new router with all handlers
func NewRouter(weatherService weather_service.IWeatherService, subscriptionService subscription_service.ISubscriptionService, mailerService mailer_service.IMailerService) http.Handler {
	router := &Router{
		mux:                 http.NewServeMux(),
		weatherHandler:      handlers.NewWeatherHandler(weatherService),
		subscriptionHandler: handlers.NewSubscriptionHandler(subscriptionService, mailerService),
	}

	router.setupRoutes()

	// Apply middleware
	return middleware.Chain(
		router.mux,
		middleware.CORS(),
		middleware.Logging(),
		middleware.Recovery(),
	)
}

// setupRoutes configures all application routes
func (r *Router) setupRoutes() {
	// Weather routes
	r.mux.HandleFunc("/api/weather", r.weatherHandler.GetWeather)

	// Subscription routes
	r.mux.HandleFunc("/api/subscribe", r.methodFilter("POST", r.subscriptionHandler.Subscribe))
	r.mux.HandleFunc("/api/confirm/", r.methodFilter("GET", r.subscriptionHandler.Confirm))
	r.mux.HandleFunc("/api/unsubscribe/", r.methodFilter("GET", r.subscriptionHandler.Unsubscribe))
}

// methodFilter filters HTTP methods for handlers
func (r *Router) methodFilter(allowedMethod string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != allowedMethod {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, req)
	}
}
