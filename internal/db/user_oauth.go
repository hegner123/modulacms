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

func (d Database) DeleteUserOauth(id types.UserOauthID) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, mdb.DeleteUserOauthParams{UserOAuthID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}
	return nil
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

func (d MysqlDatabase) CreateUserOauth(s CreateUserOauthParams) (*UserOauth, error) {
	params := d.MapCreateUserOauthParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserOauth(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUserOauth.\n %v\n", err)
		return nil, e
	}
	row, err := queries.GetUserOauth(d.Context, mdbm.GetUserOauthParams{UserOAuthID: params.UserOAuthID})
	if err != nil {
		return nil, fmt.Errorf("Failed to get last inserted UserOauth: %v\n", err)
	}
	userOauth := d.MapUserOauth(row)
	return &userOauth, nil
}

func (d MysqlDatabase) DeleteUserOauth(id types.UserOauthID) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, mdbm.DeleteUserOauthParams{UserOAuthID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}
	return nil
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

func (d PsqlDatabase) DeleteUserOauth(id types.UserOauthID) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteUserOauth(d.Context, mdbp.DeleteUserOauthParams{UserOAuthID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete UserOauth: %v ", id)
	}
	return nil
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
