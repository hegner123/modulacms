package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMediaFolderTools(srv *server.MCPServer, backend MediaFolderBackend) {
	srv.AddTool(
		mcp.NewTool("list_media_folders",
			mcp.WithDescription("List media folders. Returns root folders by default, or children of a given parent folder."),
			mcp.WithString("parent_id", mcp.Description("Parent folder ID (ULID). If omitted, returns root folders.")),
		),
		handleListMediaFolders(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_media_folder",
			mcp.WithDescription("Get a media folder by ID. Returns folder details including name, parent, and timestamps."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media folder ID (ULID)")),
		),
		handleGetMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_media_folder",
			mcp.WithDescription("Create a new media folder. Folders can be nested up to 10 levels deep. Names must be unique within a parent."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Folder name")),
			mcp.WithString("parent_id", mcp.Description("Parent folder ID (ULID). Omit to create at root.")),
		),
		handleCreateMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_media_folder",
			mcp.WithDescription("Update a media folder. Rename or move to a different parent. Only provided fields are changed."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media folder ID (ULID)")),
			mcp.WithString("name", mcp.Description("New folder name")),
			mcp.WithString("parent_id", mcp.Description("New parent folder ID (ULID). Use empty string to move to root.")),
		),
		handleUpdateMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_media_folder",
			mcp.WithDescription("Delete a media folder. Fails if the folder contains child folders or media items, returning the counts."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media folder ID (ULID)")),
		),
		handleDeleteMediaFolder(backend),
	)

	srv.AddTool(
		mcp.NewTool("move_media_to_folder",
			mcp.WithDescription("Move one or more media items to a folder. Set folder_id to empty string to move to root (unfiled). Maximum 100 items per batch."),
			mcp.WithObject("media_ids", mcp.Required(), mcp.Description("Array of media IDs (ULIDs) to move")),
			mcp.WithString("folder_id", mcp.Description("Target folder ID (ULID). Omit or empty string to move to root.")),
		),
		handleMoveMediaToFolder(backend),
	)

	// get_media_folder_tree
	srv.AddTool(
		mcp.NewTool("get_media_folder_tree",
			mcp.WithDescription("Get the full media folder tree hierarchy."),
		),
		handleGetMediaFolderTree(backend),
	)

	// list_media_in_folder
	srv.AddTool(
		mcp.NewTool("list_media_in_folder",
			mcp.WithDescription("List media files within a specific folder."),
			mcp.WithString("folder_id", mcp.Required(), mcp.Description("Media folder ID (ULID)")),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleListMediaInFolder(backend),
	)
}

func handleListMediaFolders(backend MediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		parentID := req.GetString("parent_id", "")
		data, err := backend.ListMediaFolders(ctx, parentID)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetMediaFolder(backend MediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetMediaFolder(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateMediaFolder(backend MediaFolderBackend) server.ToolHandlerFunc {
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
		data, err := backend.CreateMediaFolder(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateMediaFolder(backend MediaFolderBackend) server.ToolHandlerFunc {
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
		data, err := backend.UpdateMediaFolder(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteMediaFolder(backend MediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.DeleteMediaFolder(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleMoveMediaToFolder(backend MediaFolderBackend) server.ToolHandlerFunc {
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
		data, err := backend.MoveMediaToFolder(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetMediaFolderTree(backend MediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.GetMediaFolderTree(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListMediaInFolder(backend MediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		folderID, err := req.RequireString("folder_id")
		if err != nil {
			return mcp.NewToolResultError("folder_id is required"), nil
		}
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.ListMediaInFolder(ctx, folderID, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
