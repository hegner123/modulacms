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

// AdminContentFields represents a single content field entry in the admin namespace.
type AdminContentFields struct {
	AdminContentFieldID types.AdminContentFieldID  `json:"admin_content_field_id"`
	AdminRouteID        types.NullableAdminRouteID `json:"admin_route_id"`
	AdminContentDataID  types.NullableAdminContentID `json:"admin_content_data_id"`
	AdminFieldID        types.NullableAdminFieldID `json:"admin_field_id"`
	AdminFieldValue     string                     `json:"admin_field_value"`
	AuthorID            types.NullableUserID       `json:"author_id"`
	DateCreated         types.Timestamp            `json:"date_created"`
	DateModified        types.Timestamp            `json:"date_modified"`
}

// CreateAdminContentFieldParams contains the fields required to create a new AdminContentFields record.
type CreateAdminContentFieldParams struct {
	AdminRouteID       types.NullableAdminRouteID   `json:"admin_route_id"`
	AdminContentDataID types.NullableAdminContentID `json:"admin_content_data_id"`
	AdminFieldID       types.NullableAdminFieldID   `json:"admin_field_id"`
	AdminFieldValue    string                       `json:"admin_field_value"`
	AuthorID           types.NullableUserID         `json:"author_id"`
	DateCreated        types.Timestamp              `json:"date_created"`
	DateModified       types.Timestamp              `json:"date_modified"`
}

// UpdateAdminContentFieldParams contains the fields required to update an existing AdminContentFields record.
type UpdateAdminContentFieldParams struct {
	AdminRouteID        types.NullableAdminRouteID   `json:"admin_route_id"`
	AdminContentDataID  types.NullableAdminContentID `json:"admin_content_data_id"`
	AdminFieldID        types.NullableAdminFieldID   `json:"admin_field_id"`
	AdminFieldValue     string                       `json:"admin_field_value"`
	AuthorID            types.NullableUserID         `json:"author_id"`
	DateCreated         types.Timestamp              `json:"date_created"`
	DateModified        types.Timestamp              `json:"date_modified"`
	AdminContentFieldID types.AdminContentFieldID    `json:"admin_content_field_id"`
}

// ListAdminContentFieldsByRoutePaginatedParams contains parameters for paginated route-based AdminContentFields retrieval.
type ListAdminContentFieldsByRoutePaginatedParams struct {
	AdminRouteID types.NullableAdminRouteID
	Limit        int64
	Offset       int64
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapAdminContentFieldJSON converts AdminContentFields to ContentFieldsJSON for tree building.
// Maps admin field value into the public ContentFieldsJSON shape so BuildNodes works unchanged.
func MapAdminContentFieldJSON(a AdminContentFields) ContentFieldsJSON {
	return ContentFieldsJSON{
		ContentFieldID: 0,
		RouteID:        0,
		ContentDataID:  0,
		FieldID:        0,
		FieldValue:     a.AdminFieldValue,
		AuthorID:       0,
		DateCreated:    a.DateCreated.String(),
		DateModified:   a.DateModified.String(),
	}
}

// MapStringAdminContentField converts AdminContentFields to StringAdminContentFields for table display
func MapStringAdminContentField(a AdminContentFields) StringAdminContentFields {
	return StringAdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID.String(),
		AdminRouteID:        a.AdminRouteID.String(),
		AdminContentDataID:  a.AdminContentDataID.String(),
		AdminFieldID:        a.AdminFieldID.String(),
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID.String(),
		DateCreated:         a.DateCreated.String(),
		DateModified:        a.DateModified.String(),
		History:             "", // History field removed
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapAdminContentField converts a sqlc-generated mdb.AdminContentFields to the wrapper AdminContentFields type.
func (d Database) MapAdminContentField(a mdb.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}

// MapCreateAdminContentFieldParams converts wrapper CreateAdminContentFieldParams to sqlc params for SQLite.
func (d Database) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdb.CreateAdminContentFieldParams {
	return mdb.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}

// MapUpdateAdminContentFieldParams converts wrapper UpdateAdminContentFieldParams to sqlc params for SQLite.
func (d Database) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdb.UpdateAdminContentFieldParams {
	return mdb.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// QUERIES

// CountAdminContentFields returns the total number of AdminContentFields records.
func (d Database) CountAdminContentFields() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
// CreateAdminContentField inserts a new AdminContentFields record.
func (d Database) CreateAdminContentField(ctx context.Context, ac audited.AuditContext, s CreateAdminContentFieldParams) (*AdminContentFields, error) {
	cmd := d.NewAdminContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentField: %w", err)
	}
	r := d.MapAdminContentField(result)
	return &r, nil
}
// CreateAdminContentFieldTable creates the database table for AdminContentFields entities.
func (d Database) CreateAdminContentFieldTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
// DeleteAdminContentField removes an AdminContentFields record.
func (d Database) DeleteAdminContentField(ctx context.Context, ac audited.AuditContext, id types.AdminContentFieldID) error {
	cmd := d.DeleteAdminContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
// GetAdminContentField retrieves an AdminContentFields by ID.
func (d Database) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdb.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
// ListAdminContentFields returns all AdminContentFields records.
func (d Database) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
// ListAdminContentFieldsByRoute returns AdminContentFields records filtered by route ID.
func (d Database) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdb.ListAdminContentFieldsByRouteParams{AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID(id), Valid: id != ""}})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
// ListAdminContentFieldsPaginated returns AdminContentFields records with pagination.
func (d Database) ListAdminContentFieldsPaginated(params PaginationParams) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsPaginated(d.Context, mdb.ListAdminContentFieldsPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields paginated: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
// ListAdminContentFieldsByRoutePaginated returns AdminContentFields records filtered by route with pagination.
func (d Database) ListAdminContentFieldsByRoutePaginated(params ListAdminContentFieldsByRoutePaginatedParams) (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoutePaginated(d.Context, mdb.ListAdminContentFieldsByRoutePaginatedParams{
		AdminRouteID: params.AdminRouteID,
		Limit:        params.Limit,
		Offset:       params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by route paginated: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
// UpdateAdminContentField modifies an existing AdminContentFields record.
func (d Database) UpdateAdminContentField(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentFieldParams) (*string, error) {
	cmd := d.UpdateAdminContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentFieldID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapAdminContentField converts a sqlc-generated mdbm.AdminContentFields to the wrapper AdminContentFields type.
func (d MysqlDatabase) MapAdminContentField(a mdbm.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
// MapCreateAdminContentFieldParams converts wrapper CreateAdminContentFieldParams to sqlc params for MySQL.
func (d MysqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbm.CreateAdminContentFieldParams {
	return mdbm.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
// MapUpdateAdminContentFieldParams converts wrapper UpdateAdminContentFieldParams to sqlc params for MySQL.
func (d MysqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbm.UpdateAdminContentFieldParams {
	return mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// QUERIES

func (d MysqlDatabase) CountAdminContentFields() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d MysqlDatabase) CreateAdminContentField(ctx context.Context, ac audited.AuditContext, s CreateAdminContentFieldParams) (*AdminContentFields, error) {
	cmd := d.NewAdminContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentField: %w", err)
	}
	r := d.MapAdminContentField(result)
	return &r, nil
}

func (d MysqlDatabase) CreateAdminContentFieldTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d MysqlDatabase) DeleteAdminContentField(ctx context.Context, ac audited.AuditContext, id types.AdminContentFieldID) error {
	cmd := d.DeleteAdminContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d MysqlDatabase) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdbm.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdbm.ListAdminContentFieldsByRouteParams{AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID(id), Valid: id != ""}})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFieldsPaginated(params PaginationParams) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsPaginated(d.Context, mdbm.ListAdminContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields paginated: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) ListAdminContentFieldsByRoutePaginated(params ListAdminContentFieldsByRoutePaginatedParams) (*[]AdminContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoutePaginated(d.Context, mdbm.ListAdminContentFieldsByRoutePaginatedParams{
		AdminRouteID: params.AdminRouteID,
		Limit:        int32(params.Limit),
		Offset:       int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by route paginated: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d MysqlDatabase) UpdateAdminContentField(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentFieldParams) (*string, error) {
	cmd := d.UpdateAdminContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentFieldID)
	return &msg, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS

// MapAdminContentField converts a sqlc-generated mdbp.AdminContentFields to the wrapper AdminContentFields type.
func (d PsqlDatabase) MapAdminContentField(a mdbp.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: a.AdminContentFieldID,
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
// MapCreateAdminContentFieldParams converts wrapper CreateAdminContentFieldParams to sqlc params for PostgreSQL.
func (d PsqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbp.CreateAdminContentFieldParams {
	return mdbp.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
	}
}
// MapUpdateAdminContentFieldParams converts wrapper UpdateAdminContentFieldParams to sqlc params for PostgreSQL.
func (d PsqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbp.UpdateAdminContentFieldParams {
	return mdbp.UpdateAdminContentFieldParams{
		AdminRouteID:        a.AdminRouteID,
		AdminContentDataID:  a.AdminContentDataID,
		AdminFieldID:        a.AdminFieldID,
		AdminFieldValue:     a.AdminFieldValue,
		AuthorID:            a.AuthorID,
		DateCreated:         a.DateCreated,
		DateModified:        a.DateModified,
		AdminContentFieldID: a.AdminContentFieldID,
	}
}

// /QUERIES

func (d PsqlDatabase) CountAdminContentFields() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminContentField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}
func (d PsqlDatabase) CreateAdminContentField(ctx context.Context, ac audited.AuditContext, s CreateAdminContentFieldParams) (*AdminContentFields, error) {
	cmd := d.NewAdminContentFieldCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminContentField: %w", err)
	}
	r := d.MapAdminContentField(result)
	return &r, nil
}
func (d PsqlDatabase) CreateAdminContentFieldTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminContentFieldTable(d.Context)
	return err
}
func (d PsqlDatabase) DeleteAdminContentField(ctx context.Context, ac audited.AuditContext, id types.AdminContentFieldID) error {
	cmd := d.DeleteAdminContentFieldCmd(ctx, ac, id)
	return audited.Delete(cmd)
}
func (d PsqlDatabase) GetAdminContentField(id types.AdminContentFieldID) (*AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, mdbp.GetAdminContentFieldParams{AdminContentFieldID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFieldsByRoute(id string) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoute(d.Context, mdbp.ListAdminContentFieldsByRouteParams{AdminRouteID: types.NullableAdminRouteID{ID: types.AdminRouteID(id), Valid: id != ""}})
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFieldsPaginated(params PaginationParams) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsPaginated(d.Context, mdbp.ListAdminContentFieldsPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields paginated: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) ListAdminContentFieldsByRoutePaginated(params ListAdminContentFieldsByRoutePaginatedParams) (*[]AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminContentFieldsByRoutePaginated(d.Context, mdbp.ListAdminContentFieldsByRoutePaginatedParams{
		AdminRouteID: params.AdminRouteID,
		Limit:        int32(params.Limit),
		Offset:       int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminContentFields by route paginated: %v", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d PsqlDatabase) UpdateAdminContentField(ctx context.Context, ac audited.AuditContext, s UpdateAdminContentFieldParams) (*string, error) {
	cmd := d.UpdateAdminContentFieldCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminContentField: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.AdminContentFieldID)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMANDS — SQLITE
//////////////////////////////

// NewAdminContentFieldCmd is an audited create command for admin_content_fields (SQLite).
type NewAdminContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewAdminContentFieldCmd) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c NewAdminContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c NewAdminContentFieldCmd) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c NewAdminContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c NewAdminContentFieldCmd) TableName() string                     { return "admin_content_fields" }
// Params returns the command parameters.
func (c NewAdminContentFieldCmd) Params() any                           { return c.params }
// GetID returns the ID from the created entity.
func (c NewAdminContentFieldCmd) GetID(row mdb.AdminContentFields) string {
	return string(row.AdminContentFieldID)
}

// Execute runs the audited command.
func (c NewAdminContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminContentFields, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminContentField(ctx, mdb.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
	})
}

// NewAdminContentFieldCmd creates a new audited create command for SQLite.
func (d Database) NewAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentFieldParams) NewAdminContentFieldCmd {
	return NewAdminContentFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateAdminContentFieldCmd is an audited update command for admin_content_fields (SQLite).
type UpdateAdminContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateAdminContentFieldCmd) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c UpdateAdminContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c UpdateAdminContentFieldCmd) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c UpdateAdminContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c UpdateAdminContentFieldCmd) TableName() string                     { return "admin_content_fields" }
// Params returns the command parameters.
func (c UpdateAdminContentFieldCmd) Params() any                           { return c.params }
// GetID returns the ID of the entity being updated.
func (c UpdateAdminContentFieldCmd) GetID() string {
	return string(c.params.AdminContentFieldID)
}

// GetBefore retrieves the entity state before the update.
func (c UpdateAdminContentFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentField(ctx, mdb.GetAdminContentFieldParams{AdminContentFieldID: c.params.AdminContentFieldID})
}

// Execute runs the audited command.
func (c UpdateAdminContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminContentField(ctx, mdb.UpdateAdminContentFieldParams{
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
		AdminContentFieldID: c.params.AdminContentFieldID,
	})
}

// UpdateAdminContentFieldCmd creates a new audited update command for SQLite.
func (d Database) UpdateAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentFieldParams) UpdateAdminContentFieldCmd {
	return UpdateAdminContentFieldCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// DeleteAdminContentFieldCmd is an audited delete command for admin_content_fields (SQLite).
type DeleteAdminContentFieldCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteAdminContentFieldCmd) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c DeleteAdminContentFieldCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c DeleteAdminContentFieldCmd) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c DeleteAdminContentFieldCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c DeleteAdminContentFieldCmd) TableName() string                     { return "admin_content_fields" }
// GetID returns the ID of the entity being deleted.
func (c DeleteAdminContentFieldCmd) GetID() string                         { return string(c.id) }

// GetBefore retrieves the entity state before deletion.
func (c DeleteAdminContentFieldCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminContentFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminContentField(ctx, mdb.GetAdminContentFieldParams{AdminContentFieldID: c.id})
}

// Execute runs the audited command.
func (c DeleteAdminContentFieldCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminContentField(ctx, mdb.DeleteAdminContentFieldParams{AdminContentFieldID: c.id})
}

// DeleteAdminContentFieldCmd creates a new audited delete command for SQLite.
func (d Database) DeleteAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentFieldID) DeleteAdminContentFieldCmd {
	return DeleteAdminContentFieldCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — MYSQL
//////////////////////////////

// NewAdminContentFieldCmdMysql is an audited create command for admin_content_fields (MySQL).
type NewAdminContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewAdminContentFieldCmdMysql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c NewAdminContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c NewAdminContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c NewAdminContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c NewAdminContentFieldCmdMysql) TableName() string                     { return "admin_content_fields" }
// Params returns the command parameters.
func (c NewAdminContentFieldCmdMysql) Params() any                           { return c.params }
// GetID returns the ID from the created entity.
func (c NewAdminContentFieldCmdMysql) GetID(row mdbm.AdminContentFields) string {
	return string(row.AdminContentFieldID)
}

// Execute runs the audited command.
func (c NewAdminContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentFields, error) {
	id := types.NewAdminContentFieldID()
	queries := mdbm.New(tx)
	err := queries.CreateAdminContentField(ctx, mdbm.CreateAdminContentFieldParams{
		AdminContentFieldID: id,
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
	})
	if err != nil {
		return mdbm.AdminContentFields{}, fmt.Errorf("execute create admin_content_fields: %w", err)
	}
	return queries.GetAdminContentField(ctx, mdbm.GetAdminContentFieldParams{AdminContentFieldID: id})
}

// NewAdminContentFieldCmd creates a new audited create command for MySQL.
func (d MysqlDatabase) NewAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentFieldParams) NewAdminContentFieldCmdMysql {
	return NewAdminContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateAdminContentFieldCmdMysql is an audited update command for admin_content_fields (MySQL).
type UpdateAdminContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateAdminContentFieldCmdMysql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c UpdateAdminContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c UpdateAdminContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c UpdateAdminContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c UpdateAdminContentFieldCmdMysql) TableName() string                     { return "admin_content_fields" }
// Params returns the command parameters.
func (c UpdateAdminContentFieldCmdMysql) Params() any                           { return c.params }
// GetID returns the ID of the entity being updated.
func (c UpdateAdminContentFieldCmdMysql) GetID() string {
	return string(c.params.AdminContentFieldID)
}

// GetBefore retrieves the entity state before the update.
func (c UpdateAdminContentFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentField(ctx, mdbm.GetAdminContentFieldParams{AdminContentFieldID: c.params.AdminContentFieldID})
}

// Execute runs the audited command.
func (c UpdateAdminContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminContentField(ctx, mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
		AdminContentFieldID: c.params.AdminContentFieldID,
	})
}

// UpdateAdminContentFieldCmd creates a new audited update command for MySQL.
func (d MysqlDatabase) UpdateAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentFieldParams) UpdateAdminContentFieldCmdMysql {
	return UpdateAdminContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// DeleteAdminContentFieldCmdMysql is an audited delete command for admin_content_fields (MySQL).
type DeleteAdminContentFieldCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteAdminContentFieldCmdMysql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c DeleteAdminContentFieldCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c DeleteAdminContentFieldCmdMysql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c DeleteAdminContentFieldCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c DeleteAdminContentFieldCmdMysql) TableName() string                     { return "admin_content_fields" }
// GetID returns the ID of the entity being deleted.
func (c DeleteAdminContentFieldCmdMysql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the entity state before deletion.
func (c DeleteAdminContentFieldCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminContentFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminContentField(ctx, mdbm.GetAdminContentFieldParams{AdminContentFieldID: c.id})
}

// Execute runs the audited command.
func (c DeleteAdminContentFieldCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminContentField(ctx, mdbm.DeleteAdminContentFieldParams{AdminContentFieldID: c.id})
}

// DeleteAdminContentFieldCmd creates a new audited delete command for MySQL.
func (d MysqlDatabase) DeleteAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentFieldID) DeleteAdminContentFieldCmdMysql {
	return DeleteAdminContentFieldCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// AUDITED COMMANDS — POSTGRES
//////////////////////////////

// NewAdminContentFieldCmdPsql is an audited create command for admin_content_fields (PostgreSQL).
type NewAdminContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewAdminContentFieldCmdPsql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c NewAdminContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c NewAdminContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c NewAdminContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c NewAdminContentFieldCmdPsql) TableName() string                     { return "admin_content_fields" }
// Params returns the command parameters.
func (c NewAdminContentFieldCmdPsql) Params() any                           { return c.params }
// GetID returns the ID from the created entity.
func (c NewAdminContentFieldCmdPsql) GetID(row mdbp.AdminContentFields) string {
	return string(row.AdminContentFieldID)
}

// Execute runs the audited command.
func (c NewAdminContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentFields, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminContentField(ctx, mdbp.CreateAdminContentFieldParams{
		AdminContentFieldID: types.NewAdminContentFieldID(),
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
	})
}

// NewAdminContentFieldCmd creates a new audited create command for PostgreSQL.
func (d PsqlDatabase) NewAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminContentFieldParams) NewAdminContentFieldCmdPsql {
	return NewAdminContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateAdminContentFieldCmdPsql is an audited update command for admin_content_fields (PostgreSQL).
type UpdateAdminContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminContentFieldParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateAdminContentFieldCmdPsql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c UpdateAdminContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c UpdateAdminContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c UpdateAdminContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c UpdateAdminContentFieldCmdPsql) TableName() string                     { return "admin_content_fields" }
// Params returns the command parameters.
func (c UpdateAdminContentFieldCmdPsql) Params() any                           { return c.params }
// GetID returns the ID of the entity being updated.
func (c UpdateAdminContentFieldCmdPsql) GetID() string {
	return string(c.params.AdminContentFieldID)
}

// GetBefore retrieves the entity state before the update.
func (c UpdateAdminContentFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentField(ctx, mdbp.GetAdminContentFieldParams{AdminContentFieldID: c.params.AdminContentFieldID})
}

// Execute runs the audited command.
func (c UpdateAdminContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminContentField(ctx, mdbp.UpdateAdminContentFieldParams{
		AdminRouteID:        c.params.AdminRouteID,
		AdminContentDataID:  c.params.AdminContentDataID,
		AdminFieldID:        c.params.AdminFieldID,
		AdminFieldValue:     c.params.AdminFieldValue,
		AuthorID:            c.params.AuthorID,
		DateCreated:         c.params.DateCreated,
		DateModified:        c.params.DateModified,
		AdminContentFieldID: c.params.AdminContentFieldID,
	})
}

// UpdateAdminContentFieldCmd creates a new audited update command for PostgreSQL.
func (d PsqlDatabase) UpdateAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminContentFieldParams) UpdateAdminContentFieldCmdPsql {
	return UpdateAdminContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// DeleteAdminContentFieldCmdPsql is an audited delete command for admin_content_fields (PostgreSQL).
type DeleteAdminContentFieldCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminContentFieldID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteAdminContentFieldCmdPsql) Context() context.Context              { return c.ctx }
// AuditContext returns the audit context.
func (c DeleteAdminContentFieldCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
// Connection returns the database connection.
func (c DeleteAdminContentFieldCmdPsql) Connection() *sql.DB                   { return c.conn }
// Recorder returns the change event recorder.
func (c DeleteAdminContentFieldCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
// TableName returns the table name.
func (c DeleteAdminContentFieldCmdPsql) TableName() string                     { return "admin_content_fields" }
// GetID returns the ID of the entity being deleted.
func (c DeleteAdminContentFieldCmdPsql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the entity state before deletion.
func (c DeleteAdminContentFieldCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminContentFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminContentField(ctx, mdbp.GetAdminContentFieldParams{AdminContentFieldID: c.id})
}

// Execute runs the audited command.
func (c DeleteAdminContentFieldCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminContentField(ctx, mdbp.DeleteAdminContentFieldParams{AdminContentFieldID: c.id})
}

// DeleteAdminContentFieldCmd creates a new audited delete command for PostgreSQL.
func (d PsqlDatabase) DeleteAdminContentFieldCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminContentFieldID) DeleteAdminContentFieldCmdPsql {
	return DeleteAdminContentFieldCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
