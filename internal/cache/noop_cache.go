package cache

import (
	"context"
	"time"

	"fmt"
	"weatherapi/internal/contracts"
)

// NoopWeatherCache — реалізація WeatherCache, яка нічого не зберігає і не повертає.
type NoopWeatherCache struct{}

// Get завжди повертає помилку кеш-місу.
func (NoopWeatherCache) Get(ctx context.Context, city string) (contracts.WeatherData, error) {
	return contracts.WeatherData{}, fmt.Errorf("noop cache miss for city: %s", city)
}

// Set нічого не зберігає.
func (NoopWeatherCache) Set(ctx context.Context, city string, data contracts.WeatherData, expiration time.Duration) error {
	return nil
}

// Delete нічого не видаляє.
func (NoopWeatherCache) Delete(ctx context.Context, city string) error {
	return nil
}

// Exists завжди повертає false.
func (NoopWeatherCache) Exists(ctx context.Context, city string) (bool, error) {
	return false, nil
}

// Health завжди успішний.
func (NoopWeatherCache) Health(ctx context.Context) error {
	return nil
}

// Close нічого не закриває.
func (NoopWeatherCache) Close() error {
	return nil
}

// GetStats повертає порожній map.
func (NoopWeatherCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
