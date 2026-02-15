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

// AdminRoutes represents a CMS admin route resource.
type AdminRoutes struct {
	AdminRouteID types.AdminRouteID   `json:"admin_route_id"`
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}

// CreateAdminRouteParams contains parameters for creating an admin route.
type CreateAdminRouteParams struct {
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
}

// UpdateAdminRouteParams contains parameters for updating an admin route.
type UpdateAdminRouteParams struct {
	Slug         types.Slug           `json:"slug"`
	Title        string               `json:"title"`
	Status       int64                `json:"status"`
	AuthorID     types.NullableUserID `json:"author_id"`
	DateCreated  types.Timestamp      `json:"date_created"`
	DateModified types.Timestamp      `json:"date_modified"`
	Slug_2       types.Slug           `json:"slug_2"`
}

// UtilityGetAdminRoutesRow contains the result of retrieving admin routes.
type UtilityGetAdminRoutesRow struct {
	AdminRouteID types.AdminRouteID `json:"admin_route_id"`
	Slug         types.Slug         `json:"slug"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringAdminRoute converts AdminRoutes to StringAdminRoutes for table display.
func MapStringAdminRoute(a AdminRoutes) StringAdminRoutes {
	return StringAdminRoutes{
		AdminRouteID: a.AdminRouteID.String(),
		Slug:         string(a.Slug),
		Title:        a.Title,
		Status:       fmt.Sprintf("%d", a.Status),
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

// MapAdminRoute converts a sqlc-generated SQLite type to the wrapper type.
func (d Database) MapAdminRoute(a mdb.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateAdminRouteParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdb.CreateAdminRouteParams {
	return mdb.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminRouteParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdb.UpdateAdminRouteParams {
	return mdb.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

// QUERIES

// CountAdminRoutes returns the total count of admin routes.
func (d Database) CountAdminRoutes() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminRoute inserts a new admin route record.
func (d Database) CreateAdminRoute(ctx context.Context, ac audited.AuditContext, s CreateAdminRouteParams) (*AdminRoutes, error) {
	cmd := d.NewAdminRouteCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminRoute: %w", err)
	}
	r := d.MapAdminRoute(result)
	return &r, nil
}

// CreateAdminRouteTable creates the admin_routes table.
func (d Database) CreateAdminRouteTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateAdminRouteTable(d.Context)
	return err
}

// DeleteAdminRoute removes an admin route record.
func (d Database) DeleteAdminRoute(ctx context.Context, ac audited.AuditContext, id types.AdminRouteID) error {
	cmd := d.DeleteAdminRouteCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminRoute retrieves an admin route by slug.
func (d Database) GetAdminRoute(slug types.Slug) (*AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, mdb.GetAdminRouteBySlugParams{Slug: slug})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}

// ListAdminRoutes returns all admin routes.
func (d Database) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminRoutesPaginated returns a page of admin routes.
func (d Database) ListAdminRoutesPaginated(params PaginationParams) (*[]AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminRoutePaginated(d.Context, mdb.ListAdminRoutePaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminRoutes paginated: %v", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminRoute modifies an existing admin route record.
func (d Database) UpdateAdminRoute(ctx context.Context, ac audited.AuditContext, s UpdateAdminRouteParams) (*string, error) {
	cmd := d.UpdateAdminRouteCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminRoute: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapAdminRoute converts a sqlc-generated MySQL type to the wrapper type.
func (d MysqlDatabase) MapAdminRoute(a mdbm.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateAdminRouteParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbm.CreateAdminRouteParams {
	return mdbm.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminRouteParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbm.UpdateAdminRouteParams {
	return mdbm.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

// QUERIES

// CountAdminRoutes returns the total count of admin routes.
func (d MysqlDatabase) CountAdminRoutes() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountAdminroute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminRoute inserts a new admin route record.
func (d MysqlDatabase) CreateAdminRoute(ctx context.Context, ac audited.AuditContext, s CreateAdminRouteParams) (*AdminRoutes, error) {
	cmd := d.NewAdminRouteCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminRoute: %w", err)
	}
	r := d.MapAdminRoute(result)
	return &r, nil
}

// CreateAdminRouteTable creates the admin_routes table.
func (d MysqlDatabase) CreateAdminRouteTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateAdminRouteTable(d.Context)
	return err
}

// DeleteAdminRoute removes an admin route record.
func (d MysqlDatabase) DeleteAdminRoute(ctx context.Context, ac audited.AuditContext, id types.AdminRouteID) error {
	cmd := d.DeleteAdminRouteCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminRoute retrieves an admin route by slug.
func (d MysqlDatabase) GetAdminRoute(slug types.Slug) (*AdminRoutes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, mdbm.GetAdminRouteBySlugParams{Slug: slug})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}

// ListAdminRoutes returns all admin routes.
func (d MysqlDatabase) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminRoutesPaginated returns a page of admin routes.
func (d MysqlDatabase) ListAdminRoutesPaginated(params PaginationParams) (*[]AdminRoutes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminRoutePaginated(d.Context, mdbm.ListAdminRoutePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminRoutes paginated: %v", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminRoute modifies an existing admin route record.
func (d MysqlDatabase) UpdateAdminRoute(ctx context.Context, ac audited.AuditContext, s UpdateAdminRouteParams) (*string, error) {
	cmd := d.UpdateAdminRouteCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminRoute: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapAdminRoute converts a sqlc-generated PostgreSQL type to the wrapper type.
func (d PsqlDatabase) MapAdminRoute(a mdbp.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateAdminRouteParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbp.CreateAdminRouteParams {
	return mdbp.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateAdminRouteParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbp.UpdateAdminRouteParams {
	return mdbp.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

// QUERIES

// CountAdminRoutes returns the total count of admin routes.
func (d PsqlDatabase) CountAdminRoutes() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountAdminroute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateAdminRoute inserts a new admin route record.
func (d PsqlDatabase) CreateAdminRoute(ctx context.Context, ac audited.AuditContext, s CreateAdminRouteParams) (*AdminRoutes, error) {
	cmd := d.NewAdminRouteCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create adminRoute: %w", err)
	}
	r := d.MapAdminRoute(result)
	return &r, nil
}

// CreateAdminRouteTable creates the admin_routes table.
func (d PsqlDatabase) CreateAdminRouteTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateAdminRouteTable(d.Context)
	return err
}

// DeleteAdminRoute removes an admin route record.
func (d PsqlDatabase) DeleteAdminRoute(ctx context.Context, ac audited.AuditContext, id types.AdminRouteID) error {
	cmd := d.DeleteAdminRouteCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetAdminRoute retrieves an admin route by slug.
func (d PsqlDatabase) GetAdminRoute(slug types.Slug) (*AdminRoutes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, mdbp.GetAdminRouteBySlugParams{Slug: slug})
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}

// ListAdminRoutes returns all admin routes.
func (d PsqlDatabase) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListAdminRoutesPaginated returns a page of admin routes.
func (d PsqlDatabase) ListAdminRoutesPaginated(params PaginationParams) (*[]AdminRoutes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminRoutePaginated(d.Context, mdbp.ListAdminRoutePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminRoutes paginated: %v", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminRoute modifies an existing admin route record.
func (d PsqlDatabase) UpdateAdminRoute(ctx context.Context, ac audited.AuditContext, s UpdateAdminRouteParams) (*string, error) {
	cmd := d.UpdateAdminRouteCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update adminRoute: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &msg, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

// NewAdminRouteCmd is an audited command for create on admin_routes.
type NewAdminRouteCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewAdminRouteCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewAdminRouteCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminRouteCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminRouteCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c NewAdminRouteCmd) TableName() string                     { return "admin_routes" }

// Params returns the command parameters.
func (c NewAdminRouteCmd) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewAdminRouteCmd) GetID(u mdb.AdminRoutes) string        { return string(u.AdminRouteID) }

// Execute executes the create command within a transaction.
func (c NewAdminRouteCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.AdminRoutes, error) {
	queries := mdb.New(tx)
	return queries.CreateAdminRoute(ctx, mdb.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       c.params.Status,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewAdminRouteCmd creates a new create command for admin routes.
func (d Database) NewAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminRouteParams) NewAdminRouteCmd {
	return NewAdminRouteCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateAdminRouteCmd is an audited command for update on admin_routes.
type UpdateAdminRouteCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateAdminRouteCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateAdminRouteCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminRouteCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminRouteCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c UpdateAdminRouteCmd) TableName() string                     { return "admin_routes" }

// Params returns the command parameters.
func (c UpdateAdminRouteCmd) Params() any                           { return c.params }

// GetID returns the record ID to update.
func (c UpdateAdminRouteCmd) GetID() string                         { return string(c.params.Slug_2) }

// GetBefore retrieves the record before modification.
func (c UpdateAdminRouteCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminRoutes, error) {
	queries := mdb.New(tx)
	return queries.GetAdminRouteBySlug(ctx, mdb.GetAdminRouteBySlugParams{Slug: c.params.Slug_2})
}

// Execute executes the update command within a transaction.
func (c UpdateAdminRouteCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminRoute(ctx, mdb.UpdateAdminRouteParams{
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       c.params.Status,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		Slug_2:       c.params.Slug_2,
	})
}

// UpdateAdminRouteCmd creates a new update command for admin routes.
func (d Database) UpdateAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminRouteParams) UpdateAdminRouteCmd {
	return UpdateAdminRouteCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteAdminRouteCmd is an audited command for delete on admin_routes.
type DeleteAdminRouteCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminRouteID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteAdminRouteCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteAdminRouteCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminRouteCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminRouteCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c DeleteAdminRouteCmd) TableName() string                     { return "admin_routes" }

// GetID returns the record ID to delete.
func (c DeleteAdminRouteCmd) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminRouteCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminRoutes, error) {
	queries := mdb.New(tx)
	return queries.GetAdminRouteById(ctx, mdb.GetAdminRouteByIdParams{AdminRouteID: c.id})
}

// Execute executes the delete command within a transaction.
func (c DeleteAdminRouteCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteAdminRoute(ctx, mdb.DeleteAdminRouteParams{AdminRouteID: c.id})
}

// DeleteAdminRouteCmd creates a new delete command for admin routes.
func (d Database) DeleteAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminRouteID) DeleteAdminRouteCmd {
	return DeleteAdminRouteCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewAdminRouteCmdMysql is an audited command for create on admin_routes for MySQL.
type NewAdminRouteCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewAdminRouteCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewAdminRouteCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminRouteCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminRouteCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c NewAdminRouteCmdMysql) TableName() string                     { return "admin_routes" }

// Params returns the command parameters.
func (c NewAdminRouteCmdMysql) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewAdminRouteCmdMysql) GetID(u mdbm.AdminRoutes) string      { return string(u.AdminRouteID) }

// Execute executes the create command within a transaction.
func (c NewAdminRouteCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.AdminRoutes, error) {
	queries := mdbm.New(tx)
	id := types.NewAdminRouteID()
	err := queries.CreateAdminRoute(ctx, mdbm.CreateAdminRouteParams{
		AdminRouteID: id,
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
	if err != nil {
		return mdbm.AdminRoutes{}, err
	}
	return queries.GetAdminRouteById(ctx, mdbm.GetAdminRouteByIdParams{AdminRouteID: id})
}

// NewAdminRouteCmd creates a new create command for admin routes.
func (d MysqlDatabase) NewAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminRouteParams) NewAdminRouteCmdMysql {
	return NewAdminRouteCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateAdminRouteCmdMysql is an audited command for update on admin_routes for MySQL.
type UpdateAdminRouteCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateAdminRouteCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateAdminRouteCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminRouteCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminRouteCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c UpdateAdminRouteCmdMysql) TableName() string                     { return "admin_routes" }

// Params returns the command parameters.
func (c UpdateAdminRouteCmdMysql) Params() any                           { return c.params }

// GetID returns the record ID to update.
func (c UpdateAdminRouteCmdMysql) GetID() string                         { return string(c.params.Slug_2) }

// GetBefore retrieves the record before modification.
func (c UpdateAdminRouteCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminRoutes, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminRouteBySlug(ctx, mdbm.GetAdminRouteBySlugParams{Slug: c.params.Slug_2})
}

// Execute executes the update command within a transaction.
func (c UpdateAdminRouteCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminRoute(ctx, mdbm.UpdateAdminRouteParams{
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		Slug_2:       c.params.Slug_2,
	})
}

// UpdateAdminRouteCmd creates a new update command for admin routes.
func (d MysqlDatabase) UpdateAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminRouteParams) UpdateAdminRouteCmdMysql {
	return UpdateAdminRouteCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteAdminRouteCmdMysql is an audited command for delete on admin_routes for MySQL.
type DeleteAdminRouteCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminRouteID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteAdminRouteCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteAdminRouteCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminRouteCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminRouteCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c DeleteAdminRouteCmdMysql) TableName() string                     { return "admin_routes" }

// GetID returns the record ID to delete.
func (c DeleteAdminRouteCmdMysql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminRouteCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminRoutes, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminRouteById(ctx, mdbm.GetAdminRouteByIdParams{AdminRouteID: c.id})
}

// Execute executes the delete command within a transaction.
func (c DeleteAdminRouteCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteAdminRoute(ctx, mdbm.DeleteAdminRouteParams{AdminRouteID: c.id})
}

// DeleteAdminRouteCmd creates a new delete command for admin routes.
func (d MysqlDatabase) DeleteAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminRouteID) DeleteAdminRouteCmdMysql {
	return DeleteAdminRouteCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewAdminRouteCmdPsql is an audited command for create on admin_routes for PostgreSQL.
type NewAdminRouteCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateAdminRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewAdminRouteCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewAdminRouteCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewAdminRouteCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewAdminRouteCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c NewAdminRouteCmdPsql) TableName() string                     { return "admin_routes" }

// Params returns the command parameters.
func (c NewAdminRouteCmdPsql) Params() any                           { return c.params }

// GetID returns the ID from the created record.
func (c NewAdminRouteCmdPsql) GetID(u mdbp.AdminRoutes) string      { return string(u.AdminRouteID) }

// Execute executes the create command within a transaction.
func (c NewAdminRouteCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.AdminRoutes, error) {
	queries := mdbp.New(tx)
	return queries.CreateAdminRoute(ctx, mdbp.CreateAdminRouteParams{
		AdminRouteID: types.NewAdminRouteID(),
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewAdminRouteCmd creates a new create command for admin routes.
func (d PsqlDatabase) NewAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateAdminRouteParams) NewAdminRouteCmdPsql {
	return NewAdminRouteCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateAdminRouteCmdPsql is an audited command for update on admin_routes for PostgreSQL.
type UpdateAdminRouteCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateAdminRouteCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateAdminRouteCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminRouteCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminRouteCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c UpdateAdminRouteCmdPsql) TableName() string                     { return "admin_routes" }

// Params returns the command parameters.
func (c UpdateAdminRouteCmdPsql) Params() any                           { return c.params }

// GetID returns the record ID to update.
func (c UpdateAdminRouteCmdPsql) GetID() string                         { return string(c.params.Slug_2) }

// GetBefore retrieves the record before modification.
func (c UpdateAdminRouteCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminRoutes, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminRouteBySlug(ctx, mdbp.GetAdminRouteBySlugParams{Slug: c.params.Slug_2})
}

// Execute executes the update command within a transaction.
func (c UpdateAdminRouteCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminRoute(ctx, mdbp.UpdateAdminRouteParams{
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		Slug_2:       c.params.Slug_2,
	})
}

// UpdateAdminRouteCmd creates a new update command for admin routes.
func (d PsqlDatabase) UpdateAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminRouteParams) UpdateAdminRouteCmdPsql {
	return UpdateAdminRouteCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteAdminRouteCmdPsql is an audited command for delete on admin_routes for PostgreSQL.
type DeleteAdminRouteCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.AdminRouteID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteAdminRouteCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteAdminRouteCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteAdminRouteCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteAdminRouteCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for audit records.
func (c DeleteAdminRouteCmdPsql) TableName() string                     { return "admin_routes" }

// GetID returns the record ID to delete.
func (c DeleteAdminRouteCmdPsql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the record before deletion.
func (c DeleteAdminRouteCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminRoutes, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminRoute(ctx, mdbp.GetAdminRouteParams{AdminRouteID: c.id})
}

// Execute executes the delete command within a transaction.
func (c DeleteAdminRouteCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteAdminRoute(ctx, mdbp.DeleteAdminRouteParams{AdminRouteID: c.id})
}

// DeleteAdminRouteCmd creates a new delete command for admin routes.
func (d PsqlDatabase) DeleteAdminRouteCmd(ctx context.Context, auditCtx audited.AuditContext, id types.AdminRouteID) DeleteAdminRouteCmdPsql {
	return DeleteAdminRouteCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
