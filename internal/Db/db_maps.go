package db

import mdb "github.com/hegner123/modulacms/db-sqlite"

// Sqlite
func (d Database) MapAdminDatatype(a mdb.AdminDatatypes) AdminDatatypes {
	return AdminDatatypes{
		AdminDtID:    a.AdminDtID,
		AdminRouteID: a.AdminRouteID,
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

func (d Database) MapAdminField(a mdb.AdminFields) AdminFields {
	return AdminFields{
		AdminFieldID: a.AdminFieldID,
		AdminRouteID: a.AdminRouteID,
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

func (d Database) MapAdminRoute(a mdb.AdminRoutes) AdminRoutes {
	return AdminRoutes{
		AdminRouteID: a.AdminRouteID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}

func (d Database) MapContentData(a mdb.ContentData) ContentData {
	return ContentData{
		ContentDataID: a.ContentDataID,
		AdminDtID:     a.AdminDtID,
		History:       a.History,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
	}
}

func (d Database) MapContentField(a mdb.ContentFields) ContentFields {
	return ContentFields{
		ContentFieldID: a.ContentFieldID,
		ContentDataID:  a.ContentDataID,
		AdminFieldID:   a.AdminFieldID,
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}

func (d Database) MapDatatype(a mdb.Datatypes) Datatypes {
	return Datatypes{
		DatatypeID:   a.DatatypeID,
		RouteID:      a.RouteID,
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
		RouteID:      a.RouteID,
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
		MediaID:            a.MediaID,
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:             a.Author,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		Url:                a.Url,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedTablet,
		OptimizedUltraWide: a.OptimizedUltraWide,
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

func (d Database) MapRoles(a mdb.Roles) Roles {
	return Roles{
		RoleID:      a.RoleID,
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func (d Database) MapRoute(a mdb.Routes) Routes {
	routeId, ok := a.RouteID.(int64)
	if !ok {
		return Routes{}

	}
	return Routes{
		RouteID:  routeId,
		Slug:     a.Slug,
		Title:    a.Title,
		Status:   a.Status,
		Author:   a.Author,
		AuthorID: a.AuthorID,
		History:  a.History,
	}
}

func (d Database) MapTables(a mdb.Tables) Tables {
	return Tables{
		ID:       a.ID,
		Label:    a.Label,
		AuthorID: a.AuthorID,
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

func (d Database) MapListDatatypeByRouteIdRow(a mdb.ListDatatypeByRouteIdRow) ListDatatypeByRouteIdRow {
	return ListDatatypeByRouteIdRow{
		DatatypeID: a.DatatypeID,
		RouteID:    a.RouteID,
		ParentID:   a.ParentID,
		Label:      a.Label,
		Type:       a.Type,
	}
}

func (d Database) MapListFieldByRouteIdRow(a mdb.ListFieldByRouteIdRow) ListFieldByRouteIdRow {
	return ListFieldByRouteIdRow{
		FieldID:  a.FieldID,
		RouteID:  a.RouteID,
		ParentID: a.ParentID,
		Label:    a.Label,
		Data:     a.Data,
		Type:     a.Type,
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
		AdminRouteID: a.AdminRouteID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		History:      a.History,
	}
}

func (d Database) MapListAdminDatatypeByRouteIdRow(a mdb.ListAdminDatatypeByRouteIdRow) ListAdminDatatypeByRouteIdRow {
	return ListAdminDatatypeByRouteIdRow{
		AdminDtID:    a.AdminDtID,
		AdminRouteID: a.AdminRouteID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		History:      a.History,
	}
}
func (d Database) MapCreateAdminDatatypeParams(a CreateAdminDatatypeParams) mdb.CreateAdminDatatypeParams {
	return mdb.CreateAdminDatatypeParams{
		AdminRouteID: a.AdminRouteID,
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

func (d Database) MapCreateAdminFieldParams(a CreateAdminFieldParams) mdb.CreateAdminFieldParams {
	return mdb.CreateAdminFieldParams{
		AdminRouteID: a.AdminRouteID,
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
func (d Database) MapCreateAdminRouteParams(a CreateAdminRouteParams) mdb.CreateAdminRouteParams {
	return mdb.CreateAdminRouteParams{
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
	}
}
func (d Database) MapCreateContentDataParams(a CreateContentDataParams) mdb.CreateContentDataParams {
	return mdb.CreateContentDataParams{
		AdminDtID:    a.AdminDtID,
		History:      a.History,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}
func (d Database) MapCreateContentFieldParams(a CreateContentFieldParams) mdb.CreateContentFieldParams {
	return mdb.CreateContentFieldParams{
		ContentFieldID: a.ContentFieldID,
		ContentDataID:  a.ContentDataID,
		AdminFieldID:   a.AdminFieldID,
		FieldValue:     a.FieldValue,
		History:        a.History,
		DateCreated:    a.DateCreated,
		DateModified:   a.DateModified,
	}
}
func (d Database) MapCreateDatatypeParams(a CreateDatatypeParams) mdb.CreateDatatypeParams {
	return mdb.CreateDatatypeParams{
		RouteID:      a.RouteID,
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
		RouteID:      a.RouteID,
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
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:             a.Author,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		Url:                a.Url,
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedDesktop,
		OptimizedUltraWide: a.OptimizedUltraWide,
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
func (d Database) MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeParams) mdb.UpdateAdminDatatypeParams {
	return mdb.UpdateAdminDatatypeParams{
		AdminRouteID: a.AdminRouteID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
		AdminDtID:    a.AdminDtID,
	}
}
func (d Database) MapUpdateAdminFieldParams(a UpdateAdminFieldParams) mdb.UpdateAdminFieldParams {
	return mdb.UpdateAdminFieldParams{
		AdminRouteID: a.AdminRouteID,
		ParentID:     a.ParentID,
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
		AdminFieldID: a.AdminFieldID,
	}
}
func (d Database) MapUpdateAdminRouteParams(a UpdateAdminRouteParams) mdb.UpdateAdminRouteParams {
	return mdb.UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       a.Status,
		Author:       a.Author,
		AuthorID:     a.AuthorID,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
		History:      a.History,
		Slug_2:       a.Slug_2,
	}
}
func (d Database) MapUpdateContentDataParams(a UpdateContentDataParams) mdb.UpdateContentDataParams {
	return mdb.UpdateContentDataParams{
		AdminDtID:     a.AdminDtID,
		History:       a.History,
		DateCreated:   a.DateCreated,
		DateModified:  a.DateModified,
		ContentDataID: a.ContentDataID,
	}
}
func (d Database) MapUpdateContentFieldParams(a UpdateContentFieldParams) mdb.UpdateContentFieldParams {
	return mdb.UpdateContentFieldParams{
		ContentFieldID:   a.ContentFieldID,
		ContentDataID:    a.ContentDataID,
		AdminFieldID:     a.AdminFieldID,
		FieldValue:       a.FieldValue,
		History:          a.History,
		DateCreated:      a.DateCreated,
		DateModified:     a.DateModified,
		ContentFieldID_2: a.ContentFieldID_2,
	}
}
func (d Database) MapUpdateDatatypeParams(a UpdateDatatypeParams) mdb.UpdateDatatypeParams {
	return mdb.UpdateDatatypeParams{
		RouteID:      a.RouteID,
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
		RouteID:      a.RouteID,
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
		Name:               a.Name,
		DisplayName:        a.DisplayName,
		Alt:                a.Alt,
		Caption:            a.Caption,
		Description:        a.Description,
		Class:              a.Class,
		Author:             a.Author,
		AuthorID:           a.AuthorID,
		DateCreated:        a.DateCreated,
		DateModified:       a.DateModified,
		Url:                a.Url,
		Mimetype:           a.Mimetype,
		Dimensions:         a.Dimensions,
		OptimizedMobile:    a.OptimizedMobile,
		OptimizedTablet:    a.OptimizedTablet,
		OptimizedDesktop:   a.OptimizedDesktop,
		OptimizedUltraWide: a.OptimizedUltraWide,
		MediaID:            a.MediaID,
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
func (d Database) MapUtilityGetAdminDatatypesRow(a mdb.UtilityGetAdminDatatypesRow) UtilityGetAdminDatatypesRow {
	return UtilityGetAdminDatatypesRow{
		AdminDtID: a.AdminDtID,
		Label:     a.Label,
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
