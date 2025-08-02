package ctxkeys

type contextKey string

const (
	Logger       contextKey = "logger"
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
	UserIDKey    contextKey = "user_id"
)
