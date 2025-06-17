package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSubscriptionFlow(t *testing.T) {
	defer cleanupTestData()

	email := "test@example.com"

	// 1. Підписка
	t.Run("Subscribe", func(t *testing.T) {
		cleanupTestData() // очистка перед новою підпискою
		payload := map[string]string{
			"email":     email,
			"city":      "Kyiv",
			"frequency": "daily",
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := http.Post(
			testServer.URL+"/api/subscribe",
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			t.Fatalf("Failed to make subscribe request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Read the response body for debugging
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected status 200, got %d. Response: %s", resp.StatusCode, string(body))
			return
		}

		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if message, ok := response["message"]; !ok || !strings.Contains(message.(string), "confirmation") {
			t.Error("Expected confirmation message in response")
		}
	})

	// 2. Отримання токену з БД для тестування
	var token string
	// Use Bun ORM syntax instead of raw SQL
	err := container.DB.NewSelect().
		Column("token").
		Table("subscriptions").
		Where("email = ?", email).
		Scan(context.Background(), &token)
	if err != nil {
		t.Fatalf("Failed to get token from database: %v", err)
	}

	// 3. Підтвердження підписки
	t.Run("Confirm", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/api/confirm/" + token)
		if err != nil {
			t.Fatalf("Failed to make confirm request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var subscription struct {
			Confirmed bool `bun:"confirmed"`
		}
		err = container.DB.NewSelect().
			Model(&subscription).
			Column("confirmed").
			Where("email = ?", email).
			Scan(context.Background())
		if err != nil {
			t.Fatalf("Failed to check confirmation status: %v", err)
		}

		if !subscription.Confirmed {
			t.Error("Subscription should be confirmed")
		}
	})

	// 4. Відписка
	t.Run("Unsubscribe", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/api/unsubscribe/" + token)
		if err != nil {
			t.Fatalf("Failed to make unsubscribe request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Перевірка, що підписка видалена з БД
		var count int
		err = container.DB.NewSelect().
			Column("count(*)").
			Table("subscriptions").
			Where("email = ?", email).
			Scan(context.Background(), &count)
		if err != nil {
			t.Fatalf("Failed to check subscription deletion: %v", err)
		}

		if count != 0 {
			t.Error("Subscription should be deleted")
		}
	})
}

func TestInvalidSubscriptionRequests(t *testing.T) {
	defer cleanupTestData()

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name:           "Empty email",
			payload:        map[string]string{"email": "", "city": "Kyiv", "frequency": "daily"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid email format",
			payload:        map[string]string{"email": "invalid-email", "city": "Kyiv", "frequency": "daily"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing email field",
			payload:        map[string]string{"city": "Kyiv", "frequency": "daily"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing city field",
			payload:        map[string]string{"email": "test@example.com", "frequency": "daily"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing frequency field",
			payload:        map[string]string{"email": "test@example.com", "city": "Kyiv"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid frequency",
			payload:        map[string]string{"email": "test@example.com", "city": "Kyiv", "frequency": "weekly"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)

			resp, err := http.Post(
				testServer.URL+"/api/subscribe",
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				// Read response body for debugging
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

// Add cleanup function if it doesn't exist
func cleanupTestData() {
	if container != nil && container.DB != nil {
		// Clean up test data using Bun syntax
		_, err := container.DB.NewDelete().
			Table("subscriptions").
			Where("email LIKE ? OR email LIKE ?", "%@example.com", "%test%").
			Exec(context.Background())
		if err != nil {
			// Log error but don't fail the test
			println("Cleanup error:", err.Error())
		}
	}
}
