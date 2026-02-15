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

// Tokens represents an authentication token in the system.
type Tokens struct {
	ID        string               `json:"id"`
	UserID    types.NullableUserID `json:"user_id"`
	TokenType string               `json:"token_type"`
	Token     string               `json:"token"`
	IssuedAt  string               `json:"issued_at"`
	ExpiresAt types.Timestamp      `json:"expires_at"`
	Revoked   bool                 `json:"revoked"`
}

// CreateTokenParams contains the parameters for creating a new token.
type CreateTokenParams struct {
	UserID    types.NullableUserID `json:"user_id"`
	TokenType string               `json:"token_type"`
	Token     string               `json:"token"`
	IssuedAt  string               `json:"issued_at"`
	ExpiresAt types.Timestamp      `json:"expires_at"`
	Revoked   bool                 `json:"revoked"`
}

// UpdateTokenParams contains the parameters for updating an existing token.
type UpdateTokenParams struct {
	Token     string          `json:"token"`
	IssuedAt  string          `json:"issued_at"`
	ExpiresAt types.Timestamp `json:"expires_at"`
	Revoked   bool            `json:"revoked"`
	ID        string          `json:"id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringToken converts Tokens to StringTokens for table display
func MapStringToken(a Tokens) StringTokens {
	return StringTokens{
		ID:        a.ID,
		UserID:    a.UserID.String(),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt.String(),
		Revoked:   fmt.Sprintf("%t", a.Revoked),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

// MapToken converts a sqlc-generated SQLite token to the wrapper type.
func (d Database) MapToken(a mdb.Tokens) Tokens {
	return Tokens{
		ID:        a.ID,
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Token:     a.Tokens,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

// MapCreateTokenParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapCreateTokenParams(a CreateTokenParams) mdb.CreateTokenParams {
	return mdb.CreateTokenParams{
		ID:        string(types.NewTokenID()),
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Tokens:    a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

// MapUpdateTokenParams converts wrapper params to sqlc-generated SQLite params.
func (d Database) MapUpdateTokenParams(a UpdateTokenParams) mdb.UpdateTokenParams {
	return mdb.UpdateTokenParams{
		Tokens:    a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
		ID:        a.ID,
	}
}

// QUERIES

// CountTokens returns the total number of tokens in SQLite.
func (d Database) CountTokens() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateTokenTable creates the tokens table in SQLite.
func (d Database) CreateTokenTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

// CreateToken creates a new token with audit context in SQLite.
func (d Database) CreateToken(ctx context.Context, ac audited.AuditContext, s CreateTokenParams) (*Tokens, error) {
	cmd := d.NewTokenCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	r := d.MapToken(result)
	return &r, nil
}

// DeleteToken deletes a token with audit context in SQLite.
func (d Database) DeleteToken(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTokenCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetToken retrieves a token by ID from SQLite.
func (d Database) GetToken(id string) (*Tokens, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetToken(d.Context, mdb.GetTokenParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

// GetTokenByTokenValue retrieves a token by its token value from SQLite.
func (d Database) GetTokenByTokenValue(tokenValue string) (*Tokens, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTokenByTokenValue(d.Context, mdb.GetTokenByTokenValueParams{Tokens: tokenValue})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

// GetTokenByUserId retrieves all tokens for a user from SQLite.
func (d Database) GetTokenByUserId(userID types.NullableUserID) (*[]Tokens, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetTokenByUserId(d.Context, mdb.GetTokenByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListTokens retrieves all tokens from SQLite.
func (d Database) ListTokens() (*[]Tokens, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tokens: %v\n", err)
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateToken updates a token with audit context in SQLite.
func (d Database) UpdateToken(ctx context.Context, ac audited.AuditContext, s UpdateTokenParams) (*string, error) {
	cmd := d.UpdateTokenCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

// MapToken converts a sqlc-generated MySQL token to the wrapper type.
func (d MysqlDatabase) MapToken(a mdbm.Tokens) Tokens {
	return Tokens{
		ID:        a.ID,
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Token:     a.Tokens,
		IssuedAt:  a.IssuedAt.String(),
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

// MapCreateTokenParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapCreateTokenParams(a CreateTokenParams) mdbm.CreateTokenParams {
	return mdbm.CreateTokenParams{
		ID:        string(types.NewTokenID()),
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Tokens:    a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

// MapUpdateTokenParams converts wrapper params to sqlc-generated MySQL params.
func (d MysqlDatabase) MapUpdateTokenParams(a UpdateTokenParams) mdbm.UpdateTokenParams {
	return mdbm.UpdateTokenParams{
		Tokens:    a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
		ID:        a.ID,
	}
}

// QUERIES

// CountTokens returns the total number of tokens in MySQL.
func (d MysqlDatabase) CountTokens() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateTokenTable creates the tokens table in MySQL.
func (d MysqlDatabase) CreateTokenTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

// CreateToken creates a new token with audit context in MySQL.
func (d MysqlDatabase) CreateToken(ctx context.Context, ac audited.AuditContext, s CreateTokenParams) (*Tokens, error) {
	cmd := d.NewTokenCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	r := d.MapToken(result)
	return &r, nil
}

// DeleteToken deletes a token with audit context in MySQL.
func (d MysqlDatabase) DeleteToken(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTokenCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetToken retrieves a token by ID from MySQL.
func (d MysqlDatabase) GetToken(id string) (*Tokens, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetToken(d.Context, mdbm.GetTokenParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

// GetTokenByTokenValue retrieves a token by its token value from MySQL.
func (d MysqlDatabase) GetTokenByTokenValue(tokenValue string) (*Tokens, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTokenByTokenValue(d.Context, mdbm.GetTokenByTokenValueParams{Tokens: tokenValue})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

// GetTokenByUserId retrieves all tokens for a user from MySQL.
func (d MysqlDatabase) GetTokenByUserId(userID types.NullableUserID) (*[]Tokens, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetTokenByUserId(d.Context, mdbm.GetTokenByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListTokens retrieves all tokens from MySQL.
func (d MysqlDatabase) ListTokens() (*[]Tokens, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tokens: %v\n", err)
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateToken updates a token with audit context in MySQL.
func (d MysqlDatabase) UpdateToken(ctx context.Context, ac audited.AuditContext, s UpdateTokenParams) (*string, error) {
	cmd := d.UpdateTokenCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

// MapToken converts a sqlc-generated PostgreSQL token to the wrapper type.
func (d PsqlDatabase) MapToken(a mdbp.Tokens) Tokens {
	return Tokens{
		ID:        a.ID,
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Token:     a.Tokens,
		IssuedAt:  a.IssuedAt.String(),
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

// MapCreateTokenParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapCreateTokenParams(a CreateTokenParams) mdbp.CreateTokenParams {
	return mdbp.CreateTokenParams{
		ID:        string(types.NewTokenID()),
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Tokens:    a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

// MapUpdateTokenParams converts wrapper params to sqlc-generated PostgreSQL params.
func (d PsqlDatabase) MapUpdateTokenParams(a UpdateTokenParams) mdbp.UpdateTokenParams {
	return mdbp.UpdateTokenParams{
		Tokens:    a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
		ID:        a.ID,
	}
}

// QUERIES

// CountTokens returns the total number of tokens in PostgreSQL.
func (d PsqlDatabase) CountTokens() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateTokenTable creates the tokens table in PostgreSQL.
func (d PsqlDatabase) CreateTokenTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

// CreateToken creates a new token with audit context in PostgreSQL.
func (d PsqlDatabase) CreateToken(ctx context.Context, ac audited.AuditContext, s CreateTokenParams) (*Tokens, error) {
	cmd := d.NewTokenCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	r := d.MapToken(result)
	return &r, nil
}

// DeleteToken deletes a token with audit context in PostgreSQL.
func (d PsqlDatabase) DeleteToken(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTokenCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetToken retrieves a token by ID from PostgreSQL.
func (d PsqlDatabase) GetToken(id string) (*Tokens, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetToken(d.Context, mdbp.GetTokenParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

// GetTokenByTokenValue retrieves a token by its token value from PostgreSQL.
func (d PsqlDatabase) GetTokenByTokenValue(tokenValue string) (*Tokens, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetTokenByTokenValue(d.Context, mdbp.GetTokenByTokenValueParams{Tokens: tokenValue})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

// GetTokenByUserId retrieves all tokens for a user from PostgreSQL.
func (d PsqlDatabase) GetTokenByUserId(userID types.NullableUserID) (*[]Tokens, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetTokenByUserId(d.Context, mdbp.GetTokenByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
}

// ListTokens retrieves all tokens from PostgreSQL.
func (d PsqlDatabase) ListTokens() (*[]Tokens, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tokens: %v\n", err)
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
}

// UpdateToken updates a token with audit context in PostgreSQL.
func (d PsqlDatabase) UpdateToken(ctx context.Context, ac audited.AuditContext, s UpdateTokenParams) (*string, error) {
	cmd := d.UpdateTokenCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &msg, nil
}

///////////////////////////////
// AUDITED COMMAND STRUCTS
//////////////////////////////

// ===== SQLITE =====

// NewTokenCmd implements audited.CreateCommand[mdb.Tokens] for SQLite.
type NewTokenCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTokenParams
	conn     *sql.DB
}

// Context returns the context of the command.
func (c NewTokenCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c NewTokenCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c NewTokenCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c NewTokenCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the table name for the command.
func (c NewTokenCmd) TableName() string                           { return "tokens" }

// Params returns the parameters for the command.
func (c NewTokenCmd) Params() any                                 { return c.params }

// GetID extracts the ID from a token.
func (c NewTokenCmd) GetID(x mdb.Tokens) string {
	return x.ID
}

// Execute creates a new token within a transaction.
func (c NewTokenCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Tokens, error) {
	queries := mdb.New(tx)
	return queries.CreateToken(ctx, mdb.CreateTokenParams{
		ID:        string(types.NewTokenID()),
		UserID:    c.params.UserID,
		TokenType: c.params.TokenType,
		Tokens:    c.params.Token,
		IssuedAt:  c.params.IssuedAt,
		ExpiresAt: c.params.ExpiresAt,
		Revoked:   c.params.Revoked,
	})
}

// NewTokenCmd constructs a new create command for SQLite.
func (d Database) NewTokenCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTokenParams) NewTokenCmd {
	return NewTokenCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTokenCmd implements audited.UpdateCommand[mdb.Tokens] for SQLite.
type UpdateTokenCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTokenParams
	conn     *sql.DB
}

// Context returns the context of the command.
func (c UpdateTokenCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c UpdateTokenCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c UpdateTokenCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c UpdateTokenCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the table name for the command.
func (c UpdateTokenCmd) TableName() string                           { return "tokens" }

// Params returns the parameters for the command.
func (c UpdateTokenCmd) Params() any                                 { return c.params }

// GetID extracts the ID from the parameters.
func (c UpdateTokenCmd) GetID() string                               { return c.params.ID }

// GetBefore retrieves the token before modification.
func (c UpdateTokenCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tokens, error) {
	queries := mdb.New(tx)
	return queries.GetToken(ctx, mdb.GetTokenParams{ID: c.params.ID})
}

// Execute updates the token within a transaction.
func (c UpdateTokenCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateToken(ctx, mdb.UpdateTokenParams{
		Tokens:    c.params.Token,
		IssuedAt:  c.params.IssuedAt,
		ExpiresAt: c.params.ExpiresAt,
		Revoked:   c.params.Revoked,
		ID:        c.params.ID,
	})
}

// UpdateTokenCmd constructs a new update command for SQLite.
func (d Database) UpdateTokenCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTokenParams) UpdateTokenCmd {
	return UpdateTokenCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTokenCmd implements audited.DeleteCommand[mdb.Tokens] for SQLite.
type DeleteTokenCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

// Context returns the context of the command.
func (c DeleteTokenCmd) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c DeleteTokenCmd) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c DeleteTokenCmd) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c DeleteTokenCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }

// TableName returns the table name for the command.
func (c DeleteTokenCmd) TableName() string                           { return "tokens" }

// GetID returns the ID to delete.
func (c DeleteTokenCmd) GetID() string                               { return c.id }

// GetBefore retrieves the token before deletion.
func (c DeleteTokenCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tokens, error) {
	queries := mdb.New(tx)
	return queries.GetToken(ctx, mdb.GetTokenParams{ID: c.id})
}

// Execute deletes the token within a transaction.
func (c DeleteTokenCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteToken(ctx, mdb.DeleteTokenParams{ID: c.id})
}

// DeleteTokenCmd constructs a new delete command for SQLite.
func (d Database) DeleteTokenCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTokenCmd {
	return DeleteTokenCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== MYSQL =====

// NewTokenCmdMysql implements audited.CreateCommand[mdbm.Tokens] for MySQL.
type NewTokenCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTokenParams
	conn     *sql.DB
}

// Context returns the context of the command.
func (c NewTokenCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c NewTokenCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c NewTokenCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c NewTokenCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the table name for the command.
func (c NewTokenCmdMysql) TableName() string                           { return "tokens" }

// Params returns the parameters for the command.
func (c NewTokenCmdMysql) Params() any                                 { return c.params }

// GetID extracts the ID from a token.
func (c NewTokenCmdMysql) GetID(x mdbm.Tokens) string {
	return x.ID
}

// Execute creates a new token within a transaction.
func (c NewTokenCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Tokens, error) {
	id := string(types.NewTokenID())
	queries := mdbm.New(tx)
	err := queries.CreateToken(ctx, mdbm.CreateTokenParams{
		ID:        id,
		UserID:    c.params.UserID,
		TokenType: c.params.TokenType,
		Tokens:    c.params.Token,
		IssuedAt:  StringToNTime(c.params.IssuedAt).Time,
		ExpiresAt: c.params.ExpiresAt,
		Revoked:   c.params.Revoked,
	})
	if err != nil {
		return mdbm.Tokens{}, fmt.Errorf("Failed to CreateToken: %w", err)
	}
	return queries.GetToken(ctx, mdbm.GetTokenParams{ID: id})
}

// NewTokenCmd constructs a new create command for MySQL.
func (d MysqlDatabase) NewTokenCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTokenParams) NewTokenCmdMysql {
	return NewTokenCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTokenCmdMysql implements audited.UpdateCommand[mdbm.Tokens] for MySQL.
type UpdateTokenCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTokenParams
	conn     *sql.DB
}

// Context returns the context of the command.
func (c UpdateTokenCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c UpdateTokenCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c UpdateTokenCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c UpdateTokenCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the table name for the command.
func (c UpdateTokenCmdMysql) TableName() string                           { return "tokens" }

// Params returns the parameters for the command.
func (c UpdateTokenCmdMysql) Params() any                                 { return c.params }

// GetID extracts the ID from the parameters.
func (c UpdateTokenCmdMysql) GetID() string                               { return c.params.ID }

// GetBefore retrieves the token before modification.
func (c UpdateTokenCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tokens, error) {
	queries := mdbm.New(tx)
	return queries.GetToken(ctx, mdbm.GetTokenParams{ID: c.params.ID})
}

// Execute updates the token within a transaction.
func (c UpdateTokenCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateToken(ctx, mdbm.UpdateTokenParams{
		Tokens:    c.params.Token,
		IssuedAt:  StringToNTime(c.params.IssuedAt).Time,
		ExpiresAt: c.params.ExpiresAt,
		Revoked:   c.params.Revoked,
		ID:        c.params.ID,
	})
}

// UpdateTokenCmd constructs a new update command for MySQL.
func (d MysqlDatabase) UpdateTokenCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTokenParams) UpdateTokenCmdMysql {
	return UpdateTokenCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTokenCmdMysql implements audited.DeleteCommand[mdbm.Tokens] for MySQL.
type DeleteTokenCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

// Context returns the context of the command.
func (c DeleteTokenCmdMysql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c DeleteTokenCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c DeleteTokenCmdMysql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c DeleteTokenCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }

// TableName returns the table name for the command.
func (c DeleteTokenCmdMysql) TableName() string                           { return "tokens" }

// GetID returns the ID to delete.
func (c DeleteTokenCmdMysql) GetID() string                               { return c.id }

// GetBefore retrieves the token before deletion.
func (c DeleteTokenCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tokens, error) {
	queries := mdbm.New(tx)
	return queries.GetToken(ctx, mdbm.GetTokenParams{ID: c.id})
}

// Execute deletes the token within a transaction.
func (c DeleteTokenCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteToken(ctx, mdbm.DeleteTokenParams{ID: c.id})
}

// DeleteTokenCmd constructs a new delete command for MySQL.
func (d MysqlDatabase) DeleteTokenCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTokenCmdMysql {
	return DeleteTokenCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}

// ===== POSTGRESQL =====

// NewTokenCmdPsql implements audited.CreateCommand[mdbp.Tokens] for PostgreSQL.
type NewTokenCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateTokenParams
	conn     *sql.DB
}

// Context returns the context of the command.
func (c NewTokenCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c NewTokenCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c NewTokenCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c NewTokenCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the table name for the command.
func (c NewTokenCmdPsql) TableName() string                           { return "tokens" }

// Params returns the parameters for the command.
func (c NewTokenCmdPsql) Params() any                                 { return c.params }

// GetID extracts the ID from a token.
func (c NewTokenCmdPsql) GetID(x mdbp.Tokens) string {
	return x.ID
}

// Execute creates a new token within a transaction.
func (c NewTokenCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Tokens, error) {
	queries := mdbp.New(tx)
	return queries.CreateToken(ctx, mdbp.CreateTokenParams{
		ID:        string(types.NewTokenID()),
		UserID:    c.params.UserID,
		TokenType: c.params.TokenType,
		Tokens:    c.params.Token,
		IssuedAt:  StringToNTime(c.params.IssuedAt).Time,
		ExpiresAt: c.params.ExpiresAt,
		Revoked:   c.params.Revoked,
	})
}

// NewTokenCmd constructs a new create command for PostgreSQL.
func (d PsqlDatabase) NewTokenCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateTokenParams) NewTokenCmdPsql {
	return NewTokenCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// UpdateTokenCmdPsql implements audited.UpdateCommand[mdbp.Tokens] for PostgreSQL.
type UpdateTokenCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateTokenParams
	conn     *sql.DB
}

// Context returns the context of the command.
func (c UpdateTokenCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c UpdateTokenCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c UpdateTokenCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c UpdateTokenCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the table name for the command.
func (c UpdateTokenCmdPsql) TableName() string                           { return "tokens" }

// Params returns the parameters for the command.
func (c UpdateTokenCmdPsql) Params() any                                 { return c.params }

// GetID extracts the ID from the parameters.
func (c UpdateTokenCmdPsql) GetID() string                               { return c.params.ID }

// GetBefore retrieves the token before modification.
func (c UpdateTokenCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tokens, error) {
	queries := mdbp.New(tx)
	return queries.GetToken(ctx, mdbp.GetTokenParams{ID: c.params.ID})
}

// Execute updates the token within a transaction.
func (c UpdateTokenCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateToken(ctx, mdbp.UpdateTokenParams{
		Tokens:    c.params.Token,
		IssuedAt:  StringToNTime(c.params.IssuedAt).Time,
		ExpiresAt: c.params.ExpiresAt,
		Revoked:   c.params.Revoked,
		ID:        c.params.ID,
	})
}

// UpdateTokenCmd constructs a new update command for PostgreSQL.
func (d PsqlDatabase) UpdateTokenCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateTokenParams) UpdateTokenCmdPsql {
	return UpdateTokenCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection}
}

// DeleteTokenCmdPsql implements audited.DeleteCommand[mdbp.Tokens] for PostgreSQL.
type DeleteTokenCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       string
	conn     *sql.DB
}

// Context returns the context of the command.
func (c DeleteTokenCmdPsql) Context() context.Context                    { return c.ctx }

// AuditContext returns the audit context of the command.
func (c DeleteTokenCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }

// Connection returns the database connection of the command.
func (c DeleteTokenCmdPsql) Connection() *sql.DB                         { return c.conn }

// Recorder returns the change event recorder for the command.
func (c DeleteTokenCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }

// TableName returns the table name for the command.
func (c DeleteTokenCmdPsql) TableName() string                           { return "tokens" }

// GetID returns the ID to delete.
func (c DeleteTokenCmdPsql) GetID() string                               { return c.id }

// GetBefore retrieves the token before deletion.
func (c DeleteTokenCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tokens, error) {
	queries := mdbp.New(tx)
	return queries.GetToken(ctx, mdbp.GetTokenParams{ID: c.id})
}

// Execute deletes the token within a transaction.
func (c DeleteTokenCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteToken(ctx, mdbp.DeleteTokenParams{ID: c.id})
}

// DeleteTokenCmd constructs a new delete command for PostgreSQL.
func (d PsqlDatabase) DeleteTokenCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTokenCmdPsql {
	return DeleteTokenCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
