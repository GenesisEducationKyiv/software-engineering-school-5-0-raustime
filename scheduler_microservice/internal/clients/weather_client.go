package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"scheduler_microservice/internal/contracts"
	"time"
)

type weatherHttpClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewWeatherHttpClient(baseURL string) *weatherHttpClient {
	return &weatherHttpClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *weatherHttpClient) GetWeather(ctx context.Context, city string) (*contracts.WeatherData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/weather?city=%s", c.baseURL, city), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weather service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather service returned status %d", resp.StatusCode)
	}

	var data contracts.WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &data, nil
}
