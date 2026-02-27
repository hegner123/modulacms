package main

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerDeployTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("deploy_health",
			mcp.WithDescription("Check deploy sync health status. Returns status, version, and node ID."),
		),
		handleDeployHealth(client),
	)

	srv.AddTool(
		mcp.NewTool("deploy_export",
			mcp.WithDescription("Export a sync payload from the server. Optionally filter by table names."),
			mcp.WithObject("tables", mcp.Description("Array of table names to export. Omit for full export.")),
		),
		handleDeployExport(client),
	)

	srv.AddTool(
		mcp.NewTool("deploy_import",
			mcp.WithDescription("Import a sync payload into the server. The payload should be the raw JSON from deploy_export."),
			mcp.WithObject("payload", mcp.Required(), mcp.Description("Sync payload JSON from deploy_export")),
		),
		handleDeployImport(client),
	)

	srv.AddTool(
		mcp.NewTool("deploy_dry_run",
			mcp.WithDescription("Dry-run import: shows what would change without writing. Same payload format as deploy_import."),
			mcp.WithObject("payload", mcp.Required(), mcp.Description("Sync payload JSON from deploy_export")),
		),
		handleDeployDryRun(client),
	)
}

func handleDeployHealth(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Deploy.Health(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeployExport(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var tables []string
		args := req.GetArguments()
		if raw, ok := args["tables"]; ok {
			if arr, ok := raw.([]any); ok {
				for _, item := range arr {
					if s, ok := item.(string); ok {
						tables = append(tables, s)
					}
				}
			}
		}
		result, err := client.Deploy.Export(ctx, tables)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText(string(result)), nil
	}
}

func handleDeployImport(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		rawPayload, ok := args["payload"]
		if !ok {
			return mcp.NewToolResultError("payload is required"), nil
		}
		b, err := json.Marshal(rawPayload)
		if err != nil {
			return mcp.NewToolResultError("invalid payload"), nil
		}
		result, err := client.Deploy.Import(ctx, json.RawMessage(b))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDeployDryRun(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		rawPayload, ok := args["payload"]
		if !ok {
			return mcp.NewToolResultError("payload is required"), nil
		}
		b, err := json.Marshal(rawPayload)
		if err != nil {
			return mcp.NewToolResultError("invalid payload"), nil
		}
		result, err := client.Deploy.DryRunImport(ctx, json.RawMessage(b))
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}
