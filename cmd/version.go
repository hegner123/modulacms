package main

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	Long: `Print the ModulaCMS version, build commit, and build date, then exit.

Output includes the semantic version, git commit hash, and compilation timestamp
embedded at build time via ldflags.

Examples:
  modula version`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		_, err := fmt.Fprintln(cmd.OutOrStdout(), utility.GetFullVersionInfo())
		return err
	},
}
