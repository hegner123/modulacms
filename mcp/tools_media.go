package main

import (
	"context"
	"os"
	"path/filepath"

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

	// upload_media
	srv.AddTool(
		mcp.NewTool("upload_media",
			mcp.WithDescription("Upload a media file from a local file path. Returns the created media entity."),
			mcp.WithString("file_path", mcp.Required(), mcp.Description("Absolute path to the file to upload")),
			mcp.WithString("filename", mcp.Description("Override filename (defaults to base name of file_path)")),
		),
		handleUploadMedia(client),
	)

	// media_health
	srv.AddTool(
		mcp.NewTool("media_health",
			mcp.WithDescription("Check media storage health status."),
		),
		handleMediaHealth(client),
	)

	// media_cleanup
	srv.AddTool(
		mcp.NewTool("media_cleanup",
			mcp.WithDescription("Run orphaned media cleanup. Removes media records without backing files."),
		),
		handleMediaCleanup(client),
	)

	// list_media_dimensions
	srv.AddTool(
		mcp.NewTool("list_media_dimensions",
			mcp.WithDescription("List all media dimension presets."),
		),
		handleListMediaDimensions(client),
	)

	// get_media_dimension
	srv.AddTool(
		mcp.NewTool("get_media_dimension",
			mcp.WithDescription("Get a single media dimension preset by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media dimension ID (ULID)")),
		),
		handleGetMediaDimension(client),
	)

	// create_media_dimension
	srv.AddTool(
		mcp.NewTool("create_media_dimension",
			mcp.WithDescription("Create a new media dimension preset."),
			mcp.WithString("label", mcp.Description("Dimension label")),
			mcp.WithNumber("width", mcp.Description("Width in pixels")),
			mcp.WithNumber("height", mcp.Description("Height in pixels")),
		),
		handleCreateMediaDimension(client),
	)

	// update_media_dimension
	srv.AddTool(
		mcp.NewTool("update_media_dimension",
			mcp.WithDescription("Update a media dimension preset."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media dimension ID (ULID)")),
			mcp.WithString("label", mcp.Description("Dimension label")),
			mcp.WithNumber("width", mcp.Description("Width in pixels")),
			mcp.WithNumber("height", mcp.Description("Height in pixels")),
		),
		handleUpdateMediaDimension(client),
	)

	// delete_media_dimension
	srv.AddTool(
		mcp.NewTool("delete_media_dimension",
			mcp.WithDescription("Delete a media dimension preset by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Media dimension ID (ULID)")),
		),
		handleDeleteMediaDimension(client),
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

func handleUploadMedia(client *modulacms.Client) server.ToolHandlerFunc {
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
		result, err := client.MediaUpload.Upload(ctx, f, filename, nil)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleMediaHealth(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.MediaAdmin.Health(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleMediaCleanup(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.MediaAdmin.Cleanup(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleListMediaDimensions(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.MediaDimensions.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetMediaDimension(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.MediaDimensions.Get(ctx, modulacms.MediaDimensionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateMediaDimension(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params := modulacms.CreateMediaDimensionParams{
			Label: optionalStrPtr(req, "label"),
		}
		if w := optionalFloat64Ptr(req, "width"); w != nil {
			v := int64(*w)
			params.Width = &v
		}
		if h := optionalFloat64Ptr(req, "height"); h != nil {
			v := int64(*h)
			params.Height = &v
		}
		result, err := client.MediaDimensions.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateMediaDimension(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		params := modulacms.UpdateMediaDimensionParams{
			MdID:  modulacms.MediaDimensionID(id),
			Label: optionalStrPtr(req, "label"),
		}
		if w := optionalFloat64Ptr(req, "width"); w != nil {
			v := int64(*w)
			params.Width = &v
		}
		if h := optionalFloat64Ptr(req, "height"); h != nil {
			v := int64(*h)
			params.Height = &v
		}
		result, err := client.MediaDimensions.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteMediaDimension(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.MediaDimensions.Delete(ctx, modulacms.MediaDimensionID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
