package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAdminRouteTools(srv *server.MCPServer, backend AdminRouteBackend) {
	// Admin Routes
	srv.AddTool(mcp.NewTool("admin_list_routes", mcp.WithDescription("List all admin routes.")), handleAdminListRoutes(backend))
	srv.AddTool(mcp.NewTool("admin_get_route", mcp.WithDescription("Get a single admin route by slug. The admin routes API uses slug-based lookup."), mcp.WithString("slug", mcp.Required(), mcp.Description("Admin route slug (e.g. '/admin')"))), handleAdminGetRoute(backend))
	srv.AddTool(mcp.NewTool("admin_create_route", mcp.WithDescription("Create a new admin route."), mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug")), mcp.WithString("title", mcp.Required(), mcp.Description("Title")), mcp.WithNumber("status", mcp.Required(), mcp.Description("Status")), mcp.WithString("author_id", mcp.Description("Author user ID"))), handleAdminCreateRoute(backend))
	srv.AddTool(mcp.NewTool("admin_update_route", mcp.WithDescription("update an admin route by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin route ID (ULID)")), mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug")), mcp.WithString("title", mcp.Required(), mcp.Description("Title")), mcp.WithNumber("status", mcp.Required(), mcp.Description("Status")), mcp.WithString("author_id", mcp.Description("Author"))), handleAdminUpdateRoute(backend))
	srv.AddTool(mcp.NewTool("admin_delete_route", mcp.WithDescription("Delete an admin route by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin route ID"))), handleAdminDeleteRoute(backend))

	// Admin Field Types
	srv.AddTool(mcp.NewTool("admin_list_field_types", mcp.WithDescription("List all admin field types.")), handleAdminListFieldTypes(backend))
	srv.AddTool(mcp.NewTool("admin_get_field_type", mcp.WithDescription("Get a single admin field type by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field type ID (ULID)"))), handleAdminGetFieldType(backend))
	srv.AddTool(mcp.NewTool("admin_create_field_type", mcp.WithDescription("Create a new admin field type."), mcp.WithString("type", mcp.Required(), mcp.Description("Type key")), mcp.WithString("label", mcp.Required(), mcp.Description("Label"))), handleAdminCreateFieldType(backend))
	srv.AddTool(mcp.NewTool("admin_update_field_type", mcp.WithDescription("update an admin field type."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field type ID")), mcp.WithString("type", mcp.Required(), mcp.Description("Type key")), mcp.WithString("label", mcp.Required(), mcp.Description("Label"))), handleAdminUpdateFieldType(backend))
	srv.AddTool(mcp.NewTool("admin_delete_field_type", mcp.WithDescription("Delete an admin field type by ID."), mcp.WithString("id", mcp.Required(), mcp.Description("Admin field type ID"))), handleAdminDeleteFieldType(backend))
}

// --- Admin Route Handlers ---

func handleAdminListRoutes(backend AdminRouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListAdminRoutes(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetRoute(backend AdminRouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		data, err := backend.GetAdminRoute(ctx, slug)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateRoute(backend AdminRouteBackend) server.ToolHandlerFunc {
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
		params, err := marshalParams(map[string]any{
			"slug":      slug,
			"title":     title,
			"status":    status,
			"author_id": optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateAdminRoute(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateRoute(backend AdminRouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		title, err := req.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError("title is required"), nil
		}
		status := int64(req.GetFloat("status", 0))
		params, err := marshalParams(map[string]any{
			"admin_route_id": id,
			"slug":           slug,
			"title":          title,
			"status":         status,
			"author_id":      optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminRoute(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteRoute(backend AdminRouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteAdminRoute(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Admin Field Type Handlers ---

func handleAdminListFieldTypes(backend AdminRouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListAdminFieldTypes(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetFieldType(backend AdminRouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetAdminFieldType(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateFieldType(backend AdminRouteBackend) server.ToolHandlerFunc {
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
		data, err := backend.CreateAdminFieldType(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateFieldType(backend AdminRouteBackend) server.ToolHandlerFunc {
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
			"admin_field_type_id": id,
			"type":                typ,
			"label":               label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminFieldType(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteFieldType(backend AdminRouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteAdminFieldType(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
