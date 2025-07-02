package integration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestGetWeather(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "Valid city request",
			queryParams:    "?city=Kyiv",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"temperature", "humidity", "description"},
		},
		{
			name:           "Missing parameters",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty city parameter",
			queryParams:    "?city=",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := testServer.URL + "/api/weather" + tt.queryParams
			req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(bodyBytes))
			}

			if tt.expectedStatus == http.StatusOK {
				if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				var weatherData map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &weatherData); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				for _, field := range tt.expectedFields {
					if _, exists := weatherData[field]; !exists {
						t.Errorf("Expected field %s not found in response: %v", field, weatherData)
					}
				}
			}
		})
	}
}
