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

// StringWebhook is the string-typed version of Webhook for TUI display.
type StringWebhook struct {
	WebhookID    string `json:"webhook_id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	IsActive     string `json:"is_active"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// marshalEvents converts a []string to a JSON string for storage.
func marshalEvents(events []string) string {
	if events == nil {
		return "[]"
	}
	b, err := json.Marshal(events)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// unmarshalEvents converts a JSON string to a []string.
func unmarshalEvents(s string) []string {
	var events []string
	if err := json.Unmarshal([]byte(s), &events); err != nil {
		return []string{}
	}
	return events
}

// marshalHeaders converts a map[string]string to a JSON string for storage.
func marshalHeaders(headers map[string]string) string {
	if headers == nil {
		return "{}"
	}
	b, err := json.Marshal(headers)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// unmarshalHeaders converts a JSON string to a map[string]string.
func unmarshalHeaders(s string) map[string]string {
	var headers map[string]string
	if err := json.Unmarshal([]byte(s), &headers); err != nil {
		return map[string]string{}
	}
	return headers
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MapWebhook converts a sqlc-generated SQLite webhook to the wrapper type.
func (d Database) MapWebhook(a mdb.Webhooks) Webhook {
	return Webhook{
		WebhookID:    a.WebhookID,
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       unmarshalEvents(a.Events),
		IsActive:     a.IsActive != 0,
		Headers:      unmarshalHeaders(a.Headers),
		AuthorID:     a.AuthorID.ID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateWebhookParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateWebhookParams(a CreateWebhookParams) mdb.CreateWebhookParams {
	return mdb.CreateWebhookParams{
		WebhookID:    types.NewWebhookID(),
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       marshalEvents(a.Events),
		IsActive:     boolToInt64(a.IsActive),
		Headers:      marshalHeaders(a.Headers),
		AuthorID:     types.NullableUserID{ID: a.AuthorID, Valid: true},
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateWebhookParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdateWebhookParams(a UpdateWebhookParams) mdb.UpdateWebhookParams {
	return mdb.UpdateWebhookParams{
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       marshalEvents(a.Events),
		IsActive:     boolToInt64(a.IsActive),
		Headers:      marshalHeaders(a.Headers),
		DateModified: a.DateModified,
		WebhookID:    a.WebhookID,
	}
}

// CreateWebhook inserts a new webhook with audit trail.
func (d Database) CreateWebhook(ctx context.Context, ac audited.AuditContext, s CreateWebhookParams) (*Webhook, error) {
	cmd := d.NewWebhookCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	r := d.MapWebhook(result)
	return &r, nil
}

// UpdateWebhook modifies an existing webhook with audit trail.
func (d Database) UpdateWebhook(ctx context.Context, ac audited.AuditContext, s UpdateWebhookParams) error {
	cmd := d.UpdateWebhookCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// DeleteWebhook removes a webhook with audit trail.
func (d Database) DeleteWebhook(ctx context.Context, ac audited.AuditContext, id types.WebhookID) error {
	cmd := d.DeleteWebhookCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// ----- SQLite CREATE COMMAND -----

// NewWebhookCmd is an audited command for creating webhooks on SQLite.
type NewWebhookCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateWebhookParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewWebhookCmd) Context() context.Context              { return c.ctx }
func (c NewWebhookCmd) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c NewWebhookCmd) Connection() *sql.DB                   { return c.conn }
func (c NewWebhookCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewWebhookCmd) TableName() string                     { return "webhooks" }
func (c NewWebhookCmd) Params() any                           { return c.params }
func (c NewWebhookCmd) GetID(u mdb.Webhooks) string           { return string(u.WebhookID) }

func (c NewWebhookCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Webhooks, error) {
	queries := mdb.New(tx)
	return queries.CreateWebhook(ctx, mdb.CreateWebhookParams{
		WebhookID:    types.NewWebhookID(),
		Name:         c.params.Name,
		URL:          c.params.URL,
		Secret:       c.params.Secret,
		Events:       marshalEvents(c.params.Events),
		IsActive:     boolToInt64(c.params.IsActive),
		Headers:      marshalHeaders(c.params.Headers),
		AuthorID:     types.NullableUserID{ID: c.params.AuthorID, Valid: true},
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d Database) NewWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateWebhookParams) NewWebhookCmd {
	return NewWebhookCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE COMMAND -----

// UpdateWebhookCmd is an audited command for updating webhooks on SQLite.
type UpdateWebhookCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateWebhookParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateWebhookCmd) Context() context.Context              { return c.ctx }
func (c UpdateWebhookCmd) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c UpdateWebhookCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateWebhookCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateWebhookCmd) TableName() string                     { return "webhooks" }
func (c UpdateWebhookCmd) Params() any                           { return c.params }
func (c UpdateWebhookCmd) GetID() string                         { return string(c.params.WebhookID) }

func (c UpdateWebhookCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Webhooks, error) {
	queries := mdb.New(tx)
	return queries.GetWebhook(ctx, mdb.GetWebhookParams{WebhookID: c.params.WebhookID})
}

func (c UpdateWebhookCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateWebhook(ctx, mdb.UpdateWebhookParams{
		Name:         c.params.Name,
		URL:          c.params.URL,
		Secret:       c.params.Secret,
		Events:       marshalEvents(c.params.Events),
		IsActive:     boolToInt64(c.params.IsActive),
		Headers:      marshalHeaders(c.params.Headers),
		DateModified: c.params.DateModified,
		WebhookID:    c.params.WebhookID,
	})
}

func (d Database) UpdateWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateWebhookParams) UpdateWebhookCmd {
	return UpdateWebhookCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE COMMAND -----

// DeleteWebhookCmd is an audited command for deleting webhooks on SQLite.
type DeleteWebhookCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.WebhookID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteWebhookCmd) Context() context.Context              { return c.ctx }
func (c DeleteWebhookCmd) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c DeleteWebhookCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteWebhookCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteWebhookCmd) TableName() string                     { return "webhooks" }
func (c DeleteWebhookCmd) GetID() string                         { return string(c.id) }

func (c DeleteWebhookCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Webhooks, error) {
	queries := mdb.New(tx)
	return queries.GetWebhook(ctx, mdb.GetWebhookParams{WebhookID: c.id})
}

func (c DeleteWebhookCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteWebhook(ctx, mdb.DeleteWebhookParams{WebhookID: c.id})
}

func (d Database) DeleteWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, id types.WebhookID) DeleteWebhookCmd {
	return DeleteWebhookCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MapWebhook converts a sqlc-generated MySQL webhook to the wrapper type.
func (d MysqlDatabase) MapWebhook(a mdbm.Webhooks) Webhook {
	return Webhook{
		WebhookID:    a.WebhookID,
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       unmarshalEvents(a.Events),
		IsActive:     a.IsActive != 0,
		Headers:      unmarshalHeaders(a.Headers),
		AuthorID:     a.AuthorID.ID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateWebhookParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateWebhookParams(a CreateWebhookParams) mdbm.CreateWebhookParams {
	return mdbm.CreateWebhookParams{
		WebhookID:    types.NewWebhookID(),
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       marshalEvents(a.Events),
		IsActive:     boolToInt64(a.IsActive),
		Headers:      marshalHeaders(a.Headers),
		AuthorID:     types.NullableUserID{ID: a.AuthorID, Valid: true},
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateWebhookParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdateWebhookParams(a UpdateWebhookParams) mdbm.UpdateWebhookParams {
	return mdbm.UpdateWebhookParams{
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       marshalEvents(a.Events),
		IsActive:     boolToInt64(a.IsActive),
		Headers:      marshalHeaders(a.Headers),
		DateModified: a.DateModified,
		WebhookID:    a.WebhookID,
	}
}

// CreateWebhook inserts a new webhook with audit trail.
func (d MysqlDatabase) CreateWebhook(ctx context.Context, ac audited.AuditContext, s CreateWebhookParams) (*Webhook, error) {
	cmd := d.NewWebhookCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	r := d.MapWebhook(result)
	return &r, nil
}

// UpdateWebhook modifies an existing webhook with audit trail.
func (d MysqlDatabase) UpdateWebhook(ctx context.Context, ac audited.AuditContext, s UpdateWebhookParams) error {
	cmd := d.UpdateWebhookCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// DeleteWebhook removes a webhook with audit trail.
func (d MysqlDatabase) DeleteWebhook(ctx context.Context, ac audited.AuditContext, id types.WebhookID) error {
	cmd := d.DeleteWebhookCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// ----- MySQL CREATE COMMAND -----

// NewWebhookCmdMysql is an audited command for creating webhooks on MySQL.
type NewWebhookCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateWebhookParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewWebhookCmdMysql) Context() context.Context              { return c.ctx }
func (c NewWebhookCmdMysql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c NewWebhookCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewWebhookCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewWebhookCmdMysql) TableName() string                     { return "webhooks" }
func (c NewWebhookCmdMysql) Params() any                           { return c.params }
func (c NewWebhookCmdMysql) GetID(u mdbm.Webhooks) string          { return string(u.WebhookID) }

func (c NewWebhookCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Webhooks, error) {
	queries := mdbm.New(tx)
	params := mdbm.CreateWebhookParams{
		WebhookID:    types.NewWebhookID(),
		Name:         c.params.Name,
		URL:          c.params.URL,
		Secret:       c.params.Secret,
		Events:       marshalEvents(c.params.Events),
		IsActive:     boolToInt64(c.params.IsActive),
		Headers:      marshalHeaders(c.params.Headers),
		AuthorID:     types.NullableUserID{ID: c.params.AuthorID, Valid: true},
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	}
	if err := queries.CreateWebhook(ctx, params); err != nil {
		return mdbm.Webhooks{}, err
	}
	return queries.GetWebhook(ctx, mdbm.GetWebhookParams{WebhookID: params.WebhookID})
}

func (d MysqlDatabase) NewWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateWebhookParams) NewWebhookCmdMysql {
	return NewWebhookCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE COMMAND -----

// UpdateWebhookCmdMysql is an audited command for updating webhooks on MySQL.
type UpdateWebhookCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateWebhookParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateWebhookCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateWebhookCmdMysql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c UpdateWebhookCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateWebhookCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateWebhookCmdMysql) TableName() string                     { return "webhooks" }
func (c UpdateWebhookCmdMysql) Params() any                           { return c.params }
func (c UpdateWebhookCmdMysql) GetID() string                         { return string(c.params.WebhookID) }

func (c UpdateWebhookCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Webhooks, error) {
	queries := mdbm.New(tx)
	return queries.GetWebhook(ctx, mdbm.GetWebhookParams{WebhookID: c.params.WebhookID})
}

func (c UpdateWebhookCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateWebhook(ctx, mdbm.UpdateWebhookParams{
		Name:         c.params.Name,
		URL:          c.params.URL,
		Secret:       c.params.Secret,
		Events:       marshalEvents(c.params.Events),
		IsActive:     boolToInt64(c.params.IsActive),
		Headers:      marshalHeaders(c.params.Headers),
		DateModified: c.params.DateModified,
		WebhookID:    c.params.WebhookID,
	})
}

func (d MysqlDatabase) UpdateWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateWebhookParams) UpdateWebhookCmdMysql {
	return UpdateWebhookCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE COMMAND -----

// DeleteWebhookCmdMysql is an audited command for deleting webhooks on MySQL.
type DeleteWebhookCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.WebhookID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteWebhookCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteWebhookCmdMysql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c DeleteWebhookCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteWebhookCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteWebhookCmdMysql) TableName() string                     { return "webhooks" }
func (c DeleteWebhookCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteWebhookCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Webhooks, error) {
	queries := mdbm.New(tx)
	return queries.GetWebhook(ctx, mdbm.GetWebhookParams{WebhookID: c.id})
}

func (c DeleteWebhookCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteWebhook(ctx, mdbm.DeleteWebhookParams{WebhookID: c.id})
}

func (d MysqlDatabase) DeleteWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, id types.WebhookID) DeleteWebhookCmdMysql {
	return DeleteWebhookCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MapWebhook converts a sqlc-generated PostgreSQL webhook to the wrapper type.
func (d PsqlDatabase) MapWebhook(a mdbp.Webhooks) Webhook {
	return Webhook{
		WebhookID:    a.WebhookID,
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       unmarshalEvents(a.Events),
		IsActive:     a.IsActive != 0,
		Headers:      unmarshalHeaders(a.Headers),
		AuthorID:     a.AuthorID.ID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateWebhookParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateWebhookParams(a CreateWebhookParams) mdbp.CreateWebhookParams {
	return mdbp.CreateWebhookParams{
		WebhookID:    types.NewWebhookID(),
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       marshalEvents(a.Events),
		IsActive:     boolToInt64(a.IsActive),
		Headers:      marshalHeaders(a.Headers),
		AuthorID:     types.NullableUserID{ID: a.AuthorID, Valid: true},
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateWebhookParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdateWebhookParams(a UpdateWebhookParams) mdbp.UpdateWebhookParams {
	return mdbp.UpdateWebhookParams{
		Name:         a.Name,
		URL:          a.URL,
		Secret:       a.Secret,
		Events:       marshalEvents(a.Events),
		IsActive:     boolToInt64(a.IsActive),
		Headers:      marshalHeaders(a.Headers),
		DateModified: a.DateModified,
		WebhookID:    a.WebhookID,
	}
}

// CreateWebhook inserts a new webhook with audit trail.
func (d PsqlDatabase) CreateWebhook(ctx context.Context, ac audited.AuditContext, s CreateWebhookParams) (*Webhook, error) {
	cmd := d.NewWebhookCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	r := d.MapWebhook(result)
	return &r, nil
}

// UpdateWebhook modifies an existing webhook with audit trail.
func (d PsqlDatabase) UpdateWebhook(ctx context.Context, ac audited.AuditContext, s UpdateWebhookParams) error {
	cmd := d.UpdateWebhookCmd(ctx, ac, s)
	return audited.Update(cmd)
}

// DeleteWebhook removes a webhook with audit trail.
func (d PsqlDatabase) DeleteWebhook(ctx context.Context, ac audited.AuditContext, id types.WebhookID) error {
	cmd := d.DeleteWebhookCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// ----- PostgreSQL CREATE COMMAND -----

// NewWebhookCmdPsql is an audited command for creating webhooks on PostgreSQL.
type NewWebhookCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateWebhookParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewWebhookCmdPsql) Context() context.Context              { return c.ctx }
func (c NewWebhookCmdPsql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c NewWebhookCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewWebhookCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewWebhookCmdPsql) TableName() string                     { return "webhooks" }
func (c NewWebhookCmdPsql) Params() any                           { return c.params }
func (c NewWebhookCmdPsql) GetID(u mdbp.Webhooks) string          { return string(u.WebhookID) }

func (c NewWebhookCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Webhooks, error) {
	queries := mdbp.New(tx)
	return queries.CreateWebhook(ctx, mdbp.CreateWebhookParams{
		WebhookID:    types.NewWebhookID(),
		Name:         c.params.Name,
		URL:          c.params.URL,
		Secret:       c.params.Secret,
		Events:       marshalEvents(c.params.Events),
		IsActive:     boolToInt64(c.params.IsActive),
		Headers:      marshalHeaders(c.params.Headers),
		AuthorID:     types.NullableUserID{ID: c.params.AuthorID, Valid: true},
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d PsqlDatabase) NewWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateWebhookParams) NewWebhookCmdPsql {
	return NewWebhookCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE COMMAND -----

// UpdateWebhookCmdPsql is an audited command for updating webhooks on PostgreSQL.
type UpdateWebhookCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateWebhookParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateWebhookCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateWebhookCmdPsql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c UpdateWebhookCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateWebhookCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateWebhookCmdPsql) TableName() string                     { return "webhooks" }
func (c UpdateWebhookCmdPsql) Params() any                           { return c.params }
func (c UpdateWebhookCmdPsql) GetID() string                         { return string(c.params.WebhookID) }

func (c UpdateWebhookCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Webhooks, error) {
	queries := mdbp.New(tx)
	return queries.GetWebhook(ctx, mdbp.GetWebhookParams{WebhookID: c.params.WebhookID})
}

func (c UpdateWebhookCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateWebhook(ctx, mdbp.UpdateWebhookParams{
		Name:         c.params.Name,
		URL:          c.params.URL,
		Secret:       c.params.Secret,
		Events:       marshalEvents(c.params.Events),
		IsActive:     boolToInt64(c.params.IsActive),
		Headers:      marshalHeaders(c.params.Headers),
		DateModified: c.params.DateModified,
		WebhookID:    c.params.WebhookID,
	})
}

func (d PsqlDatabase) UpdateWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateWebhookParams) UpdateWebhookCmdPsql {
	return UpdateWebhookCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE COMMAND -----

// DeleteWebhookCmdPsql is an audited command for deleting webhooks on PostgreSQL.
type DeleteWebhookCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.WebhookID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteWebhookCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteWebhookCmdPsql) AuditContext() audited.AuditContext    { return c.auditCtx }
func (c DeleteWebhookCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteWebhookCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteWebhookCmdPsql) TableName() string                     { return "webhooks" }
func (c DeleteWebhookCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteWebhookCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Webhooks, error) {
	queries := mdbp.New(tx)
	return queries.GetWebhook(ctx, mdbp.GetWebhookParams{WebhookID: c.id})
}

func (c DeleteWebhookCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteWebhook(ctx, mdbp.DeleteWebhookParams{WebhookID: c.id})
}

func (d PsqlDatabase) DeleteWebhookCmd(ctx context.Context, auditCtx audited.AuditContext, id types.WebhookID) DeleteWebhookCmdPsql {
	return DeleteWebhookCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
