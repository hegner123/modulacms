package main

import (
	"github.com/hegner123/modulacms/internal/install"
	"github.com/spf13/cobra"
)

var installYes bool
var installAdminPassword string

// installCmd provides the installation wizard command for setting up ModulaCMS.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run the installation wizard",
	Long:  "Run the interactive installation process. Use --yes to accept all defaults.",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		return install.RunInstall(&verbose, &installYes, &installAdminPassword)
	},
}

func init() {
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "Accept all defaults (non-interactive)")
	installCmd.Flags().StringVar(&installAdminPassword, "admin-password", "", "System admin password (required for --yes mode)")
}
