package logging

import "context"

type MockLogger struct {
	Entries []LogEntry
}

type LogEntry struct {
	Level   string
	Source  string
	Payload any
	Error   error
}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (m *MockLogger) Info(ctx context.Context, source string, payload any) {
	m.Entries = append(m.Entries, LogEntry{"info", source, payload, nil})
}

func (m *MockLogger) Warn(ctx context.Context, source string, payload any, err error) {
	m.Entries = append(m.Entries, LogEntry{"warn", source, payload, err})
}

func (m *MockLogger) Error(ctx context.Context, source string, payload any, err error) {
	m.Entries = append(m.Entries, LogEntry{"error", source, payload, err})
}

func (m *MockLogger) Debug(ctx context.Context, source string, payload any) {
	m.Entries = append(m.Entries, LogEntry{"debug", source, payload, nil})
}
