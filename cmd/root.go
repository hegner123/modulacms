package main

import (
	"github.com/hegner123/modulacms/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgPath     string
	overlayPath string
	verbose     bool
	yesFlag     bool
)

var rootCmd = &cobra.Command{
	Use:   "modula",
	Short: "Modula - A headless CMS written in Go",
	Long: `Modula is a headless content management system that runs as a single binary.

It serves content over HTTP/HTTPS and provides an SSH-accessible terminal UI for
backend management. Content is managed via the SSH TUI, a web admin panel, or the
REST API and delivered to frontend clients over HTTP/HTTPS.

Core commands:
  init       Initialize a project (idempotent, safe to re-run)
  serve      Start all servers (HTTP, HTTPS, SSH)
  status     Show project status for the current directory
  connect    Launch the TUI for a remote or registered project
  tui        Launch the TUI locally without starting the server

Management commands:
  db         Initialize, wipe, reset, or export the database
  backup     Create, restore, or list backups
  config     Show, validate, or update configuration
  cert       Generate self-signed SSL certificates
  deploy     Export, import, and sync content between environments
  plugin     Manage Lua plugins (install, enable, reload, approve)
  pipeline   View and manage plugin pipeline entries
  scaffold   Generate deployment files (Dockerfile, docker-compose)
  mcp        Start the MCP server for AI-assisted content management
  update     Check for and apply binary updates

Global flags:
  --config   Path to modula.config.json (default: ./modula.config.json)
  --overlay  Overlay config file (merged on top of --config)
  --verbose  Enable debug-level log output
  --yes, -y  Auto-accept all prompts (non-interactive mode)`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", config.DefaultConfigFilename, "Path to configuration file")
	rootCmd.PersistentFlags().StringVar(&overlayPath, "overlay", "", "Overlay config file (merged on top of --config)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Auto-accept all prompts")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(certCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(configParentCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(pluginCmd)
	rootCmd.AddCommand(pipelineCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(scaffoldCmd)
}

// Execute runs the root CLI command and returns any error encountered.
func Execute() error {
	return rootCmd.Execute()
}
