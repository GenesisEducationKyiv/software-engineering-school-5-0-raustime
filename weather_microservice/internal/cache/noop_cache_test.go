package cache

import (
	"context"
	"testing"
	"time"

	"weather_microservice/internal/contracts"
)

func TestNoopWeatherCache(t *testing.T) {
	var c contracts.WeatherCache = NoopWeatherCache{}
	ctx := context.Background()

	t.Run("Get returns cache miss error", func(t *testing.T) {
		_, err := c.Get(ctx, "Kyiv")
		if err == nil {
			t.Error("expected error from Get, got nil")
		}
	})

	t.Run("Set does nothing and returns nil", func(t *testing.T) {
		err := c.Set(ctx, "Kyiv", contracts.WeatherData{
			Temperature: 20,
			Humidity:    50,
			Description: "Sunny",
		}, 1*time.Minute)
		if err != nil {
			t.Errorf("expected nil from Set, got: %v", err)
		}
	})

	t.Run("Delete returns nil", func(t *testing.T) {
		err := c.Delete(ctx, "Kyiv")
		if err != nil {
			t.Errorf("expected nil from Delete, got: %v", err)
		}
	})

	t.Run("Exists always returns false", func(t *testing.T) {
		exists, err := c.Exists(ctx, "Kyiv")
		if err != nil {
			t.Errorf("expected nil from Exists, got: %v", err)
		}
		if exists {
			t.Error("expected false from Exists, got true")
		}
	})

	t.Run("Health always returns nil", func(t *testing.T) {
		err := c.Health(ctx)
		if err != nil {
			t.Errorf("expected nil from Health, got: %v", err)
		}
	})

	t.Run("Close returns nil", func(t *testing.T) {
		err := c.Close()
		if err != nil {
			t.Errorf("expected nil from Close, got: %v", err)
		}
	})

	t.Run("GetStats returns empty map", func(t *testing.T) {
		stats, err := c.GetStats(ctx)
		if err != nil {
			t.Errorf("expected nil from GetStats, got: %v", err)
		}
		if stats == nil || len(stats) != 0 {
			t.Errorf("expected empty stats, got: %v", stats)
		}
	})
}
