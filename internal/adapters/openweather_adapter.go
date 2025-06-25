package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/config"
	"weatherapi/internal/contracts"
)

type OpenWeatherAdapter struct {
	config *config.Config
}

var OpenWeatherAPIBaseURL = func() string {
	return "https://api.openweathermap.org/data/2.5"
}

func NewOpenWeatherAdapter(cfg *config.Config) *OpenWeatherAdapter {
	return &OpenWeatherAdapter{
		config: cfg,
	}
}

func (a *OpenWeatherAdapter) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	if a.config.OpenWeatherKey == "" {
		return contracts.WeatherData{}, fmt.Errorf("OPENWEATHER_API_KEY is not configured")
	}

	url := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=metric",
		OpenWeatherAPIBaseURL(), city, a.config.OpenWeatherKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to get weather: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("warning: failed to close response body: %v", cerr)
		}
	}()

	// Handle 404 (city not found)
	if resp.StatusCode == http.StatusNotFound {
		var errResp struct {
			Cod     string `json:"cod"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Message == "city not found" {
			return contracts.WeatherData{}, apierrors.ErrCityNotFound
		}
		return contracts.WeatherData{}, fmt.Errorf("weather API 404: %s", errResp.Message)
	}

	if resp.StatusCode != 200 {
		return contracts.WeatherData{}, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherResp struct {
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity float64 `json:"humidity"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to decode weather response: %w", err)
	}

	if len(weatherResp.Weather) == 0 {
		return contracts.WeatherData{}, fmt.Errorf("no weather data found")
	}

	// Convert to internal format
	return contracts.WeatherData{
		Temperature: weatherResp.Main.Temp,
		Humidity:    weatherResp.Main.Humidity,
		Description: weatherResp.Weather[0].Description,
	}, nil
}
