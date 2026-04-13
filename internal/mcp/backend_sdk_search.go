package mcp

import (
	"context"
	"encoding/json"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// SearchBackend (SDK)
// ---------------------------------------------------------------------------

type sdkSearchBackend struct {
	client *modula.Client
}

func (b *sdkSearchBackend) SearchContent(ctx context.Context, query string, limit, offset int64) (json.RawMessage, error) {
	opts := &modula.SearchOptions{
		Limit:  int(limit),
		Offset: int(offset),
	}
	result, err := b.client.Search.Search(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSearchBackend) RebuildSearchIndex(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Search.Rebuild(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
