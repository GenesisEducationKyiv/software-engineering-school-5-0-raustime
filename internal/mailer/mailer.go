package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"path/filepath"
	"weatherapi/internal/services/weather_service"

	"os"
)

type confirmationData struct {
	ConfirmURL string
}

var (
	Email       EmailSender = &SMTPSender{}
	TemplateDir string
)

// SetTemplateDir sets the template directory (useful for testing)
func SetTemplateDir(dir string) {
	TemplateDir = dir
	log.Printf("DEBUG: SetTemplateDir called with: %s", dir)
}

// GetTemplatePath returns the full path to a template file
func GetTemplatePath(filename string) string {
	var dir string

	// Priority order:
	// 1. Explicitly set TemplateDir variable
	// 2. TEMPLATE_DIR environment variable
	// 3. Default path
	if TemplateDir != "" {
		dir = TemplateDir
		log.Printf("DEBUG: Using TemplateDir variable: %s", dir)
	} else if envDir := os.Getenv("TEMPLATE_DIR"); envDir != "" {
		dir = envDir
		log.Printf("DEBUG: Using TEMPLATE_DIR env var: %s", dir)
	} else {
		dir = "internal/templates"
		log.Printf("DEBUG: Using default template dir: %s", dir)
	}

	fullPath := filepath.Join(dir, filename)
	log.Printf("DEBUG: Full template path: %s", fullPath)

	// Check if file exists and log the result
	if _, err := os.Stat(fullPath); err != nil {
		log.Printf("DEBUG: Template file does not exist: %s, error: %v", fullPath, err)
	} else {
		log.Printf("DEBUG: Template file exists: %s", fullPath)
	}

	return fullPath
}

func SendConfirmationEmailWithSender(sender EmailSender, to, token string) error {
	log.Printf("DEBUG: SendConfirmationEmailWithSender called for: %s", to)

	apiHost := os.Getenv("APP_BASE_URL")
	link := fmt.Sprintf("%s/api/confirm/%s", apiHost, token)
	data := confirmationData{ConfirmURL: link}

	tmplPath := GetTemplatePath("confirmation_email.html")
	log.Printf("DEBUG: About to parse template: %s", tmplPath)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("ERROR: Failed to parse template %s: %v", tmplPath, err)
		return fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Printf("ERROR: Failed to execute template: %v", err)
		return fmt.Errorf("failed to execute template: %w", err)
	}

	log.Printf("DEBUG: Sending confirmation email via sender")
	return sender.Send(to, "Confirm your subscription", body.String())
}

// WeatherData represents weather information for an email.
type WeatherData struct {
	Description string
	Temperature float64
	Humidity    float64
}

func SendWeatherEmailWithSender(sender EmailSender, to, city string, weather *weather_service.WeatherData, baseURL, token string) error {
	log.Printf("DEBUG: SendWeatherEmailWithSender called for: %s, city: %s", to, city)

	tmplPath := GetTemplatePath("weather_email.html")
	log.Printf("DEBUG: About to parse weather template: %s", tmplPath)

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("ERROR: Failed to parse weather template %s: %v", tmplPath, err)
		return fmt.Errorf("failed to parse weather template %s: %w", tmplPath, err)
	}

	var body bytes.Buffer
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
		UnsubscribeURL: fmt.Sprintf("%s/api/unsubscribe/%s", baseURL, token),
	}
	if err := tmpl.Execute(&body, data); err != nil {
		log.Printf("ERROR: Failed to execute weather template: %v", err)
		return fmt.Errorf("failed to execute weather template: %w", err)
	}

	subject := fmt.Sprintf("Weather Update for %s", city)
	log.Printf("DEBUG: Sending weather email via sender")
	return sender.Send(to, subject, body.String())
}
