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

// UpdateDatatypeSortOrderParams specifies parameters for updating a datatype's sort order.
type UpdateDatatypeSortOrderParams struct {
	DatatypeID types.DatatypeID `json:"datatype_id"`
	SortOrder  int64            `json:"sort_order"`
}

// ===== SQLite UpdateDatatypeSortOrder Audited Command =====

type UpdateDatatypeSortOrderCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeSortOrderCmd) Context() context.Context           { return c.ctx }
func (c UpdateDatatypeSortOrderCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateDatatypeSortOrderCmd) Connection() *sql.DB                { return c.conn }
func (c UpdateDatatypeSortOrderCmd) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateDatatypeSortOrderCmd) TableName() string { return "datatypes" }
func (c UpdateDatatypeSortOrderCmd) Params() any       { return c.params }
func (c UpdateDatatypeSortOrderCmd) GetID() string     { return string(c.params.DatatypeID) }

func (c UpdateDatatypeSortOrderCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Datatypes, error) {
	queries := mdb.New(tx)
	return queries.GetDatatype(ctx, mdb.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

func (c UpdateDatatypeSortOrderCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateDatatypeSortOrder(ctx, mdb.UpdateDatatypeSortOrderParams{
		SortOrder:  c.params.SortOrder,
		DatatypeID: c.params.DatatypeID,
	})
}

func (d Database) UpdateDatatypeSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeSortOrderParams) UpdateDatatypeSortOrderCmd {
	return UpdateDatatypeSortOrderCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

func (d Database) UpdateDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeSortOrderParams) error {
	cmd := d.UpdateDatatypeSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d Database) GetMaxDatatypeSortOrder(parentID types.NullableDatatypeID) (int64, error) {
	queries := mdb.New(d.Connection)
	if !parentID.Valid {
		result, err := queries.GetMaxDatatypeRootSortOrder(d.Context)
		if err != nil {
			return 0, fmt.Errorf("failed to get max datatype root sort order: %w", err)
		}
		return coalesceToInt64(result), nil
	}
	result, err := queries.GetMaxDatatypeSortOrderByParentID(d.Context, mdb.GetMaxDatatypeSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max datatype sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ===== MySQL UpdateDatatypeSortOrder Audited Command =====

type UpdateDatatypeSortOrderCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeSortOrderCmdMysql) Context() context.Context           { return c.ctx }
func (c UpdateDatatypeSortOrderCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateDatatypeSortOrderCmdMysql) Connection() *sql.DB                { return c.conn }
func (c UpdateDatatypeSortOrderCmdMysql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateDatatypeSortOrderCmdMysql) TableName() string { return "datatypes" }
func (c UpdateDatatypeSortOrderCmdMysql) Params() any       { return c.params }
func (c UpdateDatatypeSortOrderCmdMysql) GetID() string     { return string(c.params.DatatypeID) }

func (c UpdateDatatypeSortOrderCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Datatypes, error) {
	queries := mdbm.New(tx)
	return queries.GetDatatype(ctx, mdbm.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

func (c UpdateDatatypeSortOrderCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateDatatypeSortOrder(ctx, mdbm.UpdateDatatypeSortOrderParams{
		SortOrder:  int32(c.params.SortOrder),
		DatatypeID: c.params.DatatypeID,
	})
}

func (d MysqlDatabase) UpdateDatatypeSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeSortOrderParams) UpdateDatatypeSortOrderCmdMysql {
	return UpdateDatatypeSortOrderCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

func (d MysqlDatabase) UpdateDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeSortOrderParams) error {
	cmd := d.UpdateDatatypeSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d MysqlDatabase) GetMaxDatatypeSortOrder(parentID types.NullableDatatypeID) (int64, error) {
	queries := mdbm.New(d.Connection)
	if !parentID.Valid {
		result, err := queries.GetMaxDatatypeRootSortOrder(d.Context)
		if err != nil {
			return 0, fmt.Errorf("failed to get max datatype root sort order: %w", err)
		}
		return coalesceToInt64(result), nil
	}
	result, err := queries.GetMaxDatatypeSortOrderByParentID(d.Context, mdbm.GetMaxDatatypeSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max datatype sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}

// ===== PostgreSQL UpdateDatatypeSortOrder Audited Command =====

type UpdateDatatypeSortOrderCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateDatatypeSortOrderParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateDatatypeSortOrderCmdPsql) Context() context.Context           { return c.ctx }
func (c UpdateDatatypeSortOrderCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }
func (c UpdateDatatypeSortOrderCmdPsql) Connection() *sql.DB                { return c.conn }
func (c UpdateDatatypeSortOrderCmdPsql) Recorder() audited.ChangeEventRecorder {
	return c.recorder
}
func (c UpdateDatatypeSortOrderCmdPsql) TableName() string { return "datatypes" }
func (c UpdateDatatypeSortOrderCmdPsql) Params() any       { return c.params }
func (c UpdateDatatypeSortOrderCmdPsql) GetID() string     { return string(c.params.DatatypeID) }

func (c UpdateDatatypeSortOrderCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Datatypes, error) {
	queries := mdbp.New(tx)
	return queries.GetDatatype(ctx, mdbp.GetDatatypeParams{DatatypeID: c.params.DatatypeID})
}

func (c UpdateDatatypeSortOrderCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateDatatypeSortOrder(ctx, mdbp.UpdateDatatypeSortOrderParams{
		SortOrder:  int32(c.params.SortOrder),
		DatatypeID: c.params.DatatypeID,
	})
}

func (d PsqlDatabase) UpdateDatatypeSortOrderCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateDatatypeSortOrderParams) UpdateDatatypeSortOrderCmdPsql {
	return UpdateDatatypeSortOrderCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

func (d PsqlDatabase) UpdateDatatypeSortOrder(ctx context.Context, ac audited.AuditContext, s UpdateDatatypeSortOrderParams) error {
	cmd := d.UpdateDatatypeSortOrderCmd(ctx, ac, s)
	return audited.Update(cmd)
}

func (d PsqlDatabase) GetMaxDatatypeSortOrder(parentID types.NullableDatatypeID) (int64, error) {
	queries := mdbp.New(d.Connection)
	if !parentID.Valid {
		result, err := queries.GetMaxDatatypeRootSortOrder(d.Context)
		if err != nil {
			return 0, fmt.Errorf("failed to get max datatype root sort order: %w", err)
		}
		return coalesceToInt64(result), nil
	}
	result, err := queries.GetMaxDatatypeSortOrderByParentID(d.Context, mdbp.GetMaxDatatypeSortOrderByParentIDParams{ParentID: parentID})
	if err != nil {
		return 0, fmt.Errorf("failed to get max datatype sort order by parent id: %w", err)
	}
	return coalesceToInt64(result), nil
}
