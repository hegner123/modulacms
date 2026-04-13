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
// SchemaBackend
// ---------------------------------------------------------------------------

type svcSchemaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcSchemaBackend) ListDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	if full {
		result, err := b.svc.Schema.ListDatatypesFull(ctx)
		if err != nil {
			return nil, err
		}
		return json.Marshal(result)
	}
	result, err := b.svc.Schema.ListDatatypes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) GetDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetDatatype(ctx, types.DatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) CreateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create datatype params: %w", err)
	}
	result, err := b.svc.Schema.CreateDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update datatype params: %w", err)
	}
	result, err := b.svc.Schema.UpdateDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) DeleteDatatype(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteDatatype(ctx, b.ac, types.DatatypeID(id))
}

func (b *svcSchemaBackend) ListFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Schema.ListFields(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) GetField(ctx context.Context, id string) (json.RawMessage, error) {
	// Service GetField requires roleID and isAdmin for access control.
	// In direct mode (MCP), assume admin access.
	result, err := b.svc.Schema.GetField(ctx, types.FieldID(id), "", true)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) CreateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field params: %w", err)
	}
	result, err := b.svc.Schema.CreateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field params: %w", err)
	}
	result, err := b.svc.Schema.UpdateField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) DeleteField(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteField(ctx, b.ac, types.FieldID(id))
}

func (b *svcSchemaBackend) GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetDatatypeFull(ctx, types.DatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) ListFieldTypes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Schema.ListFieldTypes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) GetFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetFieldType(ctx, types.FieldTypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) CreateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field type params: %w", err)
	}
	result, err := b.svc.Schema.CreateFieldType(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) UpdateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field type params: %w", err)
	}
	result, err := b.svc.Schema.UpdateFieldType(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcSchemaBackend) DeleteFieldType(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteFieldType(ctx, b.ac, types.FieldTypeID(id))
}

func (b *svcSchemaBackend) GetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	// Pass a zero-value nullable parent to get the global max.
	max, err := b.svc.Schema.GetMaxDatatypeSortOrder(ctx, types.NullableDatatypeID{})
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int64{"max_sort_order": max})
}

func (b *svcSchemaBackend) UpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error {
	return b.svc.Schema.UpdateDatatypeSortOrder(ctx, b.ac, db.UpdateDatatypeSortOrderParams{
		DatatypeID: types.DatatypeID(id),
		SortOrder:  sortOrder,
	})
}

func (b *svcSchemaBackend) GetFieldMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	// Pass a zero-value nullable parent to get the global max.
	max, err := b.svc.Schema.GetMaxSortOrder(ctx, types.NullableDatatypeID{})
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int64{"max_sort_order": max})
}

func (b *svcSchemaBackend) UpdateFieldSortOrder(ctx context.Context, id string, sortOrder int64) error {
	return b.svc.Schema.UpdateFieldSortOrder(ctx, b.ac, db.UpdateFieldSortOrderParams{
		FieldID:   types.FieldID(id),
		SortOrder: sortOrder,
	})
}

// ---------------------------------------------------------------------------
// AdminSchemaBackend
// ---------------------------------------------------------------------------

type svcAdminSchemaBackend struct {
	svc *service.Registry
	ac  audited.AuditContext
}

func (b *svcAdminSchemaBackend) ListAdminDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	// AdminDatatypes do not have a ListDatatypesFull equivalent.
	// The full flag is ignored; return all admin datatypes.
	result, err := b.svc.Schema.ListAdminDatatypes(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) GetAdminDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetAdminDatatype(ctx, types.AdminDatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) CreateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin datatype params: %w", err)
	}
	result, err := b.svc.Schema.CreateAdminDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin datatype params: %w", err)
	}
	result, err := b.svc.Schema.UpdateAdminDatatype(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) DeleteAdminDatatype(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteAdminDatatype(ctx, b.ac, types.AdminDatatypeID(id))
}

func (b *svcAdminSchemaBackend) ListAdminFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.svc.Schema.ListAdminFields(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) GetAdminField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.svc.Schema.GetAdminField(ctx, types.AdminFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) CreateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.CreateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin field params: %w", err)
	}
	result, err := b.svc.Schema.CreateAdminField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p db.UpdateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin field params: %w", err)
	}
	result, err := b.svc.Schema.UpdateAdminField(ctx, b.ac, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *svcAdminSchemaBackend) DeleteAdminField(ctx context.Context, id string) error {
	return b.svc.Schema.DeleteAdminField(ctx, b.ac, types.AdminFieldID(id))
}

func (b *svcAdminSchemaBackend) AdminGetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	max, err := b.svc.Schema.GetMaxAdminDatatypeSortOrder(ctx, types.NullableAdminDatatypeID{})
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int64{"max_sort_order": max})
}

func (b *svcAdminSchemaBackend) AdminUpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error {
	return b.svc.Schema.UpdateAdminDatatypeSortOrder(ctx, b.ac, db.UpdateAdminDatatypeSortOrderParams{
		AdminDatatypeID: types.AdminDatatypeID(id),
		SortOrder:       sortOrder,
	})
}
