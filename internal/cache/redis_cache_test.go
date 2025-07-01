package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"weatherapi/internal/contracts"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---- MOCK ----

type MockRedis struct {
	mock.Mock
}

func (m *MockRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return redis.NewStringResult(args.String(0), args.Error(1))
}

func (m *MockRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return redis.NewStatusResult("", args.Error(0))
}

func (m *MockRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return redis.NewIntResult(int64(args.Int(0)), args.Error(1))
}

func (m *MockRedis) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return redis.NewIntResult(int64(args.Int(0)), args.Error(1))
}

func (m *MockRedis) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return redis.NewStatusResult("", args.Error(0))
}

func (m *MockRedis) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedis) Info(ctx context.Context, section ...string) *redis.StringCmd {
	args := m.Called(ctx, section)
	return redis.NewStringResult(args.String(0), args.Error(1))
}

func (m *MockRedis) PoolStats() *redis.PoolStats {
	args := m.Called()
	return args.Get(0).(*redis.PoolStats)
}

// ---- TESTS ----

func TestRedisCache_SetAndGet(t *testing.T) {
	mockRedis := new(MockRedis)
	cache := &RedisCache{
		client: mockRedis,
		config: CacheConfig{
			IsEnabled:         true,
			DefaultExpiration: 10 * time.Minute,
		},
		metrics: NoopMetrics{},
	}

	city := "Kyiv"
	data := contracts.WeatherData{
		Temperature: 22.5,
		Humidity:    65,
		Description: "Cloudy",
	}

	jsonData, _ := json.Marshal(data)
	key := "weather:kyiv"

	mockRedis.On(
		"Set",
		mock.Anything,
		key,
		mock.MatchedBy(func(arg interface{}) bool {
			b, ok := arg.([]byte)
			return ok && string(b) == string(jsonData)
		}),
		10*time.Minute,
	).Return(nil)

	err := cache.Set(context.Background(), city, data, 0)
	assert.NoError(t, err)

	mockRedis.On("Get", mock.Anything, key).Return(string(jsonData), nil)

	got, err := cache.Get(context.Background(), city)
	assert.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestRedisCache_Get_NotFound(t *testing.T) {
	mockRedis := new(MockRedis)
	cache := &RedisCache{
		client:  mockRedis,
		config:  CacheConfig{},
		metrics: NoopMetrics{},
	}

	key := "weather:unknown"
	mockRedis.On("Get", mock.Anything, key).Return("", redis.Nil)

	_, err := cache.Get(context.Background(), "unknown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cache miss")
}

func TestRedisCache_Exists(t *testing.T) {
	mockRedis := new(MockRedis)
	cache := &RedisCache{
		client:  mockRedis,
		config:  CacheConfig{},
		metrics: NoopMetrics{},
	}

	key := "weather:kyiv"
	mockRedis.On("Exists", mock.Anything, []string{key}).Return(1, nil)

	exists, err := cache.Exists(context.Background(), "Kyiv")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestRedisCache_Delete(t *testing.T) {
	mockRedis := new(MockRedis)
	cache := &RedisCache{
		client:  mockRedis,
		config:  CacheConfig{},
		metrics: NoopMetrics{},
	}

	key := "weather:kyiv"
	mockRedis.On("Del", mock.Anything, []string{key}).Return(1, nil)

	err := cache.Delete(context.Background(), "Kyiv")
	assert.NoError(t, err)
}

func TestRedisCache_Health(t *testing.T) {
	mockRedis := new(MockRedis)
	cache := &RedisCache{
		client:  mockRedis,
		config:  CacheConfig{},
		metrics: NoopMetrics{},
	}

	mockRedis.On("Ping", mock.Anything).Return(nil)

	err := cache.Health(context.Background())
	assert.NoError(t, err)
}
