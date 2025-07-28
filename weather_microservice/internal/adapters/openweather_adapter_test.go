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

func newAdapter(t *testing.T, baseURL string) *adapters.OpenWeatherAdapter {
	t.Helper()

	adapter, err := adapters.NewOpenWeatherAdapter("fake-key", baseURL, 3*time.Second)
	require.NoError(t, err)

	return &adapter
}

func TestOpenWeatherAdapter_CityNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintln(w, `{"cod":"404","message":"city not found"}`)
	}))
	defer mockServer.Close()

	adapter := newAdapter(t, mockServer.URL)

	_, err := adapter.FetchWeather(withLogger(context.Background()), "InvalidCity")
	require.ErrorIs(t, err, apierrors.ErrCityNotFound)
}

func TestOpenWeatherAdapter_Success(t *testing.T) {
	mockResponse := `{
		"weather": [{"description": "clear sky"}],
		"main": {"temp": 25.5, "humidity": 60}
	}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, mockResponse)
	}))
	defer mockServer.Close()

	adapter := newAdapter(t, mockServer.URL)

	data, err := adapter.FetchWeather(withLogger(context.Background()), "Kyiv")
	require.NoError(t, err)
	require.Equal(t, 25.5, data.Temperature)
	require.Equal(t, 60.0, data.Humidity)
	require.Equal(t, "clear sky", data.Description)
}

func TestOpenWeatherAdapter_InvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, `invalid-json`)
	}))
	defer mockServer.Close()

	adapter := newAdapter(t, mockServer.URL)

	_, err := adapter.FetchWeather(withLogger(context.Background()), "Kyiv")
	require.Error(t, err)
	require.Contains(t, err.Error(), "decode")
}
