package integration

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWeatherService_HTTP(t *testing.T) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://weather_service:8080/api/weather?city=Kyiv")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestWeatherService_InvalidCity(t *testing.T) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://weather_service:8080/api/weather?city=InvalidCity")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestWeatherService_MissingCityParam(t *testing.T) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://weather_service:8080/api/weather")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
