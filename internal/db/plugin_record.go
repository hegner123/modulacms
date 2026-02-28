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
)

///////////////////////////////
// STRUCTS
//////////////////////////////

// Plugin represents a registered plugin in the persistent registry.
type Plugin struct {
	PluginID       types.PluginID     `json:"plugin_id"`
	Name           string             `json:"name"`
	Version        string             `json:"version"`
	Description    string             `json:"description"`
	Author         string             `json:"author"`
	Status         types.PluginStatus `json:"status"`
	Capabilities   types.JSONData     `json:"capabilities"`
	ApprovedAccess types.JSONData     `json:"approved_access"`
	ManifestHash   string             `json:"manifest_hash"`
	DateInstalled  types.Timestamp    `json:"date_installed"`
	DateModified   types.Timestamp    `json:"date_modified"`
}

// CreatePluginParams contains parameters for creating a plugin.
type CreatePluginParams struct {
	Name           string             `json:"name"`
	Version        string             `json:"version"`
	Description    string             `json:"description"`
	Author         string             `json:"author"`
	Status         types.PluginStatus `json:"status"`
	Capabilities   string             `json:"capabilities"`
	ApprovedAccess string             `json:"approved_access"`
	ManifestHash   string             `json:"manifest_hash"`
}

// UpdatePluginParams contains parameters for updating a plugin.
type UpdatePluginParams struct {
	PluginID       types.PluginID     `json:"plugin_id"`
	Version        string             `json:"version"`
	Description    string             `json:"description"`
	Author         string             `json:"author"`
	Status         types.PluginStatus `json:"status"`
	Capabilities   string             `json:"capabilities"`
	ApprovedAccess string             `json:"approved_access"`
	ManifestHash   string             `json:"manifest_hash"`
}

// parseJSONDataString converts a string to types.JSONData for passing to sqlc.
func parseJSONDataString(s string) types.JSONData {
	if s == "" {
		return types.NewJSONData(json.RawMessage("{}"))
	}
	return types.NewJSONData(json.RawMessage(s))
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapPlugin converts a sqlc-generated SQLite Plugins type to the wrapper type.
func (d Database) MapPlugin(a mdb.Plugins) Plugin {
	return Plugin{
		PluginID:       a.PluginID,
		Name:           a.Name,
		Version:        a.Version,
		Description:    a.Description,
		Author:         a.Author,
		Status:         a.Status,
		Capabilities:   a.Capabilities,
		ApprovedAccess: a.ApprovedAccess,
		ManifestHash:   a.ManifestHash,
		DateInstalled:  a.DateInstalled,
		DateModified:   a.DateModified,
	}
}

// QUERIES

// CountPlugins returns the total count of plugins.
func (d Database) CountPlugins() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountPlugins(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count plugins: %w", err)
	}
	return &c, nil
}

// CreatePluginTable creates the plugins table.
func (d Database) CreatePluginTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreatePluginsTable(d.Context)
}

// CreatePlugin inserts a new plugin and records an audit event.
func (d Database) CreatePlugin(ctx context.Context, ac audited.AuditContext, s CreatePluginParams) (*Plugin, error) {
	cmd := d.NewPluginCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}
	p := d.MapPlugin(result)
	return &p, nil
}

// DeletePlugin removes a plugin and records an audit event.
func (d Database) DeletePlugin(ctx context.Context, ac audited.AuditContext, id types.PluginID) error {
	cmd := d.DeletePluginCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetPlugin retrieves a plugin by ID.
func (d Database) GetPlugin(id types.PluginID) (*Plugin, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPlugin(d.Context, mdb.GetPluginParams{PluginID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	p := d.MapPlugin(row)
	return &p, nil
}

// GetPluginByName retrieves a plugin by its unique name.
func (d Database) GetPluginByName(name string) (*Plugin, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPluginByName(d.Context, mdb.GetPluginByNameParams{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin by name: %w", err)
	}
	p := d.MapPlugin(row)
	return &p, nil
}

// ListPlugins returns all plugins.
func (d Database) ListPlugins() (*[]Plugin, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPlugins(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}
	res := []Plugin{}
	for _, v := range rows {
		res = append(res, d.MapPlugin(v))
	}
	return &res, nil
}

// ListPluginsByStatus returns all plugins with a given status.
func (d Database) ListPluginsByStatus(status types.PluginStatus) (*[]Plugin, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPluginsByStatus(d.Context, mdb.ListPluginsByStatusParams{Status: status})
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by status: %w", err)
	}
	res := []Plugin{}
	for _, v := range rows {
		res = append(res, d.MapPlugin(v))
	}
	return &res, nil
}

// UpdatePlugin updates a plugin and records an audit event.
func (d Database) UpdatePlugin(ctx context.Context, ac audited.AuditContext, s UpdatePluginParams) error {
	cmd := d.UpdatePluginCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// UpdatePluginStatus updates a plugin's status directly (non-audited convenience method).
func (d Database) UpdatePluginStatus(ctx context.Context, ac audited.AuditContext, id types.PluginID, status types.PluginStatus) error {
	queries := mdb.New(d.Connection)
	return queries.UpdatePluginStatus(d.Context, mdb.UpdatePluginStatusParams{
		PluginID:     id,
		Status:       status,
		DateModified: types.TimestampNow(),
	})
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapPlugin converts a sqlc-generated MySQL Plugins type to the wrapper type.
func (d MysqlDatabase) MapPlugin(a mdbm.Plugins) Plugin {
	return Plugin{
		PluginID:       a.PluginID,
		Name:           a.Name,
		Version:        a.Version,
		Description:    a.Description,
		Author:         a.Author,
		Status:         a.Status,
		Capabilities:   a.Capabilities,
		ApprovedAccess: a.ApprovedAccess,
		ManifestHash:   a.ManifestHash,
		DateInstalled:  a.DateInstalled,
		DateModified:   a.DateModified,
	}
}

// QUERIES

// CountPlugins returns the total count of plugins.
func (d MysqlDatabase) CountPlugins() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountPlugins(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count plugins: %w", err)
	}
	return &c, nil
}

// CreatePluginTable creates the plugins table.
func (d MysqlDatabase) CreatePluginTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreatePluginsTable(d.Context)
}

// CreatePlugin inserts a new plugin and records an audit event.
// MySQL uses :exec (no RETURNING), so we exec then fetch by ID.
func (d MysqlDatabase) CreatePlugin(ctx context.Context, ac audited.AuditContext, s CreatePluginParams) (*Plugin, error) {
	cmd := d.NewPluginCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}
	p := d.MapPlugin(result)
	return &p, nil
}

// DeletePlugin removes a plugin and records an audit event.
func (d MysqlDatabase) DeletePlugin(ctx context.Context, ac audited.AuditContext, id types.PluginID) error {
	cmd := d.DeletePluginCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetPlugin retrieves a plugin by ID.
func (d MysqlDatabase) GetPlugin(id types.PluginID) (*Plugin, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPlugin(d.Context, mdbm.GetPluginParams{PluginID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	p := d.MapPlugin(row)
	return &p, nil
}

// GetPluginByName retrieves a plugin by its unique name.
func (d MysqlDatabase) GetPluginByName(name string) (*Plugin, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPluginByName(d.Context, mdbm.GetPluginByNameParams{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin by name: %w", err)
	}
	p := d.MapPlugin(row)
	return &p, nil
}

// ListPlugins returns all plugins.
func (d MysqlDatabase) ListPlugins() (*[]Plugin, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPlugins(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}
	res := []Plugin{}
	for _, v := range rows {
		res = append(res, d.MapPlugin(v))
	}
	return &res, nil
}

// ListPluginsByStatus returns all plugins with a given status.
func (d MysqlDatabase) ListPluginsByStatus(status types.PluginStatus) (*[]Plugin, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPluginsByStatus(d.Context, mdbm.ListPluginsByStatusParams{Status: status})
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by status: %w", err)
	}
	res := []Plugin{}
	for _, v := range rows {
		res = append(res, d.MapPlugin(v))
	}
	return &res, nil
}

// UpdatePlugin updates a plugin and records an audit event.
func (d MysqlDatabase) UpdatePlugin(ctx context.Context, ac audited.AuditContext, s UpdatePluginParams) error {
	cmd := d.UpdatePluginCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// UpdatePluginStatus updates a plugin's status directly (non-audited convenience method).
func (d MysqlDatabase) UpdatePluginStatus(ctx context.Context, ac audited.AuditContext, id types.PluginID, status types.PluginStatus) error {
	queries := mdbm.New(d.Connection)
	return queries.UpdatePluginStatus(d.Context, mdbm.UpdatePluginStatusParams{
		PluginID:     id,
		Status:       status,
		DateModified: types.TimestampNow(),
	})
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapPlugin converts a sqlc-generated PostgreSQL Plugins type to the wrapper type.
func (d PsqlDatabase) MapPlugin(a mdbp.Plugins) Plugin {
	return Plugin{
		PluginID:       a.PluginID,
		Name:           a.Name,
		Version:        a.Version,
		Description:    a.Description,
		Author:         a.Author,
		Status:         a.Status,
		Capabilities:   a.Capabilities,
		ApprovedAccess: a.ApprovedAccess,
		ManifestHash:   a.ManifestHash,
		DateInstalled:  a.DateInstalled,
		DateModified:   a.DateModified,
	}
}

// QUERIES

// CountPlugins returns the total count of plugins.
func (d PsqlDatabase) CountPlugins() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountPlugins(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count plugins: %w", err)
	}
	return &c, nil
}

// CreatePluginTable creates the plugins table and indexes.
func (d PsqlDatabase) CreatePluginTable() error {
	queries := mdbp.New(d.Connection)
	if err := queries.CreatePluginsTable(d.Context); err != nil {
		return err
	}
	if err := queries.CreatePluginsIndexName(d.Context); err != nil {
		return err
	}
	if err := queries.CreatePluginsIndexStatus(d.Context); err != nil {
		return err
	}
	return nil
}

// CreatePlugin inserts a new plugin and records an audit event.
func (d PsqlDatabase) CreatePlugin(ctx context.Context, ac audited.AuditContext, s CreatePluginParams) (*Plugin, error) {
	cmd := d.NewPluginCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}
	p := d.MapPlugin(result)
	return &p, nil
}

// DeletePlugin removes a plugin and records an audit event.
func (d PsqlDatabase) DeletePlugin(ctx context.Context, ac audited.AuditContext, id types.PluginID) error {
	cmd := d.DeletePluginCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetPlugin retrieves a plugin by ID.
func (d PsqlDatabase) GetPlugin(id types.PluginID) (*Plugin, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPlugin(d.Context, mdbp.GetPluginParams{PluginID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	p := d.MapPlugin(row)
	return &p, nil
}

// GetPluginByName retrieves a plugin by its unique name.
func (d PsqlDatabase) GetPluginByName(name string) (*Plugin, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPluginByName(d.Context, mdbp.GetPluginByNameParams{Name: name})
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin by name: %w", err)
	}
	p := d.MapPlugin(row)
	return &p, nil
}

// ListPlugins returns all plugins.
func (d PsqlDatabase) ListPlugins() (*[]Plugin, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPlugins(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}
	res := []Plugin{}
	for _, v := range rows {
		res = append(res, d.MapPlugin(v))
	}
	return &res, nil
}

// ListPluginsByStatus returns all plugins with a given status.
func (d PsqlDatabase) ListPluginsByStatus(status types.PluginStatus) (*[]Plugin, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPluginsByStatus(d.Context, mdbp.ListPluginsByStatusParams{Status: status})
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by status: %w", err)
	}
	res := []Plugin{}
	for _, v := range rows {
		res = append(res, d.MapPlugin(v))
	}
	return &res, nil
}

// UpdatePlugin updates a plugin and records an audit event.
func (d PsqlDatabase) UpdatePlugin(ctx context.Context, ac audited.AuditContext, s UpdatePluginParams) error {
	cmd := d.UpdatePluginCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// UpdatePluginStatus updates a plugin's status directly (non-audited convenience method).
func (d PsqlDatabase) UpdatePluginStatus(ctx context.Context, ac audited.AuditContext, id types.PluginID, status types.PluginStatus) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdatePluginStatus(d.Context, mdbp.UpdatePluginStatusParams{
		PluginID:     id,
		Status:       status,
		DateModified: types.TimestampNow(),
	})
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// ----- SQLite CREATE -----

// NewPluginCmd is an audited command for creating a plugin in SQLite.
type NewPluginCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePluginParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewPluginCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewPluginCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewPluginCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewPluginCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewPluginCmd) TableName() string { return "plugins" }

// Params returns the command parameters.
func (c NewPluginCmd) Params() any { return c.params }

// GetID returns the ID of a created plugin.
func (c NewPluginCmd) GetID(x mdb.Plugins) string {
	return x.PluginID.String()
}

// Execute creates a plugin in the database.
func (c NewPluginCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Plugins, error) {
	queries := mdb.New(tx)
	now := types.TimestampNow()
	return queries.CreatePlugin(ctx, mdb.CreatePluginParams{
		PluginID:       types.NewPluginID(),
		Name:           c.params.Name,
		Version:        c.params.Version,
		Description:    c.params.Description,
		Author:         c.params.Author,
		Status:         c.params.Status,
		Capabilities:   parseJSONDataString(c.params.Capabilities),
		ApprovedAccess: parseJSONDataString(c.params.ApprovedAccess),
		ManifestHash:   c.params.ManifestHash,
		DateInstalled:  now,
		DateModified:   now,
	})
}

// NewPluginCmd returns a new create command for a plugin.
func (d Database) NewPluginCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePluginParams) NewPluginCmd {
	return NewPluginCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdatePluginCmd is an audited command for updating a plugin in SQLite.
type UpdatePluginCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePluginParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdatePluginCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdatePluginCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdatePluginCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdatePluginCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdatePluginCmd) TableName() string { return "plugins" }

// Params returns the command parameters.
func (c UpdatePluginCmd) Params() any { return c.params }

// GetID returns the plugin ID for this command.
func (c UpdatePluginCmd) GetID() string { return string(c.params.PluginID) }

// GetBefore retrieves the plugin before the update.
func (c UpdatePluginCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Plugins, error) {
	queries := mdb.New(tx)
	return queries.GetPlugin(ctx, mdb.GetPluginParams{PluginID: c.params.PluginID})
}

// Execute updates the plugin in the database.
func (c UpdatePluginCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdatePlugin(ctx, mdb.UpdatePluginParams{
		Version:        c.params.Version,
		Description:    c.params.Description,
		Author:         c.params.Author,
		Status:         c.params.Status,
		Capabilities:   parseJSONDataString(c.params.Capabilities),
		ApprovedAccess: parseJSONDataString(c.params.ApprovedAccess),
		ManifestHash:   c.params.ManifestHash,
		DateModified:   types.TimestampNow(),
		PluginID:       c.params.PluginID,
	})
}

// UpdatePluginCmd creates a command for updating a plugin.
func (d Database) UpdatePluginCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePluginParams) UpdatePluginCmd {
	return UpdatePluginCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeletePluginCmd is an audited command for deleting a plugin in SQLite.
type DeletePluginCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PluginID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeletePluginCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeletePluginCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeletePluginCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeletePluginCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeletePluginCmd) TableName() string { return "plugins" }

// GetID returns the plugin ID being deleted.
func (c DeletePluginCmd) GetID() string { return c.id.String() }

// GetBefore retrieves the plugin before deletion.
func (c DeletePluginCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Plugins, error) {
	queries := mdb.New(tx)
	return queries.GetPlugin(ctx, mdb.GetPluginParams{PluginID: c.id})
}

// Execute deletes a plugin from the database.
func (c DeletePluginCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeletePlugin(ctx, mdb.DeletePluginParams{PluginID: c.id})
}

// DeletePluginCmd returns a new delete command for a plugin.
func (d Database) DeletePluginCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PluginID) DeletePluginCmd {
	return DeletePluginCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ===== MYSQL =====

// ----- MySQL CREATE -----

// NewPluginCmdMysql is an audited command for creating a plugin in MySQL.
type NewPluginCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePluginParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewPluginCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewPluginCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewPluginCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewPluginCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewPluginCmdMysql) TableName() string { return "plugins" }

// Params returns the command parameters.
func (c NewPluginCmdMysql) Params() any { return c.params }

// GetID returns the ID of a created plugin.
func (c NewPluginCmdMysql) GetID(x mdbm.Plugins) string {
	return x.PluginID.String()
}

// Execute creates a plugin in the database.
// MySQL uses :exec (no RETURNING), so we exec then fetch by ID.
func (c NewPluginCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Plugins, error) {
	id := types.NewPluginID()
	now := types.TimestampNow()
	queries := mdbm.New(tx)
	err := queries.CreatePlugin(ctx, mdbm.CreatePluginParams{
		PluginID:       id,
		Name:           c.params.Name,
		Version:        c.params.Version,
		Description:    c.params.Description,
		Author:         c.params.Author,
		Status:         c.params.Status,
		Capabilities:   parseJSONDataString(c.params.Capabilities),
		ApprovedAccess: parseJSONDataString(c.params.ApprovedAccess),
		ManifestHash:   c.params.ManifestHash,
		DateInstalled:  now,
		DateModified:   now,
	})
	if err != nil {
		return mdbm.Plugins{}, fmt.Errorf("failed to create plugin: %w", err)
	}
	return queries.GetPlugin(ctx, mdbm.GetPluginParams{PluginID: id})
}

// NewPluginCmd returns a new create command for a plugin.
func (d MysqlDatabase) NewPluginCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePluginParams) NewPluginCmdMysql {
	return NewPluginCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdatePluginCmdMysql is an audited command for updating a plugin in MySQL.
type UpdatePluginCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePluginParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdatePluginCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdatePluginCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdatePluginCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdatePluginCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdatePluginCmdMysql) TableName() string { return "plugins" }

// Params returns the command parameters.
func (c UpdatePluginCmdMysql) Params() any { return c.params }

// GetID returns the plugin ID for this command.
func (c UpdatePluginCmdMysql) GetID() string { return string(c.params.PluginID) }

// GetBefore retrieves the plugin before the update.
func (c UpdatePluginCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Plugins, error) {
	queries := mdbm.New(tx)
	return queries.GetPlugin(ctx, mdbm.GetPluginParams{PluginID: c.params.PluginID})
}

// Execute updates the plugin in the database.
func (c UpdatePluginCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdatePlugin(ctx, mdbm.UpdatePluginParams{
		Version:        c.params.Version,
		Description:    c.params.Description,
		Author:         c.params.Author,
		Status:         c.params.Status,
		Capabilities:   parseJSONDataString(c.params.Capabilities),
		ApprovedAccess: parseJSONDataString(c.params.ApprovedAccess),
		ManifestHash:   c.params.ManifestHash,
		DateModified:   types.TimestampNow(),
		PluginID:       c.params.PluginID,
	})
}

// UpdatePluginCmd creates a command for updating a plugin.
func (d MysqlDatabase) UpdatePluginCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePluginParams) UpdatePluginCmdMysql {
	return UpdatePluginCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeletePluginCmdMysql is an audited command for deleting a plugin in MySQL.
type DeletePluginCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PluginID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeletePluginCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeletePluginCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeletePluginCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeletePluginCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeletePluginCmdMysql) TableName() string { return "plugins" }

// GetID returns the plugin ID being deleted.
func (c DeletePluginCmdMysql) GetID() string { return c.id.String() }

// GetBefore retrieves the plugin before deletion.
func (c DeletePluginCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Plugins, error) {
	queries := mdbm.New(tx)
	return queries.GetPlugin(ctx, mdbm.GetPluginParams{PluginID: c.id})
}

// Execute deletes a plugin from the database.
func (c DeletePluginCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeletePlugin(ctx, mdbm.DeletePluginParams{PluginID: c.id})
}

// DeletePluginCmd returns a new delete command for a plugin.
func (d MysqlDatabase) DeletePluginCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PluginID) DeletePluginCmdMysql {
	return DeletePluginCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ===== POSTGRESQL =====

// ----- PostgreSQL CREATE -----

// NewPluginCmdPsql is an audited command for creating a plugin in PostgreSQL.
type NewPluginCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePluginParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c NewPluginCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewPluginCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewPluginCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewPluginCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c NewPluginCmdPsql) TableName() string { return "plugins" }

// Params returns the command parameters.
func (c NewPluginCmdPsql) Params() any { return c.params }

// GetID returns the ID of a created plugin.
func (c NewPluginCmdPsql) GetID(x mdbp.Plugins) string {
	return x.PluginID.String()
}

// Execute creates a plugin in the database.
func (c NewPluginCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Plugins, error) {
	queries := mdbp.New(tx)
	now := types.TimestampNow()
	return queries.CreatePlugin(ctx, mdbp.CreatePluginParams{
		PluginID:       types.NewPluginID(),
		Name:           c.params.Name,
		Version:        c.params.Version,
		Description:    c.params.Description,
		Author:         c.params.Author,
		Status:         c.params.Status,
		Capabilities:   parseJSONDataString(c.params.Capabilities),
		ApprovedAccess: parseJSONDataString(c.params.ApprovedAccess),
		ManifestHash:   c.params.ManifestHash,
		DateInstalled:  now,
		DateModified:   now,
	})
}

// NewPluginCmd returns a new create command for a plugin.
func (d PsqlDatabase) NewPluginCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePluginParams) NewPluginCmdPsql {
	return NewPluginCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdatePluginCmdPsql is an audited command for updating a plugin in PostgreSQL.
type UpdatePluginCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePluginParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c UpdatePluginCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdatePluginCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdatePluginCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdatePluginCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c UpdatePluginCmdPsql) TableName() string { return "plugins" }

// Params returns the command parameters.
func (c UpdatePluginCmdPsql) Params() any { return c.params }

// GetID returns the plugin ID for this command.
func (c UpdatePluginCmdPsql) GetID() string { return string(c.params.PluginID) }

// GetBefore retrieves the plugin before the update.
func (c UpdatePluginCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Plugins, error) {
	queries := mdbp.New(tx)
	return queries.GetPlugin(ctx, mdbp.GetPluginParams{PluginID: c.params.PluginID})
}

// Execute updates the plugin in the database.
func (c UpdatePluginCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdatePlugin(ctx, mdbp.UpdatePluginParams{
		Version:        c.params.Version,
		Description:    c.params.Description,
		Author:         c.params.Author,
		Status:         c.params.Status,
		Capabilities:   parseJSONDataString(c.params.Capabilities),
		ApprovedAccess: parseJSONDataString(c.params.ApprovedAccess),
		ManifestHash:   c.params.ManifestHash,
		DateModified:   types.TimestampNow(),
		PluginID:       c.params.PluginID,
	})
}

// UpdatePluginCmd creates a command for updating a plugin.
func (d PsqlDatabase) UpdatePluginCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePluginParams) UpdatePluginCmdPsql {
	return UpdatePluginCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeletePluginCmdPsql is an audited command for deleting a plugin in PostgreSQL.
type DeletePluginCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PluginID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command's context.
func (c DeletePluginCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeletePluginCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeletePluginCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeletePluginCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name.
func (c DeletePluginCmdPsql) TableName() string { return "plugins" }

// GetID returns the plugin ID being deleted.
func (c DeletePluginCmdPsql) GetID() string { return c.id.String() }

// GetBefore retrieves the plugin before deletion.
func (c DeletePluginCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Plugins, error) {
	queries := mdbp.New(tx)
	return queries.GetPlugin(ctx, mdbp.GetPluginParams{PluginID: c.id})
}

// Execute deletes a plugin from the database.
func (c DeletePluginCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeletePlugin(ctx, mdbp.DeletePluginParams{PluginID: c.id})
}

// DeletePluginCmd returns a new delete command for a plugin.
func (d PsqlDatabase) DeletePluginCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PluginID) DeletePluginCmdPsql {
	return DeletePluginCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
