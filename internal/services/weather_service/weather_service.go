package weather_service

import (
	"context"
	"errors"

	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
)

// WeatherServiceProvider defines the interface for weather service
type WeatherServiceProvider interface {
	GetWeather(ctx context.Context, city string) (contracts.WeatherData, error)
}

// WeatherAPIProvider defines weather API interface with context support
type WeatherAPIProvider interface {
	FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error)
}

// WeatherService implements WeatherServiceProvider
type WeatherService struct {
	weatherAPI WeatherAPIProvider
}

// NewWeatherService creates a new weatherService with the provided API
func NewWeatherService(api WeatherAPIProvider) WeatherService {
	return WeatherService{
		weatherAPI: api,
	}
}

// GetWeather retrieves weather data for a city
func (s WeatherService) GetWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	// Validate input
	if city == "" {
		return contracts.WeatherData{}, apierrors.ErrInvalidCity
	}

	// Pass context to API call for proper cancellation/timeout handling
	data, err := s.weatherAPI.FetchWeather(ctx, city)
	if err != nil {
		if errors.Is(err, apierrors.ErrCityNotFound) {
			return contracts.WeatherData{}, apierrors.ErrCityNotFound
		}
		return contracts.WeatherData{}, err
	}

	return data, nil
}
