package main

import (
	"github.com/spf13/cobra"

	mcpserver "github.com/hegner123/modulacms/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp [project] [environment]",
	Short: "Start the MCP server over stdio",
	Long: `Start the Model Context Protocol server for AI-assisted content management.

The MCP server loads the project registry (~/.modula/configs.json) and connects
to a running Modula instance via the Go SDK, exposing 40+ tools for content
CRUD, schema management, media, users, roles, permissions, configuration, and
import.

Connection is resolved from the project registry:

  modula mcp                         Auto-detect project from working directory
  modula mcp mysite                  Connect to project "mysite" (default env)
  modula mcp mysite production       Connect to "mysite" production environment

The URL and API key are read from the project's modula.config.json (port field
for URL, mcp_api_key for authentication). When overlays are configured, the
overlay values take precedence.

Three connection management tools are always available:
  list_projects    — list all registered projects and environments
  switch_project   — change the active CMS connection at runtime
  get_connection   — show the current project, environment, and URL`,
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var project, env string
		if len(args) > 0 {
			project = args[0]
		}
		if len(args) > 1 {
			env = args[1]
		}

		return mcpserver.ServeWithRegistry(project, env)
	},
}
