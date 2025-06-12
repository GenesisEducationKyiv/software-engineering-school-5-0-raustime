package services

import (
	"context"
	"fmt"
	"os"

	"weatherapi/internal/mailer"
)

// MailerService defines mailer service interface
type MailerService interface {
	SendConfirmationEmail(ctx context.Context, email, token string) error
	SendWeatherEmail(ctx context.Context, email, city string, weather *WeatherData, token string) error
}

// mailerService implements MailerService
type mailerService struct {
	emailSender mailer.EmailSender
}

// NewMailerService creates a new mailer service
func NewMailerService(emailSender mailer.EmailSender) MailerService {
	return &mailerService{
		emailSender: emailSender,
	}
}

// SendConfirmationEmail sends confirmation email
func (s *mailerService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	return mailer.SendConfirmationEmailWithSender(s.emailSender, email, token)
}

// SendWeatherEmail sends weather update email
func (s *mailerService) SendWeatherEmail(ctx context.Context, email, city string, weather *WeatherData, token string) error {
	baseURL := os.Getenv("APP_BASE_URL")
	
	weatherData := &openweatherapi.WeatherData{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}
	
	return mailer.SendWeatherEmailWithSender(s.emailSender, email, city, weatherData, baseURL, token)
}