package weather_service

import (
	"context"

	api_errors "weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/services/weather_service/chain"
)

// WeatherServiceProvider defines the interface for weather service
type WeatherServiceProvider interface {
	GetWeather(ctx context.Context, city string) (contracts.WeatherData, error)
}

// WeatherService implements WeatherServiceProvider using Chain of Responsibility
type WeatherService struct {
	weatherChain *chain.WeatherChain
}

// NewWeatherService creates a new weatherService with the provided chain
func NewWeatherService(weatherChain *chain.WeatherChain) WeatherService {
	return WeatherService{
		weatherChain: weatherChain,
	}
}

// GetWeather retrieves weather data for a city using the chain of providers
func (s WeatherService) GetWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	// Validate input
	if city == "" {
		return contracts.WeatherData{}, api_errors.ErrInvalidCity
	}

	// Use the chain to get weather data
	data, err := s.weatherChain.GetWeather(ctx, city)
	if err != nil {
		return contracts.WeatherData{}, err
	}

	return data, nil
}
