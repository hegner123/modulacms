package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTableTools(srv *server.MCPServer, backend TableBackend) {
	srv.AddTool(
		mcp.NewTool("list_tables",
			mcp.WithDescription("List all CMS metadata tables."),
		),
		handleListTables(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_table",
			mcp.WithDescription("Get a single table by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Table ID (ULID)")),
		),
		handleGetTable(backend),
	)

	srv.AddTool(
		mcp.NewTool("create_table",
			mcp.WithDescription("Create a new table record."),
			mcp.WithString("label", mcp.Required(), mcp.Description("Table label")),
		),
		handleCreateTable(backend),
	)

	srv.AddTool(
		mcp.NewTool("update_table",
			mcp.WithDescription("update a table record."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Table ID (ULID)")),
			mcp.WithString("label", mcp.Required(), mcp.Description("New table label")),
		),
		handleUpdateTable(backend),
	)

	srv.AddTool(
		mcp.NewTool("delete_table",
			mcp.WithDescription("Delete a table by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Table ID (ULID)")),
		),
		handleDeleteTable(backend),
	)
}

func handleListTables(backend TableBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListTables(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetTable(backend TableBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		data, err := backend.GetTable(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleCreateTable(backend TableBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"label": label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.CreateTable(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleUpdateTable(backend TableBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params, err := marshalParams(map[string]any{
			"id":    id,
			"label": label,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.UpdateTable(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeleteTable(backend TableBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = backend.DeleteTable(ctx, id)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
