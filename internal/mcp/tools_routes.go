package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modula "github.com/hegner123/modulacms/sdks/go"
)

func registerRouteTools(srv *server.MCPServer, client *modula.Client) {
	srv.AddTool(
		mcp.NewTool("list_routes",
			mcp.WithDescription("List all routes."),
		),
		handleListRoutes(client),
	)

	srv.AddTool(
		mcp.NewTool("get_route",
			mcp.WithDescription("Get a single route by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Route ID (ULID)")),
		),
		handleGetRoute(client),
	)

	srv.AddTool(
		mcp.NewTool("create_route",
			mcp.WithDescription("Create a new route."),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug (e.g. 'about' or 'blog/my-post')")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Route title")),
			mcp.WithNumber("status", mcp.Required(), mcp.Description("Route status (positive integer)")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleCreateRoute(client),
	)

	srv.AddTool(
		mcp.NewTool("update_route",
			mcp.WithDescription("Update a route by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Route ID (ULID)")),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Route title")),
			mcp.WithNumber("status", mcp.Required(), mcp.Description("Route status (positive integer)")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateRoute(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_route",
			mcp.WithDescription("Delete a route by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Route ID (ULID)")),
		),
		handleDeleteRoute(client),
	)
}

func handleListRoutes(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Routes.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetRoute(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Routes.Get(ctx, modula.RouteID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateRoute(client *modula.Client) server.ToolHandlerFunc {
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
		params := modula.CreateRouteParams{
			Slug:     modula.Slug(slug),
			Title:    title,
			Status:   status,
			AuthorID: optionalIDPtr[modula.UserID](req, "author_id"),
		}
		result, err := client.Routes.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateRoute(client *modula.Client) server.ToolHandlerFunc {
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
		params := modula.UpdateRouteParams{
			RouteID:  modula.RouteID(id),
			Slug:     modula.Slug(slug),
			Title:    title,
			Status:   status,
			AuthorID: optionalIDPtr[modula.UserID](req, "author_id"),
		}
		result, err := client.Routes.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteRoute(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Routes.Delete(ctx, modula.RouteID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
