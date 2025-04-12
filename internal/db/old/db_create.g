package db

import (
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)


func (d Database) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminRoute  %v \n", err)
	}

	return d.MapContentData(row)
}

func (d Database) CreateContentField(s CreateContentFieldParams) ContentFields {
	params := d.MapCreateContentFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateContentField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateAdminRoute  %v \n", err)
	}

	return d.MapContentField(row)
}

func (d Database) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype  %v \n", err)
	}

	return d.MapDatatype(row)
}

func (d Database) CreateField(s CreateFieldParams) Fields {
	params := d.MapCreateFieldParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateField  %v \n", err)
	}

	return d.MapField(row)
}

func (d Database) CreateMedia(s CreateMediaParams) Media {
	params := d.MapCreateMediaParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateMedia(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMedia.\n%v \n", err)
	}

	return d.MapMedia(row)
}

func (d Database) CreateMediaDimension(s CreateMediaDimensionParams) MediaDimensions {
	params := d.MapCreateMediaDimensionParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateMediaDimension(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMediaDimension.\n%v \n", err)
	}

	return d.MapMediaDimension(row)
}

func (d Database) CreatePermission(s CreatePermissionParams) Permissions {
	params := d.MapCreatePermissionParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreatePermission(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreatePermission.\n %v\n", err)
	}

	return d.MapPermission(row)
}

func (d Database) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRoute.\n %v\n", err)
	}

	return d.MapRoles(row)
}

func (d Database) CreateRoute(s CreateRouteParams) Routes {
	params := d.MapCreateRouteParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRoute.\n %v\n", err)
	}

	return d.MapRoute(row)
}

func (d Database) CreateTable(s string) Tables {
	params := ns(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateTable(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateTable.\n %v\n", err)
	}

	return d.MapTables(row)
}

func (d Database) CreateToken(s CreateTokenParams) Tokens {
	params := d.MapCreateTokenParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateToken(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateToken.\n %v\n", err)
	}

	return d.MapToken(row)
}

func (d Database) CreateUser(s CreateUserParams) (*Users, error) {
	params := d.MapCreateUserParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateUser(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUser.\n %v\n", err)
		return nil, e
	}
	u := d.MapUser(row)
	return &u, nil
}
func (d Database) CreateUserOauth(s CreateUserOauthParams) (*UserOauth, error) {
	params := d.MapCreateUserOauthParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateUserOauth(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUserOauth.\n %v\n", err)
		return nil, e
	}
	u := d.MapUserOauth(row)
	return &u, nil
}
func (d Database) CreateSession(s CreateSessionParams) (*Sessions, error) {
	params := d.MapCreateSessionParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateSession(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUserOauth.\n %v\n", err)
		return nil, e
	}
	u := d.MapSession(row)
	return &u, nil
}
