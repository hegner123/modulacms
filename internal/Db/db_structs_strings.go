package db

import "strconv"

func MapStringAdminContentData(a AdminContentData) StringAdminContentData {
	return StringAdminContentData{
		AdminContentDataID: strconv.FormatInt(a.AdminContentDataID, 10),
		AdminRouteID:       strconv.FormatInt(a.AdminRouteID, 10),
		AdminDatatypeID:    strconv.FormatInt(a.AdminDatatypeID, 10),
		History:            a.History.String,
		DateCreated:        a.DateCreated.String,
		DateModified:       a.DateModified.String,
	}
}

func MapStringAdminContentField(a AdminContentFields) StringAdminContentFields {
	return StringAdminContentFields{
		AdminContentFieldID: strconv.FormatInt(a.AdminContentFieldID, 10),
		AdminContentDataID:  strconv.FormatInt(a.AdminContentDataID, 10),
		AdminFieldID:        strconv.FormatInt(a.AdminFieldID, 10),
		AdminFieldValue:     a.AdminFieldValue,
		History:             a.History.String,
		DateCreated:         a.DateCreated.String,
		DateModified:        a.DateModified.String,
	}
}

func MapStringAdminDatatype(a AdminDatatypes) StringAdminDatatypes {
	return StringAdminDatatypes{
		AdminDatatypeID: strconv.FormatInt(a.AdminDatatypeID, 10),
		ParentID:        strconv.FormatInt(a.ParentID.Int64, 10),
		Label:           a.Label,
		Type:            a.Type,
		Author:          a.Author,
		AuthorID:        strconv.FormatInt(a.AuthorID, 10),
		DateCreated:     a.DateCreated.String,
		DateModified:    a.DateModified.String,
		History:         a.History.String,
	}
}

func MapStringAdminField(a AdminFields) StringAdminFields {
	return StringAdminFields{
		AdminFieldID: strconv.FormatInt(a.AdminFieldID, 10),
		ParentID:     strconv.FormatInt(a.ParentID.Int64, 10),
		Label:        AssertString(a.Label),
		Data:         AssertString(a.Data),
		Type:         AssertString(a.Type),
		Author:       AssertString(a.Author),
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
		History:      a.History.String,
	}
}

func MapStringAdminRoute(a AdminRoutes) StringAdminRoutes {
	return StringAdminRoutes{
		AdminRouteID: strconv.FormatInt(a.AdminRouteID, 10),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       strconv.FormatInt(a.Status, 10),
		Author:       AssertString(a.Author),
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
		History:      a.History.String,
	}
}

func MapStringContentData(a ContentData) StringContentData {
	return StringContentData{
		ContentDataID: strconv.FormatInt(a.ContentDataID, 10),
		DatatypeID:    strconv.FormatInt(a.DatatypeID, 10),
		History:       a.History.String,
		DateCreated:   a.DateCreated.String,
		DateModified:  a.DateModified.String,
	}
}

func MapStringContentField(a ContentFields) StringContentFields {
	return StringContentFields{
		ContentFieldID: strconv.FormatInt(a.ContentFieldID, 10),
		ContentDataID:  strconv.FormatInt(a.ContentDataID, 10),
		FieldID:        strconv.FormatInt(a.FieldID, 10),
		FieldValue:     a.FieldValue,
		History:        a.History.String,
		DateCreated:    a.DateCreated.String,
		DateModified:   a.DateModified.String,
	}
}

func MapStringDatatype(a Datatypes) StringDatatypes {
	return StringDatatypes{
		DatatypeID:   strconv.FormatInt(a.DatatypeID, 10),
		ParentID:     strconv.FormatInt(a.ParentID.Int64, 10),
		Label:        a.Label,
		Type:         a.Type,
		Author:       AssertString(a.Author),
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
		History:      a.History.String,
	}
}

func MapStringField(a Fields) StringFields {
	return StringFields{
		FieldID:      strconv.FormatInt(a.FieldID, 10),
		ParentID:     strconv.FormatInt(a.ParentID.Int64, 10),
		Label:        AssertString(a.Label),
		Data:         a.Data,
		Type:         a.Type,
		Author:       AssertString(a.Author),
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
		History:      a.History.String,
	}
}

func MapStringMedia(a Media) StringMedia {
	return StringMedia{
		MediaID:      strconv.FormatInt(a.MediaID, 10),
		Name:         a.Name.String,
		DisplayName:  a.DisplayName.String,
		Alt:          a.Alt.String,
		Caption:      a.Caption.String,
		Description:  a.Description.String,
		Class:        a.Class.String,
		Mimetype:     a.Mimetype.String,
		Dimensions:   a.Dimensions.String,
		Url:          a.Url.String,
		Srcset:       a.Srcset.String,
        Author:       AssertString(a.Author),
        AuthorID:     strconv.FormatInt(a.AuthorID, 10),
        DateCreated:  a.DateCreated.String,
        DateModified: a.DateModified.String,
	}
}

func MapStringMediaDimension(a MediaDimensions) StringMediaDimensions {
	return StringMediaDimensions{
		MdID:        strconv.FormatInt(a.MdID, 10),
		Label:       a.Label.String,
		Width:       strconv.FormatInt(a.Width.Int64, 10),
		Height:      strconv.FormatInt(a.Height.Int64, 10),
		AspectRatio: a.AspectRatio.String,
	}

}

func MapStringRoles(a Roles) StringRoles {
	return StringRoles{
		RoleID:      strconv.FormatInt(a.RoleID, 10),
		Label:       a.Label,
		Permissions: a.Permissions,
	}
}

func MapStringRoute(a Routes) StringRoutes {
	return StringRoutes{
		RouteID:      strconv.FormatInt(a.RouteID, 10),
		Slug:         a.Slug,
		Title:        a.Title,
		Status:       strconv.FormatInt(a.Status, 10),
		Author:       a.Author,
		AuthorID:     strconv.FormatInt(a.AuthorID, 10),
		History:      a.History.String,
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
	}
}

func MapStringTables(a Tables) StringTables {
	return StringTables{
		ID:       strconv.FormatInt(a.ID, 10),
		Label:    a.Label.String,
		AuthorID: strconv.FormatInt(a.AuthorID, 10),
	}
}
func MapStringSession(a Sessions) StringSessions {
	return StringSessions{
		SessionID:   strconv.FormatInt(a.SessionID, 10),
		UserID:      strconv.FormatInt(a.UserID, 10),
		CreatedAt:   a.CreatedAt.String,
		ExpiresAt:   a.ExpiresAt.String,
		LastAccess:  a.LastAccess.String,
		IpAddress:   a.IpAddress.String,
		UserAgent:   a.UserAgent.String,
		SessionData: a.SessionData.String,
	}
}

func MapStringToken(a Tokens) StringTokens {
	return StringTokens{
		ID:        strconv.FormatInt(a.ID, 10),
		UserID:    strconv.FormatInt(a.UserID, 10),
		TokenType: a.TokenType,
		Token:     a.Token,
		IssuedAt:  a.IssuedAt,
		ExpiresAt: a.ExpiresAt,
		Revoked:   strconv.FormatBool(a.Revoked.Bool),
	}
}

func MapStringUser(a Users) StringUsers {
	return StringUsers{
		UserID:       strconv.FormatInt(a.UserID, 10),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         strconv.FormatInt(a.Role, 10),
		References:   AssertString(a.References),
		DateCreated:  a.DateCreated.String,
		DateModified: a.DateModified.String,
	}
}
func MapStringUserOauth(a UserOauth) StringUserOauth {
	return StringUserOauth{
		UserOauthID:         strconv.FormatInt(a.UserOauthID, 10),
		UserID:              strconv.FormatInt(a.UserID, 10),
		OauthProvider:       a.OauthProvider,
		OauthProviderUserID: a.OauthProviderUserID,
		AccessToken:         a.AccessToken.String,
		RefreshToken:        a.RefreshToken.String,
		TokenExpiresAt:      a.TokenExpiresAt.String,
		DateCreated:         a.DateCreated.String,
	}
}

type StringAdminContentData struct {
	AdminContentDataID string `json:"admin_content_data_id"`
	AdminRouteID       string `json:"admin_route_id"`
	AdminDatatypeID    string `json:"admin_datatype_id"`
	History            string `json:"history"`
	DateCreated        string `json:"date_created"`
	DateModified       string `json:"date_modified"`
}

type StringAdminContentFields struct {
	AdminContentFieldID string `json:"admin_content_field_id"`
	AdminRouteID        string `json:"admin_route_id"`
	AdminContentDataID  string `json:"admin_content_data_id"`
	AdminFieldID        string `json:"admin_field_id"`
	AdminFieldValue     string `json:"admin_field_value"`
	History             string `json:"history"`
	DateCreated         string `json:"date_created"`
	DateModified        string `json:"date_modified"`
}

type StringAdminDatatypes struct {
	AdminDatatypeID string `json:"admin_datatype_id"`
	ParentID        string `json:"parent_id"`
	Label           string `json:"label"`
	Type            string `json:"type"`
	Author          string `json:"author"`
	AuthorID        string `json:"author_id"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
	History         string `json:"history"`
}

type StringAdminFields struct {
	AdminFieldID string `json:"admin_field_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type StringAdminRoutes struct {
	AdminRouteID string `json:"admin_route_id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	History      string `json:"history"`
}

type StringContentData struct {
	ContentDataID string `json:"content_data_id"`
	RouteID       string `json:"route_id"`
	DatatypeID    string `json:"datatype_id"`
	History       string `json:"history"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
}

type StringContentFields struct {
	ContentFieldID string `json:"content_field_id"`
	RouteID        string `json:"route_id"`
	ContentDataID  string `json:"content_data_id"`
	FieldID        string `json:"field_id"`
	FieldValue     string `json:"field_value"`
	History        string `json:"history"`
	DateCreated    string `json:"date_created"`
	DateModified   string `json:"date_modified"`
}

type StringDatatypes struct {
	DatatypeID   string `json:"datatype_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type StringFields struct {
	FieldID      string `json:"field_id"`
	ParentID     string `json:"parent_id"`
	Label        string `json:"label"`
	Data         string `json:"data"`
	Type         string `json:"type"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type StringMedia struct {
	MediaID      string `json:"media_id"`
	Name         string `json:"name"`
	DisplayName  string `json:"display_name"`
	Alt          string `json:"alt"`
	Caption      string `json:"caption"`
	Description  string `json:"description"`
	Class        string `json:"class"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
	Mimetype     string `json:"mimetype"`
	Dimensions   string `json:"dimensions"`
	Url          string `json:"url"`
	Srcset       string `json:"srcset"`
}

type StringMediaDimensions struct {
	MdID        string `json:"md_id"`
	Label       string `json:"label"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	AspectRatio string `json:"aspect_ratio"`
}

type StringRoles struct {
	RoleID      string `json:"role_id"`
	Label       string `json:"label"`
	Permissions string `json:"permissions"`
}

type StringRoutes struct {
	RouteID      string `json:"route_id"`
	Author       string `json:"author"`
	AuthorID     string `json:"author_id"`
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	History      string `json:"history"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type StringSessions struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at"`
	LastAccess  string `json:"last_access"`
	IpAddress   string `json:"ip_address"`
	UserAgent   string `json:"user_agent"`
	SessionData string `json:"session_data"`
}

type StringTables struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	AuthorID string `json:"author_id"`
}

type StringTokens struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	IssuedAt  string `json:"issued_at"`
	ExpiresAt string `json:"expires_at"`
	Revoked   string `json:"revoked"`
}

type StringUsers struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Hash         string `json:"hash"`
	Role         string `json:"role"`
	References   string `json:"references"`
	DateCreated  string `json:"date_created"`
	DateModified string `json:"date_modified"`
}

type StringUserOauth struct {
	UserOauthID         string `json:"user_oauth_id"`
	UserID              string `json:"user_id"`
	OauthProvider       string `json:"oauth_provider"`
	OauthProviderUserID string `json:"oauth_provider_user_id"`
	AccessToken         string `json:"access_token"`
	RefreshToken        string `json:"refresh_token"`
	TokenExpiresAt      string `json:"token_expires_at"`
	DateCreated         string `json:"date_created"`
}
