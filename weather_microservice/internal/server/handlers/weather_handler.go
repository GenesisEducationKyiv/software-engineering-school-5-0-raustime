package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"weather_microservice/internal/apierrors"
	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/weather_service"
)

// WeatherHandler handles weather-related requests.
type WeatherHandler struct {
	weatherService weather_service.WeatherService
}

// NewWeatherHandler creates a new weather handler.
func NewWeatherHandler(weatherService weather_service.WeatherService) WeatherHandler {
	return WeatherHandler{
		weatherService: weatherService,
	}
}

// GetWeather handles weather requests.
func (h WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logging.FromContext(ctx)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		logger.Warn(ctx, "http:GetWeather", nil, errors.New("method not allowed"))
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "City parameter is required", http.StatusBadRequest)
		logger.Warn(ctx, "http:GetWeather", nil, errors.New("missing city parameter"))
		return
	}

	weather, err := h.weatherService.GetWeather(ctx, city)
	if err != nil {
		switch {
		case errors.Is(err, apierrors.ErrCityNotFound):
			http.Error(w, "City not found", http.StatusNotFound)
			logger.Warn(ctx, "http:GetWeather", map[string]string{"city": city}, err)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			logger.Error(ctx, "http:GetWeather", map[string]string{"city": city}, err)
		}
		return
	}

	response := contracts.WeatherData{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}

	logger.Info(ctx, "http:GetWeather", map[string]interface{}{
		"city":        city,
		"temperature": response.Temperature,
		"humidity":    response.Humidity,
		"description": response.Description,
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		logger.Error(ctx, "http:GetWeather", map[string]string{"city": city}, err)
		return
	}
}
