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

// MapRole converts a sqlc-generated type to the wrapper type.
func (d Database) MapRole(a mdb.Roles) Roles {
	return Roles{
		RoleID:          a.RoleID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected != 0,
	}
}

// MapCreateRoleParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateRoleParams(a CreateRoleParams) mdb.CreateRoleParams {
	sp := int64(0)
	if a.SystemProtected {
		sp = 1
	}
	return mdb.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           a.Label,
		SystemProtected: sp,
	}
}

// MapUpdateRoleParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdateRoleParams(a UpdateRoleParams) mdb.UpdateRoleParams {
	sp := int64(0)
	if a.SystemProtected {
		sp = 1
	}
	return mdb.UpdateRoleParams{
		Label:           a.Label,
		SystemProtected: sp,
		RoleID:          a.RoleID,
	}
}

///////////////////////////////
// MYSQL MAPPERS
//////////////////////////////

// MapRole converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapRole(a mdbm.Roles) Roles {
	return Roles{
		RoleID:          a.RoleID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapCreateRoleParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbm.CreateRoleParams {
	return mdbm.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapUpdateRoleParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbm.UpdateRoleParams {
	return mdbm.UpdateRoleParams{
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
		RoleID:          a.RoleID,
	}
}

///////////////////////////////
// POSTGRESQL MAPPERS
//////////////////////////////

// MapRole converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapRole(a mdbp.Roles) Roles {
	return Roles{
		RoleID:          a.RoleID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapCreateRoleParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbp.CreateRoleParams {
	return mdbp.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapUpdateRoleParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbp.UpdateRoleParams {
	return mdbp.UpdateRoleParams{
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
		RoleID:          a.RoleID,
	}
}

///////////////////////////////
// SQLITE QUERIES
//////////////////////////////

// CreateRole creates a new audited role record.
func (d Database) CreateRole(ctx context.Context, ac audited.AuditContext, s CreateRoleParams) (*Roles, error) {
	cmd := d.NewRoleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	r := d.MapRole(result)
	return &r, nil
}

// DeleteRole deletes a role by ID with auditing.
func (d Database) DeleteRole(ctx context.Context, ac audited.AuditContext, id types.RoleID) error {
	cmd := d.DeleteRoleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateRole updates an existing role with auditing.
func (d Database) UpdateRole(ctx context.Context, ac audited.AuditContext, s UpdateRoleParams) (*string, error) {
	cmd := d.UpdateRoleCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL QUERIES
//////////////////////////////

// CreateRole creates a new audited role record.
func (d MysqlDatabase) CreateRole(ctx context.Context, ac audited.AuditContext, s CreateRoleParams) (*Roles, error) {
	cmd := d.NewRoleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	r := d.MapRole(result)
	return &r, nil
}

// DeleteRole deletes a role by ID with auditing.
func (d MysqlDatabase) DeleteRole(ctx context.Context, ac audited.AuditContext, id types.RoleID) error {
	cmd := d.DeleteRoleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateRole updates an existing role with auditing.
func (d MysqlDatabase) UpdateRole(ctx context.Context, ac audited.AuditContext, s UpdateRoleParams) (*string, error) {
	cmd := d.UpdateRoleCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES QUERIES
//////////////////////////////

// CreateRole creates a new audited role record.
func (d PsqlDatabase) CreateRole(ctx context.Context, ac audited.AuditContext, s CreateRoleParams) (*Roles, error) {
	cmd := d.NewRoleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	r := d.MapRole(result)
	return &r, nil
}

// DeleteRole deletes a role by ID with auditing.
func (d PsqlDatabase) DeleteRole(ctx context.Context, ac audited.AuditContext, id types.RoleID) error {
	cmd := d.DeleteRoleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// UpdateRole updates an existing role with auditing.
func (d PsqlDatabase) UpdateRole(ctx context.Context, ac audited.AuditContext, s UpdateRoleParams) (*string, error) {
	cmd := d.UpdateRoleCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewRoleCmd is an audited command for creating a role in SQLite.
type NewRoleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRoleParams
	conn     *sql.DB
}

func (c NewRoleCmd) Context() context.Context              { return c.ctx }
func (c NewRoleCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewRoleCmd) Connection() *sql.DB                   { return c.conn }
func (c NewRoleCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c NewRoleCmd) TableName() string                     { return "roles" }
func (c NewRoleCmd) Params() any                           { return c.params }

func (c NewRoleCmd) GetID(x mdb.Roles) string {
	return x.RoleID.String()
}

func (c NewRoleCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Roles, error) {
	sp := int64(0)
	if c.params.SystemProtected {
		sp = 1
	}
	queries := mdb.New(tx)
	return queries.CreateRole(ctx, mdb.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           c.params.Label,
		SystemProtected: sp,
	})
}

func (d Database) NewRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRoleParams) NewRoleCmd {
	return NewRoleCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateRoleCmd is an audited command for updating a role in SQLite.
type UpdateRoleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRoleParams
	conn     *sql.DB
}

func (c UpdateRoleCmd) Context() context.Context              { return c.ctx }
func (c UpdateRoleCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateRoleCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateRoleCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c UpdateRoleCmd) TableName() string                     { return "roles" }
func (c UpdateRoleCmd) Params() any                           { return c.params }
func (c UpdateRoleCmd) GetID() string                         { return c.params.RoleID.String() }

func (c UpdateRoleCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Roles, error) {
	queries := mdb.New(tx)
	return queries.GetRole(ctx, mdb.GetRoleParams{RoleID: c.params.RoleID})
}

func (c UpdateRoleCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	sp := int64(0)
	if c.params.SystemProtected {
		sp = 1
	}
	queries := mdb.New(tx)
	return queries.UpdateRole(ctx, mdb.UpdateRoleParams{
		Label:           c.params.Label,
		SystemProtected: sp,
		RoleID:          c.params.RoleID,
	})
}

func (d Database) UpdateRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRoleParams) UpdateRoleCmd {
	return UpdateRoleCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRoleCmd is an audited command for deleting a role in SQLite.
type DeleteRoleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RoleID
	conn     *sql.DB
}

func (c DeleteRoleCmd) Context() context.Context              { return c.ctx }
func (c DeleteRoleCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteRoleCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteRoleCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c DeleteRoleCmd) TableName() string                     { return "roles" }
func (c DeleteRoleCmd) GetID() string                         { return c.id.String() }

func (c DeleteRoleCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Roles, error) {
	queries := mdb.New(tx)
	return queries.GetRole(ctx, mdb.GetRoleParams{RoleID: c.id})
}

func (c DeleteRoleCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteRole(ctx, mdb.DeleteRoleParams{RoleID: c.id})
}

func (d Database) DeleteRoleCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RoleID) DeleteRoleCmd {
	return DeleteRoleCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewRoleCmdMysql is an audited command for creating a role in MySQL.
type NewRoleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRoleParams
	conn     *sql.DB
}

func (c NewRoleCmdMysql) Context() context.Context              { return c.ctx }
func (c NewRoleCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewRoleCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewRoleCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c NewRoleCmdMysql) TableName() string                     { return "roles" }
func (c NewRoleCmdMysql) Params() any                           { return c.params }

func (c NewRoleCmdMysql) GetID(x mdbm.Roles) string {
	return x.RoleID.String()
}

func (c NewRoleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Roles, error) {
	id := types.NewRoleID()
	queries := mdbm.New(tx)
	err := queries.CreateRole(ctx, mdbm.CreateRoleParams{
		RoleID:          id,
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
	})
	if err != nil {
		return mdbm.Roles{}, fmt.Errorf("Failed to CreateRole: %w", err)
	}
	return queries.GetRole(ctx, mdbm.GetRoleParams{RoleID: id})
}

func (d MysqlDatabase) NewRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRoleParams) NewRoleCmdMysql {
	return NewRoleCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateRoleCmdMysql is an audited command for updating a role in MySQL.
type UpdateRoleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRoleParams
	conn     *sql.DB
}

func (c UpdateRoleCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateRoleCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateRoleCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateRoleCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c UpdateRoleCmdMysql) TableName() string                     { return "roles" }
func (c UpdateRoleCmdMysql) Params() any                           { return c.params }
func (c UpdateRoleCmdMysql) GetID() string                         { return c.params.RoleID.String() }

func (c UpdateRoleCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Roles, error) {
	queries := mdbm.New(tx)
	return queries.GetRole(ctx, mdbm.GetRoleParams{RoleID: c.params.RoleID})
}

func (c UpdateRoleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateRole(ctx, mdbm.UpdateRoleParams{
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
		RoleID:          c.params.RoleID,
	})
}

func (d MysqlDatabase) UpdateRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRoleParams) UpdateRoleCmdMysql {
	return UpdateRoleCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRoleCmdMysql is an audited command for deleting a role in MySQL.
type DeleteRoleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RoleID
	conn     *sql.DB
}

func (c DeleteRoleCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteRoleCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteRoleCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteRoleCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c DeleteRoleCmdMysql) TableName() string                     { return "roles" }
func (c DeleteRoleCmdMysql) GetID() string                         { return c.id.String() }

func (c DeleteRoleCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Roles, error) {
	queries := mdbm.New(tx)
	return queries.GetRole(ctx, mdbm.GetRoleParams{RoleID: c.id})
}

func (c DeleteRoleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteRole(ctx, mdbm.DeleteRoleParams{RoleID: c.id})
}

func (d MysqlDatabase) DeleteRoleCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RoleID) DeleteRoleCmdMysql {
	return DeleteRoleCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewRoleCmdPsql is an audited command for creating a role in PostgreSQL.
type NewRoleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRoleParams
	conn     *sql.DB
}

func (c NewRoleCmdPsql) Context() context.Context              { return c.ctx }
func (c NewRoleCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewRoleCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewRoleCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c NewRoleCmdPsql) TableName() string                     { return "roles" }
func (c NewRoleCmdPsql) Params() any                           { return c.params }

func (c NewRoleCmdPsql) GetID(x mdbp.Roles) string {
	return x.RoleID.String()
}

func (c NewRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.CreateRole(ctx, mdbp.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
	})
}

func (d PsqlDatabase) NewRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRoleParams) NewRoleCmdPsql {
	return NewRoleCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateRoleCmdPsql is an audited command for updating a role in PostgreSQL.
type UpdateRoleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRoleParams
	conn     *sql.DB
}

func (c UpdateRoleCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateRoleCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateRoleCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateRoleCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c UpdateRoleCmdPsql) TableName() string                     { return "roles" }
func (c UpdateRoleCmdPsql) Params() any                           { return c.params }
func (c UpdateRoleCmdPsql) GetID() string                         { return c.params.RoleID.String() }

func (c UpdateRoleCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.GetRole(ctx, mdbp.GetRoleParams{RoleID: c.params.RoleID})
}

func (c UpdateRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateRole(ctx, mdbp.UpdateRoleParams{
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
		RoleID:          c.params.RoleID,
	})
}

func (d PsqlDatabase) UpdateRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRoleParams) UpdateRoleCmdPsql {
	return UpdateRoleCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRoleCmdPsql is an audited command for deleting a role in PostgreSQL.
type DeleteRoleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RoleID
	conn     *sql.DB
}

func (c DeleteRoleCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteRoleCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteRoleCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteRoleCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c DeleteRoleCmdPsql) TableName() string                     { return "roles" }
func (c DeleteRoleCmdPsql) GetID() string                         { return c.id.String() }

func (c DeleteRoleCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.GetRole(ctx, mdbp.GetRoleParams{RoleID: c.id})
}

func (c DeleteRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteRole(ctx, mdbp.DeleteRoleParams{RoleID: c.id})
}

func (d PsqlDatabase) DeleteRoleCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RoleID) DeleteRoleCmdPsql {
	return DeleteRoleCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
