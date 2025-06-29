package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"weatherapi/internal/contracts"

	"weatherapi/internal/db/models"

	"github.com/stretchr/testify/assert"
)

// MockMailerService — мок реалізації MailerServiceProvider
type MockMailerService struct{}

func (m *MockMailerService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	return nil
}

func (m *MockMailerService) SendWeatherEmail(ctx context.Context, email, city string, weather contracts.WeatherData, token string) error {
	return nil
}

func TestSubscriptionFlow(t *testing.T) {

	// Change to project root
	originalDir, _ := os.Getwd()
	_ = os.Chdir("../..")
	defer func() {
		_ = os.Chdir(originalDir) // Restore original directory
		cleanupTestData()
	}()

	email := "test@example.com"

	t.Run("Subscribe", func(t *testing.T) {
		cleanupTestData()

		payload := map[string]string{
			"email":     email,
			"city":      "Kyiv",
			"frequency": "daily",
		}
		jsonData, _ := json.Marshal(payload)

		resp, err := http.Post(testServer.URL+"/api/subscribe", "application/json", bytes.NewBuffer(jsonData))
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Just check the status code - no JSON response expected
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Optionally verify the subscription was created in the database
		count, err := container.DB.NewSelect().
			Model((*models.Subscription)(nil)).
			Where("email = ?", email).
			Count(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "Subscription should be created in database")
	})

	// 2. Отримання токену з БД
	var token string
	err := container.DB.NewSelect().
		Model((*models.Subscription)(nil)).
		Column("token").
		Where("email = ?", email).
		Scan(context.Background(), &token)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 3. Підтвердження підписки
	t.Run("Confirm", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/api/confirm/" + token)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Just check the status code - no JSON response expected
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the subscription is confirmed in the database
		var subscription models.Subscription
		err = container.DB.NewSelect().
			Model(&subscription).
			Column("confirmed").
			Where("email = ?", email).
			Scan(context.Background())
		assert.NoError(t, err)
		assert.True(t, subscription.Confirmed)
	})

	// 4. Відписка
	t.Run("Unsubscribe", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/api/unsubscribe/" + token)
		assert.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Just check the status code - no JSON response expected
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify the subscription is deleted from the database
		count, err := container.DB.NewSelect().
			Model((*models.Subscription)(nil)).
			Where("email = ?", email).
			Count(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

}

func TestInvalidSubscriptionRequests(t *testing.T) {
	defer cleanupTestData()

	tests := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{"Empty email", map[string]string{"email": "", "city": "Kyiv", "frequency": "daily"}, http.StatusBadRequest},
		{"Invalid email format", map[string]string{"email": "invalid-email", "city": "Kyiv", "frequency": "daily"}, http.StatusBadRequest},
		{"Missing email field", map[string]string{"city": "Kyiv", "frequency": "daily"}, http.StatusBadRequest},
		{"Missing city field", map[string]string{"email": "test@example.com", "frequency": "daily"}, http.StatusBadRequest},
		{"Missing frequency field", map[string]string{"email": "test@example.com", "city": "Kyiv"}, http.StatusBadRequest},
		{"Invalid frequency", map[string]string{"email": "test@example.com", "city": "Kyiv", "frequency": "weekly"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			resp, err := http.Post(testServer.URL+"/api/subscribe", "application/json", bytes.NewBuffer(jsonData))
			assert.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			body, _ := io.ReadAll(resp.Body)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Response: %s", string(body))
		})
	}
}

func cleanupTestData() {
	if container != nil && container.DB != nil {
		_, err := container.DB.NewDelete().
			Table("subscriptions").
			Where("email LIKE ? OR email LIKE ?", "%@example.com", "%test%").
			Exec(context.Background())
		if err != nil {
			println("Cleanup error:", err.Error())
		}
	}
}
