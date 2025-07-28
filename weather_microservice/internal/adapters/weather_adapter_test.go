package adapters_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"weather_microservice/internal/adapters"
	"weather_microservice/internal/apierrors"
)

func newWeatherAdapter(t *testing.T, baseURL string) *adapters.WeatherAdapter {
	t.Helper()

	adapter, err := adapters.NewWeatherAdapter("fake-key", baseURL, 3*time.Second)
	require.NoError(t, err)

	return &adapter
}

func TestWeatherAdapter_CityNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error":{"code":1006,"message":"No matching location found."}}`))
	}))
	defer mockServer.Close()

	adapter := newOpenWeatherAdapter(t, mockServer.URL)

	_, err := adapter.FetchWeather(withLogger(context.Background()), "InvalidCity")
	require.ErrorIs(t, err, apierrors.ErrCityNotFound)
}

func TestWeatherAdapter_Success(t *testing.T) {
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

	adapter := newOpenWeatherAdapter(t, mockServer.URL)

	data, err := adapter.FetchWeather(withLogger(context.Background()), "Kyiv")
	require.NoError(t, err)
	require.Equal(t, 21.1, data.Temperature)
	require.Equal(t, 72.0, data.Humidity)
	require.Equal(t, "Partly cloudy", data.Description)
}

func TestWeatherAdapter_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `not-json`)
	}))
	defer mockServer.Close()

	adapter := newOpenWeatherAdapter(t, mockServer.URL)

	_, err := adapter.FetchWeather(withLogger(context.Background()), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode")
}

func TestWeatherAdapter_EmptyResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(``))
	}))
	defer mockServer.Close()

	adapter := newAdapter(t, mockServer.URL)

	_, err := adapter.FetchWeather(withLogger(context.Background()), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "empty response body")
}
