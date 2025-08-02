package integration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWeatherService_DefaultProvider(t *testing.T) {
	city := "Kyiv"
	provider := detectProvider(city)
	require.NotEmpty(t, provider, "default provider not detected")

	before := fetchMetric("weather_api_requests_total", provider, city)

	resp, err := client.Get("http://weather_service:8080/api/weather?city=" + city)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	time.Sleep(500 * time.Millisecond)

	after := fetchMetric("weather_api_requests_total", provider, city)
	require.Greater(t, after, before, fmt.Sprintf("metric should increment for %s:%s", provider, city))
}

func TestWeatherService_WithProviderPrefix(t *testing.T) {
	tests := []struct {
		provider string
		city     string
		expected int
	}{
		{"weatherapi", "Kyiv", http.StatusOK},
		{"openweather", "Kyiv", http.StatusOK},
		{"weatherapi", "InvalidCity", http.StatusNotFound},
		{"openweather", "InvalidCity", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.provider+"_"+tt.city, func(t *testing.T) {
			fullCity := fmt.Sprintf("%s-%s", tt.provider, tt.city)

			resp, err := client.Get("http://weather_service:8080/api/weather?city=" + fullCity)
			require.NoError(t, err)
			require.Equal(t, tt.expected, resp.StatusCode)
			resp.Body.Close()

			time.Sleep(1 * time.Second)

			metric := "weather_api_requests_total"
			if tt.expected != http.StatusOK {
				metric = "weather_api_failures_total"
			}

			count := fetchMetric(metric, tt.provider, tt.city)
			require.GreaterOrEqual(t, count, 1, fmt.Sprintf("expected %s > 0 for %s:%s", metric, tt.provider, tt.city))
		})
	}
}

func fetchMetric(metric, provider, city string) int {
	resp, err := client.Get("http://weather_service:8080/metrics")
	if err != nil {
		return -1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1
	}

	// Пошук exact match з правильним порядком лейблів
	// Прометей генерує їх у форматі: metric{label1="val1",label2="val2"} <value>
	prefix := fmt.Sprintf(`%s{provider="%s",city="%s"} `, metric, provider, city)
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, prefix) {
			var count int
			fmt.Sscanf(line[len(prefix):], "%d", &count)
			return count
		}
	}
	return 0
}

func detectProvider(city string) string {
	providers := []string{"weatherapi", "openweather"}
	for _, p := range providers {
		fullCity := fmt.Sprintf("%s-%s", p, city)
		resp, err := client.Get("http://weather_service:8080/api/weather?city=" + fullCity)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return p
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	return ""
}
