package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/http2"

	mailerv1 "scheduler_microservice/gen/go/mailer/v1"
	mailerv1connect "scheduler_microservice/gen/go/mailer/v1/mailerv1connect"
	"scheduler_microservice/internal/contracts"
)

type mailerClient struct {
	client mailerv1connect.MailerServiceClient
}

func NewMailerClient(baseURL string) *mailerClient {
	httpClient := newH2CClient()
	client := mailerv1connect.NewMailerServiceClient(httpClient, baseURL)
	return &mailerClient{client: client}
}

func (c *mailerClient) SendWeatherEmail(ctx context.Context, to, city string, data *contracts.WeatherData, token string) error {
	stream := c.client.SendEmails(ctx)

	emailReq := &mailerv1.EmailRequest{
		RequestId:   uuid.NewString(),
		To:          to,
		City:        city,
		Token:       token,
		Temperature: float32(data.Temperature),
		Humidity:    float32(data.Humidity),
		Description: data.Description,
		IsConfirmation: false,
	}

	if err := stream.Send(emailReq); err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}

	// Отримуємо відповідь перед закриттям
	resp, err := stream.Receive()
	if err != nil {
		return fmt.Errorf("receive error: %w", err)
	}

	_ = stream.CloseRequest()

	if !resp.Delivered {
		return fmt.Errorf("delivery failed: %s", resp.Error)
	}

	log.Printf("✅ weather email sent to %s for city %s", to, city)
	return nil
}

func newH2CClient() *http.Client {
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}
