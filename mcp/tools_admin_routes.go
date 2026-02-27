package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerAdminRouteTools(srv *server.MCPServer, client *modulacms.Client) {
	// Admin Routes
	srv.AddTool(mcp.NewTool("admin_list_routes", mcp.WithDescription("List all admin routes.")), handleAdminListRoutes(client))
	srv.AddTool(mcp.NewTool("admin_get_route", mcp.WithDescription("Get a single admin route by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin route ID (ULID)"))), handleAdminGetRoute(client))
	srv.AddTool(mcp.NewTool("admin_create_route", mcp.WithDescription("Create a new admin route."), mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug")), mcp.WithString("title", mcp.Required(), mcp.Description("Title")), mcp.WithNumber("status", mcp.Required(), mcp.Description("Status")), mcp.WithString("author_id", mcp.Description("Author user ID"))), handleAdminCreateRoute(client))
	srv.AddTool(mcp.NewTool("admin_update_route", mcp.WithDescription("Update an admin route."), mcp.WithString("slug", mcp.Required(), mcp.Description("Current slug")), mcp.WithString("title", mcp.Required(), mcp.Description("Title")), mcp.WithNumber("status", mcp.Required(), mcp.Description("Status")), mcp.WithString("new_slug", mcp.Description("New slug if renaming")), mcp.WithString("author_id", mcp.Description("Author"))), handleAdminUpdateRoute(client))
	srv.AddTool(mcp.NewTool("admin_delete_route", mcp.WithDescription("Delete an admin route by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin route ID"))), handleAdminDeleteRoute(client))

	// Admin Field Types
	srv.AddTool(mcp.NewTool("admin_list_field_types", mcp.WithDescription("List all admin field types.")), handleAdminListFieldTypes(client))
	srv.AddTool(mcp.NewTool("admin_get_field_type", mcp.WithDescription("Get a single admin field type by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field type ID (ULID)"))), handleAdminGetFieldType(client))
	srv.AddTool(mcp.NewTool("admin_create_field_type", mcp.WithDescription("Create a new admin field type."), mcp.WithString("type", mcp.Required(), mcp.Description("Type key")), mcp.WithString("label", mcp.Required(), mcp.Description("Label"))), handleAdminCreateFieldType(client))
	srv.AddTool(mcp.NewTool("admin_update_field_type", mcp.WithDescription("Update an admin field type."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field type ID")), mcp.WithString("type", mcp.Required(), mcp.Description("Type key")), mcp.WithString("label", mcp.Required(), mcp.Description("Label"))), handleAdminUpdateFieldType(client))
	srv.AddTool(mcp.NewTool("admin_delete_field_type", mcp.WithDescription("Delete an admin field type by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field type ID"))), handleAdminDeleteFieldType(client))
}

// --- Admin Route Handlers ---

func handleAdminListRoutes(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.AdminRoutes.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminGetRoute(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.AdminRoutes.Get(ctx, modulacms.AdminRouteID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminCreateRoute(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError("title is required"), nil
		}
		status := int64(req.GetFloat("status", 0))
		params := modulacms.CreateAdminRouteParams{
			Slug:     modulacms.Slug(slug),
			Title:    title,
			Status:   status,
			AuthorID: optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.AdminRoutes.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminUpdateRoute(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError("title is required"), nil
		}
		status := int64(req.GetFloat("status", 0))
		params := modulacms.UpdateAdminRouteParams{
			Slug:     modulacms.Slug(slug),
			Title:    title,
			Status:   status,
			AuthorID: optionalIDPtr[modulacms.UserID](req, "author_id"),
			Slug2:    modulacms.Slug(req.GetString("new_slug", "")),
		}
		result, err := client.AdminRoutes.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminDeleteRoute(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.AdminRoutes.Delete(ctx, modulacms.AdminRouteID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Admin Field Type Handlers ---

func handleAdminListFieldTypes(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.AdminFieldTypes.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminGetFieldType(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.AdminFieldTypes.Get(ctx, modulacms.AdminFieldTypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminCreateFieldType(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		typ, err := req.RequireString("type")
		if err != nil {
			return mcp.NewToolResultError("type is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modulacms.CreateAdminFieldTypeParams{
			Type:  typ,
			Label: label,
		}
		result, err := client.AdminFieldTypes.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminUpdateFieldType(client *modulacms.Client) server.ToolHandlerFunc {
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
		params := modulacms.UpdateAdminFieldTypeParams{
			AdminFieldTypeID: modulacms.AdminFieldTypeID(id),
			Type:             typ,
			Label:            label,
		}
		result, err := client.AdminFieldTypes.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminDeleteFieldType(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.AdminFieldTypes.Delete(ctx, modulacms.AdminFieldTypeID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
