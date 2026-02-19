package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerMediaTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("list_media",
			mcp.WithDescription("List media assets with pagination. Media files must be uploaded through the CMS web interface or API directly. This MCP server can view and update media metadata but cannot upload new files."),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleListMedia(client),
	)

	srv.AddTool(
		mcp.NewTool("get_media",
			mcp.WithDescription("Get a single media asset by ID. Returns metadata including URL, dimensions, alt text, focal point, etc."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media ID (ULID)")),
		),
		handleGetMedia(client),
	)

	srv.AddTool(
		mcp.NewTool("update_media",
			mcp.WithDescription("Update media asset metadata. Only provided fields are changed; omitted fields remain unchanged."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media ID (ULID)")),
			mcp.WithString("name", mcp.Description("File name")),
			mcp.WithString("display_name", mcp.Description("Display name")),
			mcp.WithString("alt", mcp.Description("Alt text for accessibility")),
			mcp.WithString("caption", mcp.Description("Caption text")),
			mcp.WithString("description", mcp.Description("Description text")),
			mcp.WithString("class", mcp.Description("CSS class")),
			mcp.WithNumber("focal_x", mcp.Description("Focal point X coordinate (0.0 to image width)")),
			mcp.WithNumber("focal_y", mcp.Description("Focal point Y coordinate (0.0 to image height)")),
		),
		handleUpdateMedia(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_media",
			mcp.WithDescription("Delete a media asset by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media ID (ULID)")),
		),
		handleDeleteMedia(client),
	)
}

func handleListMedia(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		result, err := client.Media.ListPaginated(ctx, modulacms.PaginationParams{
			Limit: limit, Offset: offset,
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetMedia(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Media.Get(ctx, modulacms.MediaID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateMedia(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params := modulacms.UpdateMediaParams{
			MediaID:     modulacms.MediaID(id),
			Name:        optionalStrPtr(req, "name"),
			DisplayName: optionalStrPtr(req, "display_name"),
			Alt:         optionalStrPtr(req, "alt"),
			Caption:     optionalStrPtr(req, "caption"),
			Description: optionalStrPtr(req, "description"),
			Class:       optionalStrPtr(req, "class"),
			FocalX:      optionalFloat64Ptr(req, "focal_x"),
			FocalY:      optionalFloat64Ptr(req, "focal_y"),
		}
		result, err := client.Media.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteMedia(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Media.Delete(ctx, modulacms.MediaID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
