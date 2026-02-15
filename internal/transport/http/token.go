package httptransport

import (
	"context"
	"time"
)

type TokenIssuer interface {
	Issue(ctx context.Context, tenantID, userID string) (string, time.Time, error)
}
