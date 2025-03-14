package db

import (
	"encoding/json"

	mdbp "github.com/hegner123/modulacms/db-psql"
	"github.com/sqlc-dev/pqtype"
)

// Sqlite
func (d PsqlDatabase) MapAdminContentData(a mdbp.AdminContentData) AdminContentData {
	return AdminContentData{
		AdminDatatypeID: int64(a.AdminDatatypeID.Int32),
		AdminRouteID:    int64(a.AdminRouteID.Int32),
		DateCreated:     Ns(nt(a.DateCreated)),
		DateModified:    Ns(nt(a.DateModified)),
		History:         a.History,
	}
}
func (d PsqlDatabase) MapAdminContentField(a mdbp.AdminContentFields) AdminContentFields {
	return AdminContentFields{
		AdminContentFieldID: int64(a.AdminContentFieldID),
		AdminContentDataID:  int64(a.AdminContentDataID),
		AdminFieldID:        int64(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History,
		DateCreated:         Ns(nt(a.DateCreated)),
		DateModified:        Ns(nt(a.DateModified)),
	}
}
func (d PsqlDatabase) MapAdminDatatype(a mdbp.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDatatypeID: int64(a.AdminDatatypeID),
		ParentID:        Ni64(int64(a.ParentID.Int32)),
		Label:           a.Label,
		Type:            a.Type,
		Author:          a.Author,
		AuthorID:        int64(a.AuthorID),
		DateCreated:     Ns(nt(a.DateCreated)),
		DateModified:    Ns(nt(a.DateModified)),
		History:         a.History,
	}
}

func (d PsqlDatabase) MapAdminField(a mdbp.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: int64(a.AdminFieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: Ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d PsqlDatabase) MapAdminRoute(a mdbp.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: int64(a.AdminRouteID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: Ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d PsqlDatabase) MapContentData(a mdbp.ContentData) ContentData {
	return ContentData{
		ContentDataID: int64(a.ContentDataID),
		RouteID:       int64(a.RouteID.Int32),
		DatatypeID:    int64(a.DatatypeID.Int32),
		History:       a.History,
		DateCreated:   Ns(nt(a.DateCreated)),
		DateModified:  Ns(nt(a.DateModified)),
	}
}

func (d PsqlDatabase) MapContentField(a mdbp.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: int64(a.ContentFieldID),
		RouteID:        int64(a.RouteID.Int32),
		ContentDataID:  int64(a.ContentDataID),
		FieldID:        int64(a.FieldID),
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    Ns(nt(a.DateCreated)),
		DateModified:   Ns(nt(a.DateModified)),
	}
}

func (d PsqlDatabase) MapDatatype(a mdbp.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   int64(a.DatatypeID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: Ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d PsqlDatabase) MapField(a mdbp.Fields) Fields {
	return Fields{
		FieldID:      int64(a.FieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: Ns(nt(a.DateModified)),
		History:      a.History,
	}
}

func (d PsqlDatabase) MapMedia(a mdbp.Media) Media {
	return Media{
		MediaID:            int64(a.MediaID),
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:             a.Author,
		AuthorID:           int64(a.AuthorID),
		DateCreated:        Ns(nt(a.DateCreated)),
		DateModified:       Ns(nt(a.DateModified)),
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		Url:                a.Url,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedTablet,
		OptimizedUltraWide: a.OptimizedUltraWide,
	}
}

func (d PsqlDatabase) MapMediaDimension(a mdbp.MediaDimensions) MediaDimensions {
	return MediaDimensions{
		MdID:        int64(a.MdID),
		Label:       a.Label,
		Width:       Ni64(int64(a.Width.Int32)),
		Height:      Ni64(int64(a.Height.Int32)),
		AspectRatio: a.AspectRatio,
	}

}

func (d PsqlDatabase) MapRoles(a mdbp.Roles) Roles {
	return Roles{
		RoleID:      int64(a.RoleID),
		Label:       a.Label,
		Permissions: pString(a.Permissions),
	}
}

func (d PsqlDatabase) MapRoute(a mdbp.Routes) Routes {
	return Routes{
		RouteID:      int64(a.RouteID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int64(a.Status),
		Author:       a.Author,
		AuthorID:     int64(a.AuthorID),
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}
func (d PsqlDatabase) MapSession(a mdbp.Sessions) Sessions {
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

func (d PsqlDatabase) MapTables(a mdbp.Tables) Tables {
	return Tables{
		ID:       int64(a.ID),
		Label:    a.Label,
		AuthorID: int64(a.AuthorID),
	}
}

func (d PsqlDatabase) MapToken(a mdbp.Tokens) Tokens {
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

func (d PsqlDatabase) MapUser(a mdbp.Users) Users {
	return Users{
		UserID:       int64(a.UserID),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int64(a.Role.Int32),
		DateCreated:  Ns(nt(a.DateCreated)),
		DateModified: Ns(nt(a.DateModified)),
	}
}

func (d PsqlDatabase) MapUserOauth(a mdbp.UserOauth) UserOauth {
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

func (d PsqlDatabase) MapListAdminFieldsByDatatypeIDRow(a mdbp.ListAdminFieldsByDatatypeIDRow) ListAdminFieldsByDatatypeIDRow {
	return ListAdminFieldsByDatatypeIDRow{
		AdminFieldID: int64(a.AdminFieldID),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		History:      a.History,
	}
}

func (d PsqlDatabase) MapCreateAdminContentDataParams(a CreateAdminContentDataParams) mdbp.CreateAdminContentDataParams {
	return mdbp.CreateAdminContentDataParams{
		AdminRouteID:    Ni32(a.AdminRouteID),
		AdminDatatypeID: Ni32(a.AdminDatatypeID),
		History:         a.History,
		DateCreated:     sTime(a.DateCreated.String),
		DateModified:    sTime(a.DateModified.String),
	}
}
func (d PsqlDatabase) MapCreateAdminContentFieldParams(a CreateAdminContentFieldParams) mdbp.CreateAdminContentFieldParams {
	return mdbp.CreateAdminContentFieldParams{
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

func (d PsqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbp.CreateAdminDatatypeParams {
	return mdbp.CreateAdminDatatypeParams{
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

func (d PsqlDatabase) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdbp.CreateAdminFieldParams {
	return mdbp.CreateAdminFieldParams{
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
func (d PsqlDatabase) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdbp.CreateAdminRouteParams {
	return mdbp.CreateAdminRouteParams{
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
func (d PsqlDatabase) MapCreateContentDataParams(a CreateContentDataParams) mdbp.CreateContentDataParams {
	return mdbp.CreateContentDataParams{
		RouteID:      Ni32(a.RouteID),
		DatatypeID:   Ni32(a.DatatypeID),
		History:      a.History,
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
	}
}
func (d PsqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbp.CreateContentFieldParams {
	return mdbp.CreateContentFieldParams{
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
func (d PsqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbp.CreateDatatypeParams {
	return mdbp.CreateDatatypeParams{
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

func (d PsqlDatabase) MapCreateFieldParams(a CreateFieldParams) mdbp.CreateFieldParams {
	return mdbp.CreateFieldParams{
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
func (d PsqlDatabase) MapCreateMediaParams(a CreateMediaParams) mdbp.CreateMediaParams {
	return mdbp.CreateMediaParams{
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:             AssertString(a.Author),
		AuthorID:           int32(a.AuthorID),
		DateCreated:        sTime(a.DateCreated.String),
		DateModified:       sTime(a.DateModified.String),
		Url:                a.Url,
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedDesktop,
		OptimizedUltraWide: a.OptimizedUltraWide,
	}
}
func (d PsqlDatabase) MapCreateMediaDimensionParams(a CreateMediaDimensionParams) mdbp.CreateMediaDimensionParams {
	return mdbp.CreateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
	}
}
func (d PsqlDatabase) MapCreateRoleParams(a CreateRoleParams) mdbp.CreateRoleParams {
	return mdbp.CreateRoleParams{
		Label:       a.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(a.Permissions)},
	}
}
func (d PsqlDatabase) MapCreateRouteParams(a CreateRouteParams) mdbp.CreateRouteParams {
	return mdbp.CreateRouteParams{
		Author:       a.Author,
		AuthorID:     int32(a.AuthorID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

func (d PsqlDatabase) MapCreateSessionParams(a CreateSessionParams) mdbp.CreateSessionParams {
	return mdbp.CreateSessionParams{
		UserID:      int32(a.UserID),
		CreatedAt:   sTime(a.CreatedAt.String),
		ExpiresAt:   sTime(a.ExpiresAt.String),
		LastAccess:  sTime(a.ExpiresAt.String),
		IpAddress:   a.IpAddress,
		UserAgent:   a.UserAgent,
		SessionData: a.SessionData,
	}
}

func (d PsqlDatabase) MapCreateTokenParams(a CreateTokenParams) mdbp.CreateTokenParams {
	return mdbp.CreateTokenParams{
		UserID:    int32(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  sTime(a.IssuedAt).Time,
		ExpiresAt: sTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
	}
}
func (d PsqlDatabase) MapCreateUserParams(a CreateUserParams) mdbp.CreateUserParams {
	return mdbp.CreateUserParams{
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Ni32(a.Role),
	}
}

func (d PsqlDatabase) MapCreateUserOauthParams(a CreateUserOauthParams) mdbp.CreateUserOauthParams {
	return mdbp.CreateUserOauthParams{
		UserID:              int32(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken,
		RefreshToken:        a.RefreshToken,
		TokenExpiresAt:      sTime(a.TokenExpiresAt.String),
		DateCreated:         sTime(a.DateCreated.String),
	}
}

func (d PsqlDatabase) MapUpdateAdminContentDataParams(a UpdateAdminContentDataParams) mdbp.UpdateAdminContentDataParams {
	return mdbp.UpdateAdminContentDataParams{
		AdminRouteID:       Ni32(a.AdminRouteID),
		AdminDatatypeID:    Ni32(a.AdminDatatypeID),
		History:            a.History,
		DateCreated:        sTime(a.DateCreated.String),
		DateModified:       sTime(a.DateModified.String),
		AdminContentDataID: int32(a.AdminContentDataID),
	}
}
func (d PsqlDatabase) MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldParams) mdbp.UpdateAdminContentFieldParams {
	return mdbp.UpdateAdminContentFieldParams{
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
func (d PsqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbp.UpdateAdminDatatypeParams {
	return mdbp.UpdateAdminDatatypeParams{
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
func (d PsqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbp.UpdateAdminFieldParams {
	return mdbp.UpdateAdminFieldParams{
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
func (d PsqlDatabase) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdbp.UpdateAdminRouteParams {
	return mdbp.UpdateAdminRouteParams{
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
func (d PsqlDatabase) MapUpdateContentDataParams(a UpdateContentDataParams) mdbp.UpdateContentDataParams {
	return mdbp.UpdateContentDataParams{
		RouteID:       Ni32(a.RouteID),
		DatatypeID:    Ni32(a.DatatypeID),
		History:       a.History,
		DateCreated:   sTime(a.DateCreated.String),
		DateModified:  sTime(a.DateModified.String),
		ContentDataID: int32(a.ContentDataID),
	}
}
func (d PsqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbp.UpdateContentFieldParams {
	return mdbp.UpdateContentFieldParams{
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
func (d PsqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbp.UpdateDatatypeParams {
	return mdbp.UpdateDatatypeParams{
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
func (d PsqlDatabase) MapUpdateFieldParams(a UpdateFieldParams) mdbp.UpdateFieldParams {
	return mdbp.UpdateFieldParams{
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
func (d PsqlDatabase) MapUpdateMediaParams(a UpdateMediaParams) mdbp.UpdateMediaParams {
	return mdbp.UpdateMediaParams{
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:             AssertString(a.Author),
		AuthorID:           int32(a.AuthorID),
		DateCreated:        sTime(a.DateCreated.String),
		DateModified:       sTime(a.DateModified.String),
		Url:                a.Url,
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedDesktop,
		OptimizedUltraWide: a.OptimizedUltraWide,
		MediaID:            int32(a.MediaID),
	}
}
func (d PsqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbp.UpdateMediaDimensionParams {
	return mdbp.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
		MdID:        int32(a.MdID),
	}
}
func (d PsqlDatabase) MapUpdateRoleParams(a UpdateRoleParams) mdbp.UpdateRoleParams {
	return mdbp.UpdateRoleParams{
		Label:       a.Label,
		Permissions: pqtype.NullRawMessage{RawMessage: json.RawMessage(a.Permissions)},
		RoleID:      int32(a.RoleID),
	}
}
func (d PsqlDatabase) MapUpdateRouteParams(a UpdateRouteParams) mdbp.UpdateRouteParams {
	return mdbp.UpdateRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       int32(a.Status),
		History:      a.History,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		Slug_2:       a.Slug_2,
	}
}
func (d PsqlDatabase) MapUpdateSessionParams(a UpdateSessionParams) mdbp.UpdateSessionParams {
	return mdbp.UpdateSessionParams{
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
func (d PsqlDatabase) MapUpdateTableParams(a UpdateTableParams) mdbp.UpdateTableParams {
	return mdbp.UpdateTableParams{
		Label: a.Label,
		ID:    int32(a.ID),
	}
}
func (d PsqlDatabase) MapUpdateTokenParams(a UpdateTokenParams) mdbp.UpdateTokenParams {
	return mdbp.UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  sTime(a.IssuedAt).Time,
		ExpiresAt: sTime(a.ExpiresAt).Time,
		Revoked:   a.Revoked,
		ID:        int32(a.ID),
	}
}
func (d PsqlDatabase) MapUpdateUserParams(a UpdateUserParams) mdbp.UpdateUserParams {
	return mdbp.UpdateUserParams{
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
func (d PsqlDatabase) MapUpdateUserOauthParams(a UpdateUserOauthParams) mdbp.UpdateUserOauthParams {
	return mdbp.UpdateUserOauthParams{
		AccessToken:    a.AccessToken,
		RefreshToken:   a.RefreshToken,
		TokenExpiresAt: sTime(a.TokenExpiresAt.String),
		UserOauthID:    int32(a.UserOauthID),
	}
}
