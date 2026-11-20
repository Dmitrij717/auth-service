package middleware

import (
	"net/http"
	"time"

	"github.com/Dmitrij717/auth-service/internal/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.statusCode != 0 {
		return
	}

	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(p []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}

	return rw.ResponseWriter.Write(p)
}

func HTTPMetricMiddleware(
	m *metrics.Metrics,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w}
		start := time.Now()

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		method := r.Method

		code := rw.statusCode
		if rw.statusCode == 0 {
			code = http.StatusOK
		}

		route := r.Pattern
		if route == "" {
			route = "unmatched"
		}

		m.ObserveHTTPRequest(
			method,
			route,
			code,
			duration,
		)
	})
}
