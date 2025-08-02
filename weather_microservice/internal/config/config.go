package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppBaseURL             string
	Port                   string
	GRPCPort               string
	ExtAPITimeout          time.Duration
	OpenWeatherBaseURL     string
	OpenWeatherKey         string
	WeatherBaseURL         string
	WeatherKey             string
	SubscriptionServiceURL string
	NATSUrl                string
	Environment            string
	Cache                  CacheConfig
	LogPath                string
	LogLevelDefault        string
}

type CacheConfig struct {
	Enabled    bool
	Expiration time.Duration
	Redis      RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
	Timeout  time.Duration
}

// Load завантажує конфігурацію з змінних оточення.
func Load() *Config {

	// Redis + Cache.
	cacheEnabled := strings.ToLower(getEnv("CACHE_ENABLED", "false"))
	enabled := cacheEnabled == "true" || cacheEnabled == "1" || cacheEnabled == "yes"

	expirationMinutes, err := strconv.Atoi(getEnv("CACHE_EXPIRATION_MINUTES", "10"))
	if err != nil {
		fmt.Printf("Invalid CACHE_EXPIRATION_MINUTES, using default 10 minutes: %v\n", err)
		expirationMinutes = 10
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	redisPoolSize, _ := strconv.Atoi(getEnv("REDIS_POOL_SIZE", "10"))
	redisTimeoutSec, _ := strconv.Atoi(getEnv("REDIS_TIMEOUT_SECONDS", "5"))

	cacheConfig := CacheConfig{
		Enabled:    enabled,
		Expiration: time.Duration(expirationMinutes) * time.Minute,
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
			PoolSize: redisPoolSize,
			Timeout:  time.Duration(redisTimeoutSec) * time.Second,
		},
	}

	ExtAPITimeoutSec, _ := strconv.Atoi(getEnv("EXT_API_TIMEOUT_SECONDS", "10"))

	return &Config{
		AppBaseURL:             getEnv("APP_BASE_URL", "http://localhost:8080"),
		Port:                   getEnv("PORT", "8080"),
		GRPCPort:               getEnv("GRPC_PORT", "8081"),
		ExtAPITimeout:          time.Duration(ExtAPITimeoutSec) * time.Second,
		OpenWeatherBaseURL:     getEnv("OPENWEATHER_BASE_URL", "https://api.openweathermap.org/data/2.5"),
		OpenWeatherKey:         getEnv("OPENWEATHER_API_KEY", ""),
		WeatherBaseURL:         getEnv("OPENWEATHER_BASE_URL", "https://api.weatherapi.com/v1"),
		WeatherKey:             getEnv("WEATHER_API_KEY", ""),
		SubscriptionServiceURL: getEnv("SUBSCRIPTION_SERVICE_URL", "http://localhost:8091"),
		NATSUrl:                getEnv("NATS_URL", "nats://localhost:4222"),
		Environment:            strings.ToLower(getEnv("ENVIRONMENT", "development")),
		Cache:                  cacheConfig,
		LogPath:                getEnv("LOG_PATH", "weather.log"),
		LogLevelDefault:        getEnv("LOG_LEVEL_DEFAULT", "Info"),
	}

}

func LoadTestConfig() *Config {
	cfg := Load()
	if cfg != nil {
		cfg.Environment = "test"
	}
	return cfg
}

// IsProduction перевіряє чи додаток працює в продакшен режимі.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment перевіряє чи додаток працює в режимі розробки.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsTest перевіряє чи додаток працює в тестовому режимі.
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

// Validate перевіряє чи всі обов'язкові конфігурації встановлені.
func (c *Config) Validate() error {
	var errors []string

	if c.OpenWeatherKey == "" {
		errors = append(errors, "OPENWEATHER_API_KEY is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// getEnv отримує значення змінної оточення або повертає значення за замовчуванням.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
