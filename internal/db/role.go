package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/sqlc-dev/pqtype"
)

///////////////////////////////
// STRUCTS
//////////////////////////////

type Roles struct {
	RoleID      types.RoleID `json:"role_id"`
	Label       string       `json:"label"`
	Permissions string       `json:"permissions"`
}

type CreateRoleParams struct {
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}

type UpdateRoleParams struct {
	Label       string       `json:"label"`
	Permissions string       `json:"permissions"`
	RoleID      types.RoleID `json:"role_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapRole(a mdb.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func (d Database) MapCreateRoleParams(a CreateRoleParams) mdb.CreateRoleParams {
	return mdb.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func (d Database) MapUpdateRoleParams(a UpdateRoleParams) mdb.UpdateRoleParams {
	return mdb.UpdateRoleParams{
		Label:       a.Label,
		Permissions: a.Permissions,
		RoleID:      a.RoleID,
	}
}

// QUERIES

func (d Database) CountRoles() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateRoleTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d Database) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRole: %v\n", err)
	}
	return d.MapRole(row)
}

func (d Database) DeleteRole(id types.RoleID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteRole(d.Context, mdb.DeleteRoleParams{RoleID: id})
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", id)
	}
	return nil
}

func (d Database) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdb.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

func (d Database) ListRoles() (*[]Roles, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRole(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update role, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapRole(a mdbm.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: a.Permissions.String,
	}
}

func (d MysqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbm.CreateRoleParams {
	return mdbm.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       a.Label,
		Permissions: StringToNullString(a.Permissions),
	}
}

func (d MysqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbm.UpdateRoleParams {
	return mdbm.UpdateRoleParams{
		Label:       a.Label,
		Permissions: StringToNullString(a.Permissions),
		RoleID:      a.RoleID,
	}
}

// QUERIES

func (d MysqlDatabase) CountRoles() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateRoleTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRole: %v\n", err)
	}
	row, err := queries.GetRole(d.Context, mdbm.GetRoleParams{RoleID: params.RoleID})
	if err != nil {
		fmt.Printf("Failed to get last inserted Role: %v\n", err)
	}
	return d.MapRole(row)
}

func (d MysqlDatabase) DeleteRole(id types.RoleID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteRole(d.Context, mdbm.DeleteRoleParams{RoleID: id})
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", id)
	}
	return nil
}

func (d MysqlDatabase) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdbm.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

func (d MysqlDatabase) ListRoles() (*[]Roles, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRole(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update role, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapRole(a mdbp.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: string(a.Permissions.RawMessage),
	}
}

func (d PsqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbp.CreateRoleParams {
	return mdbp.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       a.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(a.Permissions)},
	}
}

func (d PsqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbp.UpdateRoleParams {
	return mdbp.UpdateRoleParams{
		Label:       a.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(a.Permissions)},
		RoleID:      a.RoleID,
	}
}

// QUERIES

func (d PsqlDatabase) CountRoles() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateRoleTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRole: %v\n", err)
	}
	return d.MapRole(row)
}

func (d PsqlDatabase) DeleteRole(id types.RoleID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteRole(d.Context, mdbp.DeleteRoleParams{RoleID: id})
	if err != nil {
		return fmt.Errorf("failed to delete role: %v", id)
	}
	return nil
}

func (d PsqlDatabase) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdbp.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

func (d PsqlDatabase) ListRoles() (*[]Roles, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRole(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update role, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewRoleCmd implements audited.CreateCommand[mdb.Roles] for SQLite.
type NewRoleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRoleParams
	conn     *sql.DB
}

func (c NewRoleCmd) Context() context.Context                    { return c.ctx }
func (c NewRoleCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewRoleCmd) Connection() *sql.DB                         { return c.conn }
func (c NewRoleCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c NewRoleCmd) TableName() string                           { return "roles" }
func (c NewRoleCmd) Params() any                                 { return c.params }

func (c NewRoleCmd) GetID(x mdb.Roles) string {
	return x.RoleID.String()
}

func (c NewRoleCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Roles, error) {
	queries := mdb.New(tx)
	return queries.CreateRole(ctx, mdb.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       c.params.Label,
		Permissions: c.params.Permissions,
	})
}

func (d Database) NewRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRoleParams) NewRoleCmd {
	return NewRoleCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateRoleCmd implements audited.UpdateCommand[mdb.Roles] for SQLite.
type UpdateRoleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRoleParams
	conn     *sql.DB
}

func (c UpdateRoleCmd) Context() context.Context                    { return c.ctx }
func (c UpdateRoleCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateRoleCmd) Connection() *sql.DB                         { return c.conn }
func (c UpdateRoleCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c UpdateRoleCmd) TableName() string                           { return "roles" }
func (c UpdateRoleCmd) Params() any                                 { return c.params }
func (c UpdateRoleCmd) GetID() string                               { return c.params.RoleID.String() }

func (c UpdateRoleCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Roles, error) {
	queries := mdb.New(tx)
	return queries.GetRole(ctx, mdb.GetRoleParams{RoleID: c.params.RoleID})
}

func (c UpdateRoleCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateRole(ctx, mdb.UpdateRoleParams{
		Label:       c.params.Label,
		Permissions: c.params.Permissions,
		RoleID:      c.params.RoleID,
	})
}

func (d Database) UpdateRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRoleParams) UpdateRoleCmd {
	return UpdateRoleCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRoleCmd implements audited.DeleteCommand[mdb.Roles] for SQLite.
type DeleteRoleCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RoleID
	conn     *sql.DB
}

func (c DeleteRoleCmd) Context() context.Context                    { return c.ctx }
func (c DeleteRoleCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteRoleCmd) Connection() *sql.DB                         { return c.conn }
func (c DeleteRoleCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c DeleteRoleCmd) TableName() string                           { return "roles" }
func (c DeleteRoleCmd) GetID() string                               { return c.id.String() }

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

// NewRoleCmdMysql implements audited.CreateCommand[mdbm.Roles] for MySQL.
type NewRoleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRoleParams
	conn     *sql.DB
}

func (c NewRoleCmdMysql) Context() context.Context                    { return c.ctx }
func (c NewRoleCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewRoleCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c NewRoleCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c NewRoleCmdMysql) TableName() string                           { return "roles" }
func (c NewRoleCmdMysql) Params() any                                 { return c.params }

func (c NewRoleCmdMysql) GetID(x mdbm.Roles) string {
	return x.RoleID.String()
}

func (c NewRoleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Roles, error) {
	id := types.NewRoleID()
	queries := mdbm.New(tx)
	err := queries.CreateRole(ctx, mdbm.CreateRoleParams{
		RoleID:      id,
		Label:       c.params.Label,
		Permissions: StringToNullString(c.params.Permissions),
	})
	if err != nil {
		return mdbm.Roles{}, fmt.Errorf("Failed to CreateRole: %w", err)
	}
	return queries.GetRole(ctx, mdbm.GetRoleParams{RoleID: id})
}

func (d MysqlDatabase) NewRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRoleParams) NewRoleCmdMysql {
	return NewRoleCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateRoleCmdMysql implements audited.UpdateCommand[mdbm.Roles] for MySQL.
type UpdateRoleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRoleParams
	conn     *sql.DB
}

func (c UpdateRoleCmdMysql) Context() context.Context                    { return c.ctx }
func (c UpdateRoleCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateRoleCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c UpdateRoleCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c UpdateRoleCmdMysql) TableName() string                           { return "roles" }
func (c UpdateRoleCmdMysql) Params() any                                 { return c.params }
func (c UpdateRoleCmdMysql) GetID() string                               { return c.params.RoleID.String() }

func (c UpdateRoleCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Roles, error) {
	queries := mdbm.New(tx)
	return queries.GetRole(ctx, mdbm.GetRoleParams{RoleID: c.params.RoleID})
}

func (c UpdateRoleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateRole(ctx, mdbm.UpdateRoleParams{
		Label:       c.params.Label,
		Permissions: StringToNullString(c.params.Permissions),
		RoleID:      c.params.RoleID,
	})
}

func (d MysqlDatabase) UpdateRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRoleParams) UpdateRoleCmdMysql {
	return UpdateRoleCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRoleCmdMysql implements audited.DeleteCommand[mdbm.Roles] for MySQL.
type DeleteRoleCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RoleID
	conn     *sql.DB
}

func (c DeleteRoleCmdMysql) Context() context.Context                    { return c.ctx }
func (c DeleteRoleCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteRoleCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c DeleteRoleCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c DeleteRoleCmdMysql) TableName() string                           { return "roles" }
func (c DeleteRoleCmdMysql) GetID() string                               { return c.id.String() }

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

// NewRoleCmdPsql implements audited.CreateCommand[mdbp.Roles] for PostgreSQL.
type NewRoleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRoleParams
	conn     *sql.DB
}

func (c NewRoleCmdPsql) Context() context.Context                    { return c.ctx }
func (c NewRoleCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewRoleCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c NewRoleCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c NewRoleCmdPsql) TableName() string                           { return "roles" }
func (c NewRoleCmdPsql) Params() any                                 { return c.params }

func (c NewRoleCmdPsql) GetID(x mdbp.Roles) string {
	return x.RoleID.String()
}

func (c NewRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.CreateRole(ctx, mdbp.CreateRoleParams{
		RoleID:      types.NewRoleID(),
		Label:       c.params.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(c.params.Permissions)},
	})
}

func (d PsqlDatabase) NewRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRoleParams) NewRoleCmdPsql {
	return NewRoleCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateRoleCmdPsql implements audited.UpdateCommand[mdbp.Roles] for PostgreSQL.
type UpdateRoleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateRoleParams
	conn     *sql.DB
}

func (c UpdateRoleCmdPsql) Context() context.Context                    { return c.ctx }
func (c UpdateRoleCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateRoleCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c UpdateRoleCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c UpdateRoleCmdPsql) TableName() string                           { return "roles" }
func (c UpdateRoleCmdPsql) Params() any                                 { return c.params }
func (c UpdateRoleCmdPsql) GetID() string                               { return c.params.RoleID.String() }

func (c UpdateRoleCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.GetRole(ctx, mdbp.GetRoleParams{RoleID: c.params.RoleID})
}

func (c UpdateRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateRole(ctx, mdbp.UpdateRoleParams{
		Label:       c.params.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(c.params.Permissions)},
		RoleID:      c.params.RoleID,
	})
}

func (d PsqlDatabase) UpdateRoleCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateRoleParams) UpdateRoleCmdPsql {
	return UpdateRoleCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRoleCmdPsql implements audited.DeleteCommand[mdbp.Roles] for PostgreSQL.
type DeleteRoleCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RoleID
	conn     *sql.DB
}

func (c DeleteRoleCmdPsql) Context() context.Context                    { return c.ctx }
func (c DeleteRoleCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteRoleCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c DeleteRoleCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c DeleteRoleCmdPsql) TableName() string                           { return "roles" }
func (c DeleteRoleCmdPsql) GetID() string                               { return c.id.String() }

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
