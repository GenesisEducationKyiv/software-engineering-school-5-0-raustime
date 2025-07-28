package logging

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"weather_microservice/internal/contracts"
)

type WeatherLogger interface {
	LogResponse(provider string, data contracts.WeatherData, err error)
}

type ZapWeatherLogger struct {
	logger *zap.Logger
}

func NewZapWeatherLogger(logPath string) *ZapWeatherLogger {
	writerSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // days
		Compress:   true,
	})

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "msg"

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writerSyncer,
		zap.InfoLevel,
	)

	return &ZapWeatherLogger{
		logger: zap.New(core),
	}
}

func (z *ZapWeatherLogger) LogResponse(provider string, data contracts.WeatherData, err error) {
	fields := []zap.Field{
		zap.String("provider", provider),
		zap.Time("timestamp", time.Now()),
	}

	if err != nil {
		fields = append(fields, zap.Bool("success", false), zap.String("error", err.Error()))
		z.logger.Error("weather response failed", fields...)
	} else {
		fields = append(fields,
			zap.Bool("success", true),
			zap.Float64("temperature", data.Temperature),
			zap.Float64("humidity", data.Humidity),
			zap.String("description", data.Description),
		)
		z.logger.Info("weather response received", fields...)
	}
}

// MockLogger для тестування.
type MockLogger struct {
	LoggedResponses []LogEntry
}

type LogEntry struct {
	Provider string
	Data     contracts.WeatherData
	Error    error
	Success  bool
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
		Success:  err == nil,
	})
}

// ===== 🧪 Test helper methods =====

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
