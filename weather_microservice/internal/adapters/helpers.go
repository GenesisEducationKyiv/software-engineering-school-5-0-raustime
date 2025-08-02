package adapters

import (
	"context"
	"fmt"

	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/metrics"
)

func fail(ctx context.Context, provider, city, msg string, err error) (contracts.WeatherData, error) {
	logging.Error(ctx, "adapter:"+provider, map[string]string{"city": city}, err)
	metrics.WeatherFailures.WithLabelValues(provider, city).Inc()
	return contracts.WeatherData{}, fmt.Errorf("%s: %w", msg, err)
}
