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

type CreateSessionParams struct {
	UserID      types.NullableUserID `json:"user_id"`
	CreatedAt   types.Timestamp      `json:"created_at"`
	ExpiresAt   types.Timestamp      `json:"expires_at"`
	LastAccess  sql.NullString       `json:"last_access"`
	IpAddress   sql.NullString       `json:"ip_address"`
	UserAgent   sql.NullString       `json:"user_agent"`
	SessionData sql.NullString       `json:"session_data"`
}

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

func (d Database) CountSessions() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateSessionTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

func (d Database) CreateSession(s CreateSessionParams) (*Sessions, error) {
	params := d.MapCreateSessionParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateSession(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateSession.\n %v\n", err)
		return nil, e
	}
	session := d.MapSession(row)
	return &session, nil
}

func (d Database) DeleteSession(id types.SessionID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteSession(d.Context, mdb.DeleteSessionParams{SessionID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}
	return nil
}

func (d Database) GetSession(id types.SessionID) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetSession(d.Context, mdb.GetSessionParams{SessionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

func (d Database) GetSessionByUserId(userID types.NullableUserID) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, mdb.GetSessionByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(rows)
	return &res, nil
}

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

func (d Database) UpdateSession(s UpdateSessionParams) (*string, error) {
	params := d.MapUpdateSessionParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateSession(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update session, %v", err)
	}
	u := fmt.Sprintf("Successfully updated session %v\n", s.SessionID)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

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

func (d MysqlDatabase) CountSessions() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateSessionTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateSession(s CreateSessionParams) (*Sessions, error) {
	params := d.MapCreateSessionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateSession(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateSession.\n %v\n", err)
		return nil, e
	}
	row, err := queries.GetSession(d.Context, mdbm.GetSessionParams{SessionID: params.SessionID})
	if err != nil {
		return nil, fmt.Errorf("Failed to get last inserted Session: %v\n", err)
	}
	session := d.MapSession(row)
	return &session, nil
}

func (d MysqlDatabase) DeleteSession(id types.SessionID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteSession(d.Context, mdbm.DeleteSessionParams{SessionID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetSession(id types.SessionID) (*Sessions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetSession(d.Context, mdbm.GetSessionParams{SessionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

func (d MysqlDatabase) GetSessionByUserId(userID types.NullableUserID) (*Sessions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, mdbm.GetSessionByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(rows)
	return &res, nil
}

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

func (d MysqlDatabase) UpdateSession(s UpdateSessionParams) (*string, error) {
	params := d.MapUpdateSessionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateSession(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update session, %v", err)
	}
	u := fmt.Sprintf("Successfully updated session %v\n", s.SessionID)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

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

func (d PsqlDatabase) CountSessions() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateSessionTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateSessionTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateSession(s CreateSessionParams) (*Sessions, error) {
	params := d.MapCreateSessionParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateSession(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateSession.\n %v\n", err)
		return nil, e
	}
	session := d.MapSession(row)
	return &session, nil
}

func (d PsqlDatabase) DeleteSession(id types.SessionID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteSession(d.Context, mdbp.DeleteSessionParams{SessionID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetSession(id types.SessionID) (*Sessions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetSession(d.Context, mdbp.GetSessionParams{SessionID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

func (d PsqlDatabase) GetSessionByUserId(userID types.NullableUserID) (*Sessions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, mdbp.GetSessionByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapSession(rows)
	return &res, nil
}

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

func (d PsqlDatabase) UpdateSession(s UpdateSessionParams) (*string, error) {
	params := d.MapUpdateSessionParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateSession(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update session, %v", err)
	}
	u := fmt.Sprintf("Successfully updated session %v\n", s.SessionID)
	return &u, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewSessionCmd implements audited.CreateCommand[mdb.Sessions] for SQLite.
type NewSessionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateSessionParams
	conn     *sql.DB
}

func (c NewSessionCmd) Context() context.Context                    { return c.ctx }
func (c NewSessionCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewSessionCmd) Connection() *sql.DB                         { return c.conn }
func (c NewSessionCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c NewSessionCmd) TableName() string                           { return "sessions" }
func (c NewSessionCmd) Params() any                                 { return c.params }

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

// UpdateSessionCmd implements audited.UpdateCommand[mdb.Sessions] for SQLite.
type UpdateSessionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateSessionParams
	conn     *sql.DB
}

func (c UpdateSessionCmd) Context() context.Context                    { return c.ctx }
func (c UpdateSessionCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateSessionCmd) Connection() *sql.DB                         { return c.conn }
func (c UpdateSessionCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c UpdateSessionCmd) TableName() string                           { return "sessions" }
func (c UpdateSessionCmd) Params() any                                 { return c.params }
func (c UpdateSessionCmd) GetID() string                               { return c.params.SessionID.String() }

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

// DeleteSessionCmd implements audited.DeleteCommand[mdb.Sessions] for SQLite.
type DeleteSessionCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.SessionID
	conn     *sql.DB
}

func (c DeleteSessionCmd) Context() context.Context                    { return c.ctx }
func (c DeleteSessionCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteSessionCmd) Connection() *sql.DB                         { return c.conn }
func (c DeleteSessionCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c DeleteSessionCmd) TableName() string                           { return "sessions" }
func (c DeleteSessionCmd) GetID() string                               { return c.id.String() }

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

// NewSessionCmdMysql implements audited.CreateCommand[mdbm.Sessions] for MySQL.
type NewSessionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateSessionParams
	conn     *sql.DB
}

func (c NewSessionCmdMysql) Context() context.Context                    { return c.ctx }
func (c NewSessionCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewSessionCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c NewSessionCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c NewSessionCmdMysql) TableName() string                           { return "sessions" }
func (c NewSessionCmdMysql) Params() any                                 { return c.params }

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

// UpdateSessionCmdMysql implements audited.UpdateCommand[mdbm.Sessions] for MySQL.
type UpdateSessionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateSessionParams
	conn     *sql.DB
}

func (c UpdateSessionCmdMysql) Context() context.Context                    { return c.ctx }
func (c UpdateSessionCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateSessionCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c UpdateSessionCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c UpdateSessionCmdMysql) TableName() string                           { return "sessions" }
func (c UpdateSessionCmdMysql) Params() any                                 { return c.params }
func (c UpdateSessionCmdMysql) GetID() string                               { return c.params.SessionID.String() }

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

// DeleteSessionCmdMysql implements audited.DeleteCommand[mdbm.Sessions] for MySQL.
type DeleteSessionCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.SessionID
	conn     *sql.DB
}

func (c DeleteSessionCmdMysql) Context() context.Context                    { return c.ctx }
func (c DeleteSessionCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteSessionCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c DeleteSessionCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c DeleteSessionCmdMysql) TableName() string                           { return "sessions" }
func (c DeleteSessionCmdMysql) GetID() string                               { return c.id.String() }

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

// NewSessionCmdPsql implements audited.CreateCommand[mdbp.Sessions] for PostgreSQL.
type NewSessionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateSessionParams
	conn     *sql.DB
}

func (c NewSessionCmdPsql) Context() context.Context                    { return c.ctx }
func (c NewSessionCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewSessionCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c NewSessionCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c NewSessionCmdPsql) TableName() string                           { return "sessions" }
func (c NewSessionCmdPsql) Params() any                                 { return c.params }

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

// UpdateSessionCmdPsql implements audited.UpdateCommand[mdbp.Sessions] for PostgreSQL.
type UpdateSessionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateSessionParams
	conn     *sql.DB
}

func (c UpdateSessionCmdPsql) Context() context.Context                    { return c.ctx }
func (c UpdateSessionCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateSessionCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c UpdateSessionCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c UpdateSessionCmdPsql) TableName() string                           { return "sessions" }
func (c UpdateSessionCmdPsql) Params() any                                 { return c.params }
func (c UpdateSessionCmdPsql) GetID() string                               { return c.params.SessionID.String() }

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

// DeleteSessionCmdPsql implements audited.DeleteCommand[mdbp.Sessions] for PostgreSQL.
type DeleteSessionCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.SessionID
	conn     *sql.DB
}

func (c DeleteSessionCmdPsql) Context() context.Context                    { return c.ctx }
func (c DeleteSessionCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteSessionCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c DeleteSessionCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c DeleteSessionCmdPsql) TableName() string                           { return "sessions" }
func (c DeleteSessionCmdPsql) GetID() string                               { return c.id.String() }

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
