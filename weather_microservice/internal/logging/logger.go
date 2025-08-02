package logging

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"weather_microservice/internal/pkg/ctxkeys"
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

func NewZapWeatherLogger(logPath string, levelStr string) *ZapWeatherLogger {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(levelStr)); err != nil {
		lvl = zapcore.InfoLevel
	}
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
		lvl,
	)

	return &ZapWeatherLogger{
		logger: zap.New(core, zap.AddCaller()),
	}
}
func (z *ZapWeatherLogger) Info(ctx context.Context, source string, payload any) {
	z.logger.Info("info", buildFields(ctx, source, payload, nil)...)
}

func (z *ZapWeatherLogger) Warn(ctx context.Context, source string, payload any, err error) {
	z.logger.Warn("warn", buildFields(ctx, source, payload, err)...)
}

func (z *ZapWeatherLogger) Error(ctx context.Context, source string, payload any, err error) {
	z.logger.Error("error", buildFields(ctx, source, payload, err)...)
}

func (z *ZapWeatherLogger) Debug(ctx context.Context, source string, payload any) {
	z.logger.Debug("debug", buildFields(ctx, source, payload, nil)...)
}

func buildFields(ctx context.Context, source string, payload any, err error) []zap.Field {
	fields := []zap.Field{
		zap.String("source", source),
		zap.Time("timestamp", time.Now()),
		zap.Bool("success", err == nil),
	}

	if traceID, ok := ctx.Value(ctxkeys.TraceIDKey).(string); ok {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if reqID, ok := ctx.Value(ctxkeys.RequestIDKey).(string); ok {
		fields = append(fields, zap.String("request_id", reqID))
	}
	if userID, ok := ctx.Value(ctxkeys.UserIDKey).(string); ok {
		fields = append(fields, zap.String("user_id", userID))
	}

	if payload != nil {
		fields = append(fields, zap.Any("payload", payload))
	}
	if err != nil {
		fields = append(fields, zap.String("error", err.Error()))
	}
	return fields
}
