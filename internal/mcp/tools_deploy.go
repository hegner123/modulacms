package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerDeployTools(srv *server.MCPServer, backend DeployBackend) {
	srv.AddTool(
		mcp.NewTool("deploy_health",
			mcp.WithDescription("Check deploy sync health status. Returns status, version, and node ID."),
		),
		handleDeployHealth(backend),
	)

	srv.AddTool(
		mcp.NewTool("deploy_export",
			mcp.WithDescription("Export a sync payload from the server. Optionally filter by table names."),
			mcp.WithObject("tables", mcp.Description("Array of table names to export. Omit for full export.")),
		),
		handleDeployExport(backend),
	)

	srv.AddTool(
		mcp.NewTool("deploy_import",
			mcp.WithDescription("Import a sync payload into the server. The payload should be the raw JSON from deploy_export."),
			mcp.WithObject("payload", mcp.Required(), mcp.Description("Sync payload JSON from deploy_export")),
		),
		handleDeployImport(backend),
	)

	srv.AddTool(
		mcp.NewTool("deploy_dry_run",
			mcp.WithDescription("Dry-run import: shows what would change without writing. Same payload format as deploy_import."),
			mcp.WithObject("payload", mcp.Required(), mcp.Description("Sync payload JSON from deploy_export")),
		),
		handleDeployDryRun(backend),
	)
}

func handleDeployHealth(backend DeployBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.DeployHealth(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeployExport(backend DeployBackend) server.ToolHandlerFunc {
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
		data, err := backend.DeployExport(ctx, tables)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeployImport(backend DeployBackend) server.ToolHandlerFunc {
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
		data, err := backend.DeployImport(ctx, json.RawMessage(b))
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDeployDryRun(backend DeployBackend) server.ToolHandlerFunc {
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
		data, err := backend.DeployDryRun(ctx, json.RawMessage(b))
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}
