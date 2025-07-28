package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"weather_microservice/internal/apierrors"
	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
)

type OpenWeatherAdapter struct {
	configApiKey           string
	configApiBaseURL       string
	configApiServerTimeout time.Duration
}

func NewOpenWeatherAdapter(apikey string, ApiBaseURL string, ApiServerTimeout time.Duration) (OpenWeatherAdapter, error) {
	if apikey == "" {
		return OpenWeatherAdapter{}, fmt.Errorf("OPENWEATHER_API_KEY is not configured")
	}
	return OpenWeatherAdapter{
		configApiKey:           apikey,
		configApiBaseURL:       ApiBaseURL,
		configApiServerTimeout: ApiServerTimeout,
	}, nil
}

func (a *OpenWeatherAdapter) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	if city == "" {
		err := fmt.Errorf("empty city provided")
		logging.Error(ctx, "adapter:OpenWeather", nil, err)
		return contracts.WeatherData{}, err
	}
	qCity := url.QueryEscape(city)
	url := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=metric",
		a.configApiBaseURL, qCity, a.configApiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		logging.Error(ctx, "adapter:OpenWeather", nil, err)
		return contracts.WeatherData{}, err
	}

	client := &http.Client{
		Timeout: a.configApiServerTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get weather: %w", err)
		logging.Error(ctx, "adapter:OpenWeather", nil, err)
		return contracts.WeatherData{}, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logging.Warn(ctx, "adapter:OpenWeather", nil, fmt.Errorf("failed to close response body: %w", cerr))
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
			logging.Warn(ctx, "adapter:OpenWeather", nil, apierrors.ErrCityNotFound)
			return contracts.WeatherData{}, apierrors.ErrCityNotFound
		}
		err := fmt.Errorf("weather API 404: %s", errResp.Message)
		logging.Error(ctx, "adapter:OpenWeather", nil, err)
		return contracts.WeatherData{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("weather API returned status %d", resp.StatusCode)
		logging.Error(ctx, "adapter:OpenWeather", nil, err)
		return contracts.WeatherData{}, err
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
		err = fmt.Errorf("failed to decode weather response: %w", err)
		logging.Error(ctx, "adapter:OpenWeather", nil, err)
		return contracts.WeatherData{}, err
	}

	if len(weatherResp.Weather) == 0 {
		err := fmt.Errorf("no weather data found")
		logging.Error(ctx, "adapter:OpenWeather", nil, err)
		return contracts.WeatherData{}, err
	}

	data := contracts.WeatherData{
		Temperature: weatherResp.Main.Temp,
		Humidity:    weatherResp.Main.Humidity,
		Description: weatherResp.Weather[0].Description,
	}
	logging.Info(ctx, "adapter:OpenWeather", data)
	return data, nil
}
