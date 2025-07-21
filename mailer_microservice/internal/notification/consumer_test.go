package notification_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"mailer_microservice/internal/mailer_service"
	"mailer_microservice/internal/notification"
)

func TestNotificationConsumer_HandleMessage(t *testing.T) {
	mockMailer := mailer_service.NewMockMailerService()
	consumer := notification.NewNotificationConsumer(mockMailer)

	msg := `{"to":"test@example.com","subject":"Hello","body":"World"}`
	consumer.HandleMessage(context.Background(), []byte(msg))

	require.True(t, mockMailer.HasEmailBeenSentTo("test@example.com"))
	require.True(t, mockMailer.HasEmailWithSubject("Hello"))
}
