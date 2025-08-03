package integration

import (
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var client = http.Client{Timeout: 5 * time.Second}

func TestMain(m *testing.M) {
	os.Setenv("ENABLE_FALLBACK_CHAIN", "false")
	os.Exit(m.Run())
}

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

func TestWeatherService_SameCityMultipleRequests(t *testing.T) {
	for i := 0; i < 3; i++ {
		resp, err := client.Get("http://weather_service:8080/api/weather?city=Kyiv")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestWeatherService_InvalidEndpoint(t *testing.T) {
	resp, err := client.Get("http://weather_service:8080/api/invalid")
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestMetricsEndpoint(t *testing.T) {
	// Очікуємо запуск сервера (до 3 секунд)
	var resp *http.Response
	var err error
	for i := 0; i < 6; i++ {
		resp, err = client.Get("http://weather_service:8080/metrics")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	text := string(body)
	require.Contains(t, text, "weather_cache_hits_total")
	require.Contains(t, text, "weather_api_requests_total")
}
