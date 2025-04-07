package db

import (
	_ "embed"
	"fmt"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
)


func (d MysqlDatabase) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ContentDataID)
	return &u, nil
}
func (d MysqlDatabase) UpdateContentField(s UpdateContentFieldParams) (*string, error) {
	params := d.MapUpdateContentFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ContentFieldID)
	return &u, nil
}

func (d MysqlDatabase) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d MysqlDatabase) UpdateField(s UpdateFieldParams) (*string, error) {
	params := d.MapUpdateFieldParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d MysqlDatabase) UpdateMedia(s UpdateMediaParams) (*string, error) {
	params := d.MapUpdateMediaParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateMedia(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}

func (d MysqlDatabase) UpdateMediaDimension(s UpdateMediaDimensionParams) (*string, error) {
	params := d.MapUpdateMediaDimensionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateMediaDimension(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media dimension, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d MysqlDatabase) UpdatePermission(s UpdatePermissionParams) (*string, error) {
	params := d.MapUpdatePermissionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdatePermission(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update permision, %v", err)
	}
	u := fmt.Sprintf("Successfully updated permision %v\n", s.Label)
	return &u, nil
}
func (d MysqlDatabase) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d MysqlDatabase) UpdateRoute(s UpdateRouteParams) (*string, error) {
	params := d.MapUpdateRouteParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

func (d MysqlDatabase) UpdateTable(s UpdateTableParams) (*string, error) {
	params := d.MapUpdateTableParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateTable(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update table, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d MysqlDatabase) UpdateToken(s UpdateTokenParams) (*string, error) {
	params := d.MapUpdateTokenParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateToken(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update token, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

func (d MysqlDatabase) UpdateUser(s UpdateUserParams) (*string, error) {
	params := d.MapUpdateUserParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateUser(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}
func (d MysqlDatabase) UpdateUserOauth(s UpdateUserOauthParams) (*string, error) {
	params := d.MapUpdateUserOauthParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateUserOauth(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user oauth, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.UserOauthID)
	return &u, nil
}

func (d MysqlDatabase) UpdateSession(s UpdateSessionParams) (*string, error) {
	params := d.MapUpdateSessionParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateSession(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user oauth, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &u, nil
}
