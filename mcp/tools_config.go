package main

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modulacms "github.com/hegner123/modulacms/sdks/go"
)

func registerConfigTools(srv *server.MCPServer, client *modulacms.Client) {
	srv.AddTool(
		mcp.NewTool("get_config",
			mcp.WithDescription("Get the current server configuration (sensitive values are redacted). Optionally filter by category."),
			mcp.WithString("category", mcp.Description("Filter config fields by category")),
		),
		handleGetConfig(client),
	)

	srv.AddTool(
		mcp.NewTool("get_config_meta",
			mcp.WithDescription("Get config field metadata: json_key, label, category, hot_reloadable, sensitive, required, description. Call this first to discover what config keys exist before using update_config."),
		),
		handleGetConfigMeta(client),
	)

	srv.AddTool(
		mcp.NewTool("update_config",
			mcp.WithDescription("Update server configuration. Pass a JSON object of key-value pairs. Use get_config_meta first to discover valid keys. Non-hot-reloadable changes may require a server restart."),
			mcp.WithObject("updates", mcp.Required(), mcp.Description("JSON object of config key-value pairs to update")),
		),
		handleUpdateConfig(client),
	)
}

func handleGetConfig(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		category := req.GetString("category", "")
		result, err := client.Config.Get(ctx, category)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetConfigMeta(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Config.Meta(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleUpdateConfig(client *modulacms.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		updates, ok := args["updates"].(map[string]any)
		if !ok {
			return mcp.NewToolResultError("updates must be a JSON object of key-value pairs"), nil
		}
		result, err := client.Config.Update(ctx, updates)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}
