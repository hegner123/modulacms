package db

import (
	"encoding/json"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)

// Mysql
func (d MysqlDatabase) MapAdminContentData(a mdbm.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminRouteID:       int64(a.AdminRouteID.Int32),
		AdminContentDataID: int64(a.AdminContentDataID),
		AdminDatatypeID:    int64(a.AdminDatatypeID.Int32),
		History:            a.History,
		DateCreated:        Ns(nt(a.DateCreated)),
		DateModified:       ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapAdminContentField(a mdbm.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminRouteID:        int64(a.AdminRouteID.Int32),
		AdminContentFieldID: int64(a.AdminContentFieldID),
		AdminContentDataID:  int64(a.AdminContentDataID),
		AdminFieldID:        int64(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         Ns(nt(a.DateCreated)),
		DateModified:        ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapAdminDatatype(a mdbm.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: int64(a.AdminDatatypeID),
		ParentID:        Ni64(int64(a.ParentID.Int32)),
		Label:           a.Label,
		Type:            a.Type,
		Author:          a.Author,
		AuthorID:        int64(a.AuthorID),
		DateCreated:     Ns(nt(a.DateCreated)),
		DateModified:    ns(nt(a.DateModified)),
		History:         a.History,
	}
}

func (d MysqlDatabase) MapAdminField(a mdbm.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: int64(a.AdminFieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapAdminRoute(a mdbm.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: int64(a.AdminRouteID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapContentData(a mdbm.ContentData) ContentData {
	return ContentData{
		RouteID:       int64(a.RouteID.Int32),
		ContentDataID: int64(a.ContentDataID),
		DatatypeID:    int64(a.DatatypeID.Int32),
		History:       a.History,
		DateCreated:   Ns(nt(a.DateCreated)),
		DateModified:  ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapContentField(a mdbm.ContentFields) ContentFields {
	return ContentFields{
		RouteID:        int64(a.RouteID.Int32),
		ContentFieldID: int64(a.ContentFieldID),
		ContentDataID:  int64(a.ContentDataID),
		FieldID:        int64(a.FieldID),
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    Ns(nt(a.DateCreated)),
		DateModified:   ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapDatatype(a mdbm.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   int64(a.DatatypeID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapField(a mdbm.Fields) Fields {
	return Fields{
		FieldID:      int64(a.FieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapMedia(a mdbm.Media) Media {
	return Media{
		MediaID:      int64(a.MediaID),
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
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapMediaDimension(a mdbm.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        int64(a.MdID),
		Label:       a.Label,
		Width:       Ni64(int64(a.Width.Int32)),
		Height:      Ni64(int64(a.Height.Int32)),
		AspectRatio: a.AspectRatio,
	}

}
func (d MysqlDatabase) MapPermissions(a mdbm.Permissions) Permissions {
	return Permissions{
		PermissionID: int64(a.PermissionID),
		TableID:      int64(a.TableID),
		Label:        a.Label,
		Mode:         int64(a.Mode),
	}
}

func (d MysqlDatabase) MapRoles(a mdbm.Roles) Roles {
	return Roles{
		RoleID:      int64(a.RoleID),
		Label:       a.Label,
		Permissions: jrS(a.Permissions),
	}
}

func (d MysqlDatabase) MapRoute(a mdbm.Routes) Routes {
	return Routes{
		RouteID:  int64(a.RouteID),
		Slug:     a.Slug,
		Title:    a.Title,
		Status:   int64(a.Status),
		Author:   a.Author,
		AuthorID: int64(a.AuthorID),
		History:  a.History,
	}
}
func (d MysqlDatabase) MapSession(a mdbm.Sessions) Sessions {
	return Sessions{
		SessionID:   int64(a.SessionID),
		UserID:      int64(a.UserID),
		CreatedAt:   Ns(a.CreatedAt.Time.String()),
		ExpiresAt:   Ns(a.ExpiresAt.Time.String()),
		LastAccess:  Ns(a.LastAccess.Time.String()),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d MysqlDatabase) MapTables(a mdbm.Tables) Tables {
	return Tables{
		ID:       int64(a.ID),
		Label:    a.Label,
		AuthorID: int64(a.AuthorID),
	}
}

func (d MysqlDatabase) MapToken(a mdbm.Tokens) Tokens {
	return Tokens{
		ID:        int64(a.ID),
		UserID:    int64(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt.String(),
		ExpiresAt: a.ExpiresAt.String(),
		Revoked:   a.Revoked,
	}
}

func (d MysqlDatabase) MapUser(a mdbm.Users) Users {
	return Users{
		UserID:       int64(a.UserID),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int64(a.Role.Int32),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: ns(nt(a.DateModified)),
	}
}
func (d MysqlDatabase) MapUserOauth(a mdbm.UserOauth) UserOauth {
	return UserOauth{
		UserOauthID:         int64(a.UserOauthID),
		UserID:              int64(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      Ns(a.TokenExpiresAt.Time.String()),
		DateCreated:         Ns(a.DateCreated.Time.String()),
	}
}

func (d MysqlDatabase) MapListAdminFieldsByDatatypeIDRow(a mdbm.ListAdminFieldsByDatatypeIDRow) ListAdminFieldsByDatatypeIDRow {
	return ListAdminFieldsByDatatypeIDRow{
		AdminFieldID: int64(a.AdminFieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		History:      a.History,
	}
}

func (d MysqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbm.CreateAdminContentDataParams {
	return mdbm.CreateAdminContentDataParams{
		AdminRouteID:    Ni32(a.AdminRouteID),
		AdminDatatypeID: Ni32(a.AdminDatatypeID),
		History:         a.History,
		DateCreated:     sTime(a.DateCreated.String),
		DateModified:    sTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbm.CreateAdminContentFieldParams {
	return mdbm.CreateAdminContentFieldParams{
		AdminRouteID:        Ni32(a.AdminRouteID),
		AdminContentFieldID: int32(a.AdminContentFieldID),
		AdminContentDataID:  int32(a.AdminContentDataID),
		AdminFieldID:        int32(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         sTime(a.DateCreated.String),
		DateModified:        sTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbm.CreateAdminDatatypeParams {
	return mdbm.CreateAdminDatatypeParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbm.CreateAdminFieldParams {
	return mdbm.CreateAdminFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbm.CreateAdminRouteParams {
	return mdbm.CreateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		History:      a.History,
	}
}

func (d MysqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbm.CreateContentDataParams {
	return mdbm.CreateContentDataParams{
		RouteID:      Ni32(a.RouteID),
		DatatypeID:   Ni32(a.DatatypeID),
		History:      a.History,
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbm.CreateContentFieldParams {
	return mdbm.CreateContentFieldParams{
		RouteID:        Ni32(a.RouteID),
		ContentFieldID: int32(a.ContentFieldID),
		ContentDataID:  int32(a.ContentDataID),
		FieldID:        int32(a.FieldID),
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    sTime(a.DateCreated.String),
		DateModified:   sTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbm.CreateDatatypeParams {
	return mdbm.CreateDatatypeParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		History:      a.History,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbm.CreateFieldParams {
	return mdbm.CreateFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		History:      a.History,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapCreateMediaParams(a CreateMediaParams) mdbm.CreateMediaParams {
	return mdbm.CreateMediaParams{
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
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
	}
}

func (d MysqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbm.CreateMediaDimensionParams {
	return mdbm.CreateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
	}
}
func (d MysqlDatabase) MapCreatePermissionParams(a CreatePermissionParams) mdbm.CreatePermissionParams {
	return mdbm.CreatePermissionParams{
		TableID: int32(a.TableID),
		Label:   a.Label,
		Mode:    int32(a.Mode),
	}
}

func (d MysqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbm.CreateRoleParams {
	return mdbm.CreateRoleParams{
		Label:       a.Label,
		Permissions: json.RawMessage(a.Permissions),
	}
}

func (d MysqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbm.CreateRouteParams {
	return mdbm.CreateRouteParams{
		Author:       a.Author,
		AuthorID:     int32(a.AuthorID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		History:      a.History,
		DateCreated:  NSt(a.DateCreated),
		DateModified: NSt(a.DateModified),
	}
}
func (d MysqlDatabase) MapCreateSessionParams(a CreateSessionParams) mdbm.CreateSessionParams {
	return mdbm.CreateSessionParams{
		UserID:      int32(a.UserID),
		CreatedAt:   sTime(a.CreatedAt.String),
		ExpiresAt:   sTime(a.ExpiresAt.String),
		LastAccess:  sTime(a.LastAccess.String),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d MysqlDatabase) MapCreateTokenParams(a CreateTokenParams) mdbm.CreateTokenParams {
	return mdbm.CreateTokenParams{
		UserID:    int32(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  sTime(a.IssuedAt).Time,
		ExpiresAt: sTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
	}
}

func (d MysqlDatabase) MapCreateUserParams(a CreateUserParams) mdbm.CreateUserParams {
	return mdbm.CreateUserParams{
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Ni32(a.Role),
	}
}
func (d MysqlDatabase) MapCreateUserOauthParams(a CreateUserOauthParams) mdbm.CreateUserOauthParams {
	return mdbm.CreateUserOauthParams{
		UserID:              int32(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      sTime(a.TokenExpiresAt.String),
		DateCreated:         sTime(a.DateCreated.String),
	}
}

func (d MysqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbm.UpdateAdminContentDataParams {
	return mdbm.UpdateAdminContentDataParams{
		AdminRouteID:       Ni32(a.AdminRouteID),
		AdminDatatypeID:    Ni32(a.AdminDatatypeID),
		History:            a.History,
		DateCreated:        sTime(a.DateCreated.String),
		DateModified:       sTime(a.DateModified.String),
		AdminContentDataID: int32(a.AdminContentDataID),
	}
}

func (d MysqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbm.UpdateAdminContentFieldParams {
	return mdbm.UpdateAdminContentFieldParams{
		AdminRouteID:          Ni32(a.AdminRouteID),
		AdminContentFieldID:   int32(a.AdminContentFieldID),
		AdminContentDataID:    int32(a.AdminContentDataID),
		AdminFieldID:          int32(a.AdminFieldID),
		AdminFieldValue:       a.AdminFieldValue,
		History:               a.History,
		DateCreated:           sTime(a.DateCreated.String),
		DateModified:          sTime(a.DateModified.String),
		AdminContentFieldID_2: int32(a.AdminContentFieldID_2),
	}
}

func (d MysqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbm.UpdateAdminDatatypeParams {
	return mdbm.UpdateAdminDatatypeParams{
		ParentID:        Ni32(a.ParentID.Int64),
		Label:           a.Label,
		Type:            a.Type,
		Author:          a.Author,
		AuthorID:        int32(a.AuthorID),
		DateCreated:     sTime(a.DateCreated.String),
		DateModified:    sTime(a.DateModified.String),
		History:         a.History,
		AdminDatatypeID: int32(a.AdminDatatypeID),
	}
}

func (d MysqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbm.UpdateAdminFieldParams {
	return mdbm.UpdateAdminFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		History:      a.History,
		AdminFieldID: int32(a.AdminFieldID),
	}
}

func (d MysqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbm.UpdateAdminRouteParams {
	return mdbm.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		History:      a.History,
		Slug_2:       a.Slug_2,
	}
}

func (d MysqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbm.UpdateContentDataParams {
	return mdbm.UpdateContentDataParams{
		RouteID:       Ni32(a.RouteID),
		DatatypeID:    Ni32(a.DatatypeID),
		History:       a.History,
		DateCreated:   sTime(a.DateCreated.String),
		DateModified:  sTime(a.DateModified.String),
		ContentDataID: int32(a.ContentDataID),
	}
}

func (d MysqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbm.UpdateContentFieldParams {
	return mdbm.UpdateContentFieldParams{
		RouteID:          Ni32(a.RouteID),
		ContentFieldID:   int32(a.ContentFieldID),
		ContentDataID:    int32(a.ContentDataID),
		FieldID:          int32(a.FieldID),
		FieldValue:       a.FieldValue,
		History:          a.History,
		DateCreated:      sTime(a.DateCreated.String),
		DateModified:     sTime(a.DateModified.String),
		ContentFieldID_2: int32(a.ContentFieldID_2),
	}
}

func (d MysqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbm.UpdateDatatypeParams {
	return mdbm.UpdateDatatypeParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		History:      a.History,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		DatatypeID:   int32(a.DatatypeID),
	}
}

func (d MysqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbm.UpdateFieldParams {
	return mdbm.UpdateFieldParams{
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		History:      a.History,
		FieldID:      int32(a.FieldID),
	}
}

func (d MysqlDatabase) MapUpdateMediaParams(a UpdateMediaParams) mdbm.UpdateMediaParams {
	return mdbm.UpdateMediaParams{
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
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		MediaID:      int32(a.MediaID),
	}
}

func (d MysqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbm.UpdateMediaDimensionParams {
	return mdbm.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
		MdID:        int32(a.MdID),
	}
}

func (d MysqlDatabase) MapUpdatePermissionParams(a UpdatePermissionParams) mdbm.UpdatePermissionParams {
	return mdbm.UpdatePermissionParams{
		TableID: int32(a.TableID),
		Label:   a.Label,
		Mode:    int32(a.Mode),
	}
}

func (d MysqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbm.UpdateRoleParams {
	return mdbm.UpdateRoleParams{
		Label:       a.Label,
		Permissions: json.RawMessage(a.Permissions),
		RoleID:      int32(a.RoleID),
	}
}

func (d MysqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbm.UpdateRouteParams {
	return mdbm.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		History:      a.History,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  NSt(a.DateCreated),
		DateModified: NSt(a.DateModified),
		Slug_2:       a.Slug_2,
	}
}
func (d MysqlDatabase) MapUpdateSessionParams(a UpdateSessionParams) mdbm.UpdateSessionParams {
	return mdbm.UpdateSessionParams{
		UserID:      int32(a.UserID),
		CreatedAt:   sTime(a.CreatedAt.String),
		ExpiresAt:   sTime(a.ExpiresAt.String),
		LastAccess:  sTime(a.LastAccess.String),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
		SessionID:   int32(Nsi(a.SessionID).Int64),
	}
}

func (d MysqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbm.UpdateTableParams {
	return mdbm.UpdateTableParams{
		Label: a.Label,
		ID:    int32(a.ID),
	}
}

func (d MysqlDatabase) MapUpdateTokenParams(a UpdateTokenParams) mdbm.UpdateTokenParams {
	return mdbm.UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  sTime(a.IssuedAt).Time,
		ExpiresAt: sTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
		ID:        int32(a.ID),
	}
}

func (d MysqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbm.UpdateUserParams {
	return mdbm.UpdateUserParams{
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Ni32(a.Role),
		UserID:       int32(a.UserID),
	}
}
func (d MysqlDatabase) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdbm.UpdateUserOauthParams {
	return mdbm.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: sTime(a.TokenExpiresAt.String),
		UserOauthID:    int32(a.UserOauthID),
	}
}
