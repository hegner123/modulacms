package main

import (
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
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

		cfg, err := loadConfigPtr()
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

		cfg, err := loadConfigPtr()
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

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Update a configuration field and save to disk",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		key := args[0]
		value := args[1]

		// Validate the key exists in the field registry
		if _, ok := config.FieldByKey(key); !ok {
			return fmt.Errorf("unknown config key %q; run 'config show' to see available keys", key)
		}

		mgr, err := loadConfig()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		updates := map[string]any{key: value}
		result, err := mgr.Update(updates)
		if err != nil {
			return fmt.Errorf("updating config: %w", err)
		}

		if !result.Valid {
			return fmt.Errorf("validation failed:\n  %s", strings.Join(result.Errors, "\n  "))
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Updated %s = %q\n", key, value)

		if len(result.RestartRequired) > 0 {
			fmt.Fprintf(cmd.OutOrStderr(), "Note: the following fields require a server restart to take effect:\n  %s\n", strings.Join(result.RestartRequired, "\n  "))
		}

		return nil
	},
}

func init() {
	configParentCmd.AddCommand(configShowCmd)
	configParentCmd.AddCommand(configValidateCmd)
	configParentCmd.AddCommand(configSetCmd)
}
