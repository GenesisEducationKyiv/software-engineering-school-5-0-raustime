package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"weather_microservice/internal/apierrors"
	"weather_microservice/internal/contracts"
)

const OPENWEATHER_SERVER_TIMEOUT = 10 * time.Second

type OpenWeatherAdapter struct {
	configApiKey string
}

var OpenWeatherAPIBaseURL = func() string {
	return "https://api.openweathermap.org/data/2.5"
}

func NewOpenWeatherAdapter(apikey string) (OpenWeatherAdapter, error) {
	if apikey == "" {
		return OpenWeatherAdapter{}, fmt.Errorf("OPENWEATHER_API_KEY is not configured")
	}
	return OpenWeatherAdapter{
		configApiKey: apikey,
	}, nil
}

func (a *OpenWeatherAdapter) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	if city == "" {
		return contracts.WeatherData{}, fmt.Errorf("empty city provided")
	}
	qCity := url.QueryEscape(city)
	url := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=metric",
		OpenWeatherAPIBaseURL(), qCity, a.configApiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: OPENWEATHER_SERVER_TIMEOUT,
	}

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
