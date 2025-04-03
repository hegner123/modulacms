package db

type DBTable string

const (
	Admin_content_data    DBTable = "admin_content_data"
	Admin_content_fields  DBTable = "admin_content_fields"
	Admin_datatype        DBTable = "admin_datatypes"
	Admin_datatype_fields DBTable = "admin_datatype_fields"
	Admin_field           DBTable = "admin_fields"
	Admin_route           DBTable = "admin_routes"
	Content_data          DBTable = "content_data"
	Content_fields        DBTable = "content_fields"
	Datatype_fields       DBTable = "datatype_fields"
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
