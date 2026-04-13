package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// SearchBackend (Proxy)
// ---------------------------------------------------------------------------

type proxySearchBackend struct{ p *proxyBackends }

func (b *proxySearchBackend) SearchContent(ctx context.Context, query string, limit, offset int64) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Search.SearchContent(ctx, query, limit, offset)
}

func (b *proxySearchBackend) RebuildSearchIndex(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Search.RebuildSearchIndex(ctx)
}
