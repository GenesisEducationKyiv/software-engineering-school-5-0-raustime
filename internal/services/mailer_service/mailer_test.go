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

var (
	cfg         *config.Config
	tmpDir      string
	mockSender  *mailer_service.MockSender
	service     mailer_service.MailerService
	weatherData contracts.WeatherData
)

func TestMain(m *testing.M) {
	// Налаштування спільних змінних для всіх тестів
	var err error
	cfg, err = config.LoadTestConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load test config: %v", err))
	}

	// Створюємо тимчасову директорію для шаблонів
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

	// Створюємо сервіс
	service = mailer_service.NewMailerService(mockSender, cfg.AppBaseURL)
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
	os.RemoveAll(tmpDir)

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
	assert.Contains(t, mockSender.LastBody, fmt.Sprintf("%s/api/confirm/abc123", cfg.AppBaseURL))
}

func TestSendConfirmationEmailWithTestSMTPUser(t *testing.T) {
	resetMockSender()

	err := service.SendConfirmationEmail(context.Background(), cfg.SMTPUser, "abc123")

	assert.NoError(t, err)
	assert.Equal(t, cfg.SMTPUser, mockSender.LastTo)
	assert.Equal(t, "Confirm your subscription", mockSender.LastSubject)
	assert.Contains(t, mockSender.LastBody, fmt.Sprintf("%s/api/confirm/abc123", cfg.AppBaseURL))
}

func TestSendWeatherEmail(t *testing.T) {
	resetMockSender()

	err := service.SendWeatherEmail(context.Background(), "user@example.com", "Kyiv", weatherData, "xyz789")

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mockSender.LastTo)
	assert.Contains(t, mockSender.LastSubject, "Kyiv")
	assert.Contains(t, mockSender.LastBody, weatherData.Description)
	assert.Contains(t, mockSender.LastBody, "21.5")
	assert.Contains(t, mockSender.LastBody, fmt.Sprintf("%s/api/unsubscribe/xyz789", cfg.AppBaseURL))
}

func TestSendWeatherEmailWithTestSMTPUser(t *testing.T) {
	resetMockSender()

	err := service.SendWeatherEmail(context.Background(), cfg.SMTPUser, "Kyiv", weatherData, "xyz789")

	assert.NoError(t, err)
	assert.Equal(t, cfg.SMTPUser, mockSender.LastTo)
	assert.Contains(t, mockSender.LastSubject, "Kyiv")
	assert.Contains(t, mockSender.LastBody, weatherData.Description)
	assert.Contains(t, mockSender.LastBody, "21.5")
	assert.Contains(t, mockSender.LastBody, fmt.Sprintf("%s/api/unsubscribe/xyz789", cfg.AppBaseURL))
}

func TestInvalidTemplateHandling(t *testing.T) {
	resetMockSender()

	// Створюємо окрему директорію з некоректним шаблоном
	invalidTmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(invalidTmpDir, "confirmation_email.html"), []byte("{{.MissingField}}"), 0644)
	assert.NoError(t, err)

	// Створюємо окремий сервіс з некоректним шаблоном
	invalidService := mailer_service.NewMailerService(mockSender, cfg.AppBaseURL)
	invalidService.SetTemplateDir(invalidTmpDir)

	err = invalidService.SendConfirmationEmail(context.Background(), "user@example.com", "badtoken")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render confirmation template")
}

func TestSendMultipleEmails(t *testing.T) {
	resetMockSender()

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
	resetMockSender()

	mockSender.SetShouldFail(true)
	mockSender.SetErrorMessage("SMTP server unavailable")

	err := service.SendConfirmationEmail(context.Background(), "user@example.com", "token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SMTP server unavailable")

	// Перевіряємо що email не був "відправлений"
	assert.Equal(t, 0, mockSender.GetSentEmailsCount())
}

func TestMockSenderReset(t *testing.T) {
	resetMockSender()

	// Відправляємо email
	err := service.SendConfirmationEmail(context.Background(), "user@example.com", "token")
	assert.NoError(t, err)
	assert.Equal(t, 1, mockSender.GetSentEmailsCount())

	// Очищаємо історію
	mockSender.Clear()
	assert.Equal(t, 0, mockSender.GetSentEmailsCount())
	assert.Empty(t, mockSender.LastTo)

	// Скидаємо повністю та налаштовуємо помилку
	mockSender.SetShouldFail(true)
	mockSender.Reset()
	assert.False(t, mockSender.ShouldFail)
}

func TestConfigValues(t *testing.T) {
	// Тест для перевірки що конфігурація завантажується правильно
	assert.Equal(t, "https://test.com", cfg.AppBaseURL)
	assert.Equal(t, "test@example.com", cfg.SMTPUser)
	assert.Equal(t, "test-smtp.com", cfg.SMTPHost)
	assert.Equal(t, 587, cfg.SMTPPort)
}

func TestServiceConfiguration(t *testing.T) {
	// Тест для перевірки правильного налаштування сервісу
	assert.NotNil(t, service)
	assert.Equal(t, tmpDir, service.TemplateDir)

	// Перевіряємо що sender правильно встановлений
	sender := service.GetEmailSender()
	assert.NotNil(t, sender)
	assert.IsType(t, &mailer_service.MockSender{}, sender)
}
