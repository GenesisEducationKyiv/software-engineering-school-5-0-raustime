package adapters

import (
	"context"
	"errors"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/extapi/openweatherapi"
)

type OpenWeatherAdapter struct{}

func (a OpenWeatherAdapter) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	// Use the new context-aware function
	data, err := openweatherapi.FetchWeatherWithContext(ctx, city)
	if err != nil {
		if errors.Is(err, apierrors.ErrCityNotFound) {
			return contracts.WeatherData{}, apierrors.ErrCityNotFound
		}
		return contracts.WeatherData{}, err
	}
	return contracts.WeatherData{
		Temperature: data.Main.Temp,
		Humidity:    data.Main.Humidity,
		Description: data.Weather[0].Description,
	}, nil
}
