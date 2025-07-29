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

func TestWeatherService_DefaultProvider_NoPrefix(t *testing.T) {
	client := http.Client{Timeout: 5 * time.Second}

	before := fetchMetric("weather_api_requests_total", "weatherapi", "Kyiv")

	resp, err := client.Get("http://weather_service:8080/api/weather?city=Kyiv")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	time.Sleep(1 * time.Second)

	after := fetchMetric("weather_api_requests_total", "weatherapi", "Kyiv")
	require.Greater(t, after, before, "weather_api_requests_total should increment for default provider without prefix")
}

func fetchMetric(metric, provider, city string) int {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://weather_service:8080/metrics")
	if err != nil {
		return -1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1
	}

	linePrefix := fmt.Sprintf(`%s{city="%s",provider="%s"} `, metric, city, provider)
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, linePrefix) {
			var count int
			fmt.Sscanf(line[len(linePrefix):], "%d", &count)
			return count
		}
	}
	return 0
}
