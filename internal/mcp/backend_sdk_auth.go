package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// AuthBackend (SDK)
// ---------------------------------------------------------------------------

type sdkAuthBackend struct {
	client *modula.Client
}

func (b *sdkAuthBackend) RegisterUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateUserParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal register user params: %w", err)
	}
	result, err := b.client.Auth.Register(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAuthBackend) RequestPasswordReset(ctx context.Context, email string) (json.RawMessage, error) {
	result, err := b.client.Auth.RequestPasswordReset(ctx, modula.RequestPasswordResetParams{
		Email: email,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
