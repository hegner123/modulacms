package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerPluginTools(srv *server.MCPServer, backend PluginBackend) {
	srv.AddTool(
		mcp.NewTool("list_plugins",
			mcp.WithDescription("List all installed plugins with their status."),
		),
		handleListPlugins(backend),
	)

	srv.AddTool(
		mcp.NewTool("get_plugin",
			mcp.WithDescription("Get detailed info for a specific plugin."),
			mcp.WithString("name", mcp.Required(), mcp.Description("plugin name")),
		),
		handleGetPlugin(backend),
	)

	srv.AddTool(
		mcp.NewTool("reload_plugin",
			mcp.WithDescription("Reload a plugin from disk."),
			mcp.WithString("name", mcp.Required(), mcp.Description("plugin name")),
		),
		handleReloadPlugin(backend),
	)

	srv.AddTool(
		mcp.NewTool("enable_plugin",
			mcp.WithDescription("Enable a disabled plugin."),
			mcp.WithString("name", mcp.Required(), mcp.Description("plugin name")),
		),
		handleEnablePlugin(backend),
	)

	srv.AddTool(
		mcp.NewTool("disable_plugin",
			mcp.WithDescription("Disable an active plugin."),
			mcp.WithString("name", mcp.Required(), mcp.Description("plugin name")),
		),
		handleDisablePlugin(backend),
	)

	srv.AddTool(
		mcp.NewTool("plugin_cleanup_dry_run",
			mcp.WithDescription("List orphaned plugin tables without dropping them."),
		),
		handlePluginCleanupDryRun(backend),
	)

	srv.AddTool(
		mcp.NewTool("plugin_cleanup_drop",
			mcp.WithDescription("Drop orphaned plugin tables."),
			mcp.WithBoolean("confirm", mcp.Required(), mcp.Description("Must be true to confirm dropping tables")),
			mcp.WithObject("tables", mcp.Required(), mcp.Description("Array of table names to drop")),
		),
		handlePluginCleanupDrop(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_plugin_routes",
			mcp.WithDescription("List all plugin-registered HTTP routes with their approval status."),
		),
		handleListPluginRoutes(backend),
	)

	srv.AddTool(
		mcp.NewTool("approve_plugin_routes",
			mcp.WithDescription("Approve one or more plugin routes. Each route requires plugin, method, and path."),
			mcp.WithObject("routes", mcp.Required(), mcp.Description("Array of {plugin, method, path} objects to approve")),
		),
		handleApprovePluginRoutes(backend),
	)

	srv.AddTool(
		mcp.NewTool("revoke_plugin_routes",
			mcp.WithDescription("Revoke approval for one or more plugin routes."),
			mcp.WithObject("routes", mcp.Required(), mcp.Description("Array of {plugin, method, path} objects to revoke")),
		),
		handleRevokePluginRoutes(backend),
	)

	srv.AddTool(
		mcp.NewTool("list_plugin_hooks",
			mcp.WithDescription("List all plugin-registered hooks with their approval status."),
		),
		handleListPluginHooks(backend),
	)

	srv.AddTool(
		mcp.NewTool("approve_plugin_hooks",
			mcp.WithDescription("Approve one or more plugin hooks. Each hook requires plugin, event, and table."),
			mcp.WithObject("hooks", mcp.Required(), mcp.Description("Array of {plugin, event, table} objects to approve")),
		),
		handleApprovePluginHooks(backend),
	)

	srv.AddTool(
		mcp.NewTool("revoke_plugin_hooks",
			mcp.WithDescription("Revoke approval for one or more plugin hooks."),
			mcp.WithObject("hooks", mcp.Required(), mcp.Description("Array of {plugin, event, table} objects to revoke")),
		),
		handleRevokePluginHooks(backend),
	)
}

func handleListPlugins(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListPlugins(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleGetPlugin(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		data, err := backend.GetPlugin(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleReloadPlugin(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		data, err := backend.ReloadPlugin(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleEnablePlugin(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		data, err := backend.EnablePlugin(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleDisablePlugin(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		data, err := backend.DisablePlugin(ctx, name)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handlePluginCleanupDryRun(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.PluginCleanupDryRun(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handlePluginCleanupDrop(backend PluginBackend) server.ToolHandlerFunc {
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
		params, err := marshalParams(map[string]any{
			"confirm": confirm,
			"tables":  tables,
		})
		if err != nil {
			return nil, err
		}
		data, err := backend.PluginCleanupDrop(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleListPluginRoutes(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListPluginRoutes(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func marshalApprovalItems(req mcp.CallToolRequest, key string) (json.RawMessage, error) {
	args := req.GetArguments()
	raw, ok := args[key]
	if !ok {
		return nil, nil
	}
	var b []byte
	if s, ok := raw.(string); ok {
		b = []byte(s)
	} else {
		var err error
		b, err = json.Marshal(raw)
		if err != nil {
			return nil, err
		}
	}
	return json.RawMessage(b), nil
}

func handleApprovePluginRoutes(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := marshalApprovalItems(req, "routes")
		if err != nil {
			return mcp.NewToolResultError("invalid routes: " + err.Error()), nil
		}
		if len(params) == 0 {
			return mcp.NewToolResultError("routes array is required"), nil
		}
		err = backend.ApprovePluginRoutes(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("approved"), nil
	}
}

func handleRevokePluginRoutes(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := marshalApprovalItems(req, "routes")
		if err != nil {
			return mcp.NewToolResultError("invalid routes: " + err.Error()), nil
		}
		if len(params) == 0 {
			return mcp.NewToolResultError("routes array is required"), nil
		}
		err = backend.RevokePluginRoutes(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("revoked"), nil
	}
}

func handleListPluginHooks(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		data, err := backend.ListPluginHooks(ctx)
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleApprovePluginHooks(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := marshalApprovalItems(req, "hooks")
		if err != nil {
			return mcp.NewToolResultError("invalid hooks: " + err.Error()), nil
		}
		if len(params) == 0 {
			return mcp.NewToolResultError("hooks array is required"), nil
		}
		err = backend.ApprovePluginHooks(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("approved"), nil
	}
}

func handleRevokePluginHooks(backend PluginBackend) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		params, err := marshalApprovalItems(req, "hooks")
		if err != nil {
			return mcp.NewToolResultError("invalid hooks: " + err.Error()), nil
		}
		if len(params) == 0 {
			return mcp.NewToolResultError("hooks array is required"), nil
		}
		err = backend.RevokePluginHooks(ctx, params)
		if err != nil {
			return errResult(err), nil
		}
		return mcp.NewToolResultText("revoked"), nil
	}
}
