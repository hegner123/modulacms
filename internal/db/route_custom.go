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
// SQLITE MAPPERS
//////////////////////////////

// MapRoute converts a sqlc-generated SQLite route to the wrapper type.
func (d Database) MapRoute(a mdb.Routes) Routes {
	return Routes{
		RouteID:      a.RouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateRouteParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateRouteParams(a CreateRouteParams) mdb.CreateRouteParams {
	return mdb.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateRouteParams converts wrapper params to sqlc-generated SQLite update params.
func (d Database) MapUpdateRouteParams(a UpdateRouteParams) mdb.UpdateRouteParams {
	return mdb.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

///////////////////////////////
// SQLITE QUERIES
//////////////////////////////

// CreateRoute creates a new audited route record.
func (d Database) CreateRoute(ctx context.Context, ac audited.AuditContext, s CreateRouteParams) (*Routes, error) {
	cmd := d.NewRouteCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}
	r := d.MapRoute(result)
	return &r, nil
}

// DeleteRoute deletes a route by ID with auditing.
func (d Database) DeleteRoute(ctx context.Context, ac audited.AuditContext, id types.RouteID) error {
	cmd := d.DeleteRouteCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetRouteID retrieves a route ID by slug.
func (d Database) GetRouteID(slug string) (*types.RouteID, error) {
	queries := mdb.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, mdb.GetRouteIDBySlugParams{Slug: types.Slug(slug)})
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// ListRoutesByDatatype returns routes for a specific datatype.
func (d Database) ListRoutesByDatatype(datatypeID types.DatatypeID) (*[]Routes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRoutesByDatatype(d.Context, mdb.ListRoutesByDatatypeParams{
		DatatypeID: types.NullableDatatypeID{ID: datatypeID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get routes by datatype: %w", err)
	}
	res := make([]Routes, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapRoute(v))
	}
	return &res, nil
}

// ListRoutesPaginated returns routes with pagination.
func (d Database) ListRoutesPaginated(params PaginationParams) (*[]Routes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRoutePaginated(d.Context, mdb.ListRoutePaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes paginated: %v", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateRoute updates an existing route with auditing.
func (d Database) UpdateRoute(ctx context.Context, ac audited.AuditContext, s UpdateRouteParams) (*string, error) {
	cmd := d.UpdateRouteCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update route: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &msg, nil
}

///////////////////////////////
// MYSQL MAPPERS
//////////////////////////////

// MapRoute converts a sqlc-generated MySQL route to the wrapper type.
func (d MysqlDatabase) MapRoute(a mdbm.Routes) Routes {
	return Routes{
		RouteID:      a.RouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateRouteParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbm.CreateRouteParams {
	return mdbm.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateRouteParams converts wrapper params to sqlc-generated MySQL update params.
func (d MysqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbm.UpdateRouteParams {
	return mdbm.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

///////////////////////////////
// MYSQL QUERIES
//////////////////////////////

// CreateRoute creates a new audited route record.
func (d MysqlDatabase) CreateRoute(ctx context.Context, ac audited.AuditContext, s CreateRouteParams) (*Routes, error) {
	cmd := d.NewRouteCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}
	r := d.MapRoute(result)
	return &r, nil
}

// DeleteRoute deletes a route by ID with auditing.
func (d MysqlDatabase) DeleteRoute(ctx context.Context, ac audited.AuditContext, id types.RouteID) error {
	cmd := d.DeleteRouteCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetRouteID retrieves a route ID by slug.
func (d MysqlDatabase) GetRouteID(slug string) (*types.RouteID, error) {
	queries := mdbm.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, mdbm.GetRouteIDBySlugParams{Slug: types.Slug(slug)})
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// ListRoutesByDatatype returns routes for a specific datatype.
func (d MysqlDatabase) ListRoutesByDatatype(datatypeID types.DatatypeID) (*[]Routes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRoutesByDatatype(d.Context, mdbm.ListRoutesByDatatypeParams{
		DatatypeID: types.NullableDatatypeID{ID: datatypeID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get routes by datatype: %w", err)
	}
	res := make([]Routes, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapRoute(v))
	}
	return &res, nil
}

// ListRoutesPaginated returns routes with pagination.
func (d MysqlDatabase) ListRoutesPaginated(params PaginationParams) (*[]Routes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRoutePaginated(d.Context, mdbm.ListRoutePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes paginated: %v", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateRoute updates an existing route with auditing.
func (d MysqlDatabase) UpdateRoute(ctx context.Context, ac audited.AuditContext, s UpdateRouteParams) (*string, error) {
	cmd := d.UpdateRouteCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update route: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &msg, nil
}

///////////////////////////////
// POSTGRES MAPPERS
//////////////////////////////

// MapRoute converts a sqlc-generated PostgreSQL route to the wrapper type.
func (d PsqlDatabase) MapRoute(a mdbp.Routes) Routes {
	return Routes{
		RouteID:      a.RouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateRouteParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbp.CreateRouteParams {
	return mdbp.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateRouteParams converts wrapper params to sqlc-generated PostgreSQL update params.
func (d PsqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbp.UpdateRouteParams {
	return mdbp.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

///////////////////////////////
// POSTGRES QUERIES
//////////////////////////////

// CreateRoute creates a new audited route record.
func (d PsqlDatabase) CreateRoute(ctx context.Context, ac audited.AuditContext, s CreateRouteParams) (*Routes, error) {
	cmd := d.NewRouteCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create route: %w", err)
	}
	r := d.MapRoute(result)
	return &r, nil
}

// DeleteRoute deletes a route by ID with auditing.
func (d PsqlDatabase) DeleteRoute(ctx context.Context, ac audited.AuditContext, id types.RouteID) error {
	cmd := d.DeleteRouteCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetRouteID retrieves a route ID by slug.
func (d PsqlDatabase) GetRouteID(slug string) (*types.RouteID, error) {
	queries := mdbp.New(d.Connection)
	id, err := queries.GetRouteIDBySlug(d.Context, mdbp.GetRouteIDBySlugParams{Slug: types.Slug(slug)})
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// ListRoutesByDatatype returns routes for a specific datatype.
func (d PsqlDatabase) ListRoutesByDatatype(datatypeID types.DatatypeID) (*[]Routes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRoutesByDatatype(d.Context, mdbp.ListRoutesByDatatypeParams{
		DatatypeID: types.NullableDatatypeID{ID: datatypeID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get routes by datatype: %w", err)
	}
	res := make([]Routes, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapRoute(v))
	}
	return &res, nil
}

// ListRoutesPaginated returns routes with pagination.
func (d PsqlDatabase) ListRoutesPaginated(params PaginationParams) (*[]Routes, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRoutePaginated(d.Context, mdbp.ListRoutePaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes paginated: %v", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateRoute updates an existing route with auditing.
func (d PsqlDatabase) UpdateRoute(ctx context.Context, ac audited.AuditContext, s UpdateRouteParams) (*string, error) {
	cmd := d.UpdateRouteCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update route: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ----- SQLite CREATE -----

// NewRouteCmd is an audited command for creating a route.
type NewRouteCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewRouteCmd) Context() context.Context              { return c.ctx }
func (c NewRouteCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewRouteCmd) Connection() *sql.DB                   { return c.conn }
func (c NewRouteCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewRouteCmd) TableName() string                     { return "routes" }
func (c NewRouteCmd) Params() any                           { return c.params }
func (c NewRouteCmd) GetID(r mdb.Routes) string             { return string(r.RouteID) }

func (c NewRouteCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Routes, error) {
	queries := mdb.New(tx)
	return queries.CreateRoute(ctx, mdb.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       c.params.Status,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewRouteCmd creates a new SQLite route creation command.
func (d Database) NewRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRouteParams) NewRouteCmd {
	return NewRouteCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateRouteCmd is an audited command for updating a route.
type UpdateRouteCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateRouteCmd) Context() context.Context              { return c.ctx }
func (c UpdateRouteCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateRouteCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateRouteCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateRouteCmd) TableName() string                     { return "routes" }
func (c UpdateRouteCmd) Params() any                           { return c.params }
func (c UpdateRouteCmd) GetID() string                         { return string(c.params.Slug_2) }

func (c UpdateRouteCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Routes, error) {
	queries := mdb.New(tx)
	routeID, err := queries.GetRouteIDBySlug(ctx, mdb.GetRouteIDBySlugParams{Slug: c.params.Slug_2})
	if err != nil {
		return mdb.Routes{}, fmt.Errorf("get route id by slug for before snapshot: %w", err)
	}
	return queries.GetRoute(ctx, mdb.GetRouteParams{RouteID: routeID})
}

func (c UpdateRouteCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateRoute(ctx, mdb.UpdateRouteParams{
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       c.params.Status,
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		Slug_2:       c.params.Slug_2,
	})
}

// UpdateRouteCmd creates a new SQLite route update command.
func (d Database) UpdateRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRouteParams) UpdateRouteCmd {
	return UpdateRouteCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteRouteCmd is an audited command for deleting a route.
type DeleteRouteCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RouteID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteRouteCmd) Context() context.Context              { return c.ctx }
func (c DeleteRouteCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteRouteCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteRouteCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteRouteCmd) TableName() string                     { return "routes" }
func (c DeleteRouteCmd) GetID() string                         { return string(c.id) }

func (c DeleteRouteCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Routes, error) {
	queries := mdb.New(tx)
	return queries.GetRoute(ctx, mdb.GetRouteParams{RouteID: c.id})
}

func (c DeleteRouteCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteRoute(ctx, mdb.DeleteRouteParams{RouteID: c.id})
}

// DeleteRouteCmd creates a new SQLite route deletion command.
func (d Database) DeleteRouteCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RouteID) DeleteRouteCmd {
	return DeleteRouteCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewRouteCmdMysql is an audited command for creating a route on MySQL.
type NewRouteCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewRouteCmdMysql) Context() context.Context              { return c.ctx }
func (c NewRouteCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewRouteCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewRouteCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewRouteCmdMysql) TableName() string                     { return "routes" }
func (c NewRouteCmdMysql) Params() any                           { return c.params }
func (c NewRouteCmdMysql) GetID(r mdbm.Routes) string            { return string(r.RouteID) }

func (c NewRouteCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Routes, error) {
	queries := mdbm.New(tx)
	params := mdbm.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	}
	if err := queries.CreateRoute(ctx, params); err != nil {
		return mdbm.Routes{}, err
	}
	return queries.GetRoute(ctx, mdbm.GetRouteParams{RouteID: params.RouteID})
}

// NewRouteCmd creates a new MySQL route creation command.
func (d MysqlDatabase) NewRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRouteParams) NewRouteCmdMysql {
	return NewRouteCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateRouteCmdMysql is an audited command for updating a route on MySQL.
type UpdateRouteCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateRouteCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateRouteCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateRouteCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateRouteCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateRouteCmdMysql) TableName() string                     { return "routes" }
func (c UpdateRouteCmdMysql) Params() any                           { return c.params }
func (c UpdateRouteCmdMysql) GetID() string                         { return string(c.params.Slug_2) }

func (c UpdateRouteCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Routes, error) {
	queries := mdbm.New(tx)
	routeID, err := queries.GetRouteIDBySlug(ctx, mdbm.GetRouteIDBySlugParams{Slug: c.params.Slug_2})
	if err != nil {
		return mdbm.Routes{}, fmt.Errorf("get route id by slug for before snapshot: %w", err)
	}
	return queries.GetRoute(ctx, mdbm.GetRouteParams{RouteID: routeID})
}

func (c UpdateRouteCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateRoute(ctx, mdbm.UpdateRouteParams{
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		Slug_2:       c.params.Slug_2,
	})
}

// UpdateRouteCmd creates a new MySQL route update command.
func (d MysqlDatabase) UpdateRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRouteParams) UpdateRouteCmdMysql {
	return UpdateRouteCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteRouteCmdMysql is an audited command for deleting a route on MySQL.
type DeleteRouteCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RouteID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteRouteCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteRouteCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteRouteCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteRouteCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteRouteCmdMysql) TableName() string                     { return "routes" }
func (c DeleteRouteCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteRouteCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Routes, error) {
	queries := mdbm.New(tx)
	return queries.GetRoute(ctx, mdbm.GetRouteParams{RouteID: c.id})
}

func (c DeleteRouteCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteRoute(ctx, mdbm.DeleteRouteParams{RouteID: c.id})
}

// DeleteRouteCmd creates a new MySQL route deletion command.
func (d MysqlDatabase) DeleteRouteCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RouteID) DeleteRouteCmdMysql {
	return DeleteRouteCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewRouteCmdPsql is an audited command for creating a route on PostgreSQL.
type NewRouteCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewRouteCmdPsql) Context() context.Context              { return c.ctx }
func (c NewRouteCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewRouteCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewRouteCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewRouteCmdPsql) TableName() string                     { return "routes" }
func (c NewRouteCmdPsql) Params() any                           { return c.params }
func (c NewRouteCmdPsql) GetID(r mdbp.Routes) string            { return string(r.RouteID) }

func (c NewRouteCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Routes, error) {
	queries := mdbp.New(tx)
	return queries.CreateRoute(ctx, mdbp.CreateRouteParams{
		RouteID:      types.NewRouteID(),
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

// NewRouteCmd creates a new PostgreSQL route creation command.
func (d PsqlDatabase) NewRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRouteParams) NewRouteCmdPsql {
	return NewRouteCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateRouteCmdPsql is an audited command for updating a route on PostgreSQL.
type UpdateRouteCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRouteParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateRouteCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateRouteCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateRouteCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateRouteCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateRouteCmdPsql) TableName() string                     { return "routes" }
func (c UpdateRouteCmdPsql) Params() any                           { return c.params }
func (c UpdateRouteCmdPsql) GetID() string                         { return string(c.params.Slug_2) }

func (c UpdateRouteCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Routes, error) {
	queries := mdbp.New(tx)
	routeID, err := queries.GetRouteIDBySlug(ctx, mdbp.GetRouteIDBySlugParams{Slug: c.params.Slug_2})
	if err != nil {
		return mdbp.Routes{}, fmt.Errorf("get route id by slug for before snapshot: %w", err)
	}
	return queries.GetRoute(ctx, mdbp.GetRouteParams{RouteID: routeID})
}

func (c UpdateRouteCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateRoute(ctx, mdbp.UpdateRouteParams{
		Slug:         c.params.Slug,
		Title:        c.params.Title,
		Status:       int32(c.params.Status),
		AuthorID:     c.params.AuthorID,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		Slug_2:       c.params.Slug_2,
	})
}

// UpdateRouteCmd creates a new PostgreSQL route update command.
func (d PsqlDatabase) UpdateRouteCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRouteParams) UpdateRouteCmdPsql {
	return UpdateRouteCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteRouteCmdPsql is an audited command for deleting a route on PostgreSQL.
type DeleteRouteCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RouteID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteRouteCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteRouteCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteRouteCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteRouteCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteRouteCmdPsql) TableName() string                     { return "routes" }
func (c DeleteRouteCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteRouteCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Routes, error) {
	queries := mdbp.New(tx)
	return queries.GetRoute(ctx, mdbp.GetRouteParams{RouteID: c.id})
}

func (c DeleteRouteCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteRoute(ctx, mdbp.DeleteRouteParams{RouteID: c.id})
}

// DeleteRouteCmd creates a new PostgreSQL route deletion command.
func (d PsqlDatabase) DeleteRouteCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RouteID) DeleteRouteCmdPsql {
	return DeleteRouteCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
