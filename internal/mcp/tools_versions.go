package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerVersionTools(srv *server.MCPServer, backend VersionBackend) {
	// --- Public Content Versions ---

	srv.AddTool(
		mcp.NewTool("list_content_versions",
			mcp.WithDescription("List all versions for a content item."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
		),
		handleListContentVersions(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_content_version",
			mcp.WithDescription("Get a single content version by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content version ID (ULID)")),
		),
		handleGetContentVersion(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_content_version",
			mcp.WithDescription("Create a new version snapshot of a content item."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithString("label", mcp.Description("Optional human-readable label for the version")),
		),
		handleCreateContentVersion(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_content_version",
			mcp.WithDescription("Delete a content version by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content version ID (ULID)")),
		),
		handleDeleteContentVersion(backend),
	)

	srv.AddTool(
		mcp.NewTool("restore_content_version",
			mcp.WithDescription("Restore a content item to a previous version."),
			mcp.WithString("version_id", mcp.Required(), mcp.Description("Content version ID to restore from")),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Content data ID to restore")),
		),
		handleRestoreContentVersion(backend),
	)

	// --- Admin Content Versions ---

	srv.AddTool(
		mcp.NewTool("admin_list_content_versions",
			mcp.WithDescription("List all versions for an admin content item."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
		),
		handleAdminListContentVersions(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_content_version",
			mcp.WithDescription("Get a single admin content version by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content version ID (ULID)")),
		),
		handleAdminGetContentVersion(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_create_content_version",
			mcp.WithDescription("Create a new version snapshot of an admin content item."),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
			mcp.WithString("label", mcp.Description("Optional human-readable label for the version")),
		),
		handleAdminCreateContentVersion(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_content_version",
			mcp.WithDescription("Delete an admin content version by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content version ID (ULID)")),
		),
		handleAdminDeleteContentVersion(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_restore_content_version",
			mcp.WithDescription("Restore an admin content item to a previous version."),
			mcp.WithString("version_id", mcp.Required(), mcp.Description("Admin content version ID to restore from")),
			mcp.WithString("content_id", mcp.Required(), mcp.Description("Admin content data ID to restore")),
		),
		handleAdminRestoreContentVersion(backend),
	)
}

// --- Public Content Version Handlers ---

func handleListContentVersions(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		data, err := backend.ListVersions(ctx, contentID)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetVersion(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		label := req.GetString("label", "")
		params, err := marshalParams(map[string]any{
			"content_data_id": contentID,
			"label":           label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateVersion(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		if err := backend.DeleteVersion(ctx, id); err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleRestoreContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		versionID, err := req.RequireString("version_id")
		if err != nil {
			return mcp.NewToolResultError("version_id is required"), nil
		}
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"content_data_id":    contentID,
			"content_version_id": versionID,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.RestoreVersion(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Admin Content Version Handlers ---

func handleAdminListContentVersions(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		data, err := backend.AdminListVersions(ctx, contentID)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.AdminGetVersion(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		label := req.GetString("label", "")
		params, err := marshalParams(map[string]any{
			"admin_content_data_id": contentID,
			"label":                 label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminCreateVersion(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		if err := backend.AdminDeleteVersion(ctx, id); err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleAdminRestoreContentVersion(backend VersionBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		versionID, err := req.RequireString("version_id")
		if err != nil {
			return mcp.NewToolResultError("version_id is required"), nil
		}
		contentID, err := req.RequireString("content_id")
		if err != nil {
			return mcp.NewToolResultError("content_id is required"), nil
		}
		params, err := json.Marshal(map[string]any{
			"admin_content_data_id":    contentID,
			"admin_content_version_id": versionID,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.AdminRestoreVersion(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
