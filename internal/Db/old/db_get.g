package db

import (
	_ "embed"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)



func (d Database) GetContentData(id int64) (*ContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentData(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

func (d Database) GetContentField(id int64) (*ContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetContentField(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d Database) GetDatatype(id int64) (*Datatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d Database) GetField(id int64) (*Fields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetField(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}

func (d Database) GetMedia(id int64) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMedia(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) GetMediaByName(name string) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, Ns(name))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) GetMediaByURL(url string) (*Media, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, Ns(url))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d Database) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
	return &res, nil
}

func (d Database) GetPermission(id int64) (*Permissions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetPermission(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapPermission(row)
	return &res, nil
}

func (d Database) GetRole(id int64) (*Roles, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRole(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapRoles(row)
	return &res, nil
}

func (d Database) GetRoute(slug string) (*Routes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRoute(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}

func (d Database) GetRouteID(slug string) (*int64, error) {
	queries := mdb.New(d.Connection)
	id, err := queries.GetRouteID(d.Context, slug)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (d Database) GetTable(id int64) (*Tables, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTable(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapTables(row)
	return &res, nil
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

func (d Database) GetTokenByUserId(userId int64) (*[]Tokens, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetTokensByUserId(d.Context, userId)
	if err != nil {
		return nil, err
	}
	res := []Tokens{}
	for _, v := range row {
		m := d.MapToken(v)
		res = append(res, m)
	}
	return &res, nil
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

func (d Database) GetSession(id int64) (*Sessions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetSession(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}

func (d Database) GetSessionsByUserId(id int64) (*[]Sessions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetSessionsByUserId(d.Context, id)
	if err != nil {
		return nil, err
	}
	sessions := []Sessions{}
	for _, v := range row {
		s := d.MapSession(v)
		sessions = append(sessions, s)
	}
	return &sessions, nil
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
