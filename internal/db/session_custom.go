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

// MapSession converts a sqlc-generated SQLite session to the wrapper type.
func (d Database) MapSession(a mdb.Sessions) Sessions {
	return Sessions{
		SessionID:   a.SessionID,
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  a.LastAccess,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

// MapCreateSessionParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateSessionParams(a CreateSessionParams) mdb.CreateSessionParams {
	return mdb.CreateSessionParams{
		SessionID:   types.NewSessionID(),
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  a.LastAccess,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

// MapUpdateSessionParams converts wrapper params to sqlc-generated SQLite update params.
func (d Database) MapUpdateSessionParams(a UpdateSessionParams) mdb.UpdateSessionParams {
	return mdb.UpdateSessionParams{
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  a.LastAccess,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
		SessionID:   a.SessionID,
	}
}

///////////////////////////////
// MYSQL MAPPERS
//////////////////////////////

// MapSession converts a sqlc-generated MySQL session to the wrapper type.
func (d MysqlDatabase) MapSession(a mdbm.Sessions) Sessions {
	return Sessions{
		SessionID:   a.SessionID,
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  StringToNullString(a.LastAccess.String()),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

// MapCreateSessionParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateSessionParams(a CreateSessionParams) mdbm.CreateSessionParams {
	return mdbm.CreateSessionParams{
		SessionID:   types.NewSessionID(),
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  StringToNTime(a.LastAccess.String).Time,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

// MapUpdateSessionParams converts wrapper params to sqlc-generated MySQL update params.
func (d MysqlDatabase) MapUpdateSessionParams(a UpdateSessionParams) mdbm.UpdateSessionParams {
	return mdbm.UpdateSessionParams{
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  StringToNTime(a.LastAccess.String).Time,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
		SessionID:   a.SessionID,
	}
}

///////////////////////////////
// POSTGRESQL MAPPERS
//////////////////////////////

// MapSession converts a sqlc-generated PostgreSQL session to the wrapper type.
func (d PsqlDatabase) MapSession(a mdbp.Sessions) Sessions {
	return Sessions{
		SessionID:   a.SessionID,
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  StringToNullString(NullTimeToString(a.LastAccess)),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

// MapCreateSessionParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateSessionParams(a CreateSessionParams) mdbp.CreateSessionParams {
	return mdbp.CreateSessionParams{
		SessionID:   types.NewSessionID(),
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  StringToNTime(a.LastAccess.String),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

// MapUpdateSessionParams converts wrapper params to sqlc-generated PostgreSQL update params.
func (d PsqlDatabase) MapUpdateSessionParams(a UpdateSessionParams) mdbp.UpdateSessionParams {
	return mdbp.UpdateSessionParams{
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  StringToNTime(a.LastAccess.String),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
		SessionID:   a.SessionID,
	}
}

///////////////////////////////
// SQLITE QUERIES
//////////////////////////////

// CreateSession creates a new session in SQLite via audited command.
func (d Database) CreateSession(ctx context.Context, ac audited.AuditContext, s CreateSessionParams) (*Sessions, error) {
	cmd := d.NewSessionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	r := d.MapSession(result)
	return &r, nil
}

// UpdateSession updates a session in SQLite via audited command.
func (d Database) UpdateSession(ctx context.Context, ac audited.AuditContext, s UpdateSessionParams) (*string, error) {
	cmd := d.UpdateSessionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &msg, nil
}

// DeleteSession deletes a session in SQLite via audited command.
func (d Database) DeleteSession(ctx context.Context, ac audited.AuditContext, id types.SessionID) error {
	cmd := d.DeleteSessionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

///////////////////////////////
// MYSQL QUERIES
//////////////////////////////

// CreateSession creates a new session in MySQL via audited command.
func (d MysqlDatabase) CreateSession(ctx context.Context, ac audited.AuditContext, s CreateSessionParams) (*Sessions, error) {
	cmd := d.NewSessionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	r := d.MapSession(result)
	return &r, nil
}

// UpdateSession updates a session in MySQL via audited command.
func (d MysqlDatabase) UpdateSession(ctx context.Context, ac audited.AuditContext, s UpdateSessionParams) (*string, error) {
	cmd := d.UpdateSessionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &msg, nil
}

// DeleteSession deletes a session in MySQL via audited command.
func (d MysqlDatabase) DeleteSession(ctx context.Context, ac audited.AuditContext, id types.SessionID) error {
	cmd := d.DeleteSessionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

///////////////////////////////
// POSTGRES QUERIES
//////////////////////////////

// CreateSession creates a new session in PostgreSQL via audited command.
func (d PsqlDatabase) CreateSession(ctx context.Context, ac audited.AuditContext, s CreateSessionParams) (*Sessions, error) {
	cmd := d.NewSessionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	r := d.MapSession(result)
	return &r, nil
}

// UpdateSession updates a session in PostgreSQL via audited command.
func (d PsqlDatabase) UpdateSession(ctx context.Context, ac audited.AuditContext, s UpdateSessionParams) (*string, error) {
	cmd := d.UpdateSessionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &msg, nil
}

// DeleteSession deletes a session in PostgreSQL via audited command.
func (d PsqlDatabase) DeleteSession(ctx context.Context, ac audited.AuditContext, id types.SessionID) error {
	cmd := d.DeleteSessionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewSessionCmd is an audited command for creating sessions in SQLite.
type NewSessionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateSessionParams
	conn     *sql.DB
}

func (c NewSessionCmd) Context() context.Context              { return c.ctx }
func (c NewSessionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewSessionCmd) Connection() *sql.DB                   { return c.conn }
func (c NewSessionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c NewSessionCmd) TableName() string                     { return "sessions" }
func (c NewSessionCmd) Params() any                           { return c.params }

func (c NewSessionCmd) GetID(x mdb.Sessions) string {
	return x.SessionID.String()
}

func (c NewSessionCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Sessions, error) {
	queries := mdb.New(tx)
	return queries.CreateSession(ctx, mdb.CreateSessionParams{
		SessionID:   types.NewSessionID(),
		UserID:      c.params.UserID,
		CreatedAt:   c.params.CreatedAt,
		ExpiresAt:   c.params.ExpiresAt,
		LastAccess:  c.params.LastAccess,
		IpAddress:   c.params.IpAddress,
		UserAgent:   c.params.UserAgent,
		SessionData: c.params.SessionData,
	})
}

func (d Database) NewSessionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateSessionParams) NewSessionCmd {
	return NewSessionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateSessionCmd is an audited command for updating sessions in SQLite.
type UpdateSessionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateSessionParams
	conn     *sql.DB
}

func (c UpdateSessionCmd) Context() context.Context              { return c.ctx }
func (c UpdateSessionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateSessionCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateSessionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c UpdateSessionCmd) TableName() string                     { return "sessions" }
func (c UpdateSessionCmd) Params() any                           { return c.params }
func (c UpdateSessionCmd) GetID() string                         { return c.params.SessionID.String() }

func (c UpdateSessionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Sessions, error) {
	queries := mdb.New(tx)
	return queries.GetSession(ctx, mdb.GetSessionParams{SessionID: c.params.SessionID})
}

func (c UpdateSessionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateSession(ctx, mdb.UpdateSessionParams{
		UserID:      c.params.UserID,
		CreatedAt:   c.params.CreatedAt,
		ExpiresAt:   c.params.ExpiresAt,
		LastAccess:  c.params.LastAccess,
		IpAddress:   c.params.IpAddress,
		UserAgent:   c.params.UserAgent,
		SessionData: c.params.SessionData,
		SessionID:   c.params.SessionID,
	})
}

func (d Database) UpdateSessionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateSessionParams) UpdateSessionCmd {
	return UpdateSessionCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteSessionCmd is an audited command for deleting sessions in SQLite.
type DeleteSessionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.SessionID
	conn     *sql.DB
}

func (c DeleteSessionCmd) Context() context.Context              { return c.ctx }
func (c DeleteSessionCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteSessionCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteSessionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }
func (c DeleteSessionCmd) TableName() string                     { return "sessions" }
func (c DeleteSessionCmd) GetID() string                         { return c.id.String() }

func (c DeleteSessionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Sessions, error) {
	queries := mdb.New(tx)
	return queries.GetSession(ctx, mdb.GetSessionParams{SessionID: c.id})
}

func (c DeleteSessionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteSession(ctx, mdb.DeleteSessionParams{SessionID: c.id})
}

func (d Database) DeleteSessionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.SessionID) DeleteSessionCmd {
	return DeleteSessionCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewSessionCmdMysql is an audited command for creating sessions in MySQL.
type NewSessionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateSessionParams
	conn     *sql.DB
}

func (c NewSessionCmdMysql) Context() context.Context              { return c.ctx }
func (c NewSessionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewSessionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewSessionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c NewSessionCmdMysql) TableName() string                     { return "sessions" }
func (c NewSessionCmdMysql) Params() any                           { return c.params }

func (c NewSessionCmdMysql) GetID(x mdbm.Sessions) string {
	return x.SessionID.String()
}

func (c NewSessionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Sessions, error) {
	id := types.NewSessionID()
	queries := mdbm.New(tx)
	err := queries.CreateSession(ctx, mdbm.CreateSessionParams{
		SessionID:   id,
		UserID:      c.params.UserID,
		CreatedAt:   c.params.CreatedAt,
		ExpiresAt:   c.params.ExpiresAt,
		LastAccess:  StringToNTime(c.params.LastAccess.String).Time,
		IpAddress:   c.params.IpAddress,
		UserAgent:   c.params.UserAgent,
		SessionData: c.params.SessionData,
	})
	if err != nil {
		return mdbm.Sessions{}, fmt.Errorf("Failed to CreateSession: %w", err)
	}
	return queries.GetSession(ctx, mdbm.GetSessionParams{SessionID: id})
}

func (d MysqlDatabase) NewSessionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateSessionParams) NewSessionCmdMysql {
	return NewSessionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateSessionCmdMysql is an audited command for updating sessions in MySQL.
type UpdateSessionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateSessionParams
	conn     *sql.DB
}

func (c UpdateSessionCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateSessionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateSessionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateSessionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c UpdateSessionCmdMysql) TableName() string                     { return "sessions" }
func (c UpdateSessionCmdMysql) Params() any                           { return c.params }
func (c UpdateSessionCmdMysql) GetID() string                         { return c.params.SessionID.String() }

func (c UpdateSessionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Sessions, error) {
	queries := mdbm.New(tx)
	return queries.GetSession(ctx, mdbm.GetSessionParams{SessionID: c.params.SessionID})
}

func (c UpdateSessionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateSession(ctx, mdbm.UpdateSessionParams{
		UserID:      c.params.UserID,
		CreatedAt:   c.params.CreatedAt,
		ExpiresAt:   c.params.ExpiresAt,
		LastAccess:  StringToNTime(c.params.LastAccess.String).Time,
		IpAddress:   c.params.IpAddress,
		UserAgent:   c.params.UserAgent,
		SessionData: c.params.SessionData,
		SessionID:   c.params.SessionID,
	})
}

func (d MysqlDatabase) UpdateSessionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateSessionParams) UpdateSessionCmdMysql {
	return UpdateSessionCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteSessionCmdMysql is an audited command for deleting sessions in MySQL.
type DeleteSessionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.SessionID
	conn     *sql.DB
}

func (c DeleteSessionCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteSessionCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteSessionCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteSessionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }
func (c DeleteSessionCmdMysql) TableName() string                     { return "sessions" }
func (c DeleteSessionCmdMysql) GetID() string                         { return c.id.String() }

func (c DeleteSessionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Sessions, error) {
	queries := mdbm.New(tx)
	return queries.GetSession(ctx, mdbm.GetSessionParams{SessionID: c.id})
}

func (c DeleteSessionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteSession(ctx, mdbm.DeleteSessionParams{SessionID: c.id})
}

func (d MysqlDatabase) DeleteSessionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.SessionID) DeleteSessionCmdMysql {
	return DeleteSessionCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewSessionCmdPsql is an audited command for creating sessions in PostgreSQL.
type NewSessionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateSessionParams
	conn     *sql.DB
}

func (c NewSessionCmdPsql) Context() context.Context              { return c.ctx }
func (c NewSessionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewSessionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewSessionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c NewSessionCmdPsql) TableName() string                     { return "sessions" }
func (c NewSessionCmdPsql) Params() any                           { return c.params }

func (c NewSessionCmdPsql) GetID(x mdbp.Sessions) string {
	return x.SessionID.String()
}

func (c NewSessionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Sessions, error) {
	queries := mdbp.New(tx)
	return queries.CreateSession(ctx, mdbp.CreateSessionParams{
		SessionID:   types.NewSessionID(),
		UserID:      c.params.UserID,
		CreatedAt:   c.params.CreatedAt,
		ExpiresAt:   c.params.ExpiresAt,
		LastAccess:  StringToNTime(c.params.LastAccess.String),
		IpAddress:   c.params.IpAddress,
		UserAgent:   c.params.UserAgent,
		SessionData: c.params.SessionData,
	})
}

func (d PsqlDatabase) NewSessionCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateSessionParams) NewSessionCmdPsql {
	return NewSessionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateSessionCmdPsql is an audited command for updating sessions in PostgreSQL.
type UpdateSessionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateSessionParams
	conn     *sql.DB
}

func (c UpdateSessionCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateSessionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateSessionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateSessionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c UpdateSessionCmdPsql) TableName() string                     { return "sessions" }
func (c UpdateSessionCmdPsql) Params() any                           { return c.params }
func (c UpdateSessionCmdPsql) GetID() string                         { return c.params.SessionID.String() }

func (c UpdateSessionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Sessions, error) {
	queries := mdbp.New(tx)
	return queries.GetSession(ctx, mdbp.GetSessionParams{SessionID: c.params.SessionID})
}

func (c UpdateSessionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateSession(ctx, mdbp.UpdateSessionParams{
		UserID:      c.params.UserID,
		CreatedAt:   c.params.CreatedAt,
		ExpiresAt:   c.params.ExpiresAt,
		LastAccess:  StringToNTime(c.params.LastAccess.String),
		IpAddress:   c.params.IpAddress,
		UserAgent:   c.params.UserAgent,
		SessionData: c.params.SessionData,
		SessionID:   c.params.SessionID,
	})
}

func (d PsqlDatabase) UpdateSessionCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateSessionParams) UpdateSessionCmdPsql {
	return UpdateSessionCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteSessionCmdPsql is an audited command for deleting sessions in PostgreSQL.
type DeleteSessionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.SessionID
	conn     *sql.DB
}

func (c DeleteSessionCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteSessionCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteSessionCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteSessionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }
func (c DeleteSessionCmdPsql) TableName() string                     { return "sessions" }
func (c DeleteSessionCmdPsql) GetID() string                         { return c.id.String() }

func (c DeleteSessionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Sessions, error) {
	queries := mdbp.New(tx)
	return queries.GetSession(ctx, mdbp.GetSessionParams{SessionID: c.id})
}

func (c DeleteSessionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteSession(ctx, mdbp.DeleteSessionParams{SessionID: c.id})
}

func (d PsqlDatabase) DeleteSessionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.SessionID) DeleteSessionCmdPsql {
	return DeleteSessionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
