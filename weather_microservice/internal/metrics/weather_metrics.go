package metrics

import "github.com/prometheus/client_golang/prometheus"

var WeatherRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "weather_api_requests_total",
		Help: "Total number of weather API requests",
	},
	[]string{"provider", "city"},
)

var WeatherFailures = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "weather_api_failures_total",
		Help: "Total number of failed weather API requests",
	},
	[]string{"provider", "city"},
)

func RegisterWeatherMetrics() {
	prometheus.MustRegister(WeatherRequests, WeatherFailures)
}
