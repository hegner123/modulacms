package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerAdminContentTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("admin_list_content",
			mcp.WithDescription("List admin content data entries with pagination."),
			mcp.WithNumber("limit", mcp.Description("Max items (default 20)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleAdminListContent(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_content",
			mcp.WithDescription("Get a single admin content data entry by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
		),
		handleAdminGetContent(client),
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
		handleAdminCreateContent(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_update_content",
			mcp.WithDescription("Update an existing admin content data entry."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
			mcp.WithString("status", mcp.Required(), mcp.Description("Content status"), mcp.Enum("draft", "published", "archived", "pending")),
			mcp.WithString("parent_id", mcp.Description("Parent admin content ID")),
			mcp.WithString("admin_route_id", mcp.Description("Admin route ID")),
			mcp.WithString("admin_datatype_id", mcp.Description("Admin datatype ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleAdminUpdateContent(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_content",
			mcp.WithDescription("Delete an admin content data entry by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content data ID (ULID)")),
		),
		handleAdminDeleteContent(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_reorder_content",
			mcp.WithDescription("Atomically reorder admin content siblings under a parent."),
			mcp.WithString("parent_id", mcp.Description("Parent admin content ID")),
			mcp.WithObject("ordered_ids", mcp.Required(), mcp.Description("Array of admin content IDs in desired order")),
		),
		handleAdminReorderContent(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_move_content",
			mcp.WithDescription("Move an admin content node to a new parent."),
			mcp.WithString("node_id", mcp.Required(), mcp.Description("Admin content ID to move")),
			mcp.WithString("new_parent_id", mcp.Description("New parent admin content ID")),
			mcp.WithNumber("position", mcp.Required(), mcp.Description("Zero-based position")),
		),
		handleAdminMoveContent(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_list_content_fields",
			mcp.WithDescription("List admin content field records with pagination."),
			mcp.WithNumber("limit", mcp.Description("Max items (default 20)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleAdminListContentFields(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_get_content_field",
			mcp.WithDescription("Get a single admin content field by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content field ID (ULID)")),
		),
		handleAdminGetContentField(client),
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
		handleAdminCreateContentField(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_update_content_field",
			mcp.WithDescription("Update an existing admin content field."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content field ID (ULID)")),
			mcp.WithString("admin_field_value", mcp.Required(), mcp.Description("New field value")),
			mcp.WithString("admin_content_data_id", mcp.Description("Admin content data ID")),
			mcp.WithString("admin_field_id", mcp.Description("Admin field ID")),
			mcp.WithString("admin_route_id", mcp.Description("Admin route ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleAdminUpdateContentField(client),
	)

	srv.AddTool(
		mcp.NewTool("admin_delete_content_field",
			mcp.WithDescription("Delete an admin content field by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Admin content field ID (ULID)")),
		),
		handleAdminDeleteContentField(client),
	)
}

func handleAdminListContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		result, err := client.AdminContentData.ListPaginated(ctx, modulacms.PaginationParams{Limit: limit, Offset: offset})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminGetContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.AdminContentData.Get(ctx, modulacms.AdminContentID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminCreateContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := req.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError("status is required"), nil
		}
		params := modulacms.CreateAdminContentDataParams{
			ParentID:        optionalIDPtr[modulacms.AdminContentID](req, "parent_id"),
			AdminRouteID:    optionalIDPtr[modulacms.AdminRouteID](req, "admin_route_id"),
			AdminDatatypeID: optionalIDPtr[modulacms.AdminDatatypeID](req, "admin_datatype_id"),
			AuthorID:        optionalIDPtr[modulacms.UserID](req, "author_id"),
			Status:          modulacms.ContentStatus(status),
		}
		result, err := client.AdminContentData.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminUpdateContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		status, err := req.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError("status is required"), nil
		}
		params := modulacms.UpdateAdminContentDataParams{
			AdminContentDataID: modulacms.AdminContentID(id),
			ParentID:           optionalIDPtr[modulacms.AdminContentID](req, "parent_id"),
			AdminRouteID:       optionalIDPtr[modulacms.AdminRouteID](req, "admin_route_id"),
			AdminDatatypeID:    optionalIDPtr[modulacms.AdminDatatypeID](req, "admin_datatype_id"),
			AuthorID:           optionalIDPtr[modulacms.UserID](req, "author_id"),
			Status:             modulacms.ContentStatus(status),
		}
		result, err := client.AdminContentData.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminDeleteContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.AdminContentData.Delete(ctx, modulacms.AdminContentID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

func handleAdminReorderContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		rawIDs, ok := args["ordered_ids"].([]any)
		if !ok || len(rawIDs) == 0 {
			return mcp.NewToolResultError("ordered_ids must be a non-empty array"), nil
		}
		ids := make([]modulacms.AdminContentID, 0, len(rawIDs))
		for _, raw := range rawIDs {
			s, ok := raw.(string)
			if !ok {
				return mcp.NewToolResultError("each ordered_id must be a string"), nil
			}
			ids = append(ids, modulacms.AdminContentID(s))
		}
		params := modulacms.AdminContentReorderRequest{
			ParentID:   optionalIDPtr[modulacms.AdminContentID](req, "parent_id"),
			OrderedIDs: ids,
		}
		result, err := client.AdminContentReorder.Reorder(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminMoveContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodeID, err := req.RequireString("node_id")
		if err != nil {
			return mcp.NewToolResultError("node_id is required"), nil
		}
		position := int(req.GetFloat("position", 0))
		params := modulacms.AdminContentMoveRequest{
			NodeID:      modulacms.AdminContentID(nodeID),
			NewParentID: optionalIDPtr[modulacms.AdminContentID](req, "new_parent_id"),
			Position:    position,
		}
		result, err := client.AdminContentReorder.Move(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminListContentFields(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		result, err := client.AdminContentFields.ListPaginated(ctx, modulacms.PaginationParams{Limit: limit, Offset: offset})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminGetContentField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.AdminContentFields.Get(ctx, modulacms.AdminContentFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminCreateContentField(client *modulacms.Client) server.ToolHandlerFunc {
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
		contentDataID := modulacms.AdminContentID(cdID)
		fieldID := modulacms.AdminFieldID(fID)
		params := modulacms.CreateAdminContentFieldParams{
			AdminContentDataID: &contentDataID,
			AdminFieldID:       &fieldID,
			AdminFieldValue:    fVal,
			AdminRouteID:       optionalIDPtr[modulacms.AdminRouteID](req, "admin_route_id"),
			AuthorID:           optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.AdminContentFields.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminUpdateContentField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		fVal, err := req.RequireString("admin_field_value")
		if err != nil {
			return mcp.NewToolResultError("admin_field_value is required"), nil
		}
		cfID := modulacms.AdminContentFieldID(id)
		existing, err := client.AdminContentFields.Get(ctx, cfID)
		if err != nil {
			return errResult(err), nil
		}
		params := modulacms.UpdateAdminContentFieldParams{
			AdminContentFieldID: cfID,
			AdminContentDataID:  existing.AdminContentDataID,
			AdminFieldID:        existing.AdminFieldID,
			AdminFieldValue:     fVal,
			AdminRouteID:        existing.AdminRouteID,
			AuthorID:            existing.AuthorID,
		}
		if v := optionalIDPtr[modulacms.AdminContentID](req, "admin_content_data_id"); v != nil {
			params.AdminContentDataID = v
		}
		if v := optionalIDPtr[modulacms.AdminFieldID](req, "admin_field_id"); v != nil {
			params.AdminFieldID = v
		}
		if v := optionalIDPtr[modulacms.AdminRouteID](req, "admin_route_id"); v != nil {
			params.AdminRouteID = v
		}
		if v := optionalIDPtr[modulacms.UserID](req, "author_id"); v != nil {
			params.AuthorID = v
		}
		result, err := client.AdminContentFields.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleAdminDeleteContentField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.AdminContentFields.Delete(ctx, modulacms.AdminContentFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
