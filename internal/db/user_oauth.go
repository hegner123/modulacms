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

type UserOauth struct {
	UserOauthID         types.UserOauthID    `json:"user_oauth_id"`
	UserID              types.NullableUserID `json:"user_id"`
	OauthProvider       string               `json:"oauth_provider"`
	OauthProviderUserID string               `json:"oauth_provider_user_id"`
	AccessToken         string               `json:"access_token"`
	RefreshToken        string               `json:"refresh_token"`
	TokenExpiresAt      string               `json:"token_expires_at"`
	DateCreated         types.Timestamp      `json:"date_created"`
}

type CreateUserOauthParams struct {
	UserID              types.NullableUserID `json:"user_id"`
	OauthProvider       string               `json:"oauth_provider"`
	OauthProviderUserID string               `json:"oauth_provider_user_id"`
	AccessToken         string               `json:"access_token"`
	RefreshToken        string               `json:"refresh_token"`
	TokenExpiresAt      string               `json:"token_expires_at"`
	DateCreated         types.Timestamp      `json:"date_created"`
}

type UpdateUserOauthParams struct {
	AccessToken    string            `json:"access_token"`
	RefreshToken   string            `json:"refresh_token"`
	TokenExpiresAt string            `json:"token_expires_at"`
	UserOauthID    types.UserOauthID `json:"user_oauth_id"`
}

// FormParams and HistoryEntry variants removed - use typed params directly

// GENERIC section removed - FormParams deprecated
// Use types package for direct type conversion

// MapStringUserOauth converts UserOauth to StringUserOauth for table display
func MapStringUserOauth(a UserOauth) StringUserOauth {
	return StringUserOauth{
		UserOauthID:         a.UserOauthID.String(),
		UserID:              a.UserID.String(),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapUserOauth(a mdb.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         a.UserOAuthID,
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OAuthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated,
	}
}

func (d Database) MapCreateUserOauthParams(a CreateUserOauthParams) mdb.CreateUserOauthParams {
	return mdb.CreateUserOauthParams{
		UserOAuthID:         types.NewUserOauthID(),
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OAuthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated,
	}
}

func (d Database) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdb.UpdateUserOauthParams {
	return mdb.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: a.TokenExpiresAt,
		UserOAuthID:    a.UserOauthID,
	}
}

// QUERIES

func (d Database) CountUserOauths() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountUserOauths(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateUserOauthTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserOauthTable(d.Context)
	return err
}

func (d Database) CreateUserOauth(ctx context.Context, ac audited.AuditContext, s CreateUserOauthParams) (*UserOauth, error) {
	cmd := d.NewUserOauthCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create userOauth: %w", err)
	}
	r := d.MapUserOauth(result)
	return &r, nil
}

func (d Database) DeleteUserOauth(ctx context.Context, ac audited.AuditContext, id types.UserOauthID) error {
	cmd := d.DeleteUserOauthCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetUserOauth(id types.UserOauthID) (*UserOauth, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, mdb.GetUserOauthParams{UserOAuthID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d Database) GetUserOauthByUserId(userID types.NullableUserID) (*UserOauth, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserOauthByUserId(d.Context, mdb.GetUserOauthByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d Database) GetUserOauthByProviderID(provider string, providerUserID string) (*UserOauth, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserOauthByProviderID(d.Context, mdb.GetUserOauthByProviderIDParams{
		OauthProvider:       provider,
		OAuthProviderUserID: providerUserID,
	})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d Database) ListUserOauths() (*[]UserOauth, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListUserOauth(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get UserOauths: %v\n", err)
	}
	res := []UserOauth{}
	for _, v := range rows {
		m := d.MapUserOauth(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateUserOauth(ctx context.Context, ac audited.AuditContext, s UpdateUserOauthParams) (*string, error) {
	cmd := d.UpdateUserOauthCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update userOauth: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.UserOauthID)
	return &msg, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapUserOauth(a mdbm.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         a.UserOAuthID,
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OAuthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt.String(),
		DateCreated:         a.DateCreated,
	}
}

func (d MysqlDatabase) MapCreateUserOauthParams(a CreateUserOauthParams) mdbm.CreateUserOauthParams {
	return mdbm.CreateUserOauthParams{
		UserOAuthID:         types.NewUserOauthID(),
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OAuthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      ParseTime(a.TokenExpiresAt),
		DateCreated:         a.DateCreated,
	}
}

func (d MysqlDatabase) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdbm.UpdateUserOauthParams {
	return mdbm.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: ParseTime(a.TokenExpiresAt),
		UserOAuthID:    a.UserOauthID,
	}
}

// QUERIES

func (d MysqlDatabase) CountUserOauths() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountUserOauths(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateUserOauthTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserOauthTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateUserOauth(ctx context.Context, ac audited.AuditContext, s CreateUserOauthParams) (*UserOauth, error) {
	cmd := d.NewUserOauthCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create userOauth: %w", err)
	}
	r := d.MapUserOauth(result)
	return &r, nil
}

func (d MysqlDatabase) DeleteUserOauth(ctx context.Context, ac audited.AuditContext, id types.UserOauthID) error {
	cmd := d.DeleteUserOauthCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetUserOauth(id types.UserOauthID) (*UserOauth, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, mdbm.GetUserOauthParams{UserOAuthID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d MysqlDatabase) GetUserOauthByUserId(userID types.NullableUserID) (*UserOauth, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserOauthByUserId(d.Context, mdbm.GetUserOauthByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d MysqlDatabase) GetUserOauthByProviderID(provider string, providerUserID string) (*UserOauth, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserOauthByProviderID(d.Context, mdbm.GetUserOauthByProviderIDParams{
		OauthProvider:       provider,
		OAuthProviderUserID: providerUserID,
	})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d MysqlDatabase) ListUserOauths() (*[]UserOauth, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListUserOauth(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get UserOauths: %v\n", err)
	}
	res := []UserOauth{}
	for _, v := range rows {
		m := d.MapUserOauth(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateUserOauth(ctx context.Context, ac audited.AuditContext, s UpdateUserOauthParams) (*string, error) {
	cmd := d.UpdateUserOauthCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update userOauth: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.UserOauthID)
	return &msg, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapUserOauth(a mdbp.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         a.UserOAuthID,
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OAuthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt.String(),
		DateCreated:         a.DateCreated,
	}
}

func (d PsqlDatabase) MapCreateUserOauthParams(a CreateUserOauthParams) mdbp.CreateUserOauthParams {
	return mdbp.CreateUserOauthParams{
		UserOAuthID:         types.NewUserOauthID(),
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OAuthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      ParseTime(a.TokenExpiresAt),
		DateCreated:         a.DateCreated,
	}
}

func (d PsqlDatabase) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdbp.UpdateUserOauthParams {
	return mdbp.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: ParseTime(a.TokenExpiresAt),
		UserOAuthID:    a.UserOauthID,
	}
}

// QUERIES

func (d PsqlDatabase) CountUserOauths() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountUserOauths(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateUserOauthTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserOauthTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateUserOauth(ctx context.Context, ac audited.AuditContext, s CreateUserOauthParams) (*UserOauth, error) {
	cmd := d.NewUserOauthCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create userOauth: %w", err)
	}
	r := d.MapUserOauth(result)
	return &r, nil
}

func (d PsqlDatabase) DeleteUserOauth(ctx context.Context, ac audited.AuditContext, id types.UserOauthID) error {
	cmd := d.DeleteUserOauthCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetUserOauth(id types.UserOauthID) (*UserOauth, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, mdbp.GetUserOauthParams{UserOAuthID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d PsqlDatabase) GetUserOauthByUserId(userID types.NullableUserID) (*UserOauth, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserOauthByUserId(d.Context, mdbp.GetUserOauthByUserIdParams{UserID: userID})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d PsqlDatabase) GetUserOauthByProviderID(provider string, providerUserID string) (*UserOauth, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserOauthByProviderID(d.Context, mdbp.GetUserOauthByProviderIDParams{
		OauthProvider:       provider,
		OAuthProviderUserID: providerUserID,
	})
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}

func (d PsqlDatabase) ListUserOauths() (*[]UserOauth, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListUserOauth(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get UserOauths: %v\n", err)
	}
	res := []UserOauth{}
	for _, v := range rows {
		m := d.MapUserOauth(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateUserOauth(ctx context.Context, ac audited.AuditContext, s UpdateUserOauthParams) (*string, error) {
	cmd := d.UpdateUserOauthCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update userOauth: %w", err)
	}
	msg := fmt.Sprintf("Successfully updated %v\n", s.UserOauthID)
	return &msg, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

type NewUserOauthCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserOauthParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserOauthCmd) Context() context.Context              { return c.ctx }
func (c NewUserOauthCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserOauthCmd) Connection() *sql.DB                   { return c.conn }
func (c NewUserOauthCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserOauthCmd) TableName() string                     { return "user_oauth" }
func (c NewUserOauthCmd) Params() any                           { return c.params }
func (c NewUserOauthCmd) GetID(u mdb.UserOauth) string          { return string(u.UserOAuthID) }

func (c NewUserOauthCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.UserOauth, error) {
	queries := mdb.New(tx)
	return queries.CreateUserOauth(ctx, mdb.CreateUserOauthParams{
		UserOAuthID:         types.NewUserOauthID(),
		UserID:              c.params.UserID,
		OauthProvider:       c.params.OauthProvider,
		OAuthProviderUserID: c.params.OauthProviderUserID,
		AccessToken:         c.params.AccessToken,
		RefreshToken:        c.params.RefreshToken,
		TokenExpiresAt:      c.params.TokenExpiresAt,
		DateCreated:         c.params.DateCreated,
	})
}

func (d Database) NewUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserOauthParams) NewUserOauthCmd {
	return NewUserOauthCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

type UpdateUserOauthCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserOauthParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserOauthCmd) Context() context.Context              { return c.ctx }
func (c UpdateUserOauthCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserOauthCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserOauthCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserOauthCmd) TableName() string                     { return "user_oauth" }
func (c UpdateUserOauthCmd) Params() any                           { return c.params }
func (c UpdateUserOauthCmd) GetID() string                         { return string(c.params.UserOauthID) }

func (c UpdateUserOauthCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.UserOauth, error) {
	queries := mdb.New(tx)
	return queries.GetUserOauth(ctx, mdb.GetUserOauthParams{UserOAuthID: c.params.UserOauthID})
}

func (c UpdateUserOauthCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateUserOauth(ctx, mdb.UpdateUserOauthParams{
		AccessToken:    c.params.AccessToken,
		RefreshToken:   c.params.RefreshToken,
		TokenExpiresAt: c.params.TokenExpiresAt,
		UserOAuthID:    c.params.UserOauthID,
	})
}

func (d Database) UpdateUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserOauthParams) UpdateUserOauthCmd {
	return UpdateUserOauthCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteUserOauthCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserOauthID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserOauthCmd) Context() context.Context              { return c.ctx }
func (c DeleteUserOauthCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserOauthCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserOauthCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserOauthCmd) TableName() string                     { return "user_oauth" }
func (c DeleteUserOauthCmd) GetID() string                         { return string(c.id) }

func (c DeleteUserOauthCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.UserOauth, error) {
	queries := mdb.New(tx)
	return queries.GetUserOauth(ctx, mdb.GetUserOauthParams{UserOAuthID: c.id})
}

func (c DeleteUserOauthCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteUserOauth(ctx, mdb.DeleteUserOauthParams{UserOAuthID: c.id})
}

func (d Database) DeleteUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserOauthID) DeleteUserOauthCmd {
	return DeleteUserOauthCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewUserOauthCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserOauthParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserOauthCmdMysql) Context() context.Context              { return c.ctx }
func (c NewUserOauthCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserOauthCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewUserOauthCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserOauthCmdMysql) TableName() string                     { return "user_oauth" }
func (c NewUserOauthCmdMysql) Params() any                           { return c.params }
func (c NewUserOauthCmdMysql) GetID(u mdbm.UserOauth) string        { return string(u.UserOAuthID) }

func (c NewUserOauthCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.UserOauth, error) {
	queries := mdbm.New(tx)
	id := types.NewUserOauthID()
	err := queries.CreateUserOauth(ctx, mdbm.CreateUserOauthParams{
		UserOAuthID:         id,
		UserID:              c.params.UserID,
		OauthProvider:       c.params.OauthProvider,
		OAuthProviderUserID: c.params.OauthProviderUserID,
		AccessToken:         c.params.AccessToken,
		RefreshToken:        c.params.RefreshToken,
		TokenExpiresAt:      ParseTime(c.params.TokenExpiresAt),
		DateCreated:         c.params.DateCreated,
	})
	if err != nil {
		return mdbm.UserOauth{}, err
	}
	return queries.GetUserOauth(ctx, mdbm.GetUserOauthParams{UserOAuthID: id})
}

func (d MysqlDatabase) NewUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserOauthParams) NewUserOauthCmdMysql {
	return NewUserOauthCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

type UpdateUserOauthCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserOauthParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserOauthCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateUserOauthCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserOauthCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserOauthCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserOauthCmdMysql) TableName() string                     { return "user_oauth" }
func (c UpdateUserOauthCmdMysql) Params() any                           { return c.params }
func (c UpdateUserOauthCmdMysql) GetID() string                         { return string(c.params.UserOauthID) }

func (c UpdateUserOauthCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.UserOauth, error) {
	queries := mdbm.New(tx)
	return queries.GetUserOauth(ctx, mdbm.GetUserOauthParams{UserOAuthID: c.params.UserOauthID})
}

func (c UpdateUserOauthCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateUserOauth(ctx, mdbm.UpdateUserOauthParams{
		AccessToken:    c.params.AccessToken,
		RefreshToken:   c.params.RefreshToken,
		TokenExpiresAt: ParseTime(c.params.TokenExpiresAt),
		UserOAuthID:    c.params.UserOauthID,
	})
}

func (d MysqlDatabase) UpdateUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserOauthParams) UpdateUserOauthCmdMysql {
	return UpdateUserOauthCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteUserOauthCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserOauthID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserOauthCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteUserOauthCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserOauthCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserOauthCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserOauthCmdMysql) TableName() string                     { return "user_oauth" }
func (c DeleteUserOauthCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteUserOauthCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.UserOauth, error) {
	queries := mdbm.New(tx)
	return queries.GetUserOauth(ctx, mdbm.GetUserOauthParams{UserOAuthID: c.id})
}

func (c DeleteUserOauthCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteUserOauth(ctx, mdbm.DeleteUserOauthParams{UserOAuthID: c.id})
}

func (d MysqlDatabase) DeleteUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserOauthID) DeleteUserOauthCmdMysql {
	return DeleteUserOauthCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewUserOauthCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserOauthParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserOauthCmdPsql) Context() context.Context              { return c.ctx }
func (c NewUserOauthCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserOauthCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewUserOauthCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserOauthCmdPsql) TableName() string                     { return "user_oauth" }
func (c NewUserOauthCmdPsql) Params() any                           { return c.params }
func (c NewUserOauthCmdPsql) GetID(u mdbp.UserOauth) string        { return string(u.UserOAuthID) }

func (c NewUserOauthCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.UserOauth, error) {
	queries := mdbp.New(tx)
	return queries.CreateUserOauth(ctx, mdbp.CreateUserOauthParams{
		UserOAuthID:         types.NewUserOauthID(),
		UserID:              c.params.UserID,
		OauthProvider:       c.params.OauthProvider,
		OAuthProviderUserID: c.params.OauthProviderUserID,
		AccessToken:         c.params.AccessToken,
		RefreshToken:        c.params.RefreshToken,
		TokenExpiresAt:      ParseTime(c.params.TokenExpiresAt),
		DateCreated:         c.params.DateCreated,
	})
}

func (d PsqlDatabase) NewUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserOauthParams) NewUserOauthCmdPsql {
	return NewUserOauthCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

type UpdateUserOauthCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserOauthParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserOauthCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateUserOauthCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserOauthCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserOauthCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserOauthCmdPsql) TableName() string                     { return "user_oauth" }
func (c UpdateUserOauthCmdPsql) Params() any                           { return c.params }
func (c UpdateUserOauthCmdPsql) GetID() string                         { return string(c.params.UserOauthID) }

func (c UpdateUserOauthCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.UserOauth, error) {
	queries := mdbp.New(tx)
	return queries.GetUserOauth(ctx, mdbp.GetUserOauthParams{UserOAuthID: c.params.UserOauthID})
}

func (c UpdateUserOauthCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateUserOauth(ctx, mdbp.UpdateUserOauthParams{
		AccessToken:    c.params.AccessToken,
		RefreshToken:   c.params.RefreshToken,
		TokenExpiresAt: ParseTime(c.params.TokenExpiresAt),
		UserOAuthID:    c.params.UserOauthID,
	})
}

func (d PsqlDatabase) UpdateUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserOauthParams) UpdateUserOauthCmdPsql {
	return UpdateUserOauthCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteUserOauthCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserOauthID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserOauthCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteUserOauthCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserOauthCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserOauthCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserOauthCmdPsql) TableName() string                     { return "user_oauth" }
func (c DeleteUserOauthCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteUserOauthCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.UserOauth, error) {
	queries := mdbp.New(tx)
	return queries.GetUserOauth(ctx, mdbp.GetUserOauthParams{UserOAuthID: c.id})
}

func (c DeleteUserOauthCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteUserOauth(ctx, mdbp.DeleteUserOauthParams{UserOAuthID: c.id})
}

func (d PsqlDatabase) DeleteUserOauthCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserOauthID) DeleteUserOauthCmdPsql {
	return DeleteUserOauthCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
