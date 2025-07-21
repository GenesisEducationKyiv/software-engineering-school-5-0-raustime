package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"weather_microservice/internal/apierrors"
	"weather_microservice/internal/contracts"
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
func (h WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) { // value receiver
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "City parameter is required", http.StatusBadRequest)
		return
	}

	weather, err := h.weatherService.GetWeather(r.Context(), city)
	if err != nil {
		switch {
		case errors.Is(err, apierrors.ErrCityNotFound):
			http.Error(w, "City not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	response := contracts.WeatherData{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
