package logging

import (
	"context"
	"weather_microservice/internal/pkg/ctxkeys"
)

type nopLogger struct{}

func (nopLogger) Info(context.Context, string, any)         {}
func (nopLogger) Warn(context.Context, string, any, error)  {}
func (nopLogger) Error(context.Context, string, any, error) {}
func (nopLogger) Debug(context.Context, string, any)        {}

var defaultLogger = nopLogger{}

// FromContext returns Logger from context, or nil if not present.
func FromContext(ctx context.Context) Logger {
	if v := ctx.Value(ctxkeys.Logger); v != nil {
		if logger, ok := v.(Logger); ok {
			return logger
		}
	}
	return defaultLogger
}

// Info logs an info-level message from context.
func Info(ctx context.Context, source string, payload any) {
	if logger := FromContext(ctx); logger != nil {
		logger.Info(ctx, source, payload)
	}
}

// Warn logs a warn-level message from context.
func Warn(ctx context.Context, source string, payload any, err error) {
	if logger := FromContext(ctx); logger != nil {
		logger.Warn(ctx, source, payload, err)
	}
}

// Error logs an error-level message from context.
func Error(ctx context.Context, source string, payload any, err error) {
	if logger := FromContext(ctx); logger != nil {
		logger.Error(ctx, source, payload, err)
	}
}

// Debug logs a debug-level message from context.
func Debug(ctx context.Context, source string, payload any) {
	if logger := FromContext(ctx); logger != nil {
		logger.Debug(ctx, source, payload)
	}
}
