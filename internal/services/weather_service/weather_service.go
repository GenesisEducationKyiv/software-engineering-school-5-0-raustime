package weather_service

import (
	"context"
	"time"

	api_errors "weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/services/weather_service/chain"
)

// WeatherServiceProvider defines the interface for weather service.
type WeatherServiceProvider interface {
	GetWeather(ctx context.Context, city string) (contracts.WeatherData, error)
}

// WeatherService implements WeatherServiceProvider using Chain of Responsibility.
type WeatherService struct {
	weatherChain    *chain.WeatherChain
	cache           contracts.WeatherCache
	cacheExpiration time.Duration
}

// NewWeatherService creates a new weatherService with the provided chain.
func NewWeatherService(
	weatherChain *chain.WeatherChain,
	cache contracts.WeatherCache,
	cacheExpiration time.Duration,
) WeatherService {
	return WeatherService{
		weatherChain:    weatherChain,
		cache:           cache,
		cacheExpiration: cacheExpiration,
	}
}

// GetWeather retrieves weather data for a city using the chain of providers.
func (s WeatherService) GetWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	// Validate input.
	if city == "" {
		return contracts.WeatherData{}, api_errors.ErrInvalidCity
	}

	// Try getting from cache.
	if cachedData, err := s.cache.Get(ctx, city); err == nil {
		return cachedData, nil
	}

	// Use the chain to get weather data if cache miss or cache disabled.
	data, err := s.weatherChain.GetWeather(ctx, city)
	if err != nil {
		return contracts.WeatherData{}, err
	}
	// The cache implementation will handle whether caching is enabled or not.
	_ = s.cache.Set(ctx, city, data, s.cacheExpiration)

	return data, nil
}
