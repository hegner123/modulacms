package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/deploy"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Export, import, snapshot, push, and pull content data",
}

// --- deploy export ---

var deployExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export content data to a JSON file",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		outFile, _ := cmd.Flags().GetString("file")
		tablesFlag, _ := cmd.Flags().GetString("tables")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if outFile == "" {
			return fmt.Errorf("--file is required")
		}

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		_ = mgr
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		var tables []db.DBTable
		if tablesFlag != "" {
			for _, name := range strings.Split(tablesFlag, ",") {
				name = strings.TrimSpace(name)
				t, vErr := db.ValidateTableName(name)
				if vErr != nil {
					return vErr
				}
				tables = append(tables, t)
			}
		}

		ctx := context.Background()
		manifest, actualPath, err := deploy.ExportToFile(ctx, driver, tables, outFile)
		if err != nil {
			return fmt.Errorf("export failed: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(manifest, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		utility.DefaultLogger.Info("Export complete",
			"file", actualPath,
			"tables", len(manifest.Tables),
			"version", manifest.Version,
		)
		for _, t := range manifest.Tables {
			utility.DefaultLogger.Info("  "+t, "rows", manifest.RowCounts[t])
		}

		return nil
	},
}

// --- deploy import ---

var deployImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import content data from a JSON export file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		inFile := args[0]
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		skipBackup, _ := cmd.Flags().GetBool("skip-backup")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		cfg, err := mgr.Config()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		if dryRun {
			// Dry run: validate and report impact without modifying the database.
			payload, lErr := loadPayloadFromFile(inFile)
			if lErr != nil {
				return lErr
			}
			result := deploy.BuildDryRunResult(payload, driver)

			if jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return nil
			}

			printDryRunResult(result)
			return nil
		}

		ctx := context.Background()
		result, err := deploy.ImportFromFile(ctx, *cfg, driver, inFile, skipBackup)
		if err != nil {
			if result != nil && jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
			}
			return fmt.Errorf("import failed: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		utility.DefaultLogger.Info("Import complete",
			"tables", len(result.TablesAffected),
			"duration", result.Duration,
			"backup", result.BackupPath,
			"snapshot", result.SnapshotID,
		)
		for _, t := range result.TablesAffected {
			utility.DefaultLogger.Info("  "+t, "rows", result.RowCounts[t])
		}
		if len(result.Warnings) > 0 {
			for _, w := range result.Warnings {
				utility.DefaultLogger.Warn("  "+w, nil)
			}
		}

		return nil
	},
}

// --- deploy snapshot ---

var deploySnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Manage import snapshots",
}

var deploySnapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available snapshots",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		jsonOutput, _ := cmd.Flags().GetBool("json")

		snapshotDir := "./deploy/snapshots"
		snapshots, err := deploy.ListSnapshots(snapshotDir)
		if err != nil {
			return fmt.Errorf("list snapshots: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(snapshots, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(snapshots) == 0 {
			utility.DefaultLogger.Info("No snapshots found")
			return nil
		}

		fmt.Printf("%-28s %-22s %-8s %s\n", "ID", "Timestamp", "Tables", "Size")
		fmt.Println("----------------------------+----------------------+--------+----------")
		for _, s := range snapshots {
			fmt.Printf("%-28s %-22s %-8d %s\n",
				s.ID,
				s.Timestamp.Format("2006-01-02 15:04:05"),
				len(s.Tables),
				formatBytes(s.SizeBytes),
			)
		}

		return nil
	},
}

var deploySnapshotShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show snapshot details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		jsonOutput, _ := cmd.Flags().GetBool("json")
		snapshotDir := "./deploy/snapshots"

		payload, err := deploy.LoadSnapshot(snapshotDir, args[0])
		if err != nil {
			return fmt.Errorf("load snapshot: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(payload.Manifest, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		m := payload.Manifest
		utility.DefaultLogger.Info("Snapshot details",
			"timestamp", m.Timestamp,
			"version", m.Version,
			"schema", m.SchemaVersion[:12]+"...",
			"strategy", string(m.Strategy),
		)
		for _, t := range m.Tables {
			utility.DefaultLogger.Info("  "+t, "rows", m.RowCounts[t])
		}
		utility.DefaultLogger.Info("User refs", "count", len(payload.UserRefs))

		return nil
	},
}

var deploySnapshotRestoreCmd = &cobra.Command{
	Use:   "restore <id>",
	Short: "Restore a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		skipBackup, _ := cmd.Flags().GetBool("skip-backup")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		snapshotDir := "./deploy/snapshots"

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		cfg, err := mgr.Config()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		ctx := context.Background()

		// Load and import (RestoreSnapshot creates a backup by default).
		payload, lErr := deploy.LoadSnapshot(snapshotDir, args[0])
		if lErr != nil {
			return fmt.Errorf("load snapshot: %w", lErr)
		}

		result, err := deploy.ImportPayload(ctx, *cfg, driver, payload, skipBackup)
		if err != nil {
			if result != nil && jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
			}
			return fmt.Errorf("restore failed: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		utility.DefaultLogger.Info("Snapshot restored",
			"id", args[0],
			"tables", len(result.TablesAffected),
			"duration", result.Duration,
		)

		return nil
	},
}

// --- deploy pull ---

var deployPullCmd = &cobra.Command{
	Use:   "pull <source>",
	Short: "Pull data from a remote environment and apply locally",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		envName := args[0]
		tablesFlag, _ := cmd.Flags().GetString("tables")
		skipBackup, _ := cmd.Flags().GetBool("skip-backup")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		cfg, err := mgr.Config()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		tables := parseTablesFlag(tablesFlag)
		if tables == nil && tablesFlag != "" {
			return fmt.Errorf("invalid --tables flag")
		}

		ctx := context.Background()
		result, err := deploy.Pull(ctx, *cfg, driver, envName, tables, skipBackup, dryRun)
		if err != nil {
			if result != nil && jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
			}
			return fmt.Errorf("pull failed: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if dryRun {
			printDryRunResult(result)
			return nil
		}

		utility.DefaultLogger.Info("Pull complete",
			"source", envName,
			"tables", len(result.TablesAffected),
			"duration", result.Duration,
		)
		for _, t := range result.TablesAffected {
			utility.DefaultLogger.Info("  "+t, "rows", result.RowCounts[t])
		}
		if len(result.Warnings) > 0 {
			for _, w := range result.Warnings {
				utility.DefaultLogger.Warn("  "+w, nil)
			}
		}

		return nil
	},
}

// --- deploy push ---

var deployPushCmd = &cobra.Command{
	Use:   "push <target>",
	Short: "Export local data and push to a remote environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		envName := args[0]
		tablesFlag, _ := cmd.Flags().GetString("tables")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer func() {
			if cerr := db.CloseDB(); cerr != nil {
				utility.DefaultLogger.Error("Database pool close error", cerr)
			}
		}()

		cfg, err := mgr.Config()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		tables := parseTablesFlag(tablesFlag)
		if tables == nil && tablesFlag != "" {
			return fmt.Errorf("invalid --tables flag")
		}

		ctx := context.Background()
		result, err := deploy.Push(ctx, *cfg, driver, envName, tables, dryRun)
		if err != nil {
			if result != nil && jsonOutput {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
			}
			return fmt.Errorf("push failed: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if dryRun {
			printDryRunResult(result)
			return nil
		}

		utility.DefaultLogger.Info("Push complete",
			"target", envName,
			"tables", len(result.TablesAffected),
			"duration", result.Duration,
		)
		for _, t := range result.TablesAffected {
			utility.DefaultLogger.Info("  "+t, "rows", result.RowCounts[t])
		}
		if len(result.Warnings) > 0 {
			for _, w := range result.Warnings {
				utility.DefaultLogger.Warn("  "+w, nil)
			}
		}

		return nil
	},
}

// --- deploy env ---

var deployEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage deploy environments",
}

var deployEnvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured deploy environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		jsonOutput, _ := cmd.Flags().GetBool("json")

		cfg, err := loadConfigPtr()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		envs := cfg.Deploy_Environments
		if jsonOutput {
			// Redact API keys in JSON output.
			type safeEnv struct {
				Name   string `json:"name"`
				URL    string `json:"url"`
				HasKey bool   `json:"has_api_key"`
			}
			safe := make([]safeEnv, len(envs))
			for i, e := range envs {
				safe[i] = safeEnv{Name: e.Name, URL: e.URL, HasKey: e.APIKey != ""}
			}
			data, _ := json.MarshalIndent(safe, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		if len(envs) == 0 {
			utility.DefaultLogger.Info("No deploy environments configured")
			utility.DefaultLogger.Info("Add environments to config.json under deploy_environments")
			return nil
		}

		fmt.Printf("%-16s %-40s %s\n", "Name", "URL", "API Key")
		fmt.Println("----------------+----------------------------------------+--------")
		for _, e := range envs {
			keyStatus := "missing"
			if e.APIKey != "" {
				keyStatus = "set"
			}
			fmt.Printf("%-16s %-40s %s\n", e.Name, e.URL, keyStatus)
		}

		return nil
	},
}

var deployEnvTestCmd = &cobra.Command{
	Use:   "test <name>",
	Short: "Test connectivity to a deploy environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		jsonOutput, _ := cmd.Flags().GetBool("json")

		cfg, err := loadConfigPtr()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		ctx := context.Background()
		health, err := deploy.TestEnvConnection(ctx, *cfg, args[0])
		if err != nil {
			if jsonOutput {
				data, _ := json.MarshalIndent(map[string]string{"error": err.Error()}, "", "  ")
				fmt.Println(string(data))
			}
			return fmt.Errorf("connection test failed: %w", err)
		}

		if jsonOutput {
			data, _ := json.MarshalIndent(health, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		utility.DefaultLogger.Info("Connection successful",
			"environment", args[0],
			"status", health.Status,
			"version", health.Version,
			"node_id", health.NodeID,
		)

		return nil
	},
}

// printDryRunResult outputs a human-readable dry-run impact report.
func printDryRunResult(result *deploy.SyncResult) {
	if result.Success {
		utility.DefaultLogger.Info("Dry run: validation passed")
	} else {
		utility.DefaultLogger.Warn("Dry run: validation failed", nil, "errors", len(result.Errors))
	}

	utility.DefaultLogger.Info("Impact report",
		"tables", len(result.TablesAffected),
		"strategy", string(result.Strategy),
	)
	for _, t := range result.TablesAffected {
		utility.DefaultLogger.Info("  "+t, "rows", result.RowCounts[t])
	}

	if len(result.Warnings) > 0 {
		for _, w := range result.Warnings {
			utility.DefaultLogger.Warn("  "+w, nil)
		}
	}

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			utility.DefaultLogger.Warn(fmt.Sprintf("  [%s] %s: %s", e.Phase, e.Table, e.Message), nil)
		}
	}
}

// parseTablesFlag splits a comma-separated table names string into validated DBTable values.
// Returns nil for empty input. Returns nil with non-empty input if any name is invalid.
func parseTablesFlag(flag string) []db.DBTable {
	if flag == "" {
		return nil
	}
	var tables []db.DBTable
	for _, name := range strings.Split(flag, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		t, vErr := db.ValidateTableName(name)
		if vErr != nil {
			return nil
		}
		tables = append(tables, t)
	}
	return tables
}

// loadPayloadFromFile is a helper that reads and decodes a SyncPayload from a file path.
// Handles gzip-compressed files (.json.gz or gzip magic bytes).
func loadPayloadFromFile(path string) (*deploy.SyncPayload, error) {
	data, err := deploy.ReadPayloadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var payload deploy.SyncPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &payload, nil
}

func init() {
	// deploy export flags
	deployExportCmd.Flags().String("file", "", "Output file path (required)")
	deployExportCmd.Flags().String("tables", "", "Comma-separated table names (default: all sync tables)")
	deployExportCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy import flags
	deployImportCmd.Flags().Bool("dry-run", false, "Validate only, do not import")
	deployImportCmd.Flags().Bool("skip-backup", false, "Skip pre-import backup")
	deployImportCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy pull flags
	deployPullCmd.Flags().String("tables", "", "Comma-separated table names (default: all sync tables)")
	deployPullCmd.Flags().Bool("skip-backup", false, "Skip pre-import backup")
	deployPullCmd.Flags().Bool("dry-run", false, "Validate only, show impact report without importing")
	deployPullCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy push flags
	deployPushCmd.Flags().String("tables", "", "Comma-separated table names (default: all sync tables)")
	deployPushCmd.Flags().Bool("dry-run", false, "Validate only, show impact report without importing")
	deployPushCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy snapshot list flags
	deploySnapshotListCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy snapshot show flags
	deploySnapshotShowCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy snapshot restore flags
	deploySnapshotRestoreCmd.Flags().Bool("skip-backup", false, "Skip pre-restore backup")
	deploySnapshotRestoreCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy env list flags
	deployEnvListCmd.Flags().Bool("json", false, "Output as JSON")

	// deploy env test flags
	deployEnvTestCmd.Flags().Bool("json", false, "Output as JSON")

	// Wire subcommands
	deploySnapshotCmd.AddCommand(deploySnapshotListCmd)
	deploySnapshotCmd.AddCommand(deploySnapshotShowCmd)
	deploySnapshotCmd.AddCommand(deploySnapshotRestoreCmd)

	deployEnvCmd.AddCommand(deployEnvListCmd)
	deployEnvCmd.AddCommand(deployEnvTestCmd)

	deployCmd.AddCommand(deployExportCmd)
	deployCmd.AddCommand(deployImportCmd)
	deployCmd.AddCommand(deployPullCmd)
	deployCmd.AddCommand(deployPushCmd)
	deployCmd.AddCommand(deploySnapshotCmd)
	deployCmd.AddCommand(deployEnvCmd)
}
