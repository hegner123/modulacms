package db

import (
	"encoding/json"

	mdbm "github.com/hegner123/modulacms/db-mysql"
)

// Sqlite
func (d MysqlDatabase) MapAdminDatatype(a mdbm.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDtID:    int64(a.AdminDtID),
		AdminRouteID: Ni64(int64(a.AdminRouteID.Int32)),
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

func (d MysqlDatabase) MapAdminField(a mdbm.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: int64(a.AdminFieldID),
		AdminRouteID: Ni64(int64(a.AdminRouteID)),
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
		ContentDataID: int64(a.ContentDataID),
		AdminDtID:     int64(a.AdminDtID.Int32),
		History:       a.History,
		DateCreated:   Ns(nt(a.DateCreated)),
		DateModified:  ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapContentField(a mdbm.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: int64(a.ContentFieldID),
		ContentDataID:  int64(a.ContentDataID),
		AdminFieldID:   int64(a.AdminFieldID),
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    Ns(nt(a.DateCreated)),
		DateModified:   ns(nt(a.DateModified)),
	}
}

func (d MysqlDatabase) MapDatatype(a mdbm.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   int64(a.DatatypeID),
		RouteID:      Ni64(int64(a.RouteID.Int32)),
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
		RouteID:      Ni64(int64(a.RouteID.Int32)),
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
		DateModified:       ns(nt(a.DateModified)),
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		Url:                a.Url,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedTablet,
		OptimizedUltraWide: a.OptimizedUltraWide,
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
func (d MysqlDatabase) MapListDatatypeByRouteIdRow(a mdbm.ListDatatypeByRouteIdRow) ListDatatypeByRouteIdRow {
	return ListDatatypeByRouteIdRow{
		DatatypeID: int64(a.DatatypeID),
		RouteID:    Ni64(int64(a.RouteID.Int32)),
		ParentID:   Ni64(int64(a.ParentID.Int32)),
		Label:      a.Label,
		Type:       a.Type,
	}
}
func (d MysqlDatabase) MapListFieldByRouteIdRow(a mdbm.ListFieldByRouteIdRow) ListFieldByRouteIdRow {
	return ListFieldByRouteIdRow{
		FieldID:  int64(a.FieldID),
		RouteID:  ni64(int64(a.RouteID.Int32)),
		ParentID: Ni64(int64(a.ParentID.Int32)),
		Label:    a.Label,
		Data:     a.Data,
		Type:     a.Type,
	}
}

func (d MysqlDatabase) MapListAdminFieldsByDatatypeIDRow(a mdbm.ListAdminFieldsByDatatypeIDRow) ListAdminFieldsByDatatypeIDRow {
	return ListAdminFieldsByDatatypeIDRow{
		AdminFieldID: int64(a.AdminFieldID),
		AdminRouteID: Ni64(int64(a.AdminRouteID)),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		History:      a.History,
	}
}
func (d MysqlDatabase) MapListAdminDatatypeByRouteIdRow(a mdbm.ListAdminDatatypeByRouteIdRow) ListAdminDatatypeByRouteIdRow {
	return ListAdminDatatypeByRouteIdRow{
		AdminDtID:    int64(a.AdminDtID),
		AdminRouteID: Ni64(int64(a.AdminRouteID.Int32)),
		ParentID:     Ni64(int64(a.ParentID.Int32)),
		Label:        a.Label,
		Type:         a.Type,
		History:      a.History,
	}
}

func (d MysqlDatabase) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdbm.CreateAdminDatatypeParams {
	return mdbm.CreateAdminDatatypeParams{
		AdminRouteID: Ni32(a.AdminRouteID.Int64),
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
		AdminRouteID: int32(a.AdminRouteID.Int64),
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
		AdminDtID:    Ni32(a.AdminDtID),
		History:      a.History,
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
	}
}
func (d MysqlDatabase) MapCreateContentFieldParams(a CreateContentFieldParams) mdbm.CreateContentFieldParams {
	return mdbm.CreateContentFieldParams{
		ContentFieldID: int32(a.ContentFieldID),
		ContentDataID:  int32(a.ContentDataID),
		AdminFieldID:   int32(a.AdminFieldID),
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
	}
}
func (d MysqlDatabase) MapCreateDatatypeParams(a CreateDatatypeParams) mdbm.CreateDatatypeParams {
	return mdbm.CreateDatatypeParams{
		RouteID:      Ni32(a.RouteID.Int64),
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
		RouteID:      Ni32(a.RouteID.Int64),
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
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		Url:                a.Url,
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedDesktop,
		OptimizedUltraWide: a.OptimizedUltraWide,
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
		DateCreated:  a.DateCreated.Time,
		DateModified: a.DateModified.Time,
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
func (d MysqlDatabase) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdbm.UpdateAdminDatatypeParams {
	return mdbm.UpdateAdminDatatypeParams{
		AdminRouteID: Ni32(a.AdminRouteID.Int64),
		ParentID:     Ni32(a.ParentID.Int64),
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
		History:      a.History,
		AdminDtID:    int32(a.AdminDtID),
	}
}
func (d MysqlDatabase) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdbm.UpdateAdminFieldParams {
	return mdbm.UpdateAdminFieldParams{
		AdminRouteID: int32(a.AdminRouteID.Int64),
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
		AdminDtID:     Ni32(a.AdminDtID),
		History:       a.History,
		DateCreated:   sTime(a.DateCreated.String),
		DateModified:  sTime(a.DateModified.String),
		ContentDataID: int32(a.ContentDataID),
	}
}
func (d MysqlDatabase) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdbm.UpdateContentFieldParams {
	return mdbm.UpdateContentFieldParams{
		ContentFieldID:   int32(a.ContentFieldID),
		ContentDataID:    int32(a.ContentDataID),
		AdminFieldID:     int32(a.AdminFieldID),
		FieldValue:       a.FieldValue,
		History:          a.History,
		DateCreated:      sTime(a.DateCreated.String),
		DateModified:     sTime(a.DateModified.String),
		ContentFieldID_2: int32(a.ContentFieldID_2),
	}
}
func (d MysqlDatabase) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdbm.UpdateDatatypeParams {
	return mdbm.UpdateDatatypeParams{
		RouteID:      Ni32(a.RouteID.Int64),
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
		RouteID:      Ni32(a.RouteID.Int64),
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
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:       AssertString(a.Author),
		AuthorID:     int32(a.AuthorID),
		DateCreated:  sTime(a.DateCreated.String),
		DateModified: sTime(a.DateModified.String),
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
func (d MysqlDatabase) MapUpdateMediaDimensionParams(a UpdateMediaDimensionParams) mdbm.UpdateMediaDimensionParams {
	return mdbm.UpdateMediaDimensionParams{
		Label:       a.Label,
		Width:       Ni32(a.Width.Int64),
		Height:      Ni32(a.Height.Int64),
		AspectRatio: a.AspectRatio,
		MdID:        int32(a.MdID),
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
		DateCreated:  a.DateCreated.Time,
		DateModified: a.DateModified.Time,
		Slug_2:       a.Slug_2,
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
