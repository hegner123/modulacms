package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	mcpserver "github.com/hegner123/modulacms/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server over stdio",
	Long: `Start the Model Context Protocol server for AI-assisted content management.

The MCP server connects to a running Modula instance via the Go SDK and exposes
40+ tools for content CRUD, schema management, media, users, roles, permissions,
configuration, and import.

Connection is configured via environment variables:

  MODULA_URL      Base URL of the Modula server (e.g. http://localhost:8080)
  MODULA_API_KEY  API key for authentication

Or via flags:

  --url       Base URL of the Modula server
  --api-key   API key for authentication

Flags take precedence over environment variables.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		url, _ := cmd.Flags().GetString("url")
		apiKey, _ := cmd.Flags().GetString("api-key")

		if url == "" {
			url = os.Getenv("MODULA_URL")
		}
		if apiKey == "" {
			apiKey = os.Getenv("MODULA_API_KEY")
		}

		if url == "" || apiKey == "" {
			fmt.Fprintln(os.Stderr, "MODULA_URL and MODULA_API_KEY are required (via flags or environment variables)")
			os.Exit(1)
		}

		return mcpserver.Serve(url, apiKey)
	},
}

func init() {
	mcpCmd.Flags().String("url", "", "Base URL of the Modula server")
	mcpCmd.Flags().String("api-key", "", "API key for authentication")
}
