package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"weatherapi/internal/services"
)

// Mock Weather Service
type mockWeatherService struct {
	weather *services.Weather
	err     error
}

func (m *mockWeatherService) GetWeather(ctx context.Context, city string) (*services.Weather, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.weather, nil
}

func TestWeatherHandler_GetWeather(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		queryParams    string
		mockWeather    *services.Weather
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful weather request",
			method:      http.MethodGet,
			queryParams: "?city=Kyiv",
			mockWeather: &services.Weather{
				Temperature: 25.5,
				Humidity:    60.0,
				Description: "Sunny",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"temperature":25.5,"humidity":60,"description":"Sunny"}`,
		},
		{
			name:           "missing city parameter",
			method:         http.MethodGet,
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "City parameter is required\n",
		},
		{
			name:           "city not found",
			method:         http.MethodGet,
			queryParams:    "?city=InvalidCity",
			mockError:      services.ErrCityNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "City not found\n",
		},
		{
			name:           "internal server error",
			method:         http.MethodGet,
			queryParams:    "?city=Kyiv",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			queryParams:    "?city=Kyiv",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock service
			mockService := &mockWeatherService{
				weather: tt.mockWeather,
				err:     tt.mockError,
			}

			// Create handler
			handler := NewWeatherHandler(mockService)

			// Create request
			req := httptest.NewRequest(tt.method, "/api/weather"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			// Execute
			handler.GetWeather(w, req)

			// Assert status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Assert response body
			body := strings.TrimSpace(w.Body.String())
			expectedBody := strings.TrimSpace(tt.expectedBody)
			if body != expectedBody {
				t.Errorf("expected body %q, got %q", expectedBody, body)
			}

			// Assert content type for successful responses
			if tt.expectedStatus == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", contentType)
				}
			}
		})
	}
}