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

// ListAdminFieldByRouteIdRow represents a result row from listing admin fields by route ID.
type ListAdminFieldByRouteIdRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	ValidationID types.NullableAdminValidationID `json:"validation_id"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
}

// ListAdminFieldsByDatatypeIDRow represents a result row from listing admin fields by datatype ID.
type ListAdminFieldsByDatatypeIDRow struct {
	AdminFieldID types.AdminFieldID            `json:"admin_field_id"`
	ParentID     types.NullableAdminDatatypeID `json:"parent_id"`
	Label        string                        `json:"label"`
	Data         string                        `json:"data"`
	ValidationID types.NullableAdminValidationID `json:"validation_id"`
	UIConfig     string                        `json:"ui_config"`
	Type         types.FieldType               `json:"type"`
}

// UtilityGetAdminfieldsRow represents a result row from utility admin fields retrieval.
type UtilityGetAdminfieldsRow struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	Label        string             `json:"label"`
}

// ListAdminFieldsPaginated returns admin fields with pagination (SQLite).
func (d Database) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdb.ListAdminFieldPaginatedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %w", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// MapAdminFieldJSON converts AdminFields to FieldsJSON for tree building
// by mapping admin field ID into the public FieldsJSON shape.
func MapAdminFieldJSON(a AdminFields) FieldsJSON {
	return FieldsJSON{
		FieldID:      a.AdminFieldID.String(),
		ParentID:     a.ParentID.String(),
		SortOrder:    fmt.Sprintf("%d", a.SortOrder),
		Name:         a.Name,
		Label:        a.Label,
		Data:         a.Data,
		ValidationID: a.ValidationID.String(),
		UIConfig:     a.UIConfig,
		Type:         a.Type.String(),
		Translatable: fmt.Sprintf("%t", a.Translatable),
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

// ===== UpdateAdminFieldSortOrder Params =====

// UpdateAdminFieldSortOrderParams specifies parameters for updating an admin field's sort order.
type UpdateAdminFieldSortOrderParams struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	SortOrder    int64              `json:"sort_order"`
}

// ===== SQLite UpdateAdminFieldSortOrder Audited Command =====

// UpdateAdminFieldSortOrderCmd is an audited command for updating an admin field's sort order.
type UpdateAdminFieldSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateAdminFieldSortOrderCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateAdminFieldSortOrderCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminFieldSortOrderCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminFieldSortOrderCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdateAdminFieldSortOrderCmd) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c UpdateAdminFieldSortOrderCmd) Params() any { return c.params }

// GetID returns the admin field ID.
func (c UpdateAdminFieldSortOrderCmd) GetID() string { return string(c.params.AdminFieldID) }

// GetBefore retrieves the admin field before the update.
func (c UpdateAdminFieldSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminFields, error) {
	queries := mdb.New(tx)
	return queries.GetAdminField(ctx, mdb.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

// Execute updates the admin field's sort order in the database.
func (c UpdateAdminFieldSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminFieldSortOrder(ctx, mdb.UpdateAdminFieldSortOrderParams{
		SortOrder:    c.params.SortOrder,
		AdminFieldID: c.params.AdminFieldID,
	})
}

// UpdateAdminFieldSortOrderCmd creates a new update command for an admin field's sort order.
func (d Database) UpdateAdminFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldSortOrderParams) UpdateAdminFieldSortOrderCmd {
	return UpdateAdminFieldSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// UpdateAdminFieldSortOrder updates an admin field's sort order with audit trail (SQLite).
func (d Database) UpdateAdminFieldSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldSortOrderParams) error {
	cmd := d.UpdateAdminFieldSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// GetMaxAdminSortOrderByParentID retrieves the maximum sort order for admin fields under a parent admin datatype (SQLite).
func (d Database) GetMaxAdminSortOrderByParentID(parentID types.NullableAdminDatatypeID) (int64, error) {
	queries := mdb.New(d.Connection)
	result, err := queries.GetMaxAdminSortOrderByParentID(d.Context, mdb.GetMaxAdminSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max admin sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ===== MySQL UpdateAdminFieldSortOrder Audited Command =====

// ===== PostgreSQL UpdateAdminFieldSortOrder Audited Command =====

///////////////////////////////
// LIST ADMIN FIELDS BY DATATYPE ID
//////////////////////////////

// ListAdminFieldsByDatatypeID retrieves admin fields by their parent admin datatype ID (SQLite).
func (d Database) ListAdminFieldsByDatatypeID(datatypeID types.NullableAdminDatatypeID) (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentID(d.Context, mdb.ListAdminFieldByParentIDParams{ParentID: datatypeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin fields by datatype id: %w", err)
	}
	res := make([]AdminFields, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminField(v))
	}
	return &res, nil
}

// MYSQL

// GetMaxAdminSortOrderByParentID retrieves the maximum sort order for admin fields under a parent admin datatype (MySQL).
func (d MysqlDatabase) GetMaxAdminSortOrderByParentID(parentID types.NullableAdminDatatypeID) (int64, error) {
	queries := mdbm.New(d.Connection)
	result, err := queries.GetMaxAdminSortOrderByParentID(d.Context, mdbm.GetMaxAdminSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max admin sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ListAdminFieldsByDatatypeID retrieves admin fields by their parent admin datatype ID (MySQL).
func (d MysqlDatabase) ListAdminFieldsByDatatypeID(datatypeID types.NullableAdminDatatypeID) (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentID(d.Context, mdbm.ListAdminFieldByParentIDParams{ParentID: datatypeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin fields by datatype id: %w", err)
	}
	res := make([]AdminFields, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminField(v))
	}
	return &res, nil
}

// ListAdminFieldsPaginated returns admin fields with pagination (MySQL).
func (d MysqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbm.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %w", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminFieldSortOrderCmdMysql is an audited command for updating an admin field's sort order on MySQL.
type UpdateAdminFieldSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateAdminFieldSortOrderCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateAdminFieldSortOrderCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminFieldSortOrderCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminFieldSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}

// TableName returns the table name.
func (c UpdateAdminFieldSortOrderCmdMysql) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c UpdateAdminFieldSortOrderCmdMysql) Params() any { return c.params }

// GetID returns the admin field ID.
func (c UpdateAdminFieldSortOrderCmdMysql) GetID() string { return string(c.params.AdminFieldID) }

// GetBefore retrieves the admin field before the update.
func (c UpdateAdminFieldSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminFields, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminField(ctx, mdbm.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

// Execute updates the admin field's sort order in the database.
func (c UpdateAdminFieldSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminFieldSortOrder(ctx, mdbm.UpdateAdminFieldSortOrderParams{
		SortOrder:    int32(c.params.SortOrder),
		AdminFieldID: c.params.AdminFieldID,
	})
}

// UpdateAdminFieldSortOrderCmd creates a new update command for an admin field's sort order.
func (d MysqlDatabase) UpdateAdminFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldSortOrderParams) UpdateAdminFieldSortOrderCmdMysql {
	return UpdateAdminFieldSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// UpdateAdminFieldSortOrder updates an admin field's sort order with audit trail (MySQL).
func (d MysqlDatabase) UpdateAdminFieldSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldSortOrderParams) error {
	cmd := d.UpdateAdminFieldSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// PSQL

// GetMaxAdminSortOrderByParentID retrieves the maximum sort order for admin fields under a parent admin datatype (PostgreSQL).
func (d PsqlDatabase) GetMaxAdminSortOrderByParentID(parentID types.NullableAdminDatatypeID) (int64, error) {
	queries := mdbp.New(d.Connection)
	result, err := queries.GetMaxAdminSortOrderByParentID(d.Context, mdbp.GetMaxAdminSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max admin sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ListAdminFieldsByDatatypeID retrieves admin fields by their parent admin datatype ID (PostgreSQL).
func (d PsqlDatabase) ListAdminFieldsByDatatypeID(datatypeID types.NullableAdminDatatypeID) (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldByParentID(d.Context, mdbp.ListAdminFieldByParentIDParams{ParentID: datatypeID})
	if err != nil {
		return nil, fmt.Errorf("failed to get admin fields by datatype id: %w", err)
	}
	res := make([]AdminFields, 0, len(rows))
	for _, v := range rows {
		res = append(res, d.MapAdminField(v))
	}
	return &res, nil
}

// ListAdminFieldsPaginated returns admin fields with pagination (PostgreSQL).
func (d PsqlDatabase) ListAdminFieldsPaginated(params PaginationParams) (*[]AdminFields, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListAdminFieldPaginated(d.Context, mdbp.ListAdminFieldPaginatedParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields paginated: %w", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateAdminFieldSortOrderCmdPsql is an audited command for updating an admin field's sort order on PostgreSQL.
type UpdateAdminFieldSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminFieldSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdateAdminFieldSortOrderCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateAdminFieldSortOrderCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateAdminFieldSortOrderCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateAdminFieldSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}

// TableName returns the table name.
func (c UpdateAdminFieldSortOrderCmdPsql) TableName() string { return "admin_fields" }

// Params returns the command parameters.
func (c UpdateAdminFieldSortOrderCmdPsql) Params() any { return c.params }

// GetID returns the admin field ID.
func (c UpdateAdminFieldSortOrderCmdPsql) GetID() string { return string(c.params.AdminFieldID) }

// GetBefore retrieves the admin field before the update.
func (c UpdateAdminFieldSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminFields, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminField(ctx, mdbp.GetAdminFieldParams{AdminFieldID: c.params.AdminFieldID})
}

// Execute updates the admin field's sort order in the database.
func (c UpdateAdminFieldSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminFieldSortOrder(ctx, mdbp.UpdateAdminFieldSortOrderParams{
		SortOrder:    int32(c.params.SortOrder),
		AdminFieldID: c.params.AdminFieldID,
	})
}

// UpdateAdminFieldSortOrderCmd creates a new update command for an admin field's sort order.
func (d PsqlDatabase) UpdateAdminFieldSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminFieldSortOrderParams) UpdateAdminFieldSortOrderCmdPsql {
	return UpdateAdminFieldSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// UpdateAdminFieldSortOrder updates an admin field's sort order with audit trail (PostgreSQL).
func (d PsqlDatabase) UpdateAdminFieldSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminFieldSortOrderParams) error {
	cmd := d.UpdateAdminFieldSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}
