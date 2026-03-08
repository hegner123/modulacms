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

// UpdateAdminDatatypeSortOrderParams specifies parameters for updating an admin datatype's sort order.
type UpdateAdminDatatypeSortOrderParams struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	SortOrder       int64                 `json:"sort_order"`
}

// ===== SQLite UpdateAdminDatatypeSortOrder Audited Command =====

type UpdateAdminDatatypeSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeSortOrderCmd) Context() context.Context           { return c.ctx }
func (c UpdateAdminDatatypeSortOrderCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateAdminDatatypeSortOrderCmd) Connection() *sql.DB                { return c.conn }
func (c UpdateAdminDatatypeSortOrderCmd) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateAdminDatatypeSortOrderCmd) TableName() string { return "admin_datatypes" }
func (c UpdateAdminDatatypeSortOrderCmd) Params() any       { return c.params }
func (c UpdateAdminDatatypeSortOrderCmd) GetID() string {
	return string(c.params.AdminDatatypeID)
}

func (c UpdateAdminDatatypeSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.AdminDatatypes, error) {
	queries := mdb.New(tx)
	return queries.GetAdminDatatype(ctx, mdb.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

func (c UpdateAdminDatatypeSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateAdminDatatypeSortOrder(ctx, mdb.UpdateAdminDatatypeSortOrderParams{
		SortOrder:       c.params.SortOrder,
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

func (d Database) UpdateAdminDatatypeSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeSortOrderParams) UpdateAdminDatatypeSortOrderCmd {
	return UpdateAdminDatatypeSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

func (d Database) UpdateAdminDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeSortOrderParams) error {
	cmd := d.UpdateAdminDatatypeSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d Database) GetMaxAdminDatatypeSortOrder(parentID types.NullableAdminDatatypeID) (int64, error) {
	queries := mdb.New(d.Connection)
	if !parentID.Valid {
		result, err := queries.GetMaxAdminDatatypeRootSortOrder(d.Context)
		if err != nil {
			return 0, fmt.Errorf("failed to get max admin datatype root sort order: %w", err)
		}
		return coalesceToInt64(result), nil
	}
	result, err := queries.GetMaxAdminDatatypeSortOrderByParentID(d.Context, mdb.GetMaxAdminDatatypeSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max admin datatype sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ===== MySQL UpdateAdminDatatypeSortOrder Audited Command =====

type UpdateAdminDatatypeSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeSortOrderCmdMysql) Context() context.Context           { return c.ctx }
func (c UpdateAdminDatatypeSortOrderCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateAdminDatatypeSortOrderCmdMysql) Connection() *sql.DB                { return c.conn }
func (c UpdateAdminDatatypeSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateAdminDatatypeSortOrderCmdMysql) TableName() string { return "admin_datatypes" }
func (c UpdateAdminDatatypeSortOrderCmdMysql) Params() any       { return c.params }
func (c UpdateAdminDatatypeSortOrderCmdMysql) GetID() string {
	return string(c.params.AdminDatatypeID)
}

func (c UpdateAdminDatatypeSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.AdminDatatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetAdminDatatype(ctx, mdbm.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

func (c UpdateAdminDatatypeSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateAdminDatatypeSortOrder(ctx, mdbm.UpdateAdminDatatypeSortOrderParams{
		SortOrder:       int32(c.params.SortOrder),
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

func (d MysqlDatabase) UpdateAdminDatatypeSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeSortOrderParams) UpdateAdminDatatypeSortOrderCmdMysql {
	return UpdateAdminDatatypeSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

func (d MysqlDatabase) UpdateAdminDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeSortOrderParams) error {
	cmd := d.UpdateAdminDatatypeSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d MysqlDatabase) GetMaxAdminDatatypeSortOrder(parentID types.NullableAdminDatatypeID) (int64, error) {
	queries := mdbm.New(d.Connection)
	if !parentID.Valid {
		result, err := queries.GetMaxAdminDatatypeRootSortOrder(d.Context)
		if err != nil {
			return 0, fmt.Errorf("failed to get max admin datatype root sort order: %w", err)
		}
		return coalesceToInt64(result), nil
	}
	result, err := queries.GetMaxAdminDatatypeSortOrderByParentID(d.Context, mdbm.GetMaxAdminDatatypeSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max admin datatype sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ===== PostgreSQL UpdateAdminDatatypeSortOrder Audited Command =====

type UpdateAdminDatatypeSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateAdminDatatypeSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateAdminDatatypeSortOrderCmdPsql) Context() context.Context           { return c.ctx }
func (c UpdateAdminDatatypeSortOrderCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateAdminDatatypeSortOrderCmdPsql) Connection() *sql.DB                { return c.conn }
func (c UpdateAdminDatatypeSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateAdminDatatypeSortOrderCmdPsql) TableName() string { return "admin_datatypes" }
func (c UpdateAdminDatatypeSortOrderCmdPsql) Params() any       { return c.params }
func (c UpdateAdminDatatypeSortOrderCmdPsql) GetID() string {
	return string(c.params.AdminDatatypeID)
}

func (c UpdateAdminDatatypeSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.AdminDatatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetAdminDatatype(ctx, mdbp.GetAdminDatatypeParams{AdminDatatypeID: c.params.AdminDatatypeID})
}

func (c UpdateAdminDatatypeSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateAdminDatatypeSortOrder(ctx, mdbp.UpdateAdminDatatypeSortOrderParams{
		SortOrder:       int32(c.params.SortOrder),
		AdminDatatypeID: c.params.AdminDatatypeID,
	})
}

func (d PsqlDatabase) UpdateAdminDatatypeSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateAdminDatatypeSortOrderParams) UpdateAdminDatatypeSortOrderCmdPsql {
	return UpdateAdminDatatypeSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

func (d PsqlDatabase) UpdateAdminDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateAdminDatatypeSortOrderParams) error {
	cmd := d.UpdateAdminDatatypeSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d PsqlDatabase) GetMaxAdminDatatypeSortOrder(parentID types.NullableAdminDatatypeID) (int64, error) {
	queries := mdbp.New(d.Connection)
	if !parentID.Valid {
		result, err := queries.GetMaxAdminDatatypeRootSortOrder(d.Context)
		if err != nil {
			return 0, fmt.Errorf("failed to get max admin datatype root sort order: %w", err)
		}
		return coalesceToInt64(result), nil
	}
	result, err := queries.GetMaxAdminDatatypeSortOrderByParentID(d.Context, mdbp.GetMaxAdminDatatypeSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max admin datatype sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}
