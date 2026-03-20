package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var configParentCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long: `View, validate, and modify the Modula configuration file (modula.config.json).

Subcommands:
  show       Print the full loaded configuration as JSON
  validate   Check that required fields are present and valid
  set        Update a single configuration field and save to disk
  fields     List every available configuration field with its description

Examples:
  modula config show
  modula config validate
  modula config set port ":9090"
  modula config fields
  modula config fields --category server
  modula config fields port`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print the loaded configuration as JSON",
	Long: `Load modula.config.json and print its full contents as formatted JSON to stdout.

Reads from the path specified by --config (default: ./modula.config.json). Environment
variable references (${VAR}) in the config file are expanded before display.

When using --overlay, the default output is the merged config. Use --raw to print
the overlay file contents only (useful for seeing what a specific environment overrides).

Examples:
  modula config show
  modula config show --overlay modula.config.prod.json
  modula config show --overlay modula.config.prod.json --raw`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		rawFlag, _ := cmd.Flags().GetBool("raw")

		// --raw without --overlay is meaningless
		if rawFlag && overlayPath == "" {
			return fmt.Errorf("--raw requires --overlay to be set")
		}

		// --raw: show the overlay file contents only
		if rawFlag {
			rawCfg, err := config.NewFileProvider(overlayPath).Get()
			if err != nil {
				return fmt.Errorf("loading overlay file: %w", err)
			}
			formatted, err := utility.FormatJSON(rawCfg)
			if err != nil {
				return fmt.Errorf("formatting overlay: %w", err)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
			return err
		}

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
	Long: `Load modula.config.json and check that all required fields are present and non-empty.

Required fields: db_driver, db_url, port, ssh_port. Reports the count and
details of any validation errors found.

Examples:
  modula config validate
  modula config validate --config /etc/modula/modula.config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfigPtr()
		if err != nil {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			if jsonOutput {
				result := struct {
					Valid  bool     `json:"valid"`
					Errors []string `json:"errors"`
				}{
					Valid:  false,
					Errors: []string{err.Error()},
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
			}
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

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			result := struct {
				Valid  bool     `json:"valid"`
				Errors []string `json:"errors"`
			}{
				Valid:  len(errs) == 0,
				Errors: errs,
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
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
	Long: `Update a single field in modula.config.json and write the file back to disk.

The key must be a known configuration field name. Use "modula config fields" to
see all available keys and their descriptions. If the updated field requires a
server restart to take effect, a notice is printed.

When using --overlay, writes target the overlay file by default (operator intent
is to override). Use --base to write to the base config file instead.

Arguments:
  key     Configuration field name (e.g. port, db_driver, environment)
  value   New value for the field

Examples:
  modula config set port ":9090"
  modula config set environment "production"
  modula config set db_driver "postgres" --overlay modula.config.prod.json
  modula config set db_driver "sqlite" --overlay modula.config.prod.json --base`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		key := args[0]
		value := args[1]
		baseFlag, _ := cmd.Flags().GetBool("base")

		// --base without --overlay is meaningless
		if baseFlag && overlayPath == "" {
			return fmt.Errorf("--base requires --overlay to be set")
		}

		// Validate the key exists in the field registry
		if _, ok := config.FieldByKey(key); !ok {
			return fmt.Errorf("unknown config key %q; run 'modula config fields' to see available keys", key)
		}

		// When --base is set with a layered config, we need to load the base
		// config directly, update it, and save back to base.
		if baseFlag {
			baseMgr := config.NewManager(config.NewFileProvider(cfgPath))
			if err := baseMgr.Load(); err != nil {
				return fmt.Errorf("loading base configuration: %w", err)
			}

			updates := map[string]any{key: value}
			result, err := baseMgr.Update(updates)
			if err != nil {
				return fmt.Errorf("updating base config: %w", err)
			}
			if !result.Valid {
				return fmt.Errorf("validation failed:\n  %s", strings.Join(result.Errors, "\n  "))
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s = %q (in base config)\n", key, value)
			if len(result.RestartRequired) > 0 {
				fmt.Fprintf(cmd.OutOrStderr(), "Note: the following fields require a server restart to take effect:\n  %s\n", strings.Join(result.RestartRequired, "\n  "))
			}
			return nil
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

		target := ""
		if overlayPath != "" {
			target = " (in overlay)"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Updated %s = %q%s\n", key, value, target)

		if len(result.RestartRequired) > 0 {
			fmt.Fprintf(cmd.OutOrStderr(), "Note: the following fields require a server restart to take effect:\n  %s\n", strings.Join(result.RestartRequired, "\n  "))
		}

		return nil
	},
}

var configFieldsCmd = &cobra.Command{
	Use:   "fields [field-name]",
	Short: "List every available configuration field",
	Long: `List all configuration fields that modula.config.json accepts, grouped by category.

Each field shows its JSON key (used with "modula config set"), a description,
and flags indicating whether it is required, hot-reloadable (takes effect
without restart), or sensitive (contains secrets).

When filtering by --category or looking up a specific field by name, example
values and usage are shown for each field.

Arguments:
  field-name   Optional. A specific field key to look up (e.g. port, db_driver).
               Shows full detail including example value and config set usage.

Flags:
  --category   Filter by category (server, database, storage, cors, cookie,
               oauth, observability, email, plugin, update, misc)

Examples:
  modula config fields                     # list all fields (compact)
  modula config fields --category server   # list server fields with examples
  modula config fields --category database # list database fields with examples
  modula config fields port                # show detail for a single field`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		categoryFlag, _ := cmd.Flags().GetString("category")
		out := cmd.OutOrStdout()

		// Single field lookup
		if len(args) == 1 {
			f, ok := config.FieldByKey(args[0])
			if !ok {
				return fmt.Errorf("unknown field %q; run 'modula config fields' to see all available fields", args[0])
			}
			printFieldDetail(out, f)
			return nil
		}

		// Determine whether to show examples (filtered view) or compact (full listing)
		showExamples := categoryFlag != ""

		categories := config.AllCategories()
		if categoryFlag != "" {
			matched := false
			for _, c := range categories {
				if string(c) == categoryFlag {
					categories = []config.FieldCategory{c}
					matched = true
					break
				}
			}
			if !matched {
				var names []string
				for _, c := range config.AllCategories() {
					names = append(names, string(c))
				}
				return fmt.Errorf("unknown category %q; available: %s", categoryFlag, strings.Join(names, ", "))
			}
		}

		for i, cat := range categories {
			if i > 0 {
				fmt.Fprintln(out)
			}
			fmt.Fprintf(out, "[%s]\n", config.CategoryLabel(cat))

			fields := config.FieldsByCategory(cat)
			for _, f := range fields {
				var tags []string
				if f.Required {
					tags = append(tags, "required")
				}
				if f.HotReloadable {
					tags = append(tags, "hot-reload")
				}
				if f.Sensitive {
					tags = append(tags, "sensitive")
				}

				tagStr := ""
				if len(tags) > 0 {
					tagStr = "  [" + strings.Join(tags, ", ") + "]"
				}

				fmt.Fprintf(out, "  %-35s %s%s\n", f.JSONKey, f.Description, tagStr)

				if showExamples && f.Example != "" {
					fmt.Fprintf(out, "  %-35s example: %s\n", "", f.Example)
					fmt.Fprintf(out, "  %-35s usage:   modula config set %s %q\n", "", f.JSONKey, f.Example)
				}
			}
		}

		return nil
	},
}

// printFieldDetail prints full information for a single field, including
// its category, description, flags, example value, and config set usage.
func printFieldDetail(out io.Writer, f config.FieldMeta) {
	fmt.Fprintf(out, "Field:       %s\n", f.JSONKey)
	fmt.Fprintf(out, "Label:       %s\n", f.Label)
	fmt.Fprintf(out, "Category:    %s\n", config.CategoryLabel(f.Category))
	fmt.Fprintf(out, "Description: %s\n", f.Description)

	var flags []string
	if f.Required {
		flags = append(flags, "required")
	}
	if f.HotReloadable {
		flags = append(flags, "hot-reloadable")
	} else {
		flags = append(flags, "requires restart")
	}
	if f.Sensitive {
		flags = append(flags, "sensitive")
	}
	fmt.Fprintf(out, "Flags:       %s\n", strings.Join(flags, ", "))

	if f.Example != "" {
		fmt.Fprintf(out, "Example:     %s\n", f.Example)
		fmt.Fprintf(out, "\nUsage:\n  modula config set %s %q\n", f.JSONKey, f.Example)
	}
}

var configTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Print a complete modula.config.json template with all fields",
	Long: `Generate a modula.config.json containing every available configuration field
with its default value. Use this as a starting point for a new installation
or as a reference for all configurable options.

Examples:
  modula config template
  modula config template > modula.config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		formatted, err := utility.FormatJSON(cfg)
		if err != nil {
			return fmt.Errorf("formatting template: %w", err)
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), formatted)
		return err
	},
}

var configOverlayCmd = &cobra.Command{
	Use:   "overlay --env <name>",
	Short: "Generate a minimal overlay config file for an environment",
	Long: `Create a modula.config.<env>.json overlay file containing only the
environment field. Use this as a starting point, then add only the fields
that differ from the base config for that environment.

Examples:
  modula config overlay --env prod
  modula config overlay --env staging`,
	RunE: func(cmd *cobra.Command, args []string) error {
		envName, _ := cmd.Flags().GetString("env")
		if envName == "" {
			return fmt.Errorf("--env is required")
		}

		filename := fmt.Sprintf("modula.config.%s.json", envName)

		// Check if the file already exists
		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("%s already exists; remove it first or edit it directly", filename)
		}

		overlay := map[string]any{
			"environment": envName,
		}
		data, err := json.MarshalIndent(overlay, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling overlay: %w", err)
		}
		data = append(data, '\n')

		if err := os.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", filename)
		fmt.Fprintf(cmd.OutOrStdout(), "Usage: modula serve --overlay %s\n", filename)
		return nil
	},
}

func init() {
	configParentCmd.AddCommand(configShowCmd)
	configParentCmd.AddCommand(configValidateCmd)
	configParentCmd.AddCommand(configSetCmd)
	configParentCmd.AddCommand(configFieldsCmd)
	configParentCmd.AddCommand(configTemplateCmd)
	configParentCmd.AddCommand(configOverlayCmd)

	configValidateCmd.Flags().Bool("json", false, "Output as JSON")
	configShowCmd.Flags().Bool("raw", false, "show overlay file contents only (requires --overlay)")
	configSetCmd.Flags().Bool("base", false, "write to base config instead of overlay (requires --overlay)")
	configOverlayCmd.Flags().String("env", "", "environment name for the overlay file (required)")
	configFieldsCmd.Flags().String("category", "", "filter by category (server, database, storage, cors, cookie, oauth, observability, email, plugin, update, misc)")
}
