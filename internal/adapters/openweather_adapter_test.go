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

func TestOpenWeatherAdapter_CityNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `{"cod":"404","message":"city not found"}`)
	}))
	defer mockServer.Close()

	cfg := &config.Config{OpenWeatherKey: "fake-key"}
	adapter := adapters.NewOpenWeatherAdapter(cfg)

	originalBaseURL := adapters.OpenWeatherAPIBaseURL
	adapters.OpenWeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.OpenWeatherAPIBaseURL = originalBaseURL
	}()

	_, err := adapter.FetchWeather(context.Background(), "InvalidCity")
	require.ErrorIs(t, err, apierrors.ErrCityNotFound)
}

func TestOpenWeatherAdapter_Success(t *testing.T) {
	mockResponse := `{
		"weather": [{"description": "clear sky"}],
		"main": {"temp": 25.5, "humidity": 60}
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, mockResponse)
	}))
	defer mockServer.Close()

	cfg := &config.Config{OpenWeatherKey: "fake-key"}
	adapter := adapters.NewOpenWeatherAdapter(cfg)

	originalBaseURL := adapters.OpenWeatherAPIBaseURL
	adapters.OpenWeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.OpenWeatherAPIBaseURL = originalBaseURL
	}()

	data, err := adapter.FetchWeather(context.Background(), "Kyiv")
	require.NoError(t, err)
	require.Equal(t, 25.5, data.Temperature)
	require.Equal(t, 60.0, data.Humidity)
	require.Equal(t, "clear sky", data.Description)
}

func TestOpenWeatherAdapter_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `invalid-json`)
	}))
	defer mockServer.Close()

	cfg := &config.Config{OpenWeatherKey: "fake-key"}
	adapter := adapters.NewOpenWeatherAdapter(cfg)

	originalBaseURL := adapters.OpenWeatherAPIBaseURL
	adapters.OpenWeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.OpenWeatherAPIBaseURL = originalBaseURL
	}()

	_, err := adapter.FetchWeather(context.Background(), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode")
}
