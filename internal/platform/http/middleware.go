package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	stdhttp "net/http"
	"time"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

type responseRecorder struct {
	stdhttp.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func requestIDMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		requestID := newRequestID()
		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		w.Header().Set("X-Request-Id", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func recovererMiddleware(logger *log.Logger) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Printf("panic request_id=%s method=%s path=%s err=%v", requestIDFromContext(r.Context()), r.Method, r.URL.Path, recovered)
					writeError(w, stdhttp.StatusInternalServerError, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func requestLoggerMiddleware(logger *log.Logger) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			startedAt := time.Now()
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     stdhttp.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			logger.Printf(
				"request_id=%s method=%s path=%s status=%d duration=%s remote=%s",
				requestIDFromContext(r.Context()),
				r.Method,
				r.URL.Path,
				recorder.statusCode,
				time.Since(startedAt).Round(time.Millisecond),
				r.RemoteAddr,
			)
		})
	}
}

func requestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDContextKey).(string)
	if !ok {
		return ""
	}

	return value
}

func newRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err == nil {
		return hex.EncodeToString(buffer)
	}

	return fmt.Sprintf("fallback-%d", time.Now().UTC().UnixNano())
}
