package clients

import (
	"context"
	"fmt"
	"scheduler_microservice/internal/contracts"
	"time"

	"net/http"
	mailerv1 "scheduler_microservice/gen/go/mailer/v1"
	mailerv1connect "scheduler_microservice/gen/go/mailer/v1/mailerv1connect"
)

type mailerClient struct {
	client mailerv1connect.MailerServiceClient
}

func NewMailerClient(httpClient *http.Client, baseURL string) *mailerClient {
	client := mailerv1connect.NewMailerServiceClient(httpClient, baseURL)
	return &mailerClient{client: client}
}

func (c *mailerClient) SendWeatherEmail(ctx context.Context, to, city string, data *contracts.WeatherData, token string) error {
	// Create bidirectional stream
	stream := c.client.SendEmails(ctx)

	// Generate a unique request ID
	requestID := fmt.Sprintf("weather-%s-%d", to, time.Now().Unix())

	// Send the email request
	err := stream.Send(&mailerv1.EmailRequest{
		RequestId:   requestID,
		To:          to,
		City:        city,
		Token:       token,
		Temperature: float32(data.Temperature),
		Humidity:    float32(data.Humidity),
		Description: data.Description,
	})
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}

	// Close the send side to indicate we're done sending
	if err := stream.CloseRequest(); err != nil {
		return fmt.Errorf("failed to close request stream: %w", err)
	}

	// Read the response
	response, err := stream.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	// Check if email was delivered successfully
	if !response.Delivered {
		return fmt.Errorf("email delivery failed: %s", response.Error)
	}

	return nil
}
