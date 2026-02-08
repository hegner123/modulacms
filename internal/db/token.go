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

type Tokens struct {
	ID        string               `json:"id"`
	UserID    types.NullableUserID `json:"user_id"`
	TokenType string               `json:"token_type"`
	Token     string               `json:"token"`
	IssuedAt  string               `json:"issued_at"`
	ExpiresAt types.Timestamp      `json:"expires_at"`
	Revoked   bool                 `json:"revoked"`
}

type CreateTokenParams struct {
	UserID    types.NullableUserID `json:"user_id"`
	TokenType string               `json:"token_type"`
	Token     string               `json:"token"`
	IssuedAt  string               `json:"issued_at"`
	ExpiresAt types.Timestamp      `json:"expires_at"`
	Revoked   bool                 `json:"revoked"`
}

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

func (d Database) CountTokens() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateTokenTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

func (d Database) CreateToken(ctx context.Context, ac audited.AuditContext, s CreateTokenParams) (*Tokens, error) {
	cmd := d.NewTokenCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	r := d.MapToken(result)
	return &r, nil
}

func (d Database) DeleteToken(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTokenCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetToken(id string) (*Tokens, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetToken(d.Context, mdb.GetTokenParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

func (d Database) GetTokenByTokenValue(tokenValue string) (*Tokens, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTokenByTokenValue(d.Context, mdb.GetTokenByTokenValueParams{Tokens: tokenValue})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

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

func (d MysqlDatabase) CountTokens() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateTokenTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateToken(ctx context.Context, ac audited.AuditContext, s CreateTokenParams) (*Tokens, error) {
	cmd := d.NewTokenCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	r := d.MapToken(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteToken(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTokenCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetToken(id string) (*Tokens, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetToken(d.Context, mdbm.GetTokenParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

func (d MysqlDatabase) GetTokenByTokenValue(tokenValue string) (*Tokens, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTokenByTokenValue(d.Context, mdbm.GetTokenByTokenValueParams{Tokens: tokenValue})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

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

func (d PsqlDatabase) CountTokens() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountToken(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateTokenTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateTokenTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateToken(ctx context.Context, ac audited.AuditContext, s CreateTokenParams) (*Tokens, error) {
	cmd := d.NewTokenCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	r := d.MapToken(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteToken(ctx context.Context, ac audited.AuditContext, id string) error {
	cmd := d.DeleteTokenCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetToken(id string) (*Tokens, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetToken(d.Context, mdbp.GetTokenParams{ID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

func (d PsqlDatabase) GetTokenByTokenValue(tokenValue string) (*Tokens, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetTokenByTokenValue(d.Context, mdbp.GetTokenByTokenValueParams{Tokens: tokenValue})
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

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

func (c NewTokenCmd) Context() context.Context                    { return c.ctx }
func (c NewTokenCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewTokenCmd) Connection() *sql.DB                         { return c.conn }
func (c NewTokenCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c NewTokenCmd) TableName() string                           { return "tokens" }
func (c NewTokenCmd) Params() any                                 { return c.params }

func (c NewTokenCmd) GetID(x mdb.Tokens) string {
	return x.ID
}

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

func (c UpdateTokenCmd) Context() context.Context                    { return c.ctx }
func (c UpdateTokenCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateTokenCmd) Connection() *sql.DB                         { return c.conn }
func (c UpdateTokenCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c UpdateTokenCmd) TableName() string                           { return "tokens" }
func (c UpdateTokenCmd) Params() any                                 { return c.params }
func (c UpdateTokenCmd) GetID() string                               { return c.params.ID }

func (c UpdateTokenCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tokens, error) {
	queries := mdb.New(tx)
	return queries.GetToken(ctx, mdb.GetTokenParams{ID: c.params.ID})
}

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

func (c DeleteTokenCmd) Context() context.Context                    { return c.ctx }
func (c DeleteTokenCmd) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteTokenCmd) Connection() *sql.DB                         { return c.conn }
func (c DeleteTokenCmd) Recorder() audited.ChangeEventRecorder       { return SQLiteRecorder }
func (c DeleteTokenCmd) TableName() string                           { return "tokens" }
func (c DeleteTokenCmd) GetID() string                               { return c.id }

func (c DeleteTokenCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Tokens, error) {
	queries := mdb.New(tx)
	return queries.GetToken(ctx, mdb.GetTokenParams{ID: c.id})
}

func (c DeleteTokenCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteToken(ctx, mdb.DeleteTokenParams{ID: c.id})
}

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

func (c NewTokenCmdMysql) Context() context.Context                    { return c.ctx }
func (c NewTokenCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewTokenCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c NewTokenCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c NewTokenCmdMysql) TableName() string                           { return "tokens" }
func (c NewTokenCmdMysql) Params() any                                 { return c.params }

func (c NewTokenCmdMysql) GetID(x mdbm.Tokens) string {
	return x.ID
}

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

func (c UpdateTokenCmdMysql) Context() context.Context                    { return c.ctx }
func (c UpdateTokenCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateTokenCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c UpdateTokenCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c UpdateTokenCmdMysql) TableName() string                           { return "tokens" }
func (c UpdateTokenCmdMysql) Params() any                                 { return c.params }
func (c UpdateTokenCmdMysql) GetID() string                               { return c.params.ID }

func (c UpdateTokenCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tokens, error) {
	queries := mdbm.New(tx)
	return queries.GetToken(ctx, mdbm.GetTokenParams{ID: c.params.ID})
}

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

func (c DeleteTokenCmdMysql) Context() context.Context                    { return c.ctx }
func (c DeleteTokenCmdMysql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteTokenCmdMysql) Connection() *sql.DB                         { return c.conn }
func (c DeleteTokenCmdMysql) Recorder() audited.ChangeEventRecorder       { return MysqlRecorder }
func (c DeleteTokenCmdMysql) TableName() string                           { return "tokens" }
func (c DeleteTokenCmdMysql) GetID() string                               { return c.id }

func (c DeleteTokenCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Tokens, error) {
	queries := mdbm.New(tx)
	return queries.GetToken(ctx, mdbm.GetTokenParams{ID: c.id})
}

func (c DeleteTokenCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteToken(ctx, mdbm.DeleteTokenParams{ID: c.id})
}

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

func (c NewTokenCmdPsql) Context() context.Context                    { return c.ctx }
func (c NewTokenCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c NewTokenCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c NewTokenCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c NewTokenCmdPsql) TableName() string                           { return "tokens" }
func (c NewTokenCmdPsql) Params() any                                 { return c.params }

func (c NewTokenCmdPsql) GetID(x mdbp.Tokens) string {
	return x.ID
}

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

func (c UpdateTokenCmdPsql) Context() context.Context                    { return c.ctx }
func (c UpdateTokenCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c UpdateTokenCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c UpdateTokenCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c UpdateTokenCmdPsql) TableName() string                           { return "tokens" }
func (c UpdateTokenCmdPsql) Params() any                                 { return c.params }
func (c UpdateTokenCmdPsql) GetID() string                               { return c.params.ID }

func (c UpdateTokenCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tokens, error) {
	queries := mdbp.New(tx)
	return queries.GetToken(ctx, mdbp.GetTokenParams{ID: c.params.ID})
}

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

func (c DeleteTokenCmdPsql) Context() context.Context                    { return c.ctx }
func (c DeleteTokenCmdPsql) AuditContext() audited.AuditContext           { return c.auditCtx }
func (c DeleteTokenCmdPsql) Connection() *sql.DB                         { return c.conn }
func (c DeleteTokenCmdPsql) Recorder() audited.ChangeEventRecorder       { return PsqlRecorder }
func (c DeleteTokenCmdPsql) TableName() string                           { return "tokens" }
func (c DeleteTokenCmdPsql) GetID() string                               { return c.id }

func (c DeleteTokenCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Tokens, error) {
	queries := mdbp.New(tx)
	return queries.GetToken(ctx, mdbp.GetTokenParams{ID: c.id})
}

func (c DeleteTokenCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteToken(ctx, mdbp.DeleteTokenParams{ID: c.id})
}

func (d PsqlDatabase) DeleteTokenCmd(ctx context.Context, auditCtx audited.AuditContext, id string) DeleteTokenCmdPsql {
	return DeleteTokenCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection}
}
