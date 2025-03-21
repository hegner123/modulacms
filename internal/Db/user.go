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
type Users struct {
	UserID       int64          `json:"user_id"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	References   any            `json:"references"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateUserParams struct {
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
}

type UpdateUserParams struct {
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	UserID       int64          `json:"user_id"`
}

type UsersHistoryEntry struct {
	UserID       int64          `json:"user_id"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	References   any            `json:"references"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}

type CreateUserFormParams struct {
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
}

type UpdateUserFormParams struct {
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
	UserID       string `json:"user_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateUserParams(a CreateUserFormParams) CreateUserParams {
	return CreateUserParams{
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Si(a.Role),
	}
}

func MapUpdateUserParams(a UpdateUserFormParams) UpdateUserParams {
	return UpdateUserParams{
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Si(a.Role),
		UserID:       Si(a.UserID),
	}
}

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

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapUser(a mdb.Users) Users {
	return Users{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		References:   a.References,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateUserParams(a CreateUserParams) mdb.CreateUserParams {
	return mdb.CreateUserParams{
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
	}
}

func (d Database) MapUpdateUserParams(a UpdateUserParams) mdb.UpdateUserParams {
	return mdb.UpdateUserParams{
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		UserID:       a.UserID,
	}
}

// /QUERIES
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

func (d Database) DeleteUser(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUser(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}
	return nil
}

func (d Database) GetUser(id int64) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUser(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}
func (d Database) GetUserByEmail(email string) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, email)
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

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapUser(a mdbm.Users) Users {
	return Users{
		UserID:       int64(a.UserID),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int64(a.Role.Int32),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: Ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapCreateUserParams(a CreateUserParams) mdbm.CreateUserParams {
	return mdbm.CreateUserParams{
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Ni32(a.Role),
	}
}

func (d MysqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbm.UpdateUserParams {
	return mdbm.UpdateUserParams{
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Ni32(a.Role),
		UserID:       int32(a.UserID),
	}
}

// /QUERIES
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

func (d MysqlDatabase) DeleteUser(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetUser(id int64) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUser(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}
func (d MysqlDatabase) GetUserByEmail(email string) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, email)
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

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapUser(a mdbp.Users) Users {
	return Users{
		UserID:       int64(a.UserID),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int64(a.Role.Int32),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: Ns(nt(a.DateModified)),
	}
}

func (d PsqlDatabase) MapCreateUserParams(a CreateUserParams) mdbp.CreateUserParams {
	return mdbp.CreateUserParams{
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Ni32(a.Role),
	}
}

func (d PsqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbp.UpdateUserParams {
	return mdbp.UpdateUserParams{
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Ni32(a.Role),
		UserID:       int32(a.UserID),
	}
}

// /QUERIES
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

func (d PsqlDatabase) DeleteUser(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUser(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete User: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetUser(id int64) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUser(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}
func (d PsqlDatabase) GetUserByEmail(email string) (*Users, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserByEmail(d.Context, email)
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
