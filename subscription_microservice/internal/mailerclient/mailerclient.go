package mailerclient

import (
	"context"
	"fmt"
	"log"
	"net/http"

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
		http.DefaultClient, // use connect-go HTTP client
		addr,
	)

	return &MailerClient{client: client}, nil
}

func (m *MailerClient) SendConfirmationEmail(ctx context.Context, email, token string) error {
	stream := m.client.SendEmails(ctx)

	err := stream.Send(&mailerv1.EmailRequest{
		To:    email,
		Token: token,
		// інші поля залишити порожніми або за замовчуванням
	})

	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}

	resp, err := stream.Receive()
	if err != nil {
		return fmt.Errorf("receive error: %w", err)
	}

	if !resp.Delivered {
		return fmt.Errorf("delivery failed: %s", resp.Error)
	}

	log.Printf("✅ confirmation email sent to %s", email)
	return nil
}
