package mailer_service

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"html/template"
	"path/filepath"

	"mailer_microservice/internal/contracts"
)

// MailerService defines mailer service interface.
type MailerServiceProvider interface {
	SendConfirmationEmail(ctx context.Context, email, token string) error
	SendWeatherEmail(ctx context.Context, email, city string, weather contracts.WeatherData, token string) error
}

// MailerService implements MailerServiceProvider.
type MailerService struct {
	emailSender contracts.EmailSenderProvider
	appBaseURL  string
	TemplateDir string
}

// NewMailerService creates a new mailer service.
func NewMailerService(emailSender contracts.EmailSenderProvider, baseURL string) *MailerService {
	return &MailerService{
		emailSender: emailSender,
		appBaseURL:  baseURL,
		TemplateDir: "internal/templates", // default template directory
	}
}

// SetTemplateDir sets custom template directory.
func (s *MailerService) SetTemplateDir(dir string) {
	s.TemplateDir = dir
}

// GetEmailSender returns the underlying email sender (useful for testing).
func (s *MailerService) GetEmailSender() contracts.EmailSenderProvider {
	return s.emailSender
}

// SendConfirmationEmail sends confirmation email.
func (s *MailerService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	link := fmt.Sprintf("%s/api/confirm/%s", s.appBaseURL, token)

	data := struct {
		ConfirmURL string
	}{
		ConfirmURL: link,
	}

	body, err := s.renderTemplate("confirmation_email.html", data)
	if err != nil {
		log.Printf("[MailerService] ❌ failed to render confirmation template: %v", err)
		return fmt.Errorf("failed to render confirmation template: %w", err)
	}
	log.Printf("[MailerService] 📩 sending confirmation email to %s with link: %s", email, link)
	if err := s.emailSender.Send(email, "Confirm your subscription", body); err != nil {
		log.Printf("[MailerService] ❌ failed to send confirmation email to %s: %v", email, err)
		return err
	}
	log.Printf("[MailerService] ✅ confirmation email sent to %s", email)
	return nil
}

// SendWeatherEmail sends weather update email.
func (s *MailerService) SendWeatherEmail(ctx context.Context, email, city string, weather contracts.WeatherData, token string) error {
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
		log.Printf("[MailerService] ❌ failed to render weather template: %v", err)
		return fmt.Errorf("failed to render weather template: %w", err)
	}

	subject := fmt.Sprintf("Weather Update for %s", city)
	log.Printf("[MailerService] 📩 sending weather email to %s for city %s", email, city)

	if err := s.emailSender.Send(email, subject, body); err != nil {
		log.Printf("[MailerService] ❌ failed to send weather email to %s: %v", email, err)
		return err
	}

	log.Printf("[MailerService] ✅ weather email sent to %s", email)
	return nil
}

// renderTemplate renders HTML template with data.
func (s *MailerService) renderTemplate(templateName string, data interface{}) (string, error) {
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
