package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rec, r)
			rid, _ := RequestIDFromContext(r.Context())
			if rid == "" {
				// fallback to header if context doesn't have the value
				rid = r.Header.Get("X-Request-ID")
			}
			tenant, _ := TenantID(r.Context())
			if tenant == "" {
				tenant = r.Header.Get("X-Tenant-ID")
			}
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", rid,
				"tenant_id", tenant,
			)
		})
	}
}
