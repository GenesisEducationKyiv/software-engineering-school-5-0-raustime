package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"weather_microservice/internal/logging"
	"weather_microservice/internal/pkg/ctxkeys"
)

// Middleware represents a middleware function.
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middleware functions.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// CORS adds CORS headers.
func CORS() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, Accept, X-Trace-Id")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Length, X-Trace-Id")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "43200")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Trace injects trace_id into context and response headers.
func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get("X-Trace-Id")
			if traceID == "" {
				traceID = uuid.NewString()
			}

			// Expose trace ID in response
			w.Header().Set("X-Trace-Id", traceID)

			ctx := context.WithValue(r.Context(), ctxkeys.TraceIDKey, traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logging logs structured HTTP requests using ZapWeatherLogger.
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lw := &loggingWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(lw, r)

			logger := logging.FromContext(r.Context())
			if logger != nil {
				logger.Info(r.Context(), "http:Request", map[string]interface{}{
					"method":      r.Method,
					"path":        r.URL.Path,
					"status":      lw.statusCode,
					"duration_ms": time.Since(start).Milliseconds(),
				})
			}
		})
	}
}

// Recovery recovers from panics and logs them.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger := logging.FromContext(r.Context())
					if logger != nil {
						logger.Error(r.Context(), "http:Recovery", nil, err.(error))
					}
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type loggingWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggingWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}
