package adapters_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"weatherapi/internal/adapters"
	"weatherapi/internal/apierrors"
)

func TestWeatherAPIAdapter_CityNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error":{"code":1006,"message":"No matching location found."}}`))
	}))
	defer mockServer.Close()

	adapter, err := adapters.NewWeatherAPIAdapter("fake-key")

	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	originalBaseURL := adapters.WeatherAPIBaseURL
	adapters.WeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.WeatherAPIBaseURL = originalBaseURL
	}()

	_, err = adapter.FetchWeather(context.Background(), "InvalidCity")
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer mockServer.Close()

	adapter, err := adapters.NewWeatherAPIAdapter("fake-key")
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

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
		_, _ = fmt.Fprintln(w, `not-json`)
	}))
	defer mockServer.Close()

	adapter, err := adapters.NewWeatherAPIAdapter("fake-key")

	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	originalBaseURL := adapters.WeatherAPIBaseURL
	adapters.WeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.WeatherAPIBaseURL = originalBaseURL
	}()

	_, err = adapter.FetchWeather(context.Background(), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode")
}

func TestWeatherAPIAdapter_EmptyResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(``))
	}))
	defer mockServer.Close()

	adapter, err := adapters.NewWeatherAPIAdapter("fake-key")

	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	originalBaseURL := adapters.WeatherAPIBaseURL
	adapters.WeatherAPIBaseURL = func() string {
		return mockServer.URL
	}
	defer func() {
		adapters.WeatherAPIBaseURL = originalBaseURL
	}()

	_, err = adapter.FetchWeather(context.Background(), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty response body")
}
