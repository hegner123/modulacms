package main

// Entities defines all entity configurations for code generation.
var Entities = []Entity{
	// ========================
	// FULL GENERATION ENTITIES
	// ========================

	// Users
	{
		Name:               "Users",
		Singular:           "User",
		Plural:             "Users",
		SqlcTypeName:       "Users",
		TableName:          "users",
		IDType:             "types.UserID",
		IDField:            "UserID",
		NewIDFunc:          "types.NewUserID()",
		UpdateSuccessField: "s.Username",
		StringTypeName:     "StringUsers",
		Fields: []Field{
			{AppName: "UserID", Type: "types.UserID", JSONTag: "user_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "Username", Type: "string", JSONTag: "username", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Name", Type: "string", JSONTag: "name", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Email", Type: "types.Email", JSONTag: "email", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Hash", Type: "string", JSONTag: "hash", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Role", SqlcName: "Roles", Type: "string", JSONTag: "role", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "GetUserByEmail",
				SqlcName:    "GetUserByEmail",
				ReturnsList: false,
				Params: []ExtraQueryParam{
					{ParamName: "email", ParamType: "types.Email", SqlcField: "Email"},
				},
			},
		},
		OutputFile: "user_gen.go",
	},

	// Media
	{
		Name:               "Media",
		Singular:           "Media",
		Plural:             "Media",
		SqlcTypeName:       "Media",
		TableName:          "media",
		IDType:             "types.MediaID",
		IDField:            "MediaID",
		NewIDFunc:          "types.NewMediaID()",
		HasPaginated:       true,
		UpdateSuccessField: "s.MediaID",
		StringTypeName:     "StringMedia",
		Fields: []Field{
			{AppName: "MediaID", Type: "types.MediaID", JSONTag: "media_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "Name", Type: "sql.NullString", JSONTag: "name", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "DisplayName", Type: "sql.NullString", JSONTag: "display_name", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Alt", Type: "sql.NullString", JSONTag: "alt", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Caption", Type: "sql.NullString", JSONTag: "caption", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Description", Type: "sql.NullString", JSONTag: "description", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Class", Type: "sql.NullString", JSONTag: "class", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Mimetype", Type: "sql.NullString", JSONTag: "mimetype", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Dimensions", Type: "sql.NullString", JSONTag: "dimensions", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "URL", Type: "types.URL", JSONTag: "url", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Srcset", Type: "sql.NullString", JSONTag: "srcset", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "FocalX", Type: "types.NullableFloat64", JSONTag: "focal_x", InCreate: true, InUpdate: true, StringConvert: "sprintfFloat64"},
			{AppName: "FocalY", Type: "types.NullableFloat64", JSONTag: "focal_y", InCreate: true, InUpdate: true, StringConvert: "sprintfFloat64"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "GetMediaByName",
				SqlcName:    "GetMediaByName",
				ReturnsList: false,
				Params: []ExtraQueryParam{
					{ParamName: "name", ParamType: "string", SqlcField: "Name", WrapExpr: "StringToNullString(%s)"},
				},
			},
			{
				MethodName:  "GetMediaByURL",
				SqlcName:    "GetMediaByUrl",
				ReturnsList: false,
				Params: []ExtraQueryParam{
					{ParamName: "url", ParamType: "types.URL", SqlcField: "URL"},
				},
			},
		},
		OutputFile: "media_gen.go",
	},

	// Tables
	{
		Name:                "Tables",
		Singular:            "Table",
		Plural:              "Tables",
		SqlcTypeName:        "Tables",
		TableName:           "tables",
		IDType:              "string",
		IDField:             "ID",
		NewIDFunc:           "string(types.NewTableID())",
		UpdateSuccessField:  "s.ID",
		SqlcCreateTableName: "CreateTablesTable",
		SqlcCountName:       "CountTables",
		StringTypeName:      "StringTables",
		Fields: []Field{
			{AppName: "ID", Type: "string", JSONTag: "id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "string"},
			{AppName: "Label", Type: "string", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: false, InUpdate: false, StringConvert: "toString"},
		},
		OutputFile: "table_gen.go",
	},

	// ContentData
	{
		Name:               "ContentData",
		Singular:           "ContentData",
		Plural:             "ContentData",
		SqlcTypeName:       "ContentData",
		TableName:          "content_data",
		IDType:             "types.ContentID",
		IDField:            "ContentDataID",
		NewIDFunc:          "types.NewContentID()",
		HasPaginated:       true,
		UpdateSuccessField: "s.ContentDataID",
		StringTypeName:     "StringContentData",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "ContentDataID", Type: "types.ContentID", JSONTag: "content_data_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "ParentID", Type: "types.NullableContentID", JSONTag: "parent_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "FirstChildID", Type: "types.NullableContentID", JSONTag: "first_child_id", InCreate: true, InUpdate: true, StringConvert: "nullableIDToEmpty"},
			{AppName: "NextSiblingID", Type: "types.NullableContentID", JSONTag: "next_sibling_id", InCreate: true, InUpdate: true, StringConvert: "nullableIDToEmpty"},
			{AppName: "PrevSiblingID", Type: "types.NullableContentID", JSONTag: "prev_sibling_id", InCreate: true, InUpdate: true, StringConvert: "nullableIDToEmpty"},
			{AppName: "RouteID", Type: "types.NullableRouteID", JSONTag: "route_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DatatypeID", Type: "types.NullableDatatypeID", JSONTag: "datatype_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AuthorID", Type: "types.UserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Status", Type: "types.ContentStatus", JSONTag: "status", InCreate: true, InUpdate: true, StringConvert: "cast"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListContentDataByRoute",
				SqlcName:    "ListContentDataByRoute",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "routeID", ParamType: "types.NullableRouteID", SqlcField: "RouteID"},
				},
			},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListContentDataByRoutePaginated",
				SqlcName:      "ListContentDataByRoutePaginated",
				AppParamsType: "ListContentDataByRoutePaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "RouteID", ParamType: "types.NullableRouteID", SqlcField: "RouteID"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListContentDataByRoutePaginatedParams",
				Fields: []ExtraParamField{
					{Name: "RouteID", Type: "types.NullableRouteID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "content_data_gen.go",
	},

	// ContentFields
	{
		Name:                  "ContentFields",
		Singular:              "ContentField",
		Plural:                "ContentFields",
		SqlcTypeName:          "ContentFields",
		TableName:             "content_fields",
		IDType:                "types.ContentFieldID",
		IDField:               "ContentFieldID",
		NewIDFunc:             "types.NewContentFieldID()",
		UpdateSuccessField:    "s.ContentFieldID",
		SqlcListName:          "ListContentFields",
		SqlcListPaginatedName: "ListContentFieldsPaginated",
		StringTypeName:        "StringContentFields",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "ContentFieldID", Type: "types.ContentFieldID", JSONTag: "content_field_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "RouteID", Type: "types.NullableRouteID", JSONTag: "route_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "ContentDataID", Type: "types.NullableContentID", JSONTag: "content_data_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "FieldID", Type: "types.NullableFieldID", JSONTag: "field_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "FieldValue", Type: "string", JSONTag: "field_value", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListContentFieldsByRoute",
				SqlcName:    "ListContentFieldsByRoute",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "routeID", ParamType: "types.NullableRouteID", SqlcField: "RouteID"},
				},
			},
			{
				MethodName:  "ListContentFieldsByContentData",
				SqlcName:    "ListContentFieldsByContentData",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "contentDataID", ParamType: "types.NullableContentID", SqlcField: "ContentDataID"},
				},
			},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListContentFieldsByRoutePaginated",
				SqlcName:      "ListContentFieldsByRoutePaginated",
				AppParamsType: "ListContentFieldsByRoutePaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "RouteID", ParamType: "types.NullableRouteID", SqlcField: "RouteID"},
				},
			},
			{
				MethodName:    "ListContentFieldsByContentDataPaginated",
				SqlcName:      "ListContentFieldsByContentDataPaginated",
				AppParamsType: "ListContentFieldsByContentDataPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "ContentDataID", ParamType: "types.NullableContentID", SqlcField: "ContentDataID"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListContentFieldsByRoutePaginatedParams",
				Fields: []ExtraParamField{
					{Name: "RouteID", Type: "types.NullableRouteID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
			{
				TypeName: "ListContentFieldsByContentDataPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "ContentDataID", Type: "types.NullableContentID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "content_field_gen.go",
	},

	// AdminContentData
	{
		Name:               "AdminContentData",
		Singular:           "AdminContentData",
		Plural:             "AdminContentData",
		SqlcTypeName:       "AdminContentData",
		TableName:          "admin_content_data",
		IDType:             "types.AdminContentID",
		IDField:            "AdminContentDataID",
		NewIDFunc:          "types.NewAdminContentID()",
		HasPaginated:       true,
		UpdateSuccessField: "s.AdminContentDataID",
		StringTypeName:     "StringAdminContentData",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "AdminContentDataID", Type: "types.AdminContentID", JSONTag: "admin_content_data_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "ParentID", Type: "types.NullableAdminContentID", JSONTag: "parent_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "FirstChildID", Type: "types.NullableAdminContentID", JSONTag: "first_child_id", InCreate: true, InUpdate: true, StringConvert: ""},
			{AppName: "NextSiblingID", Type: "types.NullableAdminContentID", JSONTag: "next_sibling_id", InCreate: true, InUpdate: true, StringConvert: ""},
			{AppName: "PrevSiblingID", Type: "types.NullableAdminContentID", JSONTag: "prev_sibling_id", InCreate: true, InUpdate: true, StringConvert: ""},
			{AppName: "AdminRouteID", Type: "types.NullableAdminRouteID", JSONTag: "admin_route_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AdminDatatypeID", Type: "types.NullableAdminDatatypeID", JSONTag: "admin_datatype_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AuthorID", Type: "types.UserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Status", Type: "types.ContentStatus", JSONTag: "status", InCreate: true, InUpdate: true, StringConvert: "cast"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListAdminContentDataByRoute",
				SqlcName:    "ListAdminContentDataByRoute",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "routeID", ParamType: "types.NullableAdminRouteID", SqlcField: "AdminRouteID"},
				},
			},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListAdminContentDataByRoutePaginated",
				SqlcName:      "ListAdminContentDataByRoutePaginated",
				AppParamsType: "ListAdminContentDataByRoutePaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "AdminRouteID", ParamType: "types.NullableAdminRouteID", SqlcField: "AdminRouteID"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListAdminContentDataByRoutePaginatedParams",
				Fields: []ExtraParamField{
					{Name: "AdminRouteID", Type: "types.NullableAdminRouteID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "admin_content_data_gen.go",
	},

	// AdminContentFields
	{
		Name:                  "AdminContentFields",
		Singular:              "AdminContentField",
		Plural:                "AdminContentFields",
		SqlcTypeName:          "AdminContentFields",
		TableName:             "admin_content_fields",
		IDType:                "types.AdminContentFieldID",
		IDField:               "AdminContentFieldID",
		NewIDFunc:             "types.NewAdminContentFieldID()",
		UpdateSuccessField:    "s.AdminContentFieldID",
		SqlcListName:          "ListAdminContentFields",
		SqlcListPaginatedName: "ListAdminContentFieldsPaginated",
		StringTypeName:        "StringAdminContentFields",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "AdminContentFieldID", Type: "types.AdminContentFieldID", JSONTag: "admin_content_field_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "AdminRouteID", Type: "types.NullableAdminRouteID", JSONTag: "admin_route_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AdminContentDataID", Type: "types.NullableAdminContentID", JSONTag: "admin_content_data_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AdminFieldID", Type: "types.NullableAdminFieldID", JSONTag: "admin_field_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AdminFieldValue", Type: "string", JSONTag: "admin_field_value", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListAdminContentFieldsByRoute",
				SqlcName:    "ListAdminContentFieldsByRoute",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "routeID", ParamType: "types.NullableAdminRouteID", SqlcField: "AdminRouteID"},
				},
			},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListAdminContentFieldsByRoutePaginated",
				SqlcName:      "ListAdminContentFieldsByRoutePaginated",
				AppParamsType: "ListAdminContentFieldsByRoutePaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "AdminRouteID", ParamType: "types.NullableAdminRouteID", SqlcField: "AdminRouteID"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListAdminContentFieldsByRoutePaginatedParams",
				Fields: []ExtraParamField{
					{Name: "AdminRouteID", Type: "types.NullableAdminRouteID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "admin_content_field_gen.go",
	},

	// AdminDatatypes
	{
		Name:               "AdminDatatypes",
		Singular:           "AdminDatatype",
		Plural:             "AdminDatatypes",
		SqlcTypeName:       "AdminDatatypes",
		TableName:          "admin_datatypes",
		IDType:             "types.AdminDatatypeID",
		IDField:            "AdminDatatypeID",
		NewIDFunc:          "types.NewAdminDatatypeID()",
		UpdateSuccessField: "s.Label",
		StringTypeName:     "StringAdminDatatypes",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "AdminDatatypeID", Type: "types.AdminDatatypeID", JSONTag: "admin_datatype_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "ParentID", Type: "types.NullableAdminDatatypeID", JSONTag: "parent_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Label", Type: "string", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Type", Type: "string", JSONTag: "type", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "AuthorID", Type: "types.UserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListAdminDatatypeChildrenPaginated",
				SqlcName:      "ListAdminDatatypeChildrenPaginated",
				AppParamsType: "ListAdminDatatypeChildrenPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "ParentID", ParamType: "types.AdminDatatypeID", SqlcField: "ParentID", WrapExpr: "types.NullableAdminDatatypeID{ID: params.%s, Valid: true}"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListAdminDatatypeChildrenPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "ParentID", Type: "types.AdminDatatypeID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "admin_datatype_gen.go",
	},

	// AdminFields
	{
		Name:               "AdminFields",
		Singular:           "AdminField",
		Plural:             "AdminFields",
		SqlcTypeName:       "AdminFields",
		TableName:          "admin_fields",
		IDType:             "types.AdminFieldID",
		IDField:            "AdminFieldID",
		NewIDFunc:          "types.NewAdminFieldID()",
		UpdateSuccessField: "s.Label",
		StringTypeName:     "StringAdminFields",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "AdminFieldID", Type: "types.AdminFieldID", JSONTag: "admin_field_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "ParentID", Type: "types.NullableAdminDatatypeID", JSONTag: "parent_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Label", Type: "string", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Data", Type: "string", JSONTag: "data", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Validation", Type: "string", JSONTag: "validation", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "UIConfig", SqlcName: "UiConfig", Type: "string", JSONTag: "ui_config", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Type", Type: "types.FieldType", JSONTag: "type", InCreate: true, InUpdate: true, StringConvert: "cast"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListAdminFieldsByParentIDPaginated",
				SqlcName:      "ListAdminFieldByParentIDPaginated",
				AppParamsType: "ListAdminFieldsByParentIDPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "ParentID", ParamType: "types.AdminDatatypeID", SqlcField: "ParentID", WrapExpr: "types.NullableAdminDatatypeID{ID: params.%s, Valid: true}"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListAdminFieldsByParentIDPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "ParentID", Type: "types.AdminDatatypeID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "admin_field_gen.go",
	},

	// Datatypes (CallerSuppliedID)
	{
		Name:               "Datatypes",
		Singular:           "Datatype",
		Plural:             "Datatypes",
		SqlcTypeName:       "Datatypes",
		TableName:          "datatypes",
		IDType:             "types.DatatypeID",
		IDField:            "DatatypeID",
		NewIDFunc:          "types.NewDatatypeID()",
		CallerSuppliedID:   true,
		UpdateSuccessField: "s.Label",
		StringTypeName:     "StringDatatypes",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "DatatypeID", Type: "types.DatatypeID", JSONTag: "datatype_id", IsPrimaryID: true, InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "ParentID", Type: "types.NullableDatatypeID", JSONTag: "parent_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Label", Type: "string", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Type", Type: "string", JSONTag: "type", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "AuthorID", Type: "types.UserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListDatatypeChildren",
				SqlcName:    "ListDatatypeChildren",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "parentID", ParamType: "types.DatatypeID", SqlcField: "ParentID", WrapExpr: "types.NullableDatatypeID{ID: %s, Valid: true}"},
				},
			},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListDatatypeChildrenPaginated",
				SqlcName:      "ListDatatypeChildrenPaginated",
				AppParamsType: "ListDatatypeChildrenPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "ParentID", ParamType: "types.DatatypeID", SqlcField: "ParentID", WrapExpr: "types.NullableDatatypeID{ID: params.%s, Valid: true}"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListDatatypeChildrenPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "ParentID", Type: "types.DatatypeID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "datatype_gen.go",
	},

	// Fields (CallerSuppliedID)
	{
		Name:               "Fields",
		Singular:           "Field",
		Plural:             "Fields",
		SqlcTypeName:       "Fields",
		TableName:          "fields",
		IDType:             "types.FieldID",
		IDField:            "FieldID",
		NewIDFunc:          "types.NewFieldID()",
		CallerSuppliedID:   true,
		UpdateSuccessField: "s.Label",
		StringTypeName:     "StringFields",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "FieldID", Type: "types.FieldID", JSONTag: "field_id", IsPrimaryID: true, InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "ParentID", Type: "types.NullableDatatypeID", JSONTag: "parent_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Label", Type: "string", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Data", Type: "string", JSONTag: "data", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Validation", Type: "string", JSONTag: "validation", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "UIConfig", SqlcName: "UiConfig", Type: "string", JSONTag: "ui_config", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Type", Type: "types.FieldType", JSONTag: "type", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListFieldsByDatatypeID",
				SqlcName:    "ListFieldByDatatypeID",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "datatypeID", ParamType: "types.NullableDatatypeID", SqlcField: "ParentID"},
				},
			},
		},
		OutputFile: "field_gen.go",
	},

	// DatatypeFields (CallerSuppliedID, no sqlc Get, special sqlc type name)
	{
		Name:                "DatatypeFields",
		Singular:            "DatatypeField",
		Plural:              "DatatypeFields",
		SqlcTypeName:        "DatatypesFields",
		TableName:           "datatypes_fields",
		IDType:              "string",
		IDField:             "ID",
		NewIDFunc:           "string(types.NewDatatypeFieldID())",
		CallerSuppliedID:    true,
		SkipMappers:         true,
		SkipAuditedCommands: true,
		SkipGet:             true,
		SqlcCreateTableName: "CreateDatatypesFieldsTable",
		UpdateSuccessField:  "s.ID",
		StringTypeName:      "StringDatatypeFields",
		Fields: []Field{
			{AppName: "ID", Type: "string", JSONTag: "id", IsPrimaryID: true, InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "DatatypeID", Type: "types.DatatypeID", JSONTag: "datatype_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "FieldID", Type: "types.FieldID", JSONTag: "field_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "SortOrder", Type: "int64", JSONTag: "sort_order", InCreate: true, InUpdate: true, StringConvert: "sprintf"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListDatatypeFieldByDatatypeID",
				SqlcName:    "ListDatatypeFieldByDatatypeID",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "datatypeID", ParamType: "types.DatatypeID", SqlcField: "DatatypeID"},
				},
			},
			{
				MethodName:  "ListDatatypeFieldByFieldID",
				SqlcName:    "ListDatatypeFieldByFieldID",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "fieldID", ParamType: "types.FieldID", SqlcField: "FieldID"},
				},
			},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListDatatypeFieldByDatatypeIDPaginated",
				SqlcName:      "ListDatatypeFieldByDatatypeIDPaginated",
				AppParamsType: "ListDatatypeFieldByDatatypeIDPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "DatatypeID", ParamType: "types.DatatypeID", SqlcField: "DatatypeID"},
				},
			},
			{
				MethodName:    "ListDatatypeFieldByFieldIDPaginated",
				SqlcName:      "ListDatatypeFieldByFieldIDPaginated",
				AppParamsType: "ListDatatypeFieldByFieldIDPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "FieldID", ParamType: "types.FieldID", SqlcField: "FieldID"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListDatatypeFieldByDatatypeIDPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "DatatypeID", Type: "types.DatatypeID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
			{
				TypeName: "ListDatatypeFieldByFieldIDPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "FieldID", Type: "types.FieldID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "datatype_field_gen.go",
	},

	// AdminDatatypeFields (no sqlc Get, special sqlc type name)
	{
		Name:                "AdminDatatypeFields",
		Singular:            "AdminDatatypeField",
		Plural:              "AdminDatatypeFields",
		SqlcTypeName:        "AdminDatatypesFields",
		TableName:           "admin_datatypes_fields",
		IDType:              "string",
		IDField:             "ID",
		NewIDFunc:           "string(types.NewAdminDatatypeFieldID())",
		SkipAuditedCommands: true,
		SkipGet:             true,
		SqlcCreateTableName: "CreateAdminDatatypesFieldsTable",
		UpdateSuccessField:  "s.ID",
		StringTypeName:      "StringAdminDatatypeFields",
		Fields: []Field{
			{AppName: "ID", Type: "string", JSONTag: "id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "string"},
			{AppName: "AdminDatatypeID", Type: "types.AdminDatatypeID", JSONTag: "admin_datatype_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "AdminFieldID", Type: "types.AdminFieldID", JSONTag: "admin_field_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "ListAdminDatatypeFieldByDatatypeID",
				SqlcName:    "ListAdminDatatypeFieldByDatatypeID",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "adminDatatypeID", ParamType: "types.AdminDatatypeID", SqlcField: "AdminDatatypeID"},
				},
			},
			{
				MethodName:  "ListAdminDatatypeFieldByFieldID",
				SqlcName:    "ListAdminDatatypeFieldByFieldID",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "adminFieldID", ParamType: "types.AdminFieldID", SqlcField: "AdminFieldID"},
				},
			},
		},
		PaginatedExtraQueries: []PaginatedExtraQuery{
			{
				MethodName:    "ListAdminDatatypeFieldByDatatypeIDPaginated",
				SqlcName:      "ListAdminDatatypeFieldByDatatypeIDPaginated",
				AppParamsType: "ListAdminDatatypeFieldByDatatypeIDPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "AdminDatatypeID", ParamType: "types.AdminDatatypeID", SqlcField: "AdminDatatypeID"},
				},
			},
			{
				MethodName:    "ListAdminDatatypeFieldByFieldIDPaginated",
				SqlcName:      "ListAdminDatatypeFieldByFieldIDPaginated",
				AppParamsType: "ListAdminDatatypeFieldByFieldIDPaginatedParams",
				FilterFields: []ExtraQueryParam{
					{ParamName: "AdminFieldID", ParamType: "types.AdminFieldID", SqlcField: "AdminFieldID"},
				},
			},
		},
		ExtraParamStructs: []ExtraParamStruct{
			{
				TypeName: "ListAdminDatatypeFieldByDatatypeIDPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "AdminDatatypeID", Type: "types.AdminDatatypeID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
			{
				TypeName: "ListAdminDatatypeFieldByFieldIDPaginatedParams",
				Fields: []ExtraParamField{
					{Name: "AdminFieldID", Type: "types.AdminFieldID"},
					{Name: "Limit", Type: "int64"},
					{Name: "Offset", Type: "int64"},
				},
			},
		},
		OutputFile: "admin_datatype_field_gen.go",
	},

	// ===================================================
	// SKIP MAPPERS + AUDITED COMMANDS (per-driver diffs)
	// ===================================================

	// Roles (SystemProtected: bool vs int64 per-driver)
	{
		Name:                "Roles",
		Singular:            "Role",
		Plural:              "Roles",
		SqlcTypeName:        "Roles",
		TableName:           "roles",
		IDType:              "types.RoleID",
		IDField:             "RoleID",
		NewIDFunc:           "types.NewRoleID()",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		UpdateSuccessField:  "s.Label",
		StringTypeName:      "StringRoles",
		Fields: []Field{
			{AppName: "RoleID", Type: "types.RoleID", JSONTag: "role_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "Label", Type: "string", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "SystemProtected", Type: "bool", JSONTag: "system_protected", InCreate: true, InUpdate: true, StringConvert: ""},
		},
		OutputFile: "role_gen.go",
	},

	// Permissions (SystemProtected: bool vs int64 per-driver)
	{
		Name:                "Permissions",
		Singular:            "Permission",
		Plural:              "Permissions",
		SqlcTypeName:        "Permissions",
		TableName:           "permissions",
		IDType:              "types.PermissionID",
		IDField:             "PermissionID",
		NewIDFunc:           "types.NewPermissionID()",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		UpdateSuccessField:  "s.Label",
		StringTypeName:      "StringPermissions",
		Fields: []Field{
			{AppName: "PermissionID", Type: "types.PermissionID", JSONTag: "permission_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "Label", Type: "string", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "SystemProtected", Type: "bool", JSONTag: "system_protected", InCreate: true, InUpdate: true, StringConvert: ""},
		},
		OutputFile: "permission_gen.go",
	},

	// Tokens (Token->Tokens field remap, IssuedAt per-driver time conversion)
	{
		Name:                "Tokens",
		Singular:            "Token",
		Plural:              "Tokens",
		SqlcTypeName:        "Tokens",
		TableName:           "tokens",
		IDType:              "string",
		IDField:             "ID",
		NewIDFunc:           "string(types.NewTokenID())",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		UpdateSuccessField:  "s.ID",
		StringTypeName:      "StringTokens",
		Fields: []Field{
			{AppName: "ID", Type: "string", JSONTag: "id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "string"},
			{AppName: "UserID", Type: "types.NullableUserID", JSONTag: "user_id", InCreate: true, InUpdate: false, StringConvert: "toString"},
			{AppName: "TokenType", Type: "string", JSONTag: "token_type", InCreate: true, InUpdate: false, StringConvert: "string"},
			{AppName: "Token", Type: "string", JSONTag: "token", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "IssuedAt", Type: "string", JSONTag: "issued_at", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "ExpiresAt", Type: "types.Timestamp", JSONTag: "expires_at", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Revoked", Type: "bool", JSONTag: "revoked", InCreate: true, InUpdate: true, StringConvert: "sprintfBool"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "GetTokenByTokenValue",
				SqlcName:    "GetTokenByTokenValue",
				ReturnsList: false,
				Params: []ExtraQueryParam{
					{ParamName: "tokenValue", ParamType: "string", SqlcField: "Tokens"},
				},
			},
			{
				MethodName:  "GetTokenByUserId",
				SqlcName:    "GetTokenByUserId",
				ReturnsList: true,
				Params: []ExtraQueryParam{
					{ParamName: "userID", ParamType: "types.NullableUserID", SqlcField: "UserID"},
				},
			},
		},
		OutputFile: "token_gen.go",
	},

	// Sessions (LastAccess per-driver time conversion)
	{
		Name:                "Sessions",
		Singular:            "Session",
		Plural:              "Sessions",
		SqlcTypeName:        "Sessions",
		TableName:           "sessions",
		IDType:              "types.SessionID",
		IDField:             "SessionID",
		NewIDFunc:           "types.NewSessionID()",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		UpdateSuccessField:  "s.SessionID",
		StringTypeName:      "StringSessions",
		Fields: []Field{
			{AppName: "SessionID", Type: "types.SessionID", JSONTag: "session_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "UserID", Type: "types.NullableUserID", JSONTag: "user_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "ExpiresAt", Type: "types.Timestamp", JSONTag: "expires_at", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "LastAccess", Type: "sql.NullString", JSONTag: "last_access", InCreate: true, InUpdate: true, StringConvert: "nullToEmpty"},
			{AppName: "IpAddress", Type: "sql.NullString", JSONTag: "ip_address", InCreate: true, InUpdate: true, StringConvert: "nullToEmpty"},
			{AppName: "UserAgent", Type: "sql.NullString", JSONTag: "user_agent", InCreate: true, InUpdate: true, StringConvert: "nullToEmpty"},
			{AppName: "SessionData", Type: "sql.NullString", JSONTag: "session_data", InCreate: true, InUpdate: true, StringConvert: "nullToEmpty"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "GetSessionByUserId",
				SqlcName:    "GetSessionByUserId",
				ReturnsList: false,
				Params: []ExtraQueryParam{
					{ParamName: "userID", ParamType: "types.NullableUserID", SqlcField: "UserID"},
				},
			},
		},
		OutputFile: "session_gen.go",
	},

	// MediaDimensions (Width/Height int64 vs int32 per-driver)
	{
		Name:                "MediaDimensions",
		Singular:            "MediaDimension",
		Plural:              "MediaDimensions",
		SqlcTypeName:        "MediaDimensions",
		TableName:           "media_dimensions",
		IDType:              "string",
		IDField:             "MdID",
		NewIDFunc:           "string(types.NewMediaDimensionID())",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		UpdateSuccessField:  "s.MdID",
		StringTypeName:      "StringMediaDimensions",
		Fields: []Field{
			{AppName: "MdID", Type: "string", JSONTag: "md_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "string"},
			{AppName: "Label", Type: "sql.NullString", JSONTag: "label", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Width", Type: "sql.NullInt64", JSONTag: "width", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "Height", Type: "sql.NullInt64", JSONTag: "height", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
			{AppName: "AspectRatio", Type: "sql.NullString", JSONTag: "aspect_ratio", InCreate: true, InUpdate: true, StringConvert: "nullToString"},
		},
		OutputFile: "media_dimension_gen.go",
	},

	// UserOauth (OAuthID casing per-driver, TokenExpiresAt per-driver)
	{
		Name:                "UserOauth",
		Singular:            "UserOauth",
		Plural:              "UserOauths",
		SqlcTypeName:        "UserOauth",
		TableName:           "user_oauth",
		IDType:              "types.UserOauthID",
		IDField:             "UserOauthID",
		NewIDFunc:           "types.NewUserOauthID()",
		SqlcCountName:       "CountUserOauths",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		SkipGet:             true,
		UpdateSuccessField:  "s.UserOauthID",
		StringTypeName:      "StringUserOauth",
		Fields: []Field{
			{AppName: "UserOauthID", Type: "types.UserOauthID", JSONTag: "user_oauth_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
			{AppName: "UserID", Type: "types.NullableUserID", JSONTag: "user_id", InCreate: true, InUpdate: false, StringConvert: "toString"},
			{AppName: "OauthProvider", Type: "string", JSONTag: "oauth_provider", InCreate: true, InUpdate: false, StringConvert: "string"},
			{AppName: "OauthProviderUserID", Type: "string", JSONTag: "oauth_provider_user_id", InCreate: true, InUpdate: false, StringConvert: "string"},
			{AppName: "AccessToken", Type: "string", JSONTag: "access_token", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "RefreshToken", Type: "string", JSONTag: "refresh_token", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "TokenExpiresAt", Type: "string", JSONTag: "token_expires_at", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: false, StringConvert: "toString"},
		},
		ExtraQueries: []ExtraQuery{
			{
				MethodName:  "GetUserOauthByUserId",
				SqlcName:    "GetUserOauthByUserId",
				ReturnsList: false,
				Params: []ExtraQueryParam{
					{ParamName: "userID", ParamType: "types.NullableUserID", SqlcField: "UserID"},
				},
			},
			{
				MethodName:  "GetUserOauthByProviderID",
				SqlcName:    "GetUserOauthByProviderID",
				ReturnsList: false,
				Params: []ExtraQueryParam{
					{ParamName: "provider", ParamType: "string", SqlcField: "OauthProvider"},
					{ParamName: "providerUserID", ParamType: "string", SqlcField: "OAuthProviderUserID"},
				},
			},
		},
		OutputFile: "user_oauth_gen.go",
	},

	// Routes (Slug_2 update WHERE, Status int conversion per-driver)
	{
		Name:                "Routes",
		Singular:            "Route",
		Plural:              "Routes",
		SqlcTypeName:        "Routes",
		TableName:           "routes",
		IDType:              "types.RouteID",
		IDField:             "RouteID",
		NewIDFunc:           "types.NewRouteID()",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		UpdateSuccessField:  "s.Slug",
		StringTypeName:      "StringRoutes",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "RouteID", Type: "types.RouteID", JSONTag: "route_id", IsPrimaryID: true, InCreate: false, InUpdate: false, StringConvert: "toString"},
			{AppName: "Slug", Type: "types.Slug", JSONTag: "slug", InCreate: true, InUpdate: true, StringConvert: "cast"},
			{AppName: "Title", Type: "string", JSONTag: "title", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Status", Type: "int64", JSONTag: "status", InCreate: true, InUpdate: true, StringConvert: "sprintf"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Slug_2", Type: "types.Slug", JSONTag: "slug_2", InCreate: false, InUpdate: true, UpdateParamsOnly: true},
		},
		OutputFile: "route_gen.go",
	},

	// AdminRoutes (Slug_2 update WHERE, Status int conversion, no sqlc Get)
	{
		Name:                "AdminRoutes",
		Singular:            "AdminRoute",
		Plural:              "AdminRoutes",
		SqlcTypeName:        "AdminRoutes",
		TableName:           "admin_routes",
		IDType:              "types.AdminRouteID",
		IDField:             "AdminRouteID",
		NewIDFunc:           "types.NewAdminRouteID()",
		SkipMappers:         true,
		SkipAuditedCommands: true,
		SkipGet:             true,
		SqlcCountName:       "CountAdminRoute",
		UpdateSuccessField:  "s.Slug",
		StringTypeName:      "StringAdminRoutes",
		ExtraStringFields: []ExtraStringField{
			{Name: "History", Value: `""`},
		},
		Fields: []Field{
			{AppName: "AdminRouteID", Type: "types.AdminRouteID", JSONTag: "admin_route_id", IsPrimaryID: true, InCreate: false, InUpdate: false, StringConvert: "toString"},
			{AppName: "Slug", Type: "types.Slug", JSONTag: "slug", InCreate: true, InUpdate: true, StringConvert: "cast"},
			{AppName: "Title", Type: "string", JSONTag: "title", InCreate: true, InUpdate: true, StringConvert: "string"},
			{AppName: "Status", Type: "int64", JSONTag: "status", InCreate: true, InUpdate: true, StringConvert: "sprintf"},
			{AppName: "AuthorID", Type: "types.NullableUserID", JSONTag: "author_id", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
			{AppName: "Slug_2", Type: "types.Slug", JSONTag: "slug_2", InCreate: false, InUpdate: true, UpdateParamsOnly: true},
		},
		OutputFile: "admin_route_gen.go",
	},
}
