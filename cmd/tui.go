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
	// Bare "tui" delegates to the new TUI package
	RunE: func(cmd *cobra.Command, args []string) error {
		return tuiDefaultCmd.RunE(cmd, args)
	},
}

var tuiDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Launch the new TUI (default)",
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
		model, _ := tui.InitialModel(&verbose, cfg, driver, utility.DefaultLogger, nil, mgr, nil)
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
