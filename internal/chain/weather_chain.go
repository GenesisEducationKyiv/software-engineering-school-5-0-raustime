package chain

import (
	"context"
	"fmt"
	"log"
	"weatherapi/internal/contracts"
)

// WeatherHandler defines the interface for weather handlers in the chain
type WeatherHandler interface {
	SetNext(handler WeatherHandler) WeatherHandler
	Handle(ctx context.Context, city string) (contracts.WeatherData, error)
	GetProviderName() string
}

// BaseWeatherHandler provides common functionality for all handlers
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
	data, err := h.api.FetchWeather(ctx, city)

	if err != nil {
		log.Printf("Provider %s failed: %v", h.name, err)
		if h.next != nil {
			log.Printf("Trying next provider in chain...")
			return h.next.Handle(ctx, city)
		}
		return contracts.WeatherData{}, fmt.Errorf("all weather providers failed, last error from %s: %w", h.name, err)
	}

	log.Printf("Successfully got weather data from provider: %s", h.name)
	return data, nil
}

// WeatherChain manages the chain of weather providers
type WeatherChain struct {
	firstHandler WeatherHandler
	logger       WeatherLogger
}

type WeatherLogger interface {
	LogResponse(provider string, data contracts.WeatherData, err error)
}

func NewWeatherChain(logger WeatherLogger) *WeatherChain {
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

	data, err := c.firstHandler.Handle(ctx, city)

	// Логування результату на рівні ланцюга
	if c.logger != nil {
		providerName := "unknown"
		if c.firstHandler != nil {
			providerName = c.firstHandler.GetProviderName()
		}
		c.logger.LogResponse(providerName, data, err)
	}

	return data, err
}
