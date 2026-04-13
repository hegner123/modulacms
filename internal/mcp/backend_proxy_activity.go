package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// ActivityBackend (Proxy)
// ---------------------------------------------------------------------------

type proxyActivityBackend struct{ p *proxyBackends }

func (b *proxyActivityBackend) ListRecentActivity(ctx context.Context, limit int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Activity.ListRecentActivity(ctx, limit)
}
