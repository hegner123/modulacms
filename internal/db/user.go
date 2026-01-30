package db

import (
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
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

func (d Database) DeleteUser(id types.UserID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUser(d.Context, mdb.DeleteUserParams{UserID: id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", id)
	}
	return nil
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

func (d MysqlDatabase) DeleteUser(id types.UserID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUser(d.Context, mdbm.DeleteUserParams{UserID: id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", id)
	}
	return nil
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

func (d PsqlDatabase) DeleteUser(id types.UserID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUser(d.Context, mdbp.DeleteUserParams{UserID: id})
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", id)
	}
	return nil
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
