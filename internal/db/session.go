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

// Sessions represents a user session with metadata.
type Sessions struct {
	SessionID   types.SessionID      `json:"session_id"`
	UserID      types.NullableUserID `json:"user_id"`
	CreatedAt   types.Timestamp      `json:"created_at"`
	ExpiresAt   types.Timestamp      `json:"expires_at"`
	LastAccess  sql.NullString       `json:"last_access"`
	IpAddress   sql.NullString       `json:"ip_address"`
	UserAgent   sql.NullString       `json:"user_agent"`
	SessionData sql.NullString       `json:"session_data"`
}

// CreateSessionParams holds parameters for creating a new session.
type CreateSessionParams struct {
	UserID      types.NullableUserID `json:"user_id"`
	CreatedAt   types.Timestamp      `json:"created_at"`
	ExpiresAt   types.Timestamp      `json:"expires_at"`
	LastAccess  sql.NullString       `json:"last_access"`
	IpAddress   sql.NullString       `json:"ip_address"`
	UserAgent   sql.NullString       `json:"user_agent"`
	SessionData sql.NullString       `json:"session_data"`
}

// UpdateSessionParams holds parameters for updating an existing session.
type UpdateSessionParams struct {
	UserID      types.NullableUserID `json:"user_id"`
	CreatedAt   types.Timestamp      `json:"created_at"`
	ExpiresAt   types.Timestamp      `json:"expires_at"`
	LastAccess  sql.NullString       `json:"last_access"`
	IpAddress   sql.NullString       `json:"ip_address"`
	UserAgent   sql.NullString       `json:"user_agent"`
	SessionData sql.NullString       `json:"session_data"`
	SessionID   types.SessionID      `json:"session_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringSession converts Sessions to StringSessions for table display
func MapStringSession(a Sessions) StringSessions {
	lastAccess := ""
	if a.LastAccess.Valid {
		lastAccess = a.LastAccess.String
	}
	ipAddress := ""
	if a.IpAddress.Valid {
		ipAddress = a.IpAddress.String
	}
	userAgent := ""
	if a.UserAgent.Valid {
		userAgent = a.UserAgent.String
	}
	sessionData := ""
	if a.SessionData.Valid {
		sessionData = a.SessionData.String
	}
	return StringSessions{
		SessionID:   a.SessionID.String(),
		UserID:      a.UserID.String(),
		CreatedAt:   a.CreatedAt.String(),
		ExpiresAt:   a.ExpiresAt.String(),
		LastAccess:  lastAccess,
		IpAddress:   ipAddress,
		UserAgent:   userAgent,
		SessionData: sessionData,
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

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

// QUERIES

// CountSessions returns the total number of sessions.
func (d Database) CountSessions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateSessionTable creates the sessions table.
func (d Database) CreateSessionTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

// CreateSession creates a new session with audit tracking.
func (d Database) CreateSession(ctx context.Context, ac audited.AuditContext, s CreateSessionParams) (*Sessions, error) {
	cmd := d.NewSessionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	r := d.MapSession(result)
	return &r, nil
}

// DeleteSession deletes a session with audit tracking.
func (d Database) DeleteSession(ctx context.Context, ac audited.AuditContext, id types.SessionID) error {
	cmd := d.DeleteSessionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetSession retrieves a session by ID.
func (d Database) GetSession(id types.SessionID) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetSession(d.Context, mdb.GetSessionParams{SessionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

// GetSessionByUserId retrieves a session by user ID.
func (d Database) GetSessionByUserId(userID types.NullableUserID) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, mdb.GetSessionByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(rows)
	return &res, nil
}

// ListSessions returns all sessions.
func (d Database) ListSessions() (*[]Sessions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Sessions: %v\n", err)
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateSession updates an existing session with audit tracking.
func (d Database) UpdateSession(ctx context.Context, ac audited.AuditContext, s UpdateSessionParams) (*string, error) {
	cmd := d.UpdateSessionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

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

// QUERIES

// CountSessions returns the total number of sessions.
func (d MysqlDatabase) CountSessions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateSessionTable creates the sessions table.
func (d MysqlDatabase) CreateSessionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

// CreateSession creates a new session with audit tracking.
func (d MysqlDatabase) CreateSession(ctx context.Context, ac audited.AuditContext, s CreateSessionParams) (*Sessions, error) {
	cmd := d.NewSessionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	r := d.MapSession(result)
	return &r, nil
}

// DeleteSession deletes a session with audit tracking.
func (d MysqlDatabase) DeleteSession(ctx context.Context, ac audited.AuditContext, id types.SessionID) error {
	cmd := d.DeleteSessionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetSession retrieves a session by ID.
func (d MysqlDatabase) GetSession(id types.SessionID) (*Sessions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetSession(d.Context, mdbm.GetSessionParams{SessionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

// GetSessionByUserId retrieves a session by user ID.
func (d MysqlDatabase) GetSessionByUserId(userID types.NullableUserID) (*Sessions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, mdbm.GetSessionByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(rows)
	return &res, nil
}

// ListSessions returns all sessions.
func (d MysqlDatabase) ListSessions() (*[]Sessions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Sessions: %v\n", err)
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateSession updates an existing session with audit tracking.
func (d MysqlDatabase) UpdateSession(ctx context.Context, ac audited.AuditContext, s UpdateSessionParams) (*string, error) {
	cmd := d.UpdateSessionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

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

// QUERIES

// CountSessions returns the total number of sessions.
func (d PsqlDatabase) CountSessions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateSessionTable creates the sessions table.
func (d PsqlDatabase) CreateSessionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

// CreateSession creates a new session with audit tracking.
func (d PsqlDatabase) CreateSession(ctx context.Context, ac audited.AuditContext, s CreateSessionParams) (*Sessions, error) {
	cmd := d.NewSessionCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	r := d.MapSession(result)
	return &r, nil
}

// DeleteSession deletes a session with audit tracking.
func (d PsqlDatabase) DeleteSession(ctx context.Context, ac audited.AuditContext, id types.SessionID) error {
	cmd := d.DeleteSessionCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetSession retrieves a session by ID.
func (d PsqlDatabase) GetSession(id types.SessionID) (*Sessions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetSession(d.Context, mdbp.GetSessionParams{SessionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

// GetSessionByUserId retrieves a session by user ID.
func (d PsqlDatabase) GetSessionByUserId(userID types.NullableUserID) (*Sessions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, mdbp.GetSessionByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(rows)
	return &res, nil
}

// ListSessions returns all sessions.
func (d PsqlDatabase) ListSessions() (*[]Sessions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Sessions: %v\n", err)
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateSession updates an existing session with audit tracking.
func (d PsqlDatabase) UpdateSession(ctx context.Context, ac audited.AuditContext, s UpdateSessionParams) (*string, error) {
	cmd := d.UpdateSessionCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &msg, nil
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

// Context returns the command context.
func (c NewSessionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewSessionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewSessionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewSessionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c NewSessionCmd) TableName() string { return "sessions" }

// Params returns the command parameters.
func (c NewSessionCmd) Params() any { return c.params }

// GetID extracts the ID from a created session.
func (c NewSessionCmd) GetID(x mdb.Sessions) string {
	return x.SessionID.String()
}

// Execute performs the create operation.
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

// NewSessionCmd creates a new create command.
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

// Context returns the command context.
func (c UpdateSessionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateSessionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateSessionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateSessionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c UpdateSessionCmd) TableName() string { return "sessions" }

// Params returns the command parameters.
func (c UpdateSessionCmd) Params() any { return c.params }

// GetID returns the session ID being updated.
func (c UpdateSessionCmd) GetID() string { return c.params.SessionID.String() }

// GetBefore retrieves the pre-update session state.
func (c UpdateSessionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Sessions, error) {
	queries := mdb.New(tx)
	return queries.GetSession(ctx, mdb.GetSessionParams{SessionID: c.params.SessionID})
}

// Execute performs the update operation.
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

// UpdateSessionCmd creates a new update command.
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

// Context returns the command context.
func (c DeleteSessionCmd) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteSessionCmd) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteSessionCmd) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteSessionCmd) Recorder() audited.ChangeEventRecorder { return SQLiteRecorder }

// TableName returns the table name.
func (c DeleteSessionCmd) TableName() string { return "sessions" }

// GetID returns the session ID being deleted.
func (c DeleteSessionCmd) GetID() string { return c.id.String() }

// GetBefore retrieves the session before deletion.
func (c DeleteSessionCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Sessions, error) {
	queries := mdb.New(tx)
	return queries.GetSession(ctx, mdb.GetSessionParams{SessionID: c.id})
}

// Execute performs the delete operation.
func (c DeleteSessionCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteSession(ctx, mdb.DeleteSessionParams{SessionID: c.id})
}

// DeleteSessionCmd creates a new delete command.
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

// Context returns the command context.
func (c NewSessionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewSessionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewSessionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewSessionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c NewSessionCmdMysql) TableName() string { return "sessions" }

// Params returns the command parameters.
func (c NewSessionCmdMysql) Params() any { return c.params }

// GetID extracts the ID from a created session.
func (c NewSessionCmdMysql) GetID(x mdbm.Sessions) string {
	return x.SessionID.String()
}

// Execute performs the create operation.
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

// NewSessionCmd creates a new create command.
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

// Context returns the command context.
func (c UpdateSessionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateSessionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateSessionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateSessionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c UpdateSessionCmdMysql) TableName() string { return "sessions" }

// Params returns the command parameters.
func (c UpdateSessionCmdMysql) Params() any { return c.params }

// GetID returns the session ID being updated.
func (c UpdateSessionCmdMysql) GetID() string { return c.params.SessionID.String() }

// GetBefore retrieves the pre-update session state.
func (c UpdateSessionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Sessions, error) {
	queries := mdbm.New(tx)
	return queries.GetSession(ctx, mdbm.GetSessionParams{SessionID: c.params.SessionID})
}

// Execute performs the update operation.
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

// UpdateSessionCmd creates a new update command.
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

// Context returns the command context.
func (c DeleteSessionCmdMysql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteSessionCmdMysql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteSessionCmdMysql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteSessionCmdMysql) Recorder() audited.ChangeEventRecorder { return MysqlRecorder }

// TableName returns the table name.
func (c DeleteSessionCmdMysql) TableName() string { return "sessions" }

// GetID returns the session ID being deleted.
func (c DeleteSessionCmdMysql) GetID() string { return c.id.String() }

// GetBefore retrieves the session before deletion.
func (c DeleteSessionCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Sessions, error) {
	queries := mdbm.New(tx)
	return queries.GetSession(ctx, mdbm.GetSessionParams{SessionID: c.id})
}

// Execute performs the delete operation.
func (c DeleteSessionCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteSession(ctx, mdbm.DeleteSessionParams{SessionID: c.id})
}

// DeleteSessionCmd creates a new delete command.
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

// Context returns the command context.
func (c NewSessionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c NewSessionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c NewSessionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c NewSessionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c NewSessionCmdPsql) TableName() string { return "sessions" }

// Params returns the command parameters.
func (c NewSessionCmdPsql) Params() any { return c.params }

// GetID extracts the ID from a created session.
func (c NewSessionCmdPsql) GetID(x mdbp.Sessions) string {
	return x.SessionID.String()
}

// Execute performs the create operation.
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

// NewSessionCmd creates a new create command.
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

// Context returns the command context.
func (c UpdateSessionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateSessionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateSessionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateSessionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c UpdateSessionCmdPsql) TableName() string { return "sessions" }

// Params returns the command parameters.
func (c UpdateSessionCmdPsql) Params() any { return c.params }

// GetID returns the session ID being updated.
func (c UpdateSessionCmdPsql) GetID() string { return c.params.SessionID.String() }

// GetBefore retrieves the pre-update session state.
func (c UpdateSessionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Sessions, error) {
	queries := mdbp.New(tx)
	return queries.GetSession(ctx, mdbp.GetSessionParams{SessionID: c.params.SessionID})
}

// Execute performs the update operation.
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

// UpdateSessionCmd creates a new update command.
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

// Context returns the command context.
func (c DeleteSessionCmdPsql) Context() context.Context { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteSessionCmdPsql) AuditContext() audited.AuditContext { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteSessionCmdPsql) Connection() *sql.DB { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteSessionCmdPsql) Recorder() audited.ChangeEventRecorder { return PsqlRecorder }

// TableName returns the table name.
func (c DeleteSessionCmdPsql) TableName() string { return "sessions" }

// GetID returns the session ID being deleted.
func (c DeleteSessionCmdPsql) GetID() string { return c.id.String() }

// GetBefore retrieves the session before deletion.
func (c DeleteSessionCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Sessions, error) {
	queries := mdbp.New(tx)
	return queries.GetSession(ctx, mdbp.GetSessionParams{SessionID: c.id})
}

// Execute performs the delete operation.
func (c DeleteSessionCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteSession(ctx, mdbp.DeleteSessionParams{SessionID: c.id})
}

// DeleteSessionCmd creates a new delete command.
func (d PsqlDatabase) DeleteSessionCmd(ctx context.Context, auditCtx audited.AuditContext, id types.SessionID) DeleteSessionCmdPsql {
	return DeleteSessionCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
