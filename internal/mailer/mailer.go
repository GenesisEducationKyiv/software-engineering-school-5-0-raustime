package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"os"
	"weatherapi/internal/openweatherapi"
)

type confirmationData struct {
	ConfirmURL string
}

var (
	Email       EmailSender = &SMTPSender{}
	TemplateDir string      = getTemplateDir()
)

func getTemplateDir() string {
	if dir := os.Getenv("TEMPLATE_DIR"); dir != "" {
		return dir
	}
	return "internal/templates"
}

func SendConfirmationEmailWithSender(sender EmailSender, to, token string) error {
	apiHost := os.Getenv("APP_BASE_URL")
	link := fmt.Sprintf("%s/api/confirm/%s", apiHost, token)
	data := confirmationData{ConfirmURL: link}

	tmplPath := filepath.Join(TemplateDir, "confirmation_email.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}
	return sender.Send(to, "Confirm your subscription", body.String())
}

func SendWeatherEmailWithSender(sender EmailSender, to, city string, weather *openweatherapi.WeatherData, baseURL, token string) error {
	tmplPath := filepath.Join(TemplateDir, "weather_email.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return err
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
		return err
	}

	subject := fmt.Sprintf("Weather Update for %s", city)
	return sender.Send(to, subject, body.String())
}
