package main

import (
	"context"
	"net/url"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerAdminSchemaTools(srv *server.MCPServer, client *modulacms.Client) {
	// Admin Datatypes
	srv.AddTool(mcp.NewTool("admin_list_datatypes", mcp.WithDescription("List all admin datatypes."), mcp.WithBoolean("full", mcp.Description("Include linked fields"))), handleAdminListDatatypes(client))
	srv.AddTool(mcp.NewTool("admin_get_datatype", mcp.WithDescription("Get a single admin datatype by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype ID (ULID)"))), handleAdminGetDatatype(client))
	srv.AddTool(mcp.NewTool("admin_create_datatype", mcp.WithDescription("Create a new admin datatype."), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("type", mcp.Required(), mcp.Description("Type")), mcp.WithString("parent_id", mcp.Description("Parent admin datatype ID")), mcp.WithString("author_id", mcp.Description("Author user ID"))), handleAdminCreateDatatype(client))
	srv.AddTool(mcp.NewTool("admin_update_datatype", mcp.WithDescription("Update an admin datatype."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype ID")), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("type", mcp.Required(), mcp.Description("Type")), mcp.WithString("parent_id", mcp.Description("Parent")), mcp.WithString("author_id", mcp.Description("Author"))), handleAdminUpdateDatatype(client))
	srv.AddTool(mcp.NewTool("admin_delete_datatype", mcp.WithDescription("Delete an admin datatype by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype ID"))), handleAdminDeleteDatatype(client))

	// Admin Datatype Fields
	srv.AddTool(mcp.NewTool("admin_list_datatype_fields", mcp.WithDescription("List admin datatype-field links."), mcp.WithString("admin_datatype_id", mcp.Description("Filter by admin datatype ID")), mcp.WithString("admin_field_id", mcp.Description("Filter by admin field ID"))), handleAdminListDatatypeFields(client))
	srv.AddTool(mcp.NewTool("admin_get_datatype_field", mcp.WithDescription("Get a single admin datatype-field link."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype-field ID"))), handleAdminGetDatatypeField(client))
	srv.AddTool(mcp.NewTool("admin_create_datatype_field", mcp.WithDescription("Link an admin field to an admin datatype."), mcp.WithString("admin_datatype_id", mcp.Required(), mcp.Description("Admin datatype ID")), mcp.WithString("admin_field_id", mcp.Required(), mcp.Description("Admin field ID"))), handleAdminCreateDatatypeField(client))
	srv.AddTool(mcp.NewTool("admin_update_datatype_field", mcp.WithDescription("Update an admin datatype-field link."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype-field ID")), mcp.WithString("admin_datatype_id", mcp.Required(), mcp.Description("Admin datatype ID")), mcp.WithString("admin_field_id", mcp.Required(), mcp.Description("Admin field ID"))), handleAdminUpdateDatatypeField(client))
	srv.AddTool(mcp.NewTool("admin_delete_datatype_field", mcp.WithDescription("Delete an admin datatype-field link."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin datatype-field ID"))), handleAdminDeleteDatatypeField(client))

	// Admin Fields
	srv.AddTool(mcp.NewTool("admin_list_fields", mcp.WithDescription("List all admin field definitions.")), handleAdminListFields(client))
	srv.AddTool(mcp.NewTool("admin_get_field", mcp.WithDescription("Get a single admin field by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field ID (ULID)"))), handleAdminGetField(client))
	srv.AddTool(mcp.NewTool("admin_create_field", mcp.WithDescription("Create a new admin field."), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("field_type", mcp.Required(), mcp.Description("Field type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "relation", "json", "richtext", "slug", "email", "url")), mcp.WithString("parent_id", mcp.Description("Parent admin datatype ID")), mcp.WithString("data", mcp.Description("Additional data (JSON)")), mcp.WithString("validation", mcp.Description("Validation rules (JSON)")), mcp.WithString("ui_config", mcp.Description("UI config (JSON)")), mcp.WithString("author_id", mcp.Description("Author user ID"))), handleAdminCreateField(client))
	srv.AddTool(mcp.NewTool("admin_update_field", mcp.WithDescription("Update an admin field."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field ID")), mcp.WithString("name", mcp.Description("Machine-readable name")), mcp.WithString("label", mcp.Required(), mcp.Description("Label")), mcp.WithString("field_type", mcp.Required(), mcp.Description("Field type"), mcp.Enum("text", "textarea", "number", "date", "datetime", "boolean", "select", "media", "relation", "json", "richtext", "slug", "email", "url")), mcp.WithString("parent_id", mcp.Description("Parent")), mcp.WithString("data", mcp.Description("Data")), mcp.WithString("validation", mcp.Description("Validation")), mcp.WithString("ui_config", mcp.Description("UI config")), mcp.WithString("author_id", mcp.Description("Author"))), handleAdminUpdateField(client))
	srv.AddTool(mcp.NewTool("admin_delete_field", mcp.WithDescription("Delete an admin field by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field ID"))), handleAdminDeleteField(client))
}

// --- Admin Datatype Handlers ---

func handleAdminListDatatypes(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if req.GetBool("full", false) {
			params := url.Values{}
			params.Set("full", "true")
			result, err := client.AdminDatatypes.RawList(ctx, params)
			if err != nil {
				return errResult(err), nil
			}
			return mcp.NewToolResultText(string(result)), nil
		}
		result, err := client.AdminDatatypes.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminGetDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.AdminDatatypes.Get(ctx, modulacms.AdminDatatypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminCreateDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		params := modulacms.CreateAdminDatatypeParams{
			Name:     req.GetString("name", ""),
			Label:    label,
			Type:     typ,
			ParentID: optionalIDPtr[modulacms.AdminDatatypeID](req, "parent_id"),
			AuthorID: optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.AdminDatatypes.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminUpdateDatatype(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.UpdateAdminDatatypeParams{
			AdminDatatypeID: modulacms.AdminDatatypeID(id),
			Name:            req.GetString("name", ""),
			Label:           label,
			Type:            typ,
			ParentID:        optionalIDPtr[modulacms.AdminDatatypeID](req, "parent_id"),
			AuthorID:        optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.AdminDatatypes.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminDeleteDatatype(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.AdminDatatypes.Delete(ctx, modulacms.AdminDatatypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Admin Datatype Field Handlers ---

func handleAdminListDatatypeFields(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := url.Values{}
		if v := req.GetString("admin_datatype_id", ""); v != "" {
			params.Set("admin_datatype_id", v)
		}
		if v := req.GetString("admin_field_id", ""); v != "" {
			params.Set("admin_field_id", v)
		}
		result, err := client.AdminDatatypeFields.RawList(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleAdminGetDatatypeField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.AdminDatatypeFields.Get(ctx, modulacms.AdminDatatypeFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminCreateDatatypeField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dtID, err := req.RequireString("admin_datatype_id")
		if err != nil {
			return mcp.NewToolResultError("admin_datatype_id is required"), nil
		}
		fID, err := req.RequireString("admin_field_id")
		if err != nil {
			return mcp.NewToolResultError("admin_field_id is required"), nil
		}
		params := modulacms.CreateAdminDatatypeFieldParams{
			AdminDatatypeID: modulacms.AdminDatatypeID(dtID),
			AdminFieldID:    modulacms.AdminFieldID(fID),
		}
		result, err := client.AdminDatatypeFields.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminUpdateDatatypeField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		dtID, err := req.RequireString("admin_datatype_id")
		if err != nil {
			return mcp.NewToolResultError("admin_datatype_id is required"), nil
		}
		fID, err := req.RequireString("admin_field_id")
		if err != nil {
			return mcp.NewToolResultError("admin_field_id is required"), nil
		}
		params := modulacms.UpdateAdminDatatypeFieldParams{
			ID:              modulacms.AdminDatatypeFieldID(id),
			AdminDatatypeID: modulacms.AdminDatatypeID(dtID),
			AdminFieldID:    modulacms.AdminFieldID(fID),
		}
		result, err := client.AdminDatatypeFields.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminDeleteDatatypeField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.AdminDatatypeFields.Delete(ctx, modulacms.AdminDatatypeFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Admin Field Handlers ---

func handleAdminListFields(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.AdminFields.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminGetField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.AdminFields.Get(ctx, modulacms.AdminFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminCreateField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		ft, err := req.RequireString("field_type")
		if err != nil {
			return mcp.NewToolResultError("field_type is required"), nil
		}
		params := modulacms.CreateAdminFieldParams{
			Name:       req.GetString("name", ""),
			Label:      label,
			Type:       modulacms.FieldType(ft),
			ParentID:   optionalIDPtr[modulacms.AdminDatatypeID](req, "parent_id"),
			Data:       req.GetString("data", ""),
			Validation: req.GetString("validation", ""),
			UIConfig:   req.GetString("ui_config", ""),
			AuthorID:   optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.AdminFields.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminUpdateField(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.UpdateAdminFieldParams{
			AdminFieldID: modulacms.AdminFieldID(id),
			Name:         req.GetString("name", ""),
			Label:        label,
			Type:         modulacms.FieldType(ft),
			ParentID:     optionalIDPtr[modulacms.AdminDatatypeID](req, "parent_id"),
			Data:         req.GetString("data", ""),
			Validation:   req.GetString("validation", ""),
			UIConfig:     req.GetString("ui_config", ""),
			AuthorID:     optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.AdminFields.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminDeleteField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.AdminFields.Delete(ctx, modulacms.AdminFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
