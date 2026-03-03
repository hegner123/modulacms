package main

import (
	"os"
	"syscall"

	"github.com/hegner123/modulacms/internal/tui"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal UI without the server",
	Long: `Launch the Bubbletea terminal UI for local content management without starting
the HTTP, HTTPS, or SSH servers. Connects directly to the database configured
in config.json.

Running bare "modula tui" launches the default (v2) panel-based TUI.
Use "modula tui v1" to launch the original tree-based TUI.

Subcommands:
  default    Launch the v2 panel TUI (this is the default)
  v1         Launch the original v1 tree TUI

Examples:
  modula tui                          # launch default TUI
  modula tui default                  # same as above
  modula tui v1                       # launch the v1 TUI`,
	// Bare "tui" delegates to the new TUI package
	RunE: func(cmd *cobra.Command, args []string) error {
		return tuiDefaultCmd.RunE(cmd, args)
	},
}

var tuiDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Launch the new TUI (default)",
	Long: `Launch the v2 panel-based Bubbletea TUI with direct database access.

Reads config.json, opens the database, and starts the interactive terminal
interface. No HTTP or SSH servers are started.

Examples:
  modula tui default
  modula tui default --config /path/to/config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		mgr, _, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		cfg, err := mgr.Config()
		if err != nil {
			return err
		}
		model, _ := tui.NewPanelModel(&verbose, cfg)
		if _, ok := tui.PanelRun(&model); !ok {
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

var tuiV1Cmd = &cobra.Command{
	Use:   "v1",
	Short: "Launch the original v1 TUI",
	Long: `Launch the original v1 tree-based Bubbletea TUI with direct database access.

This is the legacy TUI interface. Use "modula tui" or "modula tui default"
for the current panel-based interface.

Examples:
  modula tui v1
  modula tui v1 --config /path/to/config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

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

func init() {
	tuiCmd.AddCommand(tuiDefaultCmd)
	tuiCmd.AddCommand(tuiV1Cmd)
}
