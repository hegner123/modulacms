package main

import (
	"os"
	"syscall"

	"github.com/hegner123/modulacms/internal/cli"
	"github.com/hegner123/modulacms/internal/db"
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

		cfg, _, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		model, _ := tui.InitialModel(&verbose, cfg)
		if _, ok := tui.Run(&model); !ok {
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

		cfg, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		model, _ := cli.InitialModel(&verbose, cfg, driver, utility.DefaultLogger)
		if _, ok := cli.CliRun(&model); !ok {
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
