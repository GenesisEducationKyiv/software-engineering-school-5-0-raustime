package openweatherapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

var ErrCityNotFound = errors.New("city not found")

type WeatherResponse struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp     float64 `json:"temp"`
		Humidity float64 `json:"humidity"`
	} `json:"main"`
}

type WeatherData struct {
	Description string
	Temperature float64
	Humidity    float64
}

func FetchWeather(city string) (*WeatherData, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENWEATHER_API_KEY is not set")
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", city, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather: %w", err)
	}
	defer resp.Body.Close()

	// special case for 404
	if resp.StatusCode == http.StatusNotFound {
		var errResp struct {
			Cod     string `json:"cod"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Message == "city not found" {
			return nil, ErrCityNotFound
		}
		return nil, fmt.Errorf("weather API 404: %s", errResp.Message)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	if len(weatherResp.Weather) == 0 {
		return nil, fmt.Errorf("no weather data found")
	}

	data := &WeatherData{
		Description: weatherResp.Weather[0].Description,
		Temperature: weatherResp.Main.Temp,
		Humidity:    weatherResp.Main.Humidity,
	}

	return data, nil
}
