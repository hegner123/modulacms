package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAdminMediaFolderTools(srv *server.MCPServer, backend AdminMediaFolderBackend) {
	srv.AddTool(
		mcp.NewTool("admin_list_media_folders",
			mcp.WithDescription("List admin media folders. Returns root folders by default, or children of a given parent folder."),
			mcp.WithString("parent_id", mcp.Description("Parent folder ID (ULID). If omitted, returns root folders.")),
		),
		handleAdminListMediaFolders(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_media_folder",
			mcp.WithDescription("Get an admin media folder by ID. Returns folder details including name, parent, and timestamps."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin media folder ID (ULID)")),
		),
		handleAdminGetMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_create_media_folder",
			mcp.WithDescription("Create a new admin media folder. Folders can be nested up to 10 levels deep. Names must be unique within a parent."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Folder name")),
			mcp.WithString("parent_id", mcp.Description("Parent folder ID (ULID). Omit to create at root.")),
		),
		handleAdminCreateMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_update_media_folder",
			mcp.WithDescription("Update an admin media folder. Rename or move to a different parent. Only provided fields are changed."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin media folder ID (ULID)")),
			mcp.WithString("name", mcp.Description("New folder name")),
			mcp.WithString("parent_id", mcp.Description("New parent folder ID (ULID). Use empty string to move to root.")),
		),
		handleAdminUpdateMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_media_folder",
			mcp.WithDescription("Delete an admin media folder. Fails if the folder contains child folders or media items, returning the counts."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin media folder ID (ULID)")),
		),
		handleAdminDeleteMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_move_media_to_folder",
			mcp.WithDescription("Move one or more admin media items to a folder. Set folder_id to empty string to move to root (unfiled). Maximum 100 items per batch."),
			mcp.WithObject("media_ids", mcp.Required(), mcp.Description("Array of admin media IDs (ULIDs) to move")),
			mcp.WithString("folder_id", mcp.Description("Target folder ID (ULID). Omit or empty string to move to root.")),
		),
		handleAdminMoveMediaToFolder(backend),
	)
}

func handleAdminListMediaFolders(backend AdminMediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parentID := req.GetString("parent_id", "")
		data, err := backend.ListAdminMediaFolders(ctx, parentID)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetMediaFolder(backend AdminMediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetAdminMediaFolder(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateMediaFolder(backend AdminMediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"name":      name,
			"parent_id": optionalStrPtr(req, "parent_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateAdminMediaFolder(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateMediaFolder(backend AdminMediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"folder_id": id,
			"name":      optionalStrPtr(req, "name"),
			"parent_id": optionalStrPtr(req, "parent_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminMediaFolder(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteMediaFolder(backend AdminMediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.DeleteAdminMediaFolder(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminMoveMediaToFolder(backend AdminMediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		mediaIDsRaw, ok := args["media_ids"]
		if !ok {
			return mcp.NewToolResultError("media_ids is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"media_ids": mediaIDsRaw,
			"folder_id": optionalStrPtr(req, "folder_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.MoveAdminMediaToFolder(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
