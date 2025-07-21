package notification_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"mailer_microservice/internal/contracts"
	"mailer_microservice/internal/notification"
	"mailer_microservice/internal/services/mailer_service"
)

func TestHandleMessage_Custom_Success(t *testing.T) {
	mockMailer := mailer_service.NewMockMailerService()
	consumer := notification.NewNotificationConsumer(mockMailer)

	msg := contracts.NotificationMessage{
		Type:    "custom",
		To:      "test@example.com",
		Subject: "Hi",
		Body:    "<b>Hello</b>",
	}
	data, _ := json.Marshal(msg)

	err := consumer.HandleMessage(context.Background(), data)
	require.NoError(t, err)
	require.Equal(t, "test@example.com", mockMailer.LastTo)
	require.Equal(t, "Hi", mockMailer.LastSubject)
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	mockMailer := mailer_service.NewMockMailerService()
	consumer := notification.NewNotificationConsumer(mockMailer)

	badJSON := []byte(`{ this is not valid }`)

	err := consumer.HandleMessage(context.Background(), badJSON)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid JSON")
}

func TestHandleMessage_Weather_MissingData(t *testing.T) {
	mockMailer := mailer_service.NewMockMailerService()
	consumer := notification.NewNotificationConsumer(mockMailer)

	msg := contracts.NotificationMessage{
		Type: "weather",
		To:   "weather@example.com",
		City: "Lviv",
		// Weather is nil
	}
	data, _ := json.Marshal(msg)

	err := consumer.HandleMessage(context.Background(), data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing weather")
}
