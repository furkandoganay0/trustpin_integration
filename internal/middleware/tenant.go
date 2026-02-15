package middleware

import (
	"net/http"
	"strings"
)

func EnforceTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerTenant := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
		ctxTenant, _ := TenantID(r.Context())
		ctxTenant = strings.TrimSpace(ctxTenant)
		if headerTenant == "" || ctxTenant == "" || headerTenant != ctxTenant {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
