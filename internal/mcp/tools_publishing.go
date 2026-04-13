package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerPublishingTools(srv *server.MCPServer, backend PublishingBackend) {
	srv.AddTool(
		mcp.NewTool("publish_content",
			mcp.WithDescription("Publish a content item. Creates a published snapshot."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithString("locale", mcp.Description("Locale code (default 'en')")),
		),
		handlePublishContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("unpublish_content",
			mcp.WithDescription("Unpublish a content item. Removes it from the public API."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithString("locale", mcp.Description("Locale code (default 'en')")),
		),
		handleUnpublishContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("schedule_content",
			mcp.WithDescription("Schedule a content item for future publishing. publish_at must be RFC3339 format."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithString("publish_at", mcp.Required(), mcp.Description("RFC3339 timestamp for scheduled publication")),
		),
		handleScheduleContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_publish_content",
			mcp.WithDescription("Publish an admin content item."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
			mcp.WithString("locale", mcp.Description("Locale code (default 'en')")),
		),
		handleAdminPublishContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_unpublish_content",
			mcp.WithDescription("Unpublish an admin content item."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
			mcp.WithString("locale", mcp.Description("Locale code (default 'en')")),
		),
		handleAdminUnpublishContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_schedule_content",
			mcp.WithDescription("Schedule an admin content item for future publishing. publish_at must be RFC3339 format."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
			mcp.WithString("publish_at", mcp.Required(), mcp.Description("RFC3339 timestamp for scheduled publication")),
		),
		handleAdminScheduleContent(backend),
	)
}

func handlePublishContent(backend PublishingBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		locale := req.GetString("locale", "")
		params, err := marshalParams(map[string]any{
			"content_data_id": contentID,
			"locale":          locale,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.PublishContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUnpublishContent(backend PublishingBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		locale := req.GetString("locale", "")
		params, err := marshalParams(map[string]any{
			"content_data_id": contentID,
			"locale":          locale,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UnpublishContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleScheduleContent(backend PublishingBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		publishAt, err := req.RequireString("publish_at")
		if err != nil {
			return mcp.NewToolResultError("publish_at is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"content_data_id": contentID,
			"publish_at":      publishAt,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.ScheduleContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminPublishContent(backend PublishingBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		locale := req.GetString("locale", "")
		params, err := marshalParams(map[string]any{
			"admin_content_data_id": contentID,
			"locale":                locale,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminPublishContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUnpublishContent(backend PublishingBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		locale := req.GetString("locale", "")
		params, err := marshalParams(map[string]any{
			"admin_content_data_id": contentID,
			"locale":                locale,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminUnpublishContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminScheduleContent(backend PublishingBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		publishAt, err := req.RequireString("publish_at")
		if err != nil {
			return mcp.NewToolResultError("publish_at is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"admin_content_data_id": contentID,
			"publish_at":            publishAt,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminScheduleContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
