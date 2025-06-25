package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
)

const WEATHER_SERVER_TIMEOUT = 10 * time.Second

type WeatherAPIAdapter struct {
	configApiKey string
}

var WeatherAPIBaseURL = func() string {
	return "https://api.weatherapi.com/v1"
}

func NewWeatherAPIAdapter(apikey string) *WeatherAPIAdapter {
	return &WeatherAPIAdapter{
		configApiKey: apikey,
	}
}

func (a *WeatherAPIAdapter) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	if a.configApiKey == "" {
		return contracts.WeatherData{}, fmt.Errorf("WEATHER_API_KEY is not configured")
	}
	if city == "" {
		return contracts.WeatherData{}, fmt.Errorf("empty city provided")
	}
	qCity := url.QueryEscape(city)
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", WeatherAPIBaseURL(), a.configApiKey, qCity)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: WEATHER_SERVER_TIMEOUT,
	}

	resp, err := client.Do(req)

	if err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to get weather from WeatherAPI: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("WeatherAPI status: %s (%d)", resp.Status, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to read WeatherAPI response body: %w", err)
	}

	log.Printf("WeatherAPI raw response: %s", string(body))

	if len(body) == 0 {
		return contracts.WeatherData{}, fmt.Errorf("empty response body from WeatherAPI")
	}

	var weatherResp struct {
		Current struct {
			TempC     float64 `json:"temp_c"`
			Humidity  float64 `json:"humidity"`
			Condition struct {
				Text string `json:"text"`
			} `json:"condition"`
		} `json:"current"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return contracts.WeatherData{}, fmt.Errorf("failed to decode WeatherAPI response: %w", err)
	}

	// Check for API error
	if weatherResp.Error.Code != 0 {
		if weatherResp.Error.Code == 1006 { // No matching location found
			return contracts.WeatherData{}, apierrors.ErrCityNotFound
		}
		return contracts.WeatherData{}, fmt.Errorf("WeatherAPI error: %s", weatherResp.Error.Message)
	}

	// Convert to internal format
	return contracts.WeatherData{
		Temperature: weatherResp.Current.TempC,
		Humidity:    weatherResp.Current.Humidity,
		Description: weatherResp.Current.Condition.Text,
	}, nil
}
