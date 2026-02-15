package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// UserSshKeys represents an SSH public key for a user
type UserSshKeys struct {
	SshKeyID    string
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

// CreateUserSshKey creates a new SSH key with audit trail.
func (d Database) CreateUserSshKey(ctx context.Context, ac audited.AuditContext, params CreateUserSshKeyParams) (*UserSshKeys, error) {
	cmd := d.NewUserSshKeyCmd(ctx, ac, params)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create userSshKey: %w", err)
	}
	r := d.MapUserSshKeys(result)
	return &r, nil
}

// GetUserSshKey retrieves an SSH key by ID.
func (d Database) GetUserSshKey(id string) (*UserSshKeys, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserSshKey(d.Context, mdb.GetUserSshKeyParams{SSHKeyID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

// GetUserSshKeyByFingerprint retrieves an SSH key by its fingerprint.
func (d Database) GetUserSshKeyByFingerprint(fingerprint string) (*UserSshKeys, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserSshKeyByFingerprint(d.Context, mdb.GetUserSshKeyByFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

// GetUserBySSHFingerprint retrieves a user by their SSH key fingerprint.
func (d Database) GetUserBySSHFingerprint(fingerprint string) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserBySSHFingerprint(d.Context, mdb.GetUserBySSHFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUserSshKeys lists all SSH keys for a user.
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

// UpdateUserSshKeyLastUsed updates the last used timestamp for an SSH key.
func (d Database) UpdateUserSshKeyLastUsed(id string, lastUsed string) error {
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

// UpdateUserSshKeyLabel updates an SSH key's label with audit trail.
func (d Database) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {
	cmd := d.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)
	return audited.Update(cmd)
}

// DeleteUserSshKey deletes an SSH key with audit trail.
func (d Database) DeleteUserSshKey(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteUserSshKeyCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// MapUserSshKeys converts a sqlc-generated UserSshKeys to the wrapper type.
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

// CreateUserSshKeyTable creates the user_ssh_keys table.
func (d Database) CreateUserSshKeyTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserSshKeyTable(d.Context)
	return err
}

// CountUserSshKeys returns the total count of SSH keys.
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

// CreateUserSshKey creates a new SSH key with audit trail.
func (d MysqlDatabase) CreateUserSshKey(ctx context.Context, ac audited.AuditContext, params CreateUserSshKeyParams) (*UserSshKeys, error) {
	cmd := d.NewUserSshKeyCmd(ctx, ac, params)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create userSshKey: %w", err)
	}
	r := d.MapUserSshKeys(result)
	return &r, nil
}

// GetUserSshKey retrieves an SSH key by ID.
func (d MysqlDatabase) GetUserSshKey(id string) (*UserSshKeys, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserSshKey(d.Context, mdbm.GetUserSshKeyParams{SSHKeyID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

// GetUserSshKeyByFingerprint retrieves an SSH key by its fingerprint.
func (d MysqlDatabase) GetUserSshKeyByFingerprint(fingerprint string) (*UserSshKeys, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserSshKeyByFingerprint(d.Context, mdbm.GetUserSshKeyByFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

// GetUserBySSHFingerprint retrieves a user by their SSH key fingerprint.
func (d MysqlDatabase) GetUserBySSHFingerprint(fingerprint string) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserBySSHFingerprint(d.Context, mdbm.GetUserBySSHFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUserSshKeys lists all SSH keys for a user.
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

// UpdateUserSshKeyLastUsed updates the last used timestamp for an SSH key.
func (d MysqlDatabase) UpdateUserSshKeyLastUsed(id string, lastUsed string) error {
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
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key last used: %v", err)
	}
	return nil
}

// UpdateUserSshKeyLabel updates an SSH key's label with audit trail.
func (d MysqlDatabase) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {
	cmd := d.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)
	return audited.Update(cmd)
}

// DeleteUserSshKey deletes an SSH key with audit trail.
func (d MysqlDatabase) DeleteUserSshKey(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteUserSshKeyCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// MapUserSshKeys converts a sqlc-generated UserSshKeys to the wrapper type.
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

// CreateUserSshKeyTable creates the user_ssh_keys table.
func (d MysqlDatabase) CreateUserSshKeyTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserSshKeyTable(d.Context)
	return err
}

// CountUserSshKeys returns the total count of SSH keys.
func (d MysqlDatabase) CountUserSshKeys() (*int64, error) {
	queries := mdbm.New(d.Connection)
	count, err := queries.CountUserSshKeys(d.Context)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// ============================================================================
// PostgreSQL Implementation
// ============================================================================

// CreateUserSshKey creates a new SSH key with audit trail.
func (d PsqlDatabase) CreateUserSshKey(ctx context.Context, ac audited.AuditContext, params CreateUserSshKeyParams) (*UserSshKeys, error) {
	cmd := d.NewUserSshKeyCmd(ctx, ac, params)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create userSshKey: %w", err)
	}
	r := d.MapUserSshKeys(result)
	return &r, nil
}

// GetUserSshKey retrieves an SSH key by ID.
func (d PsqlDatabase) GetUserSshKey(id string) (*UserSshKeys, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserSshKey(d.Context, mdbp.GetUserSshKeyParams{SSHKeyID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

// GetUserSshKeyByFingerprint retrieves an SSH key by its fingerprint.
func (d PsqlDatabase) GetUserSshKeyByFingerprint(fingerprint string) (*UserSshKeys, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserSshKeyByFingerprint(d.Context, mdbp.GetUserSshKeyByFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUserSshKeys(row)
	return &res, nil
}

// GetUserBySSHFingerprint retrieves a user by their SSH key fingerprint.
func (d PsqlDatabase) GetUserBySSHFingerprint(fingerprint string) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserBySSHFingerprint(d.Context, mdbp.GetUserBySSHFingerprintParams{Fingerprint: fingerprint})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUserSshKeys lists all SSH keys for a user.
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

// UpdateUserSshKeyLastUsed updates the last used timestamp for an SSH key.
func (d PsqlDatabase) UpdateUserSshKeyLastUsed(id string, lastUsed string) error {
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
		SSHKeyID: id,
	})
	if err != nil {
		return fmt.Errorf("failed to update SSH key last used: %v", err)
	}
	return nil
}

// UpdateUserSshKeyLabel updates an SSH key's label with audit trail.
func (d PsqlDatabase) UpdateUserSshKeyLabel(ctx context.Context, ac audited.AuditContext, id string, label string) error {
	cmd := d.UpdateUserSshKeyLabelCmd(ctx, ac, id, label)
	return audited.Update(cmd)
}

// DeleteUserSshKey deletes an SSH key with audit trail.
func (d PsqlDatabase) DeleteUserSshKey(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteUserSshKeyCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// MapUserSshKeys converts a sqlc-generated UserSshKeys to the wrapper type.
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

// CreateUserSshKeyTable creates the user_ssh_keys table.
func (d PsqlDatabase) CreateUserSshKeyTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserSshKeyTable(d.Context)
	return err
}

// CountUserSshKeys returns the total count of SSH keys.
func (d PsqlDatabase) CountUserSshKeys() (*int64, error) {
	queries := mdbp.New(d.Connection)
	count, err := queries.CountUserSshKeys(d.Context)
	if err != nil {
		return nil, err
	}
	return &count, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

// NewUserSshKeyCmd is an audited command for creating a user SSH key.
type NewUserSshKeyCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserSshKeyParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserSshKeyCmd) Context() context.Context              { return c.ctx }
func (c NewUserSshKeyCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserSshKeyCmd) Connection() *sql.DB                   { return c.conn }
func (c NewUserSshKeyCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserSshKeyCmd) TableName() string                     { return "user_ssh_keys" }
func (c NewUserSshKeyCmd) Params() any                           { return c.params }
func (c NewUserSshKeyCmd) GetID(u mdb.UserSshKeys) string        { return u.SSHKeyID }

func (c NewUserSshKeyCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.UserSshKeys, error) {
	queries := mdb.New(tx)
	return queries.CreateUserSshKey(ctx, mdb.CreateUserSshKeyParams{
		SSHKeyID:    string(types.NewUserSshKeyID()),
		UserID:      c.params.UserID,
		PublicKey:   c.params.PublicKey,
		KeyType:     c.params.KeyType,
		Fingerprint: c.params.Fingerprint,
		Label:       sql.NullString{String: c.params.Label, Valid: c.params.Label != ""},
		DateCreated: c.params.DateCreated,
	})
}

// NewUserSshKeyCmd creates a new SSH key creation command.
func (d Database) NewUserSshKeyCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserSshKeyParams) NewUserSshKeyCmd {
	return NewUserSshKeyCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteUserSshKeyCmd is an audited command for deleting a user SSH key.
type DeleteUserSshKeyCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserSshKeyCmd) Context() context.Context              { return c.ctx }
func (c DeleteUserSshKeyCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserSshKeyCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserSshKeyCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserSshKeyCmd) TableName() string                     { return "user_ssh_keys" }
func (c DeleteUserSshKeyCmd) GetID() string                         { return c.id }

func (c DeleteUserSshKeyCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.UserSshKeys, error) {
	queries := mdb.New(tx)
	return queries.GetUserSshKey(ctx, mdb.GetUserSshKeyParams{SSHKeyID: c.id})
}

func (c DeleteUserSshKeyCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteUserSshKey(ctx, mdb.DeleteUserSshKeyParams{SSHKeyID: c.id})
}

// DeleteUserSshKeyCmd creates a new SSH key deletion command.
func (d Database) DeleteUserSshKeyCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteUserSshKeyCmd {
	return DeleteUserSshKeyCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE LABEL -----

// UpdateUserSshKeyLabelCmd is an audited command for updating a user SSH key's label.
type UpdateUserSshKeyLabelCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	label    string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserSshKeyLabelCmd) Context() context.Context              { return c.ctx }
func (c UpdateUserSshKeyLabelCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserSshKeyLabelCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserSshKeyLabelCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserSshKeyLabelCmd) TableName() string                     { return "user_ssh_keys" }
func (c UpdateUserSshKeyLabelCmd) Params() any {
	return map[string]any{"id": c.id, "label": c.label}
}
func (c UpdateUserSshKeyLabelCmd) GetID() string { return c.id }

func (c UpdateUserSshKeyLabelCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.UserSshKeys, error) {
	queries := mdb.New(tx)
	return queries.GetUserSshKey(ctx, mdb.GetUserSshKeyParams{SSHKeyID: c.id})
}

func (c UpdateUserSshKeyLabelCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateUserSshKeyLabel(ctx, mdb.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: c.label, Valid: c.label != ""},
		SSHKeyID: c.id,
	})
}

// UpdateUserSshKeyLabelCmd creates a new SSH key label update command.
func (d Database) UpdateUserSshKeyLabelCmd(ctx context.Context, auditCtx audited.AuditContext, id string, label string) UpdateUserSshKeyLabelCmd {
	return UpdateUserSshKeyLabelCmd{ctx: ctx, auditCtx: auditCtx, id: id, label: label, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewUserSshKeyCmdMysql is an audited command for creating a user SSH key on MySQL.
type NewUserSshKeyCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserSshKeyParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserSshKeyCmdMysql) Context() context.Context              { return c.ctx }
func (c NewUserSshKeyCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserSshKeyCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewUserSshKeyCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserSshKeyCmdMysql) TableName() string                     { return "user_ssh_keys" }
func (c NewUserSshKeyCmdMysql) Params() any                           { return c.params }
func (c NewUserSshKeyCmdMysql) GetID(u mdbm.UserSshKeys) string      { return u.SSHKeyID }

func (c NewUserSshKeyCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.UserSshKeys, error) {
	queries := mdbm.New(tx)
	sshKeyID := string(types.NewUserSshKeyID())
	_, err := queries.CreateUserSshKey(ctx, mdbm.CreateUserSshKeyParams{
		SSHKeyID:    sshKeyID,
		UserID:      c.params.UserID,
		PublicKey:   c.params.PublicKey,
		KeyType:     c.params.KeyType,
		Fingerprint: c.params.Fingerprint,
		Label:       sql.NullString{String: c.params.Label, Valid: c.params.Label != ""},
		DateCreated: c.params.DateCreated,
	})
	if err != nil {
		return mdbm.UserSshKeys{}, err
	}
	return queries.GetUserSshKey(ctx, mdbm.GetUserSshKeyParams{SSHKeyID: sshKeyID})
}

// NewUserSshKeyCmd creates a new SSH key creation command for MySQL.
func (d MysqlDatabase) NewUserSshKeyCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserSshKeyParams) NewUserSshKeyCmdMysql {
	return NewUserSshKeyCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteUserSshKeyCmdMysql is an audited command for deleting a user SSH key on MySQL.
type DeleteUserSshKeyCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserSshKeyCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteUserSshKeyCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserSshKeyCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserSshKeyCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserSshKeyCmdMysql) TableName() string                     { return "user_ssh_keys" }
func (c DeleteUserSshKeyCmdMysql) GetID() string                         { return c.id }

func (c DeleteUserSshKeyCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.UserSshKeys, error) {
	queries := mdbm.New(tx)
	return queries.GetUserSshKey(ctx, mdbm.GetUserSshKeyParams{SSHKeyID: c.id})
}

func (c DeleteUserSshKeyCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteUserSshKey(ctx, mdbm.DeleteUserSshKeyParams{SSHKeyID: c.id})
}

// DeleteUserSshKeyCmd creates a new SSH key deletion command for MySQL.
func (d MysqlDatabase) DeleteUserSshKeyCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteUserSshKeyCmdMysql {
	return DeleteUserSshKeyCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE LABEL -----

// UpdateUserSshKeyLabelCmdMysql is an audited command for updating a user SSH key's label on MySQL.
type UpdateUserSshKeyLabelCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	label    string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserSshKeyLabelCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateUserSshKeyLabelCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserSshKeyLabelCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserSshKeyLabelCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserSshKeyLabelCmdMysql) TableName() string                     { return "user_ssh_keys" }
func (c UpdateUserSshKeyLabelCmdMysql) Params() any {
	return map[string]any{"id": c.id, "label": c.label}
}
func (c UpdateUserSshKeyLabelCmdMysql) GetID() string { return c.id }

func (c UpdateUserSshKeyLabelCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.UserSshKeys, error) {
	queries := mdbm.New(tx)
	return queries.GetUserSshKey(ctx, mdbm.GetUserSshKeyParams{SSHKeyID: c.id})
}

func (c UpdateUserSshKeyLabelCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateUserSshKeyLabel(ctx, mdbm.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: c.label, Valid: c.label != ""},
		SSHKeyID: c.id,
	})
}

// UpdateUserSshKeyLabelCmd creates a new SSH key label update command for MySQL.
func (d MysqlDatabase) UpdateUserSshKeyLabelCmd(ctx context.Context, auditCtx audited.AuditContext, id string, label string) UpdateUserSshKeyLabelCmdMysql {
	return UpdateUserSshKeyLabelCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, label: label, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewUserSshKeyCmdPsql is an audited command for creating a user SSH key on PostgreSQL.
type NewUserSshKeyCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserSshKeyParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserSshKeyCmdPsql) Context() context.Context              { return c.ctx }
func (c NewUserSshKeyCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserSshKeyCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewUserSshKeyCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserSshKeyCmdPsql) TableName() string                     { return "user_ssh_keys" }
func (c NewUserSshKeyCmdPsql) Params() any                           { return c.params }
func (c NewUserSshKeyCmdPsql) GetID(u mdbp.UserSshKeys) string      { return u.SSHKeyID }

func (c NewUserSshKeyCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.UserSshKeys, error) {
	queries := mdbp.New(tx)
	return queries.CreateUserSshKey(ctx, mdbp.CreateUserSshKeyParams{
		SSHKeyID:    string(types.NewUserSshKeyID()),
		UserID:      c.params.UserID,
		PublicKey:   c.params.PublicKey,
		KeyType:     c.params.KeyType,
		Fingerprint: c.params.Fingerprint,
		Label:       sql.NullString{String: c.params.Label, Valid: c.params.Label != ""},
		DateCreated: c.params.DateCreated,
	})
}

// NewUserSshKeyCmd creates a new SSH key creation command for PostgreSQL.
func (d PsqlDatabase) NewUserSshKeyCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserSshKeyParams) NewUserSshKeyCmdPsql {
	return NewUserSshKeyCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteUserSshKeyCmdPsql is an audited command for deleting a user SSH key on PostgreSQL.
type DeleteUserSshKeyCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserSshKeyCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteUserSshKeyCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserSshKeyCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserSshKeyCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserSshKeyCmdPsql) TableName() string                     { return "user_ssh_keys" }
func (c DeleteUserSshKeyCmdPsql) GetID() string                         { return c.id }

func (c DeleteUserSshKeyCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.UserSshKeys, error) {
	queries := mdbp.New(tx)
	return queries.GetUserSshKey(ctx, mdbp.GetUserSshKeyParams{SSHKeyID: c.id})
}

func (c DeleteUserSshKeyCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteUserSshKey(ctx, mdbp.DeleteUserSshKeyParams{SSHKeyID: c.id})
}

// DeleteUserSshKeyCmd creates a new SSH key deletion command for PostgreSQL.
func (d PsqlDatabase) DeleteUserSshKeyCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteUserSshKeyCmdPsql {
	return DeleteUserSshKeyCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE LABEL -----

// UpdateUserSshKeyLabelCmdPsql is an audited command for updating a user SSH key's label on PostgreSQL.
type UpdateUserSshKeyLabelCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	label    string
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserSshKeyLabelCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateUserSshKeyLabelCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserSshKeyLabelCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserSshKeyLabelCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserSshKeyLabelCmdPsql) TableName() string                     { return "user_ssh_keys" }
func (c UpdateUserSshKeyLabelCmdPsql) Params() any {
	return map[string]any{"id": c.id, "label": c.label}
}
func (c UpdateUserSshKeyLabelCmdPsql) GetID() string { return c.id }

func (c UpdateUserSshKeyLabelCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.UserSshKeys, error) {
	queries := mdbp.New(tx)
	return queries.GetUserSshKey(ctx, mdbp.GetUserSshKeyParams{SSHKeyID: c.id})
}

func (c UpdateUserSshKeyLabelCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateUserSshKeyLabel(ctx, mdbp.UpdateUserSshKeyLabelParams{
		Label:    sql.NullString{String: c.label, Valid: c.label != ""},
		SSHKeyID: c.id,
	})
}

// UpdateUserSshKeyLabelCmd creates a new SSH key label update command for PostgreSQL.
func (d PsqlDatabase) UpdateUserSshKeyLabelCmd(ctx context.Context, auditCtx audited.AuditContext, id string, label string) UpdateUserSshKeyLabelCmdPsql {
	return UpdateUserSshKeyLabelCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, label: label, conn: d.Connection, recorder: PsqlRecorder}
}
