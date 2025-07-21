package notification

import (
	"context"
	"encoding/json"
	"log"

	"mailer_microservice/internal/mailer_service"
)

type NotificationConsumer struct {
	mailer mailer_service.MailerServiceProvider
}

func NewNotificationConsumer(mailer mailer_service.MailerServiceProvider) *NotificationConsumer {
	return &NotificationConsumer{mailer: mailer}
}

type NotificationMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (c *NotificationConsumer) HandleMessage(ctx context.Context, msg []byte) {
	var notif NotificationMessage
	if err := json.Unmarshal(msg, &notif); err != nil {
		log.Printf("❌ Failed to decode message: %v", err)
		return
	}

	err := c.mailer.SendEmail(ctx, notif.To, notif.Subject, notif.Body)
	if err != nil {
		log.Printf("❌ Failed to send email: %v", err)
		return
	}

	log.Printf("✅ Email sent to %s", notif.To)
}
