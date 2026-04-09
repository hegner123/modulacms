package main

import (
	"fmt"
	"os"
	"path/filepath"

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

If positional arguments are given, the config is resolved from the project
registry (~/.modula/configs.json). The project's base config is loaded first,
then the environment overlay is layered on top.

Examples:
  modula tui                                        # local config
  modula tui mysite                                 # base + default env overlay
  modula tui mysite production                      # base + production overlay
  modula tui --config base.json --overlay prod.json # explicit files`,
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

			cfgPath = resolved.Base
			if resolved.Overlay != "" {
				overlayPath = resolved.Overlay
			}
			utility.DefaultLogger.Info("using config from registry", "project", projName, "environment", envName, "config", cfgPath)
		}

		// Root the process in the config file's directory so relative paths
		// (db_url, cert_dir, log_path, etc.) resolve predictably.
		absCfg, absErr := filepath.Abs(cfgPath)
		if absErr != nil {
			return fmt.Errorf("resolving config path: %w", absErr)
		}
		cfgDir := filepath.Dir(absCfg)

		// Resolve overlay path to absolute before chdir changes the working directory.
		if overlayPath != "" {
			absOverlay, oErr := filepath.Abs(overlayPath)
			if oErr != nil {
				return fmt.Errorf("resolving overlay path: %w", oErr)
			}
			overlayPath = absOverlay
		}

		if err := os.Chdir(cfgDir); err != nil {
			return fmt.Errorf("changing to config directory %s: %w", cfgDir, err)
		}
		cfgPath = filepath.Base(absCfg)

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
			os.Exit(1)
		}

		return nil
	},
}
