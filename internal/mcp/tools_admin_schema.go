package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAdminSchemaTools(srv *server.MCPServer, backend AdminSchemaBackend) {
	// Admin Datatypes
	srv.AddTool(mcp.NewTool("admin_list_datatypes", mcp.WithDescription("List all admin datatypes."), mcp.WithBoolean("full", mcp.Description("Include linked fields"))), handleAdminListDatatypes(backend))
	srv.AddTool(mcp.NewTool("admin_get_datatype", mcp.WithDescription("Get a single admin datatype by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype ID (ULID)"))), handleAdminGetDatatype(backend))
	srv.AddTool(mcp.NewTool("admin_create_datatype", mcp.WithDescription("Create a new admin datatype."), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("type", mcp.Required(), mcp.Description("Type")), mcp.WithString("parent_id", mcp.Description("Parent admin datatype ID")), mcp.WithString("author_id", mcp.Description("Author user ID"))), handleAdminCreateDatatype(backend))
	srv.AddTool(mcp.NewTool("admin_update_datatype", mcp.WithDescription("Update an admin datatype."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype ID")), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("type", mcp.Required(), mcp.Description("Type")), mcp.WithString("parent_id", mcp.Description("Parent")), mcp.WithString("author_id", mcp.Description("Author"))), handleAdminUpdateDatatype(backend))
	srv.AddTool(mcp.NewTool("admin_delete_datatype", mcp.WithDescription("Delete an admin datatype by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype ID"))), handleAdminDeleteDatatype(backend))

	// Admin Datatype Sort Ordering
	srv.AddTool(mcp.NewTool("admin_get_datatype_max_sort_order", mcp.WithDescription("Get the maximum sort order value across admin datatypes.")), handleAdminGetDatatypeMaxSortOrder(backend))
	srv.AddTool(mcp.NewTool("admin_update_datatype_sort_order", mcp.WithDescription("Update an admin datatype's sort order."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype ID (ULID)")), mcp.WithNumber("sort_order", mcp.Required(), mcp.Description("New sort order value"))), handleAdminUpdateDatatypeSortOrder(backend))

	// Admin Fields
	srv.AddTool(mcp.NewTool("admin_list_fields", mcp.WithDescription("List all admin field definitions.")), handleAdminListFields(backend))
	srv.AddTool(mcp.NewTool("admin_get_field", mcp.WithDescription("Get a single admin field by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field ID (ULID)"))), handleAdminGetField(backend))
	srv.AddTool(mcp.NewTool("admin_create_field", mcp.WithDescription("Create a new admin field."), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("field_type", mcp.Required(), mcp.Description("Field type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "relation", "json", "richtext", "slug", "email", "url")), mcp.WithString("parent_id", mcp.Description("Parent admin datatype ID")), mcp.WithString("data", mcp.Description("Additional data (JSON)")), mcp.WithString("validation", mcp.Description("Validation rules (JSON)")), mcp.WithString("ui_config", mcp.Description("UI config (JSON)")), mcp.WithString("author_id", mcp.Description("Author user ID"))), handleAdminCreateField(backend))
	srv.AddTool(mcp.NewTool("admin_update_field", mcp.WithDescription("Update an admin field."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field ID")), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("field_type", mcp.Required(), mcp.Description("Field type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "relation", "json", "richtext", "slug", "email", "url")), mcp.WithString("parent_id", mcp.Description("Parent")), mcp.WithString("data", mcp.Description("Data")), mcp.WithString("validation", mcp.Description("Validation")), mcp.WithString("ui_config", mcp.Description("UI config")), mcp.WithString("author_id", mcp.Description("Author"))), handleAdminUpdateField(backend))
	srv.AddTool(mcp.NewTool("admin_delete_field", mcp.WithDescription("Delete an admin field by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field ID"))), handleAdminDeleteField(backend))
}

// --- Admin Datatype Handlers ---

func handleAdminListDatatypes(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		full := req.GetBool("full", false)
		data, err := backend.ListAdminDatatypes(ctx, full)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetDatatype(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetAdminDatatype(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateDatatype(backend AdminSchemaBackend) server.ToolHandlerFunc {
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
		data, err := backend.CreateAdminDatatype(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateDatatype(backend AdminSchemaBackend) server.ToolHandlerFunc {
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
			"admin_datatype_id": id,
			"name":              optionalStrPtr(req, "name"),
			"label":             label,
			"type":              typ,
			"parent_id":         optionalStrPtr(req, "parent_id"),
			"author_id":         optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminDatatype(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteDatatype(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteAdminDatatype(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Admin Datatype Sort Ordering Handlers ---

func handleAdminGetDatatypeMaxSortOrder(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.AdminGetDatatypeMaxSortOrder(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateDatatypeSortOrder(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		sortOrder := int64(req.GetFloat("sort_order", 0))
		if err := backend.AdminUpdateDatatypeSortOrder(ctx, id, sortOrder); err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("updated"), nil
	}
}

// --- Admin Field Handlers ---

func handleAdminListFields(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListAdminFields(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetField(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetAdminField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateField(backend AdminSchemaBackend) server.ToolHandlerFunc {
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
		data, err := backend.CreateAdminField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateField(backend AdminSchemaBackend) server.ToolHandlerFunc {
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
			"admin_field_id": id,
			"name":           optionalStrPtr(req, "name"),
			"label":          label,
			"type":           ft,
			"parent_id":      optionalStrPtr(req, "parent_id"),
			"data":           optionalStrPtr(req, "data"),
			"validation":     optionalStrPtr(req, "validation"),
			"ui_config":      optionalStrPtr(req, "ui_config"),
			"author_id":      optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteField(backend AdminSchemaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteAdminField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
