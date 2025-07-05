package config

import "os"

type Config struct {
	NATSURL  string
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
}

func Load() Config {
	return Config{
		NATSURL:  os.Getenv("NATS_URL"), // example: nats://nats:4222
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: 587, // або з env
		SMTPUser: os.Getenv("SMTP_USER"),
		SMTPPass: os.Getenv("SMTP_PASS"),
	}
}
