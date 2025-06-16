package mailer_service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"weatherapi/internal/contracts"
	"weatherapi/internal/services/mailer_service"

	"github.com/stretchr/testify/assert"
)

type MockSender struct {
	LastTo      string
	LastSubject string
	LastBody    string
}

func (m *MockSender) Send(to, subject, body string) error {
	m.LastTo = to
	m.LastSubject = subject
	m.LastBody = body
	return nil
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

func TestSendConfirmationEmail(t *testing.T) {
	tmpDir := t.TempDir()
	err := setupTemplates(tmpDir)
	assert.NoError(t, err)

	mockSender := &MockSender{}
	service := mailer_service.NewMailerService(mockSender, "https://test.com")
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

	mockSender := &MockSender{}
	service := mailer_service.NewMailerService(mockSender, "https://test.com")
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
	tmpDir := t.TempDir()
	os.MkdirAll(tmpDir, 0755)
	err := os.WriteFile(filepath.Join(tmpDir, "confirmation_email.html"), []byte("{{.MissingField}}"), 0644)
	assert.NoError(t, err)

	mockSender := &MockSender{}
	service := mailer_service.NewMailerService(mockSender, "https://test.com")
	service.SetTemplateDir(tmpDir)

	err = service.SendConfirmationEmail(context.Background(), "user@example.com", "badtoken")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute template")
}
