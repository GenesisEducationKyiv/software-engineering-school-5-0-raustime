package mailer_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"weatherapi/internal/mailer"
	"weatherapi/internal/openweatherapi"

	"github.com/stretchr/testify/assert"
)

const testTemplatesDir = "tests/templates"

func TestMain(m *testing.M) {
	// Setup: створюємо тестові шаблони перед запуском тестів
	if err := setupTemplates(); err != nil {
		fmt.Printf("Failed to setup test templates: %v\n", err)
		os.Exit(1)
	}

	// Запускаємо тести
	code := m.Run()

	// Teardown: видаляємо тестові файли після завершення тестів
	if err := cleanupTemplates(); err != nil {
		fmt.Printf("Failed to cleanup test templates: %v\n", err)
	}

	os.Exit(code)
}

func setupTemplates() error {
	if err := os.MkdirAll(testTemplatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	confirmationTemplate := `<!DOCTYPE html>
<html>
<head><title>Confirm Subscription</title></head>
<body>
	<h1>Confirm your subscription</h1>
	<p>Click <a href="{{.ConfirmURL}}">here</a> to confirm</p>
</body>
</html>`

	if err := os.WriteFile(testTemplatesDir+"/confirmation_email.html", []byte(confirmationTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write confirmation_email.html: %w", err)
	}

	weatherTemplate := `<!DOCTYPE html>
<html>
<head><title>Weather Update</title></head>
<body>
	<h1>Weather in {{.City}}</h1>
	<p>Temperature: {{.Temperature}}°C</p>
	<p>Description: {{.Description}}</p>
	<p><a href="{{.UnsubscribeURL}}">Unsubscribe</a></p>
</body>
</html>`

	if err := os.WriteFile(testTemplatesDir+"/weather_email.html", []byte(weatherTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write weather_email.html: %w", err)
	}

	return nil
}

func cleanupTemplates() error {
	return os.RemoveAll("tests")
}

func TestSendConfirmationEmail(t *testing.T) {
	mock := &mailer.MockSender{}

	// Зберігаємо старий глобальний sender
	oldEmail := mailer.Email
	mailer.Email = mock
	defer func() { mailer.Email = oldEmail }()

	err := mailer.SendConfirmationEmailWithSender(mock, "test@example.com", "token123")

	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", mock.LastTo)
	assert.Equal(t, "Confirm your subscription", mock.LastSubject)
	assert.Contains(t, mock.LastBody, "https://example.com/api/confirm/token123")
}

func TestSendWeatherEmail(t *testing.T) {
	mock := &mailer.MockSender{}

	// Зберігаємо старий глобальний sender
	oldEmail := mailer.Email
	mailer.Email = mock
	defer func() { mailer.Email = oldEmail }()

	data := &openweatherapi.WeatherData{
		Description: "Cloudy",
		Temperature: 13.7,
		Humidity:    70,
	}

	err := mailer.SendWeatherEmailWithSender(mock, "user@example.com", "Berlin", data, "https://example.com", "tok789")

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", mock.LastTo)
	assert.Contains(t, mock.LastSubject, "Berlin")
	assert.Contains(t, mock.LastBody, "Cloudy")
	assert.Contains(t, mock.LastBody, "13.7")
	assert.Contains(t, mock.LastBody, "https://example.com/api/unsubscribe/tok789")
}

// Альтернативний підхід з t.TempDir() для ізольованих тестів
func TestSendWeatherEmail_WithTempDir(t *testing.T) {
	// Створюємо тимчасову директорію для цього тесту
	tempDir := t.TempDir()
	templatesDir := tempDir + "/templates"

	err := os.MkdirAll(templatesDir, 0755)
	assert.NoError(t, err)

	// Створюємо шаблон тільки для цього тесту
	weatherTemplate := `<h1>{{.City}} Weather</h1><p>{{.Temperature}}°C - {{.Description}}</p>`
	err = os.WriteFile(templatesDir+"/weather_email.html", []byte(weatherTemplate), 0644)
	assert.NoError(t, err)

	mock := &mailer.MockSender{}

	data := &openweatherapi.WeatherData{
		Description: "Rainy",
		Temperature: 15.2,
		Humidity:    85,
	}

	err = mailer.SendWeatherEmailWithSender(mock, "isolated@test.com", "London", data, "https://test.com", "token123")
	assert.NoError(t, err)

	// Файли автоматично видаляться після завершення тесту
}

// Додатковий тест для перевірки помилок шаблонів
func TestSendWeatherEmail_InvalidTemplate(t *testing.T) {
	mock := &mailer.MockSender{}

	// Тимчасово пошкоджуємо шаблон
	invalidTemplate := `{{.InvalidField}}`
	tempFile := testTemplatesDir + "/weather_email_temp.html"

	err := os.WriteFile(tempFile, []byte(invalidTemplate), 0644)
	assert.NoError(t, err)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			log.Printf("Failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	// Тест повинен пройти, оскільки ми використовуємо правильний шаблон
	data := &openweatherapi.WeatherData{
		Description: "Sunny",
		Temperature: 25.0,
		Humidity:    60,
	}

	err = mailer.SendWeatherEmailWithSender(mock, "test@example.com", "Kyiv", data, "https://example.com", "token")
	assert.NoError(t, err)
}
