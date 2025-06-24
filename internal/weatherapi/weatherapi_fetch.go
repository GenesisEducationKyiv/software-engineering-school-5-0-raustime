package weatherapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"weatherapi/internal/apierrors"
	"weatherapi/internal/contracts"
)

type WeatherAPIResponse struct {
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

func FetchWeather(city string) (*contracts.WeatherData, error) {
	return FetchWeatherWithContext(context.Background(), city)
}

func FetchWeatherWithContext(ctx context.Context, city string) (*contracts.WeatherData, error) {
	apiKey := os.Getenv("WEATHERAPI_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WEATHERAPI_KEY is not set")
	}

	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather from WeatherAPI: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("warning: failed to close response body: %v", cerr)
		}
	}()

	var weatherResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode WeatherAPI response: %w", err)
	}

	// Check for API error
	if weatherResp.Error.Code != 0 {
		if weatherResp.Error.Code == 1006 { // No matching location found
			return nil, apierrors.ErrCityNotFound
		}
		return nil, fmt.Errorf("WeatherAPI error: %s", weatherResp.Error.Message)
	}

	data := &contracts.WeatherData{
		Description: weatherResp.Current.Condition.Text,
		Temperature: weatherResp.Current.TempC,
		Humidity:    weatherResp.Current.Humidity,
	}

	return data, nil
}
