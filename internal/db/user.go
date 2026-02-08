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

type Users struct {
	UserID       types.UserID    `json:"user_id"`
	Username     string          `json:"username"`
	Name         string          `json:"name"`
	Email        types.Email     `json:"email"`
	Hash         string          `json:"hash"`
	Role         string          `json:"role"`
	DateCreated  types.Timestamp `json:"date_created"`
	DateModified types.Timestamp `json:"date_modified"`
}

type CreateUserParams struct {
	Username     string          `json:"username"`
	Name         string          `json:"name"`
	Email        types.Email     `json:"email"`
	Hash         string          `json:"hash"`
	Role         string          `json:"role"`
	DateCreated  types.Timestamp `json:"date_created"`
	DateModified types.Timestamp `json:"date_modified"`
}

type UpdateUserParams struct {
	Username     string          `json:"username"`
	Name         string          `json:"name"`
	Email        types.Email     `json:"email"`
	Hash         string          `json:"hash"`
	Role         string          `json:"role"`
	DateCreated  types.Timestamp `json:"date_created"`
	DateModified types.Timestamp `json:"date_modified"`
	UserID       types.UserID    `json:"user_id"`
}

// FormParams and JSON variants removed - use typed params directly

// GENERIC section removed - FormParams and JSON variants deprecated
// Use types package for direct type conversion

// MapStringUser converts a Users struct to a StringUsers struct.
func MapStringUser(a Users) StringUsers {
	return StringUsers{
		UserID:       a.UserID.String(),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email.String(),
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated.String(),
		DateModified: a.DateModified.String(),
	}
}

///////////////////////////////
// SQLITE
//////////////////////////////

// MAPS

func (d Database) MapUser(a mdb.Users) Users {
	return Users{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Roles,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateUserParams(a CreateUserParams) mdb.CreateUserParams {
	return mdb.CreateUserParams{
		UserID:       types.NewUserID(),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Roles:        a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapUpdateUserParams(a UpdateUserParams) mdb.UpdateUserParams {
	return mdb.UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Roles:        a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		UserID:       a.UserID,
	}
}

// QUERIES

func (d Database) CountUsers() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateUserTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

func (d Database) CreateUser(ctx context.Context, ac audited.AuditContext, s CreateUserParams) (*Users, error) {
	cmd := d.NewUserCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	u := d.MapUser(result)
	return &u, nil
}

func (d Database) DeleteUser(ctx context.Context, ac audited.AuditContext, id types.UserID) error {
	cmd := d.DeleteUserCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d Database) GetUser(id types.UserID) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUser(d.Context, mdb.GetUserParams{UserID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d Database) GetUserByEmail(email types.Email) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, mdb.GetUserByEmailParams{Email: email})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d Database) ListUsers() (*[]Users, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Users{}
	for _, v := range rows {
		m := d.MapUser(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) UpdateUser(ctx context.Context, ac audited.AuditContext, s UpdateUserParams) (*string, error) {
	cmd := d.UpdateUserCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Username)
	return &u, nil
}

///////////////////////////////
// MYSQL
//////////////////////////////

// MAPS

func (d MysqlDatabase) MapUser(a mdbm.Users) Users {
	return Users{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Roles,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapCreateUserParams(a CreateUserParams) mdbm.CreateUserParams {
	return mdbm.CreateUserParams{
		UserID:       types.NewUserID(),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Roles:        a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d MysqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbm.UpdateUserParams {
	return mdbm.UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Roles:        a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		UserID:       a.UserID,
	}
}

// QUERIES

func (d MysqlDatabase) CountUsers() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateUserTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateUser(ctx context.Context, ac audited.AuditContext, s CreateUserParams) (*Users, error) {
	cmd := d.NewUserCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	u := d.MapUser(result)
	return &u, nil
}

func (d MysqlDatabase) DeleteUser(ctx context.Context, ac audited.AuditContext, id types.UserID) error {
	cmd := d.DeleteUserCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d MysqlDatabase) GetUser(id types.UserID) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUser(d.Context, mdbm.GetUserParams{UserID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d MysqlDatabase) GetUserByEmail(email types.Email) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, mdbm.GetUserByEmailParams{Email: email})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d MysqlDatabase) ListUsers() (*[]Users, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Users{}
	for _, v := range rows {
		m := d.MapUser(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d MysqlDatabase) UpdateUser(ctx context.Context, ac audited.AuditContext, s UpdateUserParams) (*string, error) {
	cmd := d.UpdateUserCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Username)
	return &u, nil
}

///////////////////////////////
// POSTGRES
//////////////////////////////

// MAPS

func (d PsqlDatabase) MapUser(a mdbp.Users) Users {
	return Users{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Roles,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateUserParams(a CreateUserParams) mdbp.CreateUserParams {
	return mdbp.CreateUserParams{
		UserID:       types.NewUserID(),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Roles:        a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbp.UpdateUserParams {
	return mdbp.UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Roles:        a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		UserID:       a.UserID,
	}
}

// QUERIES

func (d PsqlDatabase) CountUsers() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateUserTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateUser(ctx context.Context, ac audited.AuditContext, s CreateUserParams) (*Users, error) {
	cmd := d.NewUserCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	u := d.MapUser(result)
	return &u, nil
}

func (d PsqlDatabase) DeleteUser(ctx context.Context, ac audited.AuditContext, id types.UserID) error {
	cmd := d.DeleteUserCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

func (d PsqlDatabase) GetUser(id types.UserID) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUser(d.Context, mdbp.GetUserParams{UserID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d PsqlDatabase) GetUserByEmail(email types.Email) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, mdbp.GetUserByEmailParams{Email: email})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

func (d PsqlDatabase) ListUsers() (*[]Users, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Users{}
	for _, v := range rows {
		m := d.MapUser(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d PsqlDatabase) UpdateUser(ctx context.Context, ac audited.AuditContext, s UpdateUserParams) (*string, error) {
	cmd := d.UpdateUserCmd(ctx, ac, s)
	if err := audited.Update(cmd); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Username)
	return &u, nil
}

// ========== AUDITED COMMAND TYPES ==========

// ----- SQLite CREATE -----

type NewUserCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserCmd) Context() context.Context              { return c.ctx }
func (c NewUserCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserCmd) Connection() *sql.DB                   { return c.conn }
func (c NewUserCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserCmd) TableName() string                     { return "users" }
func (c NewUserCmd) Params() any                           { return c.params }
func (c NewUserCmd) GetID(u mdb.Users) string              { return string(u.UserID) }

func (c NewUserCmd) Execute(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
	queries := mdb.New(tx)
	return queries.CreateUser(ctx, mdb.CreateUserParams{
		UserID:       types.NewUserID(),
		Username:     c.params.Username,
		Name:         c.params.Name,
		Email:        c.params.Email,
		Hash:         c.params.Hash,
		Roles:        c.params.Role,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d Database) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
	return NewUserCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

type UpdateUserCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserCmd) Context() context.Context              { return c.ctx }
func (c UpdateUserCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserCmd) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserCmd) TableName() string                     { return "users" }
func (c UpdateUserCmd) Params() any                           { return c.params }
func (c UpdateUserCmd) GetID() string                         { return string(c.params.UserID) }

func (c UpdateUserCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
	queries := mdb.New(tx)
	return queries.GetUser(ctx, mdb.GetUserParams{UserID: c.params.UserID})
}

func (c UpdateUserCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.UpdateUser(ctx, mdb.UpdateUserParams{
		Username:     c.params.Username,
		Name:         c.params.Name,
		Email:        c.params.Email,
		Hash:         c.params.Hash,
		Roles:        c.params.Role,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		UserID:       c.params.UserID,
	})
}

func (d Database) UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserParams) UpdateUserCmd {
	return UpdateUserCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

type DeleteUserCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserCmd) Context() context.Context              { return c.ctx }
func (c DeleteUserCmd) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserCmd) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserCmd) TableName() string                     { return "users" }
func (c DeleteUserCmd) GetID() string                         { return string(c.id) }

func (c DeleteUserCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
	queries := mdb.New(tx)
	return queries.GetUser(ctx, mdb.GetUserParams{UserID: c.id})
}

func (c DeleteUserCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteUser(ctx, mdb.DeleteUserParams{UserID: c.id})
}

func (d Database) DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmd {
	return DeleteUserCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

type NewUserCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserCmdMysql) Context() context.Context              { return c.ctx }
func (c NewUserCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c NewUserCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserCmdMysql) TableName() string                     { return "users" }
func (c NewUserCmdMysql) Params() any                           { return c.params }
func (c NewUserCmdMysql) GetID(u mdbm.Users) string             { return string(u.UserID) }

func (c NewUserCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Users, error) {
	queries := mdbm.New(tx)
	params := mdbm.CreateUserParams{
		UserID:       types.NewUserID(),
		Username:     c.params.Username,
		Name:         c.params.Name,
		Email:        c.params.Email,
		Hash:         c.params.Hash,
		Roles:        c.params.Role,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	}
	if err := queries.CreateUser(ctx, params); err != nil {
		return mdbm.Users{}, err
	}
	return queries.GetUser(ctx, mdbm.GetUserParams{UserID: params.UserID})
}

func (d MysqlDatabase) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmdMysql {
	return NewUserCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

type UpdateUserCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserCmdMysql) Context() context.Context              { return c.ctx }
func (c UpdateUserCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserCmdMysql) TableName() string                     { return "users" }
func (c UpdateUserCmdMysql) Params() any                           { return c.params }
func (c UpdateUserCmdMysql) GetID() string                         { return string(c.params.UserID) }

func (c UpdateUserCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Users, error) {
	queries := mdbm.New(tx)
	return queries.GetUser(ctx, mdbm.GetUserParams{UserID: c.params.UserID})
}

func (c UpdateUserCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.UpdateUser(ctx, mdbm.UpdateUserParams{
		Username:     c.params.Username,
		Name:         c.params.Name,
		Email:        c.params.Email,
		Hash:         c.params.Hash,
		Roles:        c.params.Role,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		UserID:       c.params.UserID,
	})
}

func (d MysqlDatabase) UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserParams) UpdateUserCmdMysql {
	return UpdateUserCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

type DeleteUserCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserCmdMysql) Context() context.Context              { return c.ctx }
func (c DeleteUserCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserCmdMysql) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserCmdMysql) TableName() string                     { return "users" }
func (c DeleteUserCmdMysql) GetID() string                         { return string(c.id) }

func (c DeleteUserCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Users, error) {
	queries := mdbm.New(tx)
	return queries.GetUser(ctx, mdbm.GetUserParams{UserID: c.id})
}

func (c DeleteUserCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteUser(ctx, mdbm.DeleteUserParams{UserID: c.id})
}

func (d MysqlDatabase) DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmdMysql {
	return DeleteUserCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

type NewUserCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c NewUserCmdPsql) Context() context.Context              { return c.ctx }
func (c NewUserCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c NewUserCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c NewUserCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c NewUserCmdPsql) TableName() string                     { return "users" }
func (c NewUserCmdPsql) Params() any                           { return c.params }
func (c NewUserCmdPsql) GetID(u mdbp.Users) string             { return string(u.UserID) }

func (c NewUserCmdPsql) Execute(ctx context.Context, tx audited.DBTX) (mdbp.Users, error) {
	queries := mdbp.New(tx)
	return queries.CreateUser(ctx, mdbp.CreateUserParams{
		UserID:       types.NewUserID(),
		Username:     c.params.Username,
		Name:         c.params.Name,
		Email:        c.params.Email,
		Hash:         c.params.Hash,
		Roles:        c.params.Role,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
	})
}

func (d PsqlDatabase) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmdPsql {
	return NewUserCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

type UpdateUserCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c UpdateUserCmdPsql) Context() context.Context              { return c.ctx }
func (c UpdateUserCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c UpdateUserCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c UpdateUserCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c UpdateUserCmdPsql) TableName() string                     { return "users" }
func (c UpdateUserCmdPsql) Params() any                           { return c.params }
func (c UpdateUserCmdPsql) GetID() string                         { return string(c.params.UserID) }

func (c UpdateUserCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Users, error) {
	queries := mdbp.New(tx)
	return queries.GetUser(ctx, mdbp.GetUserParams{UserID: c.params.UserID})
}

func (c UpdateUserCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.UpdateUser(ctx, mdbp.UpdateUserParams{
		Username:     c.params.Username,
		Name:         c.params.Name,
		Email:        c.params.Email,
		Hash:         c.params.Hash,
		Roles:        c.params.Role,
		DateCreated:  c.params.DateCreated,
		DateModified: c.params.DateModified,
		UserID:       c.params.UserID,
	})
}

func (d PsqlDatabase) UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserParams) UpdateUserCmdPsql {
	return UpdateUserCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

type DeleteUserCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

func (c DeleteUserCmdPsql) Context() context.Context              { return c.ctx }
func (c DeleteUserCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }
func (c DeleteUserCmdPsql) Connection() *sql.DB                   { return c.conn }
func (c DeleteUserCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }
func (c DeleteUserCmdPsql) TableName() string                     { return "users" }
func (c DeleteUserCmdPsql) GetID() string                         { return string(c.id) }

func (c DeleteUserCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Users, error) {
	queries := mdbp.New(tx)
	return queries.GetUser(ctx, mdbp.GetUserParams{UserID: c.id})
}

func (c DeleteUserCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteUser(ctx, mdbp.DeleteUserParams{UserID: c.id})
}

func (d PsqlDatabase) DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmdPsql {
	return DeleteUserCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
