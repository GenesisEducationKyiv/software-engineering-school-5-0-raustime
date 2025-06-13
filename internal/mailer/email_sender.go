package mailer

type EmailSender interface {
	Send(to string, subject string, htmlBody string) error
}
