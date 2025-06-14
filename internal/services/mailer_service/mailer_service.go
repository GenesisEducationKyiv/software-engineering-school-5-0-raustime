package mailer_service

import (
	"context"

	"weatherapi/internal/mailer"
	"weatherapi/internal/services/weather_service"
)

// MailerService defines mailer service interface
type IMailerService interface {
	SendConfirmationEmail(ctx context.Context, email, token string) error
	SendWeatherEmail(ctx context.Context, email, city string, weather *weather_service.WeatherData, token string) error
}

// mailerService implements MailerService
type mailerService struct {
	emailSender mailer.EmailSender
	AppBaseURL  string
}

// NewMailerService creates a new mailer service
func NewMailerService(emailSender mailer.EmailSender, baseURL string) IMailerService {
	return &mailerService{
		emailSender: emailSender,
		AppBaseURL:  baseURL,
	}
}

// SendConfirmationEmail sends confirmation email
func (s *mailerService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	return mailer.SendConfirmationEmailWithSender(s.emailSender, email, token)
}

// SendWeatherEmail sends weather update email
func (s *mailerService) SendWeatherEmail(ctx context.Context, email, city string, weather *weather_service.WeatherData, token string) error {
	baseURL := s.AppBaseURL
	weatherData := &weather_service.WeatherData{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description,
	}

	return mailer.SendWeatherEmailWithSender(s.emailSender, email, city, weatherData, baseURL, token)
}
