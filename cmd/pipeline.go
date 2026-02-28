package main

import (
	"context"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// pipelineCmd is the parent command for all pipeline management operations.
var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Pipeline management commands",
}

// pipelineListCmd lists all pipeline entries with a formatted table.
// Offline: uses loadConfigAndDB() for direct DB access.
var pipelineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all pipeline entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		_, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		pipelines, err := driver.ListPipelines()
		if err != nil {
			return fmt.Errorf("listing pipelines: %w", err)
		}

		if pipelines == nil || len(*pipelines) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No pipelines found.")
			return nil
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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		tableName := args[0]

		_, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		pipelines, err := driver.ListPipelinesByTable(tableName)
		if err != nil {
			return fmt.Errorf("listing pipelines for table %q: %w", tableName, err)
		}

		if pipelines == nil || len(*pipelines) == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "No pipelines found for table %q.\n", tableName)
			return nil
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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		pipelineID, err := types.ParsePipelineID(args[0])
		if err != nil {
			return fmt.Errorf("invalid pipeline ID %q: %w", args[0], err)
		}

		_, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		ctx := context.Background()
		ac := audited.Ctx(types.NodeID(""), types.UserID(""), "pipeline-enable", "cli")

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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		pipelineID, err := types.ParsePipelineID(args[0])
		if err != nil {
			return fmt.Errorf("invalid pipeline ID %q: %w", args[0], err)
		}

		_, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		ctx := context.Background()
		ac := audited.Ctx(types.NodeID(""), types.UserID(""), "pipeline-disable", "cli")

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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		pipelineID, err := types.ParsePipelineID(args[0])
		if err != nil {
			return fmt.Errorf("invalid pipeline ID %q: %w", args[0], err)
		}

		_, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		ctx := context.Background()
		ac := audited.Ctx(types.NodeID(""), types.UserID(""), "pipeline-remove", "cli")

		if deleteErr := driver.DeletePipeline(ctx, ac, pipelineID); deleteErr != nil {
			return fmt.Errorf("removing pipeline %s: %w", pipelineID, deleteErr)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Pipeline %s removed.\n", pipelineID)
		return nil
	},
}

func init() {
	pipelineCmd.AddCommand(pipelineListCmd)
	pipelineCmd.AddCommand(pipelineShowCmd)
	pipelineCmd.AddCommand(pipelineEnableCmd)
	pipelineCmd.AddCommand(pipelineDisableCmd)
	pipelineCmd.AddCommand(pipelineRemoveCmd)
}
