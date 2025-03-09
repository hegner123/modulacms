package db

import (
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func (d Database) ListAdminContentData() (*[]AdminContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentData{}
	for _, v := range rows {
		m := d.MapAdminContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminContentFields() (*[]AdminContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []AdminContentFields{}
	for _, v := range rows {
		m := d.MapAdminContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Datatypes: %v\n", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminFields() (*[]AdminFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Fields: %v\n", err)
	}
	res := []AdminFields{}
	for _, v := range rows {
		m := d.MapAdminField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Admin Routes: %v\n", err)
	}
	res := []AdminRoutes{}
	for _, v := range rows {
		m := d.MapAdminRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentData() (*[]ContentData, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentData(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []ContentData{}
	for _, v := range rows {
		m := d.MapContentData(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListContentFields() (*[]ContentFields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListContentFields(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []ContentFields{}
	for _, v := range rows {
		m := d.MapContentField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListDatatypes() (*[]Datatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatype(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Datatypes: %v\n", err)
	}
	res := []Datatypes{}
	for _, v := range rows {
		m := d.MapDatatype(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListFields() (*[]Fields, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListField(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Fields: %v\n", err)
	}
	res := []Fields{}
	for _, v := range rows {
		m := d.MapField(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListMedia() (*[]Media, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMedia(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Medias: %v\n", err)
	}
	res := []Media{}
	for _, v := range rows {
		m := d.MapMedia(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListMediaDimensions() (*[]MediaDimensions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListMediaDimension(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get MediaDimensions: %v\n", err)
	}
	res := []MediaDimensions{}
	for _, v := range rows {
		m := d.MapMediaDimension(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListRoles() (*[]Roles, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRole(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	res := []Roles{}
	for _, v := range rows {
		m := d.MapRoles(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListRoutes() (*[]Routes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListRoute(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Routes: %v\n", err)
	}
	res := []Routes{}
	for _, v := range rows {
		m := d.MapRoute(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListTables() (*[]Tables, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListTable(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Tables: %v\n", err)
	}
	res := []Tables{}
	for _, v := range rows {
		m := d.MapTables(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListTokens() (*[]Tokens, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListTokens(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Tokens{}
	for _, v := range rows {
		m := d.MapToken(v)
		res = append(res, m)
	}
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

func (d Database) ListTokenDependencies(id int64) {
	// TODO implement dependency checking for delete candidate
}

func (d Database) ListDatatypeById(routeId int64) (*[]ListDatatypeByRouteIdRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListDatatypeByRouteId(d.Context, ni64(routeId))
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []ListDatatypeByRouteIdRow{}
	for _, v := range rows {
		m := d.MapListDatatypeByRouteIdRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListFieldByRouteId(routeId int64) (*[]ListFieldByRouteIdRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListFieldByRouteId(d.Context, ni64(routeId))
	if err != nil {
		return nil, fmt.Errorf("failed to get fields with route %d: %v\n", routeId, err)
	}
	res := []ListFieldByRouteIdRow{}
	for _, v := range rows {
		m := d.MapListFieldByRouteIdRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminFieldsByDatatypeID(admin_datatype_id int64) (*[]ListAdminFieldsByDatatypeIDRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminFieldsByDatatypeID(d.Context, ni64(admin_datatype_id))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminFields by AdminDatatypes id: %v\n ", err)
	}
	res := []ListAdminFieldsByDatatypeIDRow{}
	for _, v := range rows {
		m := d.MapListAdminFieldsByDatatypeIDRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypeByAdminRouteId(adminRouteId int64) (*[]ListAdminDatatypeByRouteIdRow, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeByRouteId(d.Context, ni64(adminRouteId))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes by AdminRouteId %v\n", err)
	}
	res := []ListAdminDatatypeByRouteIdRow{}
	for _, v := range rows {
		m := d.MapListAdminDatatypeByRouteIdRow(v)
		res = append(res, m)
	}
	return &res, nil
}

func (d Database) ListAdminDatatypeChildren(parentId int64) (*[]AdminDatatypes, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListAdminDatatypeChildren(d.Context, ni64(parentId))
	if err != nil {
		return nil, fmt.Errorf("failed to get AdminDatatypes by AdminRouteId %v\n", err)
	}
	res := []AdminDatatypes{}
	for _, v := range rows {
		m := d.MapAdminDatatype(v)
		res = append(res, m)
	}
	return &res, nil

}
func (d Database) ListUserOauths() (*[]UserOauth, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListUserOauth(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []UserOauth{}
	for _, v := range rows {
		m := d.MapUserOauth(v)
		res = append(res, m)
	}
	return &res, nil
}
func (d Database) ListSessions() (*[]Sessions, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListSessions(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get Users: %v\n", err)
	}
	res := []Sessions{}
	for _, v := range rows {
		m := d.MapSession(v)
		res = append(res, m)
	}
	return &res, nil
}
