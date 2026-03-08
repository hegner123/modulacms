package service

import (
	"context"
	"fmt"
	"time"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// ---------------------------------------------------------------------------
// Store sub-interfaces (composable, testable)
// ---------------------------------------------------------------------------

// DatatypeStore defines database operations for public datatypes.
type DatatypeStore interface {
	ListDatatypes() (*[]db.Datatypes, error)
	GetDatatype(types.DatatypeID) (*db.Datatypes, error)
	CreateDatatype(context.Context, audited.AuditContext, db.CreateDatatypeParams) (*db.Datatypes, error)
	UpdateDatatype(context.Context, audited.AuditContext, db.UpdateDatatypeParams) (*string, error)
	DeleteDatatype(context.Context, audited.AuditContext, types.DatatypeID) error
	ListDatatypesPaginated(db.PaginationParams) (*[]db.Datatypes, error)
	CountDatatypes() (*int64, error)
	UpdateDatatypeSortOrder(context.Context, audited.AuditContext, db.UpdateDatatypeSortOrderParams) error
	GetMaxDatatypeSortOrder(types.NullableDatatypeID) (int64, error)
}

// FieldStore defines database operations for public fields.
type FieldStore interface {
	ListFields() (*[]db.Fields, error)
	GetField(types.FieldID) (*db.Fields, error)
	CreateField(context.Context, audited.AuditContext, db.CreateFieldParams) (*db.Fields, error)
	UpdateField(context.Context, audited.AuditContext, db.UpdateFieldParams) (*string, error)
	DeleteField(context.Context, audited.AuditContext, types.FieldID) error
	ListFieldsPaginated(db.PaginationParams) (*[]db.Fields, error)
	CountFields() (*int64, error)
	ListFieldsByDatatypeID(types.NullableDatatypeID) (*[]db.Fields, error)
	UpdateFieldSortOrder(context.Context, audited.AuditContext, db.UpdateFieldSortOrderParams) error
	GetMaxSortOrderByParentID(types.NullableDatatypeID) (int64, error)
}

// FieldTypeStore defines database operations for public field types.
type FieldTypeStore interface {
	ListFieldTypes() (*[]db.FieldTypes, error)
	GetFieldType(types.FieldTypeID) (*db.FieldTypes, error)
	GetFieldTypeByType(string) (*db.FieldTypes, error)
	CreateFieldType(context.Context, audited.AuditContext, db.CreateFieldTypeParams) (*db.FieldTypes, error)
	UpdateFieldType(context.Context, audited.AuditContext, db.UpdateFieldTypeParams) (*string, error)
	DeleteFieldType(context.Context, audited.AuditContext, types.FieldTypeID) error
}

// AdminDatatypeStore defines database operations for admin datatypes.
type AdminDatatypeStore interface {
	ListAdminDatatypes() (*[]db.AdminDatatypes, error)
	GetAdminDatatypeById(types.AdminDatatypeID) (*db.AdminDatatypes, error)
	CreateAdminDatatype(context.Context, audited.AuditContext, db.CreateAdminDatatypeParams) (*db.AdminDatatypes, error)
	UpdateAdminDatatype(context.Context, audited.AuditContext, db.UpdateAdminDatatypeParams) (*string, error)
	DeleteAdminDatatype(context.Context, audited.AuditContext, types.AdminDatatypeID) error
	ListAdminDatatypesPaginated(db.PaginationParams) (*[]db.AdminDatatypes, error)
	CountAdminDatatypes() (*int64, error)
	UpdateAdminDatatypeSortOrder(context.Context, audited.AuditContext, db.UpdateAdminDatatypeSortOrderParams) error
	GetMaxAdminDatatypeSortOrder(types.NullableAdminDatatypeID) (int64, error)
}

// AdminFieldStore defines database operations for admin fields.
type AdminFieldStore interface {
	ListAdminFields() (*[]db.AdminFields, error)
	GetAdminField(types.AdminFieldID) (*db.AdminFields, error)
	CreateAdminField(context.Context, audited.AuditContext, db.CreateAdminFieldParams) (*db.AdminFields, error)
	UpdateAdminField(context.Context, audited.AuditContext, db.UpdateAdminFieldParams) (*string, error)
	DeleteAdminField(context.Context, audited.AuditContext, types.AdminFieldID) error
	ListAdminFieldsPaginated(db.PaginationParams) (*[]db.AdminFields, error)
	CountAdminFields() (*int64, error)
}

// AdminFieldTypeStore defines database operations for admin field types.
type AdminFieldTypeStore interface {
	ListAdminFieldTypes() (*[]db.AdminFieldTypes, error)
	GetAdminFieldType(types.AdminFieldTypeID) (*db.AdminFieldTypes, error)
	GetAdminFieldTypeByType(string) (*db.AdminFieldTypes, error)
	CreateAdminFieldType(context.Context, audited.AuditContext, db.CreateAdminFieldTypeParams) (*db.AdminFieldTypes, error)
	UpdateAdminFieldType(context.Context, audited.AuditContext, db.UpdateAdminFieldTypeParams) (*string, error)
	DeleteAdminFieldType(context.Context, audited.AuditContext, types.AdminFieldTypeID) error
}

// SchemaStore composes all sub-interfaces into a single store contract.
// The DbDriver satisfies this implicitly.
type SchemaStore interface {
	DatatypeStore
	FieldStore
	FieldTypeStore
	AdminDatatypeStore
	AdminFieldStore
	AdminFieldTypeStore
}

// ---------------------------------------------------------------------------
// SchemaService
// ---------------------------------------------------------------------------

// SchemaService manages datatypes, fields, field types, and their admin
// variants. It centralises validation, default-value assignment, and error
// wrapping that was previously duplicated across admin panel and REST API
// handlers.
type SchemaService struct {
	store      SchemaStore
	fullDriver db.DbDriver
}

// NewSchemaService creates a SchemaService. Both arguments will typically be
// the same DbDriver instance; fullDriver is kept separate because
// AssembleDatatypeFullView needs methods outside SchemaStore's scope.
func NewSchemaService(store SchemaStore, driver db.DbDriver) *SchemaService {
	return &SchemaService{store: store, fullDriver: driver}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// nilSafeSlice converts a *[]T to []T, returning an empty non-nil slice when
// the pointer is nil.
func nilSafeSlice[T any](p *[]T) []T {
	if p == nil {
		return []T{}
	}
	return *p
}

// derefCount safely dereferences a *int64 count pointer, returning 0 for nil.
func derefCount(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// nowUTC returns a types.Timestamp for the current moment in UTC.
func nowUTC() types.Timestamp {
	return types.NewTimestamp(time.Now().UTC())
}

// ---------------------------------------------------------------------------
// Public Datatypes
// ---------------------------------------------------------------------------

// ListDatatypes returns all public datatypes.
func (s *SchemaService) ListDatatypes(ctx context.Context) ([]db.Datatypes, error) {
	items, err := s.store.ListDatatypes()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list datatypes: %w", err)}
	}
	return nilSafeSlice(items), nil
}

// GetDatatype returns a single public datatype by ID.
func (s *SchemaService) GetDatatype(ctx context.Context, id types.DatatypeID) (*db.Datatypes, error) {
	dt, err := s.store.GetDatatype(id)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("get datatype: %w", err)}
	}
	if dt == nil {
		return nil, &NotFoundError{Resource: "datatype", ID: id.String()}
	}
	return dt, nil
}

// GetDatatypeFull returns a datatype with all its field definitions.
func (s *SchemaService) GetDatatypeFull(ctx context.Context, id types.DatatypeID) (*db.DatatypeFullView, error) {
	view, err := db.AssembleDatatypeFullView(s.fullDriver, id)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("get datatype full: %w", err)}
	}
	if view == nil {
		return nil, &NotFoundError{Resource: "datatype", ID: id.String()}
	}
	return view, nil
}

// CreateDatatype validates, sets defaults, and creates a public datatype.
func (s *SchemaService) CreateDatatype(ctx context.Context, ac audited.AuditContext, params db.CreateDatatypeParams) (*db.Datatypes, error) {
	var ve ValidationError
	if params.Label == "" {
		ve.Add("label", "Label is required")
	}
	if params.Type == "" {
		ve.Add("type", "Type is required")
	} else if err := types.ValidateUserDatatypeType(params.Type); err != nil {
		ve.Add("type", err.Error())
	}
	if ve.HasErrors() {
		return nil, &ve
	}

	if params.DatatypeID.IsZero() {
		params.DatatypeID = types.NewDatatypeID()
	}
	now := nowUTC()
	if !params.DateCreated.Valid {
		params.DateCreated = now
	}
	if !params.DateModified.Valid {
		params.DateModified = now
	}

	dt, err := s.store.CreateDatatype(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create datatype: %w", err)}
	}
	return dt, nil
}

// UpdateDatatype validates, preserves immutable fields, updates, and returns
// the refreshed entity.
func (s *SchemaService) UpdateDatatype(ctx context.Context, ac audited.AuditContext, params db.UpdateDatatypeParams) (*db.Datatypes, error) {
	if params.Type != "" {
		if err := types.ValidateUserDatatypeType(params.Type); err != nil {
			return nil, NewValidationError("type", err.Error())
		}
	}

	existing, err := s.store.GetDatatype(params.DatatypeID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch existing datatype: %w", err)}
	}
	if existing == nil {
		return nil, &NotFoundError{Resource: "datatype", ID: params.DatatypeID.String()}
	}

	// Preserve immutable fields.
	params.ParentID = existing.ParentID
	params.SortOrder = existing.SortOrder
	params.AuthorID = existing.AuthorID
	params.DateCreated = existing.DateCreated
	params.DateModified = nowUTC()

	if _, err = s.store.UpdateDatatype(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update datatype: %w", err)}
	}

	updated, err := s.store.GetDatatype(params.DatatypeID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated datatype: %w", err)}
	}
	if updated == nil {
		return nil, &NotFoundError{Resource: "datatype", ID: params.DatatypeID.String()}
	}
	return updated, nil
}

// UpdateDatatypeSortOrder updates the sort order for a specific datatype.
func (s *SchemaService) UpdateDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, params db.UpdateDatatypeSortOrderParams) error {
	if err := params.DatatypeID.Validate(); err != nil {
		return NewValidationError("datatype_id", err.Error())
	}
	if err := s.store.UpdateDatatypeSortOrder(ctx, ac, params); err != nil {
		return &InternalError{Err: fmt.Errorf("update datatype sort order: %w", err)}
	}
	return nil
}

// GetMaxDatatypeSortOrder returns the maximum sort order for datatypes under a parent.
func (s *SchemaService) GetMaxDatatypeSortOrder(ctx context.Context, parentID types.NullableDatatypeID) (int64, error) {
	max, err := s.store.GetMaxDatatypeSortOrder(parentID)
	if err != nil {
		return 0, &InternalError{Err: fmt.Errorf("get max datatype sort order: %w", err)}
	}
	return max, nil
}

// DeleteDatatype removes a public datatype by ID.
func (s *SchemaService) DeleteDatatype(ctx context.Context, ac audited.AuditContext, id types.DatatypeID) error {
	if err := id.Validate(); err != nil {
		return NewValidationError("datatype_id", err.Error())
	}
	if err := s.store.DeleteDatatype(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete datatype: %w", err)}
	}
	return nil
}

// ListDatatypesPaginated returns a paginated list of public datatypes.
func (s *SchemaService) ListDatatypesPaginated(ctx context.Context, p db.PaginationParams) (db.PaginatedResponse[db.Datatypes], error) {
	items, err := s.store.ListDatatypesPaginated(p)
	if err != nil {
		return db.PaginatedResponse[db.Datatypes]{}, &InternalError{Err: fmt.Errorf("list datatypes paginated: %w", err)}
	}

	total, err := s.store.CountDatatypes()
	if err != nil {
		return db.PaginatedResponse[db.Datatypes]{}, &InternalError{Err: fmt.Errorf("count datatypes: %w", err)}
	}

	return db.PaginatedResponse[db.Datatypes]{
		Data:   nilSafeSlice(items),
		Total:  derefCount(total),
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}

// ---------------------------------------------------------------------------
// Public Fields
// ---------------------------------------------------------------------------

// ListFields returns all public fields.
func (s *SchemaService) ListFields(ctx context.Context) ([]db.Fields, error) {
	items, err := s.store.ListFields()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list fields: %w", err)}
	}
	return nilSafeSlice(items), nil
}

// ListFieldsFiltered returns public fields filtered by role access.
func (s *SchemaService) ListFieldsFiltered(ctx context.Context, roleID string, isAdmin bool) ([]db.Fields, error) {
	items, err := s.store.ListFields()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list fields: %w", err)}
	}
	return db.FilterFieldsByRole(nilSafeSlice(items), roleID, isAdmin), nil
}

// GetField returns a single field, checking role-based access.
func (s *SchemaService) GetField(ctx context.Context, id types.FieldID, roleID string, isAdmin bool) (*db.Fields, error) {
	f, err := s.store.GetField(id)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("get field: %w", err)}
	}
	if f == nil {
		return nil, &NotFoundError{Resource: "field", ID: id.String()}
	}
	if !db.IsFieldAccessible(*f, roleID, isAdmin) {
		return nil, &ForbiddenError{Message: fmt.Sprintf("field %s not accessible to role %s", id, roleID)}
	}
	return f, nil
}

// CreateField validates, sets defaults, and creates a public field.
func (s *SchemaService) CreateField(ctx context.Context, ac audited.AuditContext, params db.CreateFieldParams) (*db.Fields, error) {
	var ve ValidationError
	if params.Label == "" {
		ve.Add("label", "Label is required")
	}
	if params.Type == "" {
		ve.Add("type", "Type is required")
	} else if _, lookupErr := s.store.GetFieldTypeByType(string(params.Type)); lookupErr != nil {
		ve.Add("type", "Invalid field type")
	}

	// Validate validation config JSON.
	if params.Validation != "" && params.Validation != types.EmptyJSON {
		vc, vcErr := types.ParseValidationConfig(params.Validation)
		if vcErr != nil {
			ve.Add("validation", vcErr.Error())
		} else if vcValErr := types.ValidateValidationConfig(vc); vcValErr != nil {
			ve.Add("validation", vcValErr.Error())
		}
	}

	if ve.HasErrors() {
		return nil, &ve
	}

	// Defaults.
	if params.FieldID.IsZero() {
		params.FieldID = types.NewFieldID()
	}
	now := nowUTC()
	if !params.DateCreated.Valid {
		params.DateCreated = now
	}
	if !params.DateModified.Valid {
		params.DateModified = now
	}
	if params.Data == "" {
		params.Data = types.EmptyJSON
	}
	if params.Validation == "" {
		params.Validation = types.EmptyJSON
	}
	if params.UIConfig == "" {
		params.UIConfig = types.EmptyJSON
	}

	f, err := s.store.CreateField(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create field: %w", err)}
	}
	return f, nil
}

// UpdateField validates, preserves immutable fields, updates, and returns
// the refreshed entity.
func (s *SchemaService) UpdateField(ctx context.Context, ac audited.AuditContext, params db.UpdateFieldParams) (*db.Fields, error) {
	var ve ValidationError
	if params.Label == "" {
		ve.Add("label", "Label is required")
	}
	if params.Type == "" {
		ve.Add("type", "Type is required")
	} else if _, lookupErr := s.store.GetFieldTypeByType(string(params.Type)); lookupErr != nil {
		ve.Add("type", "Invalid field type")
	}

	// Validate validation config JSON.
	if params.Validation != "" && params.Validation != types.EmptyJSON {
		vc, vcErr := types.ParseValidationConfig(params.Validation)
		if vcErr != nil {
			ve.Add("validation", vcErr.Error())
		} else if vcValErr := types.ValidateValidationConfig(vc); vcValErr != nil {
			ve.Add("validation", vcValErr.Error())
		}
	}

	if ve.HasErrors() {
		return nil, &ve
	}

	existing, err := s.store.GetField(params.FieldID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch existing field: %w", err)}
	}
	if existing == nil {
		return nil, &NotFoundError{Resource: "field", ID: params.FieldID.String()}
	}

	// Preserve immutable fields.
	params.ParentID = existing.ParentID
	params.AuthorID = existing.AuthorID
	params.SortOrder = existing.SortOrder
	params.DateCreated = existing.DateCreated
	params.DateModified = nowUTC()

	// Default empty JSON.
	if params.Data == "" {
		params.Data = types.EmptyJSON
	}
	if params.Validation == "" {
		params.Validation = types.EmptyJSON
	}
	if params.UIConfig == "" {
		params.UIConfig = types.EmptyJSON
	}

	if _, err = s.store.UpdateField(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update field: %w", err)}
	}

	updated, err := s.store.GetField(params.FieldID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated field: %w", err)}
	}
	if updated == nil {
		return nil, &NotFoundError{Resource: "field", ID: params.FieldID.String()}
	}
	return updated, nil
}

// DeleteField removes a public field by ID.
func (s *SchemaService) DeleteField(ctx context.Context, ac audited.AuditContext, id types.FieldID) error {
	if err := id.Validate(); err != nil {
		return NewValidationError("field_id", err.Error())
	}
	if err := s.store.DeleteField(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete field: %w", err)}
	}
	return nil
}

// ListFieldsPaginated returns a paginated, role-filtered list of public fields.
func (s *SchemaService) ListFieldsPaginated(ctx context.Context, p db.PaginationParams, roleID string, isAdmin bool) (db.PaginatedResponse[db.Fields], error) {
	items, err := s.store.ListFieldsPaginated(p)
	if err != nil {
		return db.PaginatedResponse[db.Fields]{}, &InternalError{Err: fmt.Errorf("list fields paginated: %w", err)}
	}

	total, err := s.store.CountFields()
	if err != nil {
		return db.PaginatedResponse[db.Fields]{}, &InternalError{Err: fmt.Errorf("count fields: %w", err)}
	}

	filtered := db.FilterFieldsByRole(nilSafeSlice(items), roleID, isAdmin)

	return db.PaginatedResponse[db.Fields]{
		Data:   filtered,
		Total:  derefCount(total),
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}

// UpdateFieldSortOrder updates the sort order for a specific field.
func (s *SchemaService) UpdateFieldSortOrder(ctx context.Context, ac audited.AuditContext, params db.UpdateFieldSortOrderParams) error {
	if err := params.FieldID.Validate(); err != nil {
		return NewValidationError("field_id", err.Error())
	}
	if err := s.store.UpdateFieldSortOrder(ctx, ac, params); err != nil {
		return &InternalError{Err: fmt.Errorf("update field sort order: %w", err)}
	}
	return nil
}

// GetMaxSortOrder returns the maximum sort order for fields under a parent.
func (s *SchemaService) GetMaxSortOrder(ctx context.Context, parentID types.NullableDatatypeID) (int64, error) {
	max, err := s.store.GetMaxSortOrderByParentID(parentID)
	if err != nil {
		return 0, &InternalError{Err: fmt.Errorf("get max sort order: %w", err)}
	}
	return max, nil
}

// ListFieldsByDatatypeID returns fields belonging to a specific datatype.
func (s *SchemaService) ListFieldsByDatatypeID(ctx context.Context, parentID types.NullableDatatypeID) ([]db.Fields, error) {
	items, err := s.store.ListFieldsByDatatypeID(parentID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list fields by datatype: %w", err)}
	}
	return nilSafeSlice(items), nil
}

// ---------------------------------------------------------------------------
// Public Field Types
// ---------------------------------------------------------------------------

// ListFieldTypes returns all public field types.
func (s *SchemaService) ListFieldTypes(ctx context.Context) ([]db.FieldTypes, error) {
	items, err := s.store.ListFieldTypes()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list field types: %w", err)}
	}
	return nilSafeSlice(items), nil
}

// GetFieldType returns a single public field type by ID.
func (s *SchemaService) GetFieldType(ctx context.Context, id types.FieldTypeID) (*db.FieldTypes, error) {
	ft, err := s.store.GetFieldType(id)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("get field type: %w", err)}
	}
	if ft == nil {
		return nil, &NotFoundError{Resource: "field_type", ID: id.String()}
	}
	return ft, nil
}

// CreateFieldType creates a public field type.
func (s *SchemaService) CreateFieldType(ctx context.Context, ac audited.AuditContext, params db.CreateFieldTypeParams) (*db.FieldTypes, error) {
	ft, err := s.store.CreateFieldType(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create field type: %w", err)}
	}
	return ft, nil
}

// UpdateFieldType updates a field type and returns the refreshed entity.
func (s *SchemaService) UpdateFieldType(ctx context.Context, ac audited.AuditContext, params db.UpdateFieldTypeParams) (*db.FieldTypes, error) {
	if _, err := s.store.UpdateFieldType(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update field type: %w", err)}
	}

	updated, err := s.store.GetFieldType(params.FieldTypeID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated field type: %w", err)}
	}
	if updated == nil {
		return nil, &NotFoundError{Resource: "field_type", ID: params.FieldTypeID.String()}
	}
	return updated, nil
}

// DeleteFieldType removes a public field type by ID.
func (s *SchemaService) DeleteFieldType(ctx context.Context, ac audited.AuditContext, id types.FieldTypeID) error {
	if err := id.Validate(); err != nil {
		return NewValidationError("field_type_id", err.Error())
	}
	if err := s.store.DeleteFieldType(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete field type: %w", err)}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Admin Datatypes
// ---------------------------------------------------------------------------

// ListAdminDatatypes returns all admin datatypes.
func (s *SchemaService) ListAdminDatatypes(ctx context.Context) ([]db.AdminDatatypes, error) {
	items, err := s.store.ListAdminDatatypes()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list admin datatypes: %w", err)}
	}
	return nilSafeSlice(items), nil
}

// GetAdminDatatype returns a single admin datatype by ID.
func (s *SchemaService) GetAdminDatatype(ctx context.Context, id types.AdminDatatypeID) (*db.AdminDatatypes, error) {
	dt, err := s.store.GetAdminDatatypeById(id)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("get admin datatype: %w", err)}
	}
	if dt == nil {
		return nil, &NotFoundError{Resource: "admin_datatype", ID: id.String()}
	}
	return dt, nil
}

// CreateAdminDatatype validates and creates an admin datatype.
// Admin datatypes skip ValidateUserDatatypeType.
func (s *SchemaService) CreateAdminDatatype(ctx context.Context, ac audited.AuditContext, params db.CreateAdminDatatypeParams) (*db.AdminDatatypes, error) {
	var ve ValidationError
	if params.Label == "" {
		ve.Add("label", "Label is required")
	}
	if params.Type == "" {
		ve.Add("type", "Type is required")
	}
	if ve.HasErrors() {
		return nil, &ve
	}

	now := nowUTC()
	if !params.DateCreated.Valid {
		params.DateCreated = now
	}
	if !params.DateModified.Valid {
		params.DateModified = now
	}

	dt, err := s.store.CreateAdminDatatype(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create admin datatype: %w", err)}
	}
	return dt, nil
}

// UpdateAdminDatatype updates an admin datatype and returns the refreshed entity.
func (s *SchemaService) UpdateAdminDatatype(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminDatatypeParams) (*db.AdminDatatypes, error) {
	existing, err := s.store.GetAdminDatatypeById(params.AdminDatatypeID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch existing admin datatype: %w", err)}
	}
	if existing == nil {
		return nil, &NotFoundError{Resource: "admin_datatype", ID: params.AdminDatatypeID.String()}
	}

	params.ParentID = existing.ParentID
	params.SortOrder = existing.SortOrder
	params.AuthorID = existing.AuthorID
	params.DateCreated = existing.DateCreated
	params.DateModified = nowUTC()

	if _, err = s.store.UpdateAdminDatatype(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update admin datatype: %w", err)}
	}

	updated, err := s.store.GetAdminDatatypeById(params.AdminDatatypeID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated admin datatype: %w", err)}
	}
	if updated == nil {
		return nil, &NotFoundError{Resource: "admin_datatype", ID: params.AdminDatatypeID.String()}
	}
	return updated, nil
}

// UpdateAdminDatatypeSortOrder updates the sort order for a specific admin datatype.
func (s *SchemaService) UpdateAdminDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminDatatypeSortOrderParams) error {
	if err := params.AdminDatatypeID.Validate(); err != nil {
		return NewValidationError("admin_datatype_id", err.Error())
	}
	if err := s.store.UpdateAdminDatatypeSortOrder(ctx, ac, params); err != nil {
		return &InternalError{Err: fmt.Errorf("update admin datatype sort order: %w", err)}
	}
	return nil
}

// GetMaxAdminDatatypeSortOrder returns the maximum sort order for admin datatypes under a parent.
func (s *SchemaService) GetMaxAdminDatatypeSortOrder(ctx context.Context, parentID types.NullableAdminDatatypeID) (int64, error) {
	max, err := s.store.GetMaxAdminDatatypeSortOrder(parentID)
	if err != nil {
		return 0, &InternalError{Err: fmt.Errorf("get max admin datatype sort order: %w", err)}
	}
	return max, nil
}

// DeleteAdminDatatype removes an admin datatype by ID.
func (s *SchemaService) DeleteAdminDatatype(ctx context.Context, ac audited.AuditContext, id types.AdminDatatypeID) error {
	if err := id.Validate(); err != nil {
		return NewValidationError("admin_datatype_id", err.Error())
	}
	if err := s.store.DeleteAdminDatatype(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete admin datatype: %w", err)}
	}
	return nil
}

// ListAdminDatatypesPaginated returns a paginated list of admin datatypes.
func (s *SchemaService) ListAdminDatatypesPaginated(ctx context.Context, p db.PaginationParams) (db.PaginatedResponse[db.AdminDatatypes], error) {
	items, err := s.store.ListAdminDatatypesPaginated(p)
	if err != nil {
		return db.PaginatedResponse[db.AdminDatatypes]{}, &InternalError{Err: fmt.Errorf("list admin datatypes paginated: %w", err)}
	}

	total, err := s.store.CountAdminDatatypes()
	if err != nil {
		return db.PaginatedResponse[db.AdminDatatypes]{}, &InternalError{Err: fmt.Errorf("count admin datatypes: %w", err)}
	}

	return db.PaginatedResponse[db.AdminDatatypes]{
		Data:   nilSafeSlice(items),
		Total:  derefCount(total),
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}

// ---------------------------------------------------------------------------
// Admin Fields
// ---------------------------------------------------------------------------

// ListAdminFields returns all admin fields.
func (s *SchemaService) ListAdminFields(ctx context.Context) ([]db.AdminFields, error) {
	items, err := s.store.ListAdminFields()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list admin fields: %w", err)}
	}
	return nilSafeSlice(items), nil
}

// GetAdminField returns a single admin field by ID.
func (s *SchemaService) GetAdminField(ctx context.Context, id types.AdminFieldID) (*db.AdminFields, error) {
	f, err := s.store.GetAdminField(id)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("get admin field: %w", err)}
	}
	if f == nil {
		return nil, &NotFoundError{Resource: "admin_field", ID: id.String()}
	}
	return f, nil
}

// CreateAdminField validates, sets defaults, and creates an admin field.
func (s *SchemaService) CreateAdminField(ctx context.Context, ac audited.AuditContext, params db.CreateAdminFieldParams) (*db.AdminFields, error) {
	var ve ValidationError
	if params.Label == "" {
		ve.Add("label", "Label is required")
	}
	if params.Type == "" {
		ve.Add("type", "Type is required")
	} else if _, lookupErr := s.store.GetAdminFieldTypeByType(string(params.Type)); lookupErr != nil {
		ve.Add("type", "Invalid field type")
	}
	if ve.HasErrors() {
		return nil, &ve
	}

	now := nowUTC()
	if !params.DateCreated.Valid {
		params.DateCreated = now
	}
	if !params.DateModified.Valid {
		params.DateModified = now
	}
	if params.Validation == "" {
		params.Validation = types.EmptyJSON
	}
	if params.UIConfig == "" {
		params.UIConfig = types.EmptyJSON
	}

	f, err := s.store.CreateAdminField(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create admin field: %w", err)}
	}
	return f, nil
}

// UpdateAdminField validates, preserves immutable fields, updates, and returns
// the refreshed entity.
func (s *SchemaService) UpdateAdminField(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminFieldParams) (*db.AdminFields, error) {
	existing, err := s.store.GetAdminField(params.AdminFieldID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch existing admin field: %w", err)}
	}
	if existing == nil {
		return nil, &NotFoundError{Resource: "admin_field", ID: params.AdminFieldID.String()}
	}

	params.ParentID = existing.ParentID
	params.AuthorID = existing.AuthorID
	params.DateCreated = existing.DateCreated
	params.DateModified = nowUTC()

	if params.Validation == "" {
		params.Validation = types.EmptyJSON
	}
	if params.UIConfig == "" {
		params.UIConfig = types.EmptyJSON
	}

	if _, err = s.store.UpdateAdminField(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update admin field: %w", err)}
	}

	updated, err := s.store.GetAdminField(params.AdminFieldID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated admin field: %w", err)}
	}
	if updated == nil {
		return nil, &NotFoundError{Resource: "admin_field", ID: params.AdminFieldID.String()}
	}
	return updated, nil
}

// DeleteAdminField removes an admin field by ID.
func (s *SchemaService) DeleteAdminField(ctx context.Context, ac audited.AuditContext, id types.AdminFieldID) error {
	if err := id.Validate(); err != nil {
		return NewValidationError("admin_field_id", err.Error())
	}
	if err := s.store.DeleteAdminField(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete admin field: %w", err)}
	}
	return nil
}

// ListAdminFieldsPaginated returns a paginated list of admin fields.
func (s *SchemaService) ListAdminFieldsPaginated(ctx context.Context, p db.PaginationParams) (db.PaginatedResponse[db.AdminFields], error) {
	items, err := s.store.ListAdminFieldsPaginated(p)
	if err != nil {
		return db.PaginatedResponse[db.AdminFields]{}, &InternalError{Err: fmt.Errorf("list admin fields paginated: %w", err)}
	}

	total, err := s.store.CountAdminFields()
	if err != nil {
		return db.PaginatedResponse[db.AdminFields]{}, &InternalError{Err: fmt.Errorf("count admin fields: %w", err)}
	}

	return db.PaginatedResponse[db.AdminFields]{
		Data:   nilSafeSlice(items),
		Total:  derefCount(total),
		Limit:  p.Limit,
		Offset: p.Offset,
	}, nil
}

// ---------------------------------------------------------------------------
// Admin Field Types
// ---------------------------------------------------------------------------

// ListAdminFieldTypes returns all admin field types.
func (s *SchemaService) ListAdminFieldTypes(ctx context.Context) ([]db.AdminFieldTypes, error) {
	items, err := s.store.ListAdminFieldTypes()
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("list admin field types: %w", err)}
	}
	return nilSafeSlice(items), nil
}

// GetAdminFieldType returns a single admin field type by ID.
func (s *SchemaService) GetAdminFieldType(ctx context.Context, id types.AdminFieldTypeID) (*db.AdminFieldTypes, error) {
	ft, err := s.store.GetAdminFieldType(id)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("get admin field type: %w", err)}
	}
	if ft == nil {
		return nil, &NotFoundError{Resource: "admin_field_type", ID: id.String()}
	}
	return ft, nil
}

// CreateAdminFieldType creates an admin field type.
func (s *SchemaService) CreateAdminFieldType(ctx context.Context, ac audited.AuditContext, params db.CreateAdminFieldTypeParams) (*db.AdminFieldTypes, error) {
	ft, err := s.store.CreateAdminFieldType(ctx, ac, params)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("create admin field type: %w", err)}
	}
	return ft, nil
}

// UpdateAdminFieldType updates an admin field type and returns the refreshed entity.
func (s *SchemaService) UpdateAdminFieldType(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminFieldTypeParams) (*db.AdminFieldTypes, error) {
	if _, err := s.store.UpdateAdminFieldType(ctx, ac, params); err != nil {
		return nil, &InternalError{Err: fmt.Errorf("update admin field type: %w", err)}
	}

	updated, err := s.store.GetAdminFieldType(params.AdminFieldTypeID)
	if err != nil {
		return nil, &InternalError{Err: fmt.Errorf("fetch updated admin field type: %w", err)}
	}
	if updated == nil {
		return nil, &NotFoundError{Resource: "admin_field_type", ID: params.AdminFieldTypeID.String()}
	}
	return updated, nil
}

// DeleteAdminFieldType removes an admin field type by ID.
func (s *SchemaService) DeleteAdminFieldType(ctx context.Context, ac audited.AuditContext, id types.AdminFieldTypeID) error {
	if err := id.Validate(); err != nil {
		return NewValidationError("admin_field_type_id", err.Error())
	}
	if err := s.store.DeleteAdminFieldType(ctx, ac, id); err != nil {
		return &InternalError{Err: fmt.Errorf("delete admin field type: %w", err)}
	}
	return nil
}
