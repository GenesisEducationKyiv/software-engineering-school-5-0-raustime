package bootstrap

import (
	"context"
	"fmt"
	"weather_microservice/internal/adapters"
	"weather_microservice/internal/cache"
	"weather_microservice/internal/chain"
	"weather_microservice/internal/config"
	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/metrics"
	"weather_microservice/internal/weather_service"
)

func InitWeatherService(ctx context.Context, cfg *config.Config) (weather_service.WeatherService, error) {
	if cfg == nil {
		return weather_service.WeatherService{}, fmt.Errorf("failed to load config")
	}

	logger := logging.FromContext(ctx)
	if logger == nil {
		return weather_service.WeatherService{}, fmt.Errorf("logger not found in context")
	}

	// Register Prometheus metrics
	cacheRawMetrics := metrics.NewCacheMetrics()
	cacheMetrics := metrics.NewCacheMetricsAdapter(cacheRawMetrics)
	metrics.RegisterAll(cacheRawMetrics)
	metrics.RegisterWeatherMetrics()

	// Setup cache
	var redisCache contracts.WeatherCache
	if cfg.Cache.Enabled {
		redisCache = cache.NewRedisCache(
			ctx,
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
			cacheMetrics,
		)
	} else {
		redisCache = cache.NoopWeatherCache{}
	}

	// Setup adapters
	openWeather, err := adapters.NewOpenWeatherAdapter(cfg.OpenWeatherKey, cfg.OpenWeatherBaseURL, cfg.ExtAPITimeout)
	if err != nil {
		logger.Error(ctx, "bootstrap", nil, err)
		return weather_service.WeatherService{}, fmt.Errorf("failed to create OpenWeather adapter: %w", err)
	}

	weatherAPI, err := adapters.NewWeatherAdapter(cfg.WeatherKey, cfg.WeatherBaseURL, cfg.ExtAPITimeout)
	if err != nil {
		logger.Error(ctx, "bootstrap", nil, err)
		return weather_service.WeatherService{}, fmt.Errorf("failed to create WeatherAPI adapter: %w", err)
	}

	// Setup chain
	weatherChain := chain.NewWeatherChain(logger)

	owHandler := chain.NewBaseWeatherHandler(&openWeather, "openweather")
	waHandler := chain.NewBaseWeatherHandler(&weatherAPI, "weatherapi")
	owHandler.SetNext(waHandler)
	weatherChain.SetFirstHandler(owHandler)

	return weather_service.NewWeatherService(weatherChain, redisCache, cfg.Cache.Expiration), nil
}
