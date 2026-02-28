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

// Pipeline represents a plugin pipeline registration in the persistent registry.
type Pipeline struct {
	PipelineID   types.PipelineID `json:"pipeline_id"`
	PluginID     types.PluginID   `json:"plugin_id"`
	TableName    string           `json:"table_name"`
	Operation    string           `json:"operation"`
	PluginName   string           `json:"plugin_name"`
	Handler      string           `json:"handler"`
	Priority     int              `json:"priority"`
	Enabled      bool             `json:"enabled"`
	Config       types.JSONData   `json:"config"`
	DateCreated  types.Timestamp  `json:"date_created"`
	DateModified types.Timestamp  `json:"date_modified"`
}

// CreatePipelineParams contains parameters for creating a pipeline.
type CreatePipelineParams struct {
	PluginID   types.PluginID `json:"plugin_id"`
	TableName  string         `json:"table_name"`
	Operation  string         `json:"operation"`
	PluginName string         `json:"plugin_name"`
	Handler    string         `json:"handler"`
	Priority   int            `json:"priority"`
	Enabled    bool           `json:"enabled"`
	Config     types.JSONData `json:"config"`
}

// UpdatePipelineParams contains parameters for updating a pipeline.
type UpdatePipelineParams struct {
	PipelineID types.PipelineID `json:"pipeline_id"`
	TableName  string           `json:"table_name"`
	Operation  string           `json:"operation"`
	Handler    string           `json:"handler"`
	Priority   int              `json:"priority"`
	Enabled    bool             `json:"enabled"`
	Config     types.JSONData   `json:"config"`
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapPipeline converts a sqlc-generated SQLite Pipelines type to the wrapper type.
func (d Database) MapPipeline(a mdb.Pipelines) Pipeline {
	return Pipeline{
		PipelineID:   a.PipelineID,
		PluginID:     a.PluginID,
		TableName:    a.TableName,
		Operation:    a.Operation,
		PluginName:   a.PluginName,
		Handler:      a.Handler,
		Priority:     int(a.Priority),
		Enabled:      a.Enabled != 0,
		Config:       a.Config,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// QUERIES

// CountPipelines returns the total count of pipelines.
func (d Database) CountPipelines() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count pipelines: %w", err)
	}
	return &c, nil
}

// CreatePipelineTable creates the pipelines table.
func (d Database) CreatePipelineTable() error {
	queries := mdb.New(d.Connection)
	return queries.CreatePipelinesTable(d.Context)
}

// CreatePipeline inserts a new pipeline and records an audit event.
func (d Database) CreatePipeline(ctx context.Context, ac audited.AuditContext, s CreatePipelineParams) (*Pipeline, error) {
	cmd := d.NewPipelineCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}
	p := d.MapPipeline(result)
	return &p, nil
}

// DeletePipeline removes a pipeline and records an audit event.
func (d Database) DeletePipeline(ctx context.Context, ac audited.AuditContext, id types.PipelineID) error {
	cmd := d.DeletePipelineCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// DeletePipelinesByPluginID removes all pipelines for a given plugin.
// This is a non-audited bulk delete used for cascade cleanup.
func (d Database) DeletePipelinesByPluginID(ctx context.Context, ac audited.AuditContext, pluginID types.PluginID) error {
	queries := mdb.New(d.Connection)
	return queries.DeletePipelinesByPluginID(d.Context, mdb.DeletePipelinesByPluginIDParams{PluginID: pluginID})
}

// GetPipeline retrieves a pipeline by ID.
func (d Database) GetPipeline(id types.PipelineID) (*Pipeline, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPipeline(d.Context, mdb.GetPipelineParams{PipelineID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}
	p := d.MapPipeline(row)
	return &p, nil
}

// ListPipelines returns all pipelines ordered by table, operation, priority.
func (d Database) ListPipelines() (*[]Pipeline, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByTable returns all pipelines for a given table name.
func (d Database) ListPipelinesByTable(tableName string) (*[]Pipeline, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPipelinesByTable(d.Context, mdb.ListPipelinesByTableParams{TableName: tableName})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by table: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByPluginID returns all pipelines for a given plugin.
func (d Database) ListPipelinesByPluginID(pluginID types.PluginID) (*[]Pipeline, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPipelinesByPluginID(d.Context, mdb.ListPipelinesByPluginIDParams{PluginID: pluginID})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by plugin id: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByTableOperation returns pipelines for a given table and operation.
func (d Database) ListPipelinesByTableOperation(tableName string, operation string) (*[]Pipeline, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListPipelinesByTableOperation(d.Context, mdb.ListPipelinesByTableOperationParams{
		TableName: tableName,
		Operation: operation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by table operation: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListEnabledPipelines returns all enabled pipelines.
func (d Database) ListEnabledPipelines() (*[]Pipeline, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListEnabledPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled pipelines: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// UpdatePipeline updates a pipeline's fields and records an audit event.
func (d Database) UpdatePipeline(ctx context.Context, ac audited.AuditContext, s UpdatePipelineParams) error {
	cmd := d.UpdatePipelineCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// UpdatePipelineEnabled updates a pipeline's enabled flag directly (non-audited toggle).
func (d Database) UpdatePipelineEnabled(ctx context.Context, ac audited.AuditContext, id types.PipelineID, enabled bool) error {
	queries := mdb.New(d.Connection)
	var enabledInt int64
	if enabled {
		enabledInt = 1
	}
	return queries.UpdatePipelineEnabled(d.Context, mdb.UpdatePipelineEnabledParams{
		Enabled:      enabledInt,
		DateModified: types.TimestampNow(),
		PipelineID:   id,
	})
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapPipeline converts a sqlc-generated MySQL Pipelines type to the wrapper type.
func (d MysqlDatabase) MapPipeline(a mdbm.Pipelines) Pipeline {
	return Pipeline{
		PipelineID:   a.PipelineID,
		PluginID:     a.PluginID,
		TableName:    a.TableName,
		Operation:    a.Operation,
		PluginName:   a.PluginName,
		Handler:      a.Handler,
		Priority:     int(a.Priority),
		Enabled:      a.Enabled != 0,
		Config:       a.Config,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// QUERIES

// CountPipelines returns the total count of pipelines.
func (d MysqlDatabase) CountPipelines() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count pipelines: %w", err)
	}
	return &c, nil
}

// CreatePipelineTable creates the pipelines table.
func (d MysqlDatabase) CreatePipelineTable() error {
	queries := mdbm.New(d.Connection)
	return queries.CreatePipelinesTable(d.Context)
}

// CreatePipeline inserts a new pipeline and records an audit event.
// MySQL uses :exec (no RETURNING), so we exec then fetch by ID.
func (d MysqlDatabase) CreatePipeline(ctx context.Context, ac audited.AuditContext, s CreatePipelineParams) (*Pipeline, error) {
	cmd := d.NewPipelineCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}
	p := d.MapPipeline(result)
	return &p, nil
}

// DeletePipeline removes a pipeline and records an audit event.
func (d MysqlDatabase) DeletePipeline(ctx context.Context, ac audited.AuditContext, id types.PipelineID) error {
	cmd := d.DeletePipelineCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// DeletePipelinesByPluginID removes all pipelines for a given plugin.
// This is a non-audited bulk delete used for cascade cleanup.
func (d MysqlDatabase) DeletePipelinesByPluginID(ctx context.Context, ac audited.AuditContext, pluginID types.PluginID) error {
	queries := mdbm.New(d.Connection)
	return queries.DeletePipelinesByPluginID(d.Context, mdbm.DeletePipelinesByPluginIDParams{PluginID: pluginID})
}

// GetPipeline retrieves a pipeline by ID.
func (d MysqlDatabase) GetPipeline(id types.PipelineID) (*Pipeline, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPipeline(d.Context, mdbm.GetPipelineParams{PipelineID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}
	p := d.MapPipeline(row)
	return &p, nil
}

// ListPipelines returns all pipelines ordered by table, operation, priority.
func (d MysqlDatabase) ListPipelines() (*[]Pipeline, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByTable returns all pipelines for a given table name.
func (d MysqlDatabase) ListPipelinesByTable(tableName string) (*[]Pipeline, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPipelinesByTable(d.Context, mdbm.ListPipelinesByTableParams{TableName: tableName})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by table: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByPluginID returns all pipelines for a given plugin.
func (d MysqlDatabase) ListPipelinesByPluginID(pluginID types.PluginID) (*[]Pipeline, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPipelinesByPluginID(d.Context, mdbm.ListPipelinesByPluginIDParams{PluginID: pluginID})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by plugin id: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByTableOperation returns pipelines for a given table and operation.
func (d MysqlDatabase) ListPipelinesByTableOperation(tableName string, operation string) (*[]Pipeline, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListPipelinesByTableOperation(d.Context, mdbm.ListPipelinesByTableOperationParams{
		TableName: tableName,
		Operation: operation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by table operation: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListEnabledPipelines returns all enabled pipelines.
func (d MysqlDatabase) ListEnabledPipelines() (*[]Pipeline, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListEnabledPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled pipelines: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// UpdatePipeline updates a pipeline's fields and records an audit event.
func (d MysqlDatabase) UpdatePipeline(ctx context.Context, ac audited.AuditContext, s UpdatePipelineParams) error {
	cmd := d.UpdatePipelineCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// UpdatePipelineEnabled updates a pipeline's enabled flag directly (non-audited toggle).
func (d MysqlDatabase) UpdatePipelineEnabled(ctx context.Context, ac audited.AuditContext, id types.PipelineID, enabled bool) error {
	queries := mdbm.New(d.Connection)
	var enabledInt int8
	if enabled {
		enabledInt = 1
	}
	return queries.UpdatePipelineEnabled(d.Context, mdbm.UpdatePipelineEnabledParams{
		Enabled:      enabledInt,
		DateModified: types.TimestampNow(),
		PipelineID:   id,
	})
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapPipeline converts a sqlc-generated PostgreSQL Pipelines type to the wrapper type.
func (d PsqlDatabase) MapPipeline(a mdbp.Pipelines) Pipeline {
	return Pipeline{
		PipelineID:   a.PipelineID,
		PluginID:     a.PluginID,
		TableName:    a.TableName,
		Operation:    a.Operation,
		PluginName:   a.PluginName,
		Handler:      a.Handler,
		Priority:     int(a.Priority),
		Enabled:      a.Enabled,
		Config:       a.Config,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// QUERIES

// CountPipelines returns the total count of pipelines.
func (d PsqlDatabase) CountPipelines() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to count pipelines: %w", err)
	}
	return &c, nil
}

// CreatePipelineTable creates the pipelines table and indexes.
func (d PsqlDatabase) CreatePipelineTable() error {
	queries := mdbp.New(d.Connection)
	if err := queries.CreatePipelinesTable(d.Context); err != nil {
		return err
	}
	if err := queries.CreatePipelinesIndexPlugin(d.Context); err != nil {
		return err
	}
	if err := queries.CreatePipelinesIndexTable(d.Context); err != nil {
		return err
	}
	if err := queries.CreatePipelinesIndexUnique(d.Context); err != nil {
		return err
	}
	return nil
}

// CreatePipeline inserts a new pipeline and records an audit event.
func (d PsqlDatabase) CreatePipeline(ctx context.Context, ac audited.AuditContext, s CreatePipelineParams) (*Pipeline, error) {
	cmd := d.NewPipelineCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline: %w", err)
	}
	p := d.MapPipeline(result)
	return &p, nil
}

// DeletePipeline removes a pipeline and records an audit event.
func (d PsqlDatabase) DeletePipeline(ctx context.Context, ac audited.AuditContext, id types.PipelineID) error {
	cmd := d.DeletePipelineCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// DeletePipelinesByPluginID removes all pipelines for a given plugin.
// This is a non-audited bulk delete used for cascade cleanup.
func (d PsqlDatabase) DeletePipelinesByPluginID(ctx context.Context, ac audited.AuditContext, pluginID types.PluginID) error {
	queries := mdbp.New(d.Connection)
	return queries.DeletePipelinesByPluginID(d.Context, mdbp.DeletePipelinesByPluginIDParams{PluginID: pluginID})
}

// GetPipeline retrieves a pipeline by ID.
func (d PsqlDatabase) GetPipeline(id types.PipelineID) (*Pipeline, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPipeline(d.Context, mdbp.GetPipelineParams{PipelineID: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline: %w", err)
	}
	p := d.MapPipeline(row)
	return &p, nil
}

// ListPipelines returns all pipelines ordered by table, operation, priority.
func (d PsqlDatabase) ListPipelines() (*[]Pipeline, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByTable returns all pipelines for a given table name.
func (d PsqlDatabase) ListPipelinesByTable(tableName string) (*[]Pipeline, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPipelinesByTable(d.Context, mdbp.ListPipelinesByTableParams{TableName: tableName})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by table: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByPluginID returns all pipelines for a given plugin.
func (d PsqlDatabase) ListPipelinesByPluginID(pluginID types.PluginID) (*[]Pipeline, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPipelinesByPluginID(d.Context, mdbp.ListPipelinesByPluginIDParams{PluginID: pluginID})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by plugin id: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListPipelinesByTableOperation returns pipelines for a given table and operation.
func (d PsqlDatabase) ListPipelinesByTableOperation(tableName string, operation string) (*[]Pipeline, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListPipelinesByTableOperation(d.Context, mdbp.ListPipelinesByTableOperationParams{
		TableName: tableName,
		Operation: operation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pipelines by table operation: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// ListEnabledPipelines returns all enabled pipelines.
func (d PsqlDatabase) ListEnabledPipelines() (*[]Pipeline, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListEnabledPipelines(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list enabled pipelines: %w", err)
	}
	res := []Pipeline{}
	for _, v := range rows {
		res = append(res, d.MapPipeline(v))
	}
	return &res, nil
}

// UpdatePipeline updates a pipeline's fields and records an audit event.
func (d PsqlDatabase) UpdatePipeline(ctx context.Context, ac audited.AuditContext, s UpdatePipelineParams) error {
	cmd := d.UpdatePipelineCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// UpdatePipelineEnabled updates a pipeline's enabled flag directly (non-audited toggle).
func (d PsqlDatabase) UpdatePipelineEnabled(ctx context.Context, ac audited.AuditContext, id types.PipelineID, enabled bool) error {
	queries := mdbp.New(d.Connection)
	return queries.UpdatePipelineEnabled(d.Context, mdbp.UpdatePipelineEnabledParams{
		PipelineID:   id,
		Enabled:      enabled,
		DateModified: types.TimestampNow(),
	})
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewPipelineCmd is an audited command for creating a pipeline in SQLite.
type NewPipelineCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePipelineParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c NewPipelineCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewPipelineCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewPipelineCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewPipelineCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c NewPipelineCmd) TableName() string { return "pipelines" }

// Params returns the command parameters.
func (c NewPipelineCmd) Params() any { return c.params }

// GetID returns the ID of a created pipeline.
func (c NewPipelineCmd) GetID(x mdb.Pipelines) string {
	return x.PipelineID.String()
}

// Execute creates a pipeline in the database.
func (c NewPipelineCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Pipelines, error) {
	queries := mdb.New(tx)
	now := types.TimestampNow()
	var enabledInt int64
	if c.params.Enabled {
		enabledInt = 1
	}
	return queries.CreatePipeline(ctx, mdb.CreatePipelineParams{
		PipelineID:   types.NewPipelineID(),
		PluginID:     c.params.PluginID,
		TableName:    c.params.TableName,
		Operation:    c.params.Operation,
		PluginName:   c.params.PluginName,
		Handler:      c.params.Handler,
		Priority:     int64(c.params.Priority),
		Enabled:      enabledInt,
		Config:       c.params.Config,
		DateCreated:  now,
		DateModified: now,
	})
}

// NewPipelineCmd returns a new create command for a pipeline.
func (d Database) NewPipelineCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePipelineParams) NewPipelineCmd {
	return NewPipelineCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdatePipelineCmd is an audited command for updating a pipeline in SQLite.
type UpdatePipelineCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePipelineParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c UpdatePipelineCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdatePipelineCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdatePipelineCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdatePipelineCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c UpdatePipelineCmd) TableName() string { return "pipelines" }

// Params returns the command parameters.
func (c UpdatePipelineCmd) Params() any { return c.params }

// GetID returns the pipeline ID being updated.
func (c UpdatePipelineCmd) GetID() string { return c.params.PipelineID.String() }

// GetBefore retrieves the pipeline before the update.
func (c UpdatePipelineCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Pipelines, error) {
	queries := mdb.New(tx)
	return queries.GetPipeline(ctx, mdb.GetPipelineParams{PipelineID: c.params.PipelineID})
}

// Execute updates the pipeline in the database.
func (c UpdatePipelineCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	var enabledInt int64
	if c.params.Enabled {
		enabledInt = 1
	}
	return queries.UpdatePipeline(ctx, mdb.UpdatePipelineParams{
		TableName:    c.params.TableName,
		Operation:    c.params.Operation,
		Handler:      c.params.Handler,
		Priority:     int64(c.params.Priority),
		Enabled:      enabledInt,
		Config:       c.params.Config,
		DateModified: types.TimestampNow(),
		PipelineID:   c.params.PipelineID,
	})
}

// UpdatePipelineCmd returns a new update command for a pipeline.
func (d Database) UpdatePipelineCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePipelineParams) UpdatePipelineCmd {
	return UpdatePipelineCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeletePipelineCmd is an audited command for deleting a pipeline in SQLite.
type DeletePipelineCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PipelineID
	conn     *sql.DB
}

// Context returns the command's context.
func (c DeletePipelineCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeletePipelineCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeletePipelineCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeletePipelineCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c DeletePipelineCmd) TableName() string { return "pipelines" }

// GetID returns the pipeline ID being deleted.
func (c DeletePipelineCmd) GetID() string { return c.id.String() }

// GetBefore retrieves the pipeline before deletion.
func (c DeletePipelineCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Pipelines, error) {
	queries := mdb.New(tx)
	return queries.GetPipeline(ctx, mdb.GetPipelineParams{PipelineID: c.id})
}

// Execute deletes a pipeline from the database.
func (c DeletePipelineCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeletePipeline(ctx, mdb.DeletePipelineParams{PipelineID: c.id})
}

// DeletePipelineCmd returns a new delete command for a pipeline.
func (d Database) DeletePipelineCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PipelineID) DeletePipelineCmd {
	return DeletePipelineCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewPipelineCmdMysql is an audited command for creating a pipeline in MySQL.
type NewPipelineCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePipelineParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c NewPipelineCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewPipelineCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewPipelineCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewPipelineCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c NewPipelineCmdMysql) TableName() string { return "pipelines" }

// Params returns the command parameters.
func (c NewPipelineCmdMysql) Params() any { return c.params }

// GetID returns the ID of a created pipeline.
func (c NewPipelineCmdMysql) GetID(x mdbm.Pipelines) string {
	return x.PipelineID.String()
}

// Execute creates a pipeline in the database.
// MySQL uses :exec (no RETURNING), so we exec then fetch by ID.
func (c NewPipelineCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Pipelines, error) {
	id := types.NewPipelineID()
	now := types.TimestampNow()
	var enabledInt int8
	if c.params.Enabled {
		enabledInt = 1
	}
	queries := mdbm.New(tx)
	err := queries.CreatePipeline(ctx, mdbm.CreatePipelineParams{
		PipelineID:   id,
		PluginID:     c.params.PluginID,
		TableName:    c.params.TableName,
		Operation:    c.params.Operation,
		PluginName:   c.params.PluginName,
		Handler:      c.params.Handler,
		Priority:     int32(c.params.Priority),
		Enabled:      enabledInt,
		Config:       c.params.Config,
		DateCreated:  now,
		DateModified: now,
	})
	if err != nil {
		return mdbm.Pipelines{}, fmt.Errorf("failed to create pipeline: %w", err)
	}
	return queries.GetPipeline(ctx, mdbm.GetPipelineParams{PipelineID: id})
}

// NewPipelineCmd returns a new create command for a pipeline.
func (d MysqlDatabase) NewPipelineCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePipelineParams) NewPipelineCmdMysql {
	return NewPipelineCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdatePipelineCmdMysql is an audited command for updating a pipeline in MySQL.
type UpdatePipelineCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePipelineParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c UpdatePipelineCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdatePipelineCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdatePipelineCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdatePipelineCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c UpdatePipelineCmdMysql) TableName() string { return "pipelines" }

// Params returns the command parameters.
func (c UpdatePipelineCmdMysql) Params() any { return c.params }

// GetID returns the pipeline ID being updated.
func (c UpdatePipelineCmdMysql) GetID() string { return c.params.PipelineID.String() }

// GetBefore retrieves the pipeline before the update.
func (c UpdatePipelineCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Pipelines, error) {
	queries := mdbm.New(tx)
	return queries.GetPipeline(ctx, mdbm.GetPipelineParams{PipelineID: c.params.PipelineID})
}

// Execute updates the pipeline in the database.
func (c UpdatePipelineCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	var enabledInt int8
	if c.params.Enabled {
		enabledInt = 1
	}
	return queries.UpdatePipeline(ctx, mdbm.UpdatePipelineParams{
		TableName:    c.params.TableName,
		Operation:    c.params.Operation,
		Handler:      c.params.Handler,
		Priority:     int32(c.params.Priority),
		Enabled:      enabledInt,
		Config:       c.params.Config,
		DateModified: types.TimestampNow(),
		PipelineID:   c.params.PipelineID,
	})
}

// UpdatePipelineCmd returns a new update command for a pipeline.
func (d MysqlDatabase) UpdatePipelineCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePipelineParams) UpdatePipelineCmdMysql {
	return UpdatePipelineCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeletePipelineCmdMysql is an audited command for deleting a pipeline in MySQL.
type DeletePipelineCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PipelineID
	conn     *sql.DB
}

// Context returns the command's context.
func (c DeletePipelineCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeletePipelineCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeletePipelineCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeletePipelineCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c DeletePipelineCmdMysql) TableName() string { return "pipelines" }

// GetID returns the pipeline ID being deleted.
func (c DeletePipelineCmdMysql) GetID() string { return c.id.String() }

// GetBefore retrieves the pipeline before deletion.
func (c DeletePipelineCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Pipelines, error) {
	queries := mdbm.New(tx)
	return queries.GetPipeline(ctx, mdbm.GetPipelineParams{PipelineID: c.id})
}

// Execute deletes a pipeline from the database.
func (c DeletePipelineCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeletePipeline(ctx, mdbm.DeletePipelineParams{PipelineID: c.id})
}

// DeletePipelineCmd returns a new delete command for a pipeline.
func (d MysqlDatabase) DeletePipelineCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PipelineID) DeletePipelineCmdMysql {
	return DeletePipelineCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewPipelineCmdPsql is an audited command for creating a pipeline in PostgreSQL.
type NewPipelineCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreatePipelineParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c NewPipelineCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewPipelineCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewPipelineCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewPipelineCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c NewPipelineCmdPsql) TableName() string { return "pipelines" }

// Params returns the command parameters.
func (c NewPipelineCmdPsql) Params() any { return c.params }

// GetID returns the ID of a created pipeline.
func (c NewPipelineCmdPsql) GetID(x mdbp.Pipelines) string {
	return x.PipelineID.String()
}

// Execute creates a pipeline in the database.
func (c NewPipelineCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Pipelines, error) {
	queries := mdbp.New(tx)
	now := types.TimestampNow()
	return queries.CreatePipeline(ctx, mdbp.CreatePipelineParams{
		PipelineID:   types.NewPipelineID(),
		PluginID:     c.params.PluginID,
		TableName:    c.params.TableName,
		Operation:    c.params.Operation,
		PluginName:   c.params.PluginName,
		Handler:      c.params.Handler,
		Priority:     int32(c.params.Priority),
		Enabled:      c.params.Enabled,
		Config:       c.params.Config,
		DateCreated:  now,
		DateModified: now,
	})
}

// NewPipelineCmd returns a new create command for a pipeline.
func (d PsqlDatabase) NewPipelineCmd(ctx context.Context, auditCtx audited.AuditContext, params CreatePipelineParams) NewPipelineCmdPsql {
	return NewPipelineCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdatePipelineCmdPsql is an audited command for updating a pipeline in PostgreSQL.
type UpdatePipelineCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdatePipelineParams
	conn     *sql.DB
}

// Context returns the command's context.
func (c UpdatePipelineCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdatePipelineCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdatePipelineCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdatePipelineCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c UpdatePipelineCmdPsql) TableName() string { return "pipelines" }

// Params returns the command parameters.
func (c UpdatePipelineCmdPsql) Params() any { return c.params }

// GetID returns the pipeline ID being updated.
func (c UpdatePipelineCmdPsql) GetID() string { return c.params.PipelineID.String() }

// GetBefore retrieves the pipeline before the update.
func (c UpdatePipelineCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Pipelines, error) {
	queries := mdbp.New(tx)
	return queries.GetPipeline(ctx, mdbp.GetPipelineParams{PipelineID: c.params.PipelineID})
}

// Execute updates the pipeline in the database.
func (c UpdatePipelineCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdatePipeline(ctx, mdbp.UpdatePipelineParams{
		PipelineID:   c.params.PipelineID,
		TableName:    c.params.TableName,
		Operation:    c.params.Operation,
		Handler:      c.params.Handler,
		Priority:     int32(c.params.Priority),
		Enabled:      c.params.Enabled,
		Config:       c.params.Config,
		DateModified: types.TimestampNow(),
	})
}

// UpdatePipelineCmd returns a new update command for a pipeline.
func (d PsqlDatabase) UpdatePipelineCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdatePipelineParams) UpdatePipelineCmdPsql {
	return UpdatePipelineCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeletePipelineCmdPsql is an audited command for deleting a pipeline in PostgreSQL.
type DeletePipelineCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.PipelineID
	conn     *sql.DB
}

// Context returns the command's context.
func (c DeletePipelineCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeletePipelineCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeletePipelineCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeletePipelineCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c DeletePipelineCmdPsql) TableName() string { return "pipelines" }

// GetID returns the pipeline ID being deleted.
func (c DeletePipelineCmdPsql) GetID() string { return c.id.String() }

// GetBefore retrieves the pipeline before deletion.
func (c DeletePipelineCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Pipelines, error) {
	queries := mdbp.New(tx)
	return queries.GetPipeline(ctx, mdbp.GetPipelineParams{PipelineID: c.id})
}

// Execute deletes a pipeline from the database.
func (c DeletePipelineCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeletePipeline(ctx, mdbp.DeletePipelineParams{PipelineID: c.id})
}

// DeletePipelineCmd returns a new delete command for a pipeline.
func (d PsqlDatabase) DeletePipelineCmd(ctx context.Context, auditCtx audited.AuditContext, id types.PipelineID) DeletePipelineCmdPsql {
	return DeletePipelineCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
