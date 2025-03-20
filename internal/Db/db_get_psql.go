package db

import (
	_ "embed"

	mdbp "github.com/hegner123/modulacms/db-psql"
)

func (d PsqlDatabase) GetAdminContentData(id int64) (*AdminContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}

func (d PsqlDatabase) GetAdminContentField(id int64) (*AdminContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}

func (d PsqlDatabase) GetAdminDatatypeGlobalId() (*AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetGlobalAdminDatatypeId(d.Context)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) GetAdminDatatypeById(id int64) (*AdminDatatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) GetAdminField(id int64) (*AdminFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d PsqlDatabase) GetAdminRoute(slug string) (*AdminRoutes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}

func (d PsqlDatabase) GetContentData(id int64) (*ContentData, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentData(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentData(row)
	return &res, nil
}

func (d PsqlDatabase) GetContentField(id int64) (*ContentFields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetContentField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapContentField(row)
	return &res, nil
}

func (d PsqlDatabase) GetDatatype(id int64) (*Datatypes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetDatatype(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapDatatype(row)
	return &res, nil
}

func (d PsqlDatabase) GetField(id int64) (*Fields, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetField(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapField(row)
	return &res, nil
}

func (d PsqlDatabase) GetMedia(id int64) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMedia(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) GetMediaByName(name string) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaByName(d.Context, Ns(name))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) GetMediaByURL(url string) (*Media, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaByUrl(d.Context, Ns(url))
	if err != nil {
		return nil, err
	}
	res := d.MapMedia(row)
	return &res, nil
}

func (d PsqlDatabase) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
	return &res, nil
}

func (d PsqlDatabase) GetPermission(id int64) (*Permissions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetPermission(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapPermissions(row)
	return &res, nil
}
func (d PsqlDatabase) GetRole(id int64) (*Roles, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRole(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapRoles(row)
	return &res, nil
}

func (d PsqlDatabase) GetRoute(slug string) (*Routes, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRoute(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapRoute(row)
	return &res, nil
}
func (d PsqlDatabase) GetRouteID(slug string) (*int64, error) {
	queries := mdbp.New(d.Connection)
	id, err := queries.GetRouteID(d.Context, slug)
	if err != nil {
		return nil, err
	}
	id64 := int64(id)
	return &id64, nil
}

func (d PsqlDatabase) GetTable(id int64) (*Tables, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetTable(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapTables(row)
	return &res, nil
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

func (d PsqlDatabase) GetTokenByUserId(userId int64) (*[]Tokens, error) {
	queries := mdbp.New(d.Connection)
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
func (d PsqlDatabase) GetSession(id int64) (*Sessions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetSession(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapSession(row)
	return &res, nil
}
func (d PsqlDatabase) GetSessionsByUserId(id int64) (*[]Sessions, error) {
	queries := mdbp.New(d.Connection)
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

func (d PsqlDatabase) GetUserOauth(id int64) (*UserOauth, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetUserOauth(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapUserOauth(row)
	return &res, nil
}
