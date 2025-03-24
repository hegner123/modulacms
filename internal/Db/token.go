package db

import (
	"database/sql"
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/db-mysql"
	mdbp "github.com/hegner123/modulacms/db-psql"
	mdb "github.com/hegner123/modulacms/db-sqlite"
)

///////////////////////////////
//STRUCTS
//////////////////////////////
type Tokens struct {
	ID        int64        `json:"id"`
	UserID    int64        `json:"user_id"`
	TokenType string       `json:"token_type"`
	Token     string       `json:"token"`
	IssuedAt  string       `json:"issued_at"`
	ExpiresAt string       `json:"expires_at"`
	Revoked   sql.NullBool `json:"revoked"`
}

type CreateTokenParams struct {
	UserID    int64        `json:"user_id"`
	TokenType string       `json:"token_type"`
	Token     string       `json:"token"`
	IssuedAt  string       `json:"issued_at"`
	ExpiresAt string       `json:"expires_at"`
	Revoked   sql.NullBool `json:"revoked"`
}

type UpdateTokenParams struct {
	Token     string       `json:"token"`
	IssuedAt  string       `json:"issued_at"`
	ExpiresAt string       `json:"expires_at"`
	Revoked   sql.NullBool `json:"revoked"`
	ID        int64        `json:"id"`
}

type TokensHistoryEntry struct {
	ID        int64        `json:"id"`
	UserID    int64        `json:"user_id"`
	TokenType string       `json:"token_type"`
	Token     string       `json:"token"`
	IssuedAt  string       `json:"issued_at"`
	ExpiresAt string       `json:"expires_at"`
	Revoked   sql.NullBool `json:"revoked"`
}

type CreateTokenFormParams struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
	Revoked   string `json:"revoked"`
}

type UpdateTokenFormParams struct {
	Token     string `json:"token"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
	Revoked   string `json:"revoked"`
	ID        string `json:"id"`
}

///////////////////////////////
//GENERIC
//////////////////////////////

func MapCreateTokenParams(a CreateTokenFormParams) CreateTokenParams {
	revoked := false
	if a.Revoked == "true" {
		revoked = true
	}
	return CreateTokenParams{
		UserID:    Si(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   sql.NullBool{Bool: revoked, Valid: true},
	}
}

func MapUpdateTokenParams(a UpdateTokenFormParams) UpdateTokenParams {
	revoked := false
	if a.Revoked == "true" {
		revoked = true
	}
	return UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   sql.NullBool{Bool: revoked, Valid: true},
		ID:        Si(a.ID),
	}
}

func MapStringToken(a Tokens) StringTokens {
	return StringTokens{
		ID:        strconv.FormatInt(a.ID, 10),
		UserID:    strconv.FormatInt(a.UserID, 10),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   strconv.FormatBool(a.Revoked.Bool),
	}
}

///////////////////////////////
//SQLITE
//////////////////////////////

///MAPS
func (d Database) MapToken(a mdb.Tokens) Tokens {
	return Tokens{
		ID:        a.ID,
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

func (d Database) MapCreateTokenParams(a CreateTokenParams) mdb.CreateTokenParams {
	return mdb.CreateTokenParams{
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

func (d Database) MapUpdateTokenParams(a UpdateTokenParams) mdb.UpdateTokenParams {
	return mdb.UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
		ID:        a.ID,
	}
}

///QUERIES
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

func (d Database) DeleteToken(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteToken(d.Context, int64(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}
	return nil
}

func (d Database) GetToken(id int64) (*Tokens, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetToken(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

func (d Database) GetTokenByUserId(userID int64) (*[]Tokens, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.GetTokenByUserId(d.Context, userID)
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
//MYSQL
//////////////////////////////

///MAPS
func (d MysqlDatabase) MapToken(a mdbm.Tokens) Tokens {
	return Tokens{
		ID:        int64(a.ID),
		UserID:    int64(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt.String(),
		ExpiresAt: a.ExpiresAt.String(),
		Revoked:   a.Revoked,
	}
}

func (d MysqlDatabase) MapCreateTokenParams(a CreateTokenParams) mdbm.CreateTokenParams {
	return mdbm.CreateTokenParams{
		UserID:    int32(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: StringToNTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
	}
}

func (d MysqlDatabase) MapUpdateTokenParams(a UpdateTokenParams) mdbm.UpdateTokenParams {
	return mdbm.UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: StringToNTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
		ID:        int32(a.ID),
	}
}

///QUERIES
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
	row, err := queries.GetLastToken(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Token: %v\n", err)
	}
	return d.MapToken(row)
}

func (d MysqlDatabase) DeleteToken(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteToken(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetToken(id int64) (*Tokens, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetToken(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

func (d MysqlDatabase) GetTokenByUserId(userID int64) (*[]Tokens, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.GetTokenByUserId(d.Context, int32(userID))
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
//POSTGRES
//////////////////////////////

///MAPS
func (d PsqlDatabase) MapToken(a mdbp.Tokens) Tokens {
	return Tokens{
		ID:        int64(a.ID),
		UserID:    int64(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt.String(),
		ExpiresAt: a.ExpiresAt.String(),
		Revoked:   a.Revoked,
	}
}

func (d PsqlDatabase) MapCreateTokenParams(a CreateTokenParams) mdbp.CreateTokenParams {
	return mdbp.CreateTokenParams{
		UserID:    int32(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: StringToNTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
	}
}

func (d PsqlDatabase) MapUpdateTokenParams(a UpdateTokenParams) mdbp.UpdateTokenParams {
	return mdbp.UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  StringToNTime(a.IssuedAt).Time,
		ExpiresAt: StringToNTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
		ID:        int32(a.ID),
	}
}

///QUERIES
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

func (d PsqlDatabase) DeleteToken(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteToken(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Token: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetToken(id int64) (*Tokens, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetToken(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapToken(row)
	return &res, nil
}

func (d PsqlDatabase) GetTokenByUserId(userID int64) (*[]Tokens, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.GetTokenByUserId(d.Context, int32(userID))
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
