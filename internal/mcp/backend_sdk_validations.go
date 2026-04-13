package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// sdkValidationBackend implements ValidationBackend via the Go SDK client.
// ---------------------------------------------------------------------------

type sdkValidationBackend struct {
	client *modula.Client
}

// --- Public ---

func (b *sdkValidationBackend) ListValidations(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Validations.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) GetValidation(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Validations.Get(ctx, modula.ValidationID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) CreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create validation params: %w", err)
	}
	result, err := b.client.Validations.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) UpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update validation params: %w", err)
	}
	result, err := b.client.Validations.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) DeleteValidation(ctx context.Context, id string) error {
	return b.client.Validations.Delete(ctx, modula.ValidationID(id))
}

func (b *sdkValidationBackend) SearchValidations(ctx context.Context, query string) (json.RawMessage, error) {
	result, err := b.client.Validations.SearchByName(ctx, query)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// --- Admin ---

func (b *sdkValidationBackend) AdminListValidations(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.AdminValidations.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) AdminGetValidation(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminValidations.Get(ctx, modula.AdminValidationID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) AdminCreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin validation params: %w", err)
	}
	result, err := b.client.AdminValidations.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) AdminUpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin validation params: %w", err)
	}
	result, err := b.client.AdminValidations.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkValidationBackend) AdminDeleteValidation(ctx context.Context, id string) error {
	return b.client.AdminValidations.Delete(ctx, modula.AdminValidationID(id))
}

func (b *sdkValidationBackend) AdminSearchValidations(ctx context.Context, query string) (json.RawMessage, error) {
	result, err := b.client.AdminValidations.SearchByName(ctx, query)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
