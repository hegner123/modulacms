package main

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version and exit",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		_, err := fmt.Fprintln(cmd.OutOrStdout(), utility.GetFullVersionInfo())
		return err
	},
}
