package main

import (
	"errors"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"github.com/hegner123/modulacms/internal/auth"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

// dbCmd is the root command for database management operations.
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "database management commands",
	Long: `Initialize, inspect, and manage the Modula database.

Supports SQLite, MySQL, and PostgreSQL as configured in modula.config.json.

Subcommands:
  init           Create all tables and seed bootstrap data (roles, permissions, admin user)
  wipe           Drop all tables (prompts for confirmation)
  wipe-redeploy  Drop all tables, recreate schema, and re-seed with new admin password
  reset          Delete the SQLite database file (SQLite only)
  export         Dump the database to a SQL file

Examples:
  modula db init
  modula db wipe
  modula db wipe-redeploy
  modula db reset
  modula db export`,
}

// dbInitCmd creates database tables and bootstrap data after prompting for admin credentials.
var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create database tables and bootstrap data",
	Long: `Create all database tables and seed bootstrap data for a fresh installation.

Prompts for a system admin password (minimum 8 characters) and confirmation.
Creates tables, default roles (admin, editor, viewer), permissions, and the
system admin user.

Use this when modula.config.json and the database already exist but tables have not
been created (e.g. after a manual database reset or fresh PostgreSQL/MySQL setup).

Examples:
  modula db init
  modula db init --config /etc/modula/modula.config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfigPtr()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		// Collect admin password for the bootstrap user
		adminPassword := ""
		adminConfirm := ""
		pwForm := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("System admin password (min 8 characters)").
				Value(&adminPassword).
				EchoMode(huh.EchoModePassword).
				Validate(install.ValidatePassword),
			huh.NewInput().
				Title("Confirm admin password").
				Value(&adminConfirm).
				EchoMode(huh.EchoModePassword).
				Validate(install.ValidatePassword),
		))
		if err := pwForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				utility.DefaultLogger.Info("database init cancelled")
				return nil
			}
			return fmt.Errorf("password form error: %w", err)
		}
		if adminPassword != adminConfirm {
			return fmt.Errorf("passwords do not match")
		}
		adminHash, err := auth.HashPassword(adminPassword)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		utility.DefaultLogger.Info("initializing database tables and bootstrap data...")
		if err := install.CreateDbSimple(cfgPath, cfg, adminHash); err != nil {
			return fmt.Errorf("database initialization failed: %w", err)
		}

		utility.DefaultLogger.Info("database initialization complete")
		return nil
	},
}

// dbWipeCmd drops all database tables after user confirmation.
var dbWipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Drop all database tables (data is lost)",
	Long: `Drop every table in the database. All data is permanently deleted.

Prompts for confirmation before proceeding. The database connection and file
remain intact; only the tables are removed. After wiping, use "modula db init"
to recreate tables.

This is a destructive, irreversible operation. Create a backup first.

Examples:
  modula db wipe`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		mgr, _, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		cfg, err := mgr.Config()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		wipeConfirm := false
		confirm := huh.NewConfirm().
			Title("WARNING: This will drop ALL tables and delete ALL data. Continue?").
			Value(&wipeConfirm)
		if err := confirm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				utility.DefaultLogger.Info("wipe cancelled")
				return nil
			}
			return fmt.Errorf("confirmation form error: %w", err)
		}
		if !wipeConfirm {
			utility.DefaultLogger.Info("wipe cancelled")
			return nil
		}

		driver := db.ConfigDB(*cfg)
		if err := driver.DropAllTables(); err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}

		utility.DefaultLogger.Info("all tables dropped successfully")
		return nil
	},
}

// dbWipeRedeployCmd drops all tables, recreates the schema, and initializes bootstrap data with a new admin password.
var dbWipeRedeployCmd = &cobra.Command{
	Use:   "wipe-redeploy",
	Short: "Drop all tables and recreate schema with bootstrap data",
	Long: `Drop all tables, recreate the full schema, and seed fresh bootstrap data.

Equivalent to running "db wipe" followed by "db init" in a single command.
Prompts for confirmation and a new system admin password. Useful for resetting
a development database to a clean state.

This is a destructive, irreversible operation. Create a backup first.

Examples:
  modula db wipe-redeploy`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		mgr, _, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		cfg, err := mgr.Config()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		wipeConfirm := false
		confirm := huh.NewConfirm().
			Title("WARNING: This will drop ALL tables, delete ALL data, and recreate the schema. Continue?").
			Value(&wipeConfirm)
		if err := confirm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				utility.DefaultLogger.Info("wipe-redeploy cancelled")
				return nil
			}
			return fmt.Errorf("confirmation form error: %w", err)
		}
		if !wipeConfirm {
			utility.DefaultLogger.Info("wipe-redeploy cancelled")
			return nil
		}

		// Collect admin password for the new bootstrap user
		adminPassword := ""
		adminConfirm := ""
		pwForm := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("New system admin password (min 8 characters)").
				Value(&adminPassword).
				EchoMode(huh.EchoModePassword).
				Validate(install.ValidatePassword),
			huh.NewInput().
				Title("Confirm admin password").
				Value(&adminConfirm).
				EchoMode(huh.EchoModePassword).
				Validate(install.ValidatePassword),
		))
		if err := pwForm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				utility.DefaultLogger.Info("wipe-redeploy cancelled")
				return nil
			}
			return fmt.Errorf("password form error: %w", err)
		}
		if adminPassword != adminConfirm {
			return fmt.Errorf("passwords do not match")
		}
		adminHash, err := auth.HashPassword(adminPassword)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		driver := db.ConfigDB(*cfg)
		if err := driver.DropAllTables(); err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
		utility.DefaultLogger.Info("all tables dropped")

		if err := driver.CreateAllTables(); err != nil {
			return fmt.Errorf("failed to recreate tables: %w", err)
		}
		if err := driver.CreateBootstrapData(adminHash); err != nil {
			return fmt.Errorf("failed to create bootstrap data: %w", err)
		}
		if err := driver.ValidateBootstrapData(); err != nil {
			return fmt.Errorf("failed to validate bootstrap data: %w", err)
		}

		utility.DefaultLogger.Info("database wiped and redeployed successfully")
		return nil
	},
}

// dbResetCmd deletes the database file (SQLite only).
var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete the database file (SQLite only)",
	Long: `Delete the SQLite database file from disk.

Only works with SQLite. The file path is read from db_url in modula.config.json.
After deletion, use "modula db init" or "modula serve" (which auto-initializes)
to create a fresh database.

This is a destructive, irreversible operation. Create a backup first.

Examples:
  modula db reset`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfigPtr()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		utility.DefaultLogger.Info("resetting database", "path", cfg.Db_URL)
		if err := os.Remove(cfg.Db_URL); err != nil {
			return fmt.Errorf("error deleting database file: %w", err)
		}

		utility.DefaultLogger.Info("database reset complete")
		return nil
	},
}

// dbExportCmd exports the database schema and data as SQL to a file.
var dbExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Dump database SQL to file",
	Long: `Export the full database schema and data as a SQL file.

By default, the output file is written to the current directory with a
timestamped name (e.g., sqlite2026-03-20T14:30:00Z.sql). Use --file to
specify a custom output path.

This produces a plain SQL dump (not a backup archive). For full backup
archives with metadata, use "modula backup create" instead.

Flags:
  --file     Output file path (optional; default: timestamped file in current directory)

Examples:
  modula db export
  modula db export --file backup.sql
  modula db export --file /tmp/dump.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		outFile, _ := cmd.Flags().GetString("file")

		mgr, _, err := loadConfigAndDB()
		if err != nil {
			return err
		}
		defer closeDBWithLog()

		cfg, err := mgr.Config()
		if err != nil {
			return fmt.Errorf("reading configuration: %w", err)
		}

		driver := db.ConfigDB(*cfg)
		if err := driver.DumpSql(*cfg, outFile); err != nil {
			return fmt.Errorf("database export failed: %w", err)
		}

		utility.DefaultLogger.Info("database export complete")
		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbWipeCmd)
	dbCmd.AddCommand(dbWipeRedeployCmd)
	dbCmd.AddCommand(dbResetCmd)
	dbCmd.AddCommand(dbExportCmd)

	dbExportCmd.Flags().String("file", "", "Output file path (default: timestamped file in current directory)")
}
