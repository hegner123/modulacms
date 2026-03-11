package mcp

import (
	"context"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modula "github.com/hegner123/modulacms/sdks/go"
)

func registerSchemaTools(srv *server.MCPServer, client *modula.Client) {
	// --- Datatypes ---

	srv.AddTool(
		mcp.NewTool("list_datatypes",
			mcp.WithDescription("List all datatypes. Set full=true to return datatypes with their linked fields joined."),
			mcp.WithBoolean("full", mcp.Description("If true, returns datatypes with their field associations included")),
		),
		handleListDatatypes(client),
	)

	srv.AddTool(
		mcp.NewTool("get_datatype",
			mcp.WithDescription("Get a single datatype by ID. Returns the datatype without its field associations. To see which fields are linked to a datatype, use list_datatype_fields with a datatype_id filter."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Datatype ID (ULID)")),
		),
		handleGetDatatype(client),
	)

	srv.AddTool(
		mcp.NewTool("create_datatype",
			mcp.WithDescription("Create a new datatype. The type field is freeform (e.g. 'page', 'component', 'block')."),
			mcp.WithString("name", mcp.Description("Machine-readable name (used as JSON key). If omitted, derived from label.")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Datatype label")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Datatype type (freeform, e.g. 'page', 'component', 'block')")),
			mcp.WithString("parent_id", mcp.Description("Parent datatype ID for hierarchical datatypes")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleCreateDatatype(client),
	)

	srv.AddTool(
		mcp.NewTool("update_datatype",
			mcp.WithDescription("Update an existing datatype. This is a full replacement."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Datatype ID (ULID)")),
			mcp.WithString("name", mcp.Description("Machine-readable name (used as JSON key). If omitted, derived from label.")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Datatype label")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Datatype type (freeform)")),
			mcp.WithString("parent_id", mcp.Description("Parent datatype ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateDatatype(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_datatype",
			mcp.WithDescription("Delete a datatype by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Datatype ID (ULID)")),
		),
		handleDeleteDatatype(client),
	)

	// --- Fields ---

	srv.AddTool(
		mcp.NewTool("list_fields",
			mcp.WithDescription("List all field definitions."),
		),
		handleListFields(client),
	)

	srv.AddTool(
		mcp.NewTool("get_field",
			mcp.WithDescription("Get a single field definition by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field ID (ULID)")),
		),
		handleGetField(client),
	)

	srv.AddTool(
		mcp.NewTool("create_field",
			mcp.WithDescription("Create a new field definition. The field_type parameter specifies the data type."),
			mcp.WithString("name", mcp.Description("Machine-readable name (used as JSON key). If omitted, derived from label.")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Field label")),
			mcp.WithString("field_type", mcp.Required(), mcp.Description("Field data type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "_id", "json", "richtext", "slug", "email", "url")),
			mcp.WithString("parent_id", mcp.Description("Parent datatype ID")),
			mcp.WithString("data", mcp.Description("Additional field data (JSON string)")),
			mcp.WithString("validation", mcp.Description("Validation rules (JSON string)")),
			mcp.WithString("ui_config", mcp.Description("UI configuration (JSON string)")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleCreateField(client),
	)

	srv.AddTool(
		mcp.NewTool("update_field",
			mcp.WithDescription("Update an existing field definition. This is a full replacement."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field ID (ULID)")),
			mcp.WithString("name", mcp.Description("Machine-readable name (used as JSON key). If omitted, derived from label.")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Field label")),
			mcp.WithString("field_type", mcp.Required(), mcp.Description("Field data type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "_id", "json", "richtext", "slug", "email", "url")),
			mcp.WithString("parent_id", mcp.Description("Parent datatype ID")),
			mcp.WithString("data", mcp.Description("Additional field data (JSON string)")),
			mcp.WithString("validation", mcp.Description("Validation rules (JSON string)")),
			mcp.WithString("ui_config", mcp.Description("UI configuration (JSON string)")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateField(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_field",
			mcp.WithDescription("Delete a field definition by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field ID (ULID)")),
		),
		handleDeleteField(client),
	)

	// --- Datatype Full ---

	srv.AddTool(
		mcp.NewTool("get_datatype_full",
			mcp.WithDescription("Get a single datatype with its linked fields joined. If id is omitted, returns all datatypes with fields."),
			mcp.WithString("id", mcp.Description("Datatype ID (ULID). Omit to list all with fields.")),
		),
		handleGetDatatypeFull(client),
	)

	// --- Field Types ---

	srv.AddTool(
		mcp.NewTool("list_field_types",
			mcp.WithDescription("List all field type definitions."),
		),
		handleListFieldTypes(client),
	)

	srv.AddTool(
		mcp.NewTool("get_field_type",
			mcp.WithDescription("Get a single field type by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field type ID (ULID)")),
		),
		handleGetFieldType(client),
	)

	srv.AddTool(
		mcp.NewTool("create_field_type",
			mcp.WithDescription("Create a new field type definition."),
			mcp.WithString("type", mcp.Required(), mcp.Description("Field type key (e.g. 'color', 'rating')")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Human-readable label")),
		),
		handleCreateFieldType(client),
	)

	srv.AddTool(
		mcp.NewTool("update_field_type",
			mcp.WithDescription("Update a field type definition."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field type ID (ULID)")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Field type key")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Human-readable label")),
		),
		handleUpdateFieldType(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_field_type",
			mcp.WithDescription("Delete a field type by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field type ID (ULID)")),
		),
		handleDeleteFieldType(client),
	)
}

// --- Datatype Handlers ---

func handleListDatatypes(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		full := req.GetBool("full", false)
		if full {
			params := url.Values{}
			params.Set("full", "true")
			result, err := client.Datatypes.RawList(ctx, params)
			if err != nil {
				return errResult(err), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		}
		result, err := client.Datatypes.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetDatatype(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Datatypes.Get(ctx, modula.DatatypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateDatatype(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		params := modula.CreateDatatypeParams{
			Name:     req.GetString("name", ""),
			Label:    label,
			Type:     typ,
			ParentID: optionalIDPtr[modula.DatatypeID](req, "parent_id"),
			AuthorID: optionalIDPtr[modula.UserID](req, "author_id"),
		}
		result, err := client.Datatypes.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateDatatype(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		params := modula.UpdateDatatypeParams{
			DatatypeID: modula.DatatypeID(id),
			Name:       req.GetString("name", ""),
			Label:      label,
			Type:       typ,
			ParentID:   optionalIDPtr[modula.DatatypeID](req, "parent_id"),
			AuthorID:   optionalIDPtr[modula.UserID](req, "author_id"),
		}
		result, err := client.Datatypes.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteDatatype(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Datatypes.Delete(ctx, modula.DatatypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Field Handlers ---

func handleListFields(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Fields.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetField(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Fields.Get(ctx, modula.FieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateField(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		ft, err := req.RequireString("field_type")
		if err != nil {
			return mcp.NewToolResultError("field_type is required"), nil
		}
		params := modula.CreateFieldParams{
			Name:       req.GetString("name", ""),
			Label:      label,
			Type:       modula.FieldType(ft),
			ParentID:   optionalIDPtr[modula.DatatypeID](req, "parent_id"),
			Data:       req.GetString("data", ""),
			Validation: req.GetString("validation", ""),
			UIConfig:   req.GetString("ui_config", ""),
			AuthorID:   optionalIDPtr[modula.UserID](req, "author_id"),
		}
		result, err := client.Fields.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateField(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		ft, err := req.RequireString("field_type")
		if err != nil {
			return mcp.NewToolResultError("field_type is required"), nil
		}
		params := modula.UpdateFieldParams{
			FieldID:    modula.FieldID(id),
			Name:       req.GetString("name", ""),
			Label:      label,
			Type:       modula.FieldType(ft),
			ParentID:   optionalIDPtr[modula.DatatypeID](req, "parent_id"),
			Data:       req.GetString("data", ""),
			Validation: req.GetString("validation", ""),
			UIConfig:   req.GetString("ui_config", ""),
			AuthorID:   optionalIDPtr[modula.UserID](req, "author_id"),
		}
		result, err := client.Fields.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteField(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Fields.Delete(ctx, modula.FieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Datatype Full Handler ---

func handleGetDatatypeFull(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := url.Values{}
		params.Set("full", "true")
		id := req.GetString("id", "")
		if id != "" {
			params.Set("q", id)
		}
		result, err := client.Datatypes.RawList(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

// --- Field Type Handlers ---

func handleListFieldTypes(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.FieldTypes.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetFieldType(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.FieldTypes.Get(ctx, modula.FieldTypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateFieldType(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modula.CreateFieldTypeParams{Type: typ, Label: label}
		result, err := client.FieldTypes.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateFieldType(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modula.UpdateFieldTypeParams{
			FieldTypeID: modula.FieldTypeID(id),
			Type:        typ,
			Label:       label,
		}
		result, err := client.FieldTypes.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteFieldType(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.FieldTypes.Delete(ctx, modula.FieldTypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
