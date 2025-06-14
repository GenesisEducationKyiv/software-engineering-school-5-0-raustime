package weather_service

import (
	"context"
	"errors"

	"weatherapi/internal/openweatherapi"
)

var (
	ErrCityNotFound = errors.New("city not found")
)

type IWeatherService interface {
	GetCurrentWeather(city string) (*WeatherData, error)
}

// WeatherData represents weather information
type WeatherData struct {
	Temperature float64
	Humidity    float64
	Description string
}

// weatherService implements WeatherService
type weatherService struct {
	weatherAPI IWeatherAPI
}

// WeatherAPI defines weather API interface
type IWeatherAPI interface {
	FetchWeather(city string) (*openweatherapi.WeatherData, error)
}

// NewWeatherService creates a new weather service
func NewWeatherService() IWeatherService {
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

// GetCurrentWeather retrieves weather data for a city (implements IWeatherService)
func (s *weatherService) GetCurrentWeather(city string) (*WeatherData, error) {
	// Use context.Background() since interface does not provide context
	return s.GetWeather(context.Background(), city)
}

// openWeatherAPIAdapter adapts the external API to our interface
type openWeatherAPIAdapter struct{}

func (a *openWeatherAPIAdapter) FetchWeather(city string) (*openweatherapi.WeatherData, error) {
	return openweatherapi.FetchWeather(city)
}
