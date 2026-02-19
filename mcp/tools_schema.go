package main

import (
	"context"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerSchemaTools(srv *server.MCPServer, client *modulacms.Client) {
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
			mcp.WithString("label", mcp.Required(), mcp.Description("Field label")),
			mcp.WithString("field_type", mcp.Required(), mcp.Description("Field data type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "relation", "json", "richtext", "slug", "email", "url")),
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
			mcp.WithString("label", mcp.Required(), mcp.Description("Field label")),
			mcp.WithString("field_type", mcp.Required(), mcp.Description("Field data type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "relation", "json", "richtext", "slug", "email", "url")),
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

	// --- Datatype-Field Links ---

	srv.AddTool(
		mcp.NewTool("list_datatype_fields",
			mcp.WithDescription("List datatype-field link records. Optionally filter by datatype_id and/or field_id to see which fields are linked to a specific datatype or vice versa."),
			mcp.WithString("datatype_id", mcp.Description("Filter by datatype ID")),
			mcp.WithString("field_id", mcp.Description("Filter by field ID")),
		),
		handleListDatatypeFields(client),
	)

	srv.AddTool(
		mcp.NewTool("link_field_to_datatype",
			mcp.WithDescription("Link a field definition to a datatype. This creates the association that allows content of this datatype to have values for this field."),
			mcp.WithString("datatype_id", mcp.Required(), mcp.Description("Datatype ID")),
			mcp.WithString("field_id", mcp.Required(), mcp.Description("Field ID")),
			mcp.WithNumber("sort_order", mcp.Required(), mcp.Description("Display order of this field within the datatype")),
		),
		handleLinkFieldToDatatype(client),
	)

	srv.AddTool(
		mcp.NewTool("unlink_field_from_datatype",
			mcp.WithDescription("Remove a field-to-datatype link. Requires the datatype-field link ID (not the datatype or field ID). Use list_datatype_fields with a datatype_id filter to find the link ID first."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Datatype-field link ID (ULID)")),
		),
		handleUnlinkFieldFromDatatype(client),
	)
}

// --- Datatype Handlers ---

func handleListDatatypes(client *modulacms.Client) server.ToolHandlerFunc {
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

func handleGetDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Datatypes.Get(ctx, modulacms.DatatypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		params := modulacms.CreateDatatypeParams{
			Label:    label,
			Type:     typ,
			ParentID: optionalIDPtr[modulacms.DatatypeID](req, "parent_id"),
			AuthorID: optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.Datatypes.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateDatatype(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.UpdateDatatypeParams{
			DatatypeID: modulacms.DatatypeID(id),
			Label:      label,
			Type:       typ,
			ParentID:   optionalIDPtr[modulacms.DatatypeID](req, "parent_id"),
			AuthorID:   optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.Datatypes.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Datatypes.Delete(ctx, modulacms.DatatypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Field Handlers ---

func handleListFields(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Fields.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Fields.Get(ctx, modulacms.FieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		ft, err := req.RequireString("field_type")
		if err != nil {
			return mcp.NewToolResultError("field_type is required"), nil
		}
		params := modulacms.CreateFieldParams{
			Label:      label,
			Type:       modulacms.FieldType(ft),
			ParentID:   optionalIDPtr[modulacms.DatatypeID](req, "parent_id"),
			Data:       req.GetString("data", ""),
			Validation: req.GetString("validation", ""),
			UIConfig:   req.GetString("ui_config", ""),
			AuthorID:   optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.Fields.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateField(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.UpdateFieldParams{
			FieldID:    modulacms.FieldID(id),
			Label:      label,
			Type:       modulacms.FieldType(ft),
			ParentID:   optionalIDPtr[modulacms.DatatypeID](req, "parent_id"),
			Data:       req.GetString("data", ""),
			Validation: req.GetString("validation", ""),
			UIConfig:   req.GetString("ui_config", ""),
			AuthorID:   optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.Fields.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Fields.Delete(ctx, modulacms.FieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Datatype-Field Link Handlers ---

func handleListDatatypeFields(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := url.Values{}
		if dtID := req.GetString("datatype_id", ""); dtID != "" {
			params.Set("datatype_id", dtID)
		}
		if fID := req.GetString("field_id", ""); fID != "" {
			params.Set("field_id", fID)
		}
		result, err := client.DatatypeFields.RawList(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleLinkFieldToDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dtID, err := req.RequireString("datatype_id")
		if err != nil {
			return mcp.NewToolResultError("datatype_id is required"), nil
		}
		fID, err := req.RequireString("field_id")
		if err != nil {
			return mcp.NewToolResultError("field_id is required"), nil
		}
		sortOrder := int64(req.GetFloat("sort_order", 0))
		params := modulacms.CreateDatatypeFieldParams{
			DatatypeID: modulacms.DatatypeID(dtID),
			FieldID:    modulacms.FieldID(fID),
			SortOrder:  sortOrder,
		}
		result, err := client.DatatypeFields.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUnlinkFieldFromDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.DatatypeFields.Delete(ctx, modulacms.DatatypeFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
