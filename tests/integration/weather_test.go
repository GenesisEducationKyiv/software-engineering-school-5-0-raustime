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

func TestWeatherService_Caching(t *testing.T) {
	client := http.Client{Timeout: 5 * time.Second}
	url := "http://weather_service:8080/api/weather?city=Kyiv"

	start1 := time.Now()
	resp1, err := client.Get(url)
	duration1 := time.Since(start1)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp1.StatusCode)

	start2 := time.Now()
	resp2, err := client.Get(url)
	duration2 := time.Since(start2)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	require.Less(t, duration2.Milliseconds(), duration1.Milliseconds())
}

func TestWeatherService_MissingCityParam(t *testing.T) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://weather_service:8080/api/weather")
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
