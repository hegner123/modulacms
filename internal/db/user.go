package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// Types

// Users represents a user record from the database.
type Users struct {
	UserID       int64          `json:"user_id"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

// CreateUserParams contains parameters needed to create a new user.
type CreateUserParams struct {
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

// UpdateUserParams contains parameters needed to update a user.
type UpdateUserParams struct {
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	UserID       int64          `json:"user_id"`
}

// UsersHistoryEntry represents a historical user record.
type UsersHistoryEntry struct {
	UserID       int64          `json:"user_id"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

// CreateUserFormParams contains form parameters for user creation.
type CreateUserFormParams struct {
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

// UpdateUserFormParams contains form parameters for user updates.
type UpdateUserFormParams struct {
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	UserID       string `json:"user_id"`
}

// UsersJSON represents a user record using types that are JSON friendly
type UsersJSON struct {
	UserID       int64      `json:"user_id"`
	Username     string     `json:"username"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Hash         string     `json:"hash"`
	Role         int64      `json:"role"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
}

// CreateUserParamsJSON contains parameters for creating a user record using JSON friendly types
type CreateUserParamsJSON struct {
	Username     string     `json:"username"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Hash         string     `json:"hash"`
	Role         int64      `json:"role"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
}

// UpdateUserParamsJSON contains parameters for user updates using JSON friendly types.
type UpdateUserParamsJSON struct {
	Username     string     `json:"username"`
	Name         string     `json:"name"`
	Email        string     `json:"email"`
	Hash         string     `json:"hash"`
	Role         int64      `json:"role"`
	DateCreated  NullString `json:"date_created"`
	DateModified NullString `json:"date_modified"`
	UserID       int64      `json:"user_id"`
}

// Generic Functions

// MapCreateUserParams converts form parameters to database parameters for user creation.
func MapCreateUserParams(a CreateUserFormParams) CreateUserParams {
	return CreateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         StringToInt64(a.Role),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
	}
}

// MapUpdateUserParams converts form parameters to database parameters for user updates.
func MapUpdateUserParams(a UpdateUserFormParams) UpdateUserParams {
	return UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         StringToInt64(a.Role),
		DateCreated:  StringToNullString(a.DateCreated),
		DateModified: StringToNullString(a.DateModified),
		UserID:       StringToInt64(a.UserID),
	}
}

// MapStringUser converts a Users struct to a StringUsers struct.
func MapStringUser(a Users) StringUsers {
	return StringUsers{
		UserID:       strconv.FormatInt(a.UserID, 10),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         strconv.FormatInt(a.Role, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
	}
}

// MapCreateUserJSONParams converts JSON parameters to database parameters for user creation.
func MapCreateUserJSONParams(a CreateUserParamsJSON) CreateUserParams {
	return CreateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
	}
}

// MapUpdateUserJSONParams converts JSON parameters to database parameters for user updates.
func MapUpdateUserJSONParams(a UpdateUserParamsJSON) UpdateUserParams {
	return UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated.NullString,
		DateModified: a.DateModified.NullString,
		UserID:       a.UserID,
	}
}

// SQLite Database Implementation

// MapUser maps SQLite user data to the common Users struct.
func (d Database) MapUser(a mdb.Users) Users {
	return Users{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapCreateUserParams maps common CreateUserParams to SQLite-specific parameters.
func (d Database) MapCreateUserParams(a CreateUserParams) mdb.CreateUserParams {
	return mdb.CreateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MapUpdateUserParams maps common UpdateUserParams to SQLite-specific parameters.
func (d Database) MapUpdateUserParams(a UpdateUserParams) mdb.UpdateUserParams {
	return mdb.UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		UserID:       a.UserID,
	}
}

// CountUsers returns the total number of users in the SQLite database.
func (d Database) CountUsers() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateUserTable creates the user table in the SQLite database.
func (d Database) CreateUserTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

// CreateUser creates a new user in the SQLite database.
func (d Database) CreateUser(s CreateUserParams) (*Users, error) {
	params := d.MapCreateUserParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateUser(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUser.\n %v\n", err)
		return nil, e
	}
	u := d.MapUser(row)
	return &u, nil
}

// DeleteUser removes a user from the SQLite database by ID.
func (d Database) DeleteUser(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUser(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}
	return nil
}

// GetUser retrieves a user by ID from the SQLite database.
func (d Database) GetUser(id int64) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUser(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// GetUserByEmail retrieves a user by email from the SQLite database.
func (d Database) GetUserByEmail(email string) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, email)
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUsers retrieves all users from the SQLite database.
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

// UpdateUser updates an existing user in the SQLite database.
func (d Database) UpdateUser(s UpdateUserParams) (*string, error) {
	params := d.MapUpdateUserParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateUser(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Username)
	return &u, nil
}

// MySQL Database Implementation

// MapUser maps MySQL user data to the common Users struct.
func (d MysqlDatabase) MapUser(a mdbm.Users) Users {
	return Users{
		UserID:       int64(a.UserID),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int64(a.Role),
		DateCreated:  StringToNullString(a.DateCreated.String()),
		DateModified: StringToNullString(a.DateModified.String()),
	}
}

// MapCreateUserParams maps common CreateUserParams to MySQL-specific parameters.
func (d MysqlDatabase) MapCreateUserParams(a CreateUserParams) mdbm.CreateUserParams {
	return mdbm.CreateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int32(a.Role),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
	}
}

// MapUpdateUserParams maps common UpdateUserParams to MySQL-specific parameters.
func (d MysqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbm.UpdateUserParams {
	return mdbm.UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int32(a.Role),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
		UserID:       int32(a.UserID),
	}
}

// CountUsers returns the total number of users in the MySQL database.
func (d MysqlDatabase) CountUsers() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateUserTable creates the user table in the MySQL database.
func (d MysqlDatabase) CreateUserTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

// CreateUser creates a new user in the MySQL database.
func (d MysqlDatabase) CreateUser(s CreateUserParams) (*Users, error) {
	params := d.MapCreateUserParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateUser(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUser.\n %v\n", err)
		return nil, e
	}
	row, err := queries.GetLastUser(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted User: %v\n", err)
	}
	u := d.MapUser(row)
	return &u, nil
}

// DeleteUser removes a user from the MySQL database by ID.
func (d MysqlDatabase) DeleteUser(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}
	return nil
}

// GetUser retrieves a user by ID from the MySQL database.
func (d MysqlDatabase) GetUser(id int64) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUser(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// GetUserByEmail retrieves a user by email from the MySQL database.
func (d MysqlDatabase) GetUserByEmail(email string) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, email)
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUsers retrieves all users from the MySQL database.
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

// UpdateUser updates an existing user in the MySQL database.
func (d MysqlDatabase) UpdateUser(s UpdateUserParams) (*string, error) {
	params := d.MapUpdateUserParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateUser(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Username)
	return &u, nil
}

// PostgreSQL Database Implementation

// MapUser maps PostgreSQL user data to the common Users struct.
func (d PsqlDatabase) MapUser(a mdbp.Users) Users {
	return Users{
		UserID:       int64(a.UserID),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int64(a.Role),
		DateCreated:  StringToNullString(NullTimeToString(a.DateCreated)),
		DateModified: StringToNullString(NullTimeToString(a.DateModified)),
	}
}

// MapCreateUserParams maps common CreateUserParams to PostgreSQL-specific parameters.
func (d PsqlDatabase) MapCreateUserParams(a CreateUserParams) mdbp.CreateUserParams {
	return mdbp.CreateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int32(a.Role),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
	}
}

// MapUpdateUserParams maps common UpdateUserParams to PostgreSQL-specific parameters.
func (d PsqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbp.UpdateUserParams {
	return mdbp.UpdateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int32(a.Role),
		DateCreated:  StringToNTime(a.DateCreated.String),
		DateModified: StringToNTime(a.DateModified.String),
		UserID:       int32(a.UserID),
	}
}

// CountUsers returns the total number of users in the PostgreSQL database.
func (d PsqlDatabase) CountUsers() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountUser(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

// CreateUserTable creates the user table in the PostgreSQL database.
func (d PsqlDatabase) CreateUserTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateUserTable(d.Context)
	return err
}

// CreateUser creates a new user in the PostgreSQL database.
func (d PsqlDatabase) CreateUser(s CreateUserParams) (*Users, error) {
	params := d.MapCreateUserParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateUser(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUser.\n %v\n", err)
		return nil, e
	}
	u := d.MapUser(row)
	return &u, nil
}

// DeleteUser removes a user from the PostgreSQL database by ID.
func (d PsqlDatabase) DeleteUser(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}
	return nil
}

// GetUser retrieves a user by ID from the PostgreSQL database.
func (d PsqlDatabase) GetUser(id int64) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUser(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// GetUserByEmail retrieves a user by email from the PostgreSQL database.
func (d PsqlDatabase) GetUserByEmail(email string) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, email)
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// ListUsers retrieves all users from the PostgreSQL database.
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

// UpdateUser updates an existing user in the PostgreSQL database.
func (d PsqlDatabase) UpdateUser(s UpdateUserParams) (*string, error) {
	params := d.MapUpdateUserParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateUser(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Username)
	return &u, nil
}
