package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-ID")
		if rid == "" {
			buf := make([]byte, 16)
			_, _ = rand.Read(buf)
			rid = hex.EncodeToString(buf)
		}
		w.Header().Set("X-Request-ID", rid)
		ctx := WithRequestID(r.Context(), rid)
		r2 := r.WithContext(ctx)
		// ensure the request header also contains the request id so downstream
		// middlewares and the logger can read it from either context or headers
		r2.Header.Set("X-Request-ID", rid)
		next.ServeHTTP(w, r2)
	})
}
