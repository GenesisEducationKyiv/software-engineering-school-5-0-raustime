// mailer_test.go
package mailer_service_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"weatherapi/internal/config"
	"weatherapi/internal/contracts"
	"weatherapi/internal/services/mailer_service"

	"github.com/stretchr/testify/assert"
)

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

func createTestConfig() *config.Config {
	return &config.Config{
		AppBaseURL:   "https://test.com",
		Environment:  "test",
		SMTPHost:     "test-smtp.com",
		SMTPPort:     587,
		SMTPUser:     "test@example.com",
		SMTPPassword: "testpass",
	}
}

func TestSendConfirmationEmail(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	cfg := createTestConfig()
	service := mailer_service.NewMailerService(cfg)
	service.SetTemplateDir(tmpDir)

	// Отримуємо mock sender для перевірки
	mockSender, ok := service.GetEmailSender().(*mailer_service.MockSender)
	assert.True(t, ok, "Expected MockSender in test environment")

	err = service.SendConfirmationEmail(context.Background(), "user@example.com", "abc123")

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mockSender.LastTo)
	assert.Equal(t, "Confirm your subscription", mockSender.LastSubject)
	assert.Contains(t, mockSender.LastBody, "https://test.com/api/confirm/abc123")
}

func TestSendConfirmationEmailWithCustomSender(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	mockSender := mailer_service.NewMockSender()
	service := mailer_service.NewMailerServiceWithSender(mockSender, "https://test.com")
	service.SetTemplateDir(tmpDir)

	err = service.SendConfirmationEmail(context.Background(), "user@example.com", "abc123")

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mockSender.LastTo)
	assert.Equal(t, "Confirm your subscription", mockSender.LastSubject)
	assert.Contains(t, mockSender.LastBody, "https://test.com/api/confirm/abc123")
}

func TestSendWeatherEmail(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	cfg := createTestConfig()
	service := mailer_service.NewMailerService(cfg)
	service.SetTemplateDir(tmpDir)

	// Отримуємо mock sender для перевірки
	mockSender, ok := service.GetEmailSender().(*mailer_service.MockSender)
	assert.True(t, ok, "Expected MockSender in test environment")

	data := contracts.WeatherData{
		Description: "Cloudy",
		Temperature: 21.5,
		Humidity:    80,
	}

	err = service.SendWeatherEmail(context.Background(), "user@example.com", "Kyiv", data, "xyz789")
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mockSender.LastTo)
	assert.Contains(t, mockSender.LastSubject, "Kyiv")
	assert.Contains(t, mockSender.LastBody, "Cloudy")
	assert.Contains(t, mockSender.LastBody, "21.5")
	assert.Contains(t, mockSender.LastBody, "https://test.com/api/unsubscribe/xyz789")
}

func TestSendWeatherEmailWithCustomSender(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	mockSender := mailer_service.NewMockSender()
	service := mailer_service.NewMailerServiceWithSender(mockSender, "https://test.com")
	service.SetTemplateDir(tmpDir)

	data := contracts.WeatherData{
		Description: "Cloudy",
		Temperature: 21.5,
		Humidity:    80,
	}

	err = service.SendWeatherEmail(context.Background(), "user@example.com", "Kyiv", data, "xyz789")
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mockSender.LastTo)
	assert.Contains(t, mockSender.LastSubject, "Kyiv")
	assert.Contains(t, mockSender.LastBody, "Cloudy")
	assert.Contains(t, mockSender.LastBody, "21.5")
	assert.Contains(t, mockSender.LastBody, "https://test.com/api/unsubscribe/xyz789")
}

func TestInvalidTemplateHandling(t *testing.T) {
	// Disable logs for this test
	os.Setenv("DISABLE_TEST_LOGS", "1")
	defer os.Unsetenv("DISABLE_TEST_LOGS")

	tmpDir := t.TempDir()
	os.MkdirAll(tmpDir, 0755)
	err := os.WriteFile(filepath.Join(tmpDir, "confirmation_email.html"), []byte("{{.MissingField}}"), 0644)
	assert.NoError(t, err)

	cfg := createTestConfig()
	service := mailer_service.NewMailerService(cfg)
	service.SetTemplateDir(tmpDir)

	err = service.SendConfirmationEmail(context.Background(), "user@example.com", "badtoken")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render confirmation template")
}

func TestConfigBasedMailerService(t *testing.T) {
	cfg := createTestConfig()
	service := mailer_service.NewMailerService(cfg)

	// Перевіряємо що в тестовому середовищі використовується MockSender
	_, ok := service.GetEmailSender().(*mailer_service.MockSender)
	assert.True(t, ok, "Expected MockSender in test environment")

	// Тестуємо з production config
	prodCfg := &config.Config{
		AppBaseURL:   "https://prod.com",
		Environment:  "production",
		SMTPHost:     "smtp.gmail.com",
		SMTPPort:     587,
		SMTPUser:     "prod@example.com",
		SMTPPassword: "prodpass",
	}
	prodService := mailer_service.NewMailerService(prodCfg)

	// В продакшені має бути SMTP sender
	_, ok = prodService.GetEmailSender().(*mailer_service.MockSender)
	assert.False(t, ok, "Expected SMTP sender in production environment")
}

// Додаткові тести з використанням розширеного функціоналу MockSender
func TestSendMultipleEmails(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	mockSender := mailer_service.NewMockSender()
	service := mailer_service.NewMailerServiceWithSender(mockSender, "https://test.com")
	service.SetTemplateDir(tmpDir)

	// Відправляємо кілька emails
	emails := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
	for i, email := range emails {
		token := fmt.Sprintf("token-%d", i)
		err := service.SendConfirmationEmail(context.Background(), email, token)
		assert.NoError(t, err)
	}

	// Перевіряємо загальну кількість
	assert.Equal(t, len(emails), mockSender.GetSentEmailsCount())

	// Перевіряємо що всі emails були відправлені
	for _, email := range emails {
		assert.True(t, mockSender.HasEmailBeenSentTo(email))
	}

	// Перевіряємо останній відправлений email
	lastEmail := mockSender.GetLastSentEmail()
	assert.NotNil(t, lastEmail)
	assert.Equal(t, "user3@example.com", lastEmail.To)
}

func TestMockSenderErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	mockSender := mailer_service.NewMockSender()
	mockSender.SetShouldFail(true)
	mockSender.SetErrorMessage("SMTP server unavailable")

	service := mailer_service.NewMailerServiceWithSender(mockSender, "https://test.com")
	service.SetTemplateDir(tmpDir)

	err = service.SendConfirmationEmail(context.Background(), "user@example.com", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP server unavailable")

	// Перевіряємо що email не був "відправлений"
	assert.Equal(t, 0, mockSender.GetSentEmailsCount())
}

func TestMockSenderReset(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	mockSender := mailer_service.NewMockSender()
	service := mailer_service.NewMailerServiceWithSender(mockSender, "https://test.com")
	service.SetTemplateDir(tmpDir)

	// Відправляємо email
	err = service.SendConfirmationEmail(context.Background(), "user@example.com", "token")
	assert.NoError(t, err)
	assert.Equal(t, 1, mockSender.GetSentEmailsCount())

	// Очищаємо історію
	mockSender.Clear()
	assert.Equal(t, 0, mockSender.GetSentEmailsCount())
	assert.Empty(t, mockSender.LastTo)

	// Скидаємо повністю
	mockSender.SetShouldFail(true)
	mockSender.Reset()
	assert.False(t, mockSender.ShouldFail)
}

func TestDifferentEnvironmentConfigs(t *testing.T) {
	// Тест для різних середовищ
	environments := []struct {
		name        string
		environment string
		expectMock  bool
	}{
		{"test", "test", true},
		{"development", "development", false},
		{"production", "production", false},
	}

	for _, env := range environments {
		t.Run(env.name, func(t *testing.T) {
			cfg := &config.Config{
				AppBaseURL:   "https://example.com",
				Environment:  env.environment,
				SMTPHost:     "smtp.example.com",
				SMTPPort:     587,
				SMTPUser:     "test@example.com",
				SMTPPassword: "password",
			}

			service := mailer_service.NewMailerService(cfg)
			_, isMock := service.GetEmailSender().(*mailer_service.MockSender)

			if env.expectMock {
				assert.True(t, isMock, "Expected MockSender for %s environment", env.environment)
			} else {
				assert.False(t, isMock, "Expected SMTP sender for %s environment", env.environment)
			}
		})
	}
}
