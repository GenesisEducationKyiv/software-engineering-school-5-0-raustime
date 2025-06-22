package weather_service

import (
	"context"

	api_errors "weatherapi/internal/apierrors"
	"weatherapi/internal/chain"
	"weatherapi/internal/contracts"
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

// Example of how to set up the chain in main.go or dependency injection
/*
package main

import (
	"context"
	"log"
	"weatherapi/internal/adapters"
	"weatherapi/internal/chain"
	"weatherapi/internal/weather_service"
)

func main() {
	// Create adapters
	openWeatherAdapter := &adapters.OpenWeatherAdapter{}
	weatherAPIAdapter := &adapters.WeatherAPIAdapter{}

	// Create handlers
	openWeatherHandler := chain.NewBaseWeatherHandler(openWeatherAdapter, "openweathermap.org")
	weatherAPIHandler := chain.NewBaseWeatherHandler(weatherAPIAdapter, "weatherapi.com")

	// Set up the chain: OpenWeather -> WeatherAPI
	openWeatherHandler.SetNext(weatherAPIHandler)

	// Create and configure the chain
	weatherChain := chain.NewWeatherChain()
	weatherChain.SetFirstHandler(openWeatherHandler)

	// Create weather service
	weatherService := weather_service.NewWeatherService(weatherChain)

	// Use the service
	ctx := context.Background()
	data, err := weatherService.GetWeather(ctx, "London")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("Weather data: %+v", data)
}
*/
