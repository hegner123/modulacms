package db

import "reflect"

type DBTable string

const (
	Admin_content_data    DBTable = "admin_content_data"
	Admin_content_fields  DBTable = "admin_content_fields"
	Admin_datatype        DBTable = "admin_datatypes"
	Admin_datatype_fields DBTable = "admin_datatypes_fields"
	Admin_field           DBTable = "admin_fields"
	Admin_route           DBTable = "admin_routes"
	Content_data          DBTable = "content_data"
	Content_fields        DBTable = "content_fields"
	Datatype_fields       DBTable = "datatypes_fields"
	Datatype              DBTable = "datatypes"
	Field                 DBTable = "fields"
	MediaT                DBTable = "media"
	Media_dimension       DBTable = "media_dimensions"
	Permission            DBTable = "permissions"
	Role                  DBTable = "roles"
	Route                 DBTable = "routes"
	Session               DBTable = "sessions"
	Table                 DBTable = "tables"
	Token                 DBTable = "tokens"
	User                  DBTable = "users"
	User_oauth            DBTable = "user_oauth"
)

// TableStructMap maps each DBTable to its associated struct type
var TableStructMap = map[DBTable]reflect.Type{
	Admin_content_data:    reflect.TypeOf(AdminContentData{}),
	Admin_content_fields:  reflect.TypeOf(AdminContentFields{}),
	Admin_datatype:        reflect.TypeOf(AdminDatatypes{}),
	Admin_datatype_fields: reflect.TypeOf(AdminDatatypeFields{}),
	Admin_field:           reflect.TypeOf(AdminFields{}),
	Admin_route:           reflect.TypeOf(AdminRoutes{}),
	Content_data:          reflect.TypeOf(ContentData{}),
	Content_fields:        reflect.TypeOf(ContentFields{}),
	Datatype_fields:       reflect.TypeOf(DatatypeFields{}),
	Datatype:              reflect.TypeOf(Datatypes{}),
	Field:                 reflect.TypeOf(Fields{}),
	MediaT:                reflect.TypeOf(Media{}),
	Media_dimension:       reflect.TypeOf(MediaDimensions{}),
	Permission:            reflect.TypeOf(Permissions{}),
	Role:                  reflect.TypeOf(Roles{}),
	Route:                 reflect.TypeOf(Routes{}),
	Session:               reflect.TypeOf(Sessions{}),
	Table:                 reflect.TypeOf(Tables{}),
	Token:                 reflect.TypeOf(Tokens{}),
	User:                  reflect.TypeOf(Users{}),
	User_oauth:            reflect.TypeOf(UserOauth{}),
}

// CastToTypedSlice casts an any return from Parse to a typed slice based on the DBTable
func CastToTypedSlice(result any, table DBTable) any {
	if result == nil {
		return nil
	}

	switch table {
	case Admin_content_data:
		if slice, ok := result.([]AdminContentData); ok {
			return slice
		}
	case Admin_content_fields:
		if slice, ok := result.([]AdminContentFields); ok {
			return slice
		}
	case Admin_datatype:
		if slice, ok := result.([]AdminDatatypes); ok {
			return slice
		}
	case Admin_datatype_fields:
		if slice, ok := result.([]AdminDatatypeFields); ok {
			return slice
		}
	case Admin_field:
		if slice, ok := result.([]AdminFields); ok {
			return slice
		}
	case Admin_route:
		if slice, ok := result.([]AdminRoutes); ok {
			return slice
		}
	case Content_data:
		if slice, ok := result.([]ContentData); ok {
			return slice
		}
	case Content_fields:
		if slice, ok := result.([]ContentFields); ok {
			return slice
		}
	case Datatype_fields:
		if slice, ok := result.([]DatatypeFields); ok {
			return slice
		}
	case Datatype:
		if slice, ok := result.([]Datatypes); ok {
			return slice
		}
	case Field:
		if slice, ok := result.([]Fields); ok {
			return slice
		}
	case MediaT:
		if slice, ok := result.([]Media); ok {
			return slice
		}
	case Media_dimension:
		if slice, ok := result.([]MediaDimensions); ok {
			return slice
		}
	case Permission:
		if slice, ok := result.([]Permissions); ok {
			return slice
		}
	case Role:
		if slice, ok := result.([]Roles); ok {
			return slice
		}
	case Route:
		if slice, ok := result.([]Routes); ok {
			return slice
		}
	case Session:
		if slice, ok := result.([]Sessions); ok {
			return slice
		}
	case Table:
		if slice, ok := result.([]Tables); ok {
			return slice
		}
	case Token:
		if slice, ok := result.([]Tokens); ok {
			return slice
		}
	case User:
		if slice, ok := result.([]Users); ok {
			return slice
		}
	case User_oauth:
		if slice, ok := result.([]UserOauth); ok {
			return slice
		}
	}

	// Return as-is if no match (could be []map[string]any from parseGeneric)
	return result
}
