package mcp

import (
	"context"
	"encoding/json"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// ActivityBackend (SDK)
// ---------------------------------------------------------------------------

type sdkActivityBackend struct {
	client *modula.Client
}

func (b *sdkActivityBackend) ListRecentActivity(ctx context.Context, limit int64) (json.RawMessage, error) {
	result, err := b.client.Activity.ListRecent(ctx, int(limit))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
