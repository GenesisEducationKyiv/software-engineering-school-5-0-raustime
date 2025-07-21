package client

import (
	"net/http"
	"weather_microservice/gen/go/weather/v1/weatherv1connect"
)

func NewWeatherClient(baseURL string) weatherv1connect.WeatherServiceClient {
	return weatherv1connect.NewWeatherServiceClient(
		http.DefaultClient,
		baseURL,
	)
}
