package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerContentTools(srv *server.MCPServer, client *modulacms.Client) {
	// --- Content Data CRUD ---

	srv.AddTool(
		mcp.NewTool("list_content",
			mcp.WithDescription("List content data entries with pagination. Returns structural metadata (IDs, status, timestamps, tree pointers) without field values. Use get_content_tree or get_page to see assembled content with field values."),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleListContent(client),
	)

	srv.AddTool(
		mcp.NewTool("get_content",
			mcp.WithDescription("Get a single content data entry by ID. Returns structural metadata without field values."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
		),
		handleGetContent(client),
	)

	srv.AddTool(
		mcp.NewTool("create_content",
			mcp.WithDescription("Create a new content data entry. author_id is NOT auto-populated; use the whoami tool to get your user ID and provide it explicitly. Status values: draft, published, archived, pending."),
			mcp.WithString("status", mcp.Required(), mcp.Description("Content status: draft, published, archived, pending"), mcp.Enum("draft", "published", "archived", "pending")),
			mcp.WithString("parent_id", mcp.Description("Parent content data ID")),
			mcp.WithString("route_id", mcp.Description("Route ID to associate with this content")),
			mcp.WithString("datatype_id", mcp.Description("Datatype ID defining the content schema")),
			mcp.WithString("author_id", mcp.Description("Author user ID (use whoami to get yours)")),
		),
		handleCreateContent(client),
	)

	srv.AddTool(
		mcp.NewTool("update_content",
			mcp.WithDescription("Update an existing content data entry. This is a full replacement — all fields are sent. Omitted pointer fields (parent_id, route_id, etc.) will be set to null. Status values: draft, published, archived, pending."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithString("status", mcp.Required(), mcp.Description("Content status: draft, published, archived, pending"), mcp.Enum("draft", "published", "archived", "pending")),
			mcp.WithString("parent_id", mcp.Description("Parent content data ID")),
			mcp.WithString("route_id", mcp.Description("Route ID")),
			mcp.WithString("datatype_id", mcp.Description("Datatype ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateContent(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_content",
			mcp.WithDescription("Delete a content data entry by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
		),
		handleDeleteContent(client),
	)

	// --- Content Delivery ---

	srv.AddTool(
		mcp.NewTool("get_page",
			mcp.WithDescription("Get assembled page content by slug via the public content delivery endpoint. Returns the full content tree with field values in the requested format."),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug of the page (e.g. 'about' or 'blog/my-post')")),
			mcp.WithString("format", mcp.Description("Response format: contentful, sanity, strapi, wordpress, clean, raw (default: server default)"), mcp.Enum("contentful", "sanity", "strapi", "wordpress", "clean", "raw")),
		),
		handleGetPage(client),
	)

	srv.AddTool(
		mcp.NewTool("get_content_tree",
			mcp.WithDescription("Get the admin content tree for a slug. Returns the content hierarchy with field values. To get the tree for a known content item, first use get_content to find its route, then use the route's slug here."),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug to retrieve the tree for")),
			mcp.WithString("format", mcp.Description("Response format: contentful, sanity, strapi, wordpress, clean, raw (default: server default)"), mcp.Enum("contentful", "sanity", "strapi", "wordpress", "clean", "raw")),
		),
		handleGetContentTree(client),
	)

	// --- Content Fields ---

	srv.AddTool(
		mcp.NewTool("list_content_fields",
			mcp.WithDescription("List content field records with pagination. Returns ALL content fields across all content items (cannot filter by content item). Use get_content_tree or get_page to see fields for a specific content item."),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleListContentFields(client),
	)

	srv.AddTool(
		mcp.NewTool("create_content_field",
			mcp.WithDescription("Create a new content field value. Links a field definition to a content data entry with a specific value."),
			mcp.WithString("content_data_id", mcp.Required(), mcp.Description("Content data ID this field belongs to")),
			mcp.WithString("field_id", mcp.Required(), mcp.Description("Field definition ID")),
			mcp.WithString("field_value", mcp.Required(), mcp.Description("The field value as a string")),
			mcp.WithString("route_id", mcp.Description("Route ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleCreateContentField(client),
	)

	srv.AddTool(
		mcp.NewTool("update_content_field",
			mcp.WithDescription("Update an existing content field value."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content field ID (ULID)")),
			mcp.WithString("field_value", mcp.Required(), mcp.Description("The new field value")),
			mcp.WithString("content_data_id", mcp.Description("Content data ID")),
			mcp.WithString("field_id", mcp.Description("Field definition ID")),
			mcp.WithString("route_id", mcp.Description("Route ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateContentField(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_content_field",
			mcp.WithDescription("Delete a content field record by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content field ID (ULID)")),
		),
		handleDeleteContentField(client),
	)

	// --- Content Batch ---

	srv.AddTool(
		mcp.NewTool("batch_update_content",
			mcp.WithDescription(`Atomically update content data and/or field values for a single content item.

Request structure:
- content_data_id: ULID of the content item (required)
- content_data: UpdateContentDataParams object (optional) — fields to update on the content data record
- fields: object mapping FieldID to new value (optional) — existing fields are updated, missing fields are created

author_id is auto-populated from the API token on the batch endpoint.

Response includes: content_data_id, content_data_updated, content_data_error, fields_updated, fields_created, fields_failed, errors.`),
			mcp.WithString("content_data_id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithObject("content_data", mcp.Description("Content data fields to update (parent_id, route_id, datatype_id, status, etc.)")),
			mcp.WithObject("fields", mcp.Description("Map of field_id to new field value. Existing fields are updated; missing fields are created.")),
		),
		handleBatchUpdateContent(client),
	)
}

// --- Content Data Handlers ---

func handleListContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		result, err := client.ContentData.ListPaginated(ctx, modulacms.PaginationParams{
			Limit: limit, Offset: offset,
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.ContentData.Get(ctx, modulacms.ContentID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := req.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError("status is required"), nil
		}
		params := modulacms.CreateContentDataParams{
			ParentID:   optionalIDPtr[modulacms.ContentID](req, "parent_id"),
			RouteID:    optionalIDPtr[modulacms.RouteID](req, "route_id"),
			DatatypeID: optionalIDPtr[modulacms.DatatypeID](req, "datatype_id"),
			AuthorID:   optionalIDPtr[modulacms.UserID](req, "author_id"),
			Status:     modulacms.ContentStatus(status),
		}
		result, err := client.ContentData.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		status, err := req.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError("status is required"), nil
		}
		params := modulacms.UpdateContentDataParams{
			ContentDataID: modulacms.ContentID(id),
			ParentID:      optionalIDPtr[modulacms.ContentID](req, "parent_id"),
			RouteID:       optionalIDPtr[modulacms.RouteID](req, "route_id"),
			DatatypeID:    optionalIDPtr[modulacms.DatatypeID](req, "datatype_id"),
			AuthorID:      optionalIDPtr[modulacms.UserID](req, "author_id"),
			Status:        modulacms.ContentStatus(status),
		}
		result, err := client.ContentData.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.ContentData.Delete(ctx, modulacms.ContentID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Content Delivery Handlers ---

func handleGetPage(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		format := req.GetString("format", "")
		result, err := client.Content.GetPage(ctx, slug, format)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleGetContentTree(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		format := req.GetString("format", "")
		result, err := client.AdminTree.Get(ctx, slug, format)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

// --- Content Field Handlers ---

func handleListContentFields(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		result, err := client.ContentFields.ListPaginated(ctx, modulacms.PaginationParams{
			Limit: limit, Offset: offset,
		})
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateContentField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cdID, err := req.RequireString("content_data_id")
		if err != nil {
			return mcp.NewToolResultError("content_data_id is required"), nil
		}
		fID, err := req.RequireString("field_id")
		if err != nil {
			return mcp.NewToolResultError("field_id is required"), nil
		}
		fVal, err := req.RequireString("field_value")
		if err != nil {
			return mcp.NewToolResultError("field_value is required"), nil
		}
		contentDataID := modulacms.ContentID(cdID)
		fieldID := modulacms.FieldID(fID)
		params := modulacms.CreateContentFieldParams{
			ContentDataID: &contentDataID,
			FieldID:       &fieldID,
			FieldValue:    fVal,
			RouteID:       optionalIDPtr[modulacms.RouteID](req, "route_id"),
			AuthorID:      optionalIDPtr[modulacms.UserID](req, "author_id"),
		}
		result, err := client.ContentFields.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateContentField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		fVal, err := req.RequireString("field_value")
		if err != nil {
			return mcp.NewToolResultError("field_value is required"), nil
		}

		cfID := modulacms.ContentFieldID(id)

		// Fetch existing record to preserve unmodified fields (PUT is full replacement)
		existing, err := client.ContentFields.Get(ctx, cfID)
		if err != nil {
			return errResult(err), nil
		}

		params := modulacms.UpdateContentFieldParams{
			ContentFieldID: cfID,
			ContentDataID:  existing.ContentDataID,
			FieldID:        existing.FieldID,
			FieldValue:     fVal,
			RouteID:        existing.RouteID,
			AuthorID:       existing.AuthorID,
		}

		// Override with explicitly provided values
		if v := optionalIDPtr[modulacms.ContentID](req, "content_data_id"); v != nil {
			params.ContentDataID = v
		}
		if v := optionalIDPtr[modulacms.FieldID](req, "field_id"); v != nil {
			params.FieldID = v
		}
		if v := optionalIDPtr[modulacms.RouteID](req, "route_id"); v != nil {
			params.RouteID = v
		}
		if v := optionalIDPtr[modulacms.UserID](req, "author_id"); v != nil {
			params.AuthorID = v
		}

		result, err := client.ContentFields.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteContentField(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.ContentFields.Delete(ctx, modulacms.ContentFieldID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Content Batch Handler ---

func handleBatchUpdateContent(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		cdID, err := req.RequireString("content_data_id")
		if err != nil {
			return mcp.NewToolResultError("content_data_id is required"), nil
		}
		body := map[string]any{
			"content_data_id": cdID,
		}
		args := req.GetArguments()
		if cd, ok := args["content_data"]; ok {
			body["content_data"] = cd
		}
		if f, ok := args["fields"]; ok {
			body["fields"] = f
		}
		result, err := client.ContentBatch.Update(ctx, body)
		if err != nil {
			return errResult(err), nil
		}
		// Pretty-print the raw JSON response
		var pretty json.RawMessage
		if jsonErr := json.Unmarshal(result, &pretty); jsonErr != nil {
			return mcp.NewToolResultText(string(result)), nil
		}
		formatted, formatErr := json.MarshalIndent(pretty, "", "  ")
		if formatErr != nil {
			return mcp.NewToolResultText(string(result)), nil
		}
		return mcp.NewToolResultText(string(formatted)), nil
	}
}
