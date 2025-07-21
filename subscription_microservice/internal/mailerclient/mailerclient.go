package mailerclient

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

	mailerv1 "subscription_microservice/gen/go/mailer/v1"
	mailerv1connect "subscription_microservice/gen/go/mailer/v1/mailerv1connect"
)

type MailerClient struct {
	client mailerv1connect.MailerServiceClient
}

func New(addr string) (*MailerClient, error) {
	if addr == "" {
		return nil, fmt.Errorf("mailer address is empty")
	}

	client := mailerv1connect.NewMailerServiceClient(
		newH2CClient(),
		addr,
	)

	return &MailerClient{client: client}, nil
}

func (m *MailerClient) SendConfirmationEmail(ctx context.Context, email, token string) error {
	stream := m.client.SendEmails(ctx)

	emailReq := &mailerv1.EmailRequest{
		RequestId: uuid.NewString(),
		To:        email,
		Token:     token,
		IsConfirmation: true,
	}

	if err := stream.Send(emailReq); err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}

	// Очікуємо відповідь перед закриттям
	resp, err := stream.Receive()
	if err != nil {
		return fmt.Errorf("receive error: %w", err)
	}

	// Тепер безпечно закрити потік
	_ = stream.CloseRequest()

	if !resp.Delivered {
		return fmt.Errorf("delivery failed: %s", resp.Error)
	}

	log.Printf("✅ confirmation email sent to %s", email)
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
