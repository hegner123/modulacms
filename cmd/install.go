package main

import (
	"github.com/hegner123/modulacms/internal/install"
	"github.com/spf13/cobra"
)

var installYes bool
var installAdminPassword string

// installCmd provides the installation wizard command for setting up Modula.
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run the installation wizard",
	Long: `Run the first-time setup wizard to create config.json, initialize the database,
and seed bootstrap data (admin user, default roles, permissions).

In interactive mode, prompts for database driver, connection details, ports,
and admin credentials. Use --yes to accept all defaults non-interactively
(requires --admin-password).

Flags:
  --yes, -y          Accept all defaults without prompting
  --admin-password   System admin password (required when using --yes)

Examples:
  modula install                                  # interactive wizard
  modula install --yes --admin-password s3cret!    # non-interactive with defaults`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		return install.RunInstall(&verbose, &installYes, &installAdminPassword)
	},
}

func init() {
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "Accept all defaults (non-interactive)")
	installCmd.Flags().StringVar(&installAdminPassword, "admin-password", "", "System admin password (required for --yes mode)")
}
