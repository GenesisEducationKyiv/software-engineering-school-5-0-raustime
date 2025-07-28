package logging

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Info(ctx context.Context, source string, payload any)
	Warn(ctx context.Context, source string, payload any, err error)
	Error(ctx context.Context, source string, payload any, err error)
	Debug(ctx context.Context, source string, payload any)
}

type LogLevel string

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

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.LevelKey = "level"
	encoderCfg.MessageKey = "msg"
	encoderCfg.CallerKey = "caller"

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		writerSyncer,
		zap.InfoLevel,
	)

	return &ZapWeatherLogger{
		logger: zap.New(core, zap.AddCaller()),
	}
}
func (z *ZapWeatherLogger) Info(ctx context.Context, source string, payload any) {
	z.logger.Info("info", buildFields(source, payload, nil)...)
}

func (z *ZapWeatherLogger) Warn(ctx context.Context, source string, payload any, err error) {
	z.logger.Warn("warn", buildFields(source, payload, err)...)
}

func (z *ZapWeatherLogger) Error(ctx context.Context, source string, payload any, err error) {
	z.logger.Error("error", buildFields(source, payload, err)...)
}

func (z *ZapWeatherLogger) Debug(ctx context.Context, source string, payload any) {
	z.logger.Debug("debug", buildFields(source, payload, nil)...)
}

func buildFields(source string, payload any, err error) []zap.Field {
	fields := []zap.Field{
		zap.String("source", source),
		zap.Time("timestamp", time.Now()),
		zap.Bool("success", err == nil),
	}

	if payload != nil {
		fields = append(fields, zap.Any("payload", payload))
	}
	if err != nil {
		fields = append(fields, zap.String("error", err.Error()))
	}
	return fields
}
