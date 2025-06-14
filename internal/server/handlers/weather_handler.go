package handlers

import (
	"encoding/json"
	"net/http"

	"weatherapi/internal/services/weather_service"
)

// WeatherHandler handles weather-related requests
type WeatherHandler struct {
	weatherService weather_service.IWeatherService
}

// NewWeatherHandler creates a new weather handler
func NewWeatherHandler(weatherService weather_service.IWeatherService) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
	}
}

// WeatherResponse represents weather API response
type WeatherResponse struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

// GetWeather handles weather requests
func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "City parameter is required", http.StatusBadRequest)
		return
	}

	weather, err := h.weatherService.GetCurrentWeather(city)
	if err != nil {
		switch err {
		case weather_service.ErrCityNotFound:
			http.Error(w, "City not found", http.StatusNotFound)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := WeatherResponse{
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
