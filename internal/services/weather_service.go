package services

import (
	"context"
	"errors"

	"weatherapi/internal/openweatherapi"
)

var (
	ErrCityNotFound = errors.New("city not found")
)

// WeatherService defines weather service interface
type WeatherService interface {
	GetWeather(ctx context.Context, city string) (*WeatherData, error)
}

// WeatherData represents weather information
type WeatherData struct {
	Temperature float64
	Humidity    float64
	Description string
}

// weatherService implements WeatherService
type weatherService struct {
	weatherAPI WeatherAPI
}

// WeatherAPI defines weather API interface
type WeatherAPI interface {
	FetchWeather(city string) (*openweatherapi.WeatherData, error)
}

// NewWeatherService creates a new weather service
func NewWeatherService() WeatherService {
	return &weatherService{
		weatherAPI: &openWeatherAPIAdapter{}, // Adapter pattern
	}
}

// GetWeather retrieves weather data for a city
func (s *weatherService) GetWeather(ctx context.Context, city string) (*WeatherData, error) {
	data, err := s.weatherAPI.FetchWeather(city)
	if err != nil {
		if errors.Is(err, openweatherapi.ErrCityNotFound) {
			return nil, ErrCityNotFound
		}
		return nil, err
	}

	return &WeatherData{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}, nil
}

// openWeatherAPIAdapter adapts the external API to our interface
type openWeatherAPIAdapter struct{}

func (a *openWeatherAPIAdapter) FetchWeather(city string) (*openweatherapi.WeatherData, error) {
	return openweatherapi.FetchWeather(city)
}
