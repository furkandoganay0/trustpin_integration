package middleware

import "context"

type ctxKey string

const (
	ctxKeyTenantID  ctxKey = "tenant_id"
	ctxKeyUserID    ctxKey = "user_id"
	ctxKeyRequestID ctxKey = "request_id"
)

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, ctxKeyTenantID, tenantID)
}

func TenantID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyTenantID).(string)
	return v, ok
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, userID)
}

func UserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyUserID).(string)
	return v, ok
}

func WithRequestID(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID, rid)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyRequestID).(string)
	return v, ok
}
