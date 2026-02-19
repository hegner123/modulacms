package main

import (
	"github.com/spf13/cobra"
)

var (
	cfgPath string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "modula",
	Short: "Modula - A headless CMS written in Go",
	Long:  "Modula serves content over HTTP/HTTPS and provides SSH access for backend management.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "config.json", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(certCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(configParentCmd)
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(pluginCmd)
}

// Execute runs the root CLI command and returns any error encountered.
func Execute() error {
	return rootCmd.Execute()
}
