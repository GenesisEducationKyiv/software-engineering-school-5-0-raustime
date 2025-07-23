package mailer_service

import (
	"context"
	"mailer_microservice/internal/contracts"
)

type MockMailerService struct {
	LastTo      string
	LastSubject string
	LastBody    string
}

func NewMockMailerService() *MockMailerService {
	return &MockMailerService{}
}

func (m *MockMailerService) GetLastBody() string {
	return m.LastBody
}

func (m *MockMailerService) SendEmail(ctx context.Context, to, subject, body string) error {
	m.LastTo = to
	m.LastSubject = subject
	m.LastBody = body
	return nil
}

func (m *MockMailerService) SendConfirmationEmail(ctx context.Context, email, token string) error {
	return nil
}

func (m *MockMailerService) SendWeatherEmail(ctx context.Context, email, city string, weather contracts.WeatherData, token string) error {
	return nil
}

func (m *MockMailerService) HasEmailBeenSentTo(email string) bool {
	return m.LastTo == email
}

func (m *MockMailerService) HasEmailWithSubject(subject string) bool {
	return m.LastSubject == subject
}
