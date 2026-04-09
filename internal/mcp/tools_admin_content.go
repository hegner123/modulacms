package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAdminContentTools(srv *server.MCPServer, backend AdminContentBackend) {
	srv.AddTool(
		mcp.NewTool("admin_list_content",
			mcp.WithDescription("List admin content data entries with pagination."),
			mcp.WithNumber("limit", mcp.Description("Max items (default 20)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleAdminListContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_content",
			mcp.WithDescription("Get a single admin content data entry by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
		),
		handleAdminGetContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_create_content",
			mcp.WithDescription("Create a new admin content data entry."),
			mcp.WithString("status", mcp.Required(), mcp.Description("Content status"), mcp.Enum("draft", "published", "archived", "pending")),
			mcp.WithString("parent_id", mcp.Description("Parent admin content ID")),
			mcp.WithString("admin_route_id", mcp.Description("Admin route ID")),
			mcp.WithString("admin_datatype_id", mcp.Description("Admin datatype ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleAdminCreateContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_update_content",
			mcp.WithDescription("update an existing admin content data entry."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
			mcp.WithString("status", mcp.Required(), mcp.Description("Content status"), mcp.Enum("draft", "published", "archived", "pending")),
			mcp.WithString("parent_id", mcp.Description("Parent admin content ID")),
			mcp.WithString("admin_route_id", mcp.Description("Admin route ID")),
			mcp.WithString("admin_datatype_id", mcp.Description("Admin datatype ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleAdminUpdateContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_content",
			mcp.WithDescription("Delete an admin content data entry by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
		),
		handleAdminDeleteContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_reorder_content",
			mcp.WithDescription("Atomically reorder admin content siblings under a parent."),
			mcp.WithString("parent_id", mcp.Description("Parent admin content ID")),
			mcp.WithObject("ordered_ids", mcp.Required(), mcp.Description("Array of admin content IDs in desired order")),
		),
		handleAdminReorderContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_move_content",
			mcp.WithDescription("Move an admin content node to a new parent."),
			mcp.WithString("node_id", mcp.Required(), mcp.Description("Admin content ID to move")),
			mcp.WithString("new_parent_id", mcp.Description("New parent admin content ID")),
			mcp.WithNumber("position", mcp.Required(), mcp.Description("Zero-based position")),
		),
		handleAdminMoveContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_list_content_fields",
			mcp.WithDescription("List admin content field records with pagination."),
			mcp.WithNumber("limit", mcp.Description("Max items (default 20)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleAdminListContentFields(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_content_field",
			mcp.WithDescription("Get a single admin content field by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content field ID (ULID)")),
		),
		handleAdminGetContentField(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_create_content_field",
			mcp.WithDescription("Create a new admin content field."),
			mcp.WithString("admin_content_data_id", mcp.Required(), mcp.Description("Admin content data ID")),
			mcp.WithString("admin_field_id", mcp.Required(), mcp.Description("Admin field ID")),
			mcp.WithString("admin_field_value", mcp.Required(), mcp.Description("Field value")),
			mcp.WithString("admin_route_id", mcp.Description("Admin route ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleAdminCreateContentField(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_update_content_field",
			mcp.WithDescription("update an existing admin content field."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content field ID (ULID)")),
			mcp.WithString("admin_field_value", mcp.Required(), mcp.Description("New field value")),
			mcp.WithString("admin_content_data_id", mcp.Description("Admin content data ID")),
			mcp.WithString("admin_field_id", mcp.Description("Admin field ID")),
			mcp.WithString("admin_route_id", mcp.Description("Admin route ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleAdminUpdateContentField(backend),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_content_field",
			mcp.WithDescription("Delete an admin content field by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content field ID (ULID)")),
		),
		handleAdminDeleteContentField(backend),
	)
}

func handleAdminListContent(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.ListAdminContent(ctx, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetContent(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetAdminContent(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateContent(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := req.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError("status is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"parent_id":         optionalStrPtr(req, "parent_id"),
			"admin_route_id":    optionalStrPtr(req, "admin_route_id"),
			"admin_datatype_id": optionalStrPtr(req, "admin_datatype_id"),
			"author_id":         optionalStrPtr(req, "author_id"),
			"status":            status,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateAdminContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateContent(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		status, err := req.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError("status is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"admin_content_data_id": id,
			"parent_id":             optionalStrPtr(req, "parent_id"),
			"admin_route_id":        optionalStrPtr(req, "admin_route_id"),
			"admin_datatype_id":     optionalStrPtr(req, "admin_datatype_id"),
			"author_id":             optionalStrPtr(req, "author_id"),
			"status":                status,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteContent(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteAdminContent(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleAdminReorderContent(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		rawIDs, ok := args["ordered_ids"].([]any)
		if !ok {
			if s, sOk := args["ordered_ids"].(string); sOk {
				var parsed []string
				if err := json.Unmarshal([]byte(s), &parsed); err != nil {
					return mcp.NewToolResultError("ordered_ids must be a JSON array of admin content IDs"), nil
				}
				rawIDs = make([]any, len(parsed))
				for i, v := range parsed {
					rawIDs[i] = v
				}
			}
		}
		if len(rawIDs) == 0 {
			return mcp.NewToolResultError("ordered_ids must be a non-empty array"), nil
		}
		ids := make([]string, 0, len(rawIDs))
		for _, raw := range rawIDs {
			s, ok := raw.(string)
			if !ok {
				return mcp.NewToolResultError("each ordered_id must be a string"), nil
			}
			ids = append(ids, s)
		}
		params, err := marshalParams(map[string]any{
			"parent_id":   optionalStrPtr(req, "parent_id"),
			"ordered_ids": ids,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.ReorderAdminContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminMoveContent(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodeID, err := req.RequireString("node_id")
		if err != nil {
			return mcp.NewToolResultError("node_id is required"), nil
		}
		position := int(req.GetFloat("position", 0))
		params, err := marshalParams(map[string]any{
			"node_id":       nodeID,
			"new_parent_id": optionalStrPtr(req, "new_parent_id"),
			"position":      position,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.MoveAdminContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminListContentFields(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.ListAdminContentFields(ctx, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminGetContentField(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetAdminContentField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminCreateContentField(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cdID, err := req.RequireString("admin_content_data_id")
		if err != nil {
			return mcp.NewToolResultError("admin_content_data_id is required"), nil
		}
		fID, err := req.RequireString("admin_field_id")
		if err != nil {
			return mcp.NewToolResultError("admin_field_id is required"), nil
		}
		fVal, err := req.RequireString("admin_field_value")
		if err != nil {
			return mcp.NewToolResultError("admin_field_value is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"admin_content_data_id": cdID,
			"admin_field_id":        fID,
			"admin_field_value":     fVal,
			"admin_route_id":        optionalStrPtr(req, "admin_route_id"),
			"author_id":             optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateAdminContentField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminUpdateContentField(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		fVal, err := req.RequireString("admin_field_value")
		if err != nil {
			return mcp.NewToolResultError("admin_field_value is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"admin_content_field_id": id,
			"admin_content_data_id":  optionalStrPtr(req, "admin_content_data_id"),
			"admin_field_id":         optionalStrPtr(req, "admin_field_id"),
			"admin_field_value":      fVal,
			"admin_route_id":         optionalStrPtr(req, "admin_route_id"),
			"author_id":              optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateAdminContentField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleAdminDeleteContentField(backend AdminContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteAdminContentField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
