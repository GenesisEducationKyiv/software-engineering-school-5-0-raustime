package bootstrap

import (
	"fmt"
	"weather_microservice/internal/adapters"
	"weather_microservice/internal/cache"
	"weather_microservice/internal/chain"
	"weather_microservice/internal/config"
	"weather_microservice/internal/weather_service"
	"weather_microservice/internal/logging"
)

func InitWeatherService(cfg *config.Config) (weather_service.WeatherService, error) {
	if cfg == nil {
		return weather_service.WeatherService{}, fmt.Errorf("failed to load config")
	}

	// Register metrics.
	metrics := cache.NewPrometheusMetrics()
	metrics.Register()

	// Setup cache
	var redisCache contracts.WeatherCache
	if cfg.Cache.Enabled {
		redisCache =  cache.NewRedisCache(
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
		if err != nil {
    		return weather_service.WeatherService{}, fmt.Errorf("failed to init redis cache: %w", err)
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

	// Setup logger
	logger := logging.NewFileWeatherLogger("weather.log")
		
	// Setup chain
	weatherChain := chain.NewWeatherChain(logger)

	owHandler := chain.NewBaseWeatherHandler(&openWeather, "openweather")
	waHandler := chain.NewBaseWeatherHandler(&weatherAPI, "weatherapi")
	owHandler.SetNext(waHandler)
	weatherChain.SetFirstHandler(owHandler)

	return weather_service.NewWeatherService(weatherChain, redisCache, cfg.Cache.Expiration), nil
}
