package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerRouteTools(srv *server.MCPServer, backend RouteBackend) {
	srv.AddTool(
		mcp.NewTool("list_routes",
			mcp.WithDescription("List all routes."),
		),
		handleListRoutes(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_route",
			mcp.WithDescription("Get a single route by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Route ID (ULID)")),
		),
		handleGetRoute(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_route",
			mcp.WithDescription("Create a new route."),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug (e.g. 'about' or 'blog/my-post')")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Route title")),
			mcp.WithNumber("status", mcp.Required(), mcp.Description("Route status (positive integer)")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleCreateRoute(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_route",
			mcp.WithDescription("update a route by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Route ID (ULID)")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Route title")),
			mcp.WithNumber("status", mcp.Required(), mcp.Description("Route status (positive integer)")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateRoute(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_route",
			mcp.WithDescription("Delete a route by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Route ID (ULID)")),
		),
		handleDeleteRoute(backend),
	)
}

func handleListRoutes(backend RouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListRoutes(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetRoute(backend RouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetRoute(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateRoute(backend RouteBackend) server.ToolHandlerFunc {
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
		data, err := backend.CreateRoute(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateRoute(backend RouteBackend) server.ToolHandlerFunc {
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
			"route_id":  id,
			"slug":      slug,
			"title":     title,
			"status":    status,
			"author_id": optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateRoute(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteRoute(backend RouteBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteRoute(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
