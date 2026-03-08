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
	Long: `Sync content data between Modula environments.

Export and import move data as JSON files. Push and pull transfer data between
the local database and a remote CMS environment configured in config.json.
Snapshots provide rollback points for import operations.

Subcommands:
  export     Export content tables to a JSON file
  import     Import content from a JSON export file
  pull       Download data from a remote environment and apply locally
  push       Upload local data to a remote environment
  snapshot   List, show, or restore import snapshots
  env        List and test configured deploy environments

All subcommands support --json for machine-readable output.

Examples:
  modula deploy export --file data.json
  modula deploy import data.json
  modula deploy pull staging
  modula deploy push production
  modula deploy snapshot list
  modula deploy env list`,
}

// --- deploy export ---

var deployExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export content data to a JSON file",
	Long: `Export content tables from the local database to a JSON file.

By default, exports all sync-eligible tables. Use --tables to export only
specific tables. The output includes a manifest with table names, row counts,
schema version, and timestamp.

Flags:
  --file     Output file path (required)
  --tables   Comma-separated table names to export (default: all sync tables)
  --json     Print the export manifest as JSON instead of log output

Examples:
  modula deploy export --file data.json
  modula deploy export --file content-only.json --tables content_data,content_tree
  modula deploy export --file data.json --json`,
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
		defer closeDBWithLog()

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
	Long: `Import content data from a previously exported JSON file into the local database.

By default, creates a pre-import backup and a snapshot for rollback. Use
--dry-run to validate the payload and see an impact report without modifying
the database. Supports gzip-compressed files (.json.gz).

Arguments:
  file   Path to the JSON export file

Flags:
  --dry-run       Validate and report impact without importing
  --skip-backup   Skip the pre-import backup
  --json          Output results as JSON

Examples:
  modula deploy import data.json
  modula deploy import data.json --dry-run
  modula deploy import data.json.gz --skip-backup`,
	Args: cobra.ExactArgs(1),
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
		defer closeDBWithLog()

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
	Long: `List, inspect, and restore import snapshots.

Snapshots are automatically created during imports and provide rollback points.
They are stored in ./deploy/snapshots/.

Subcommands:
  list      Show available snapshots with timestamps and sizes
  show      Display detailed manifest for a specific snapshot
  restore   Re-import data from a snapshot

Examples:
  modula deploy snapshot list
  modula deploy snapshot show 01HXYZ...
  modula deploy snapshot restore 01HXYZ...`,
}

var deploySnapshotListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available snapshots",
	Long: `Show all available import snapshots with their ID, timestamp, table count, and size.

Flags:
  --json   Output as JSON

Examples:
  modula deploy snapshot list
  modula deploy snapshot list --json`,
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
	Long: `Display the full manifest for a specific snapshot including timestamp, version,
schema hash, strategy, tables, row counts, and user references.

Arguments:
  id   Snapshot ULID (from "modula deploy snapshot list")

Flags:
  --json   Output manifest as JSON

Examples:
  modula deploy snapshot show 01HXYZ1234567890ABCDEFGH`,
	Args: cobra.ExactArgs(1),
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
	Long: `Restore the database to the state captured in a snapshot.

Loads the snapshot payload and imports it into the current database. A backup
is created before restoring by default.

Arguments:
  id   Snapshot ULID (from "modula deploy snapshot list")

Flags:
  --skip-backup   Skip the pre-restore backup
  --json          Output results as JSON

Examples:
  modula deploy snapshot restore 01HXYZ1234567890ABCDEFGH
  modula deploy snapshot restore 01HXYZ... --skip-backup`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		skipBackup, _ := cmd.Flags().GetBool("skip-backup")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		snapshotDir := "./deploy/snapshots"

		mgr, driver, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

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
	Long: `Download content data from a remote CMS environment and import it into the local
database.

The source environment must be configured in config.json under deploy_environments
with a name, URL, and API key. A pre-import backup is created by default.

Arguments:
  source   Environment name (from config.json deploy_environments)

Flags:
  --tables        Comma-separated table names (default: all sync tables)
  --skip-backup   Skip the pre-import backup
  --dry-run       Validate and show impact report without importing
  --json          Output results as JSON

Examples:
  modula deploy pull staging
  modula deploy pull production --tables content_data,content_tree
  modula deploy pull staging --dry-run`,
	Args: cobra.ExactArgs(1),
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
		defer closeDBWithLog()

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
	Long: `Export content data from the local database and upload it to a remote CMS
environment.

The target environment must be configured in config.json under deploy_environments
with a name, URL, and API key.

Arguments:
  target   Environment name (from config.json deploy_environments)

Flags:
  --tables     Comma-separated table names (default: all sync tables)
  --dry-run    Validate and show impact report without pushing
  --json       Output results as JSON

Examples:
  modula deploy push staging
  modula deploy push production --tables content_data,content_tree
  modula deploy push staging --dry-run`,
	Args: cobra.ExactArgs(1),
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
		defer closeDBWithLog()

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
	Long: `List and test remote deploy environments configured in config.json.

Environments are defined under deploy_environments in config.json, each with
a name, URL, and API key.

Subcommands:
  list   Show all configured environments (API keys are redacted)
  test   Test connectivity to a specific environment

Examples:
  modula deploy env list
  modula deploy env test staging`,
}

var deployEnvListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured deploy environments",
	Long: `Display all deploy environments from config.json with name, URL, and API key status.

API keys are not printed in plain text; only "set" or "missing" is shown.
Use --json for machine-readable output (keys redacted as has_api_key boolean).

Flags:
  --json   Output as JSON

Examples:
  modula deploy env list
  modula deploy env list --json`,
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
	Long: `Send a health check request to a remote deploy environment and report the result.

Tests that the URL is reachable, the API key is valid, and the remote CMS is
responding. Displays the remote server's status, version, and node ID.

Arguments:
  name   Environment name (from config.json deploy_environments)

Flags:
  --json   Output as JSON

Examples:
  modula deploy env test staging
  modula deploy env test production --json`,
	Args: cobra.ExactArgs(1),
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
