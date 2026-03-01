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

// ListFieldsPaginated returns fields with pagination (SQLite).
func (d Database) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdb.ListFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %w", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListFieldsPaginated returns fields with pagination (MySQL).
func (d MysqlDatabase) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdbm.ListFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %w", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListFieldsPaginated returns fields with pagination (PostgreSQL).
func (d PsqlDatabase) ListFieldsPaginated(params PaginationParams) (*[]Fields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListFieldPaginated(d.Context, mdbp.ListFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields paginated: %w", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

// FieldsJSON is used for JSON serialization in model package
type FieldsJSON struct {
	FieldID      string `json:"field_id"`
	ParentID     string `json:"parent_id"`
	SortOrder    string `json:"sort_order"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Validation   string `json:"validation"`
	UIConfig     string `json:"ui_config"`
	Type         string `json:"type"`
	Translatable string `json:"translatable"`
	Roles        string `json:"roles"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// MapFieldJSON converts Fields to FieldsJSON for JSON serialization
func MapFieldJSON(a Fields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.FieldID.String(),
		ParentID:     a.ParentID.String(),
		SortOrder:    fmt.Sprintf("%d", a.SortOrder),
		Name:         a.Name,
		Label:        a.Label,
		Data:         a.Data,
		Validation:   a.Validation,
		UIConfig:     a.UIConfig,
		Type:         a.Type.String(),
		Translatable: fmt.Sprintf("%d", a.Translatable),
		Roles: func() string {
			if a.Roles.Valid {
				return a.Roles.String
			}
			return ""
		}(),
		AuthorID:     a.AuthorID.String(),
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

// ===== UpdateFieldSortOrder Params =====

// UpdateFieldSortOrderParams specifies parameters for updating a field's sort order.
type UpdateFieldSortOrderParams struct {
	FieldID   types.FieldID `json:"field_id"`
	SortOrder int64         `json:"sort_order"`
}

// ===== SQLite UpdateFieldSortOrder Audited Command =====

// UpdateFieldSortOrderCmd is an audited command for updating a field's sort order.
type UpdateFieldSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateFieldSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateFieldSortOrderCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateFieldSortOrderCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateFieldSortOrderCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateFieldSortOrderCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateFieldSortOrderCmd) TableName() string { return "fields" }

// Params returns the command parameters.
func (c UpdateFieldSortOrderCmd) Params() any { return c.params }

// GetID returns the field ID.
func (c UpdateFieldSortOrderCmd) GetID() string { return string(c.params.FieldID) }

// GetBefore retrieves the field before the update.
func (c UpdateFieldSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Fields, error) {
	queries := mdb.New(tx)
	return queries.GetField(ctx, mdb.GetFieldParams{FieldID: c.params.FieldID})
}

// Execute updates the field's sort order in the database.
func (c UpdateFieldSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateFieldSortOrder(ctx, mdb.UpdateFieldSortOrderParams{
		SortOrder: c.params.SortOrder,
		FieldID:   c.params.FieldID,
	})
}

// UpdateFieldSortOrderCmd creates a new update command for a field's sort order.
func (d Database) UpdateFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateFieldSortOrderParams) UpdateFieldSortOrderCmd {
	return UpdateFieldSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateFieldSortOrder updates a field's sort order with audit trail (SQLite).
func (d Database) UpdateFieldSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateFieldSortOrderParams) error {
	cmd := d.UpdateFieldSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetMaxSortOrderByParentID retrieves the maximum sort order for fields under a parent datatype (SQLite).
func (d Database) GetMaxSortOrderByParentID(parentID types.NullableDatatypeID) (int64, error) {
	queries := mdb.New(d.Connection)
	result, err := queries.GetMaxSortOrderByParentID(d.Context, mdb.GetMaxSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ===== MySQL UpdateFieldSortOrder Audited Command =====

// UpdateFieldSortOrderCmdMysql is an audited command for updating a field's sort order on MySQL.
type UpdateFieldSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateFieldSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateFieldSortOrderCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateFieldSortOrderCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateFieldSortOrderCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateFieldSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateFieldSortOrderCmdMysql) TableName() string { return "fields" }

// Params returns the command parameters.
func (c UpdateFieldSortOrderCmdMysql) Params() any { return c.params }

// GetID returns the field ID.
func (c UpdateFieldSortOrderCmdMysql) GetID() string { return string(c.params.FieldID) }

// GetBefore retrieves the field before the update.
func (c UpdateFieldSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Fields, error) {
	queries := mdbm.New(tx)
	return queries.GetField(ctx, mdbm.GetFieldParams{FieldID: c.params.FieldID})
}

// Execute updates the field's sort order in the database.
func (c UpdateFieldSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateFieldSortOrder(ctx, mdbm.UpdateFieldSortOrderParams{
		SortOrder: int32(c.params.SortOrder),
		FieldID:   c.params.FieldID,
	})
}

// UpdateFieldSortOrderCmd creates a new update command for a field's sort order.
func (d MysqlDatabase) UpdateFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateFieldSortOrderParams) UpdateFieldSortOrderCmdMysql {
	return UpdateFieldSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateFieldSortOrder updates a field's sort order with audit trail (MySQL).
func (d MysqlDatabase) UpdateFieldSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateFieldSortOrderParams) error {
	cmd := d.UpdateFieldSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetMaxSortOrderByParentID retrieves the maximum sort order for fields under a parent datatype (MySQL).
func (d MysqlDatabase) GetMaxSortOrderByParentID(parentID types.NullableDatatypeID) (int64, error) {
	queries := mdbm.New(d.Connection)
	result, err := queries.GetMaxSortOrderByParentID(d.Context, mdbm.GetMaxSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ===== PostgreSQL UpdateFieldSortOrder Audited Command =====

// UpdateFieldSortOrderCmdPsql is an audited command for updating a field's sort order on PostgreSQL.
type UpdateFieldSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateFieldSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateFieldSortOrderCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateFieldSortOrderCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateFieldSortOrderCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateFieldSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateFieldSortOrderCmdPsql) TableName() string { return "fields" }

// Params returns the command parameters.
func (c UpdateFieldSortOrderCmdPsql) Params() any { return c.params }

// GetID returns the field ID.
func (c UpdateFieldSortOrderCmdPsql) GetID() string { return string(c.params.FieldID) }

// GetBefore retrieves the field before the update.
func (c UpdateFieldSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Fields, error) {
	queries := mdbp.New(tx)
	return queries.GetField(ctx, mdbp.GetFieldParams{FieldID: c.params.FieldID})
}

// Execute updates the field's sort order in the database.
func (c UpdateFieldSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateFieldSortOrder(ctx, mdbp.UpdateFieldSortOrderParams{
		SortOrder: int32(c.params.SortOrder),
		FieldID:   c.params.FieldID,
	})
}

// UpdateFieldSortOrderCmd creates a new update command for a field's sort order.
func (d PsqlDatabase) UpdateFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateFieldSortOrderParams) UpdateFieldSortOrderCmdPsql {
	return UpdateFieldSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateFieldSortOrder updates a field's sort order with audit trail (PostgreSQL).
func (d PsqlDatabase) UpdateFieldSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateFieldSortOrderParams) error {
	cmd := d.UpdateFieldSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetMaxSortOrderByParentID retrieves the maximum sort order for fields under a parent datatype (PostgreSQL).
func (d PsqlDatabase) GetMaxSortOrderByParentID(parentID types.NullableDatatypeID) (int64, error) {
	queries := mdbp.New(d.Connection)
	result, err := queries.GetMaxSortOrderByParentID(d.Context, mdbp.GetMaxSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// coalesceToInt64 converts the interface{} result from a COALESCE query to int64.
func coalesceToInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}
