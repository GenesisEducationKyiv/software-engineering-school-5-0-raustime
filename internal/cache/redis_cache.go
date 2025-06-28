package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"weatherapi/internal/contracts"

	"github.com/redis/go-redis/v9"
)

const (
	// Cache key prefix for weather data
	weatherCachePrefix = "weather:"

	// Default Redis configuration
	defaultRedisDB       = 0
	defaultRedisPoolSize = 10
	defaultRedisTimeout  = 5 * time.Second
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
	Info(ctx context.Context, section ...string) *redis.StringCmd
	PoolStats() *redis.PoolStats
}

// CacheConfig holds cache configuration such as default expiration
type CacheConfig struct {
	DefaultExpiration time.Duration `json:"default_expiration"`
}

// RedisCache implements WeatherCache using Redis
type RedisCache struct {
	client RedisClient
	config CacheConfig
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string        `json:"addr"`      // Redis server address (host:port)
	Password string        `json:"password"`  // Redis password (optional)
	DB       int           `json:"db"`        // Redis database number
	PoolSize int           `json:"pool_size"` // Connection pool size
	Timeout  time.Duration `json:"timeout"`   // Connection timeout
}

// DefaultRedisConfig returns default Redis configuration
func DefaultRedisConfig() RedisConfig {
	return RedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       defaultRedisDB,
		PoolSize: defaultRedisPoolSize,
		Timeout:  defaultRedisTimeout,
	}
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(redisConfig RedisConfig, cacheConfig CacheConfig) (*RedisCache, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
		PoolSize: redisConfig.PoolSize,

		// Connection timeouts
		DialTimeout:  redisConfig.Timeout,
		ReadTimeout:  redisConfig.Timeout,
		WriteTimeout: redisConfig.Timeout,

		// Pool timeouts
		PoolTimeout: redisConfig.Timeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), redisConfig.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		config: cacheConfig,
	}, nil
}

// generateCacheKey creates a cache key for a city
func (r *RedisCache) generateCacheKey(city string) string {
	// Normalize city name (lowercase, trim spaces)
	normalizedCity := strings.ToLower(strings.TrimSpace(city))
	return weatherCachePrefix + normalizedCity
}

// Get retrieves weather data from Redis cache
func (r *RedisCache) Get(ctx context.Context, city string) (contracts.WeatherData, error) {
	key := r.generateCacheKey(city)

	// Get data from Redis
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Cache miss - return empty data with no error
			return contracts.WeatherData{}, fmt.Errorf("cache miss for city: %s", city)
		}
		return contracts.WeatherData{}, fmt.Errorf("failed to get from cache: %w", err)
	}

	// Deserialize JSON data
	var data contracts.WeatherData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return data, nil
}

// Set stores weather data in Redis cache with expiration
func (r *RedisCache) Set(ctx context.Context, city string, data contracts.WeatherData, expiration time.Duration) error {
	key := r.generateCacheKey(city)

	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal weather data: %w", err)
	}

	// Use default expiration if none provided
	if expiration <= 0 {
		expiration = r.config.DefaultExpiration
	}

	// Set data in Redis with expiration
	if err := r.client.Set(ctx, key, jsonData, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Delete removes weather data from Redis cache
func (r *RedisCache) Delete(ctx context.Context, city string) error {
	key := r.generateCacheKey(city)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// Exists checks if weather data exists in Redis cache
func (r *RedisCache) Exists(ctx context.Context, city string) (bool, error) {
	key := r.generateCacheKey(city)

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache existence: %w", err)
	}

	return count > 0, nil
}

// Health checks Redis connection health
func (r *RedisCache) Health(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// GetStats returns Redis cache statistics
func (r *RedisCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Get Redis info
	info, err := r.client.Info(ctx, "memory", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis stats: %w", err)
	}

	// Get connection pool stats
	poolStats := r.client.PoolStats()

	stats := map[string]interface{}{
		"redis_info":    info,
		"pool_hits":     poolStats.Hits,
		"pool_misses":   poolStats.Misses,
		"pool_timeouts": poolStats.Timeouts,
		"pool_total":    poolStats.TotalConns,
		"pool_idle":     poolStats.IdleConns,
		"pool_stale":    poolStats.StaleConns,
	}

	return stats, nil
}
