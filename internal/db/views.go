package db

import (
	"github.com/hegner123/modulacms/internal/db/types"
)

// AuthorView is a safe-to-serialize subset of Users (excludes hash).
type AuthorView struct {
	UserID   types.UserID `json:"user_id"`
	Username string       `json:"username"`
	Name     string       `json:"name"`
	Email    types.Email  `json:"email"`
	Role     string       `json:"role"`
}

// DatatypeView is the embedded datatype within a composed response.
type DatatypeView struct {
	DatatypeID types.DatatypeID `json:"datatype_id"`
	Label      string           `json:"label"`
	Type       string           `json:"type"`
}

// FieldView pairs a field definition with its content value.
type FieldView struct {
	FieldID types.FieldID   `json:"field_id"`
	Label   string          `json:"label"`
	Type    types.FieldType `json:"type"`
	Value   string          `json:"value"`
}

// ContentDataView is a composed response for content data with embedded relations.
type ContentDataView struct {
	ContentDataID types.ContentID     `json:"content_data_id"`
	Status        types.ContentStatus `json:"status"`
	DateCreated   types.Timestamp     `json:"date_created"`
	DateModified  types.Timestamp     `json:"date_modified"`
	Datatype      *DatatypeView       `json:"datatype,omitempty"`
	Author        *AuthorView         `json:"author,omitempty"`
	Fields        []FieldView         `json:"fields"`
}

// DatatypeFieldView is a field definition with sort order for the datatype full view.
type DatatypeFieldView struct {
	FieldID    types.FieldID        `json:"field_id"`
	Label      string               `json:"label"`
	Type       types.FieldType      `json:"type"`
	Data       string               `json:"data"`
	ValidationID types.NullableValidationID `json:"validation_id"`
	UIConfig   string               `json:"ui_config"`
	SortOrder  int64                `json:"sort_order"`
	Roles      types.NullableString `json:"roles"`
}

// DatatypeFullView is a composed response for a datatype with all field definitions.
type DatatypeFullView struct {
	DatatypeID   types.DatatypeID         `json:"datatype_id"`
	Label        string                   `json:"label"`
	Type         string                   `json:"type"`
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	Author       *AuthorView              `json:"author,omitempty"`
	Fields       []DatatypeFieldView      `json:"fields"`
	DateCreated  types.Timestamp          `json:"date_created"`
	DateModified types.Timestamp          `json:"date_modified"`
}

// MapDatatypeFieldView converts a FieldWithSortOrderRow to a DatatypeFieldView.
func MapDatatypeFieldView(row FieldWithSortOrderRow) DatatypeFieldView {
	return DatatypeFieldView{
		FieldID:    row.FieldID,
		Label:      row.Label,
		Type:       row.Type,
		Data:       row.Data,
		ValidationID: row.ValidationID,
		UIConfig:   row.UIConfig,
		SortOrder:  row.SortOrder,
		Roles:      row.Roles,
	}
}

// UserOauthView is a safe subset of UserOauth (excludes access/refresh tokens).
type UserOauthView struct {
	UserOauthID         types.UserOauthID `json:"user_oauth_id"`
	OauthProvider       string            `json:"oauth_provider"`
	OauthProviderUserID string            `json:"oauth_provider_user_id"`
	TokenExpiresAt      string            `json:"token_expires_at"`
	DateCreated         types.Timestamp   `json:"date_created"`
}

// UserSshKeyView is a view of an SSH key (excludes raw public key).
type UserSshKeyView struct {
	SshKeyID    string          `json:"ssh_key_id"`
	KeyType     string          `json:"key_type"`
	Fingerprint string          `json:"fingerprint"`
	Label       string          `json:"label"`
	DateCreated types.Timestamp `json:"date_created"`
	LastUsed    string          `json:"last_used"`
}

// SessionView is a safe subset of Sessions (excludes session_data).
type SessionView struct {
	SessionID   types.SessionID `json:"session_id"`
	DateCreated types.Timestamp `json:"date_created"`
	ExpiresAt   types.Timestamp `json:"expires_at"`
	LastAccess  string          `json:"last_access"`
	IpAddress   string          `json:"ip_address"`
	UserAgent   string          `json:"user_agent"`
}

// TokenView is a safe subset of Tokens (excludes token value).
type TokenView struct {
	ID        string          `json:"id"`
	TokenType string          `json:"token_type"`
	IssuedAt  string          `json:"issued_at"`
	ExpiresAt types.Timestamp `json:"expires_at"`
	Revoked   bool            `json:"revoked"`
}

// UserFullView is a composed response for a single user with all related entities.
type UserFullView struct {
	UserID       types.UserID     `json:"user_id"`
	Username     string           `json:"username"`
	Name         string           `json:"name"`
	Email        types.Email      `json:"email"`
	RoleID       string           `json:"role_id"`
	RoleLabel    string           `json:"role_label"`
	DateCreated  types.Timestamp  `json:"date_created"`
	DateModified types.Timestamp  `json:"date_modified"`
	Oauth        *UserOauthView   `json:"oauth,omitempty"`
	SshKeys      []UserSshKeyView `json:"ssh_keys"`
	Sessions     *SessionView     `json:"sessions,omitempty"`
	Tokens       []TokenView      `json:"tokens"`
}

// MapUserOauthView converts a UserOauth to a safe view (strips tokens).
func MapUserOauthView(o UserOauth) UserOauthView {
	return UserOauthView{
		UserOauthID:         o.UserOauthID,
		OauthProvider:       o.OauthProvider,
		OauthProviderUserID: o.OauthProviderUserID,
		TokenExpiresAt:      o.TokenExpiresAt.String(),
		DateCreated:         o.DateCreated,
	}
}

// MapUserSshKeyView converts a UserSshKeys to a view (strips public key).
func MapUserSshKeyView(k UserSshKeys) UserSshKeyView {
	return UserSshKeyView{
		SshKeyID:    k.SshKeyID,
		KeyType:     k.KeyType,
		Fingerprint: k.Fingerprint,
		Label:       k.Label,
		DateCreated: k.DateCreated,
		LastUsed:    k.LastUsed,
	}
}

// MapSessionView converts a Sessions to a safe view (strips session_data).
func MapSessionView(s Sessions) SessionView {
	return SessionView{
		SessionID:   s.SessionID,
		DateCreated: s.DateCreated,
		ExpiresAt:   s.ExpiresAt,
		LastAccess:  s.LastAccess.String(),
		IpAddress:   nullStringValue(s.IpAddress),
		UserAgent:   nullStringValue(s.UserAgent),
	}
}

// MapTokenView converts a Tokens to a safe view (strips token value).
func MapTokenView(t Tokens) TokenView {
	return TokenView{
		ID:        t.ID,
		TokenType: t.TokenType,
		IssuedAt:  t.IssuedAt.String(),
		ExpiresAt: t.ExpiresAt,
		Revoked:   t.Revoked,
	}
}

// nullStringValue extracts the string from a NullString, returning "" if not valid.
func nullStringValue(ns NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// MediaFullView is a composed response for a media item with embedded author.
type MediaFullView struct {
	MediaID      types.MediaID         `json:"media_id"`
	Name         NullString            `json:"name"`
	DisplayName  NullString            `json:"display_name"`
	Alt          NullString            `json:"alt"`
	Caption      NullString            `json:"caption"`
	Description  NullString            `json:"description"`
	Mimetype     NullString            `json:"mimetype"`
	Dimensions   NullString            `json:"dimensions"`
	URL          types.URL             `json:"url"`
	Srcset       NullString            `json:"srcset"`
	FocalX       types.NullableFloat64 `json:"focal_x"`
	FocalY       types.NullableFloat64 `json:"focal_y"`
	Author       *AuthorView           `json:"author,omitempty"`
	DateCreated  types.Timestamp       `json:"date_created"`
	DateModified types.Timestamp       `json:"date_modified"`
}

// RouteContentNodeView is a flattened content tree node for the route full view.
type RouteContentNodeView struct {
	ContentDataID types.ContentID         `json:"content_data_id"`
	ParentID      types.NullableContentID `json:"parent_id"`
	DatatypeLabel string                  `json:"datatype_label"`
	DatatypeType  string                  `json:"datatype_type"`
	Status        types.ContentStatus     `json:"status"`
	DateCreated   types.Timestamp         `json:"date_created"`
	DateModified  types.Timestamp         `json:"date_modified"`
}

// RouteFullView is a composed response for a route with its content tree.
type RouteFullView struct {
	RouteID      types.RouteID          `json:"route_id"`
	Slug         types.Slug             `json:"slug"`
	Title        string                 `json:"title"`
	Status       int64                  `json:"status"`
	Author       *AuthorView            `json:"author,omitempty"`
	ContentTree  []RouteContentNodeView `json:"content_tree"`
	DateCreated  types.Timestamp        `json:"date_created"`
	DateModified types.Timestamp        `json:"date_modified"`
}

// ActivityEventView is a composed change event with actor info for dashboards.
type ActivityEventView struct {
	EventID       types.EventID   `json:"event_id"`
	TableName     string          `json:"table_name"`
	RecordID      string          `json:"record_id"`
	Operation     types.Operation `json:"operation"`
	Action        types.Action    `json:"action"`
	Actor         *AuthorView     `json:"actor,omitempty"`
	WallTimestamp types.Timestamp `json:"timestamp"`
}

// DatatypeListItemView is a datatype with field count and parent label for list views.
type DatatypeListItemView struct {
	DatatypeID   types.DatatypeID         `json:"datatype_id"`
	Label        string                   `json:"label"`
	Type         string                   `json:"type"`
	ParentID     types.NullableDatatypeID `json:"parent_id"`
	ParentLabel  string                   `json:"parent_label,omitempty"`
	FieldCount   int                      `json:"field_count"`
	Author       *AuthorView              `json:"author,omitempty"`
	DateCreated  types.Timestamp          `json:"date_created"`
	DateModified types.Timestamp          `json:"date_modified"`
}

// AdminDatatypeFullView is a composed response for an admin datatype with field definitions.
type AdminDatatypeFullView struct {
	AdminDatatypeID types.AdminDatatypeID         `json:"admin_datatype_id"`
	Label           string                        `json:"label"`
	Type            string                        `json:"type"`
	ParentID        types.NullableAdminDatatypeID `json:"parent_id"`
	SortOrder       int64                         `json:"sort_order"`
	Author          *AuthorView                   `json:"author,omitempty"`
	Fields          []AdminFieldView              `json:"fields"`
	DateCreated     types.Timestamp               `json:"date_created"`
	DateModified    types.Timestamp               `json:"date_modified"`
}

// AdminFieldView is a field definition in the admin namespace.
type AdminFieldView struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	Label        string            `json:"label"`
	Type         types.FieldType   `json:"type"`
	Data         string            `json:"data"`
	ValidationID types.NullableAdminValidationID `json:"validation_id"`
	UIConfig     string            `json:"ui_config"`
	SortOrder    int64             `json:"sort_order"`
}

// AdminContentDataView is a composed response for admin content with embedded relations.
type AdminContentDataView struct {
	AdminContentDataID types.AdminContentID     `json:"admin_content_data_id"`
	Status             types.ContentStatus      `json:"status"`
	Revision           int64                    `json:"revision"`
	DateCreated        types.Timestamp          `json:"date_created"`
	DateModified       types.Timestamp          `json:"date_modified"`
	Datatype           *AdminDatatypeView       `json:"datatype,omitempty"`
	Author             *AuthorView              `json:"author,omitempty"`
	Fields             []AdminContentFieldView  `json:"fields"`
}

// AdminDatatypeView is an embedded admin datatype within a composed response.
type AdminDatatypeView struct {
	AdminDatatypeID types.AdminDatatypeID `json:"admin_datatype_id"`
	Label           string               `json:"label"`
	Type            string               `json:"type"`
}

// AdminContentFieldView pairs an admin field definition with its content value.
type AdminContentFieldView struct {
	AdminFieldID types.AdminFieldID `json:"admin_field_id"`
	Label        string            `json:"label"`
	Type         types.FieldType   `json:"type"`
	Value        string            `json:"value"`
}

// MapMediaFullView converts a Media entity to a MediaFullView.
func MapMediaFullView(m Media) MediaFullView {
	return MediaFullView{
		MediaID:      m.MediaID,
		Name:         m.Name,
		DisplayName:  m.DisplayName,
		Alt:          m.Alt,
		Caption:      m.Caption,
		Description:  m.Description,
		Mimetype:     m.Mimetype,
		Dimensions:   m.Dimensions,
		URL:          m.URL,
		Srcset:       m.Srcset,
		FocalX:       m.FocalX,
		FocalY:       m.FocalY,
		DateCreated:  m.DateCreated,
		DateModified: m.DateModified,
	}
}

// MapAuthorView converts a Users entity to an AuthorView, excluding the hash.
func MapAuthorView(u Users) AuthorView {
	return AuthorView{
		UserID:   u.UserID,
		Username: u.Username,
		Name:     u.Name,
		Email:    u.Email,
		Role:     u.Role,
	}
}

// MapDatatypeView converts a Datatypes entity to a DatatypeView.
func MapDatatypeView(d Datatypes) DatatypeView {
	return DatatypeView{
		DatatypeID: d.DatatypeID,
		Label:      d.Label,
		Type:       d.Type,
	}
}

// MapFieldView converts a ContentFields and its associated Fields definition into a FieldView.
func MapFieldView(cf ContentFields, f Fields) FieldView {
	return FieldView{
		FieldID: f.FieldID,
		Label:   f.Label,
		Type:    f.Type,
		Value:   cf.FieldValue,
	}
}

// MapFieldViewFromRow converts a ContentFieldWithFieldRow (JOIN result) into a FieldView.
func MapFieldViewFromRow(row ContentFieldWithFieldRow) FieldView {
	return FieldView{
		FieldID: row.FFieldID,
		Label:   row.FLabel,
		Type:    row.FType,
		Value:   row.FieldValue,
	}
}
