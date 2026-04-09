package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerMediaTools(srv *server.MCPServer, backend MediaBackend, folderBackend MediaFolderBackend) {
	srv.AddTool(
		mcp.NewTool("list_media",
			mcp.WithDescription("List media assets with pagination. Media files must be uploaded through the CMS web interface or API directly. This MCP server can view and update media metadata but cannot upload new files."),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleListMedia(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_media",
			mcp.WithDescription("Get a single media asset by ID. Returns metadata including URL, dimensions, alt text, focal point, etc."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media ID (ULID)")),
		),
		handleGetMedia(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_media",
			mcp.WithDescription("update media asset metadata. Only provided fields are changed; omitted fields remain unchanged."),
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
		handleUpdateMedia(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_media",
			mcp.WithDescription("Delete a media asset by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media ID (ULID)")),
		),
		handleDeleteMedia(backend),
	)

	// upload_media
	srv.AddTool(
		mcp.NewTool("upload_media",
			mcp.WithDescription("Upload a media file from a local file path. Returns the created media entity. Optionally place the uploaded file into a media folder."),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the file to upload")),
			mcp.WithString("filename", mcp.Description("Override filename (defaults to base name of file_path)")),
			mcp.WithString("folder_id", mcp.Description("Media folder ID (ULID) to place the uploaded file into")),
		),
		handleUploadMedia(backend, folderBackend),
	)

	// media_health
	srv.AddTool(
		mcp.NewTool("media_health",
			mcp.WithDescription("Check media storage health status."),
		),
		handleMediaHealth(backend),
	)

	// media_cleanup
	srv.AddTool(
		mcp.NewTool("media_cleanup",
			mcp.WithDescription("Run orphaned media cleanup. Removes media records without backing files."),
		),
		handleMediaCleanup(backend),
	)

	// list_media_dimensions
	srv.AddTool(
		mcp.NewTool("list_media_dimensions",
			mcp.WithDescription("List all media dimension presets."),
		),
		handleListMediaDimensions(backend),
	)

	// get_media_dimension
	srv.AddTool(
		mcp.NewTool("get_media_dimension",
			mcp.WithDescription("Get a single media dimension preset by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media dimension ID (ULID)")),
		),
		handleGetMediaDimension(backend),
	)

	// create_media_dimension
	srv.AddTool(
		mcp.NewTool("create_media_dimension",
			mcp.WithDescription("Create a new media dimension preset."),
			mcp.WithString("label", mcp.Description("Dimension label")),
			mcp.WithNumber("width", mcp.Description("Width in pixels")),
			mcp.WithNumber("height", mcp.Description("Height in pixels")),
		),
		handleCreateMediaDimension(backend),
	)

	// update_media_dimension
	srv.AddTool(
		mcp.NewTool("update_media_dimension",
			mcp.WithDescription("update a media dimension preset."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media dimension ID (ULID)")),
			mcp.WithString("label", mcp.Description("Dimension label")),
			mcp.WithNumber("width", mcp.Description("Width in pixels")),
			mcp.WithNumber("height", mcp.Description("Height in pixels")),
		),
		handleUpdateMediaDimension(backend),
	)

	// delete_media_dimension
	srv.AddTool(
		mcp.NewTool("delete_media_dimension",
			mcp.WithDescription("Delete a media dimension preset by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media dimension ID (ULID)")),
		),
		handleDeleteMediaDimension(backend),
	)
}

func handleListMedia(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.ListMedia(ctx, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetMedia(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetMedia(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateMedia(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"media_id":     id,
			"name":         optionalStrPtr(req, "name"),
			"display_name": optionalStrPtr(req, "display_name"),
			"alt":          optionalStrPtr(req, "alt"),
			"caption":      optionalStrPtr(req, "caption"),
			"description":  optionalStrPtr(req, "description"),
			"class":        optionalStrPtr(req, "class"),
			"focal_x":      optionalFloat64Ptr(req, "focal_x"),
			"focal_y":      optionalFloat64Ptr(req, "focal_y"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateMedia(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteMedia(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteMedia(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleUploadMedia(backend MediaBackend, folderBackend MediaFolderBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := req.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError("file_path is required"), nil
		}
		f, err := os.Open(filePath)
		if err != nil {
			return mcp.NewToolResultError("cannot open file: " + err.Error()), nil
		}
		defer f.Close()
		filename := req.GetString("filename", "")
		if filename == "" {
			filename = filepath.Base(filePath)
		}
		data, err := backend.UploadMedia(ctx, f, filename)
		if err != nil {
			return errResult(err), nil
		}

		// If folder_id was provided, move the uploaded media into that folder.
		folderID := req.GetString("folder_id", "")
		if folderID != "" {
			// Extract the media_id from the upload response.
			var uploaded struct {
				MediaID string `json:"media_id"`
			}
			if jsonErr := json.Unmarshal(data, &uploaded); jsonErr == nil && uploaded.MediaID != "" {
				moveParams, marshalErr := marshalParams(map[string]any{
					"media_ids": []string{uploaded.MediaID},
					"folder_id": &folderID,
				})
				if marshalErr == nil {
					moveData, moveErr := folderBackend.MoveMediaToFolder(ctx, moveParams)
					if moveErr != nil {
						return errResult(moveErr), nil
					}
					// Return a combined response with the media and move result.
					var moveResult map[string]any
					json.Unmarshal(moveData, &moveResult) //nolint: best-effort parse for combined response
					var mediaResult map[string]any
					json.Unmarshal(data, &mediaResult) //nolint: best-effort parse for combined response
					if mediaResult != nil {
						mediaResult["folder_id"] = folderID
						mediaResult["move_result"] = moveResult
						combined, combErr := json.Marshal(mediaResult)
						if combErr == nil {
							return rawJSONResult(combined), nil
						}
					}
				}
			}
		}

		return rawJSONResult(data), nil
	}
}

func handleMediaHealth(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.MediaHealth(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleMediaCleanup(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.MediaCleanup(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListMediaDimensions(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListMediaDimensions(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetMediaDimension(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetMediaDimension(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateMediaDimension(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := marshalParams(map[string]any{
			"label":  optionalStrPtr(req, "label"),
			"width":  optionalFloat64Ptr(req, "width"),
			"height": optionalFloat64Ptr(req, "height"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateMediaDimension(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateMediaDimension(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"media_dimension_id": id,
			"label":              optionalStrPtr(req, "label"),
			"width":              optionalFloat64Ptr(req, "width"),
			"height":             optionalFloat64Ptr(req, "height"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateMediaDimension(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteMediaDimension(backend MediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteMediaDimension(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
