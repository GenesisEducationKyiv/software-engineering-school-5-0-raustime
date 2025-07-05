package nats

import (
	"encoding/json"
	"log"
	"mailer-service/internal/config"
	"mailer-service/internal/mailer"

	"github.com/nats-io/nats.go"
)

type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type Listener struct {
	nc     *nats.Conn
	sender *mailer.Sender
}

func NewListener(cfg config.Config) (*Listener, error) {
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, err
	}

	sender := mailer.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	return &Listener{nc: nc, sender: sender}, nil
}

func (l *Listener) Listen() error {
	_, err := l.nc.Subscribe("mailer.send", func(m *nats.Msg) {
		var email EmailMessage
		if err := json.Unmarshal(m.Data, &email); err != nil {
			log.Printf("failed to parse email message: %v", err)
			return
		}

		log.Printf("Sending email to %s", email.To)
		if err := l.sender.Send(email.To, email.Subject, email.Body); err != nil {
			log.Printf("email send error: %v", err)
		}
	})

	return err
}
