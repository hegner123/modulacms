package main

import (
	"os"
	"syscall"

	"github.com/hegner123/modulacms/internal/cli"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the terminal UI without the server",
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

		model, _ := cli.InitialModel(&verbose, cfg)
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
