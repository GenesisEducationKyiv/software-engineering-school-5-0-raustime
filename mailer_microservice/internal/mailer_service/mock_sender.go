// mock_sender.go
package mailer_service

import (
	"errors"
	"sync"

	"mailer_microservice/internal/config"
)

// MockSender implements EmailSenderProvider for testing.
type MockSender struct {
	mu           sync.Mutex
	LastTo       string
	LastSubject  string
	LastBody     string
	SentEmails   []SentEmail
	ShouldFail   bool
	ErrorMessage string
}

// SentEmail represents a sent email for testing.
type SentEmail struct {
	To      string
	Subject string
	Body    string
}

// NewMockSender creates a new mock email sender.
func NewMockSender() *MockSender {
	return &MockSender{
		SentEmails: make([]SentEmail, 0),
	}
}

// Send mocks sending an email and stores the data for verification.
func (m *MockSender) Send(to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ShouldFail {
		errorMsg := m.ErrorMessage
		if errorMsg == "" {
			errorMsg = "mock sender error"
		}
		return errors.New(errorMsg) // return після вкладеного if.
	}

	// Store for compatibility with existing tests.
	m.LastTo = to
	m.LastSubject = subject
	m.LastBody = body

	// Store in history for advanced testing.
	m.SentEmails = append(m.SentEmails, SentEmail{
		To:      to,
		Subject: subject,
		Body:    body,
	})

	return nil
}

// SetShouldFail configures the mock to return an error.
func (m *MockSender) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShouldFail = shouldFail
}

// SetErrorMessage sets custom error message.
func (m *MockSender) SetErrorMessage(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorMessage = msg
}

// GetSentEmails returns all sent emails.
func (m *MockSender) GetSentEmails() []SentEmail {
	m.mu.Lock()
	defer m.mu.Unlock()

	emails := make([]SentEmail, len(m.SentEmails))
	copy(emails, m.SentEmails)
	return emails
}

// GetSentEmailsCount returns the number of sent emails.
func (m *MockSender) GetSentEmailsCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.SentEmails)
}

// GetLastSentEmail returns the last sent email.
func (m *MockSender) GetLastSentEmail() *SentEmail {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.SentEmails) == 0 {
		return nil
	}
	return &m.SentEmails[len(m.SentEmails)-1]
}

// HasEmailBeenSentTo checks if an email was sent to the specified address.
func (m *MockSender) HasEmailBeenSentTo(email string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sentEmail := range m.SentEmails {
		if sentEmail.To == email {
			return true
		}
	}
	return false
}

// HasEmailWithSubject checks if an email with the specified subject was sent.
func (m *MockSender) HasEmailWithSubject(subject string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sentEmail := range m.SentEmails {
		if sentEmail.Subject == subject {
			return true
		}
	}
	return false
}

// Clear clears the email history.
func (m *MockSender) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LastTo = ""
	m.LastSubject = ""
	m.LastBody = ""
	m.SentEmails = m.SentEmails[:0]
}

// Reset resets the mock to its initial state.
func (m *MockSender) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LastTo = ""
	m.LastSubject = ""
	m.LastBody = ""
	m.SentEmails = m.SentEmails[:0]
	m.ShouldFail = false
	m.ErrorMessage = ""
}

// Test helper functions.

// CreateTestConfig creates a standard test configuration.
func CreateTestConfig() *config.Config {
	return &config.Config{
		AppBaseURL:   "https://test.com",
		Environment:  "test",
		SMTPHost:     "test-smtp.com",
		SMTPPort:     587,
		SMTPUser:     "test@example.com",
		SMTPPassword: "testpass",
	}
}
