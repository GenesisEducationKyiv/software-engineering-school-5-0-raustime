package mailer_test

import (
	"os"
	"testing"
	"weatherapi/internal/mailer"
	"weatherapi/internal/openweatherapi"

	"github.com/stretchr/testify/assert"
)

func setupTemplates() {
	os.MkdirAll("internal/templates", 0755)
	os.WriteFile("internal/templates/confirmation_email.html", []byte(`
		<html><body><a href="{{.ConfirmURL}}">Confirm</a></body></html>
	`), 0644)
	os.WriteFile("internal/templates/weather_email.html", []byte(`
		<html><body>{{.City}} - {{.Temperature}}°C, {{.Description}}.
		<a href="{{.UnsubscribeURL}}">Unsubscribe</a></body></html>
	`), 0644)
}

func TestSendConfirmationEmail(t *testing.T) {
	mock := &mailer.MockSender{}

	// Збережемо старий глобальний sender, щоб відновити після тесту
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

	// Збережемо старий глобальний sender, щоб відновити після тесту
	oldEmail := mailer.Email
	mailer.Email = mock
	defer func() { mailer.Email = oldEmail }()

	setupTemplates()

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
	assert.Contains(t, mock.LastBody, "https://example.com/api/unsubscribe/tok789")
}
