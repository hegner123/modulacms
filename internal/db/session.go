package db

import (
	"database/sql"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
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
