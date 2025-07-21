// mailer_test.go
package mailer_service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"mailer_microservice/internal/contracts"
	"mailer_microservice/internal/mailer_service"
)

var (
	tmpDir      string
	mockSender  *mailer_service.MockSender
	service     *mailer_service.MailerService
	weatherData contracts.WeatherData
	testBaseURL = "https://test-api.example.com"
)

func TestMain(m *testing.M) {
	// Створюємо тимчасову директорію для шаблонів
	var err error
	tmpDir, err = os.MkdirTemp("", "mailer_test_templates")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp dir: %v", err))
	}

	// Налаштовуємо шаблони
	if err := setupTemplates(tmpDir); err != nil {
		panic(fmt.Sprintf("Failed to setup templates: %v", err))
	}

	// Ініціалізуємо mock sender
	mockSender = mailer_service.NewMockSender()

	// Створюємо сервіс з тестовим base URL
	service = mailer_service.NewMailerService(mockSender, testBaseURL)
	service.SetTemplateDir(tmpDir)

	// Налаштовуємо тестові дані погоди
	weatherData = contracts.WeatherData{
		Description: "Cloudy",
		Temperature: 21.5,
		Humidity:    80,
	}

	// Запускаємо тести
	code := m.Run()

	// Очищаємо після тестів
	_ = os.RemoveAll(tmpDir)

	os.Exit(code)
}

func setupTemplates(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	confirmation := `<!DOCTYPE html>
<html><body><a href="{{.ConfirmURL}}">Confirm</a></body></html>`

	weather := `<html><body><h1>{{.City}}</h1><p>{{.Temperature}}°C</p><p>{{.Description}}</p><a href="{{.UnsubscribeURL}}">Unsubscribe</a></body></html>`

	if err := os.WriteFile(filepath.Join(dir, "confirmation_email.html"), []byte(confirmation), 0644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(dir, "weather_email.html"), []byte(weather), 0644); err != nil {
		return err
	}

	return nil
}

func resetMockSender() {
	mockSender.Reset()
}

func TestSendConfirmationEmail(t *testing.T) {
	resetMockSender()

	err := service.SendConfirmationEmail(context.Background(), "user@example.com", "abc123")

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mockSender.LastTo)
	assert.Equal(t, "Confirm your subscription", mockSender.LastSubject)
	assert.Contains(t, mockSender.LastBody, fmt.Sprintf("%s/api/confirm/abc123", testBaseURL))
}

func TestSendWeatherEmail(t *testing.T) {
	resetMockSender()

	err := service.SendWeatherEmail(context.Background(), "user@example.com", "Kyiv", weatherData, "xyz789")

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mockSender.LastTo)
	assert.Contains(t, mockSender.LastSubject, "Kyiv")
	assert.Contains(t, mockSender.LastBody, weatherData.Description)
	assert.Contains(t, mockSender.LastBody, "21.5")
	assert.Contains(t, mockSender.LastBody, fmt.Sprintf("%s/api/unsubscribe/xyz789", testBaseURL))
}

func TestSendWeatherEmailWithTestUser(t *testing.T) {
	resetMockSender()
	testEmail := "test@example.com"

	err := service.SendWeatherEmail(context.Background(), testEmail, "Kyiv", weatherData, "xyz789")

	assert.NoError(t, err)
	assert.Equal(t, testEmail, mockSender.LastTo)
	assert.Contains(t, mockSender.LastSubject, "Kyiv")
	assert.Contains(t, mockSender.LastBody, weatherData.Description)
	assert.Contains(t, mockSender.LastBody, "21.5")
	assert.Contains(t, mockSender.LastBody, fmt.Sprintf("%s/api/unsubscribe/xyz789", testBaseURL))
}

func TestInvalidTemplateHandling(t *testing.T) {
	resetMockSender()

	invalidTmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(invalidTmpDir, "confirmation_email.html"), []byte("{{.MissingField}}"), 0644)
	assert.NoError(t, err)

	invalidService := mailer_service.NewMailerService(mockSender, testBaseURL)
	invalidService.SetTemplateDir(invalidTmpDir)

	err = invalidService.SendConfirmationEmail(context.Background(), "user@example.com", "badtoken")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render confirmation template")
}

func TestSendMultipleEmails(t *testing.T) {
	resetMockSender()

	emails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for i, email := range emails {
		token := fmt.Sprintf("token-%d", i)
		err := service.SendConfirmationEmail(context.Background(), email, token)
		assert.NoError(t, err)
	}

	assert.Equal(t, len(emails), mockSender.GetSentEmailsCount())

	for _, email := range emails {
		assert.True(t, mockSender.HasEmailBeenSentTo(email))
	}

	lastEmail := mockSender.GetLastSentEmail()
	assert.NotNil(t, lastEmail)
	assert.Equal(t, "user3@example.com", lastEmail.To)
}

func TestMockSenderErrorHandling(t *testing.T) {
	resetMockSender()

	mockSender.SetShouldFail(true)
	mockSender.SetErrorMessage("SMTP server unavailable")

	err := service.SendConfirmationEmail(context.Background(), "user@example.com", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP server unavailable")
	assert.Equal(t, 0, mockSender.GetSentEmailsCount())
}

func TestMockSenderReset(t *testing.T) {
	resetMockSender()

	err := service.SendConfirmationEmail(context.Background(), "user@example.com", "token")
	assert.NoError(t, err)
	assert.Equal(t, 1, mockSender.GetSentEmailsCount())

	mockSender.Clear()
	assert.Equal(t, 0, mockSender.GetSentEmailsCount())
	assert.Empty(t, mockSender.LastTo)

	mockSender.SetShouldFail(true)
	mockSender.Reset()
	assert.False(t, mockSender.ShouldFail)
}

func TestServiceConfiguration(t *testing.T) {
	assert.NotNil(t, service)
	assert.Equal(t, tmpDir, service.TemplateDir)

	sender := service.GetEmailSender()
	assert.NotNil(t, sender)
	assert.IsType(t, &mailer_service.MockSender{}, sender)
}
