package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// AuthBackend (Service)
// ---------------------------------------------------------------------------

type svcAuthBackend struct {
	svc *service.Registry
}

func (b *svcAuthBackend) RegisterUser(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p service.RegisterInput
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal register user params: %w", err)
	}
	result, err := b.svc.Auth.Register(ctx, AuditContextFromMCP(ctx), p)
	if err != nil {
		return nil, err
	}
	result.Hash = "" // never expose password hash
	return json.Marshal(result)
}

func (b *svcAuthBackend) RequestPasswordReset(ctx context.Context, email string) (json.RawMessage, error) {
	err := b.svc.Auth.RequestPasswordReset(ctx, AuditContextFromMCP(ctx), service.PasswordResetRequestInput{
		Email: email,
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]string{
		"message": "if an account with that email exists, a reset link has been sent",
	})
}
