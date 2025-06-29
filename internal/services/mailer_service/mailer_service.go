package mailer_service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"

	"weatherapi/internal/contracts"
)

// MailerService defines mailer service interface
type MailerServiceProvider interface {
	SendConfirmationEmail(ctx context.Context, email, token string) error
	SendWeatherEmail(ctx context.Context, email, city string, weather contracts.WeatherData, token string) error
}

// MailerService implements MailerServiceProvider
type MailerService struct {
	emailSender contracts.EmailSenderProvider
	appBaseURL  string
	TemplateDir string
}

// NewMailerService creates a new mailer service
// NewMailerService creates a new mailer service with automatic sender selection based on config
func NewMailerService(emailSender contracts.EmailSenderProvider, baseURL string) MailerService {

	return MailerService{
		emailSender: emailSender,
		appBaseURL:  baseURL,
		TemplateDir: "internal/templates", // default template directory
	}
}

// SetTemplateDir sets custom template directory
func (s *MailerService) SetTemplateDir(dir string) {
	s.TemplateDir = dir
}

// GetEmailSender returns the underlying email sender (useful for testing)
func (s *MailerService) GetEmailSender() contracts.EmailSenderProvider {
	return s.emailSender
}

// SendConfirmationEmail sends confirmation email
func (s MailerService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	link := fmt.Sprintf("%s/api/confirm/%s", s.appBaseURL, token)

	data := struct {
		ConfirmURL string
	}{
		ConfirmURL: link,
	}

	body, err := s.renderTemplate("confirmation_email.html", data)
	if err != nil {
		return fmt.Errorf("failed to render confirmation template: %w", err)
	}

	return s.emailSender.Send(email, "Confirm your subscription", body)
}

// SendWeatherEmail sends weather update email
func (s MailerService) SendWeatherEmail(ctx context.Context, email, city string, weather contracts.WeatherData, token string) error {
	data := struct {
		City           string
		Description    string
		Temperature    float64
		Humidity       float64
		UnsubscribeURL string
	}{
		City:           city,
		Description:    weather.Description,
		Temperature:    weather.Temperature,
		Humidity:       weather.Humidity,
		UnsubscribeURL: fmt.Sprintf("%s/api/unsubscribe/%s", s.appBaseURL, token),
	}

	body, err := s.renderTemplate("weather_email.html", data)
	if err != nil {
		return fmt.Errorf("failed to render weather template: %w", err)
	}

	subject := fmt.Sprintf("Weather Update for %s", city)
	return s.emailSender.Send(email, subject, body)
}

// renderTemplate renders HTML template with data
func (s MailerService) renderTemplate(templateName string, data interface{}) (string, error) {
	// In renderTemplate method, replace log.Printf with:

	tmplPath := filepath.Join(s.TemplateDir, templateName)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return body.String(), nil
}
