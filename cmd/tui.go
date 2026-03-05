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

Examples:
  modula tui
  modula tui --config /path/to/config.json`,
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
