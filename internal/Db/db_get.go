package db

import (
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func (d Database) GetAdminContentData(id int64) (*AdminContentData, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentData(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentData(row)
	return &res, nil
}

func (d Database) GetAdminContentField(id int64) (*AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminContentField(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminContentField(row)
	return &res, nil
}

func (d Database) GetAdminDatatypeGlobalId() (*AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetGlobalAdminDatatypeId(d.Context)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d Database) GetAdminDatatypeById(id int64) (*AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminDatatype(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminDatatype(row)
	return &res, nil
}

func (d Database) GetRootAdIdByAdRtId(adminRtId int64) (*int64, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetRootAdminDtByAdminRtId(d.Context, ni64(adminRtId))
	if err != nil {
		fmt.Printf("adminRtId %d\n", adminRtId)
		return nil, err
	}
	res := row.AdminDatatypeID
	return &res, nil
}

func (d Database) GetAdminField(id int64) (*AdminFields, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminField(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminField(row)
	return &res, nil
}

func (d Database) GetAdminRoute(slug string) (*AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetAdminRouteBySlug(d.Context, slug)
	if err != nil {
		return nil, err
	}
	res := d.MapAdminRoute(row)
	return &res, nil
}

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

func (d Database) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
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
