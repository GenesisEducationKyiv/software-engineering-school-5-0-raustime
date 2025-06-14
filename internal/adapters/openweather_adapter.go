package adapters

import (
	"errors"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/openweatherapi"
)

type OpenWeatherAdapter struct{}

func (a *OpenWeatherAdapter) FetchWeather(city string) (*contracts.WeatherData, error) {
	data, err := openweatherapi.FetchWeather(city)
	if err != nil {
		if errors.Is(err, openweatherapi.ErrCityNotFound) {
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
