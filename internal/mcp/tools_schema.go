package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerSchemaTools(srv *server.MCPServer, backend SchemaBackend) {
	// --- Datatypes ---

	srv.AddTool(
		mcp.NewTool("list_datatypes",
			mcp.WithDescription("List all datatypes. Set full=true to return datatypes with their linked fields joined."),
			mcp.WithBoolean("full", mcp.Description("If true, returns datatypes with their field associations included")),
		),
		handleListDatatypes(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_datatype",
			mcp.WithDescription("Get a single datatype by ID. Returns the datatype without its field associations. To see which fields are linked to a datatype, use list_datatype_fields with a datatype_id filter."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Datatype ID (ULID)")),
		),
		handleGetDatatype(backend),
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
		handleCreateDatatype(backend),
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
		handleUpdateDatatype(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_datatype",
			mcp.WithDescription("Delete a datatype by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Datatype ID (ULID)")),
		),
		handleDeleteDatatype(backend),
	)

	// --- Fields ---

	srv.AddTool(
		mcp.NewTool("list_fields",
			mcp.WithDescription("List all field definitions."),
		),
		handleListFields(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_field",
			mcp.WithDescription("Get a single field definition by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field ID (ULID)")),
		),
		handleGetField(backend),
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
		handleCreateField(backend),
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
		handleUpdateField(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_field",
			mcp.WithDescription("Delete a field definition by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field ID (ULID)")),
		),
		handleDeleteField(backend),
	)

	// --- Datatype Full ---

	srv.AddTool(
		mcp.NewTool("get_datatype_full",
			mcp.WithDescription("Get a single datatype with its linked fields joined. If id is omitted, returns all datatypes with fields."),
			mcp.WithString("id", mcp.Description("Datatype ID (ULID). Omit to list all with fields.")),
		),
		handleGetDatatypeFull(backend),
	)

	// --- Field Types ---

	srv.AddTool(
		mcp.NewTool("list_field_types",
			mcp.WithDescription("List all field type definitions."),
		),
		handleListFieldTypes(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_field_type",
			mcp.WithDescription("Get a single field type by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field type ID (ULID)")),
		),
		handleGetFieldType(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_field_type",
			mcp.WithDescription("Create a new field type definition."),
			mcp.WithString("type", mcp.Required(), mcp.Description("Field type key (e.g. 'color', 'rating')")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Human-readable label")),
		),
		handleCreateFieldType(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_field_type",
			mcp.WithDescription("Update a field type definition."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field type ID (ULID)")),
			mcp.WithString("type", mcp.Required(), mcp.Description("Field type key")),
			mcp.WithString("label", mcp.Required(), mcp.Description("Human-readable label")),
		),
		handleUpdateFieldType(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_field_type",
			mcp.WithDescription("Delete a field type by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Field type ID (ULID)")),
		),
		handleDeleteFieldType(backend),
	)
}

// --- Datatype Handlers ---

func handleListDatatypes(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		full := req.GetBool("full", false)
		data, err := backend.ListDatatypes(ctx, full)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetDatatype(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetDatatype(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateDatatype(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"name":      req.GetString("name", ""),
			"label":     label,
			"type":      typ,
			"parent_id": optionalStrPtr(req, "parent_id"),
			"author_id": optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateDatatype(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateDatatype(backend SchemaBackend) server.ToolHandlerFunc {
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
		params, err := marshalParams(map[string]any{
			"datatype_id": id,
			"name":        req.GetString("name", ""),
			"label":       label,
			"type":        typ,
			"parent_id":   optionalStrPtr(req, "parent_id"),
			"author_id":   optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateDatatype(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteDatatype(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteDatatype(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Field Handlers ---

func handleListFields(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListFields(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetField(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateField(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		ft, err := req.RequireString("field_type")
		if err != nil {
			return mcp.NewToolResultError("field_type is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"name":       req.GetString("name", ""),
			"label":      label,
			"type":       ft,
			"parent_id":  optionalStrPtr(req, "parent_id"),
			"data":       req.GetString("data", ""),
			"validation": req.GetString("validation", ""),
			"ui_config":  req.GetString("ui_config", ""),
			"author_id":  optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateField(backend SchemaBackend) server.ToolHandlerFunc {
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
		params, err := marshalParams(map[string]any{
			"field_id":   id,
			"name":       req.GetString("name", ""),
			"label":      label,
			"type":       ft,
			"parent_id":  optionalStrPtr(req, "parent_id"),
			"data":       req.GetString("data", ""),
			"validation": req.GetString("validation", ""),
			"ui_config":  req.GetString("ui_config", ""),
			"author_id":  optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteField(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Datatype Full Handler ---

func handleGetDatatypeFull(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetString("id", "")
		data, err := backend.GetDatatypeFull(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Field Type Handlers ---

func handleListFieldTypes(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListFieldTypes(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetFieldType(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetFieldType(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateFieldType(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"type":  typ,
			"label": label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateFieldType(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateFieldType(backend SchemaBackend) server.ToolHandlerFunc {
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
		params, err := marshalParams(map[string]any{
			"field_type_id": id,
			"type":          typ,
			"label":         label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateFieldType(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteFieldType(backend SchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteFieldType(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
