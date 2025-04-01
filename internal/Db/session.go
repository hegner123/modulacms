package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/db-mysql"
	mdbp "github.com/hegner123/modulacms/db-psql"
	mdb "github.com/hegner123/modulacms/db-sqlite"
)

// /////////////////////////////
// STRUCTS
// ////////////////////////////
type Sessions struct {
	SessionID   int64          `json:"session_id"`
	UserID      int64          `json:"user_id"`
	CreatedAt   sql.NullString `json:"created_at"`
	ExpiresAt   sql.NullString `json:"expires_at"`
	LastAccess  sql.NullString `json:"last_access"`
	IpAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	SessionData sql.NullString `json:"session_data"`
}

type CreateSessionParams struct {
	UserID      int64          `json:"user_id"`
	CreatedAt   sql.NullString `json:"created_at"`
	ExpiresAt   sql.NullString `json:"expires_at"`
	LastAccess  sql.NullString `json:"last_access"`
	IpAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	SessionData sql.NullString `json:"session_data"`
}

type UpdateSessionParams struct {
	UserID      int64          `json:"user_id"`
	CreatedAt   sql.NullString `json:"created_at"`
	ExpiresAt   sql.NullString `json:"expires_at"`
	LastAccess  sql.NullString `json:"last_access"`
	IpAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	SessionData sql.NullString `json:"session_data"`
	SessionID   string         `json:"session_id"`
}

type SessionsHistoryEntry struct {
	SessionID   int64          `json:"session_id"`
	UserID      int64          `json:"user_id"`
	CreatedAt   sql.NullString `json:"created_at"`
	ExpiresAt   sql.NullString `json:"expires_at"`
	LastAccess  sql.NullString `json:"last_access"`
	IpAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	SessionData sql.NullString `json:"session_data"`
}

type CreateSessionFormParams struct {
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at"`
	LastAccess  string `json:"last_access"`
	IpAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	SessionData string `json:"session_data"`
}

type UpdateSessionFormParams struct {
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at"`
	LastAccess  string `json:"last_access"`
	IpAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	SessionData string `json:"session_data"`
	SessionID   string `json:"session_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateSessionParams(a CreateSessionFormParams) CreateSessionParams {
	return CreateSessionParams{
		UserID:      Si(a.UserID),
		CreatedAt:   Ns(a.CreatedAt),
		ExpiresAt:   Ns(a.ExpiresAt),
		LastAccess:  Ns(a.LastAccess),
		IpAddress:   Ns(a.IpAddress),
		UserAgent:   Ns(a.UserAgent),
		SessionData: Ns(a.SessionData),
	}
}

func MapUpdateSessionParams(a UpdateSessionFormParams) UpdateSessionParams {
	return UpdateSessionParams{
		UserID:      Si(a.UserID),
		CreatedAt:   Ns(a.CreatedAt),
		ExpiresAt:   Ns(a.ExpiresAt),
		LastAccess:  Ns(a.LastAccess),
		IpAddress:   Ns(a.IpAddress),
		UserAgent:   Ns(a.UserAgent),
		SessionData: Ns(a.SessionData),
		SessionID:   a.SessionID,
	}
}

func MapStringSession(a Sessions) StringSessions {
	return StringSessions{
		SessionID:   strconv.FormatInt(a.SessionID, 10),
		UserID:      strconv.FormatInt(a.UserID, 10),
		CreatedAt:   a.CreatedAt.String,
		ExpiresAt:   a.ExpiresAt.String,
		LastAccess:  a.LastAccess.String,
		IpAddress:   a.IpAddress.String,
		UserAgent:   a.UserAgent.String,
		SessionData: a.SessionData.String,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
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
		SessionID:   Si(a.SessionID),
	}
}

// /QUERIES
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

func (d Database) DeleteSession(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteSession(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}
	return nil
}

func (d Database) GetSession(id int64) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetSession(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

func (d Database) GetSessionsByUserId(userID int64) (*[]Sessions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, userID)
	if err != nil {
		return nil, err
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
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
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapSession(a mdbm.Sessions) Sessions {
	return Sessions{
		SessionID:   int64(a.SessionID),
		UserID:      int64(a.UserID),
		CreatedAt:   Ns(a.CreatedAt.String()),
		ExpiresAt:   Ns(a.ExpiresAt.String()),
		LastAccess:  Ns(a.LastAccess.String()),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d MysqlDatabase) MapCreateSessionParams(a CreateSessionParams) mdbm.CreateSessionParams {
	return mdbm.CreateSessionParams{
		UserID:      int32(a.UserID),
		CreatedAt:   StringToNTime(a.CreatedAt.String).Time,
		ExpiresAt:   StringToNTime(a.ExpiresAt.String).Time,
		LastAccess:  StringToNTime(a.LastAccess.String).Time,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d MysqlDatabase) MapUpdateSessionParams(a UpdateSessionParams) mdbm.UpdateSessionParams {
	return mdbm.UpdateSessionParams{
		UserID:      int32(a.UserID),
		CreatedAt:   StringToNTime(a.CreatedAt.String).Time,
		ExpiresAt:   StringToNTime(a.ExpiresAt.String).Time,
		LastAccess:  StringToNTime(a.LastAccess.String).Time,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
		SessionID:   int32(Si(a.SessionID)),
	}
}

// /QUERIES
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
	row, err := queries.GetLastSession(d.Context)
	if err != nil {
		return nil, fmt.Errorf("Failed to get last inserted Session: %v\n", err)
	}
	session := d.MapSession(row)
	return &session, nil
}

func (d MysqlDatabase) DeleteSession(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteSession(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetSession(id int64) (*Sessions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetSession(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

func (d MysqlDatabase) GetSessionsByUserId(userID int64) (*[]Sessions, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, int32(userID))
	if err != nil {
		return nil, err
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
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
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapSession(a mdbp.Sessions) Sessions {
	return Sessions{
		SessionID:   int64(a.SessionID),
		UserID:      int64(a.UserID),
		CreatedAt:   Ns(Nt(a.CreatedAt)),
		ExpiresAt:   Ns(Nt(a.ExpiresAt)),
		LastAccess:  Ns(Nt(a.LastAccess)),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d PsqlDatabase) MapCreateSessionParams(a CreateSessionParams) mdbp.CreateSessionParams {
	return mdbp.CreateSessionParams{
		UserID:      int32(a.UserID),
		CreatedAt:   StringToNTime(a.CreatedAt.String),
		ExpiresAt:   StringToNTime(a.ExpiresAt.String),
		LastAccess:  StringToNTime(a.LastAccess.String),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d PsqlDatabase) MapUpdateSessionParams(a UpdateSessionParams) mdbp.UpdateSessionParams {
	return mdbp.UpdateSessionParams{
		UserID:      int32(a.UserID),
		CreatedAt:   StringToNTime(a.CreatedAt.String),
		ExpiresAt:   StringToNTime(a.ExpiresAt.String),
		LastAccess:  StringToNTime(a.LastAccess.String),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
		SessionID:   int32(Si(a.SessionID)),
	}
}

// /QUERIES
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

func (d PsqlDatabase) DeleteSession(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteSession(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Session: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetSession(id int64) (*Sessions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetSession(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

func (d PsqlDatabase) GetSessionsByUserId(userID int64) (*[]Sessions, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetSessionByUserId(d.Context, int32(userID))
	if err != nil {
		return nil, err
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
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
