package db

import mdb "github.com/hegner123/modulacms/db-sqlite"

// Sqlite





func (d Database) MapContentData(a mdb.ContentData) ContentData {
	return ContentData{
		ContentDataID: a.ContentDataID,
		DatatypeID:    a.DatatypeID,
		History:       a.History,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

func (d Database) MapContentField(a mdb.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d Database) MapDatatype(a mdb.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}

func (d Database) MapField(a mdb.Fields) Fields {
	return Fields{
		FieldID:      a.FieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}

func (d Database) MapMedia(a mdb.Media) Media {
	return Media{
		MediaID:      a.MediaID,
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Url:          a.Url,
		Srcset:       a.Srcset,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapMediaDimension(a mdb.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        a.MdID,
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
	}

}
func (d Database) MapPermission(a mdb.Permissions) Permissions {
	return Permissions{
		PermissionID: a.PermissionID,
		TableID:      a.TableID,
		Mode:         a.Mode,
		Label:        a.Label,
	}
}

func (d Database) MapRoles(a mdb.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func (d Database) MapRoute(a mdb.Routes) Routes {
	return Routes{
		RouteID:      a.RouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapTables(a mdb.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
	}
}
func (d Database) MapSession(a mdb.Sessions) Sessions {
	return Sessions{
		SessionID:   AssertInt64(a.SessionID),
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  a.LastAccess,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d Database) MapToken(a mdb.Tokens) Tokens {
	return Tokens{
		ID:        a.ID,
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

func (d Database) MapUser(a mdb.Users) Users {
	return Users{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		References:   a.References,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}
func (d Database) MapUserOauth(a mdb.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         a.UserOauthID,
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated,
	}
}

func (d Database) MapListAdminDatatypeTreeRow(a mdb.ListAdminDatatypeTreeRow) ListAdminDatatypeTreeRow {
	return ListAdminDatatypeTreeRow{
		ChildID:     a.ChildID,
		ChildLabel:  a.ChildLabel,
		ParentID:    a.ParentID,
		ParentLabel: a.ParentLabel,
	}
}

func (d Database) MapListAdminFieldsByDatatypeIDRow(a mdb.ListAdminFieldsByDatatypeIDRow) ListAdminFieldsByDatatypeIDRow {
	return ListAdminFieldsByDatatypeIDRow{
		AdminFieldID: a.AdminFieldID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		History:      a.History,
	}
}





func (d Database) MapCreateContentDataParams(a CreateContentDataParams) mdb.CreateContentDataParams {
	return mdb.CreateContentDataParams{
		RouteID:      a.RouteID,
		DatatypeID:   a.DatatypeID,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateContentFieldParams(a CreateContentFieldParams) mdb.CreateContentFieldParams {
	return mdb.CreateContentFieldParams{
		RouteID:        a.RouteID,
		ContentFieldID: a.ContentFieldID,
		ContentDataID:  a.ContentDataID,
		FieldID:        a.FieldID,
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d Database) MapCreateDatatypeParams(a CreateDatatypeParams) mdb.CreateDatatypeParams {
	return mdb.CreateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateFieldParams(a CreateFieldParams) mdb.CreateFieldParams {
	return mdb.CreateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateMediaParams(a CreateMediaParams) mdb.CreateMediaParams {
	return mdb.CreateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Url:          a.Url,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdb.CreateMediaDimensionParams {
	return mdb.CreateMediaDimensionParams{
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
	}
}
func (d Database) MapCreatePermissionParams(a CreatePermissionParams) mdb.CreatePermissionParams {
	return mdb.CreatePermissionParams{
		TableID: a.TableID,
		Label:   a.Label,
		Mode:    a.Mode,
	}
}

func (d Database) MapCreateRoleParams(a CreateRoleParams) mdb.CreateRoleParams {
	return mdb.CreateRoleParams{
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func (d Database) MapCreateRouteParams(a CreateRouteParams) mdb.CreateRouteParams {
	return mdb.CreateRouteParams{
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d Database) MapCreateSessionParams(a CreateSessionParams) mdb.CreateSessionParams {
	return mdb.CreateSessionParams{
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  a.LastAccess,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d Database) MapCreateTokenParams(a CreateTokenParams) mdb.CreateTokenParams {
	return mdb.CreateTokenParams{
		UserID:    a.UserID,
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
	}
}

func (d Database) MapCreateUserParams(a CreateUserParams) mdb.CreateUserParams {
	return mdb.CreateUserParams{
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
	}
}
func (d Database) MapCreateUserOauthParams(a CreateUserOauthParams) mdb.CreateUserOauthParams {
	return mdb.CreateUserOauthParams{
		UserID:              a.UserID,
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      a.TokenExpiresAt,
		DateCreated:         a.DateCreated,
	}
}






func (d Database) MapUpdateContentDataParams(a UpdateContentDataParams) mdb.UpdateContentDataParams {
	return mdb.UpdateContentDataParams{
		RouteID:       a.RouteID,
		DatatypeID:    a.DatatypeID,
		History:       a.History,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		ContentDataID: a.ContentDataID,
	}
}

func (d Database) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdb.UpdateContentFieldParams {
	return mdb.UpdateContentFieldParams{
		RouteID:          a.RouteID,
		ContentFieldID:   a.ContentFieldID,
		ContentDataID:    a.ContentDataID,
		FieldID:          a.FieldID,
		FieldValue:       a.FieldValue,
		History:          a.History,
		DateCreated:      a.DateCreated,
		DateModified:     a.DateModified,
		ContentFieldID_2: a.ContentFieldID_2,
	}
}

func (d Database) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdb.UpdateDatatypeParams {
	return mdb.UpdateDatatypeParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		DatatypeID:   a.DatatypeID,
	}
}

func (d Database) MapUpdateFieldParams(a UpdateFieldParams) mdb.UpdateFieldParams {
	return mdb.UpdateFieldParams{
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
		FieldID:      a.FieldID,
	}
}

func (d Database) MapUpdateMediaParams(a UpdateMediaParams) mdb.UpdateMediaParams {
	return mdb.UpdateMediaParams{
		Name:         a.Name,
		DisplayName:  a.DisplayName,
		Alt:          a.Alt,
		Caption:      a.Caption,
		Description:  a.Description,
		Class:        a.Class,
		Url:          a.Url,
		Mimetype:     a.Mimetype,
		Dimensions:   a.Dimensions,
		Srcset:       a.Srcset,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		MediaID:      a.MediaID,
	}
}

func (d Database) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdb.UpdateMediaDimensionParams {
	return mdb.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       a.Width,
		Height:      a.Height,
		AspectRatio: a.AspectRatio,
		MdID:        a.MdID,
	}
}

func (d Database) MapUpdatePermissionParams(a UpdatePermissionParams) mdb.UpdatePermissionParams {
	return mdb.UpdatePermissionParams{
		TableID: a.TableID,
		Label:   a.Label,
		Mode:    a.Mode,
	}
}

func (d Database) MapUpdateRoleParams(a UpdateRoleParams) mdb.UpdateRoleParams {
	return mdb.UpdateRoleParams{
		Label:       a.Label,
		Permissions: a.Permissions,
		RoleID:      a.RoleID,
	}
}

func (d Database) MapUpdateRouteParams(a UpdateRouteParams) mdb.UpdateRouteParams {
	return mdb.UpdateRouteParams{
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}

func (d Database) MapUpdateSessionParams(a UpdateSessionParams) mdb.UpdateSessionParams {
	return mdb.UpdateSessionParams{
		UserID:      a.UserID,
		CreatedAt:   a.CreatedAt,
		ExpiresAt:   a.ExpiresAt,
		LastAccess:  a.LastAccess,
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
		SessionID:   Si(a.SessionID),
	}
}

func (d Database) MapUpdateTableParams(a UpdateTableParams) mdb.UpdateTableParams {
	return mdb.UpdateTableParams{
		Label: a.Label,
		ID:    a.ID,
	}
}

func (d Database) MapUpdateTokenParams(a UpdateTokenParams) mdb.UpdateTokenParams {
	return mdb.UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   a.Revoked,
		ID:        a.ID,
	}
}

func (d Database) MapUpdateUserParams(a UpdateUserParams) mdb.UpdateUserParams {
	return mdb.UpdateUserParams{
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		UserID:       a.UserID,
	}
}
func (d Database) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdb.UpdateUserOauthParams {
	return mdb.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: a.TokenExpiresAt,
		UserOauthID:    a.UserOauthID,
	}
}

func (d Database) MapUtilityGetAdminDatatypesRow(a mdb.UtilityGetAdminDatatypesRow) UtilityGetAdminDatatypesRow {
	return UtilityGetAdminDatatypesRow{
		AdminDatatypeID: a.AdminDatatypeID,
		Label:           a.Label,
	}
}

func (d Database) MapUtilityGetAdminRoutesRow(a mdb.UtilityGetAdminRoutesRow) UtilityGetAdminRoutesRow {
	return UtilityGetAdminRoutesRow{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
	}
}

func (d Database) MapUtilityGetAdminfieldsRow(a mdb.UtilityGetAdminfieldsRow) UtilityGetAdminfieldsRow {
	return UtilityGetAdminfieldsRow{
		AdminFieldID: a.AdminFieldID,
		Label:        a.Label,
	}
}

func (d Database) MapUtilityGetDatatypesRow(a mdb.UtilityGetDatatypesRow) UtilityGetDatatypesRow {
	return UtilityGetDatatypesRow{
		DatatypeID: a.DatatypeID,
		Label:      a.Label,
	}
}

func (d Database) MapUtilityGetFieldsRow(a mdb.UtilityGetFieldsRow) UtilityGetFieldsRow {
	return UtilityGetFieldsRow{
		FieldID: a.FieldID,
		Label:   a.Label,
	}
}

func (d Database) MapUtilityGetMediaRow(a mdb.UtilityGetMediaRow) UtilityGetMediaRow {
	return UtilityGetMediaRow{
		MediaID: a.MediaID,
		Name:    a.Name,
	}
}

func (d Database) MapUtilityGetMediaDimensionRow(a mdb.UtilityGetMediaDimensionRow) UtilityGetMediaDimensionRow {
	return UtilityGetMediaDimensionRow{
		MdID:  a.MdID,
		Label: a.Label,
	}
}

func (d Database) MapUtilityGetRouteRow(a mdb.UtilityGetRouteRow) UtilityGetRouteRow {
	return UtilityGetRouteRow{
		RouteID: a.RouteID,
		Slug:    a.Slug,
	}
}

func (d Database) MapUtilityGetTablesRow(a mdb.UtilityGetTablesRow) UtilityGetTablesRow {
	return UtilityGetTablesRow{
		ID:    a.ID,
		Label: a.Label,
	}
}

func (d Database) MapUtilityGetTokenRow(a mdb.UtilityGetTokenRow) UtilityGetTokenRow {
	return UtilityGetTokenRow{
		ID:     a.ID,
		UserID: a.UserID,
	}
}

func (d Database) MapUtilityGetUsersRow(a mdb.UtilityGetUsersRow) UtilityGetUsersRow {
	return UtilityGetUsersRow{
		UserID:   a.UserID,
		Username: a.Username,
	}
}
