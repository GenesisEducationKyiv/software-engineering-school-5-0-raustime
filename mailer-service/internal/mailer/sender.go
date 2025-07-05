package mailer

import (
	"fmt"
	"net/smtp"
)

type Sender struct {
	host string
	port int
	user string
	pass string
}

func New(host string, port int, user, pass string) *Sender {
	return &Sender{host, port, user, pass}
}

func (s *Sender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := "From: " + s.user + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" + body

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	return smtp.SendMail(addr, auth, s.user, []string{to}, []byte(msg))
}
