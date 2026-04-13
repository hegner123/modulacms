package mcp

import (
	"context"
	"encoding/json"

	"github.com/hegner123/modulacms/internal/search"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// SearchBackend (Service)
// ---------------------------------------------------------------------------

type svcSearchBackend struct {
	svc *service.Registry
}

func (b *svcSearchBackend) SearchContent(ctx context.Context, query string, limit, offset int64) (json.RawMessage, error) {
	opts := search.SearchOptions{
		Limit:  int(limit),
		Offset: int(offset),
	}
	result, err := b.svc.Search.Search(ctx, query, false, opts)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSearchBackend) RebuildSearchIndex(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Search.Rebuild(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
