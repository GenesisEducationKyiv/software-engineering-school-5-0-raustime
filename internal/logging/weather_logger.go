package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"weatherapi/internal/contracts"
)

type WeatherLogger interface {
	LogResponse(provider string, data contracts.WeatherData, err error)
}

type FileWeatherLogger struct {
	logFile string
}

func NewFileWeatherLogger(logFile string) *FileWeatherLogger {
	return &FileWeatherLogger{
		logFile: logFile,
	}
}

func (l *FileWeatherLogger) LogResponse(provider string, data contracts.WeatherData, err error) {
	logEntry := map[string]interface{}{
		"provider": provider,
		"success":  err == nil,
	}

	if err != nil {
		logEntry["error"] = err.Error()
	} else {
		logEntry["response"] = data
	}

	logJSON, _ := json.Marshal(logEntry)
	logMessage := string(logJSON)

	// Log to file
	l.logToFile(logMessage)

	// Also log to console for debugging
	log.Printf("%s - Response: %s", provider, logMessage)
}

func (l *FileWeatherLogger) logToFile(logMessage string) {
	file, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer func() { _ = file.Close() }()

	if _, err := file.WriteString(fmt.Sprintf("%s\n", logMessage)); err != nil {
		log.Printf("Failed to write to log file: %v", err)
	}
}

// MockLogger для тестування
type MockLogger struct {
	LoggedResponses []LogEntry
}

type LogEntry struct {
	Provider string
	Data     contracts.WeatherData
	Error    error
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		LoggedResponses: make([]LogEntry, 0),
	}
}

func (m *MockLogger) LogResponse(provider string, data contracts.WeatherData, err error) {
	m.LoggedResponses = append(m.LoggedResponses, LogEntry{
		Provider: provider,
		Data:     data,
		Error:    err,
	})

	// Optional: also log to console for debugging in tests
	log.Printf("[MOCK] %s - Success: %t", provider, err == nil)
}

// Helper methods for test assertions
func (m *MockLogger) GetLogCount() int {
	return len(m.LoggedResponses)
}

func (m *MockLogger) GetLastLog() *LogEntry {
	if len(m.LoggedResponses) == 0 {
		return nil
	}
	return &m.LoggedResponses[len(m.LoggedResponses)-1]
}

func (m *MockLogger) GetLogsByProvider(provider string) []LogEntry {
	var results []LogEntry
	for _, entry := range m.LoggedResponses {
		if entry.Provider == provider {
			results = append(results, entry)
		}
	}
	return results
}

func (m *MockLogger) Reset() {
	m.LoggedResponses = make([]LogEntry, 0)
}
