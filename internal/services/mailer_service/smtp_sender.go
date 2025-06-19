package mailer_service

import (
	"fmt"
	"log"
	"net/smtp"
)

type SMTPSender struct {
	From     string
	Password string
	Host     string
	Port     string
}

func NewSMTPSender(from, pass, host string, port string) *SMTPSender {
	return &SMTPSender{
		From:     from,
		Password: pass,
		Host:     host,
		Port:     port,
	}
}

func (s *SMTPSender) Send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	msg := fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		htmlBody

	auth := smtp.PlainAuth("", s.From, s.Password, s.Host)
	err := smtp.SendMail(addr, auth, s.From, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("Error sending HTML email to %s: %v", to, err)
	}
	return err
}
