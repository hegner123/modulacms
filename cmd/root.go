package main

import (
	"github.com/spf13/cobra"
)

var (
	cfgPath string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "modulacms",
	Short: "ModulaCMS - A headless CMS written in Go",
	Long:  "ModulaCMS serves content over HTTP/HTTPS and provides SSH access for backend management.",
	// Default action (no subcommand) starts the server
	RunE: func(cmd *cobra.Command, args []string) error {
		return serveCmd.RunE(cmd, args)
	},
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
}

func Execute() error {
	return rootCmd.Execute()
}
