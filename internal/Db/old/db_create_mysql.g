package db

import (
	_ "embed"
	"fmt"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)



func (d MysqlDatabase) CreateContentData(s CreateContentDataParams) ContentData {
	params := d.MapCreateContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentData(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentData: %v\n", err)
	}
	row, err := queries.GetLastContentData(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted ContentData: %v\n", err)
	}
	return d.MapContentData(row)
}

func (d MysqlDatabase) CreateContentField(s CreateContentFieldParams) ContentFields {
	params := d.MapCreateContentFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateContentField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateContentField: %v\n", err)
	}
	row, err := queries.GetLastContentField(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted ContentField: %v\n", err)
	}
	return d.MapContentField(row)
}

func (d MysqlDatabase) CreateDatatype(s CreateDatatypeParams) Datatypes {
	params := d.MapCreateDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateDatatype(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateDatatype: %v\n", err)
	}
	row, err := queries.GetLastDatatype(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Datatype: %v\n", err)
	}
	return d.MapDatatype(row)
}

func (d MysqlDatabase) CreateField(s CreateFieldParams) Fields {
	params := d.MapCreateFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateField(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateField: %v\n", err)
	}
	row, err := queries.GetLastField(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Field: %v\n", err)
	}
	return d.MapField(row)
}

func (d MysqlDatabase) CreateMedia(s CreateMediaParams) Media {
	params := d.MapCreateMediaParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateMedia(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMedia: %v\n", err)
	}
	row, err := queries.GetLastMedia(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Media: %v\n", err)
	}
	return d.MapMedia(row)
}

func (d MysqlDatabase) CreateMediaDimension(s CreateMediaDimensionParams) MediaDimensions {
	params := d.MapCreateMediaDimensionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateMediaDimension(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateMediaDimension: %v\n", err)
	}
	row, err := queries.GetLastMediaDimension(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted MediaDimension: %v\n", err)
	}
	return d.MapMediaDimension(row)
}

func (d MysqlDatabase) CreatePermission(s CreatePermissionParams) Permissions {
	params := d.MapCreatePermissionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreatePermission(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreatePermission.\n %v\n", err)
	}
	row, err := queries.GetLastPermission(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Permissions: %v\n", err)
	}

	return d.MapPermissions(row)
}

func (d MysqlDatabase) CreateRole(s CreateRoleParams) Roles {
	params := d.MapCreateRoleParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateRole(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRole: %v\n", err)
	}
	row, err := queries.GetLastRole(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Role: %v\n", err)
	}
	return d.MapRoles(row)
}

func (d MysqlDatabase) CreateRoute(s CreateRouteParams) Routes {
	params := d.MapCreateRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateRoute(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateRoute: %v\n", err)
	}
	row, err := queries.GetLastRoute(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Route: %v\n", err)
	}
	return d.MapRoute(row)
}

func (d MysqlDatabase) CreateTable(s string) Tables {
	params := ns(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateTable(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateTable: %v\n", err)
	}
	row, err := queries.GetLastTable(d.Context)
	if err != nil {
		fmt.Printf("Failed to get last inserted Table: %v\n", err)
	}
	return d.MapTables(row)
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

func (d MysqlDatabase) CreateUser(s CreateUserParams) (*Users, error) {
	params := d.MapCreateUserParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateUser(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateUser: %v\n", err)
	}
	row, err := queries.GetLastUser(d.Context)
	if err != nil {
		e := fmt.Errorf("Failed to get last inserted User: %v\n", err)
		return nil, e
	}
	u := d.MapUser(row)
	return &u, nil
}
func (d MysqlDatabase) CreateUserOauth(s CreateUserOauthParams) (*UserOauth, error) {
	params := d.MapCreateUserOauthParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateUserOauth(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateUserOauth.\n %v\n", err)
		return nil, e
	}
	row, err := queries.GetLastUserOauth(d.Context)
	if err != nil {
		e := fmt.Errorf("Failed to get last inserted UserOauth: %v\n", err)
		return nil, e
	}
	u := d.MapUserOauth(row)
	return &u, nil
}
func (d MysqlDatabase) CreateSession(s CreateSessionParams) (*Sessions, error) {
	params := d.MapCreateSessionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateSession(d.Context, params)
	if err != nil {
		e := fmt.Errorf("Failed to CreateSession.\n %v\n", err)
		return nil, e
	}
	row, err := queries.GetLastSession(d.Context)
	if err != nil {
		e := fmt.Errorf("Failed to get last inserted Session: %v\n", err)
		return nil, e
	}
	u := d.MapSession(row)
	return &u, nil
}
