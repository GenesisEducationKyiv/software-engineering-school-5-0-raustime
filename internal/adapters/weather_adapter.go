package adapters

import (
	"context"
	"errors"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/weatherapi"
)

type WeatherAPIAdapter struct{}

func (a WeatherAPIAdapter) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	data, err := weatherapi.FetchWeatherWithContext(ctx, city)
	if err != nil {
		if errors.Is(err, apierrors.ErrCityNotFound) {
			return contracts.WeatherData{}, apierrors.ErrCityNotFound
		}
		return contracts.WeatherData{}, err
	}
	return contracts.WeatherData{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}, nil
}
