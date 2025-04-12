package db

import (
	_ "embed"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
)



func (d MysqlDatabase) GetContentData(id int64) (*ContentData, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentData(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

func (d MysqlDatabase) GetContentField(id int64) (*ContentFields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetContentField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d MysqlDatabase) GetDatatype(id int64) (*Datatypes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d MysqlDatabase) GetField(id int64) (*Fields, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}

func (d MysqlDatabase) GetMedia(id int64) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMedia(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) GetMediaByName(name string) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, Ns(name))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) GetMediaByURL(url string) (*Media, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, Ns(url))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d MysqlDatabase) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
	return &res, nil
}

func (d MysqlDatabase) GetPermission(id int64) (*Permissions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetPermission(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapPermissions(row)
	return &res, nil
}

func (d MysqlDatabase) GetRole(id int64) (*Roles, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRole(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapRoles(row)
	return &res, nil
}

func (d MysqlDatabase) GetRoute(slug string) (*Routes, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetRoute(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}

func (d MysqlDatabase) GetRouteID(slug string) (*int64, error) {
	queries := mdbm.New(d.Connection)
	id, err := queries.GetRouteID(d.Context, slug)
	if err != nil {
		return nil, err
	}
	id64 := int64(id)
	return &id64, nil
}

func (d MysqlDatabase) GetTable(id int64) (*Tables, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTable(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapTables(row)
	return &res, nil
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

func (d MysqlDatabase) GetTokenByUserId(userId int64) (*[]Tokens, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetTokensByUserId(d.Context, int32(userId))
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
func (d MysqlDatabase) GetSession(id int64) (*Sessions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetSession(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}
func (d MysqlDatabase) GetSessionsByUserId(id int64) (*[]Sessions, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetSessionsByUserId(d.Context, int32(id))
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

func (d MysqlDatabase) GetUserOauth(id int64) (*UserOauth, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}
