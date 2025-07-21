package contracts

import (
	"context"
	"time"
)

// WeatherCache визначає інтерфейс для кешування погоди.
type WeatherCache interface {
	Get(ctx context.Context, city string) (WeatherData, error)
	Set(ctx context.Context, city string, data WeatherData, expiration time.Duration) error
	Delete(ctx context.Context, city string) error
	Exists(ctx context.Context, city string) (bool, error)
	Health(ctx context.Context) error
	Close() error
	GetStats(ctx context.Context) (map[string]interface{}, error)
}
