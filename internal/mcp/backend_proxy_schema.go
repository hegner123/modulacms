package mcp

import (
	"context"
	"encoding/json"
)

// ---------------------------------------------------------------------------
// Schema
// ---------------------------------------------------------------------------

type proxySchemaBackend struct{ p *proxyBackends }

func (b *proxySchemaBackend) ListDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.ListDatatypes(ctx, full)
}
func (b *proxySchemaBackend) GetDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetDatatype(ctx, id)
}
func (b *proxySchemaBackend) CreateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.CreateDatatype(ctx, params)
}
func (b *proxySchemaBackend) UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.UpdateDatatype(ctx, params)
}
func (b *proxySchemaBackend) DeleteDatatype(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.DeleteDatatype(ctx, id)
}
func (b *proxySchemaBackend) ListFields(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.ListFields(ctx)
}
func (b *proxySchemaBackend) GetField(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetField(ctx, id)
}
func (b *proxySchemaBackend) CreateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.CreateField(ctx, params)
}
func (b *proxySchemaBackend) UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.UpdateField(ctx, params)
}
func (b *proxySchemaBackend) DeleteField(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.DeleteField(ctx, id)
}
func (b *proxySchemaBackend) GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetDatatypeFull(ctx, id)
}
func (b *proxySchemaBackend) ListFieldTypes(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.ListFieldTypes(ctx)
}
func (b *proxySchemaBackend) GetFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetFieldType(ctx, id)
}
func (b *proxySchemaBackend) CreateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.CreateFieldType(ctx, params)
}
func (b *proxySchemaBackend) UpdateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.UpdateFieldType(ctx, params)
}
func (b *proxySchemaBackend) DeleteFieldType(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.DeleteFieldType(ctx, id)
}
func (b *proxySchemaBackend) GetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetDatatypeMaxSortOrder(ctx)
}
func (b *proxySchemaBackend) UpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.UpdateDatatypeSortOrder(ctx, id, sortOrder)
}
func (b *proxySchemaBackend) GetFieldMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.Schema.GetFieldMaxSortOrder(ctx)
}
func (b *proxySchemaBackend) UpdateFieldSortOrder(ctx context.Context, id string, sortOrder int64) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.Schema.UpdateFieldSortOrder(ctx, id, sortOrder)
}

// ---------------------------------------------------------------------------
// AdminSchema
// ---------------------------------------------------------------------------

type proxyAdminSchemaBackend struct{ p *proxyBackends }

func (b *proxyAdminSchemaBackend) ListAdminDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.ListAdminDatatypes(ctx, full)
}
func (b *proxyAdminSchemaBackend) GetAdminDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.GetAdminDatatype(ctx, id)
}
func (b *proxyAdminSchemaBackend) CreateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.CreateAdminDatatype(ctx, params)
}
func (b *proxyAdminSchemaBackend) UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.UpdateAdminDatatype(ctx, params)
}
func (b *proxyAdminSchemaBackend) DeleteAdminDatatype(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminSchema.DeleteAdminDatatype(ctx, id)
}
func (b *proxyAdminSchemaBackend) ListAdminFields(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.ListAdminFields(ctx)
}
func (b *proxyAdminSchemaBackend) GetAdminField(ctx context.Context, id string) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.GetAdminField(ctx, id)
}
func (b *proxyAdminSchemaBackend) CreateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.CreateAdminField(ctx, params)
}
func (b *proxyAdminSchemaBackend) UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.UpdateAdminField(ctx, params)
}
func (b *proxyAdminSchemaBackend) DeleteAdminField(ctx context.Context, id string) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminSchema.DeleteAdminField(ctx, id)
}
func (b *proxyAdminSchemaBackend) AdminGetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	be, err := b.p.backends()
	if err != nil {
		return nil, err
	}
	return be.AdminSchema.AdminGetDatatypeMaxSortOrder(ctx)
}
func (b *proxyAdminSchemaBackend) AdminUpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error {
	be, err := b.p.backends()
	if err != nil {
		return err
	}
	return be.AdminSchema.AdminUpdateDatatypeSortOrder(ctx, id, sortOrder)
}
