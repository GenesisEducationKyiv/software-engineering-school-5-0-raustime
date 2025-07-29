package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"weather_microservice/internal/apierrors"
	"weather_microservice/internal/contracts"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/metrics"
)

type WeatherAdapter struct {
	configApiKey           string
	configApiBaseURL       string
	configApiServerTimeout time.Duration
}

func NewWeatherAdapter(apikey string, apiBaseURL string, apiServerTimeout time.Duration) (WeatherAdapter, error) {
	if apikey == "" {
		return WeatherAdapter{}, fmt.Errorf("WEATHER_API_KEY is not configured")
	}

	return WeatherAdapter{
		configApiKey:           apikey,
		configApiBaseURL:       apiBaseURL,
		configApiServerTimeout: apiServerTimeout,
	}, nil
}

func (a *WeatherAdapter) FetchWeather(ctx context.Context, city string) (contracts.WeatherData, error) {
	if city == "" {
		err := fmt.Errorf("empty city provided")
		logging.Error(ctx, "adapter:WeatherAPI", nil, err)
		return contracts.WeatherData{}, err
	}

	metrics.WeatherRequests.WithLabelValues("weatherapi", city).Inc()

	qCity := url.QueryEscape(city)
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", a.configApiBaseURL, a.configApiKey, qCity)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		logging.Error(ctx, "adapter:WeatherAPI", nil, err)
		metrics.WeatherFailures.WithLabelValues("weatherapi", city).Inc()
		return contracts.WeatherData{}, err
	}

	client := &http.Client{Timeout: a.configApiServerTimeout}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get weather from WeatherAPI: %w", err)
		logging.Error(ctx, "adapter:WeatherAPI", nil, err)
		metrics.WeatherFailures.WithLabelValues("weatherapi", city).Inc()
		return contracts.WeatherData{}, err
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logging.Warn(ctx, "adapter:WeatherAPI", nil, fmt.Errorf("failed to close response body: %w", cerr))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read WeatherAPI response body: %w", err)
		logging.Error(ctx, "adapter:WeatherAPI", nil, err)
		metrics.WeatherFailures.WithLabelValues("weatherapi", city).Inc()
		return contracts.WeatherData{}, err
	}

	if len(body) == 0 {
		err = fmt.Errorf("empty response body from WeatherAPI")
		logging.Error(ctx, "adapter:WeatherAPI", nil, err)
		metrics.WeatherFailures.WithLabelValues("weatherapi", city).Inc()
		return contracts.WeatherData{}, err
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
		err = fmt.Errorf("failed to decode WeatherAPI response: %w", err)
		logging.Error(ctx, "adapter:WeatherAPI", nil, err)
		metrics.WeatherFailures.WithLabelValues("weatherapi", city).Inc()
		return contracts.WeatherData{}, err
	}

	if weatherResp.Error.Code != 0 {
		var err error
		if weatherResp.Error.Code == 1006 {
			err = apierrors.ErrCityNotFound
			logging.Warn(ctx, "adapter:WeatherAPI", nil, err)
		} else {
			err = fmt.Errorf("WeatherAPI error: %s", weatherResp.Error.Message)
			logging.Error(ctx, "adapter:WeatherAPI", nil, err)
		}
		metrics.WeatherFailures.WithLabelValues("weatherapi", city).Inc()
		return contracts.WeatherData{}, err
	}

	data := contracts.WeatherData{
		Temperature: weatherResp.Current.TempC,
		Humidity:    weatherResp.Current.Humidity,
		Description: weatherResp.Current.Condition.Text,
	}
	logging.Info(ctx, "adapter:WeatherAPI", data)
	return data, nil
}
