package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/hegner123/modulacms/internal/registry"
	"github.com/hegner123/modulacms/internal/tui"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui [project] [environment]",
	Short: "Launch the terminal UI without the server",
	Long: `Launch the Bubbletea terminal UI for local content management without starting
the HTTP, HTTPS, or SSH servers. Connects directly to the database configured
in modula.config.json.

If positional arguments are given, the config path is resolved from the global
project registry (~/.modula/configs.json). The first argument is the project
name and the second is the environment name. If only a project name is given,
the project's default environment is used. If neither is given, the --config
flag or ./modula.config.json is used as before.

Examples:
  modula tui
  modula tui mysite
  modula tui mysite production
  modula tui --config /path/to/modula.config.json`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		// Resolve config path from registry when positional args are given.
		// Positional args take priority over the default --config value but
		// an explicit --config flag overrides everything.
		if len(args) > 0 && !cmd.Flags().Changed("config") {
			reg, regErr := registry.Load()
			if regErr != nil {
				return fmt.Errorf("loading project registry: %w", regErr)
			}

			var projName, envName string
			projName = args[0]
			if len(args) > 1 {
				envName = args[1]
			}

			resolved, resolveErr := reg.Resolve(projName, envName)
			if resolveErr != nil {
				return fmt.Errorf("resolving config: %w", resolveErr)
			}

			cfgPath = resolved
			utility.DefaultLogger.Info("Using config from registry", "project", projName, "environment", envName, "path", cfgPath)
		}

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		cfg, err := mgr.Config()
		if err != nil {
			return err
		}
		model, _ := tui.InitialModel(&verbose, cfg, driver, utility.DefaultLogger, nil, mgr, nil, nil)
		if _, ok := tui.CliRun(&model); !ok {
			process, err := os.FindProcess(os.Getpid())
			if err != nil {
				utility.DefaultLogger.Error("", err)
				return err
			}

			if err := process.Signal(syscall.SIGTERM); err != nil {
				utility.DefaultLogger.Error("", err)
				return err
			}
		}

		return nil
	},
}
