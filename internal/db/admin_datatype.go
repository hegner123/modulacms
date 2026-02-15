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

// AdminDatatypes represents an admin datatype in the system.
type AdminDatatypes struct {
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
	ParentID        types.NullableAdminDatatypeID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
	AuthorID        types.NullableUserID    `json:"author_id"`
	DateCreated     types.Timestamp         `json:"date_created"`
	DateModified    types.Timestamp         `json:"date_modified"`
}
// CreateAdminDatatypeParams contains the parameters for creating an admin datatype.
type CreateAdminDatatypeParams struct {
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                  `json:"label"`
	Type         string                  `json:"type"`
	AuthorID     types.NullableUserID    `json:"author_id"`
	DateCreated  types.Timestamp         `json:"date_created"`
	DateModified types.Timestamp         `json:"date_modified"`
}
// ListAdminDatatypeByRouteIdRow represents a row from listing admin datatypes by route ID.
type ListAdminDatatypeByRouteIdRow struct {
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
	AdminRouteID    types.NullableRouteID   `json:"admin_route_id"`
	ParentID        types.NullableAdminDatatypeID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
}
// UpdateAdminDatatypeParams contains the parameters for updating an admin datatype.
type UpdateAdminDatatypeParams struct {
	ParentID        types.NullableAdminDatatypeID `json:"parent_id"`
	Label           string                  `json:"label"`
	Type            string                  `json:"type"`
	AuthorID        types.NullableUserID    `json:"author_id"`
	DateCreated     types.Timestamp         `json:"date_created"`
	DateModified    types.Timestamp         `json:"date_modified"`
	AdminDatatypeID types.AdminDatatypeID   `json:"admin_datatype_id"`
}
// UtilityGetAdminDatatypesRow represents a row from the utility query for admin datatypes.
type UtilityGetAdminDatatypesRow struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	Label           string                `json:"label"`
}

// ListAdminDatatypeChildrenPaginatedParams contains parameters for paginated list of admin datatype children.
type ListAdminDatatypeChildrenPaginatedParams struct {
	ParentID types.AdminDatatypeID
	Limit    int64
	Offset   int64
}

// FormParams and JSON variants removed - use typed params directly

// MapAdminDatatypeJSON converts AdminDatatypes to DatatypeJSON for tree building by mapping admin datatype ID into the public DatatypeJSON shape.
func MapAdminDatatypeJSON(a AdminDatatypes) DatatypeJSON {
	return DatatypeJSON{
		DatatypeID:   a.AdminDatatypeID.String(),
		ParentID:     a.ParentID.String(),
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// MapStringAdminDatatype converts AdminDatatypes to StringAdminDatatypes for table display.
func MapStringAdminDatatype(a AdminDatatypes) StringAdminDatatypes {
	return StringAdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID.String(),
		ParentID:        a.ParentID.String(),
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID.String(),
		DateCreated:     a.DateCreated.String(),
		DateModified:    a.DateModified.String(),
		History:         "", // History field removed from AdminDatatypes
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapAdminDatatype converts a sqlc-generated type to the wrapper type.
func (d Database) MapAdminDatatype(a mdb.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

// MapCreateAdminDatatypeParams converts wrapper params to sqlc params for insertion.
func (d Database) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdb.CreateAdminDatatypeParams {
	return mdb.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminDatatypeParams converts wrapper params to sqlc params for updates.
func (d Database) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdb.UpdateAdminDatatypeParams {
	return mdb.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

// CountAdminDatatypes returns the total count of admin datatypes.
func (d Database) CountAdminDatatypes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminDatatype inserts a new admin datatype record.
func (d Database) CreateAdminDatatype(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeParams) (*AdminDatatypes, error) {
	cmd := d.NewAdminDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatype: %w", err)
	}
	r := d.MapAdminDatatype(result)
	return &r, nil
}
// CreateAdminDatatypeTable creates the admin_datatypes table.
func (d Database) CreateAdminDatatypeTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

// DeleteAdminDatatype removes an admin datatype record.
func (d Database) DeleteAdminDatatype(ctx context.Context, ac audited.AuditContext, id types.AdminDatatypeID) error {
	cmd := d.DeleteAdminDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminDatatypeById retrieves an admin datatype by ID.
func (d Database) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdb.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

// ListAdminDatatypeGlobalId returns all admin datatypes with global scope.
func (d Database) ListAdminDatatypeGlobalId() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeGlobal(d.Context)
	if err != nil {
		return nil, err
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypes returns all admin datatype records.
func (d Database) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypesPaginated returns admin datatype records with pagination.
func (d Database) ListAdminDatatypesPaginated(params PaginationParams) (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypePaginated(d.Context, mdb.ListAdminDatatypePaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes paginated: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypeChildrenPaginated returns child admin datatypes with pagination.
func (d Database) ListAdminDatatypeChildrenPaginated(params ListAdminDatatypeChildrenPaginatedParams) (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeChildrenPaginated(d.Context, mdb.ListAdminDatatypeChildrenPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get child admin datatypes paginated: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminDatatype modifies an existing admin datatype record.
func (d Database) UpdateAdminDatatype(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapAdminDatatype converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapAdminDatatype(a mdbm.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

// MapCreateAdminDatatypeParams converts wrapper params to sqlc params for insertion.
func (d MysqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbm.CreateAdminDatatypeParams {
	return mdbm.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminDatatypeParams converts wrapper params to sqlc params for updates.
func (d MysqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbm.UpdateAdminDatatypeParams {
	return mdbm.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

// CountAdminDatatypes returns the total count of admin datatypes.
func (d MysqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminDatatype inserts a new admin datatype record.
func (d MysqlDatabase) CreateAdminDatatype(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeParams) (*AdminDatatypes, error) {
	cmd := d.NewAdminDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatype: %w", err)
	}
	r := d.MapAdminDatatype(result)
	return &r, nil
}

// CreateAdminDatatypeTable creates the admin_datatypes table.
func (d MysqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

// DeleteAdminDatatype removes an admin datatype record.
func (d MysqlDatabase) DeleteAdminDatatype(ctx context.Context, ac audited.AuditContext, id types.AdminDatatypeID) error {
	cmd := d.DeleteAdminDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminDatatypeById retrieves an admin datatype by ID.
func (d MysqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdbm.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

// ListAdminDatatypes returns all admin datatype records.
func (d MysqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypesPaginated returns admin datatype records with pagination.
func (d MysqlDatabase) ListAdminDatatypesPaginated(params PaginationParams) (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypePaginated(d.Context, mdbm.ListAdminDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes paginated: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypeChildrenPaginated returns child admin datatypes with pagination.
func (d MysqlDatabase) ListAdminDatatypeChildrenPaginated(params ListAdminDatatypeChildrenPaginatedParams) (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeChildrenPaginated(d.Context, mdbm.ListAdminDatatypeChildrenPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get child admin datatypes paginated: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminDatatype modifies an existing admin datatype record.
func (d MysqlDatabase) UpdateAdminDatatype(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapAdminDatatype converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapAdminDatatype(a mdbp.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: a.AdminDatatypeID,
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
	}
}

// MapCreateAdminDatatypeParams converts wrapper params to sqlc params for insertion.
func (d PsqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbp.CreateAdminDatatypeParams {
	return mdbp.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminDatatypeParams converts wrapper params to sqlc params for updates.
func (d PsqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbp.UpdateAdminDatatypeParams {
	return mdbp.UpdateAdminDatatypeParams{
		ParentID:        a.ParentID,
		Label:           a.Label,
		Type:            a.Type,
		AuthorID:        a.AuthorID,
		DateCreated:     a.DateCreated,
		DateModified:    a.DateModified,
		AdminDatatypeID: a.AdminDatatypeID,
	}
}

// QUERIES

// CountAdminDatatypes returns the total count of admin datatypes.
func (d PsqlDatabase) CountAdminDatatypes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminDatatype inserts a new admin datatype record.
func (d PsqlDatabase) CreateAdminDatatype(ctx context.Context, ac audited.AuditContext, s CreateAdminDatatypeParams) (*AdminDatatypes, error) {
	cmd := d.NewAdminDatatypeCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminDatatype: %w", err)
	}
	r := d.MapAdminDatatype(result)
	return &r, nil
}

// CreateAdminDatatypeTable creates the admin_datatypes table.
func (d PsqlDatabase) CreateAdminDatatypeTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminDatatypeTable(d.Context)
	return err
}

// DeleteAdminDatatype removes an admin datatype record.
func (d PsqlDatabase) DeleteAdminDatatype(ctx context.Context, ac audited.AuditContext, id types.AdminDatatypeID) error {
	cmd := d.DeleteAdminDatatypeCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminDatatypeById retrieves an admin datatype by ID.
func (d PsqlDatabase) GetAdminDatatypeById(id types.AdminDatatypeID) (*AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, mdbp.GetAdminDatatypeParams{AdminDatatypeID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

// ListAdminDatatypes returns all admin datatype records.
func (d PsqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin datatypes: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypesPaginated returns admin datatype records with pagination.
func (d PsqlDatabase) ListAdminDatatypesPaginated(params PaginationParams) (*[]AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypePaginated(d.Context, mdbp.ListAdminDatatypePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes paginated: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminDatatypeChildrenPaginated returns child admin datatypes with pagination.
func (d PsqlDatabase) ListAdminDatatypeChildrenPaginated(params ListAdminDatatypeChildrenPaginatedParams) (*[]AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminDatatypeChildrenPaginated(d.Context, mdbp.ListAdminDatatypeChildrenPaginatedParams{
		ParentID: types.NullableAdminDatatypeID{ID: params.ParentID, Valid: true},
		Limit:    int32(params.Limit),
		Offset:   int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get child admin datatypes paginated: %v", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminDatatype modifies an existing admin datatype record.
func (d PsqlDatabase) UpdateAdminDatatype(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeParams) (*string, error) {
	cmd := d.UpdateAdminDatatypeCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminDatatype: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

// NewAdminDatatypeCmd is an audited command for create on admin_datatypes.
type NewAdminDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c NewAdminDatatypeCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewAdminDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminDatatypeCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewAdminDatatypeCmd) TableName() string                     { return "admin_datatypes" }

// Params returns the command parameters.
func (c NewAdminDatatypeCmd) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewAdminDatatypeCmd) GetID(u mdb.AdminDatatypes) string     { return string(u.AdminDatatypeID) }

// Execute inserts a new admin datatype and returns the created record.
func (c NewAdminDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypes, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminDatatype(ctx, mdb.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
	})
}

// NewAdminDatatypeCmd constructs a create command for admin_datatypes.
func (d Database) NewAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeParams) NewAdminDatatypeCmd {
	return NewAdminDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateAdminDatatypeCmd is an audited command for update on admin_datatypes.
type UpdateAdminDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c UpdateAdminDatatypeCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateAdminDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminDatatypeCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateAdminDatatypeCmd) TableName() string                     { return "admin_datatypes" }

// Params returns the command parameters.
func (c UpdateAdminDatatypeCmd) Params() any                           { return c.params }

// GetID returns the record ID being updated.
func (c UpdateAdminDatatypeCmd) GetID() string                         { return string(c.params.AdminDatatypeID) }

// GetBefore retrieves the record before update.
func (c UpdateAdminDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypes, error) {
	queries := mdb.New(tx)
	return queries.GetAdminDatatype(ctx, mdb.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

// Execute updates the admin datatype record.
func (c UpdateAdminDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminDatatype(ctx, mdb.UpdateAdminDatatypeParams{
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

// UpdateAdminDatatypeCmd constructs an update command for admin_datatypes.
func (d Database) UpdateAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeParams) UpdateAdminDatatypeCmd {
	return UpdateAdminDatatypeCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteAdminDatatypeCmd is an audited command for delete on admin_datatypes.
type DeleteAdminDatatypeCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminDatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c DeleteAdminDatatypeCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteAdminDatatypeCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminDatatypeCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminDatatypeCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteAdminDatatypeCmd) TableName() string                     { return "admin_datatypes" }

// GetID returns the record ID being deleted.
func (c DeleteAdminDatatypeCmd) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminDatatypeCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypes, error) {
	queries := mdb.New(tx)
	return queries.GetAdminDatatype(ctx, mdb.GetAdminDatatypeParams{AdminDatatypeID: c.id})
}

// Execute deletes the admin datatype record.
func (c DeleteAdminDatatypeCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminDatatype(ctx, mdb.DeleteAdminDatatypeParams{AdminDatatypeID: c.id})
}

// DeleteAdminDatatypeCmd constructs a delete command for admin_datatypes.
func (d Database) DeleteAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminDatatypeID) DeleteAdminDatatypeCmd {
	return DeleteAdminDatatypeCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewAdminDatatypeCmdMysql is an audited create command for MySQL.
type NewAdminDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c NewAdminDatatypeCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewAdminDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewAdminDatatypeCmdMysql) TableName() string                     { return "admin_datatypes" }

// Params returns the command parameters.
func (c NewAdminDatatypeCmdMysql) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewAdminDatatypeCmdMysql) GetID(u mdbm.AdminDatatypes) string    { return string(u.AdminDatatypeID) }

// Execute inserts a new admin datatype and returns the created record.
func (c NewAdminDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypes, error) {
	queries := mdbm.New(tx)
	id := types.NewAdminDatatypeID()
	err := queries.CreateAdminDatatype(ctx, mdbm.CreateAdminDatatypeParams{
		AdminDatatypeID: id,
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
	})
	if err != nil {
		return mdbm.AdminDatatypes{}, err
	}
	return queries.GetAdminDatatype(ctx, mdbm.GetAdminDatatypeParams{AdminDatatypeID: id})
}

// NewAdminDatatypeCmd constructs a create command for admin_datatypes (MySQL).
func (d MysqlDatabase) NewAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeParams) NewAdminDatatypeCmdMysql {
	return NewAdminDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateAdminDatatypeCmdMysql is an audited update command for MySQL.
type UpdateAdminDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c UpdateAdminDatatypeCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateAdminDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateAdminDatatypeCmdMysql) TableName() string                     { return "admin_datatypes" }

// Params returns the command parameters.
func (c UpdateAdminDatatypeCmdMysql) Params() any                           { return c.params }

// GetID returns the record ID being updated.
func (c UpdateAdminDatatypeCmdMysql) GetID() string                         { return string(c.params.AdminDatatypeID) }

// GetBefore retrieves the record before update.
func (c UpdateAdminDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminDatatype(ctx, mdbm.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

// Execute updates the admin datatype record.
func (c UpdateAdminDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminDatatype(ctx, mdbm.UpdateAdminDatatypeParams{
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

// UpdateAdminDatatypeCmd constructs an update command for admin_datatypes (MySQL).
func (d MysqlDatabase) UpdateAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeParams) UpdateAdminDatatypeCmdMysql {
	return UpdateAdminDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteAdminDatatypeCmdMysql is an audited delete command for MySQL.
type DeleteAdminDatatypeCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminDatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c DeleteAdminDatatypeCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteAdminDatatypeCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminDatatypeCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminDatatypeCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteAdminDatatypeCmdMysql) TableName() string                     { return "admin_datatypes" }

// GetID returns the record ID being deleted.
func (c DeleteAdminDatatypeCmdMysql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminDatatypeCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminDatatype(ctx, mdbm.GetAdminDatatypeParams{AdminDatatypeID: c.id})
}

// Execute deletes the admin datatype record.
func (c DeleteAdminDatatypeCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminDatatype(ctx, mdbm.DeleteAdminDatatypeParams{AdminDatatypeID: c.id})
}

// DeleteAdminDatatypeCmd constructs a delete command for admin_datatypes (MySQL).
func (d MysqlDatabase) DeleteAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminDatatypeID) DeleteAdminDatatypeCmdMysql {
	return DeleteAdminDatatypeCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewAdminDatatypeCmdPsql is an audited create command for PostgreSQL.
type NewAdminDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c NewAdminDatatypeCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c NewAdminDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewAdminDatatypeCmdPsql) TableName() string                     { return "admin_datatypes" }

// Params returns the command parameters.
func (c NewAdminDatatypeCmdPsql) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewAdminDatatypeCmdPsql) GetID(u mdbp.AdminDatatypes) string    { return string(u.AdminDatatypeID) }

// Execute inserts a new admin datatype and returns the created record.
func (c NewAdminDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypes, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminDatatype(ctx, mdbp.CreateAdminDatatypeParams{
		AdminDatatypeID: types.NewAdminDatatypeID(),
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
	})
}

// NewAdminDatatypeCmd constructs a create command for admin_datatypes (PostgreSQL).
func (d PsqlDatabase) NewAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminDatatypeParams) NewAdminDatatypeCmdPsql {
	return NewAdminDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateAdminDatatypeCmdPsql is an audited update command for PostgreSQL.
type UpdateAdminDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c UpdateAdminDatatypeCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c UpdateAdminDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateAdminDatatypeCmdPsql) TableName() string                     { return "admin_datatypes" }

// Params returns the command parameters.
func (c UpdateAdminDatatypeCmdPsql) Params() any                           { return c.params }

// GetID returns the record ID being updated.
func (c UpdateAdminDatatypeCmdPsql) GetID() string                         { return string(c.params.AdminDatatypeID) }

// GetBefore retrieves the record before update.
func (c UpdateAdminDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminDatatype(ctx, mdbp.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

// Execute updates the admin datatype record.
func (c UpdateAdminDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminDatatype(ctx, mdbp.UpdateAdminDatatypeParams{
		ParentID:        c.params.ParentID,
		Label:           c.params.Label,
		Type:            c.params.Type,
		AuthorID:        c.params.AuthorID,
		DateCreated:     c.params.DateCreated,
		DateModified:    c.params.DateModified,
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

// UpdateAdminDatatypeCmd constructs an update command for admin_datatypes (PostgreSQL).
func (d PsqlDatabase) UpdateAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeParams) UpdateAdminDatatypeCmdPsql {
	return UpdateAdminDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteAdminDatatypeCmdPsql is an audited delete command for PostgreSQL.
type DeleteAdminDatatypeCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminDatatypeID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the context for the command.
func (c DeleteAdminDatatypeCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context for the command.
func (c DeleteAdminDatatypeCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminDatatypeCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminDatatypeCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeleteAdminDatatypeCmdPsql) TableName() string                     { return "admin_datatypes" }

// GetID returns the record ID being deleted.
func (c DeleteAdminDatatypeCmdPsql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminDatatypeCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminDatatype(ctx, mdbp.GetAdminDatatypeParams{AdminDatatypeID: c.id})
}

// Execute deletes the admin datatype record.
func (c DeleteAdminDatatypeCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminDatatype(ctx, mdbp.DeleteAdminDatatypeParams{AdminDatatypeID: c.id})
}

// DeleteAdminDatatypeCmd constructs a delete command for admin_datatypes (PostgreSQL).
func (d PsqlDatabase) DeleteAdminDatatypeCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminDatatypeID) DeleteAdminDatatypeCmdPsql {
	return DeleteAdminDatatypeCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
