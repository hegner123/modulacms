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

// Roles represents a role entity with permissions.
type Roles struct {
	RoleID          types.RoleID `json:"role_id"`
	Label           string       `json:"label"`
	SystemProtected bool         `json:"system_protected"`
}

// CreateRoleParams contains parameters for creating a role.
type CreateRoleParams struct {
	Label           string `json:"label"`
	SystemProtected bool   `json:"system_protected"`
}

// UpdateRoleParams contains parameters for updating a role.
type UpdateRoleParams struct {
	Label           string       `json:"label"`
	SystemProtected bool         `json:"system_protected"`
	RoleID          types.RoleID `json:"role_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapRole converts a sqlc-generated type to the wrapper type.
func (d Database) MapRole(a mdb.Roles) Roles {
	return Roles{
		RoleID:          a.RoleID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected != 0,
	}
}

// MapCreateRoleParams converts a sqlc-generated type to the wrapper type.
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

// MapUpdateRoleParams converts a sqlc-generated type to the wrapper type.
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

// QUERIES

// CountRoles returns the total count of roles.
func (d Database) CountRoles() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateRoleTable creates the roles table.
func (d Database) CreateRoleTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

// CreateRole inserts a new role and records an audit event.
func (d Database) CreateRole(ctx context.Context, ac audited.AuditContext, s CreateRoleParams) (*Roles, error) {
	cmd := d.NewRoleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	r := d.MapRole(result)
	return &r, nil
}

// DeleteRole removes a role and records an audit event.
func (d Database) DeleteRole(ctx context.Context, ac audited.AuditContext, id types.RoleID) error {
	cmd := d.DeleteRoleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetRole retrieves a role by ID.
func (d Database) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdb.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

// ListRoles returns all roles.
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

// UpdateRole modifies a role and records an audit event.
func (d Database) UpdateRole(ctx context.Context, ac audited.AuditContext, s UpdateRoleParams) (*string, error) {
	cmd := d.UpdateRoleCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapRole converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapRole(a mdbm.Roles) Roles {
	return Roles{
		RoleID:          a.RoleID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapCreateRoleParams converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbm.CreateRoleParams {
	return mdbm.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapUpdateRoleParams converts a sqlc-generated type to the wrapper type.
func (d MysqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbm.UpdateRoleParams {
	return mdbm.UpdateRoleParams{
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
		RoleID:          a.RoleID,
	}
}

// QUERIES

// CountRoles returns the total count of roles.
func (d MysqlDatabase) CountRoles() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateRoleTable creates the roles table.
func (d MysqlDatabase) CreateRoleTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

// CreateRole inserts a new role and records an audit event.
func (d MysqlDatabase) CreateRole(ctx context.Context, ac audited.AuditContext, s CreateRoleParams) (*Roles, error) {
	cmd := d.NewRoleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	r := d.MapRole(result)
	return &r, nil
}

// DeleteRole removes a role and records an audit event.
func (d MysqlDatabase) DeleteRole(ctx context.Context, ac audited.AuditContext, id types.RoleID) error {
	cmd := d.DeleteRoleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetRole retrieves a role by ID.
func (d MysqlDatabase) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdbm.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

// ListRoles returns all roles.
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

// UpdateRole modifies a role and records an audit event.
func (d MysqlDatabase) UpdateRole(ctx context.Context, ac audited.AuditContext, s UpdateRoleParams) (*string, error) {
	cmd := d.UpdateRoleCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapRole converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapRole(a mdbp.Roles) Roles {
	return Roles{
		RoleID:          a.RoleID,
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapCreateRoleParams converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbp.CreateRoleParams {
	return mdbp.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
	}
}

// MapUpdateRoleParams converts a sqlc-generated type to the wrapper type.
func (d PsqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbp.UpdateRoleParams {
	return mdbp.UpdateRoleParams{
		Label:           a.Label,
		SystemProtected: a.SystemProtected,
		RoleID:          a.RoleID,
	}
}

// QUERIES

// CountRoles returns the total count of roles.
func (d PsqlDatabase) CountRoles() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateRoleTable creates the roles table.
func (d PsqlDatabase) CreateRoleTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateRoleTable(d.Context)
	return err
}

// CreateRole inserts a new role and records an audit event.
func (d PsqlDatabase) CreateRole(ctx context.Context, ac audited.AuditContext, s CreateRoleParams) (*Roles, error) {
	cmd := d.NewRoleCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}
	r := d.MapRole(result)
	return &r, nil
}

// DeleteRole removes a role and records an audit event.
func (d PsqlDatabase) DeleteRole(ctx context.Context, ac audited.AuditContext, id types.RoleID) error {
	cmd := d.DeleteRoleCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetRole retrieves a role by ID.
func (d PsqlDatabase) GetRole(id types.RoleID) (*Roles, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRole(d.Context, mdbp.GetRoleParams{RoleID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapRole(row)
	return &res, nil
}

// ListRoles returns all roles.
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

// UpdateRole modifies a role and records an audit event.
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

// Context returns the command's context.
func (c NewRoleCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c NewRoleCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c NewRoleCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c NewRoleCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the table name.
func (c NewRoleCmd) TableName() string                           { return "roles" }

// Params returns the command parameters.
func (c NewRoleCmd) Params() any                                 { return c.params }

// GetID returns the ID of a role.
func (c NewRoleCmd) GetID(x mdb.Roles) string {
	return x.RoleID.String()
}

// Execute creates a role in the database.
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

// NewRoleCmd returns a new create command for a role.
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

// Context returns the command's context.
func (c UpdateRoleCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateRoleCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateRoleCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateRoleCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the table name.
func (c UpdateRoleCmd) TableName() string                           { return "roles" }

// Params returns the command parameters.
func (c UpdateRoleCmd) Params() any                                 { return c.params }

// GetID returns the role ID.
func (c UpdateRoleCmd) GetID() string                               { return c.params.RoleID.String() }

// GetBefore retrieves the role before modification.
func (c UpdateRoleCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Roles, error) {
	queries := mdb.New(tx)
	return queries.GetRole(ctx, mdb.GetRoleParams{RoleID: c.params.RoleID})
}

// Execute updates a role in the database.
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

// UpdateRoleCmd returns a new update command for a role.
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

// Context returns the command's context.
func (c DeleteRoleCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteRoleCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteRoleCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteRoleCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the table name.
func (c DeleteRoleCmd) TableName() string                           { return "roles" }

// GetID returns the role ID.
func (c DeleteRoleCmd) GetID() string                               { return c.id.String() }

// GetBefore retrieves the role before deletion.
func (c DeleteRoleCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Roles, error) {
	queries := mdb.New(tx)
	return queries.GetRole(ctx, mdb.GetRoleParams{RoleID: c.id})
}

// Execute deletes a role from the database.
func (c DeleteRoleCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteRole(ctx, mdb.DeleteRoleParams{RoleID: c.id})
}

// DeleteRoleCmd returns a new delete command for a role.
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

// Context returns the command's context.
func (c NewRoleCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c NewRoleCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c NewRoleCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c NewRoleCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the table name.
func (c NewRoleCmdMysql) TableName() string                           { return "roles" }

// Params returns the command parameters.
func (c NewRoleCmdMysql) Params() any                                 { return c.params }

// GetID returns the ID of a role.
func (c NewRoleCmdMysql) GetID(x mdbm.Roles) string {
	return x.RoleID.String()
}

// Execute creates a role in the database.
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

// NewRoleCmd returns a new create command for a role.
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

// Context returns the command's context.
func (c UpdateRoleCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateRoleCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateRoleCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateRoleCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the table name.
func (c UpdateRoleCmdMysql) TableName() string                           { return "roles" }

// Params returns the command parameters.
func (c UpdateRoleCmdMysql) Params() any                                 { return c.params }

// GetID returns the role ID.
func (c UpdateRoleCmdMysql) GetID() string                               { return c.params.RoleID.String() }

// GetBefore retrieves the role before modification.
func (c UpdateRoleCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Roles, error) {
	queries := mdbm.New(tx)
	return queries.GetRole(ctx, mdbm.GetRoleParams{RoleID: c.params.RoleID})
}

// Execute updates a role in the database.
func (c UpdateRoleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateRole(ctx, mdbm.UpdateRoleParams{
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
		RoleID:          c.params.RoleID,
	})
}

// UpdateRoleCmd returns a new update command for a role.
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

// Context returns the command's context.
func (c DeleteRoleCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteRoleCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteRoleCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteRoleCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the table name.
func (c DeleteRoleCmdMysql) TableName() string                           { return "roles" }

// GetID returns the role ID.
func (c DeleteRoleCmdMysql) GetID() string                               { return c.id.String() }

// GetBefore retrieves the role before deletion.
func (c DeleteRoleCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Roles, error) {
	queries := mdbm.New(tx)
	return queries.GetRole(ctx, mdbm.GetRoleParams{RoleID: c.id})
}

// Execute deletes a role from the database.
func (c DeleteRoleCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteRole(ctx, mdbm.DeleteRoleParams{RoleID: c.id})
}

// DeleteRoleCmd returns a new delete command for a role.
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

// Context returns the command's context.
func (c NewRoleCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c NewRoleCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c NewRoleCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c NewRoleCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the table name.
func (c NewRoleCmdPsql) TableName() string                           { return "roles" }

// Params returns the command parameters.
func (c NewRoleCmdPsql) Params() any                                 { return c.params }

// GetID returns the ID of a role.
func (c NewRoleCmdPsql) GetID(x mdbp.Roles) string {
	return x.RoleID.String()
}

// Execute creates a role in the database.
func (c NewRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.CreateRole(ctx, mdbp.CreateRoleParams{
		RoleID:          types.NewRoleID(),
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
	})
}

// NewRoleCmd returns a new create command for a role.
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

// Context returns the command's context.
func (c UpdateRoleCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateRoleCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateRoleCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateRoleCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the table name.
func (c UpdateRoleCmdPsql) TableName() string                           { return "roles" }

// Params returns the command parameters.
func (c UpdateRoleCmdPsql) Params() any                                 { return c.params }

// GetID returns the role ID.
func (c UpdateRoleCmdPsql) GetID() string                               { return c.params.RoleID.String() }

// GetBefore retrieves the role before modification.
func (c UpdateRoleCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.GetRole(ctx, mdbp.GetRoleParams{RoleID: c.params.RoleID})
}

// Execute updates a role in the database.
func (c UpdateRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateRole(ctx, mdbp.UpdateRoleParams{
		Label:           c.params.Label,
		SystemProtected: c.params.SystemProtected,
		RoleID:          c.params.RoleID,
	})
}

// UpdateRoleCmd returns a new update command for a role.
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

// Context returns the command's context.
func (c DeleteRoleCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteRoleCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteRoleCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteRoleCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the table name.
func (c DeleteRoleCmdPsql) TableName() string                           { return "roles" }

// GetID returns the role ID.
func (c DeleteRoleCmdPsql) GetID() string                               { return c.id.String() }

// GetBefore retrieves the role before deletion.
func (c DeleteRoleCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Roles, error) {
	queries := mdbp.New(tx)
	return queries.GetRole(ctx, mdbp.GetRoleParams{RoleID: c.id})
}

// Execute deletes a role from the database.
func (c DeleteRoleCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteRole(ctx, mdbp.DeleteRoleParams{RoleID: c.id})
}

// DeleteRoleCmd returns a new delete command for a role.
func (d PsqlDatabase) DeleteRoleCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RoleID) DeleteRoleCmdPsql {
	return DeleteRoleCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
