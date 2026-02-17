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

// RolePermissions represents a junction between a role and a permission.
type RolePermissions struct {
	ID           types.RolePermissionID `json:"id"`
	RoleID       types.RoleID           `json:"role_id"`
	PermissionID types.PermissionID     `json:"permission_id"`
}

// CreateRolePermissionParams contains parameters for creating a role-permission association.
type CreateRolePermissionParams struct {
	RoleID       types.RoleID      `json:"role_id"`
	PermissionID types.PermissionID `json:"permission_id"`
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapRolePermission converts a sqlc-generated SQLite RolePermissions type to the wrapper type.
func (d Database) MapRolePermission(a mdb.RolePermissions) RolePermissions {
	return RolePermissions{
		ID:           a.ID,
		RoleID:       a.RoleID,
		PermissionID: a.PermissionID,
	}
}

// QUERIES

// CountRolePermissions returns the total count of role-permission associations.
func (d Database) CountRolePermissions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountRolePermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count role permissions: %w", err)
	}
	return &c, nil
}

// CreateRolePermissionsTable creates the role_permissions table.
func (d Database) CreateRolePermissionsTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateRolePermissionsTable(d.Context)
	return err
}

// CreateRolePermission inserts a new role-permission association and records an audit event.
func (d Database) CreateRolePermission(ctx context.Context, ac audited.AuditContext, s CreateRolePermissionParams) (*RolePermissions, error) {
	cmd := d.NewRolePermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role permission: %w", err)
	}
	r := d.MapRolePermission(result)
	return &r, nil
}

// DeleteRolePermission removes a role-permission association and records an audit event.
func (d Database) DeleteRolePermission(ctx context.Context, ac audited.AuditContext, id types.RolePermissionID) error {
	cmd := d.DeleteRolePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// DeleteRolePermissionsByRoleID removes all role-permission associations for a given role.
// This is a non-audited bulk delete used during cache invalidation.
func (d Database) DeleteRolePermissionsByRoleID(ctx context.Context, ac audited.AuditContext, roleID types.RoleID) error {
	queries := mdb.New(d.Connection)
	return queries.DeleteRolePermissionByRoleID(d.Context, mdb.DeleteRolePermissionByRoleIDParams{RoleID: roleID})
}

// GetRolePermission retrieves a role-permission association by ID.
func (d Database) GetRolePermission(id types.RolePermissionID) (*RolePermissions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRolePermission(d.Context, mdb.GetRolePermissionParams{ID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get role permission: %w", err)
	}
	res := d.MapRolePermission(row)
	return &res, nil
}

// ListRolePermissions returns all role-permission associations.
func (d Database) ListRolePermissions() (*[]RolePermissions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRolePermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListRolePermissionsByRoleID returns all role-permission associations for a given role.
func (d Database) ListRolePermissionsByRoleID(roleID types.RoleID) (*[]RolePermissions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRolePermissionByRoleID(d.Context, mdb.ListRolePermissionByRoleIDParams{RoleID: roleID})
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions by role id: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListRolePermissionsByPermissionID returns all role-permission associations for a given permission.
func (d Database) ListRolePermissionsByPermissionID(permID types.PermissionID) (*[]RolePermissions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRolePermissionByPermissionID(d.Context, mdb.ListRolePermissionByPermissionIDParams{PermissionID: permID})
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions by permission id: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListPermissionLabelsByRoleID returns the permission labels associated with a role via the JOIN query.
func (d Database) ListPermissionLabelsByRoleID(roleID types.RoleID) (*[]string, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPermissionLabelsByRoleID(d.Context, mdb.ListPermissionLabelsByRoleIDParams{RoleID: roleID})
	if err != nil {
		return nil, fmt.Errorf("failed to list permission labels by role id: %w", err)
	}
	return &rows, nil
}

// GetRoleByLabel retrieves a role by its label string.
func (d Database) GetRoleByLabel(label string) (*Roles, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRoleByLabel(d.Context, mdb.GetRoleByLabelParams{Label: label})
	if err != nil {
		return nil, fmt.Errorf("failed to get role by label: %w", err)
	}
	res := d.MapRole(row)
	return &res, nil
}

// GetPermissionByLabel retrieves a permission by its label string.
func (d Database) GetPermissionByLabel(label string) (*Permissions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPermissionByLabel(d.Context, mdb.GetPermissionByLabelParams{Label: label})
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by label: %w", err)
	}
	res := d.MapPermission(row)
	return &res, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapRolePermission converts a sqlc-generated MySQL RolePermissions type to the wrapper type.
func (d MysqlDatabase) MapRolePermission(a mdbm.RolePermissions) RolePermissions {
	return RolePermissions{
		ID:           a.ID,
		RoleID:       a.RoleID,
		PermissionID: a.PermissionID,
	}
}

// QUERIES

// CountRolePermissions returns the total count of role-permission associations.
func (d MysqlDatabase) CountRolePermissions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountRolePermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count role permissions: %w", err)
	}
	return &c, nil
}

// CreateRolePermissionsTable creates the role_permissions table.
func (d MysqlDatabase) CreateRolePermissionsTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateRolePermissionsTable(d.Context)
	return err
}

// CreateRolePermission inserts a new role-permission association and records an audit event.
func (d MysqlDatabase) CreateRolePermission(ctx context.Context, ac audited.AuditContext, s CreateRolePermissionParams) (*RolePermissions, error) {
	cmd := d.NewRolePermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role permission: %w", err)
	}
	r := d.MapRolePermission(result)
	return &r, nil
}

// DeleteRolePermission removes a role-permission association and records an audit event.
func (d MysqlDatabase) DeleteRolePermission(ctx context.Context, ac audited.AuditContext, id types.RolePermissionID) error {
	cmd := d.DeleteRolePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// DeleteRolePermissionsByRoleID removes all role-permission associations for a given role.
// This is a non-audited bulk delete used during cache invalidation.
func (d MysqlDatabase) DeleteRolePermissionsByRoleID(ctx context.Context, ac audited.AuditContext, roleID types.RoleID) error {
	queries := mdbm.New(d.Connection)
	return queries.DeleteRolePermissionByRoleID(d.Context, mdbm.DeleteRolePermissionByRoleIDParams{RoleID: roleID})
}

// GetRolePermission retrieves a role-permission association by ID.
func (d MysqlDatabase) GetRolePermission(id types.RolePermissionID) (*RolePermissions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRolePermission(d.Context, mdbm.GetRolePermissionParams{ID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get role permission: %w", err)
	}
	res := d.MapRolePermission(row)
	return &res, nil
}

// ListRolePermissions returns all role-permission associations.
func (d MysqlDatabase) ListRolePermissions() (*[]RolePermissions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRolePermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListRolePermissionsByRoleID returns all role-permission associations for a given role.
func (d MysqlDatabase) ListRolePermissionsByRoleID(roleID types.RoleID) (*[]RolePermissions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRolePermissionByRoleID(d.Context, mdbm.ListRolePermissionByRoleIDParams{RoleID: roleID})
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions by role id: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListRolePermissionsByPermissionID returns all role-permission associations for a given permission.
func (d MysqlDatabase) ListRolePermissionsByPermissionID(permID types.PermissionID) (*[]RolePermissions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListRolePermissionByPermissionID(d.Context, mdbm.ListRolePermissionByPermissionIDParams{PermissionID: permID})
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions by permission id: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListPermissionLabelsByRoleID returns the permission labels associated with a role via the JOIN query.
func (d MysqlDatabase) ListPermissionLabelsByRoleID(roleID types.RoleID) (*[]string, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPermissionLabelsByRoleID(d.Context, mdbm.ListPermissionLabelsByRoleIDParams{RoleID: roleID})
	if err != nil {
		return nil, fmt.Errorf("failed to list permission labels by role id: %w", err)
	}
	return &rows, nil
}

// GetRoleByLabel retrieves a role by its label string.
func (d MysqlDatabase) GetRoleByLabel(label string) (*Roles, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRoleByLabel(d.Context, mdbm.GetRoleByLabelParams{Label: label})
	if err != nil {
		return nil, fmt.Errorf("failed to get role by label: %w", err)
	}
	res := d.MapRole(row)
	return &res, nil
}

// GetPermissionByLabel retrieves a permission by its label string.
func (d MysqlDatabase) GetPermissionByLabel(label string) (*Permissions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPermissionByLabel(d.Context, mdbm.GetPermissionByLabelParams{Label: label})
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by label: %w", err)
	}
	res := d.MapPermission(row)
	return &res, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapRolePermission converts a sqlc-generated PostgreSQL RolePermissions type to the wrapper type.
func (d PsqlDatabase) MapRolePermission(a mdbp.RolePermissions) RolePermissions {
	return RolePermissions{
		ID:           a.ID,
		RoleID:       a.RoleID,
		PermissionID: a.PermissionID,
	}
}

// QUERIES

// CountRolePermissions returns the total count of role-permission associations.
func (d PsqlDatabase) CountRolePermissions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountRolePermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count role permissions: %w", err)
	}
	return &c, nil
}

// CreateRolePermissionsTable creates the role_permissions table and indexes.
func (d PsqlDatabase) CreateRolePermissionsTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateRolePermissionsTable(d.Context)
	if err != nil {
		return err
	}
	err = queries.CreateRolePermissionsIndexRole(d.Context)
	if err != nil {
		return err
	}
	err = queries.CreateRolePermissionsIndexPermission(d.Context)
	if err != nil {
		return err
	}
	return nil
}

// CreateRolePermission inserts a new role-permission association and records an audit event.
func (d PsqlDatabase) CreateRolePermission(ctx context.Context, ac audited.AuditContext, s CreateRolePermissionParams) (*RolePermissions, error) {
	cmd := d.NewRolePermissionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create role permission: %w", err)
	}
	r := d.MapRolePermission(result)
	return &r, nil
}

// DeleteRolePermission removes a role-permission association and records an audit event.
func (d PsqlDatabase) DeleteRolePermission(ctx context.Context, ac audited.AuditContext, id types.RolePermissionID) error {
	cmd := d.DeleteRolePermissionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// DeleteRolePermissionsByRoleID removes all role-permission associations for a given role.
// This is a non-audited bulk delete used during cache invalidation.
func (d PsqlDatabase) DeleteRolePermissionsByRoleID(ctx context.Context, ac audited.AuditContext, roleID types.RoleID) error {
	queries := mdbp.New(d.Connection)
	return queries.DeleteRolePermissionByRoleID(d.Context, mdbp.DeleteRolePermissionByRoleIDParams{RoleID: roleID})
}

// GetRolePermission retrieves a role-permission association by ID.
func (d PsqlDatabase) GetRolePermission(id types.RolePermissionID) (*RolePermissions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRolePermission(d.Context, mdbp.GetRolePermissionParams{ID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get role permission: %w", err)
	}
	res := d.MapRolePermission(row)
	return &res, nil
}

// ListRolePermissions returns all role-permission associations.
func (d PsqlDatabase) ListRolePermissions() (*[]RolePermissions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRolePermission(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListRolePermissionsByRoleID returns all role-permission associations for a given role.
func (d PsqlDatabase) ListRolePermissionsByRoleID(roleID types.RoleID) (*[]RolePermissions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRolePermissionByRoleID(d.Context, mdbp.ListRolePermissionByRoleIDParams{RoleID: roleID})
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions by role id: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListRolePermissionsByPermissionID returns all role-permission associations for a given permission.
func (d PsqlDatabase) ListRolePermissionsByPermissionID(permID types.PermissionID) (*[]RolePermissions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListRolePermissionByPermissionID(d.Context, mdbp.ListRolePermissionByPermissionIDParams{PermissionID: permID})
	if err != nil {
		return nil, fmt.Errorf("failed to list role permissions by permission id: %w", err)
	}
	res := []RolePermissions{}
	for _, v := range rows {
		res = append(res, d.MapRolePermission(v))
	}
	return &res, nil
}

// ListPermissionLabelsByRoleID returns the permission labels associated with a role via the JOIN query.
func (d PsqlDatabase) ListPermissionLabelsByRoleID(roleID types.RoleID) (*[]string, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPermissionLabelsByRoleID(d.Context, mdbp.ListPermissionLabelsByRoleIDParams{RoleID: roleID})
	if err != nil {
		return nil, fmt.Errorf("failed to list permission labels by role id: %w", err)
	}
	return &rows, nil
}

// GetRoleByLabel retrieves a role by its label string.
func (d PsqlDatabase) GetRoleByLabel(label string) (*Roles, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRoleByLabel(d.Context, mdbp.GetRoleByLabelParams{Label: label})
	if err != nil {
		return nil, fmt.Errorf("failed to get role by label: %w", err)
	}
	res := d.MapRole(row)
	return &res, nil
}

// GetPermissionByLabel retrieves a permission by its label string.
func (d PsqlDatabase) GetPermissionByLabel(label string) (*Permissions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPermissionByLabel(d.Context, mdbp.GetPermissionByLabelParams{Label: label})
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by label: %w", err)
	}
	res := d.MapPermission(row)
	return &res, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewRolePermissionCmd is an audited command for creating a role-permission association in SQLite.
type NewRolePermissionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRolePermissionParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c NewRolePermissionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewRolePermissionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewRolePermissionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewRolePermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c NewRolePermissionCmd) TableName() string { return "role_permissions" }

// Params returns the command parameters.
func (c NewRolePermissionCmd) Params() any { return c.params }

// GetID returns the ID of a created role-permission association.
func (c NewRolePermissionCmd) GetID(x mdb.RolePermissions) string {
	return x.ID.String()
}

// Execute creates a role-permission association in the database.
func (c NewRolePermissionCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.RolePermissions, error) {
	queries := mdb.New(tx)
	return queries.CreateRolePermission(ctx, mdb.CreateRolePermissionParams{
		ID:           types.NewRolePermissionID(),
		RoleID:       c.params.RoleID,
		PermissionID: c.params.PermissionID,
	})
}

// NewRolePermissionCmd returns a new create command for a role-permission association.
func (d Database) NewRolePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRolePermissionParams) NewRolePermissionCmd {
	return NewRolePermissionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRolePermissionCmd is an audited command for deleting a role-permission association in SQLite.
type DeleteRolePermissionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RolePermissionID
	conn     *sql.DB
}

// Context returns the command's context.
func (c DeleteRolePermissionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteRolePermissionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteRolePermissionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteRolePermissionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c DeleteRolePermissionCmd) TableName() string { return "role_permissions" }

// GetID returns the role-permission ID being deleted.
func (c DeleteRolePermissionCmd) GetID() string { return c.id.String() }

// GetBefore retrieves the role-permission association before deletion.
func (c DeleteRolePermissionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.RolePermissions, error) {
	queries := mdb.New(tx)
	return queries.GetRolePermission(ctx, mdb.GetRolePermissionParams{ID: c.id})
}

// Execute deletes a role-permission association from the database.
func (c DeleteRolePermissionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteRolePermission(ctx, mdb.DeleteRolePermissionParams{ID: c.id})
}

// DeleteRolePermissionCmd returns a new delete command for a role-permission association.
func (d Database) DeleteRolePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RolePermissionID) DeleteRolePermissionCmd {
	return DeleteRolePermissionCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewRolePermissionCmdMysql is an audited command for creating a role-permission association in MySQL.
type NewRolePermissionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRolePermissionParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c NewRolePermissionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewRolePermissionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewRolePermissionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewRolePermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c NewRolePermissionCmdMysql) TableName() string { return "role_permissions" }

// Params returns the command parameters.
func (c NewRolePermissionCmdMysql) Params() any { return c.params }

// GetID returns the ID of a created role-permission association.
func (c NewRolePermissionCmdMysql) GetID(x mdbm.RolePermissions) string {
	return x.ID.String()
}

// Execute creates a role-permission association in the database.
// MySQL uses :exec (no RETURNING), so we exec then fetch.
func (c NewRolePermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.RolePermissions, error) {
	id := types.NewRolePermissionID()
	queries := mdbm.New(tx)
	err := queries.CreateRolePermission(ctx, mdbm.CreateRolePermissionParams{
		ID:           id,
		RoleID:       c.params.RoleID,
		PermissionID: c.params.PermissionID,
	})
	if err != nil {
		return mdbm.RolePermissions{}, fmt.Errorf("failed to create role permission: %w", err)
	}
	return queries.GetRolePermission(ctx, mdbm.GetRolePermissionParams{ID: id})
}

// NewRolePermissionCmd returns a new create command for a role-permission association.
func (d MysqlDatabase) NewRolePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRolePermissionParams) NewRolePermissionCmdMysql {
	return NewRolePermissionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRolePermissionCmdMysql is an audited command for deleting a role-permission association in MySQL.
type DeleteRolePermissionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RolePermissionID
	conn     *sql.DB
}

// Context returns the command's context.
func (c DeleteRolePermissionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteRolePermissionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteRolePermissionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteRolePermissionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c DeleteRolePermissionCmdMysql) TableName() string { return "role_permissions" }

// GetID returns the role-permission ID being deleted.
func (c DeleteRolePermissionCmdMysql) GetID() string { return c.id.String() }

// GetBefore retrieves the role-permission association before deletion.
func (c DeleteRolePermissionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.RolePermissions, error) {
	queries := mdbm.New(tx)
	return queries.GetRolePermission(ctx, mdbm.GetRolePermissionParams{ID: c.id})
}

// Execute deletes a role-permission association from the database.
func (c DeleteRolePermissionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteRolePermission(ctx, mdbm.DeleteRolePermissionParams{ID: c.id})
}

// DeleteRolePermissionCmd returns a new delete command for a role-permission association.
func (d MysqlDatabase) DeleteRolePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RolePermissionID) DeleteRolePermissionCmdMysql {
	return DeleteRolePermissionCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewRolePermissionCmdPsql is an audited command for creating a role-permission association in PostgreSQL.
type NewRolePermissionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateRolePermissionParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c NewRolePermissionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewRolePermissionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewRolePermissionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewRolePermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c NewRolePermissionCmdPsql) TableName() string { return "role_permissions" }

// Params returns the command parameters.
func (c NewRolePermissionCmdPsql) Params() any { return c.params }

// GetID returns the ID of a created role-permission association.
func (c NewRolePermissionCmdPsql) GetID(x mdbp.RolePermissions) string {
	return x.ID.String()
}

// Execute creates a role-permission association in the database.
func (c NewRolePermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.RolePermissions, error) {
	queries := mdbp.New(tx)
	return queries.CreateRolePermission(ctx, mdbp.CreateRolePermissionParams{
		ID:           types.NewRolePermissionID(),
		RoleID:       c.params.RoleID,
		PermissionID: c.params.PermissionID,
	})
}

// NewRolePermissionCmd returns a new create command for a role-permission association.
func (d PsqlDatabase) NewRolePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateRolePermissionParams) NewRolePermissionCmdPsql {
	return NewRolePermissionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteRolePermissionCmdPsql is an audited command for deleting a role-permission association in PostgreSQL.
type DeleteRolePermissionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.RolePermissionID
	conn     *sql.DB
}

// Context returns the command's context.
func (c DeleteRolePermissionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteRolePermissionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteRolePermissionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteRolePermissionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c DeleteRolePermissionCmdPsql) TableName() string { return "role_permissions" }

// GetID returns the role-permission ID being deleted.
func (c DeleteRolePermissionCmdPsql) GetID() string { return c.id.String() }

// GetBefore retrieves the role-permission association before deletion.
func (c DeleteRolePermissionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.RolePermissions, error) {
	queries := mdbp.New(tx)
	return queries.GetRolePermission(ctx, mdbp.GetRolePermissionParams{ID: c.id})
}

// Execute deletes a role-permission association from the database.
func (c DeleteRolePermissionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteRolePermission(ctx, mdbp.DeleteRolePermissionParams{ID: c.id})
}

// DeleteRolePermissionCmd returns a new delete command for a role-permission association.
func (d PsqlDatabase) DeleteRolePermissionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.RolePermissionID) DeleteRolePermissionCmdPsql {
	return DeleteRolePermissionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
