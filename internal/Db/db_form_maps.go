package db

import ()

func MapCreateAdminContentDataParams(a CreateAdminContentDataFormParams) CreateAdminContentDataParams {
	return CreateAdminContentDataParams{
		AdminRouteID:    Si(a.AdminRouteID),
		AdminDatatypeID: Si(a.AdminDatatypeID),
		History:         Ns(a.History),
		DateCreated:     Ns(a.DateCreated),
		DateModified:    Ns(a.DateModified),
	}
}
func MapCreateAdminContentFieldParams(a CreateAdminContentFieldFormParams) CreateAdminContentFieldParams {
	return CreateAdminContentFieldParams{
		AdminRouteID:        Si(a.AdminRouteID),
		AdminContentFieldID: Si(a.AdminContentFieldID),
		AdminContentDataID:  Si(a.AdminContentDataID),
		AdminFieldID:        Si(a.AdminFieldID),
		AdminFieldValue:     a.AdminFieldValue,
		History:             Ns(a.History),
		DateCreated:         Ns(a.DateCreated),
		DateModified:        Ns(a.DateModified),
	}
}

func MapCreateAdminDatatypeParams(a CreateAdminDatatypeFormParams) CreateAdminDatatypeParams {
	return CreateAdminDatatypeParams{
		ParentID:     Nsi(a.ParentID),
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     Si(a.ParentID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
	}
}

func MapCreateAdminFieldParams(a CreateAdminFieldFormParams) CreateAdminFieldParams {

	return CreateAdminFieldParams{
		ParentID:     Nsi(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
	}
}
func MapCreateAdminRouteParams(a CreateAdminRouteFormParams) CreateAdminRouteParams {
	return CreateAdminRouteParams{
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       Si(a.Status),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
	}
}
func MapCreateContentDataParams(a CreateContentDataFormParams) CreateContentDataParams {
	return CreateContentDataParams{
		RouteID:      Si(a.RouteID),
		DatatypeID:   Si(a.DatatypeID),
		History:      Ns(a.History),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
	}
}
func MapCreateContentFieldParams(a CreateContentFieldFormParams) CreateContentFieldParams {
	return CreateContentFieldParams{
		RouteID:        Si(a.RouteID),
		ContentFieldID: Si(a.ContentFieldID),
		ContentDataID:  Si(a.ContentDataID),
		FieldID:        Si(a.FieldID),
		FieldValue:     a.FieldValue,
		History:        Ns(a.History),
		DateCreated:    Ns(a.DateCreated),
		DateModified:   Ns(a.DateModified),
	}
}
func MapCreateDatatypeParams(a CreateDatatypeFormParams) CreateDatatypeParams {
	return CreateDatatypeParams{
		ParentID:     Nsi(a.ParentID),
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		History:      Ns(a.History),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
	}
}

func MapCreateFieldParams(a CreateFieldFormParams) CreateFieldParams {
	return CreateFieldParams{
		ParentID:     Nsi(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		History:      Ns(a.History),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
	}
}
func MapCreateMediaParams(a CreateMediaFormParams) CreateMediaParams {
	return CreateMediaParams{
		Name:         Ns(a.Name),
		DisplayName:  Ns(a.DisplayName),
		Alt:          Ns(a.Alt),
		Caption:      Ns(a.Caption),
		Description:  Ns(a.Description),
		Class:        Ns(a.Class),
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		Url:          Ns(a.Url),
		Mimetype:     Ns(a.Mimetype),
		Dimensions:   Ns(a.Dimensions),
		Srcset:       Ns(a.Srcset),
	}
}
func MapCreateMediaDimensionParams(a CreateMediaDimensionFormParams) CreateMediaDimensionParams {
	return CreateMediaDimensionParams{
		Label:       Ns(a.Label),
		Width:       Nsi(a.Width),
		Height:      Nsi(a.Height),
		AspectRatio: Ns(a.AspectRatio),
	}
}
func MapCreatePermissionParams(a CreatePermissionFormParams) CreatePermissionParams {
	return CreatePermissionParams{
		TableID: Si(a.TableID),
		Label:   a.Label,
		Mode:    Si(a.Mode),
	}
}
func MapCreateRoleParams(a CreateRoleFormParams) CreateRoleParams {
	return CreateRoleParams(a)
}
func MapCreateRouteParams(a CreateRouteFormParams) CreateRouteParams {
	return CreateRouteParams{
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       Si(a.Status),
		History:      Ns(a.History),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
	}
}

func MapCreateSessionParams(a CreateSessionFormParams) CreateSessionParams {
	return CreateSessionParams{
		UserID:      Si(a.UserID),
		CreatedAt:   Ns(a.CreatedAt),
		ExpiresAt:   Ns(a.ExpiresAt),
		LastAccess:  Ns(a.LastAccess),
		IpAddress:   Ns(a.IpAddress),
		UserAgent:   Ns(a.UserAgent),
		SessionData: Ns(a.SessionData),
	}
}

func MapCreateTokenParams(a CreateTokenFormParams) CreateTokenParams {
	return CreateTokenParams{
		UserID:    Si(a.UserID),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   Nb(sb(a.Revoked)),
	}
}
func MapCreateUserParams(a CreateUserFormParams) CreateUserParams {
	return CreateUserParams{
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Si(a.Role),
	}
}
func MapCreateUserOauthParams(a CreateUserOauthFormParams) CreateUserOauthParams {
	return CreateUserOauthParams{
		UserID:              Si(a.UserID),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         Ns(a.AccessToken),
		RefreshToken:        Ns(a.RefreshToken),
		TokenExpiresAt:      Ns(a.TokenExpiresAt),
		DateCreated:         Ns(a.DateCreated),
	}
}
func MapUpdateAdminContentDataParams(a UpdateAdminContentDataFormParams) UpdateAdminContentDataParams {
	return UpdateAdminContentDataParams{
		AdminRouteID:       Si(a.AdminRouteID),
		AdminDatatypeID:    Si(a.AdminDatatypeID),
		History:            Ns(a.History),
		DateCreated:        Ns(a.DateCreated),
		DateModified:       Ns(a.DateModified),
		AdminContentDataID: Si(a.AdminContentDataID),
	}
}
func MapUpdateAdminContentFieldParams(a UpdateAdminContentFieldFormParams) UpdateAdminContentFieldParams {
	return UpdateAdminContentFieldParams{
		AdminRouteID:          Si(a.AdminRouteID),
		AdminContentFieldID:   Si(a.AdminContentFieldID),
		AdminContentDataID:    Si(a.AdminContentDataID),
		AdminFieldID:          Si(a.AdminFieldID),
		AdminFieldValue:       a.AdminFieldValue,
		History:               Ns(a.History),
		DateCreated:           Ns(a.DateCreated),
		DateModified:          Ns(a.DateModified),
		AdminContentFieldID_2: Si(a.AdminContentFieldID_2),
	}
}
func MapUpdateAdminDatatypeParams(a UpdateAdminDatatypeFormParams) UpdateAdminDatatypeParams {
	return UpdateAdminDatatypeParams{
		ParentID:        Nsi(a.ParentID),
		Label:           a.Label,
		Type:            a.Type,
		Author:          a.Author,
		AuthorID:        Si(a.AuthorID),
		DateCreated:     Ns(a.DateCreated),
		DateModified:    Ns(a.DateModified),
		History:         Ns(a.History),
		AdminDatatypeID: Si(a.AdminDatatypeID),
	}
}
func MapUpdateAdminFieldParams(a UpdateAdminFieldFormParams) UpdateAdminFieldParams {
	return UpdateAdminFieldParams{
		ParentID:     Nsi(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
		AdminFieldID: Si(a.AdminFieldID),
	}
}
func MapUpdateAdminRouteParams(a UpdateAdminRouteFormParams) UpdateAdminRouteParams {
	return UpdateAdminRouteParams{
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       Si(a.Status),
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
		Slug_2:       a.Slug_2,
	}
}
func MapUpdateContentDataParams(a UpdateContentDataFormParams) UpdateContentDataParams {
	return UpdateContentDataParams{
		RouteID:       Si(a.RouteID),
		DatatypeID:    Si(a.DatatypeID),
		History:       Ns(a.History),
		DateCreated:   Ns(a.DateCreated),
		DateModified:  Ns(a.DateModified),
		ContentDataID: Si(a.ContentDataID),
	}
}
func MapUpdateContentFieldParams(a UpdateContentFieldFormParams) UpdateContentFieldParams {
	return UpdateContentFieldParams{
		RouteID:          Si(a.RouteID),
		ContentFieldID:   Si(a.ContentFieldID),
		ContentDataID:    Si(a.ContentDataID),
		FieldID:          Si(a.FieldID),
		FieldValue:       a.FieldValue,
		History:          Ns(a.History),
		DateCreated:      Ns(a.DateCreated),
		DateModified:     Ns(a.DateModified),
		ContentFieldID_2: Si(a.ContentFieldID_2),
	}
}
func MapUpdateDatatypeParams(a UpdateDatatypeFormParams) UpdateDatatypeParams {
	return UpdateDatatypeParams{
		ParentID:     Nsi(a.ParentID),
		Label:        a.Label,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		History:      Ns(a.History),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		DatatypeID:   Si(a.DatatypeID),
	}
}
func MapUpdateFieldParams(a UpdateFieldFormParams) UpdateFieldParams {
	return UpdateFieldParams{
		ParentID:     Nsi(a.ParentID),
		Label:        a.Label,
		Data:         a.Data,
		Type:         a.Type,
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		History:      Ns(a.History),
		FieldID:      Si(a.FieldID),
	}
}
func MapUpdateMediaParams(a UpdateMediaFormParams) UpdateMediaParams {
	return UpdateMediaParams{
		Name:         Ns(a.Name),
		DisplayName:  Ns(a.DisplayName),
		Alt:          Ns(a.Alt),
		Caption:      Ns(a.Caption),
		Description:  Ns(a.Description),
		Class:        Ns(a.Class),
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		Url:          Ns(a.Url),
		Mimetype:     Ns(a.Mimetype),
		Dimensions:   Ns(a.Dimensions),
		Srcset:       Ns(a.Srcset),
		MediaID:      Si(a.MediaID),
	}
}
func MapUpdateMediaDimensionParams(a UpdateMediaDimensionFormParams) UpdateMediaDimensionParams {
	return UpdateMediaDimensionParams{
		Label:       Ns(a.Label),
		Width:       Nsi(a.Width),
		Height:      Nsi(a.Height),
		AspectRatio: Ns(a.AspectRatio),
		MdID:        Si(a.MdID),
	}
}
func MapUpdatePermissionParams(a UpdatePermissionFormParams) UpdatePermissionParams {
	return UpdatePermissionParams{
		TableID: Si(a.TableID),
		Label:   a.Label,
		Mode:    Si(a.Mode),
	}
}
func MapUpdateRoleParams(a UpdateRoleFormParams) UpdateRoleParams {
	return UpdateRoleParams{
		Label:       a.Label,
		Permissions: a.Permissions,
		RoleID:      Si(a.RoleID),
	}
}
func MapUpdateRouteParams(a UpdateRouteFormParams) UpdateRouteParams {
	return UpdateRouteParams{
		Author:       a.Author,
		AuthorID:     Si(a.AuthorID),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       Si(a.Status),
		History:      Ns(a.History),
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		Slug_2:       a.Slug_2,
	}
}
func MapUpdateSessionParams(a UpdateSessionFormParams) UpdateSessionParams {
	return UpdateSessionParams{
		UserID:      Si(a.UserID),
		CreatedAt:   Ns(a.CreatedAt),
		ExpiresAt:   Ns(a.ExpiresAt),
		LastAccess:  Ns(a.LastAccess),
		IpAddress:   Ns(a.IpAddress),
		UserAgent:   Ns(a.UserAgent),
		SessionData: Ns(a.SessionData),
		SessionID:   a.Session_ID,
	}
}
func MapUpdateTableParams(a UpdateTableFormParams) UpdateTableParams {
	return UpdateTableParams{
		Label: Ns(a.Label),
		ID:    Si(a.ID),
	}
}
func MapUpdateTokenParams(a UpdateTokenFormParams) UpdateTokenParams {
	return UpdateTokenParams{
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   Nb(sb(a.Revoked)),
		ID:        Si(a.ID),
	}
}
func MapUpdateUserParams(a UpdateUserFormParams) UpdateUserParams {
	return UpdateUserParams{
		DateCreated:  Ns(a.DateCreated),
		DateModified: Ns(a.DateModified),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         Si(a.Role),
		UserID:       Si(a.UserID),
	}
}
func MapUpdateUserOauthParams(a UpdateUserOauthFormParams) UpdateUserOauthParams {
	return UpdateUserOauthParams{
		AccessToken:    Ns(a.AccessToken),
		RefreshToken:   Ns(a.RefreshToken),
		TokenExpiresAt: Ns(a.TokenExpiresAt),
		UserOauthID:    Si(a.UserOauthID),
	}
}
