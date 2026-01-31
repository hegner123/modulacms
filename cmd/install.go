package main

import (
	"github.com/hegner123/modulacms/internal/install"
	"github.com/spf13/cobra"
)

var installYes bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run the installation wizard",
	Long:  "Run the interactive installation process. Use --yes to accept all defaults.",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		return install.RunInstall(&verbose, &installYes)
	},
}

func init() {
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "Accept all defaults (non-interactive)")
}
