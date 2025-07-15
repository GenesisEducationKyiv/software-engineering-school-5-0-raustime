package config

import (
	"os"
)

type Config struct {
	Port               string
	MailerServiceURL   string
	SubscriptionURL    string
	WeatherServiceURL  string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("SCHEDULER_PORT", "8092"),
		MailerServiceURL:   getEnv("MAILER_SERVICE_URL", "http://mailer-service:8091"),
		SubscriptionURL:    getEnv("SUBSCRIPTION_SERVICE_URL", "http://subscription-service:8090"),
		WeatherServiceURL:  getEnv("WEATHER_SERVICE_URL", "http://weather-service:8080"),
	}
	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
