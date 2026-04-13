package mcp

import (
	"context"
	"encoding/json"

	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// ActivityBackend (Service)
// ---------------------------------------------------------------------------

type svcActivityBackend struct {
	svc *service.Registry
}

func (b *svcActivityBackend) ListRecentActivity(ctx context.Context, limit int64) (json.RawMessage, error) {
	result, err := b.svc.AuditLog.GetRecentActivity(ctx, limit)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
