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

func (d Database) CreateToken(s CreateTokenParams) Tokens {
	params := d.MapCreateTokenParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateToken(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateToken: %v\n", err)
	}
	return d.MapToken(row)
}

func (d Database) DeleteToken(id string) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteToken(d.Context, mdb.DeleteTokenParams{ID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}
	return nil
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

func (d Database) UpdateToken(s UpdateTokenParams) (*string, error) {
	params := d.MapUpdateTokenParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateToken(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update token, %v", err)
	}
	u := fmt.Sprintf("Successfully updated token %v\n", s.Token)
	return &u, nil
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

func (d MysqlDatabase) CreateToken(s CreateTokenParams) Tokens {
	params := d.MapCreateTokenParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateToken(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateToken: %v\n", err)
	}
	row, err := queries.GetToken(d.Context, mdbm.GetTokenParams{ID: params.ID})
	if err != nil {
		fmt.Printf("Failed to get last inserted Token: %v\n", err)
	}
	return d.MapToken(row)
}

func (d MysqlDatabase) DeleteToken(id string) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteToken(d.Context, mdbm.DeleteTokenParams{ID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}
	return nil
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

func (d MysqlDatabase) UpdateToken(s UpdateTokenParams) (*string, error) {
	params := d.MapUpdateTokenParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateToken(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update token, %v", err)
	}
	u := fmt.Sprintf("Successfully updated token %v\n", s.Token)
	return &u, nil
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

func (d PsqlDatabase) CreateToken(s CreateTokenParams) Tokens {
	params := d.MapCreateTokenParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateToken(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateToken: %v\n", err)
	}
	return d.MapToken(row)
}

func (d PsqlDatabase) DeleteToken(id string) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteToken(d.Context, mdbp.DeleteTokenParams{ID: id})
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}
	return nil
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

func (d PsqlDatabase) UpdateToken(s UpdateTokenParams) (*string, error) {
	params := d.MapUpdateTokenParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateToken(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update token, %v", err)
	}
	u := fmt.Sprintf("Successfully updated token %v\n", s.Token)
	return &u, nil
}
