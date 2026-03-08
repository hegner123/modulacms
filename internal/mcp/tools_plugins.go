package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	modula "github.com/hegner123/modulacms/sdks/go"
)

func registerPluginTools(srv *server.MCPServer, client *modula.Client) {
	srv.AddTool(
		mcp.NewTool("list_plugins",
			mcp.WithDescription("List all installed plugins with their status."),
		),
		handleListPlugins(client),
	)

	srv.AddTool(
		mcp.NewTool("get_plugin",
			mcp.WithDescription("Get detailed info for a specific plugin."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name")),
		),
		handleGetPlugin(client),
	)

	srv.AddTool(
		mcp.NewTool("reload_plugin",
			mcp.WithDescription("Reload a plugin from disk."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name")),
		),
		handleReloadPlugin(client),
	)

	srv.AddTool(
		mcp.NewTool("enable_plugin",
			mcp.WithDescription("Enable a disabled plugin."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name")),
		),
		handleEnablePlugin(client),
	)

	srv.AddTool(
		mcp.NewTool("disable_plugin",
			mcp.WithDescription("Disable an active plugin."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name")),
		),
		handleDisablePlugin(client),
	)

	srv.AddTool(
		mcp.NewTool("plugin_cleanup_dry_run",
			mcp.WithDescription("List orphaned plugin tables without dropping them."),
		),
		handlePluginCleanupDryRun(client),
	)

	srv.AddTool(
		mcp.NewTool("plugin_cleanup_drop",
			mcp.WithDescription("Drop orphaned plugin tables."),
			mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Must be true to confirm dropping tables")),
			mcp.WithObject("tables", mcp.Required(), mcp.Description("Array of table names to drop")),
		),
		handlePluginCleanupDrop(client),
	)

	srv.AddTool(
		mcp.NewTool("list_plugin_routes",
			mcp.WithDescription("List all plugin-registered HTTP routes with their approval status."),
		),
		handleListPluginRoutes(client),
	)

	srv.AddTool(
		mcp.NewTool("approve_plugin_routes",
			mcp.WithDescription("Approve one or more plugin routes. Each route requires plugin, method, and path."),
			mcp.WithObject("routes", mcp.Required(), mcp.Description("Array of {plugin, method, path} objects to approve")),
		),
		handleApprovePluginRoutes(client),
	)

	srv.AddTool(
		mcp.NewTool("revoke_plugin_routes",
			mcp.WithDescription("Revoke approval for one or more plugin routes."),
			mcp.WithObject("routes", mcp.Required(), mcp.Description("Array of {plugin, method, path} objects to revoke")),
		),
		handleRevokePluginRoutes(client),
	)

	srv.AddTool(
		mcp.NewTool("list_plugin_hooks",
			mcp.WithDescription("List all plugin-registered hooks with their approval status."),
		),
		handleListPluginHooks(client),
	)

	srv.AddTool(
		mcp.NewTool("approve_plugin_hooks",
			mcp.WithDescription("Approve one or more plugin hooks. Each hook requires plugin, event, and table."),
			mcp.WithObject("hooks", mcp.Required(), mcp.Description("Array of {plugin, event, table} objects to approve")),
		),
		handleApprovePluginHooks(client),
	)

	srv.AddTool(
		mcp.NewTool("revoke_plugin_hooks",
			mcp.WithDescription("Revoke approval for one or more plugin hooks."),
			mcp.WithObject("hooks", mcp.Required(), mcp.Description("Array of {plugin, event, table} objects to revoke")),
		),
		handleRevokePluginHooks(client),
	)
}

func handleListPlugins(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Plugins.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleGetPlugin(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		result, err := client.Plugins.Get(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleReloadPlugin(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		result, err := client.Plugins.Reload(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleEnablePlugin(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		result, err := client.Plugins.Enable(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleDisablePlugin(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		result, err := client.Plugins.Disable(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handlePluginCleanupDryRun(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.Plugins.CleanupDryRun(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handlePluginCleanupDrop(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		confirm := req.GetBool("confirm", false)
		args := req.GetArguments()
		var tables []string
		if raw, ok := args["tables"]; ok {
			if arr, ok := raw.([]any); ok {
				for _, item := range arr {
					if s, ok := item.(string); ok {
						tables = append(tables, s)
					}
				}
			}
		}
		params := modula.CleanupDropParams{
			Confirm: confirm,
			Tables:  tables,
		}
		result, err := client.Plugins.CleanupDrop(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func handleListPluginRoutes(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.PluginRoutes.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func parseRouteApprovalItems(req mcp.CallToolRequest) ([]modula.RouteApprovalItem, error) {
	args := req.GetArguments()
	raw, ok := args["routes"]
	if !ok {
		return nil, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var items []modula.RouteApprovalItem
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func handleApprovePluginRoutes(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		routes, err := parseRouteApprovalItems(req)
		if err != nil {
			return mcp.NewToolResultError("invalid routes: " + err.Error()), nil
		}
		if len(routes) == 0 {
			return mcp.NewToolResultError("routes array is required"), nil
		}
		err = client.PluginRoutes.Approve(ctx, routes)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("approved"), nil
	}
}

func handleRevokePluginRoutes(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		routes, err := parseRouteApprovalItems(req)
		if err != nil {
			return mcp.NewToolResultError("invalid routes: " + err.Error()), nil
		}
		if len(routes) == 0 {
			return mcp.NewToolResultError("routes array is required"), nil
		}
		err = client.PluginRoutes.Revoke(ctx, routes)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("revoked"), nil
	}
}

func handleListPluginHooks(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		result, err := client.PluginHooks.List(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return jsonResult(result)
	}
}

func parseHookApprovalItems(req mcp.CallToolRequest) ([]modula.HookApprovalItem, error) {
	args := req.GetArguments()
	raw, ok := args["hooks"]
	if !ok {
		return nil, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var items []modula.HookApprovalItem
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func handleApprovePluginHooks(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		hooks, err := parseHookApprovalItems(req)
		if err != nil {
			return mcp.NewToolResultError("invalid hooks: " + err.Error()), nil
		}
		if len(hooks) == 0 {
			return mcp.NewToolResultError("hooks array is required"), nil
		}
		err = client.PluginHooks.Approve(ctx, hooks)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("approved"), nil
	}
}

func handleRevokePluginHooks(client *modula.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		hooks, err := parseHookApprovalItems(req)
		if err != nil {
			return mcp.NewToolResultError("invalid hooks: " + err.Error()), nil
		}
		if len(hooks) == 0 {
			return mcp.NewToolResultError("hooks array is required"), nil
		}
		err = client.PluginHooks.Revoke(ctx, hooks)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("revoked"), nil
	}
}
