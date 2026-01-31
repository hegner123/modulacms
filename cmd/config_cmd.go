package main

import (
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var configParentCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print the loaded configuration as JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		formatted, err := utility.FormatJSON(cfg)
		if err != nil {
			return fmt.Errorf("formatting configuration: %w", err)
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
		return err
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("configuration is invalid: %w", err)
		}

		// Validate required fields
		var errs []string

		if cfg.Db_Driver == "" {
			errs = append(errs, "db_driver is required")
		}
		if cfg.Db_URL == "" {
			errs = append(errs, "db_url is required")
		}
		if cfg.Port == "" {
			errs = append(errs, "port is required")
		}
		if cfg.SSH_Port == "" {
			errs = append(errs, "ssh_port is required")
		}

		if len(errs) > 0 {
			return fmt.Errorf("configuration has %d validation error(s):\n  %s", len(errs), strings.Join(errs, "\n  "))
		}

		utility.DefaultLogger.Info("Configuration is valid")
		return nil
	},
}

func init() {
	configParentCmd.AddCommand(configShowCmd)
	configParentCmd.AddCommand(configValidateCmd)
}
