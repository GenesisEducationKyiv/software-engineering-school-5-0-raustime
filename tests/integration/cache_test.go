package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"weatherapi/internal/cache"
	"weatherapi/internal/contracts"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func TestRedisCache_Integration(t *testing.T) {
	// Load environment variables from .env file for testing
	_ = godotenv.Load()

	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup Redis client for testing (using environment variables)
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		DB:       1,  // Use a separate DB for tests
		Password: "", // Add password if needed
	})

	// Test connection
	ctx := context.Background()
	maxAttempts := 5
	var pingErr error
	for i := 0; i < maxAttempts; i++ {
		pingErr = client.Ping(ctx).Err()
		if pingErr == nil {
			break
		}
		time.Sleep(500 * time.Millisecond) // Затримка перед наступною спробою
	}
	if pingErr != nil {
		t.Fatalf("Failed to connect to Redis after %d attempts: %v", maxAttempts, pingErr)
	}

	// Clean up before and after test
	t.Cleanup(func() {
		client.FlushDB(ctx)
		_ = client.Close()
	})

	// Create cache instance
	redisConfig := cache.RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		DB:       1,
		Password: "",
		PoolSize: 10,
		Timeout:  5 * time.Second,
	}

	cacheConfig := cache.CacheConfig{
		IsEnabled:         true,
		DefaultExpiration: 1 * time.Minute,
	}

	mockMetrics := &MockMetrics{}
	cacheInstance, err := cache.NewRedisCache(redisConfig, cacheConfig, mockMetrics)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer func() { _ = cacheInstance.Close() }()

	// Test data
	testCity := "Kyiv"
	testData := contracts.WeatherData{
		Temperature: 25.5,
		Humidity:    60.0,
		Description: "Clear sky",
	}

	t.Run("Set and Get", func(t *testing.T) {
		// Set data
		err := cacheInstance.Set(ctx, testCity, testData, 1*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Get data
		result, err := cacheInstance.Get(ctx, testCity)
		if err != nil {
			t.Fatalf("Failed to get from cache: %v", err)
		}

		// Verify data
		if result.Temperature != testData.Temperature {
			t.Errorf("Expected temperature %f, got %f", testData.Temperature, result.Temperature)
		}
		if result.Humidity != testData.Humidity {
			t.Errorf("Expected humidity %f, got %f", testData.Humidity, result.Humidity)
		}
		if result.Description != testData.Description {
			t.Errorf("Expected description %s, got %s", testData.Description, result.Description)
		}
	})

	t.Run("Cache Miss", func(t *testing.T) {
		// Try to get non-existent data
		_, err := cacheInstance.Get(ctx, "NonExistentCity")
		if err == nil {
			t.Error("Expected error for cache miss")
		}
	})

	t.Run("Exists", func(t *testing.T) {
		// Set data
		err := cacheInstance.Set(ctx, testCity, testData, 1*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Check if exists
		exists, err := cacheInstance.Exists(ctx, testCity)
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if !exists {
			t.Error("Expected data to exist in cache")
		}

		// Check non-existent
		exists, err = cacheInstance.Exists(ctx, "NonExistentCity")
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists {
			t.Error("Expected data to not exist in cache")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Set data
		err := cacheInstance.Set(ctx, testCity, testData, 1*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Delete data
		err = cacheInstance.Delete(ctx, testCity)
		if err != nil {
			t.Fatalf("Failed to delete from cache: %v", err)
		}

		// Verify deletion
		exists, err := cacheInstance.Exists(ctx, testCity)
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists {
			t.Error("Expected data to be deleted from cache")
		}
	})

	t.Run("Expiration", func(t *testing.T) {
		// Set data with short expiration
		err := cacheInstance.Set(ctx, testCity, testData, 100*time.Millisecond)
		if err != nil {
			t.Fatalf("Failed to set cache: %v", err)
		}

		// Verify data exists
		exists, err := cacheInstance.Exists(ctx, testCity)
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if !exists {
			t.Error("Expected data to exist in cache")
		}

		// Wait for expiration
		time.Sleep(200 * time.Millisecond)

		// Verify data expired
		exists, err = cacheInstance.Exists(ctx, testCity)
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists {
			t.Error("Expected data to be expired")
		}
	})

	t.Run("Health Check", func(t *testing.T) {
		err := cacheInstance.Health(ctx)
		if err != nil {
			t.Errorf("Health check failed: %v", err)
		}
	})
}
