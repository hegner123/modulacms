package modulacms

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

// ClientConfig configures a new ModulaCMS client.
type ClientConfig struct {
	// BaseURL is the root URL of the ModulaCMS server (e.g. "https://cms.example.com").
	// Required.
	BaseURL string

	// APIKey is the Bearer token for authentication. Optional.
	APIKey string

	// HTTPClient is the underlying HTTP client to use. Optional.
	// Defaults to a client with a 30-second timeout.
	HTTPClient *http.Client
}

// Client is the ModulaCMS API client. It provides typed access to all API resources.
type Client struct {
	// Standard CRUD resources
	ContentData     *Resource[ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID]
	ContentFields   *Resource[ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID]
	ContentRelations *Resource[ContentRelation, CreateContentRelationParams, UpdateContentRelationParams, ContentRelationID]
	Datatypes       *Resource[Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID]
	DatatypeFields  *Resource[DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams, DatatypeFieldID]
	Fields          *Resource[Field, CreateFieldParams, UpdateFieldParams, FieldID]
	Media           *Resource[Media, any, UpdateMediaParams, MediaID]
	MediaDimensions *Resource[MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams, MediaDimensionID]
	Routes          *Resource[Route, CreateRouteParams, UpdateRouteParams, RouteID]
	Roles           *Resource[Role, CreateRoleParams, UpdateRoleParams, RoleID]
	Users           *Resource[User, CreateUserParams, UpdateUserParams, UserID]
	Tokens          *Resource[Token, CreateTokenParams, UpdateTokenParams, TokenID]
	UsersOauth      *Resource[UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID]
	Tables          *Resource[Table, CreateTableParams, UpdateTableParams, TableID]

	// Admin CRUD resources
	AdminContentData    *Resource[AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID]
	AdminContentFields  *Resource[AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID]
	AdminDatatypes      *Resource[AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID]
	AdminDatatypeFields *Resource[AdminDatatypeField, CreateAdminDatatypeFieldParams, UpdateAdminDatatypeFieldParams, AdminDatatypeFieldID]
	AdminFields         *Resource[AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID]
	AdminRoutes         *Resource[AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, AdminRouteID]

	// Specialized resources
	Auth         *AuthResource
	MediaUpload  *MediaUploadResource
	AdminTree    *AdminTreeResource
	Content      *ContentDeliveryResource
	SSHKeys      *SSHKeysResource
	Sessions     *SessionsResource
	Import       *ImportResource
	ContentBatch *ContentBatchResource
}

// NewClient creates a new ModulaCMS API client.
func NewClient(cfg ClientConfig) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("modulacms: BaseURL is required")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")

	httpCl := cfg.HTTPClient
	if httpCl == nil {
		httpCl = &http.Client{Timeout: 30 * time.Second}
	}

	h := &httpClient{
		baseURL:    baseURL,
		apiKey:     cfg.APIKey,
		httpClient: httpCl,
	}

	return &Client{
		// Standard CRUD
		ContentData:      newResource[ContentData, CreateContentDataParams, UpdateContentDataParams, ContentID](h, "/api/v1/contentdata"),
		ContentFields:    newResource[ContentField, CreateContentFieldParams, UpdateContentFieldParams, ContentFieldID](h, "/api/v1/contentfields"),
		ContentRelations: newResource[ContentRelation, CreateContentRelationParams, UpdateContentRelationParams, ContentRelationID](h, "/api/v1/contentrelations"),
		Datatypes:        newResource[Datatype, CreateDatatypeParams, UpdateDatatypeParams, DatatypeID](h, "/api/v1/datatype"),
		DatatypeFields:   newResource[DatatypeField, CreateDatatypeFieldParams, UpdateDatatypeFieldParams, DatatypeFieldID](h, "/api/v1/datatypefields"),
		Fields:           newResource[Field, CreateFieldParams, UpdateFieldParams, FieldID](h, "/api/v1/fields"),
		Media:            newResource[Media, any, UpdateMediaParams, MediaID](h, "/api/v1/media"),
		MediaDimensions:  newResource[MediaDimension, CreateMediaDimensionParams, UpdateMediaDimensionParams, MediaDimensionID](h, "/api/v1/mediadimensions"),
		Routes:           newResource[Route, CreateRouteParams, UpdateRouteParams, RouteID](h, "/api/v1/routes"),
		Roles:            newResource[Role, CreateRoleParams, UpdateRoleParams, RoleID](h, "/api/v1/roles"),
		Users:            newResource[User, CreateUserParams, UpdateUserParams, UserID](h, "/api/v1/users"),
		Tokens:           newResource[Token, CreateTokenParams, UpdateTokenParams, TokenID](h, "/api/v1/tokens"),
		UsersOauth:       newResource[UserOauth, CreateUserOauthParams, UpdateUserOauthParams, UserOauthID](h, "/api/v1/usersoauth"),
		Tables:           newResource[Table, CreateTableParams, UpdateTableParams, TableID](h, "/api/v1/tables"),

		// Admin CRUD
		AdminContentData:    newResource[AdminContentData, CreateAdminContentDataParams, UpdateAdminContentDataParams, AdminContentID](h, "/api/v1/admincontentdatas"),
		AdminContentFields:  newResource[AdminContentField, CreateAdminContentFieldParams, UpdateAdminContentFieldParams, AdminContentFieldID](h, "/api/v1/admincontentfields"),
		AdminDatatypes:      newResource[AdminDatatype, CreateAdminDatatypeParams, UpdateAdminDatatypeParams, AdminDatatypeID](h, "/api/v1/admindatatypes"),
		AdminDatatypeFields: newResource[AdminDatatypeField, CreateAdminDatatypeFieldParams, UpdateAdminDatatypeFieldParams, AdminDatatypeFieldID](h, "/api/v1/admindatatypefields"),
		AdminFields:         newResource[AdminField, CreateAdminFieldParams, UpdateAdminFieldParams, AdminFieldID](h, "/api/v1/adminfields"),
		AdminRoutes:         newResource[AdminRoute, CreateAdminRouteParams, UpdateAdminRouteParams, AdminRouteID](h, "/api/v1/adminroutes"),

		// Specialized
		Auth:         &AuthResource{http: h},
		MediaUpload:  &MediaUploadResource{http: h},
		AdminTree:    &AdminTreeResource{http: h},
		Content:      &ContentDeliveryResource{http: h},
		SSHKeys:      &SSHKeysResource{http: h},
		Sessions:     &SessionsResource{http: h},
		Import:       &ImportResource{http: h},
		ContentBatch: &ContentBatchResource{http: h},
	}, nil
}

func newResource[E any, C any, U any, ID ~string](h *httpClient, path string) *Resource[E, C, U, ID] {
	return &Resource[E, C, U, ID]{
		path: path,
		http: h,
	}
}
