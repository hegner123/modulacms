package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// AuthBackend (Proxy)
// ---------------------------------------------------------------------------

type proxyAuthBackend struct{ p *proxyBackends }

func (b *proxyAuthBackend) RegisterUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Auth.RegisterUser(ctx, params)
}

func (b *proxyAuthBackend) RequestPasswordReset(ctx context.Context, email string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Auth.RequestPasswordReset(ctx, email)
}
