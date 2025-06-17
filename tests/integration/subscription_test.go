package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"weatherapi/internal/contracts"

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
	defer cleanupTestData()

	email := "test@example.com"

	// створюємо сервіс з моком
	//service := subscription_service.NewSubscriptionService(container.DB, &MockMailerService{})

	// 1. Підписка
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
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		message, ok := response["message"]
		assert.True(t, ok)
		assert.Contains(t, message.(string), "confirmation")
	})

	// 2. Отримання токену з БД
	var token string
	err := container.DB.NewSelect().
		Column("token").
		Table("subscriptions").
		Where("email = ?", email).
		Scan(context.Background(), &token)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 3. Підтвердження підписки
	t.Run("Confirm", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/api/confirm/" + token)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var subscription struct {
			Confirmed bool `bun:"confirmed"`
		}
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
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var count int
		err = container.DB.NewSelect().
			Column("count(*)").
			Table("subscriptions").
			Where("email = ?", email).
			Scan(context.Background(), &count)
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
			defer resp.Body.Close()

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
