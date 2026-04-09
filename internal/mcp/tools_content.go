package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerContentTools(srv *server.MCPServer, backend ContentBackend) {
	// --- Content Data CRUD ---

	srv.AddTool(
		mcp.NewTool("list_content",
			mcp.WithDescription("List content data entries with pagination. Returns structural metadata (IDs, status, timestamps, tree pointers) without field values. Use get_content_tree or get_page to see assembled content with field values."),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleListContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_content",
			mcp.WithDescription("Get a single content data entry by ID. Returns structural metadata without field values."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
		),
		handleGetContent(backend),
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
		handleCreateContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_content",
			mcp.WithDescription("update an existing content data entry. This is a full replacement — all fields are sent. Omitted pointer fields (parent_id, route_id, etc.) will be set to null. Status values: draft, published, archived, pending."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
			mcp.WithString("status", mcp.Required(), mcp.Description("Content status: draft, published, archived, pending"), mcp.Enum("draft", "published", "archived", "pending")),
			mcp.WithString("parent_id", mcp.Description("Parent content data ID")),
			mcp.WithString("route_id", mcp.Description("Route ID")),
			mcp.WithString("datatype_id", mcp.Description("Datatype ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_content",
			mcp.WithDescription("Delete a content data entry by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content data ID (ULID)")),
		),
		handleDeleteContent(backend),
	)

	// --- Content Delivery ---

	srv.AddTool(
		mcp.NewTool("get_page",
			mcp.WithDescription("Get assembled page content by slug via the public content delivery endpoint. Returns the full content tree with field values in the requested format."),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug of the page (e.g. 'about' or 'blog/my-post')")),
			mcp.WithString("format", mcp.Description("Response format: contentful, sanity, strapi, wordpress, clean, raw (default: server default)"), mcp.Enum("contentful", "sanity", "strapi", "wordpress", "clean", "raw")),
			mcp.WithString("locale", mcp.Description("Locale code for localized content (e.g. 'en', 'de', 'fr')")),
		),
		handleGetPage(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_content_tree",
			mcp.WithDescription("Get the admin content tree for a slug. Returns the content hierarchy with field values. To get the tree for a known content item, first use get_content to find its route, then use the route's slug here."),
			mcp.WithString("slug", mcp.Required(), mcp.Description("URL slug to retrieve the tree for")),
			mcp.WithString("format", mcp.Description("Response format: contentful, sanity, strapi, wordpress, clean, raw (default: server default)"), mcp.Enum("contentful", "sanity", "strapi", "wordpress", "clean", "raw")),
		),
		handleGetContentTree(backend),
	)

	// --- Content Fields ---

	srv.AddTool(
		mcp.NewTool("list_content_fields",
			mcp.WithDescription("List content field records with pagination. Returns ALL content fields across all content items (cannot filter by content item). Use get_content_tree or get_page to see fields for a specific content item."),
			mcp.WithNumber("limit", mcp.Description("Max items to return (default 20, max 1000)"), mcp.DefaultNumber(20)),
			mcp.WithNumber("offset", mcp.Description("Number of items to skip (default 0)"), mcp.DefaultNumber(0)),
		),
		handleListContentFields(backend),
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
		handleCreateContentField(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_content_field",
			mcp.WithDescription("update an existing content field value."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content field ID (ULID)")),
			mcp.WithString("field_value", mcp.Required(), mcp.Description("The new field value")),
			mcp.WithString("content_data_id", mcp.Description("Content data ID")),
			mcp.WithString("field_id", mcp.Description("Field definition ID")),
			mcp.WithString("route_id", mcp.Description("Route ID")),
			mcp.WithString("author_id", mcp.Description("Author user ID")),
		),
		handleUpdateContentField(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_content_field",
			mcp.WithDescription("Delete a content field record by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content field ID (ULID)")),
		),
		handleDeleteContentField(backend),
	)

	// --- Content Field Get ---

	srv.AddTool(
		mcp.NewTool("get_content_field",
			mcp.WithDescription("Get a single content field by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Content field ID (ULID)")),
		),
		handleGetContentField(backend),
	)

	// --- Content Reorder ---

	srv.AddTool(
		mcp.NewTool("reorder_content",
			mcp.WithDescription("Atomically reorder sibling content data nodes under a parent. Provide the parent_id and an ordered list of content IDs."),
			mcp.WithString("parent_id", mcp.Description("Parent content ID (null for root-level siblings)")),
			mcp.WithObject("ordered_ids", mcp.Required(), mcp.Description("Array of content data IDs in desired order")),
		),
		handleReorderContent(backend),
	)

	srv.AddTool(
		mcp.NewTool("move_content",
			mcp.WithDescription("Move a content data node to a new parent at a given position."),
			mcp.WithString("node_id", mcp.Required(), mcp.Description("Content data ID to move")),
			mcp.WithString("new_parent_id", mcp.Description("New parent content ID (null for root)")),
			mcp.WithNumber("position", mcp.Required(), mcp.Description("Zero-based position among siblings")),
		),
		handleMoveContent(backend),
	)

	// --- Content Tree ---

	srv.AddTool(
		mcp.NewTool("save_content_tree",
			mcp.WithDescription("Atomically apply tree structure changes (creates, deletes, pointer updates) in a single request. This is the preferred way to persist block editor state."),
			mcp.WithObject("request", mcp.Required(), mcp.Description("TreeSaveRequest JSON: content_id (required), creates (array), updates (array), deletes (array of IDs)")),
		),
		handleSaveContentTree(backend),
	)

	// --- Content Heal ---

	srv.AddTool(
		mcp.NewTool("heal_content",
			mcp.WithDescription("Scan content_data and content_field rows for malformed IDs and repair them. Use dry_run=true to preview without writing."),
			mcp.WithBoolean("dry_run", mcp.Description("If true, preview repairs without writing changes (default false)")),
		),
		handleHealContent(backend),
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
		handleBatchUpdateContent(backend),
	)
}

// --- Content Data Handlers ---

func handleListContent(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.ListContent(ctx, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetContent(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetContent(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateContent(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status, err := req.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError("status is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"parent_id":   optionalStrPtr(req, "parent_id"),
			"route_id":    optionalStrPtr(req, "route_id"),
			"datatype_id": optionalStrPtr(req, "datatype_id"),
			"author_id":   optionalStrPtr(req, "author_id"),
			"status":      status,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateContent(backend ContentBackend) server.ToolHandlerFunc {
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
			"content_data_id": id,
			"parent_id":       optionalStrPtr(req, "parent_id"),
			"route_id":        optionalStrPtr(req, "route_id"),
			"datatype_id":     optionalStrPtr(req, "datatype_id"),
			"author_id":       optionalStrPtr(req, "author_id"),
			"status":          status,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteContent(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteContent(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Content Delivery Handlers ---

func handleGetPage(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		format := req.GetString("format", "")
		locale := req.GetString("locale", "")
		data, err := backend.GetPage(ctx, slug, format, locale)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetContentTree(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		slug, err := req.RequireString("slug")
		if err != nil {
			return mcp.NewToolResultError("slug is required"), nil
		}
		// The admin tree endpoint expects the slug as a path segment, not prefixed with "/".
		slug = strings.TrimPrefix(slug, "/")
		if slug == "" {
			return mcp.NewToolResultError("slug must not be empty (use the route slug without leading '/')"), nil
		}
		format := req.GetString("format", "")
		data, err := backend.GetContentTree(ctx, slug, format)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Content Field Handlers ---

func handleListContentFields(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		limit := int64(req.GetFloat("limit", 20))
		offset := int64(req.GetFloat("offset", 0))
		data, err := backend.ListContentFields(ctx, limit, offset)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateContentField(backend ContentBackend) server.ToolHandlerFunc {
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
		params, err := marshalParams(map[string]any{
			"content_data_id": cdID,
			"field_id":        fID,
			"field_value":     fVal,
			"route_id":        optionalStrPtr(req, "route_id"),
			"author_id":       optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateContentField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateContentField(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		fVal, err := req.RequireString("field_value")
		if err != nil {
			return mcp.NewToolResultError("field_value is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"content_field_id": id,
			"content_data_id":  optionalStrPtr(req, "content_data_id"),
			"field_id":         optionalStrPtr(req, "field_id"),
			"field_value":      fVal,
			"route_id":         optionalStrPtr(req, "route_id"),
			"author_id":        optionalStrPtr(req, "author_id"),
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateContentField(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteContentField(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteContentField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}

// --- Content Batch Handler ---

func handleBatchUpdateContent(backend ContentBackend) server.ToolHandlerFunc {
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
		params, err := marshalParams(body)
		if err != nil {
			return nil, err
		}
		data, err := backend.BatchUpdateContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Content Field Get Handler ---

func handleGetContentField(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetContentField(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Content Reorder Handlers ---

func handleReorderContent(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		rawIDs, ok := args["ordered_ids"].([]any)
		if !ok {
			if s, sOk := args["ordered_ids"].(string); sOk {
				var parsed []string
				if err := json.Unmarshal([]byte(s), &parsed); err != nil {
					return mcp.NewToolResultError("ordered_ids must be a JSON array of content IDs"), nil
				}
				rawIDs = make([]any, len(parsed))
				for i, v := range parsed {
					rawIDs[i] = v
				}
			}
		}
		if len(rawIDs) == 0 {
			return mcp.NewToolResultError("ordered_ids must be a non-empty array of content IDs"), nil
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
		data, err := backend.ReorderContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleMoveContent(backend ContentBackend) server.ToolHandlerFunc {
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
		data, err := backend.MoveContent(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Content Tree Handler ---

func handleSaveContentTree(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		rawReq, ok := args["request"]
		if !ok {
			return mcp.NewToolResultError("request is required"), nil
		}
		b, err := json.Marshal(rawReq)
		if err != nil {
			return mcp.NewToolResultError("invalid request object"), nil
		}
		data, err := backend.SaveContentTree(ctx, json.RawMessage(b))
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

// --- Content Heal Handler ---

func handleHealContent(backend ContentBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		dryRun := req.GetBool("dry_run", false)
		data, err := backend.HealContent(ctx, dryRun)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
