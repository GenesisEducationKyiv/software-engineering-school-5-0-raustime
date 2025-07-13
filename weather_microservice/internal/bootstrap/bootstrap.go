package bootstrap

import (
	"fmt"
	"weather_microservice/internal/adapters"
	"weather_microservice/internal/cache"
	"weather_microservice/internal/chain"
	"weather_microservice/internal/config"
	"weather_microservice/internal/contracts"
	"weather_microservice/internal/weather_service"
)

func InitWeatherService() (weather_service.WeatherService, error) {
	cfg := config.Load()
	if cfg == nil {
		return weather_service.WeatherService{}, fmt.Errorf("failed to load config")
	}

	// Register metrics.
	metrics := cache.NewPrometheusMetrics()
	metrics.Register()

	// Setup cache
	var redisCache contracts.WeatherCache
	if cfg.Cache.Enabled {
		redisCache = cache.NewRedisCache(
			cache.RedisConfig{
				Addr:     cfg.Cache.Redis.Addr,
				Password: cfg.Cache.Redis.Password,
				DB:       cfg.Cache.Redis.DB,
				PoolSize: cfg.Cache.Redis.PoolSize,
				Timeout:  cfg.Cache.Redis.Timeout,
			},
			cache.CacheConfig{
				IsEnabled:         cfg.Cache.Enabled,
				DefaultExpiration: cfg.Cache.Expiration,
			},
			metrics,
		)
		if redisCache == nil {
			return weather_service.WeatherService{}, fmt.Errorf("failed to initialize Redis cache")
		}
	} else {
		redisCache = cache.NoopWeatherCache{}
	}

	// Setup adapters
	openWeather, err := adapters.NewOpenWeatherAdapter(cfg.OpenWeatherKey)
	if err != nil {
		return weather_service.WeatherService{}, fmt.Errorf("failed to create OpenWeather adapter: %w", err)
	}

	weatherAPI, err := adapters.NewWeatherAPIAdapter(cfg.WeatherKey)
	if err != nil {
		return weather_service.WeatherService{}, fmt.Errorf("failed to create WeatherAPI adapter: %w", err)
	}

	// Setup chain
	weatherChain := chain.NewWeatherChain(nil)
	owHandler := chain.NewBaseWeatherHandler(&openWeather, "openweather")
	waHandler := chain.NewBaseWeatherHandler(&weatherAPI, "weatherapi")
	owHandler.SetNext(waHandler)
	weatherChain.SetFirstHandler(owHandler)

	return weather_service.NewWeatherService(weatherChain, redisCache, cfg.Cache.Expiration), nil
}
