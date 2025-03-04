package db

import (
	"database/sql"
	_ "embed"
	"fmt"

	mdbp "github.com/hegner123/modulacms/db-psql"
)

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

func (d PsqlDatabase) GetRootAdIdByAdRtId(adminRtId int32) (*int64, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetRootAdminDtByAdminRtId(d.Context, sql.NullInt32{Int32: adminRtId, Valid: true})
	if err != nil {
		fmt.Printf("adminRtId %d\n", adminRtId)
		return nil, err
	}
    res := int64(row.AdminDtID)
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

func (d PsqlDatabase) GetMediaDimension(id int64) (*MediaDimensions, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetMediaDimension(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapMediaDimension(row)
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
