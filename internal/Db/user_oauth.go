package db

import (
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

// /////////////////////////////
// STRUCTS
// ////////////////////////////
type UserOauth struct {
	UserOauthID         int64  `json:"user_oauth_id"`
	UserID              int64  `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}

type CreateUserOauthParams struct {
	UserID              int64  `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}

type UpdateUserOauthParams struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	TokenExpiresAt string `json:"token_expires_at"`
	UserOauthID    int64  `json:"user_oauth_id"`
}

type UserOauthHistoryEntry struct {
	UserOauthID         int64  `json:"user_oauth_id"`
	UserID              int64  `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}

type CreateUserOauthFormParams struct {
	UserID              string `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}

type UpdateUserOauthFormParams struct {
	AccessToken    string `json:"access_token"`
	RefreshToken   string `json:"refresh_token"`
	TokenExpiresAt string `json:"token_expires_at"`
	UserOauthID    string `json:"user_oauth_id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateUserOauthParams(a CreateUserOauthFormParams) CreateUserOauthParams {
	return CreateUserOauthParams{
		UserID:              StringToInt64(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated,
	}
}

func MapUpdateUserOauthParams(a UpdateUserOauthFormParams) UpdateUserOauthParams {
	return UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: a.TokenExpiresAt,
		UserOauthID:    StringToInt64(a.UserOauthID),
	}
}

func MapStringUserOauth(a UserOauth) StringUserOauth {
	return StringUserOauth{
		UserOauthID:         strconv.FormatInt(a.UserOauthID, 10),
		UserID:              strconv.FormatInt(a.UserID, 10),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated,
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

// /MAPS
func (d Database) MapUserOauth(a mdb.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         a.UserOauthID,
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated,
	}
}

func (d Database) MapCreateUserOauthParams(a CreateUserOauthParams) mdb.CreateUserOauthParams {
	return mdb.CreateUserOauthParams{
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
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
		UserOauthID:    a.UserOauthID,
	}
}

// /QUERIES
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

func (d Database) CreateUserOauth(s CreateUserOauthParams) (*UserOauth, error) {
	params := d.MapCreateUserOauthParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateUserOauth(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUserOauth.\n %v\n", err)
		return nil, e
	}
	userOauth := d.MapUserOauth(row)
	return &userOauth, nil
}

func (d Database) DeleteUserOauth(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}
	return nil
}

func (d Database) GetUserOauth(id int64) (*UserOauth, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, id)
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

func (d Database) UpdateUserOauth(s UpdateUserOauthParams) (*string, error) {
	params := d.MapUpdateUserOauthParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateUserOauth(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user oauth, %v", err)
	}
	u := fmt.Sprintf("Successfully updated user oauth %v\n", s.UserOauthID)
	return &u, nil
}

///////////////////////////////
//MYSQL
//////////////////////////////

// /MAPS
func (d MysqlDatabase) MapUserOauth(a mdbm.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         int64(a.UserOauthID),
		UserID:              int64(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt.String(),
		DateCreated:         a.DateCreated.String(),
	}
}

func (d MysqlDatabase) MapCreateUserOauthParams(a CreateUserOauthParams) mdbm.CreateUserOauthParams {
	return mdbm.CreateUserOauthParams{
		UserID:              int32(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      ParseTime(a.TokenExpiresAt),
		DateCreated:         ParseTime(a.DateCreated),
	}
}

func (d MysqlDatabase) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdbm.UpdateUserOauthParams {
	return mdbm.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: ParseTime(a.TokenExpiresAt),
		UserOauthID:    int32(a.UserOauthID),
	}
}

// /QUERIES
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

func (d MysqlDatabase) CreateUserOauth(s CreateUserOauthParams) (*UserOauth, error) {
	params := d.MapCreateUserOauthParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserOauth(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUserOauth.\n %v\n", err)
		return nil, e
	}
	row, err := queries.GetLastUserOauth(d.Context)
	if err != nil {
		return nil, fmt.Errorf("Failed to get last inserted UserOauth: %v\n", err)
	}
	userOauth := d.MapUserOauth(row)
	return &userOauth, nil
}

func (d MysqlDatabase) DeleteUserOauth(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetUserOauth(id int64) (*UserOauth, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, int32(id))
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

func (d MysqlDatabase) UpdateUserOauth(s UpdateUserOauthParams) (*string, error) {
	params := d.MapUpdateUserOauthParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateUserOauth(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user oauth, %v", err)
	}
	u := fmt.Sprintf("Successfully updated user oauth %v\n", s.UserOauthID)
	return &u, nil
}

///////////////////////////////
//POSTGRES
//////////////////////////////

// /MAPS
func (d PsqlDatabase) MapUserOauth(a mdbp.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         int64(a.UserOauthID),
		UserID:              int64(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt.String(),
		DateCreated:         a.DateCreated.String(),
	}
}

func (d PsqlDatabase) MapCreateUserOauthParams(a CreateUserOauthParams) mdbp.CreateUserOauthParams {
	return mdbp.CreateUserOauthParams{
		UserID:              int32(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      ParseTime(a.TokenExpiresAt),
		DateCreated:         ParseTime(a.DateCreated),
	}
}

func (d PsqlDatabase) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdbp.UpdateUserOauthParams {
	return mdbp.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: ParseTime(a.TokenExpiresAt),
		UserOauthID:    int32(a.UserOauthID),
	}
}

// /QUERIES
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

func (d PsqlDatabase) CreateUserOauth(s CreateUserOauthParams) (*UserOauth, error) {
	params := d.MapCreateUserOauthParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateUserOauth(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUserOauth.\n %v\n", err)
		return nil, e
	}
	userOauth := d.MapUserOauth(row)
	return &userOauth, nil
}

func (d PsqlDatabase) DeleteUserOauth(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetUserOauth(id int64) (*UserOauth, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, int32(id))
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

func (d PsqlDatabase) UpdateUserOauth(s UpdateUserOauthParams) (*string, error) {
	params := d.MapUpdateUserOauthParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateUserOauth(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user oauth, %v", err)
	}
	u := fmt.Sprintf("Successfully updated user oauth %v\n", s.UserOauthID)
	return &u, nil
}
