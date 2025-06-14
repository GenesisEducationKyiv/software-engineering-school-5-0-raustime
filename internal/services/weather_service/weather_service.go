package weather_service

import (
	"context"
	"errors"

	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
)

type IWeatherService interface {
	GetCurrentWeather(city string) (*contracts.WeatherData, error)
}

// weatherService implements IWeatherService
type weatherService struct {
	weatherAPI IWeatherAPI
}

// WeatherAPI defines weather API interface
type IWeatherAPI interface {
	FetchWeather(city string) (*contracts.WeatherData, error)
}

// NewWeatherService створює новий weatherService з переданим API
func NewWeatherService(api IWeatherAPI) IWeatherService {
	return &weatherService{
		weatherAPI: api,
	}
}

// GetWeather retrieves weather data for a city
func (s *weatherService) GetWeather(ctx context.Context, city string) (*contracts.WeatherData, error) {
	data, err := s.weatherAPI.FetchWeather(city)
	if err != nil {
		if errors.Is(err, apierrors.ErrCityNotFound) {
			return nil, apierrors.ErrCityNotFound
		}
		return nil, err
	}

	return &contracts.WeatherData{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}, nil
}

// GetCurrentWeather retrieves weather data for a city (implements IWeatherService)
func (s *weatherService) GetCurrentWeather(city string) (*contracts.WeatherData, error) {
	return s.GetWeather(context.Background(), city)
}
