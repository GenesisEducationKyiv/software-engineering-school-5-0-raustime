package weatherapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"weatherapi/internal/apierrors"
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

func FetchWeather(city string) (WeatherAPIResponse, error) {
	return FetchWeatherWithContext(context.Background(), city)
}

func FetchWeatherWithContext(ctx context.Context, city string) (WeatherAPIResponse, error) {
	apiKey := os.Getenv("WEATHERAPI_KEY")
	if apiKey == "" {
		return WeatherAPIResponse{}, fmt.Errorf("WEATHERAPI_KEY is not set")
	}

	url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, city)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return WeatherAPIResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return WeatherAPIResponse{}, fmt.Errorf("failed to get weather from WeatherAPI: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("warning: failed to close response body: %v", cerr)
		}
	}()

	var weatherResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return WeatherAPIResponse{}, fmt.Errorf("failed to decode WeatherAPI response: %w", err)
	}

	// Check for API error
	if weatherResp.Error.Code != 0 {
		if weatherResp.Error.Code == 1006 { // No matching location found
			return WeatherAPIResponse{}, apierrors.ErrCityNotFound
		}
		return WeatherAPIResponse{}, fmt.Errorf("WeatherAPI error: %s", weatherResp.Error.Message)
	}

	return weatherResp, nil
}
