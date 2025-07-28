package testutils

import (
	"context"
	"weather_microservice/internal/logging"
	"weather_microservice/internal/pkg/ctxkeys"
)

func WithMockLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxkeys.Logger, &logging.MockLogger{})
}
