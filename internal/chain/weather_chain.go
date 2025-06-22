package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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

	// Log the response
	h.logResponse(data, err)

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

func (h *BaseWeatherHandler) logResponse(data contracts.WeatherData, err error) {
	logEntry := map[string]interface{}{
		"provider": h.name,
		"success":  err == nil,
	}

	if err != nil {
		logEntry["error"] = err.Error()
	} else {
		logEntry["response"] = data
	}

	logJSON, _ := json.Marshal(logEntry)

	// Log to file
	h.logToFile(string(logJSON))

	// Also log to console for debugging
	log.Printf("%s - Response: %s", h.name, string(logJSON))
}

func (h *BaseWeatherHandler) logToFile(logMessage string) {
	file, err := os.OpenFile("weather_providers.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("%s\n", logMessage)); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

// WeatherChain manages the chain of weather providers
type WeatherChain struct {
	firstHandler WeatherHandler
}

func NewWeatherChain() *WeatherChain {
	return &WeatherChain{}
}

func (c *WeatherChain) SetFirstHandler(handler WeatherHandler) {
	c.firstHandler = handler
}

func (c *WeatherChain) GetWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	if c.firstHandler == nil {
		return contracts.WeatherData{}, fmt.Errorf("no weather providers configured")
	}

	return c.firstHandler.Handle(ctx, city)
}
