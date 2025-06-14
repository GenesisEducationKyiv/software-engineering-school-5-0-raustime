package mailer

type MockSender struct {
	LastTo      string
	LastSubject string
	LastBody    string
}

func (m *MockSender) Send(to, subject, htmlBody string) error {
	m.LastTo = to
	m.LastSubject = subject
	m.LastBody = htmlBody
	return nil
}
