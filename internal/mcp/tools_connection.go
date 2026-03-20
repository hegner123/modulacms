package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerConnectionTools(srv *server.MCPServer, cm *ConnectionManager) {
	srv.AddTool(
		mcp.NewTool("list_projects",
			mcp.WithDescription("List all registered projects and their environments from ~/.modula/configs.json. Shows which project is currently active."),
		),
		handleListProjects(cm),
	)

	srv.AddTool(
		mcp.NewTool("switch_project",
			mcp.WithDescription("Switch the active CMS connection to a different project and environment. Reads the project's config to get the URL and MCP API key."),
			mcp.WithString("project", mcp.Required(), mcp.Description("Project name from the registry")),
			mcp.WithString("environment", mcp.Description("Environment name (uses project default if omitted)")),
		),
		handleSwitchProject(cm),
	)

	srv.AddTool(
		mcp.NewTool("get_connection",
			mcp.WithDescription("Show the currently active project, environment, and CMS URL."),
		),
		handleGetConnection(cm),
	)
}

func handleListProjects(cm *ConnectionManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Reload registry to pick up changes made by other processes.
		if err := cm.ReloadRegistry(); err != nil {
			return errResult(err), nil
		}

		data, err := cm.ListProjects()
		if err != nil {
			return errResult(err), nil
		}
		return rawJSONResult(data), nil
	}
}

func handleSwitchProject(cm *ConnectionManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		project := req.GetString("project", "")
		env := req.GetString("environment", "")

		if project == "" {
			return errResult(fmt.Errorf("project name is required")), nil
		}

		// Reload registry before switching to pick up recent changes.
		if err := cm.ReloadRegistry(); err != nil {
			return errResult(err), nil
		}

		if err := cm.SwitchProject(project, env); err != nil {
			return errResult(err), nil
		}

		proj, envName, url := cm.ActiveConnection()
		result := map[string]string{
			"status":      "connected",
			"project":     proj,
			"environment": envName,
			"url":         url,
		}
		data, _ := json.Marshal(result)
		return rawJSONResult(data), nil
	}
}

func handleGetConnection(cm *ConnectionManager) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		proj, env, url := cm.ActiveConnection()
		connected := cm.Client() != nil

		result := map[string]any{
			"connected":   connected,
			"project":     proj,
			"environment": env,
			"url":         url,
		}
		data, _ := json.Marshal(result)
		return rawJSONResult(data), nil
	}
}
