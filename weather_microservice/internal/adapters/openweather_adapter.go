package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"weather_microservice/internal/apierrors"
	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/metrics"
)

var logSourceOpenWeather = "adapter:OpenWeather"

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
		return fail(ctx, "openweather", city, "invalid input", fmt.Errorf("empty city"))
	}

	metrics.WeatherRequests.WithLabelValues("openweather", city).Inc()

	url := fmt.Sprintf("%s/weather?q=%s&appid=%s&units=metric",
		a.configApiBaseURL, url.QueryEscape(city), a.configApiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fail(ctx, "openweather", city, "failed request", fmt.Errorf("failed to create request: %w", err))
	}

	client := &http.Client{
		Timeout: a.configApiServerTimeout,
	}

	resp, err := client.Do(req)

	if err != nil {
		return fail(ctx, "openweather", city, "failed api request", fmt.Errorf("failed to get weather: %w", err))
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logging.Warn(ctx, logSourceOpenWeather, nil, fmt.Errorf("failed to close response body: %w", cerr))
		}
	}()

	// Handle 404 (city not found)
	if resp.StatusCode == http.StatusNotFound {
		var errResp struct {
			Cod     string `json:"cod"`
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)

		if strings.Contains(strings.ToLower(errResp.Message), "city not found") {
			metrics.WeatherFailures.WithLabelValues("openweather", city).Inc()
			logging.Warn(ctx, logSourceOpenWeather, nil, apierrors.ErrCityNotFound)
			return contracts.WeatherData{}, apierrors.ErrCityNotFound
		}

		return fail(ctx, "openweather", city, "failed api request", fmt.Errorf("weather API 404: %s", errResp.Message))
	}

	if resp.StatusCode != http.StatusOK {
		return fail(ctx, "openweather", city, "failed api request", fmt.Errorf("weather API returned status %d", resp.StatusCode))
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
		return fail(ctx, "openweather", city, "failed api request", fmt.Errorf("failed to decode weather response: %w", err))
	}

	if len(weatherResp.Weather) == 0 {
		return fail(ctx, "openweather", city, "failed api request", apierrors.ErrNoWeatherDataFound)
	}

	data := contracts.WeatherData{
		Temperature: weatherResp.Main.Temp,
		Humidity:    weatherResp.Main.Humidity,
		Description: weatherResp.Weather[0].Description,
	}
	logging.Info(ctx, logSourceOpenWeather, data)
	return data, nil
}
