package db

import (
	_ "embed"
	"fmt"

	mdb "github.com/hegner123/modulacms/db-sqlite"
	_ "github.com/mattn/go-sqlite3"
)

func (d Database) UpdateAdminContentData(s UpdateAdminContentDataParams) (*string, error) {
	params := d.MapUpdateAdminContentDataParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.AdminDatatypeID)
	return &u, nil
}
func (d Database) UpdateAdminContentField(s UpdateAdminContentFieldParams) (*string, error) {
	params := d.MapUpdateAdminContentFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.AdminContentFieldID)
	return &u, nil
}

func (d Database) UpdateAdminDatatype(s UpdateAdminDatatypeParams) (*string, error) {
	params := d.MapUpdateAdminDatatypeParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin datatype, %v ", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdateAdminField(s UpdateAdminFieldParams) (*string, error) {
	params := d.MapUpdateAdminFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdateAdminRoute(s UpdateAdminRouteParams) (*string, error) {
	params := d.MapUpdateAdminRouteParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateAdminRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update admin route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}
func (d Database) UpdateContentData(s UpdateContentDataParams) (*string, error) {
	params := d.MapUpdateContentDataParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateContentData(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content Data id %v\n", s.DatatypeID)
	return &u, nil
}
func (d Database) UpdateContentField(s UpdateContentFieldParams) (*string, error) {
	params := d.MapUpdateContentFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateContentField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update content data, %v", err)
	}
	u := fmt.Sprintf("Successfully updated content field id %v\n", s.ContentFieldID)
	return &u, nil
}

func (d Database) UpdateDatatype(s UpdateDatatypeParams) (*string, error) {
	params := d.MapUpdateDatatypeParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateDatatype(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update datatype, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdateField(s UpdateFieldParams) (*string, error) {
	params := d.MapUpdateFieldParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateField(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update field, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdateMedia(s UpdateMediaParams) (*string, error) {
	params := d.MapUpdateMediaParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateMedia(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}

func (d Database) UpdateMediaDimension(s UpdateMediaDimensionParams) (*string, error) {
	params := d.MapUpdateMediaDimensionParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateMediaDimension(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update media dimension, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdatePermission(s UpdatePermissionParams) (*string, error) {
	params := d.MapUpdatePermissionParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdatePermission(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update permision, %v", err)
	}
	u := fmt.Sprintf("Successfully updated permision %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdateRole(s UpdateRoleParams) (*string, error) {
	params := d.MapUpdateRoleParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateRole(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update role, %v", err)
	}
	u := fmt.Sprintf("Successfully updated role %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdateRoute(s UpdateRouteParams) (*string, error) {
	params := d.MapUpdateRouteParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateRoute(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update route, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Slug)
	return &u, nil
}

func (d Database) UpdateTable(s UpdateTableParams) (*string, error) {
	params := d.MapUpdateTableParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateTable(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update table, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Label)
	return &u, nil
}

func (d Database) UpdateToken(s UpdateTokenParams) (*string, error) {
	params := d.MapUpdateTokenParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateToken(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update token, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.ID)
	return &u, nil
}

func (d Database) UpdateUser(s UpdateUserParams) (*string, error) {
	params := d.MapUpdateUserParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateUser(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.Name)
	return &u, nil
}

func (d Database) UpdateUserOauth(s UpdateUserOauthParams) (*string, error) {
	params := d.MapUpdateUserOauthParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateUserOauth(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user oauth, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.UserOauthID)
	return &u, nil
}

func (d Database) UpdateSession(s UpdateSessionParams) (*string, error) {
	params := d.MapUpdateSessionParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateSession(d.Context, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user oauth, %v", err)
	}
	u := fmt.Sprintf("Successfully updated %v\n", s.SessionID)
	return &u, nil
}
