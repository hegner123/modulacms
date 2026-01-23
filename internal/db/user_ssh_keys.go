package db

import (
	"database/sql"
	"fmt"
	"time"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	"github.com/hegner123/modulacms/internal/db/types"
)

// UserSshKeys represents an SSH public key for a user
type UserSshKeys struct {
	SshKeyID    int64
	UserID      types.NullableUserID
	PublicKey   string
	KeyType     string
	Fingerprint string
	Label       string
	DateCreated types.Timestamp
	LastUsed    string
}

// CreateUserSshKeyParams contains parameters for creating a new SSH key
type CreateUserSshKeyParams struct {
	UserID      types.NullableUserID
	PublicKey   string
	KeyType     string
	Fingerprint string
	Label       string
	DateCreated types.Timestamp
}

// ============================================================================
// SQLite Implementation
// ============================================================================

func (d Database) CreateUserSshKey(params CreateUserSshKeyParams) (*UserSshKeys, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.CreateUserSshKey(d.Context, mdb.CreateUserSshKeyParams{
		UserID:      params.UserID,
		PublicKey:   params.PublicKey,
		KeyType:     params.KeyType,
		Fingerprint: params.Fingerprint,
		Label:       sql.NullString{String: params.Label, Valid: params.Label != ""},
		DateCreated: params.DateCreated,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %v", err)
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d Database) GetUserSshKey(id int64) (*UserSshKeys, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserSshKey(d.Context, mdb.GetUserSshKeyParams{SSHKeyID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d Database) GetUserSshKeyByFingerprint(fingerprint string) (*UserSshKeys, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserSshKeyByFingerprint(d.Context, mdb.GetUserSshKeyByFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d Database) GetUserBySSHFingerprint(fingerprint string) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserBySSHFingerprint(d.Context, mdb.GetUserBySSHFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d Database) ListUserSshKeys(userID types.NullableUserID) (*[]UserSshKeys, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListUserSshKeys(d.Context, mdb.ListUserSshKeysParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %v", err)
	}
	res := []UserSshKeys{}
	for _, v := range rows {
		m := d.MapUserSshKeys(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateUserSshKeyLastUsed(id int64, lastUsed string) error {
	queries := mdb.New(d.Connection)
	err := queries.UpdateUserSshKeyLastUsed(d.Context, mdb.UpdateUserSshKeyLastUsedParams{
		LastUsed: sql.NullString{String: lastUsed, Valid: lastUsed != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key last used: %v", err)
	}
	return nil
}

func (d Database) UpdateUserSshKeyLabel(id int64, label string) error {
	queries := mdb.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdb.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}

func (d Database) DeleteUserSshKey(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUserSshKey(d.Context, mdb.DeleteUserSshKeyParams{SSHKeyID: id})
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %v", err)
	}
	return nil
}

func (d Database) MapUserSshKeys(row mdb.UserSshKeys) UserSshKeys {
	label := ""
	if row.Label.Valid {
		label = row.Label.String
	}
	lastUsed := ""
	if row.LastUsed.Valid {
		lastUsed = row.LastUsed.String
	}
	return UserSshKeys{
		SshKeyID:    row.SSHKeyID,
		UserID:      row.UserID,
		PublicKey:   row.PublicKey,
		KeyType:     row.KeyType,
		Fingerprint: row.Fingerprint,
		Label:       label,
		DateCreated: row.DateCreated,
		LastUsed:    lastUsed,
	}
}

func (d Database) CreateUserSshKeyTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserSshKeyTable(d.Context)
	return err
}

func (d Database) CountUserSshKeys() (*int64, error) {
	queries := mdb.New(d.Connection)
	count, err := queries.CountUserSshKeys(d.Context)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// ============================================================================
// MySQL Implementation
// ============================================================================

func (d MysqlDatabase) CreateUserSshKey(params CreateUserSshKeyParams) (*UserSshKeys, error) {
	queries := mdbm.New(d.Connection)

	result, err := queries.CreateUserSshKey(d.Context, mdbm.CreateUserSshKeyParams{
		UserID:      params.UserID,
		PublicKey:   params.PublicKey,
		KeyType:     params.KeyType,
		Fingerprint: params.Fingerprint,
		Label:       sql.NullString{String: params.Label, Valid: params.Label != ""},
		DateCreated: params.DateCreated,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return d.GetUserSshKey(id)
}

func (d MysqlDatabase) GetUserSshKey(id int64) (*UserSshKeys, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserSshKey(d.Context, mdbm.GetUserSshKeyParams{SSHKeyID: int32(id)})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d MysqlDatabase) GetUserSshKeyByFingerprint(fingerprint string) (*UserSshKeys, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserSshKeyByFingerprint(d.Context, mdbm.GetUserSshKeyByFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d MysqlDatabase) GetUserBySSHFingerprint(fingerprint string) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserBySSHFingerprint(d.Context, mdbm.GetUserBySSHFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d MysqlDatabase) ListUserSshKeys(userID types.NullableUserID) (*[]UserSshKeys, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListUserSshKeys(d.Context, mdbm.ListUserSshKeysParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %v", err)
	}
	res := []UserSshKeys{}
	for _, v := range rows {
		m := d.MapUserSshKeys(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateUserSshKeyLastUsed(id int64, lastUsed string) error {
	queries := mdbm.New(d.Connection)

	// Parse lastUsed string to time.Time for sql.NullTime
	var nullTime sql.NullTime
	if lastUsed != "" {
		t, err := time.Parse(time.RFC3339, lastUsed)
		if err != nil {
			return fmt.Errorf("failed to parse last_used: %v", err)
		}
		nullTime = sql.NullTime{Time: t, Valid: true}
	}

	err := queries.UpdateUserSshKeyLastUsed(d.Context, mdbm.UpdateUserSshKeyLastUsedParams{
		LastUsed: nullTime,
		SSHKeyID: int32(id),
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key last used: %v", err)
	}
	return nil
}

func (d MysqlDatabase) UpdateUserSshKeyLabel(id int64, label string) error {
	queries := mdbm.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdbm.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: int32(id),
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}

func (d MysqlDatabase) DeleteUserSshKey(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUserSshKey(d.Context, mdbm.DeleteUserSshKeyParams{SSHKeyID: int32(id)})
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %v", err)
	}
	return nil
}

func (d MysqlDatabase) MapUserSshKeys(row mdbm.UserSshKeys) UserSshKeys {
	label := ""
	if row.Label.Valid {
		label = row.Label.String
	}
	lastUsed := ""
	if row.LastUsed.Valid {
		lastUsed = row.LastUsed.Time.Format(time.RFC3339)
	}
	return UserSshKeys{
		SshKeyID:    int64(row.SSHKeyID),
		UserID:      row.UserID,
		PublicKey:   row.PublicKey,
		KeyType:     row.KeyType,
		Fingerprint: row.Fingerprint,
		Label:       label,
		DateCreated: row.DateCreated,
		LastUsed:    lastUsed,
	}
}

func (d MysqlDatabase) CreateUserSshKeyTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserSshKeyTable(d.Context)
	return err
}

func (d MysqlDatabase) CountUserSshKeys() (*int64, error) {
	queries := mdbm.New(d.Connection)
	count, err := queries.CountUserSshKeys(d.Context)
	if err != nil {
		return nil, err
	}
	countInt64 := int64(count)
	return &countInt64, nil
}

// ============================================================================
// PostgreSQL Implementation
// ============================================================================

func (d PsqlDatabase) CreateUserSshKey(params CreateUserSshKeyParams) (*UserSshKeys, error) {
	queries := mdbp.New(d.Connection)

	row, err := queries.CreateUserSshKey(d.Context, mdbp.CreateUserSshKeyParams{
		UserID:      params.UserID,
		PublicKey:   params.PublicKey,
		KeyType:     params.KeyType,
		Fingerprint: params.Fingerprint,
		Label:       sql.NullString{String: params.Label, Valid: params.Label != ""},
		DateCreated: params.DateCreated,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %v", err)
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d PsqlDatabase) GetUserSshKey(id int64) (*UserSshKeys, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserSshKey(d.Context, mdbp.GetUserSshKeyParams{SSHKeyID: int32(id)})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d PsqlDatabase) GetUserSshKeyByFingerprint(fingerprint string) (*UserSshKeys, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserSshKeyByFingerprint(d.Context, mdbp.GetUserSshKeyByFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

func (d PsqlDatabase) GetUserBySSHFingerprint(fingerprint string) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserBySSHFingerprint(d.Context, mdbp.GetUserBySSHFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d PsqlDatabase) ListUserSshKeys(userID types.NullableUserID) (*[]UserSshKeys, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListUserSshKeys(d.Context, mdbp.ListUserSshKeysParams{UserID: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to list SSH keys: %v", err)
	}
	res := []UserSshKeys{}
	for _, v := range rows {
		m := d.MapUserSshKeys(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateUserSshKeyLastUsed(id int64, lastUsed string) error {
	queries := mdbp.New(d.Connection)

	// Parse lastUsed string to time.Time for sql.NullTime
	var nullTime sql.NullTime
	if lastUsed != "" {
		t, err := time.Parse(time.RFC3339, lastUsed)
		if err != nil {
			return fmt.Errorf("failed to parse last_used: %v", err)
		}
		nullTime = sql.NullTime{Time: t, Valid: true}
	}

	err := queries.UpdateUserSshKeyLastUsed(d.Context, mdbp.UpdateUserSshKeyLastUsedParams{
		LastUsed: nullTime,
		SSHKeyID: int32(id),
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key last used: %v", err)
	}
	return nil
}

func (d PsqlDatabase) UpdateUserSshKeyLabel(id int64, label string) error {
	queries := mdbp.New(d.Connection)
	err := queries.UpdateUserSshKeyLabel(d.Context, mdbp.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: label, Valid: label != ""},
		SSHKeyID: int32(id),
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key label: %v", err)
	}
	return nil
}

func (d PsqlDatabase) DeleteUserSshKey(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUserSshKey(d.Context, mdbp.DeleteUserSshKeyParams{SSHKeyID: int32(id)})
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %v", err)
	}
	return nil
}

func (d PsqlDatabase) MapUserSshKeys(row mdbp.UserSshKeys) UserSshKeys {
	label := ""
	if row.Label.Valid {
		label = row.Label.String
	}
	lastUsed := ""
	if row.LastUsed.Valid {
		lastUsed = row.LastUsed.Time.Format(time.RFC3339)
	}
	return UserSshKeys{
		SshKeyID:    int64(row.SSHKeyID),
		UserID:      row.UserID,
		PublicKey:   row.PublicKey,
		KeyType:     row.KeyType,
		Fingerprint: row.Fingerprint,
		Label:       label,
		DateCreated: row.DateCreated,
		LastUsed:    lastUsed,
	}
}

func (d PsqlDatabase) CreateUserSshKeyTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserSshKeyTable(d.Context)
	return err
}

func (d PsqlDatabase) CountUserSshKeys() (*int64, error) {
	queries := mdbp.New(d.Connection)
	count, err := queries.CountUserSshKeys(d.Context)
	if err != nil {
		return nil, err
	}
	return &count, nil
}
