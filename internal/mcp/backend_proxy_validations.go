package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// proxyValidationBackend delegates to the current SDK client via proxyBackends.
// ---------------------------------------------------------------------------

type proxyValidationBackend struct{ p *proxyBackends }

// --- Public ---

func (b *proxyValidationBackend) ListValidations(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.ListValidations(ctx)
}

func (b *proxyValidationBackend) GetValidation(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.GetValidation(ctx, id)
}

func (b *proxyValidationBackend) CreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.CreateValidation(ctx, params)
}

func (b *proxyValidationBackend) UpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.UpdateValidation(ctx, params)
}

func (b *proxyValidationBackend) DeleteValidation(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Validations.DeleteValidation(ctx, id)
}

func (b *proxyValidationBackend) SearchValidations(ctx context.Context, query string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.SearchValidations(ctx, query)
}

// --- Admin ---

func (b *proxyValidationBackend) AdminListValidations(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.AdminListValidations(ctx)
}

func (b *proxyValidationBackend) AdminGetValidation(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.AdminGetValidation(ctx, id)
}

func (b *proxyValidationBackend) AdminCreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.AdminCreateValidation(ctx, params)
}

func (b *proxyValidationBackend) AdminUpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.AdminUpdateValidation(ctx, params)
}

func (b *proxyValidationBackend) AdminDeleteValidation(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Validations.AdminDeleteValidation(ctx, id)
}

func (b *proxyValidationBackend) AdminSearchValidations(ctx context.Context, query string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Validations.AdminSearchValidations(ctx, query)
}
