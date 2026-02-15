package db

import (
	"context"
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

// AdminFields represents an admin field in the CMS.
type AdminFields struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	Validation   string                        `json:"validation"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
	AuthorID     types.NullableUserID          `json:"author_id"`
	DateCreated  types.Timestamp               `json:"date_created"`
	DateModified types.Timestamp               `json:"date_modified"`
}

// CreateAdminFieldParams contains parameters for creating a new admin field.
type CreateAdminFieldParams struct {
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	Validation   string                        `json:"validation"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
	AuthorID     types.NullableUserID          `json:"author_id"`
	DateCreated  types.Timestamp               `json:"date_created"`
	DateModified types.Timestamp               `json:"date_modified"`
}

// UpdateAdminFieldParams contains parameters for updating an existing admin field.
type UpdateAdminFieldParams struct {
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	Validation   string                        `json:"validation"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
	AuthorID     types.NullableUserID          `json:"author_id"`
	DateCreated  types.Timestamp               `json:"date_created"`
	DateModified types.Timestamp               `json:"date_modified"`
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
}

// ListAdminFieldByRouteIdRow represents a result row from listing admin fields by route ID.
type ListAdminFieldByRouteIdRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	Validation   string                        `json:"validation"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
}

// ListAdminFieldsByParentIDPaginatedParams contains parameters for paginated listing of admin fields by parent ID.
type ListAdminFieldsByParentIDPaginatedParams struct {
	ParentID types.AdminDatatypeID
	Limit    int64
	Offset   int64
}

// ListAdminFieldsByDatatypeIDRow represents a result row from listing admin fields by datatype ID.
type ListAdminFieldsByDatatypeIDRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	Validation   string                        `json:"validation"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
}

// UtilityGetAdminfieldsRow represents a result row from utility admin fields retrieval.
type UtilityGetAdminfieldsRow struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	Label        string             `json:"label"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapAdminFieldJSON converts AdminFields to FieldsJSON for tree building by mapping admin field ID into the public FieldsJSON shape.
func MapAdminFieldJSON(a AdminFields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.AdminFieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UIConfig,
		Type:         a.Type.String(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringAdminField converts AdminFields to StringAdminFields for table display.
func MapStringAdminField(a AdminFields) StringAdminFields {
	return StringAdminFields{
		AdminFieldID: a.AdminFieldID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UIConfig,
		Type:         string(a.Type),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
		History:      "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapAdminField converts a sqlc-generated type to the wrapper type.
func (d Database) MapAdminField(a mdb.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UiConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateAdminFieldParams converts a sqlc-generated type to the wrapper type.
func (d Database) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdb.CreateAdminFieldParams {
	return mdb.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminFieldParams converts a sqlc-generated type to the wrapper type.
func (d Database) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdb.UpdateAdminFieldParams {
	return mdb.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

// CountAdminFields returns the total count of admin fields.
func (d Database) CountAdminFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminField inserts a new admin field record.
func (d Database) CreateAdminField(ctx context.Context, ac audited.AuditContext, s CreateAdminFieldParams) (*AdminFields, error) {
	cmd := d.NewAdminFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminField: %w", err)
	}
	r := d.MapAdminField(result)
	return &r, nil
}

// CreateAdminFieldTable creates the admin_fields table in the database.
func (d Database) CreateAdminFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

// DeleteAdminField removes an admin field record by ID.
func (d Database) DeleteAdminField(ctx context.Context, ac audited.AuditContext, id types.AdminFieldID) error {
	cmd := d.DeleteAdminFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminField retrieves an admin field by ID.
func (d Database) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdb.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

// ListAdminFields returns all admin field records.
func (d Database) ListAdminFields() (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsPaginated returns admin field records with pagination.
func (d Database) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdb.ListAdminFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsByParentIDPaginated returns admin fields with pagination filtered by parent ID.
func (d Database) ListAdminFieldsByParentIDPaginated(params ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentIDPaginated(d.Context, mdb.ListAdminFieldByParentIDPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by parent paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminField modifies an existing admin field record.
func (d Database) UpdateAdminField(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldParams) (*string, error) {
	cmd := d.UpdateAdminFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapAdminField converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapAdminField(a mdbm.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UiConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateAdminFieldParams converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbm.CreateAdminFieldParams {
	return mdbm.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminFieldParams converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbm.UpdateAdminFieldParams {
	return mdbm.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

// CountAdminFields returns the total count of admin fields.
func (d MysqlDatabase) CountAdminFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminField inserts a new admin field record.
func (d MysqlDatabase) CreateAdminField(ctx context.Context, ac audited.AuditContext, s CreateAdminFieldParams) (*AdminFields, error) {
	cmd := d.NewAdminFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminField: %w", err)
	}
	r := d.MapAdminField(result)
	return &r, nil
}

// CreateAdminFieldTable creates the admin_fields table in the database.
func (d MysqlDatabase) CreateAdminFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

// DeleteAdminField removes an admin field record by ID.
func (d MysqlDatabase) DeleteAdminField(ctx context.Context, ac audited.AuditContext, id types.AdminFieldID) error {
	cmd := d.DeleteAdminFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminField retrieves an admin field by ID.
func (d MysqlDatabase) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdbm.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

// ListAdminFields returns all admin field records.
func (d MysqlDatabase) ListAdminFields() (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsPaginated returns admin field records with pagination.
func (d MysqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbm.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsByParentIDPaginated returns admin fields with pagination filtered by parent ID.
func (d MysqlDatabase) ListAdminFieldsByParentIDPaginated(params ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentIDPaginated(d.Context, mdbm.ListAdminFieldByParentIDPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by parent paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminField modifies an existing admin field record.
func (d MysqlDatabase) UpdateAdminField(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldParams) (*string, error) {
	cmd := d.UpdateAdminFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapAdminField converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapAdminField(a mdbp.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UiConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateAdminFieldParams converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbp.CreateAdminFieldParams {
	return mdbp.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminFieldParams converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbp.UpdateAdminFieldParams {
	return mdbp.UpdateAdminFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UiConfig:     a.UIConfig,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		AdminFieldID: a.AdminFieldID,
	}
}

// QUERIES

// CountAdminFields returns the total count of admin fields.
func (d PsqlDatabase) CountAdminFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminField inserts a new admin field record.
func (d PsqlDatabase) CreateAdminField(ctx context.Context, ac audited.AuditContext, s CreateAdminFieldParams) (*AdminFields, error) {
	cmd := d.NewAdminFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminField: %w", err)
	}
	r := d.MapAdminField(result)
	return &r, nil
}

// CreateAdminFieldTable creates the admin_fields table in the database.
func (d PsqlDatabase) CreateAdminFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminFieldTable(d.Context)
	return err
}

// DeleteAdminField removes an admin field record by ID.
func (d PsqlDatabase) DeleteAdminField(ctx context.Context, ac audited.AuditContext, id types.AdminFieldID) error {
	cmd := d.DeleteAdminFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminField retrieves an admin field by ID.
func (d PsqlDatabase) GetAdminField(id types.AdminFieldID) (*AdminFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, mdbp.GetAdminFieldParams{AdminFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

// ListAdminFields returns all admin field records.
func (d PsqlDatabase) ListAdminFields() (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsPaginated returns admin field records with pagination.
func (d PsqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbp.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminFieldsByParentIDPaginated returns admin fields with pagination filtered by parent ID.
func (d PsqlDatabase) ListAdminFieldsByParentIDPaginated(params ListAdminFieldsByParentIDPaginatedParams) (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentIDPaginated(d.Context, mdbp.ListAdminFieldByParentIDPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by parent paginated: %v", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminField modifies an existing admin field record.
func (d PsqlDatabase) UpdateAdminField(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldParams) (*string, error) {
	cmd := d.UpdateAdminFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

// NewAdminFieldCmd is an audited command for create operations on admin_fields.
type NewAdminFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c NewAdminFieldCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewAdminFieldCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminFieldCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c NewAdminFieldCmd) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c NewAdminFieldCmd) Params() any { return c.params }

// GetID extracts the ID from the created record.
func (c NewAdminFieldCmd) GetID(u mdb.AdminFields) string { return string(u.AdminFieldID) }

// Execute performs the create operation.
func (c NewAdminFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminFields, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminField(ctx, mdb.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewAdminFieldCmd creates a new create command for admin fields.
func (d Database) NewAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminFieldParams) NewAdminFieldCmd {
	return NewAdminFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateAdminFieldCmd is an audited command for update operations on admin_fields.
type UpdateAdminFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c UpdateAdminFieldCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateAdminFieldCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminFieldCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c UpdateAdminFieldCmd) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c UpdateAdminFieldCmd) Params() any { return c.params }

// GetID returns the ID of the record to update.
func (c UpdateAdminFieldCmd) GetID() string { return string(c.params.AdminFieldID) }

// GetBefore retrieves the record before modification.
func (c UpdateAdminFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminField(ctx, mdb.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

// Execute performs the update operation.
func (c UpdateAdminFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminField(ctx, mdb.UpdateAdminFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		AdminFieldID: c.params.AdminFieldID,
	})
}

// UpdateAdminFieldCmd creates a new update command for admin fields.
func (d Database) UpdateAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldParams) UpdateAdminFieldCmd {
	return UpdateAdminFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteAdminFieldCmd is an audited command for delete operations on admin_fields.
type DeleteAdminFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c DeleteAdminFieldCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteAdminFieldCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminFieldCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c DeleteAdminFieldCmd) TableName() string { return "admin_fields" }

// GetID returns the ID of the record to delete.
func (c DeleteAdminFieldCmd) GetID() string { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminField(ctx, mdb.GetAdminFieldParams{AdminFieldID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminField(ctx, mdb.DeleteAdminFieldParams{AdminFieldID: c.id})
}

// DeleteAdminFieldCmd creates a new delete command for admin fields.
func (d Database) DeleteAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminFieldID) DeleteAdminFieldCmd {
	return DeleteAdminFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewAdminFieldCmdMysql is an audited command for create operations on admin_fields for MySQL.
type NewAdminFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c NewAdminFieldCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewAdminFieldCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminFieldCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c NewAdminFieldCmdMysql) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c NewAdminFieldCmdMysql) Params() any { return c.params }

// GetID extracts the ID from the created record.
func (c NewAdminFieldCmdMysql) GetID(u mdbm.AdminFields) string { return string(u.AdminFieldID) }

func (c NewAdminFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminFields, error) {
	queries := mdbm.New(tx)
	id := types.NewAdminFieldID()
	err := queries.CreateAdminField(ctx, mdbm.CreateAdminFieldParams{
		AdminFieldID: id,
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
	if err != nil {
		return mdbm.AdminFields{}, err
	}
	return queries.GetAdminField(ctx, mdbm.GetAdminFieldParams{AdminFieldID: id})
}

// NewAdminFieldCmd creates a new create command for admin fields.
func (d MysqlDatabase) NewAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminFieldParams) NewAdminFieldCmdMysql {
	return NewAdminFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateAdminFieldCmdMysql is an audited command for update operations on admin_fields for MySQL.
type UpdateAdminFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c UpdateAdminFieldCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateAdminFieldCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminFieldCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c UpdateAdminFieldCmdMysql) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c UpdateAdminFieldCmdMysql) Params() any { return c.params }

// GetID returns the ID of the record to update.
func (c UpdateAdminFieldCmdMysql) GetID() string { return string(c.params.AdminFieldID) }

// GetBefore retrieves the record before modification.
func (c UpdateAdminFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminField(ctx, mdbm.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

// Execute performs the update operation.
func (c UpdateAdminFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminField(ctx, mdbm.UpdateAdminFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		AdminFieldID: c.params.AdminFieldID,
	})
}

// UpdateAdminFieldCmd creates a new update command for admin fields.
func (d MysqlDatabase) UpdateAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldParams) UpdateAdminFieldCmdMysql {
	return UpdateAdminFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteAdminFieldCmdMysql is an audited command for delete operations on admin_fields for MySQL.
type DeleteAdminFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c DeleteAdminFieldCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteAdminFieldCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminFieldCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c DeleteAdminFieldCmdMysql) TableName() string { return "admin_fields" }

// GetID returns the ID of the record to delete.
func (c DeleteAdminFieldCmdMysql) GetID() string { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminField(ctx, mdbm.GetAdminFieldParams{AdminFieldID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminField(ctx, mdbm.DeleteAdminFieldParams{AdminFieldID: c.id})
}

// DeleteAdminFieldCmd creates a new delete command for admin fields.
func (d MysqlDatabase) DeleteAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminFieldID) DeleteAdminFieldCmdMysql {
	return DeleteAdminFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewAdminFieldCmdPsql is an audited command for create operations on admin_fields for PostgreSQL.
type NewAdminFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c NewAdminFieldCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewAdminFieldCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminFieldCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c NewAdminFieldCmdPsql) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c NewAdminFieldCmdPsql) Params() any { return c.params }

// GetID extracts the ID from the created record.
func (c NewAdminFieldCmdPsql) GetID(u mdbp.AdminFields) string { return string(u.AdminFieldID) }

func (c NewAdminFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminFields, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminField(ctx, mdbp.CreateAdminFieldParams{
		AdminFieldID: types.NewAdminFieldID(),
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewAdminFieldCmd creates a new create command for admin fields.
func (d PsqlDatabase) NewAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminFieldParams) NewAdminFieldCmdPsql {
	return NewAdminFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateAdminFieldCmdPsql is an audited command for update operations on admin_fields for PostgreSQL.
type UpdateAdminFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c UpdateAdminFieldCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateAdminFieldCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminFieldCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c UpdateAdminFieldCmdPsql) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c UpdateAdminFieldCmdPsql) Params() any { return c.params }

// GetID returns the ID of the record to update.
func (c UpdateAdminFieldCmdPsql) GetID() string { return string(c.params.AdminFieldID) }

// GetBefore retrieves the record before modification.
func (c UpdateAdminFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminField(ctx, mdbp.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

// Execute performs the update operation.
func (c UpdateAdminFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminField(ctx, mdbp.UpdateAdminFieldParams{
		ParentID:     c.params.ParentID,
		Label:        c.params.Label,
		Data:         c.params.Data,
		Validation:   c.params.Validation,
		UiConfig:     c.params.UIConfig,
		Type:         c.params.Type,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		AdminFieldID: c.params.AdminFieldID,
	})
}

// UpdateAdminFieldCmd creates a new update command for admin fields.
func (d PsqlDatabase) UpdateAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldParams) UpdateAdminFieldCmdPsql {
	return UpdateAdminFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteAdminFieldCmdPsql is an audited command for delete operations on admin_fields for PostgreSQL.
type DeleteAdminFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c DeleteAdminFieldCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteAdminFieldCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminFieldCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for the command.
func (c DeleteAdminFieldCmdPsql) TableName() string { return "admin_fields" }

// GetID returns the ID of the record to delete.
func (c DeleteAdminFieldCmdPsql) GetID() string { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminField(ctx, mdbp.GetAdminFieldParams{AdminFieldID: c.id})
}

// Execute performs the delete operation.
func (c DeleteAdminFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminField(ctx, mdbp.DeleteAdminFieldParams{AdminFieldID: c.id})
}

// DeleteAdminFieldCmd creates a new delete command for admin fields.
func (d PsqlDatabase) DeleteAdminFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminFieldID) DeleteAdminFieldCmdPsql {
	return DeleteAdminFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
