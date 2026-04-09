package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// pipelineCmd is the parent command for all pipeline management operations.
var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Pipeline management commands",
	Long: `View and manage plugin pipeline entries in the database.

Pipelines define the execution order of plugin handlers for content operations
(before_insert, after_update, etc.) on specific tables. Each pipeline entry
links a plugin handler to a table + operation with a priority.

All subcommands connect directly to the database (offline).

Subcommands:
  list      Show all pipeline entries in a table
  show      Show pipelines for a specific table, grouped by operation
  enable    Enable a pipeline entry by ID
  disable   Disable a pipeline entry by ID
  remove    Delete a pipeline entry by ID

Examples:
  modula pipeline list
  modula pipeline show content_data
  modula pipeline enable 01HXYZ...
  modula pipeline disable 01HXYZ...
  modula pipeline remove 01HXYZ...`,
}

// pipelineListCmd lists all pipeline entries with a formatted table.
// Offline: uses loadConfigAndDB() for direct DB access.
var pipelineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pipeline entries",
	Long: `Display all pipeline entries across all tables in a formatted table.

Shows pipeline ID, plugin name, table, operation, handler function, priority,
and enabled status for every entry in the database.

Examples:
  modula pipeline list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		_, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		pipelines, err := driver.ListPipelines()
		if err != nil {
			return fmt.Errorf("listing pipelines: %w", err)
		}

		if pipelines == nil || len(*pipelines) == 0 {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode([]struct{}{})
			}
			fmt.Fprintln(cmd.OutOrStdout(), "no pipelines found.")
			return nil
		}

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(*pipelines)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PIPELINE ID\tPLUGIN\tTABLE\tOPERATION\tHANDLER\tPRIORITY\tENABLED")
		for _, p := range *pipelines {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%t\n",
				p.PipelineID,
				p.PluginName,
				p.TableName,
				p.Operation,
				p.Handler,
				p.Priority,
				p.Enabled,
			)
		}
		return w.Flush()
	},
}

// pipelineShowCmd shows pipelines for a specific table, grouped by operation.
// Offline: uses loadConfigAndDB() for direct DB access.
var pipelineShowCmd = &cobra.Command{
	Use:   "show <table>",
	Short: "Show pipelines for a table",
	Long: `Display pipeline entries for a specific table, grouped by operation.

Arguments:
  table   Database table name (e.g. content_data, media, users)

Examples:
  modula pipeline show content_data
  modula pipeline show media`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		tableName := args[0]

		_, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		pipelines, err := driver.ListPipelinesByTable(tableName)
		if err != nil {
			return fmt.Errorf("listing pipelines for table %q: %w", tableName, err)
		}

		if pipelines == nil || len(*pipelines) == 0 {
			jsonOutput, _ := cmd.Flags().GetBool("json")
			if jsonOutput {
				return json.NewEncoder(cmd.OutOrStdout()).Encode([]struct{}{})
			}
			fmt.Fprintf(cmd.OutOrStdout(), "no pipelines found for table %q.\n", tableName)
			return nil
		}

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(*pipelines)
		}

		// Group pipelines by operation.
		grouped := make(map[string][]db.Pipeline)
		for _, p := range *pipelines {
			grouped[p.Operation] = append(grouped[p.Operation], p)
		}

		// Sort operation keys for deterministic output.
		ops := make([]string, 0, len(grouped))
		for op := range grouped {
			ops = append(ops, op)
		}
		sort.Strings(ops)

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Table: %s\n", tableName)

		for _, op := range ops {
			fmt.Fprintf(out, "\n%s:\n", op)
			for _, p := range grouped[op] {
				fmt.Fprintf(out, "  %s %s/%s (priority: %d, enabled: %t)\n",
					p.PipelineID,
					p.PluginName,
					p.Handler,
					p.Priority,
					p.Enabled,
				)
			}
		}

		return nil
	},
}

// pipelineEnableCmd enables a pipeline entry by ID.
// Offline: uses loadConfigAndDB() for direct DB access.
var pipelineEnableCmd = &cobra.Command{
	Use:   "enable <pipeline_id>",
	Short: "Enable a pipeline entry",
	Long: `Enable a disabled pipeline entry so its handler executes during content operations.

Arguments:
  pipeline_id   ULID of the pipeline entry (from "modula pipeline list")

Examples:
  modula pipeline enable 01HXYZ1234567890ABCDEFGH`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		pipelineID, err := types.ParsePipelineID(args[0])
		if err != nil {
			return fmt.Errorf("invalid pipeline ID %q: %w", args[0], err)
		}

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		ctx := context.Background()
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			return cfgErr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), types.UserID(""), "pipeline-enable", "cli")

		if updateErr := driver.UpdatePipelineEnabled(ctx, ac, pipelineID, true); updateErr != nil {
			return fmt.Errorf("enabling pipeline %s: %w", pipelineID, updateErr)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pipeline %s enabled.\n", pipelineID)
		return nil
	},
}

// pipelineDisableCmd disables a pipeline entry by ID.
// Offline: uses loadConfigAndDB() for direct DB access.
var pipelineDisableCmd = &cobra.Command{
	Use:   "disable <pipeline_id>",
	Short: "Disable a pipeline entry",
	Long: `Disable a pipeline entry so its handler is skipped during content operations.

The entry remains in the database and can be re-enabled later.

Arguments:
  pipeline_id   ULID of the pipeline entry (from "modula pipeline list")

Examples:
  modula pipeline disable 01HXYZ1234567890ABCDEFGH`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		pipelineID, err := types.ParsePipelineID(args[0])
		if err != nil {
			return fmt.Errorf("invalid pipeline ID %q: %w", args[0], err)
		}

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		ctx := context.Background()
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			return cfgErr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), types.UserID(""), "pipeline-disable", "cli")

		if updateErr := driver.UpdatePipelineEnabled(ctx, ac, pipelineID, false); updateErr != nil {
			return fmt.Errorf("disabling pipeline %s: %w", pipelineID, updateErr)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pipeline %s disabled.\n", pipelineID)
		return nil
	},
}

// pipelineRemoveCmd removes a pipeline entry by ID.
// Offline: uses loadConfigAndDB() for direct DB access.
var pipelineRemoveCmd = &cobra.Command{
	Use:   "remove <pipeline_id>",
	Short: "Remove a pipeline entry",
	Long: `Permanently delete a pipeline entry from the database.

This is irreversible. To temporarily stop a handler, use "modula pipeline disable"
instead.

Arguments:
  pipeline_id   ULID of the pipeline entry (from "modula pipeline list")

Examples:
  modula pipeline remove 01HXYZ1234567890ABCDEFGH`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		pipelineID, err := types.ParsePipelineID(args[0])
		if err != nil {
			return fmt.Errorf("invalid pipeline ID %q: %w", args[0], err)
		}

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		ctx := context.Background()
		cfg, cfgErr := mgr.Config()
		if cfgErr != nil {
			return cfgErr
		}
		ac := audited.Ctx(types.NodeID(cfg.Node_ID), types.UserID(""), "pipeline-remove", "cli")

		if deleteErr := driver.DeletePipeline(ctx, ac, pipelineID); deleteErr != nil {
			return fmt.Errorf("removing pipeline %s: %w", pipelineID, deleteErr)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pipeline %s removed.\n", pipelineID)
		return nil
	},
}

func init() {
	pipelineListCmd.Flags().Bool("json", false, "Output as JSON")
	pipelineShowCmd.Flags().Bool("json", false, "Output as JSON")

	pipelineCmd.AddCommand(pipelineListCmd)
	pipelineCmd.AddCommand(pipelineShowCmd)
	pipelineCmd.AddCommand(pipelineEnableCmd)
	pipelineCmd.AddCommand(pipelineDisableCmd)
	pipelineCmd.AddCommand(pipelineRemoveCmd)
}
