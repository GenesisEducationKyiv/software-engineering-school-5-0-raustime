package mailer_service

import (
	"context"

	"weatherapi/internal/contracts"
	"weatherapi/internal/mailer"
)

type WeatherEmailData struct {
	Temperature float64
	Humidity    float64
	Description string
}

// EmailSender defines the contract for sending emails.

// MailerService defines mailer service interface
type IMailerService interface {
	SendConfirmationEmail(ctx context.Context, email, token string) error
	SendWeatherEmail(ctx context.Context, email, city string, weather *contracts.WeatherData, token string) error
}

// mailerService implements MailerService
type mailerService struct {
	emailSender contracts.IEmailSender
	AppBaseURL  string
}

// NewMailerService creates a new mailer service
func NewMailerService(emailSender contracts.IEmailSender, baseURL string) IMailerService {
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
func (s *mailerService) SendWeatherEmail(ctx context.Context, email, city string, weather *contracts.WeatherData, token string) error {
	return mailer.SendWeatherEmailWithSender(
		s.emailSender,
		email,
		city,
		weather,
		s.AppBaseURL,
		token,
	)
}
