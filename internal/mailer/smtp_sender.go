package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type SMTPSender struct{}

func (s *SMTPSender) Send(to, subject, htmlBody string) error {
	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	msg := fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		htmlBody

	auth := smtp.PlainAuth("", from, pass, host)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		log.Printf("Error sending HTML email to %s: %v", to, err)
	}
	return err
}
