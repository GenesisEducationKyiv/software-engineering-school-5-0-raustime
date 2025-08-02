package chain

import (
	"context"
	"fmt"

	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/pkg/ctxkeys"
)

// WeatherHandler defines the interface for weather handlers in the chain.
type WeatherHandler interface {
	SetNext(handler WeatherHandler) WeatherHandler
	Handle(ctx context.Context, city string) (contracts.WeatherData, error)
	GetProviderName() string
}

// BaseWeatherHandler provides common functionality for all handlers.
type BaseWeatherHandler struct {
	next WeatherHandler
	api  WeatherAPIProvider
	name string
}

type WeatherAPIProvider interface {
	FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error)
}

func NewBaseWeatherHandler(api WeatherAPIProvider, name string) *BaseWeatherHandler {
	return &BaseWeatherHandler{
		api:  api,
		name: name,
	}
}

func (h *BaseWeatherHandler) SetNext(handler WeatherHandler) WeatherHandler {
	h.next = handler
	return handler
}

func (h *BaseWeatherHandler) GetProviderName() string {
	return h.name
}

func (h *BaseWeatherHandler) Handle(ctx context.Context, city string) (contracts.WeatherData, error) {
	cleanCity := StripProviderPrefix(city, h.name)

	data, err := h.api.FetchWeather(ctx, cleanCity)

	logger := logging.FromContext(ctx)
	if err != nil {
		logger.Error(ctx, h.name, nil, err)
		if h.next != nil {
			return h.next.Handle(ctx, city)
		}
		return contracts.WeatherData{}, fmt.Errorf("all weather providers failed, last error from %s: %w", h.name, err)
	}

	logger.Info(ctx, h.name, data)
	return data, nil
}

// WeatherChain manages the chain of weather providers.
type WeatherChain struct {
	firstHandler WeatherHandler
	logger       logging.Logger
}

func NewWeatherChain(logger logging.Logger) *WeatherChain {
	return &WeatherChain{
		logger: logger,
	}
}

func (c *WeatherChain) SetFirstHandler(handler WeatherHandler) {
	c.firstHandler = handler
}

func (c *WeatherChain) GetWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	if c.firstHandler == nil {
		return contracts.WeatherData{}, fmt.Errorf("no weather providers configured")
	}

	// Insert logger in context using a custom key type.
	ctx = context.WithValue(ctx, ctxkeys.Logger, c.logger)

	return c.firstHandler.Handle(ctx, city)
}

func StripProviderPrefix(city, provider string) string {
	prefix := provider + "-"
	if len(city) > len(prefix) && city[:len(prefix)] == prefix {
		return city[len(prefix):]
	}
	return city
}
