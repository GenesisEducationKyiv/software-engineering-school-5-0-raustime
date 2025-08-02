package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var client = http.Client{Timeout: 5 * time.Second}

func TestWeatherService_DefaultProvider_HTTP(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/weather?city=Kyiv")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWeatherService_DefaultProvider_InvalidCity(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/weather?city=InvalidCity")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestWeatherService_MissingCityParam(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/weather")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// 🔁 Нові тести з префіксами

func TestWeatherService_OpenWeatherProvider_Valid(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/weather?city=openweather-Kyiv")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWeatherService_WeatherAPIProvider_Valid(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/weather?city=weatherapi-Kyiv")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWeatherService_OpenWeatherProvider_Invalid(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/weather?city=openweather-InvalidCity")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestWeatherService_WeatherAPIProvider_Invalid(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/weather?city=weatherapi-InvalidCity")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
