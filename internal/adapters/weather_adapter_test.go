package adapters_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"weatherapi/internal/adapters"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/config"

	"github.com/stretchr/testify/require"
)

func TestWeatherAPIAdapter_CityNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"error":{"code":1006,"message":"No matching location found."}}`))
	}))
	defer mockServer.Close()

	cfg := &config.Config{WeatherKey: "fake-key"}
	adapter := adapters.NewWeatherAPIAdapter(cfg)

	originalBaseURL := adapters.WeatherAPIBaseURL
	adapters.WeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.WeatherAPIBaseURL = originalBaseURL
	}()

	_, err := adapter.FetchWeather(context.Background(), "InvalidCity")
	require.ErrorIs(t, err, apierrors.ErrCityNotFound)
}

func TestWeatherAPIAdapter_Success(t *testing.T) {
	mockResponse := `{
		"current": {
			"temp_c": 21.1,
			"humidity": 72,
			"condition": {"text": "Partly cloudy"}
		}
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, mockResponse)
	}))
	defer mockServer.Close()

	cfg := &config.Config{WeatherKey: "fake-key"}
	adapter := adapters.NewWeatherAPIAdapter(cfg)

	originalBaseURL := adapters.WeatherAPIBaseURL
	adapters.WeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.WeatherAPIBaseURL = originalBaseURL
	}()

	data, err := adapter.FetchWeather(context.Background(), "Kyiv")
	require.NoError(t, err)
	require.Equal(t, 21.1, data.Temperature)
	require.Equal(t, 72.0, data.Humidity)
	require.Equal(t, "Partly cloudy", data.Description)
}

func TestWeatherAPIAdapter_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `not-json`)
	}))
	defer mockServer.Close()

	cfg := &config.Config{WeatherKey: "fake-key"}
	adapter := adapters.NewWeatherAPIAdapter(cfg)

	originalBaseURL := adapters.WeatherAPIBaseURL
	adapters.WeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.WeatherAPIBaseURL = originalBaseURL
	}()

	_, err := adapter.FetchWeather(context.Background(), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode")
}

func TestWeatherAPIAdapter_EmptyResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(``))
	}))
	defer mockServer.Close()

	cfg := &config.Config{WeatherKey: "fake-key"}
	adapter := adapters.NewWeatherAPIAdapter(cfg)

	originalBaseURL := adapters.WeatherAPIBaseURL
	adapters.WeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.WeatherAPIBaseURL = originalBaseURL
	}()

	_, err := adapter.FetchWeather(context.Background(), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty response body")
}
