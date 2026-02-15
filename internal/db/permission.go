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

// Permissions represents a permission entity with access control information.
type Permissions struct {
	PermissionID types.PermissionID `json:"permission_id"`
	TableID      string             `json:"table_id"`
	Mode         int64              `json:"mode"`
	Label        string             `json:"label"`
}

// CreatePermissionParams contains fields for creating a new permission.
type CreatePermissionParams struct {
	TableID string `json:"table_id"`
	Mode    int64  `json:"mode"`
	Label   string `json:"label"`
}

// UpdatePermissionParams contains fields for updating an existing permission.
type UpdatePermissionParams struct {
	TableID      string             `json:"table_id"`
	Mode         int64              `json:"mode"`
	Label        string             `json:"label"`
	PermissionID types.PermissionID `json:"permission_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringPermission converts a Permissions entity to string representation for table display.
func MapStringPermission(a Permissions) StringPermissions {
	return StringPermissions{
		PermissionID: a.PermissionID.String(),
		TableID:      a.TableID,
		Mode:         fmt.Sprintf("%d", a.Mode),
		Label:        a.Label,
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapPermission converts a sqlc-generated SQLite Permissions type to the wrapper type.
func (d Database) MapPermission(a mdb.Permissions) Permissions {
	return Permissions{
		PermissionID: a.PermissionID,
		TableID:      a.TableID,
		Mode:         a.Mode,
		Label:        a.Label,
	}
}

// MapCreatePermissionParams converts a wrapper CreatePermissionParams to sqlc-generated SQLite params.
func (d Database) MapCreatePermissionParams(a CreatePermissionParams) mdb.CreatePermissionParams {
	return mdb.CreatePermissionParams{
		PermissionID: types.NewPermissionID(),
		TableID:      a.TableID,
		Mode:         a.Mode,
		Label:        a.Label,
	}
}

// MapUpdatePermissionParams converts a wrapper UpdatePermissionParams to sqlc-generated SQLite params.
func (d Database) MapUpdatePermissionParams(a UpdatePermissionParams) mdb.UpdatePermissionParams {
	return mdb.UpdatePermissionParams{
		TableID:      a.TableID,
		Mode:         a.Mode,
		Label:        a.Label,
		PermissionID: a.PermissionID,
	}
}

// QUERIES

// CountPermissions returns the total count of permissions in the database.
func (d Database) CountPermissions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreatePermissionTable creates the permissions table in the database.
func (d Database) CreatePermissionTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

// CreatePermission inserts a new permission into the database with audit trail.
func (d Database) CreatePermission(ctx context.Context, ac audited.AuditContext, s CreatePermissionParams) (*Permissions, error) {
	cmd := d.NewPermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	r := d.MapPermission(result)
	return &r, nil
}

// DeletePermission removes a permission from the database with audit trail.
func (d Database) DeletePermission(ctx context.Context, ac audited.AuditContext, id types.PermissionID) error {
	cmd := d.DeletePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetPermission retrieves a single permission by ID from the database.
func (d Database) GetPermission(id types.PermissionID) (*Permissions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPermission(d.Context, mdb.GetPermissionParams{PermissionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapPermission(row)
	return &res, nil
}

// ListPermissions returns all permissions from the database.
func (d Database) ListPermissions() (*[]Permissions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Permissions: %v\n", err)
	}
	res := []Permissions{}
	for _, v := range rows {
		m := d.MapPermission(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdatePermission modifies an existing permission in the database with audit trail.
func (d Database) UpdatePermission(ctx context.Context, ac audited.AuditContext, s UpdatePermissionParams) (*string, error) {
	cmd := d.UpdatePermissionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapPermission converts a sqlc-generated MySQL Permissions type to the wrapper type.
func (d MysqlDatabase) MapPermission(a mdbm.Permissions) Permissions {
	return Permissions{
		PermissionID: a.PermissionID,
		TableID:      a.TableID,
		Mode:         int64(a.Mode),
		Label:        a.Label,
	}
}

// MapCreatePermissionParams converts a wrapper CreatePermissionParams to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreatePermissionParams(a CreatePermissionParams) mdbm.CreatePermissionParams {
	return mdbm.CreatePermissionParams{
		PermissionID: types.NewPermissionID(),
		TableID:      a.TableID,
		Mode:         int32(a.Mode),
		Label:        a.Label,
	}
}

// MapUpdatePermissionParams converts a wrapper UpdatePermissionParams to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdatePermissionParams(a UpdatePermissionParams) mdbm.UpdatePermissionParams {
	return mdbm.UpdatePermissionParams{
		TableID:      a.TableID,
		Mode:         int32(a.Mode),
		Label:        a.Label,
		PermissionID: a.PermissionID,
	}
}

// QUERIES

// CountPermissions returns the total count of permissions in the database.
func (d MysqlDatabase) CountPermissions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreatePermissionTable creates the permissions table in the database.
func (d MysqlDatabase) CreatePermissionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

// CreatePermission inserts a new permission into the database with audit trail.
func (d MysqlDatabase) CreatePermission(ctx context.Context, ac audited.AuditContext, s CreatePermissionParams) (*Permissions, error) {
	cmd := d.NewPermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	r := d.MapPermission(result)
	return &r, nil
}

// DeletePermission removes a permission from the database with audit trail.
func (d MysqlDatabase) DeletePermission(ctx context.Context, ac audited.AuditContext, id types.PermissionID) error {
	cmd := d.DeletePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetPermission retrieves a single permission by ID from the database.
func (d MysqlDatabase) GetPermission(id types.PermissionID) (*Permissions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPermission(d.Context, mdbm.GetPermissionParams{PermissionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapPermission(row)
	return &res, nil
}

// ListPermissions returns all permissions from the database.
func (d MysqlDatabase) ListPermissions() (*[]Permissions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Permissions: %v\n", err)
	}
	res := []Permissions{}
	for _, v := range rows {
		m := d.MapPermission(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdatePermission modifies an existing permission in the database with audit trail.
func (d MysqlDatabase) UpdatePermission(ctx context.Context, ac audited.AuditContext, s UpdatePermissionParams) (*string, error) {
	cmd := d.UpdatePermissionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapPermission converts a sqlc-generated PostgreSQL Permissions type to the wrapper type.
func (d PsqlDatabase) MapPermission(a mdbp.Permissions) Permissions {
	return Permissions{
		PermissionID: a.PermissionID,
		TableID:      a.TableID,
		Mode:         int64(a.Mode),
		Label:        a.Label,
	}
}

// MapCreatePermissionParams converts a wrapper CreatePermissionParams to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreatePermissionParams(a CreatePermissionParams) mdbp.CreatePermissionParams {
	return mdbp.CreatePermissionParams{
		PermissionID: types.NewPermissionID(),
		TableID:      a.TableID,
		Mode:         int32(a.Mode),
		Label:        a.Label,
	}
}

// MapUpdatePermissionParams converts a wrapper UpdatePermissionParams to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdatePermissionParams(a UpdatePermissionParams) mdbp.UpdatePermissionParams {
	return mdbp.UpdatePermissionParams{
		TableID:      a.TableID,
		Mode:         int32(a.Mode),
		Label:        a.Label,
		PermissionID: a.PermissionID,
	}
}

// QUERIES

// CountPermissions returns the total count of permissions in the database.
func (d PsqlDatabase) CountPermissions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreatePermissionTable creates the permissions table in the database.
func (d PsqlDatabase) CreatePermissionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreatePermissionTable(d.Context)
	return err
}

// CreatePermission inserts a new permission into the database with audit trail.
func (d PsqlDatabase) CreatePermission(ctx context.Context, ac audited.AuditContext, s CreatePermissionParams) (*Permissions, error) {
	cmd := d.NewPermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	r := d.MapPermission(result)
	return &r, nil
}

// DeletePermission removes a permission from the database with audit trail.
func (d PsqlDatabase) DeletePermission(ctx context.Context, ac audited.AuditContext, id types.PermissionID) error {
	cmd := d.DeletePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetPermission retrieves a single permission by ID from the database.
func (d PsqlDatabase) GetPermission(id types.PermissionID) (*Permissions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPermission(d.Context, mdbp.GetPermissionParams{PermissionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapPermission(row)
	return &res, nil
}

// ListPermissions returns all permissions from the database.
func (d PsqlDatabase) ListPermissions() (*[]Permissions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Permissions: %v\n", err)
	}
	res := []Permissions{}
	for _, v := range rows {
		m := d.MapPermission(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdatePermission modifies an existing permission in the database with audit trail.
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

// Context returns the context for this command.
func (c NewPermissionCmd) Context() context.Context         { return c.ctx }
// AuditContext returns the audit context for this command.
func (c NewPermissionCmd) AuditContext() audited.AuditContext { return c.auditCtx }
// Connection returns the database connection for this command.
func (c NewPermissionCmd) Connection() *sql.DB               { return c.conn }
// Recorder returns the change event recorder for this command.
func (c NewPermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
// TableName returns the table name for this command.
func (c NewPermissionCmd) TableName() string                 { return "permissions" }
// Params returns the parameters for this command.
func (c NewPermissionCmd) Params() any                       { return c.params }

// GetID returns the ID of the created permission.
func (c NewPermissionCmd) GetID(x mdb.Permissions) string {
	return x.PermissionID.String()
}

// Execute performs the create operation for this command.
func (c NewPermissionCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Permissions, error) {
	queries := mdb.New(tx)
	return queries.CreatePermission(ctx, mdb.CreatePermissionParams{
		PermissionID: types.NewPermissionID(),
		TableID:      c.params.TableID,
		Mode:         c.params.Mode,
		Label:        c.params.Label,
	})
}

// NewPermissionCmd creates a new permission create command for SQLite.
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

// Context returns the context for this command.
func (c UpdatePermissionCmd) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c UpdatePermissionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c UpdatePermissionCmd) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c UpdatePermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name for this command.
func (c UpdatePermissionCmd) TableName() string                 { return "permissions" }

// Params returns the parameters for this command.
func (c UpdatePermissionCmd) Params() any                       { return c.params }

// GetID returns the permission ID being updated.
func (c UpdatePermissionCmd) GetID() string                     { return c.params.PermissionID.String() }

// GetBefore retrieves the permission state before the update.
func (c UpdatePermissionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Permissions, error) {
	queries := mdb.New(tx)
	return queries.GetPermission(ctx, mdb.GetPermissionParams{PermissionID: c.params.PermissionID})
}

// Execute performs the update operation for this command.
func (c UpdatePermissionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdatePermission(ctx, mdb.UpdatePermissionParams{
		TableID:      c.params.TableID,
		Mode:         c.params.Mode,
		Label:        c.params.Label,
		PermissionID: c.params.PermissionID,
	})
}

// UpdatePermissionCmd creates a new permission update command for SQLite.
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

// Context returns the context for this command.
func (c DeletePermissionCmd) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c DeletePermissionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c DeletePermissionCmd) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c DeletePermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name for this command.
func (c DeletePermissionCmd) TableName() string                 { return "permissions" }

// GetID returns the permission ID being deleted.
func (c DeletePermissionCmd) GetID() string                     { return c.id.String() }

// GetBefore retrieves the permission state before the delete.
func (c DeletePermissionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Permissions, error) {
	queries := mdb.New(tx)
	return queries.GetPermission(ctx, mdb.GetPermissionParams{PermissionID: c.id})
}

// Execute performs the delete operation for this command.
func (c DeletePermissionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeletePermission(ctx, mdb.DeletePermissionParams{PermissionID: c.id})
}

// DeletePermissionCmd creates a new permission delete command for SQLite.
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

// Context returns the context for this command.
func (c NewPermissionCmdMysql) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c NewPermissionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c NewPermissionCmdMysql) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c NewPermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name for this command.
func (c NewPermissionCmdMysql) TableName() string                 { return "permissions" }

// Params returns the parameters for this command.
func (c NewPermissionCmdMysql) Params() any                       { return c.params }

// GetID returns the ID of the created permission.
func (c NewPermissionCmdMysql) GetID(x mdbm.Permissions) string {
	return x.PermissionID.String()
}

// Execute performs the create operation for this command.
func (c NewPermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Permissions, error) {
	id := types.NewPermissionID()
	queries := mdbm.New(tx)
	err := queries.CreatePermission(ctx, mdbm.CreatePermissionParams{
		PermissionID: id,
		TableID:      c.params.TableID,
		Mode:         int32(c.params.Mode),
		Label:        c.params.Label,
	})
	if err != nil {
		return mdbm.Permissions{}, fmt.Errorf("Failed to CreatePermission: %w", err)
	}
	return queries.GetPermission(ctx, mdbm.GetPermissionParams{PermissionID: id})
}

// NewPermissionCmd creates a new permission create command for MySQL.
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

// Context returns the context for this command.
func (c UpdatePermissionCmdMysql) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c UpdatePermissionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c UpdatePermissionCmdMysql) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c UpdatePermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name for this command.
func (c UpdatePermissionCmdMysql) TableName() string                 { return "permissions" }

// Params returns the parameters for this command.
func (c UpdatePermissionCmdMysql) Params() any                       { return c.params }

// GetID returns the permission ID being updated.
func (c UpdatePermissionCmdMysql) GetID() string                     { return c.params.PermissionID.String() }

// GetBefore retrieves the permission state before the update.
func (c UpdatePermissionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Permissions, error) {
	queries := mdbm.New(tx)
	return queries.GetPermission(ctx, mdbm.GetPermissionParams{PermissionID: c.params.PermissionID})
}

// Execute performs the update operation for this command.
func (c UpdatePermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdatePermission(ctx, mdbm.UpdatePermissionParams{
		TableID:      c.params.TableID,
		Mode:         int32(c.params.Mode),
		Label:        c.params.Label,
		PermissionID: c.params.PermissionID,
	})
}

// UpdatePermissionCmd creates a new permission update command for MySQL.
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

// Context returns the context for this command.
func (c DeletePermissionCmdMysql) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c DeletePermissionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c DeletePermissionCmdMysql) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c DeletePermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name for this command.
func (c DeletePermissionCmdMysql) TableName() string                 { return "permissions" }

// GetID returns the permission ID being deleted.
func (c DeletePermissionCmdMysql) GetID() string                     { return c.id.String() }

// GetBefore retrieves the permission state before the delete.
func (c DeletePermissionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Permissions, error) {
	queries := mdbm.New(tx)
	return queries.GetPermission(ctx, mdbm.GetPermissionParams{PermissionID: c.id})
}

// Execute performs the delete operation for this command.
func (c DeletePermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeletePermission(ctx, mdbm.DeletePermissionParams{PermissionID: c.id})
}

// DeletePermissionCmd creates a new permission delete command for MySQL.
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

// Context returns the context for this command.
func (c NewPermissionCmdPsql) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c NewPermissionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c NewPermissionCmdPsql) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c NewPermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name for this command.
func (c NewPermissionCmdPsql) TableName() string                 { return "permissions" }

// Params returns the parameters for this command.
func (c NewPermissionCmdPsql) Params() any                       { return c.params }

// GetID returns the ID of the created permission.
func (c NewPermissionCmdPsql) GetID(x mdbp.Permissions) string {
	return x.PermissionID.String()
}

// Execute performs the create operation for this command.
func (c NewPermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Permissions, error) {
	queries := mdbp.New(tx)
	return queries.CreatePermission(ctx, mdbp.CreatePermissionParams{
		PermissionID: types.NewPermissionID(),
		TableID:      c.params.TableID,
		Mode:         int32(c.params.Mode),
		Label:        c.params.Label,
	})
}

// NewPermissionCmd creates a new permission create command for PostgreSQL.
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

// Context returns the context for this command.
func (c UpdatePermissionCmdPsql) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c UpdatePermissionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c UpdatePermissionCmdPsql) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c UpdatePermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name for this command.
func (c UpdatePermissionCmdPsql) TableName() string                 { return "permissions" }

// Params returns the parameters for this command.
func (c UpdatePermissionCmdPsql) Params() any                       { return c.params }

// GetID returns the permission ID being updated.
func (c UpdatePermissionCmdPsql) GetID() string                     { return c.params.PermissionID.String() }

// GetBefore retrieves the permission state before the update.
func (c UpdatePermissionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Permissions, error) {
	queries := mdbp.New(tx)
	return queries.GetPermission(ctx, mdbp.GetPermissionParams{PermissionID: c.params.PermissionID})
}

// Execute performs the update operation for this command.
func (c UpdatePermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdatePermission(ctx, mdbp.UpdatePermissionParams{
		TableID:      c.params.TableID,
		Mode:         int32(c.params.Mode),
		Label:        c.params.Label,
		PermissionID: c.params.PermissionID,
	})
}

// UpdatePermissionCmd creates a new permission update command for PostgreSQL.
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

// Context returns the context for this command.
func (c DeletePermissionCmdPsql) Context() context.Context         { return c.ctx }

// AuditContext returns the audit context for this command.
func (c DeletePermissionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection for this command.
func (c DeletePermissionCmdPsql) Connection() *sql.DB               { return c.conn }

// Recorder returns the change event recorder for this command.
func (c DeletePermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name for this command.
func (c DeletePermissionCmdPsql) TableName() string                 { return "permissions" }

// GetID returns the permission ID being deleted.
func (c DeletePermissionCmdPsql) GetID() string                     { return c.id.String() }

// GetBefore retrieves the permission state before the delete.
func (c DeletePermissionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Permissions, error) {
	queries := mdbp.New(tx)
	return queries.GetPermission(ctx, mdbp.GetPermissionParams{PermissionID: c.id})
}

// Execute performs the delete operation for this command.
func (c DeletePermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeletePermission(ctx, mdbp.DeletePermissionParams{PermissionID: c.id})
}

// DeletePermissionCmd creates a new permission delete command for PostgreSQL.
func (d PsqlDatabase) DeletePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PermissionID) DeletePermissionCmdPsql {
	return DeletePermissionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
