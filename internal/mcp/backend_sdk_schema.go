package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	modula "github.com/hegner123/modulacms/sdks/go"
)

// ---------------------------------------------------------------------------
// SchemaBackend
// ---------------------------------------------------------------------------

type sdkSchemaBackend struct {
	client *modula.Client
}

func (b *sdkSchemaBackend) ListDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	if full {
		return b.client.Datatypes.RawList(ctx, url.Values{"full": {"true"}})
	}
	result, err := b.client.Datatypes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) GetDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Datatypes.Get(ctx, modula.DatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) CreateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create datatype params: %w", err)
	}
	result, err := b.client.Datatypes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) UpdateDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update datatype params: %w", err)
	}
	result, err := b.client.Datatypes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) DeleteDatatype(ctx context.Context, id string) error {
	return b.client.Datatypes.Delete(ctx, modula.DatatypeID(id))
}

func (b *sdkSchemaBackend) ListFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.Fields.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) GetField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.Fields.Get(ctx, modula.FieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) CreateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field params: %w", err)
	}
	result, err := b.client.Fields.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) UpdateField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field params: %w", err)
	}
	result, err := b.client.Fields.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) DeleteField(ctx context.Context, id string) error {
	return b.client.Fields.Delete(ctx, modula.FieldID(id))
}

func (b *sdkSchemaBackend) GetDatatypeFull(ctx context.Context, id string) (json.RawMessage, error) {
	return b.client.Datatypes.RawList(ctx, url.Values{"full": {"true"}, "q": {id}})
}

func (b *sdkSchemaBackend) ListFieldTypes(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.FieldTypes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) GetFieldType(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.FieldTypes.Get(ctx, modula.FieldTypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) CreateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create field type params: %w", err)
	}
	result, err := b.client.FieldTypes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) UpdateFieldType(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateFieldTypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update field type params: %w", err)
	}
	result, err := b.client.FieldTypes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkSchemaBackend) DeleteFieldType(ctx context.Context, id string) error {
	return b.client.FieldTypes.Delete(ctx, modula.FieldTypeID(id))
}

func (b *sdkSchemaBackend) GetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	max, err := b.client.DatatypesExtra.MaxSortOrder(ctx, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int64{"max_sort_order": max})
}

func (b *sdkSchemaBackend) UpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error {
	return b.client.DatatypesExtra.UpdateSortOrder(ctx, modula.DatatypeID(id), sortOrder)
}

func (b *sdkSchemaBackend) GetFieldMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	// MaxSortOrder requires a parent datatype ID. Use a zero-value to get the global max.
	max, err := b.client.FieldsExtra.MaxSortOrder(ctx, "")
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int64{"max_sort_order": max})
}

func (b *sdkSchemaBackend) UpdateFieldSortOrder(ctx context.Context, id string, sortOrder int64) error {
	return b.client.FieldsExtra.UpdateSortOrder(ctx, modula.FieldID(id), sortOrder)
}

// ---------------------------------------------------------------------------
// AdminSchemaBackend
// ---------------------------------------------------------------------------

type sdkAdminSchemaBackend struct {
	client *modula.Client
}

func (b *sdkAdminSchemaBackend) ListAdminDatatypes(ctx context.Context, full bool) (json.RawMessage, error) {
	if full {
		return b.client.AdminDatatypes.RawList(ctx, url.Values{"full": {"true"}})
	}
	result, err := b.client.AdminDatatypes.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) GetAdminDatatype(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminDatatypes.Get(ctx, modula.AdminDatatypeID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) CreateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin datatype params: %w", err)
	}
	result, err := b.client.AdminDatatypes.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) UpdateAdminDatatype(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminDatatypeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin datatype params: %w", err)
	}
	result, err := b.client.AdminDatatypes.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) DeleteAdminDatatype(ctx context.Context, id string) error {
	return b.client.AdminDatatypes.Delete(ctx, modula.AdminDatatypeID(id))
}

func (b *sdkAdminSchemaBackend) ListAdminFields(ctx context.Context) (json.RawMessage, error) {
	result, err := b.client.AdminFields.List(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) GetAdminField(ctx context.Context, id string) (json.RawMessage, error) {
	result, err := b.client.AdminFields.Get(ctx, modula.AdminFieldID(id))
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) CreateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.CreateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal create admin field params: %w", err)
	}
	result, err := b.client.AdminFields.Create(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) UpdateAdminField(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var p modula.UpdateAdminFieldParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, fmt.Errorf("unmarshal update admin field params: %w", err)
	}
	result, err := b.client.AdminFields.Update(ctx, p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (b *sdkAdminSchemaBackend) DeleteAdminField(ctx context.Context, id string) error {
	return b.client.AdminFields.Delete(ctx, modula.AdminFieldID(id))
}

func (b *sdkAdminSchemaBackend) AdminGetDatatypeMaxSortOrder(ctx context.Context) (json.RawMessage, error) {
	max, err := b.client.AdminDatatypesExtra.MaxSortOrder(ctx, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]int64{"max_sort_order": max})
}

func (b *sdkAdminSchemaBackend) AdminUpdateDatatypeSortOrder(ctx context.Context, id string, sortOrder int64) error {
	return b.client.AdminDatatypesExtra.UpdateSortOrder(ctx, modula.AdminDatatypeID(id), sortOrder)
}
