package chain

import (
	"context"
	"fmt"
	"strings"

	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/metrics"
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
	next          WeatherHandler
	api           WeatherAPIProvider
	name          string
	allowFallback bool // Allows fallback to next handler if true
}

type WeatherAPIProvider interface {
	FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error)
}

func NewBaseWeatherHandlerWithFallback(api WeatherAPIProvider, name string, allowFallback bool) *BaseWeatherHandler {
	return &BaseWeatherHandler{
		api:           api,
		name:          name,
		allowFallback: allowFallback,
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
	logger := logging.FromContext(ctx)

	if strings.Contains(city, "-") {
		// If the city has a prefix but it's not for this handler, pass to next
		if !strings.HasPrefix(city, h.name+"-") {
			if h.next != nil {
				logger.Info(ctx, "chain:skip", map[string]string{
					"handler": h.name,
					"input":   city,
					"reason":  "prefix_mismatch",
				})
				return h.next.Handle(ctx, city)
			}
			return contracts.WeatherData{}, fmt.Errorf("no handler found for provider prefix in: %s", city)
		}
	}


	logger.Info(ctx, "chain:match", map[string]string{
		"handler": h.name,
		"input":   city,
	})

	cleanCity := StripProviderPrefix(city, h.name)

	metrics.WeatherRequests.WithLabelValues(h.name, cleanCity).Inc()

	data, err := h.api.FetchWeather(ctx, cleanCity)

	

	if err != nil {
		logger.Error(ctx, h.name, nil, err)

		// Only fallback if allowed and there's no specific provider prefix
		if h.allowFallback && !strings.Contains(city, "-") {
			if h.next != nil {
				logger.Info(ctx, "chain:fallback", map[string]string{
					"from_handler": h.name,
					"city":         city,
					"error":        err.Error(),
				})
				return h.next.Handle(ctx, city)
			}
		}

		// If there was a specific provider prefix, don't fallback - return the error
		if strings.HasPrefix(city, h.name+"-") {
			return contracts.WeatherData{}, fmt.Errorf("provider %s failed for city %s: %w", h.name, cleanCity, err)
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
	if strings.HasPrefix(city, prefix) {
		return strings.TrimPrefix(city, prefix)
	}
	return city
}

func ShouldHandleCity(city, provider string) bool {
	return strings.HasPrefix(city, provider+"-")
}