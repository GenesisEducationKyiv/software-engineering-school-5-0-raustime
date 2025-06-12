package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"weatherapi/internal/services"
)

// Mock Subscription Service
type mockSubscriptionService struct {
	subscription *services.Subscription
	err          error
}

func (m *mockSubscriptionService) CreateSubscription(ctx context.Context, email, city, frequency string) (*services.Subscription, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.subscription, nil
}

func (m *mockSubscriptionService) ConfirmSubscription(ctx context.Context, token string) error {
	return m.err
}

func (m *mockSubscriptionService) DeleteSubscription(ctx context.Context, token string) error {
	return m.err
}

// Mock Mailer Service
type mockMailerService struct {
	err error
}

func (m *mockMailerService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	return m.err
}

func TestSubscriptionHandler_Subscribe(t *testing.T) {
	tests := []struct {
		name                string
		requestBody         interface{}
		mockSubscription    *services.Subscription
		mockSubscribeError  error
		mockMailerError     error
		expectedStatus      int
		expectedBody        string
	}{
		{
			name: "successful subscription",
			requestBody: SubscriptionRequest{
				Email:     "test@example.com",
				City:      "Kyiv",
				Frequency: "daily",
			},
			mockSubscription: &services.Subscription{
				ID:        1,
				Email:     "test@example.com",
				Token:     "test-token",
				Confirmed: false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body\n",
		},
		{
			name: "missing email",
			requestBody: SubscriptionRequest{
				City:      "Kyiv",
				Frequency: "daily",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid email\n",
		},
		{
			name: "missing city",
			requestBody: SubscriptionRequest{
				Email:     "test@example.com",
				Frequency: "daily",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid city\n",
		},
		{
			name: "invalid frequency",
			requestBody: SubscriptionRequest{
				Email:     "test@example.com",
				City:      "Kyiv",
				Frequency: "weekly",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid frequency\n",
		},
		{
			name: "already subscribed",
			requestBody: SubscriptionRequest{
				Email:     "test@example.com",
				City:      "Kyiv",
				Frequency: "daily",
			},
			mockSubscribeError: services.ErrAlreadySubscribed,
			expectedStatus:     http.StatusConflict,
			expectedBody:       "Email already subscribed\n",
		},
		{
			name: "subscription service error",
			requestBody: SubscriptionRequest{
				Email:     "test@example.com",
				City:      "Kyiv",
				Frequency: "daily",
			},
			mockSubscribeError: errors.New("service error"),
			expectedStatus:     http.StatusInternalServerError,
			expectedBody:       "Internal server error\n",
		},
		{
			name: "mailer service error",
			requestBody: SubscriptionRequest{
				Email:     "test@example.com",
				City:      "Kyiv",
				Frequency: "daily",
			},
			mockSubscription: &services.Subscription{
				ID:        1,
				Email:     "test@example.com",
				Token:     "test-token",
				Confirmed: false,
			},
			mockMailerError: errors.New("mailer error"),
			expectedStatus:  http.StatusInternalServerError,
			expectedBody:    "Failed to send confirmation email\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock services
			mockSubService := &mockSubscriptionService{
				subscription: tt.mockSubscription,
				err:          tt.mockSubscribeError,
			}
			mockMailService := &mockMailerService{
				err: tt.mockMailerError,
			}

			// Create handler
			handler := NewSubscriptionHandler(mockSubService, mockMailService)

			// Create request body
			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/subscribe", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			handler.Subscribe(w, req)

			// Assert status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Assert response body if expected
			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				expectedBody := strings.TrimSpace(tt.expectedBody)
				if body != expectedBody {
					t.Errorf("expected body %q, got %q", expectedBody, body)
				}
			}
		})
	}
}

func TestSubscriptionHandler_Confirm(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful confirmation",
			path:           "/api/confirm/test-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			path:           "/api/confirm/",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Token is required\n",
		},
		{
			name:           "subscription not found",
			path:           "/api/confirm/invalid-token",
			mockError:      services.ErrSubscriptionNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Subscription not found\n",
		},
		{
			name:           "invalid token",
			path:           "/api/confirm/invalid-token",
			mockError:      services.ErrInvalidToken,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid token\n",
		},
		{
			name:           "internal server error",
			path:           "/api/confirm/test-token",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock services
			mockSubService := &mockSubscriptionService{err: tt.mockError}
			mockMailService := &mockMailerService{}

			// Create handler
			handler := NewSubscriptionHandler(mockSubService, mockMailService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			// Execute
			handler.Confirm(w, req)

			// Assert status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Assert response body if expected
			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				expectedBody := strings.TrimSpace(tt.expectedBody)
				if body != expectedBody {
					t.Errorf("expected body %q, got %q", expectedBody, body)
				}
			}
		})
	}
}

func TestSubscriptionHandler_Unsubscribe(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful unsubscribe",
			path:           "/api/unsubscribe/test-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			path:           "/api/unsubscribe/",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Token is required\n",
		},
		{
			name:           "subscription not found",
			path:           "/api/unsubscribe/invalid-token",
			mockError:      services.ErrSubscriptionNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Subscription not found\n",
		},
		{
			name:           "invalid token",
			path:           "/api/unsubscribe/invalid-token",
			mockError:      services.ErrInvalidToken,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid token\n",
		},
		{
			name:           "internal server error",
			path:           "/api/unsubscribe/test-token",
			mockError:      errors.New("service error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock services
			mockSubService := &mockSubscriptionService{err: tt.mockError}
			mockMailService := &mockMailerService{}

			// Create handler
			handler := NewSubscriptionHandler(mockSubService, mockMailService)

			// Create request
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			// Execute
			handler.Unsubscribe(w, req)

			// Assert status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Assert response body if expected
			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				expectedBody := strings.TrimSpace(tt.expectedBody)
				if body != expectedBody {
					t.Errorf("expected body %q, got %q", expectedBody, body)
				}
			}
		})
	}
}

func TestSubscriptionHandler_validateSubscriptionRequest(t *testing.T) {
	handler := &SubscriptionHandler{}

	tests := []struct {
		name        string
		request     SubscriptionRequest
		expectedErr error
	}{
		{
			name: "valid request",
			request: SubscriptionRequest{
				Email:     "test@example.com",
				City:      "Kyiv",
				Frequency: "daily",
			},
			expectedErr: nil,
		},
		{
			name: "missing email",
			request: SubscriptionRequest{
				City:      "Kyiv",
				Frequency: "daily",
			},
			expectedErr: ErrInvalidEmail,
		},
		{
			name: "missing city",
			request: SubscriptionRequest{
				Email:     "test@example.com",
				Frequency: "daily",
			},
			expectedErr: ErrInvalidCity,
		},
		{
			name: "invalid frequency",
			request: SubscriptionRequest{
				Email:     "test@example.com",
				City:      "Kyiv",
				Frequency: "weekly",
			},
			expectedErr: ErrInvalidFrequency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateSubscriptionRequest(tt.request)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}