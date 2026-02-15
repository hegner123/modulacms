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

// Users represents a user record in the database.
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

// CreateUserParams contains parameters for creating a new user.
type CreateUserParams struct {
	Username     string          `json:"username"`
	Name         string          `json:"name"`
	Email        types.Email     `json:"email"`
	Hash         string          `json:"hash"`
	Role         string          `json:"role"`
	DateCreated  types.Timestamp `json:"date_created"`
	DateModified types.Timestamp `json:"date_modified"`
}

// UpdateUserParams contains parameters for updating an existing user.
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

// MapUser converts a sqlc-generated SQLite user to the wrapper type.
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

// MapCreateUserParams converts wrapper params to sqlc-generated SQLite params.
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

// MapUpdateUserParams converts wrapper params to sqlc-generated SQLite params.
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

// CountUsers returns the total number of users in the database.
func (d Database) CountUsers() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateUserTable creates the users table in the database.
func (d Database) CreateUserTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

// CreateUser inserts a new user and records an audit event.
func (d Database) CreateUser(ctx context.Context, ac audited.AuditContext, s CreateUserParams) (*Users, error) {
	cmd := d.NewUserCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	u := d.MapUser(result)
	return &u, nil
}

// DeleteUser removes a user and records an audit event.
func (d Database) DeleteUser(ctx context.Context, ac audited.AuditContext, id types.UserID) error {
	cmd := d.DeleteUserCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetUser retrieves a user by ID.
func (d Database) GetUser(id types.UserID) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUser(d.Context, mdb.GetUserParams{UserID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// GetUserByEmail retrieves a user by email address.
func (d Database) GetUserByEmail(email types.Email) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, mdb.GetUserByEmailParams{Email: email})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUsers retrieves all users in the database.
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

// UpdateUser modifies an existing user and records an audit event.
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

// MapUser converts a sqlc-generated MySQL user to the wrapper type.
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

// MapCreateUserParams converts wrapper params to sqlc-generated MySQL params.
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

// MapUpdateUserParams converts wrapper params to sqlc-generated MySQL params.
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

// CountUsers returns the total number of users in the database.
func (d MysqlDatabase) CountUsers() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateUserTable creates the users table in the database.
func (d MysqlDatabase) CreateUserTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

// CreateUser inserts a new user and records an audit event.
func (d MysqlDatabase) CreateUser(ctx context.Context, ac audited.AuditContext, s CreateUserParams) (*Users, error) {
	cmd := d.NewUserCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	u := d.MapUser(result)
	return &u, nil
}

// DeleteUser removes a user and records an audit event.
func (d MysqlDatabase) DeleteUser(ctx context.Context, ac audited.AuditContext, id types.UserID) error {
	cmd := d.DeleteUserCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetUser retrieves a user by ID.
func (d MysqlDatabase) GetUser(id types.UserID) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUser(d.Context, mdbm.GetUserParams{UserID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// GetUserByEmail retrieves a user by email address.
func (d MysqlDatabase) GetUserByEmail(email types.Email) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, mdbm.GetUserByEmailParams{Email: email})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUsers retrieves all users in the database.
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

// UpdateUser modifies an existing user and records an audit event.
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

// MapUser converts a sqlc-generated PostgreSQL user to the wrapper type.
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

// MapCreateUserParams converts wrapper params to sqlc-generated PostgreSQL params.
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

// MapUpdateUserParams converts wrapper params to sqlc-generated PostgreSQL params.
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

// CountUsers returns the total number of users in the database.
func (d PsqlDatabase) CountUsers() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateUserTable creates the users table in the database.
func (d PsqlDatabase) CreateUserTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

// CreateUser inserts a new user and records an audit event.
func (d PsqlDatabase) CreateUser(ctx context.Context, ac audited.AuditContext, s CreateUserParams) (*Users, error) {
	cmd := d.NewUserCmd(ctx, ac, s)
	result, err := audited.Create(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	u := d.MapUser(result)
	return &u, nil
}

// DeleteUser removes a user and records an audit event.
func (d PsqlDatabase) DeleteUser(ctx context.Context, ac audited.AuditContext, id types.UserID) error {
	cmd := d.DeleteUserCmd(ctx, ac, id)
	return audited.Delete(cmd)
}

// GetUser retrieves a user by ID.
func (d PsqlDatabase) GetUser(id types.UserID) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUser(d.Context, mdbp.GetUserParams{UserID: id})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// GetUserByEmail retrieves a user by email address.
func (d PsqlDatabase) GetUserByEmail(email types.Email) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, mdbp.GetUserByEmailParams{Email: email})
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUsers retrieves all users in the database.
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

// UpdateUser modifies an existing user and records an audit event.
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

// NewUserCmd is an audited command for creating users.
type NewUserCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewUserCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewUserCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewUserCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewUserCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c NewUserCmd) TableName() string                     { return "users" }

// Params returns the parameters for this command.
func (c NewUserCmd) Params() any                           { return c.params }

// GetID extracts the ID from a user record.
func (c NewUserCmd) GetID(u mdb.Users) string              { return string(u.UserID) }

// Execute creates the user in the database.
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

// NewUserCmd creates a command for inserting a user.
func (d Database) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
	return NewUserCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite UPDATE -----

// UpdateUserCmd is an audited command for updating users.
type UpdateUserCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateUserCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateUserCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateUserCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateUserCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c UpdateUserCmd) TableName() string                     { return "users" }

// Params returns the parameters for this command.
func (c UpdateUserCmd) Params() any                           { return c.params }

// GetID returns the user ID for this command.
func (c UpdateUserCmd) GetID() string                         { return string(c.params.UserID) }

// GetBefore retrieves the user before the update.
func (c UpdateUserCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
	queries := mdb.New(tx)
	return queries.GetUser(ctx, mdb.GetUserParams{UserID: c.params.UserID})
}

// Execute updates the user in the database.
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

// UpdateUserCmd creates a command for updating a user.
func (d Database) UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserParams) UpdateUserCmd {
	return UpdateUserCmd{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- SQLite DELETE -----

// DeleteUserCmd is an audited command for deleting users.
type DeleteUserCmd struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteUserCmd) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteUserCmd) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteUserCmd) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteUserCmd) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c DeleteUserCmd) TableName() string                     { return "users" }

// GetID returns the user ID for this command.
func (c DeleteUserCmd) GetID() string                         { return string(c.id) }

// GetBefore retrieves the user before the delete.
func (c DeleteUserCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
	queries := mdb.New(tx)
	return queries.GetUser(ctx, mdb.GetUserParams{UserID: c.id})
}

// Execute deletes the user from the database.
func (c DeleteUserCmd) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdb.New(tx)
	return queries.DeleteUser(ctx, mdb.DeleteUserParams{UserID: c.id})
}

// DeleteUserCmd creates a command for deleting a user.
func (d Database) DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmd {
	return DeleteUserCmd{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: SQLiteRecorder}
}

// ----- MySQL CREATE -----

// NewUserCmdMysql is an audited command for creating users in MySQL.
type NewUserCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewUserCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewUserCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewUserCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewUserCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c NewUserCmdMysql) TableName() string                     { return "users" }

// Params returns the parameters for this command.
func (c NewUserCmdMysql) Params() any                           { return c.params }

// GetID extracts the ID from a user record.
func (c NewUserCmdMysql) GetID(u mdbm.Users) string             { return string(u.UserID) }

// Execute creates the user in the database.
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

// NewUserCmd creates a command for inserting a user.
func (d MysqlDatabase) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmdMysql {
	return NewUserCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL UPDATE -----

// UpdateUserCmdMysql is an audited command for updating users in MySQL.
type UpdateUserCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateUserCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateUserCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateUserCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateUserCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c UpdateUserCmdMysql) TableName() string                     { return "users" }

// Params returns the parameters for this command.
func (c UpdateUserCmdMysql) Params() any                           { return c.params }

// GetID returns the user ID for this command.
func (c UpdateUserCmdMysql) GetID() string                         { return string(c.params.UserID) }

// GetBefore retrieves the user before the update.
func (c UpdateUserCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Users, error) {
	queries := mdbm.New(tx)
	return queries.GetUser(ctx, mdbm.GetUserParams{UserID: c.params.UserID})
}

// Execute updates the user in the database.
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

// UpdateUserCmd creates a command for updating a user.
func (d MysqlDatabase) UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserParams) UpdateUserCmdMysql {
	return UpdateUserCmdMysql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- MySQL DELETE -----

// DeleteUserCmdMysql is an audited command for deleting users in MySQL.
type DeleteUserCmdMysql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteUserCmdMysql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteUserCmdMysql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteUserCmdMysql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteUserCmdMysql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c DeleteUserCmdMysql) TableName() string                     { return "users" }

// GetID returns the user ID for this command.
func (c DeleteUserCmdMysql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the user before the delete.
func (c DeleteUserCmdMysql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbm.Users, error) {
	queries := mdbm.New(tx)
	return queries.GetUser(ctx, mdbm.GetUserParams{UserID: c.id})
}

// Execute deletes the user from the database.
func (c DeleteUserCmdMysql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbm.New(tx)
	return queries.DeleteUser(ctx, mdbm.DeleteUserParams{UserID: c.id})
}

// DeleteUserCmd creates a command for deleting a user.
func (d MysqlDatabase) DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmdMysql {
	return DeleteUserCmdMysql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: MysqlRecorder}
}

// ----- PostgreSQL CREATE -----

// NewUserCmdPsql is an audited command for creating users in PostgreSQL.
type NewUserCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   CreateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c NewUserCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c NewUserCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c NewUserCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c NewUserCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c NewUserCmdPsql) TableName() string                     { return "users" }

// Params returns the parameters for this command.
func (c NewUserCmdPsql) Params() any                           { return c.params }

// GetID extracts the ID from a user record.
func (c NewUserCmdPsql) GetID(u mdbp.Users) string             { return string(u.UserID) }

// Execute creates the user in the database.
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

// NewUserCmd creates a command for inserting a user.
func (d PsqlDatabase) NewUserCmd(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmdPsql {
	return NewUserCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL UPDATE -----

// UpdateUserCmdPsql is an audited command for updating users in PostgreSQL.
type UpdateUserCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	params   UpdateUserParams
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c UpdateUserCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c UpdateUserCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c UpdateUserCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c UpdateUserCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c UpdateUserCmdPsql) TableName() string                     { return "users" }

// Params returns the parameters for this command.
func (c UpdateUserCmdPsql) Params() any                           { return c.params }

// GetID returns the user ID for this command.
func (c UpdateUserCmdPsql) GetID() string                         { return string(c.params.UserID) }

// GetBefore retrieves the user before the update.
func (c UpdateUserCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Users, error) {
	queries := mdbp.New(tx)
	return queries.GetUser(ctx, mdbp.GetUserParams{UserID: c.params.UserID})
}

// Execute updates the user in the database.
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

// UpdateUserCmd creates a command for updating a user.
func (d PsqlDatabase) UpdateUserCmd(ctx context.Context, auditCtx audited.AuditContext, params UpdateUserParams) UpdateUserCmdPsql {
	return UpdateUserCmdPsql{ctx: ctx, auditCtx: auditCtx, params: params, conn: d.Connection, recorder: PsqlRecorder}
}

// ----- PostgreSQL DELETE -----

// DeleteUserCmdPsql is an audited command for deleting users in PostgreSQL.
type DeleteUserCmdPsql struct {
	ctx      context.Context
	auditCtx audited.AuditContext
	id       types.UserID
	conn     *sql.DB
	recorder audited.ChangeEventRecorder
}

// Context returns the command context.
func (c DeleteUserCmdPsql) Context() context.Context              { return c.ctx }

// AuditContext returns the audit context.
func (c DeleteUserCmdPsql) AuditContext() audited.AuditContext     { return c.auditCtx }

// Connection returns the database connection.
func (c DeleteUserCmdPsql) Connection() *sql.DB                   { return c.conn }

// Recorder returns the change event recorder.
func (c DeleteUserCmdPsql) Recorder() audited.ChangeEventRecorder { return c.recorder }

// TableName returns the table name for this command.
func (c DeleteUserCmdPsql) TableName() string                     { return "users" }

// GetID returns the user ID for this command.
func (c DeleteUserCmdPsql) GetID() string                         { return string(c.id) }

// GetBefore retrieves the user before the delete.
func (c DeleteUserCmdPsql) GetBefore(ctx context.Context, tx audited.DBTX) (mdbp.Users, error) {
	queries := mdbp.New(tx)
	return queries.GetUser(ctx, mdbp.GetUserParams{UserID: c.id})
}

// Execute deletes the user from the database.
func (c DeleteUserCmdPsql) Execute(ctx context.Context, tx audited.DBTX) error {
	queries := mdbp.New(tx)
	return queries.DeleteUser(ctx, mdbp.DeleteUserParams{UserID: c.id})
}

// DeleteUserCmd creates a command for deleting a user.
func (d PsqlDatabase) DeleteUserCmd(ctx context.Context, auditCtx audited.AuditContext, id types.UserID) DeleteUserCmdPsql {
	return DeleteUserCmdPsql{ctx: ctx, auditCtx: auditCtx, id: id, conn: d.Connection, recorder: PsqlRecorder}
}
