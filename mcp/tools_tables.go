package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerTableTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("list_tables",
			mcp.WithDescription("List all CMS metadata tables."),
		),
		handleListTables(client),
	)

	srv.AddTool(
		mcp.NewTool("get_table",
			mcp.WithDescription("Get a single table by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Table ID (ULID)")),
		),
		handleGetTable(client),
	)

	srv.AddTool(
		mcp.NewTool("create_table",
			mcp.WithDescription("Create a new table record."),
			mcp.WithString("label", mcp.Required(), mcp.Description("Table label")),
		),
		handleCreateTable(client),
	)

	srv.AddTool(
		mcp.NewTool("update_table",
			mcp.WithDescription("Update a table record."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Table ID (ULID)")),
			mcp.WithString("label", mcp.Required(), mcp.Description("New table label")),
		),
		handleUpdateTable(client),
	)

	srv.AddTool(
		mcp.NewTool("delete_table",
			mcp.WithDescription("Delete a table by ID."),
			mcp.WithString("id", mcp.Required(), mcp.Description("Table ID (ULID)")),
		),
		handleDeleteTable(client),
	)
}

func handleListTables(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Tables.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetTable(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		result, err := client.Tables.Get(ctx, modulacms.TableID(id))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleCreateTable(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modulacms.CreateTableParams{Label: label}
		result, err := client.Tables.Create(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateTable(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		label, err := req.RequireString("label")
		if err != nil {
			return mcp.NewToolResultError("label is required"), nil
		}
		params := modulacms.UpdateTableParams{
			ID:    modulacms.TableID(id),
			Label: label,
		}
		result, err := client.Tables.Update(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeleteTable(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := req.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError("id is required"), nil
		}
		err = client.Tables.Delete(ctx, modulacms.TableID(id))
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("deleted"), nil
	}
}
