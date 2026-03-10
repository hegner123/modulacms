package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/registry"
	"github.com/spf13/cobra"
)

var initYes bool
var initAdminPassword string
var initProjectName string

// initCmd sets up a new Modula project and registers it in ~/.modula/configs.json.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project and register it",
	Long: `Initialize a new Modula project: create modula.config.json, set up the database,
seed bootstrap data, and register the project in ~/.modula/configs.json.

This is the same as 'install' but also registers the project in the global
registry so it can be accessed via 'modula connect'.

The project name defaults to the current directory name. The environment
is registered as "local" pointing to the modula.config.json in the current directory.

Flags:
  --yes, -y          Accept all defaults without prompting
  --admin-password   System admin password (required when using --yes)
  --name             Project name (default: current directory name)

Examples:
  modula init                                       # interactive wizard
  modula init --yes --admin-password s3cret!        # non-interactive with defaults
  modula init --name my-site --admin-password pw    # custom project name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		if err := install.RunInstall(&verbose, &initYes, &initAdminPassword); err != nil {
			return err
		}

		return registerProject()
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "Accept all defaults (non-interactive)")
	initCmd.Flags().StringVar(&initAdminPassword, "admin-password", "", "System admin password (required for --yes mode)")
	initCmd.Flags().StringVar(&initProjectName, "name", "", "Project name (default: current directory name)")
}

// registerProject loads or creates the registry and adds the current project.
func registerProject() error {
	name := initProjectName
	if name == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("cannot determine working directory: %w", err)
		}
		name = filepath.Base(cwd)
	}

	configPath := cfgPath // from root.go persistent flag

	reg, err := registry.Load()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	if err := reg.Set(name, "local", configPath); err != nil {
		return fmt.Errorf("registering project: %w", err)
	}

	// Set as default if this is the first project
	if reg.Default == "" {
		if err := reg.SetDefault(name); err != nil {
			return fmt.Errorf("setting default project: %w", err)
		}
	}

	regPath, _ := registry.Path()
	fmt.Printf("\nProject registered: %s (env: local)\n", name)
	fmt.Printf("Registry: %s\n", regPath)
	return nil
}
