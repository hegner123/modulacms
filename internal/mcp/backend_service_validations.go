package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/service"
)

// ---------------------------------------------------------------------------
// svcValidationBackend implements ValidationBackend via the service layer.
// ---------------------------------------------------------------------------

type svcValidationBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

// --- Public ---

func (b *svcValidationBackend) ListValidations(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Validations.ListValidations()
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) GetValidation(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Validations.GetValidation(types.ValidationID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) CreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create validation params: %w", err)
	}
	result, err := b.svc.Validations.CreateValidation(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) UpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update validation params: %w", err)
	}
	result, err := b.svc.Validations.UpdateValidation(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) DeleteValidation(ctx context.Context, id string) error {
	return b.svc.Validations.DeleteValidation(ctx, b.ac, types.ValidationID(id))
}

func (b *svcValidationBackend) SearchValidations(ctx context.Context, query string) (json.RawMessage, error) {
	result, err := b.svc.Validations.ListValidationsByName(query)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// --- Admin ---

func (b *svcValidationBackend) AdminListValidations(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Validations.ListAdminValidations()
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) AdminGetValidation(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Validations.GetAdminValidation(types.AdminValidationID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) AdminCreateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin validation params: %w", err)
	}
	result, err := b.svc.Validations.CreateAdminValidation(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) AdminUpdateValidation(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateAdminValidationParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin validation params: %w", err)
	}
	result, err := b.svc.Validations.UpdateAdminValidation(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcValidationBackend) AdminDeleteValidation(ctx context.Context, id string) error {
	return b.svc.Validations.DeleteAdminValidation(ctx, b.ac, types.AdminValidationID(id))
}

func (b *svcValidationBackend) AdminSearchValidations(ctx context.Context, query string) (json.RawMessage, error) {
	result, err := b.svc.Validations.ListAdminValidationsByName(query)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}
