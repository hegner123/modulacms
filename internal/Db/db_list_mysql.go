package db

import (
	_ "embed"
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)

func (d MysqlDatabase) ListAdminDatatypes() (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListAdminFields() (*[]AdminFields, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListAdminRoutes() (*[]AdminRoutes, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListContentData() (*[]ContentData, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListContentFields() (*[]ContentFields, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListContentField(d.Context)
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

func (d MysqlDatabase) ListDatatypes() (*[]Datatypes, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListFields() (*[]Fields, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListMedia() (*[]Media, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListMediaDimensions() (*[]MediaDimensions, error) {
	queries := mdbm.New(d.Connection)
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
func (d MysqlDatabase) ListRoles() (*[]Roles, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListRoutes() (*[]Routes, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListTables() (*[]Tables, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListTokens() (*[]Tokens, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListUsers() (*[]Users, error) {
	queries := mdbm.New(d.Connection)
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

func (d MysqlDatabase) ListTokenDependencies(id int64) {
	// TODO implement dependency checking for delete candidate
}

func (d MysqlDatabase) ListDatatypeById(routeId int64) (*[]ListDatatypeByRouteIdRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListDatatypeByRouteId(d.Context, Ni32(routeId))
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

func (d MysqlDatabase) ListFieldByRouteId(routeId int64) (*[]ListFieldByRouteIdRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListFieldByRouteId(d.Context, Ni32(routeId))
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

func (d MysqlDatabase) ListAdminFieldsByDatatypeID(admin_datatype_id int64) (*[]ListAdminFieldsByDatatypeIDRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminFieldsByDatatypeID(d.Context, Ni32(admin_datatype_id))
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

func (d MysqlDatabase) ListAdminDatatypeByAdminRouteId(adminRouteId int64) (*[]ListAdminDatatypeByRouteIdRow, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeByRouteId(d.Context, Ni32(adminRouteId))
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

func (d MysqlDatabase) ListAdminDatatypeChildren(parentId int64) (*[]AdminDatatypes, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListAdminDatatypeChildren(d.Context, Ni32(parentId))
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
