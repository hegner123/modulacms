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

// MapPermission converts a sqlc-generated SQLite Permissions type to the wrapper type.
func (d Database) MapPermission(a mdb.Permissions) Permissions {
	return Permissions{
		PermissionID:    a.PermissionID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected != 0,
	}
}

// MapCreatePermissionParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreatePermissionParams(a CreatePermissionParams) mdb.CreatePermissionParams {
	sp := int64(0)
	if a.SystemProtected {
		sp = 1
	}
	return mdb.CreatePermissionParams{
		PermissionID:    types.NewPermissionID(),
		Label:           a.Label,
		SystemProtected: sp,
	}
}

// MapUpdatePermissionParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdatePermissionParams(a UpdatePermissionParams) mdb.UpdatePermissionParams {
	sp := int64(0)
	if a.SystemProtected {
		sp = 1
	}
	return mdb.UpdatePermissionParams{
		Label:           a.Label,
		SystemProtected: sp,
		PermissionID:    a.PermissionID,
	}
}

///////////////////////////////
// MYSQL MAPPERS
//////////////////////////////

// MapPermission converts a sqlc-generated MySQL Permissions type to the wrapper type.
func (d MysqlDatabase) MapPermission(a mdbm.Permissions) Permissions {
	return Permissions{
		PermissionID:    a.PermissionID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapCreatePermissionParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreatePermissionParams(a CreatePermissionParams) mdbm.CreatePermissionParams {
	return mdbm.CreatePermissionParams{
		PermissionID:    types.NewPermissionID(),
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapUpdatePermissionParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdatePermissionParams(a UpdatePermissionParams) mdbm.UpdatePermissionParams {
	return mdbm.UpdatePermissionParams{
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
		PermissionID:    a.PermissionID,
	}
}

///////////////////////////////
// POSTGRESQL MAPPERS
//////////////////////////////

// MapPermission converts a sqlc-generated PostgreSQL Permissions type to the wrapper type.
func (d PsqlDatabase) MapPermission(a mdbp.Permissions) Permissions {
	return Permissions{
		PermissionID:    a.PermissionID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapCreatePermissionParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreatePermissionParams(a CreatePermissionParams) mdbp.CreatePermissionParams {
	return mdbp.CreatePermissionParams{
		PermissionID:    types.NewPermissionID(),
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapUpdatePermissionParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdatePermissionParams(a UpdatePermissionParams) mdbp.UpdatePermissionParams {
	return mdbp.UpdatePermissionParams{
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
		PermissionID:    a.PermissionID,
	}
}

///////////////////////////////
// SQLITE QUERIES
//////////////////////////////

// CreatePermission creates a new audited permission record.
func (d Database) CreatePermission(ctx context.Context, ac audited.AuditContext, s CreatePermissionParams) (*Permissions, error) {
	cmd := d.NewPermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	r := d.MapPermission(result)
	return &r, nil
}

// DeletePermission deletes a permission by ID with auditing.
func (d Database) DeletePermission(ctx context.Context, ac audited.AuditContext, id types.PermissionID) error {
	cmd := d.DeletePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdatePermission updates an existing permission with auditing.
func (d Database) UpdatePermission(ctx context.Context, ac audited.AuditContext, s UpdatePermissionParams) (*string, error) {
	cmd := d.UpdatePermissionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL QUERIES
//////////////////////////////

// CreatePermission creates a new audited permission record.
func (d MysqlDatabase) CreatePermission(ctx context.Context, ac audited.AuditContext, s CreatePermissionParams) (*Permissions, error) {
	cmd := d.NewPermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	r := d.MapPermission(result)
	return &r, nil
}

// DeletePermission deletes a permission by ID with auditing.
func (d MysqlDatabase) DeletePermission(ctx context.Context, ac audited.AuditContext, id types.PermissionID) error {
	cmd := d.DeletePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdatePermission updates an existing permission with auditing.
func (d MysqlDatabase) UpdatePermission(ctx context.Context, ac audited.AuditContext, s UpdatePermissionParams) (*string, error) {
	cmd := d.UpdatePermissionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES QUERIES
//////////////////////////////

// CreatePermission creates a new audited permission record.
func (d PsqlDatabase) CreatePermission(ctx context.Context, ac audited.AuditContext, s CreatePermissionParams) (*Permissions, error) {
	cmd := d.NewPermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	r := d.MapPermission(result)
	return &r, nil
}

// DeletePermission deletes a permission by ID with auditing.
func (d PsqlDatabase) DeletePermission(ctx context.Context, ac audited.AuditContext, id types.PermissionID) error {
	cmd := d.DeletePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdatePermission updates an existing permission with auditing.
func (d PsqlDatabase) UpdatePermission(ctx context.Context, ac audited.AuditContext, s UpdatePermissionParams) (*string, error) {
	cmd := d.UpdatePermissionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewPermissionCmd is an audited command for creating permissions in SQLite.
type NewPermissionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePermissionParams
	conn     *sql.DB
}

func (c NewPermissionCmd) Context() context.Context              { return c.ctx }
func (c NewPermissionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewPermissionCmd) Connection() *sql.DB                   { return c.conn }
func (c NewPermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c NewPermissionCmd) TableName() string                     { return "permissions" }
func (c NewPermissionCmd) Params() any                           { return c.params }

func (c NewPermissionCmd) GetID(x mdb.Permissions) string {
	return x.PermissionID.String()
}

func (c NewPermissionCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Permissions, error) {
	sp := int64(0)
	if c.params.SystemProtected {
		sp = 1
	}
	queries := mdb.New(tx)
	return queries.CreatePermission(ctx, mdb.CreatePermissionParams{
		PermissionID:    types.NewPermissionID(),
		Label:           c.params.Label,
		SystemProtected: sp,
	})
}

func (d Database) NewPermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePermissionParams) NewPermissionCmd {
	return NewPermissionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdatePermissionCmd is an audited command for updating permissions in SQLite.
type UpdatePermissionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePermissionParams
	conn     *sql.DB
}

func (c UpdatePermissionCmd) Context() context.Context              { return c.ctx }
func (c UpdatePermissionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdatePermissionCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdatePermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c UpdatePermissionCmd) TableName() string                     { return "permissions" }
func (c UpdatePermissionCmd) Params() any                           { return c.params }
func (c UpdatePermissionCmd) GetID() string                         { return c.params.PermissionID.String() }

func (c UpdatePermissionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Permissions, error) {
	queries := mdb.New(tx)
	return queries.GetPermission(ctx, mdb.GetPermissionParams{PermissionID: c.params.PermissionID})
}

func (c UpdatePermissionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	sp := int64(0)
	if c.params.SystemProtected {
		sp = 1
	}
	queries := mdb.New(tx)
	return queries.UpdatePermission(ctx, mdb.UpdatePermissionParams{
		Label:           c.params.Label,
		SystemProtected: sp,
		PermissionID:    c.params.PermissionID,
	})
}

func (d Database) UpdatePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePermissionParams) UpdatePermissionCmd {
	return UpdatePermissionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeletePermissionCmd is an audited command for deleting permissions in SQLite.
type DeletePermissionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PermissionID
	conn     *sql.DB
}

func (c DeletePermissionCmd) Context() context.Context              { return c.ctx }
func (c DeletePermissionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeletePermissionCmd) Connection() *sql.DB                   { return c.conn }
func (c DeletePermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c DeletePermissionCmd) TableName() string                     { return "permissions" }
func (c DeletePermissionCmd) GetID() string                         { return c.id.String() }

func (c DeletePermissionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Permissions, error) {
	queries := mdb.New(tx)
	return queries.GetPermission(ctx, mdb.GetPermissionParams{PermissionID: c.id})
}

func (c DeletePermissionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeletePermission(ctx, mdb.DeletePermissionParams{PermissionID: c.id})
}

func (d Database) DeletePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PermissionID) DeletePermissionCmd {
	return DeletePermissionCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewPermissionCmdMysql is an audited command for creating permissions in MySQL.
type NewPermissionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePermissionParams
	conn     *sql.DB
}

func (c NewPermissionCmdMysql) Context() context.Context              { return c.ctx }
func (c NewPermissionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewPermissionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewPermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c NewPermissionCmdMysql) TableName() string                     { return "permissions" }
func (c NewPermissionCmdMysql) Params() any                           { return c.params }

func (c NewPermissionCmdMysql) GetID(x mdbm.Permissions) string {
	return x.PermissionID.String()
}

func (c NewPermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Permissions, error) {
	id := types.NewPermissionID()
	queries := mdbm.New(tx)
	err := queries.CreatePermission(ctx, mdbm.CreatePermissionParams{
		PermissionID:    id,
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
	})
	if err != nil {
		return mdbm.Permissions{}, fmt.Errorf("Failed to CreatePermission: %w", err)
	}
	return queries.GetPermission(ctx, mdbm.GetPermissionParams{PermissionID: id})
}

func (d MysqlDatabase) NewPermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePermissionParams) NewPermissionCmdMysql {
	return NewPermissionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdatePermissionCmdMysql is an audited command for updating permissions in MySQL.
type UpdatePermissionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePermissionParams
	conn     *sql.DB
}

func (c UpdatePermissionCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdatePermissionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdatePermissionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdatePermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c UpdatePermissionCmdMysql) TableName() string                     { return "permissions" }
func (c UpdatePermissionCmdMysql) Params() any                           { return c.params }
func (c UpdatePermissionCmdMysql) GetID() string                         { return c.params.PermissionID.String() }

func (c UpdatePermissionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Permissions, error) {
	queries := mdbm.New(tx)
	return queries.GetPermission(ctx, mdbm.GetPermissionParams{PermissionID: c.params.PermissionID})
}

func (c UpdatePermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdatePermission(ctx, mdbm.UpdatePermissionParams{
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
		PermissionID:    c.params.PermissionID,
	})
}

func (d MysqlDatabase) UpdatePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePermissionParams) UpdatePermissionCmdMysql {
	return UpdatePermissionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeletePermissionCmdMysql is an audited command for deleting permissions in MySQL.
type DeletePermissionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PermissionID
	conn     *sql.DB
}

func (c DeletePermissionCmdMysql) Context() context.Context              { return c.ctx }
func (c DeletePermissionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeletePermissionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeletePermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c DeletePermissionCmdMysql) TableName() string                     { return "permissions" }
func (c DeletePermissionCmdMysql) GetID() string                         { return c.id.String() }

func (c DeletePermissionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Permissions, error) {
	queries := mdbm.New(tx)
	return queries.GetPermission(ctx, mdbm.GetPermissionParams{PermissionID: c.id})
}

func (c DeletePermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeletePermission(ctx, mdbm.DeletePermissionParams{PermissionID: c.id})
}

func (d MysqlDatabase) DeletePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PermissionID) DeletePermissionCmdMysql {
	return DeletePermissionCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewPermissionCmdPsql is an audited command for creating permissions in PostgreSQL.
type NewPermissionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePermissionParams
	conn     *sql.DB
}

func (c NewPermissionCmdPsql) Context() context.Context              { return c.ctx }
func (c NewPermissionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewPermissionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewPermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c NewPermissionCmdPsql) TableName() string                     { return "permissions" }
func (c NewPermissionCmdPsql) Params() any                           { return c.params }

func (c NewPermissionCmdPsql) GetID(x mdbp.Permissions) string {
	return x.PermissionID.String()
}

func (c NewPermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Permissions, error) {
	queries := mdbp.New(tx)
	return queries.CreatePermission(ctx, mdbp.CreatePermissionParams{
		PermissionID:    types.NewPermissionID(),
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
	})
}

func (d PsqlDatabase) NewPermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePermissionParams) NewPermissionCmdPsql {
	return NewPermissionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdatePermissionCmdPsql is an audited command for updating permissions in PostgreSQL.
type UpdatePermissionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePermissionParams
	conn     *sql.DB
}

func (c UpdatePermissionCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdatePermissionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdatePermissionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdatePermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c UpdatePermissionCmdPsql) TableName() string                     { return "permissions" }
func (c UpdatePermissionCmdPsql) Params() any                           { return c.params }
func (c UpdatePermissionCmdPsql) GetID() string                         { return c.params.PermissionID.String() }

func (c UpdatePermissionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Permissions, error) {
	queries := mdbp.New(tx)
	return queries.GetPermission(ctx, mdbp.GetPermissionParams{PermissionID: c.params.PermissionID})
}

func (c UpdatePermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdatePermission(ctx, mdbp.UpdatePermissionParams{
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
		PermissionID:    c.params.PermissionID,
	})
}

func (d PsqlDatabase) UpdatePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePermissionParams) UpdatePermissionCmdPsql {
	return UpdatePermissionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeletePermissionCmdPsql is an audited command for deleting permissions in PostgreSQL.
type DeletePermissionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PermissionID
	conn     *sql.DB
}

func (c DeletePermissionCmdPsql) Context() context.Context              { return c.ctx }
func (c DeletePermissionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeletePermissionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeletePermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c DeletePermissionCmdPsql) TableName() string                     { return "permissions" }
func (c DeletePermissionCmdPsql) GetID() string                         { return c.id.String() }

func (c DeletePermissionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Permissions, error) {
	queries := mdbp.New(tx)
	return queries.GetPermission(ctx, mdbp.GetPermissionParams{PermissionID: c.id})
}

func (c DeletePermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeletePermission(ctx, mdbp.DeletePermissionParams{PermissionID: c.id})
}

func (d PsqlDatabase) DeletePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PermissionID) DeletePermissionCmdPsql {
	return DeletePermissionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
