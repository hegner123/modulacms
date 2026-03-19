package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAdminMediaTools(srv *server.MCPServer, backend AdminMediaBackend, folderBackend AdminMediaFolderBackend) {
	srv.AddTool(
		mcp.NewTool("admin_list_media",
			mcp.WithDescription("List admin media assets with pagination. Admin media files are used by the admin panel UI itself."),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleAdminListMedia(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_media",
			mcp.WithDescription("Get a single admin media asset by ID. Returns metadata including URL, dimensions, alt text, focal point, etc."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin media ID (ULID)")),
		),
		handleAdminGetMedia(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_update_media",
			mcp.WithDescription("Update admin media asset metadata. Only provided fields are changed; omitted fields remain unchanged."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin media ID (ULID)")),
			mcp.WithString("name", mcp.Description("File name")),
			mcp.WithString("display_name", mcp.Description("Display name")),
			mcp.WithString("alt", mcp.Description("Alt text for accessibility")),
			mcp.WithString("caption", mcp.Description("Caption text")),
			mcp.WithString("description", mcp.Description("Description text")),
			mcp.WithString("class", mcp.Description("CSS class")),
			mcp.WithNumber("focal_x", mcp.Description("Focal point X coordinate (0.0 to image width)")),
			mcp.WithNumber("focal_y", mcp.Description("Focal point Y coordinate (0.0 to image height)")),
		),
		handleAdminUpdateMedia(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_media",
			mcp.WithDescription("Delete an admin media asset by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin media ID (ULID)")),
		),
		handleAdminDeleteMedia(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_upload_media",
			mcp.WithDescription("Upload a media file to the admin media library from a local file path. Returns the created admin media entity. Optionally place the uploaded file into an admin media folder."),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the file to upload")),
			mcp.WithString("filename", mcp.Description("Override filename (defaults to base name of file_path)")),
			mcp.WithString("folder_id", mcp.Description("Admin media folder ID (ULID) to place the uploaded file into")),
		),
		handleAdminUploadMedia(backend, folderBackend),
	)

	// admin_list_media_dimensions reuses the public media dimensions — they are shared.
	// The issue specifies this tool but dimensions are global, not per-content-system.
	// We register it here pointing at the same shared dimension data.
	srv.AddTool(
		mcp.NewTool("admin_list_media_dimensions",
			mcp.WithDescription("List all media dimension presets. Dimensions are shared between public and admin media."),
		),
		handleAdminListMediaDimensions(backend),
	)
}

func handleAdminListMedia(backend AdminMediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.ListAdminMedia(ctx, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetMedia(backend AdminMediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetAdminMedia(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateMedia(backend AdminMediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"admin_media_id": id,
			"name":           optionalStrPtr(req, "name"),
			"display_name":   optionalStrPtr(req, "display_name"),
			"alt":            optionalStrPtr(req, "alt"),
			"caption":        optionalStrPtr(req, "caption"),
			"description":    optionalStrPtr(req, "description"),
			"class":          optionalStrPtr(req, "class"),
			"focal_x":        optionalFloat64Ptr(req, "focal_x"),
			"focal_y":        optionalFloat64Ptr(req, "focal_y"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminMedia(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteMedia(backend AdminMediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteAdminMedia(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleAdminUploadMedia(backend AdminMediaBackend, folderBackend AdminMediaFolderBackend) server.ToolHandlerFunc {
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
		data, err := backend.UploadAdminMedia(ctx, f, filename)
		if err != nil {
			return errResult(err), nil
		}

		// If folder_id was provided, move the uploaded media into that folder.
		folderID := req.GetString("folder_id", "")
		if folderID != "" {
			// Extract the admin_media_id from the upload response.
			var uploaded struct {
				AdminMediaID string `json:"admin_media_id"`
			}
			if jsonErr := json.Unmarshal(data, &uploaded); jsonErr == nil && uploaded.AdminMediaID != "" {
				moveParams, marshalErr := marshalParams(map[string]any{
					"media_ids": []string{uploaded.AdminMediaID},
					"folder_id": &folderID,
				})
				if marshalErr == nil {
					moveData, moveErr := folderBackend.MoveAdminMediaToFolder(ctx, moveParams)
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

func handleAdminListMediaDimensions(backend AdminMediaBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListMediaDimensions(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
