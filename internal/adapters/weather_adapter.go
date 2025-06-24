package adapters

import (
	"context"
	"errors"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
	"weatherapi/internal/extapi/weatherapi"
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
		Temperature: data.Current.TempC,
		Humidity:    data.Current.Humidity,
		Description: data.Current.Condition.Text,
	}, nil

}
