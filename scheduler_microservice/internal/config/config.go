package config

import (
	"os"
)

type Config struct {
	Port              string
	MailerServiceURL  string
	SubscriptionURL   string
	WeatherServiceURL string
	NATSUrl           string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:              getEnv("SCHEDULER_PORT", "8092"),
		MailerServiceURL:  getEnv("MAILER_SERVICE_URL", "http://mailer_service:8089"),
		SubscriptionURL:   getEnv("SUBSCRIPTION_SERVICE_URL", "http://subscription_service:8091"),
		WeatherServiceURL: getEnv("WEATHER_SERVICE_URL", "http://weather_service:8080"),
		NATSUrl:           getEnv("NATS_URL", "nats://localhost:4222"),
	}
	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
