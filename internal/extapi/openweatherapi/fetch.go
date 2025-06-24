package openweatherapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"weatherapi/internal/apierrors"
)

type WeatherResponse struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity float64 `json:"humidity"`
	} `json:"main"`
}

func FetchWeather(city string) (WeatherResponse, error) {
	return FetchWeatherWithContext(context.Background(), city)
}

func FetchWeatherWithContext(ctx context.Context, city string) (WeatherResponse, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		return WeatherResponse{}, fmt.Errorf("OPENWEATHER_API_KEY is not set")
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return WeatherResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return WeatherResponse{}, fmt.Errorf("failed to get weather: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Printf("warning: failed to close response body: %v", cerr)
		}
	}()

	// special case for 404
	if resp.StatusCode == http.StatusNotFound {
		var errResp struct {
			Cod     string `json:"cod"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Message == "city not found" {
			return WeatherResponse{}, apierrors.ErrCityNotFound
		}
		return WeatherResponse{}, fmt.Errorf("weather API 404: %s", errResp.Message)
	}

	if resp.StatusCode != 200 {
		return WeatherResponse{}, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return WeatherResponse{}, fmt.Errorf("failed to decode weather response: %w", err)
	}

	if len(weatherResp.Weather) == 0 {
		return WeatherResponse{}, fmt.Errorf("no weather data found")
	}

	// return  WeatherResponse{
	// 	Description: weatherResp.Weather[0].Description,
	// 	Temperature: weatherResp.Main.Temp,
	// 	Humidity:    weatherResp.Main.Humidity,
	// }

	return weatherResp, nil
}
