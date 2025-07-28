package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"mailer_microservice/internal/contracts"
	"mailer_microservice/internal/mailer_service"
)

const (
	NotificationTypeConfirmation = "confirmation"
	NotificationTypeWeather      = "weather"
)

type NotificationConsumer struct {
	mailer mailer_service.MailerServiceProvider
}

func NewNotificationConsumer(mailer mailer_service.MailerServiceProvider) *NotificationConsumer {
	return &NotificationConsumer{mailer: mailer}
}

func (c *NotificationConsumer) HandleMessage(ctx context.Context, msg []byte) error {
	var notif contracts.NotificationMessage
	if err := json.Unmarshal(msg, &notif); err != nil {
		log.Printf("❌ Failed to decode message: %v", err)
		return fmt.Errorf("invalid JSON: %w", err)
	}

	var err error
	switch notif.Type {
	case NotificationTypeConfirmation:
		err = c.mailer.SendConfirmationEmail(ctx, notif.To, notif.Token)
	case NotificationTypeWeather:
		if notif.Weather == nil {
			return fmt.Errorf("missing weather field")
		}
		err = c.mailer.SendWeatherEmail(ctx, notif.To, notif.City, *notif.Weather, notif.Token)
	default:
		err = c.mailer.SendEmail(ctx, notif.To, notif.Subject, notif.Body)
	}

	if err != nil {
		log.Printf("❌ Failed to send %s email to %s: %v", notif.Type, notif.To, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ [%s] email sent to %s", notif.Type, notif.To)
	return nil
}
