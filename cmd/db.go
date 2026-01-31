package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/install"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}

var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create database tables and bootstrap data",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		utility.DefaultLogger.Info("Initializing database tables and bootstrap data...")
		if err := install.CreateDbSimple(cfgPath, cfg); err != nil {
			return fmt.Errorf("database initialization failed: %w", err)
		}

		utility.DefaultLogger.Info("Database initialization complete")
		return nil
	},
}

var dbWipeCmd = &cobra.Command{
	Use:   "wipe",
	Short: "Drop all database tables (data is lost)",
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

		wipeConfirm := false
		confirm := huh.NewConfirm().
			Title("WARNING: This will drop ALL tables and delete ALL data. Continue?").
			Value(&wipeConfirm)
		if err := confirm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				utility.DefaultLogger.Info("Wipe cancelled")
				return nil
			}
			return fmt.Errorf("confirmation form error: %w", err)
		}
		if !wipeConfirm {
			utility.DefaultLogger.Info("Wipe cancelled")
			return nil
		}

		driver := db.ConfigDB(*cfg)
		if err := driver.DropAllTables(); err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}

		utility.DefaultLogger.Info("All tables dropped successfully")
		return nil
	},
}

var dbWipeRedeployCmd = &cobra.Command{
	Use:   "wipe-redeploy",
	Short: "Drop all tables and recreate schema with bootstrap data",
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

		wipeConfirm := false
		confirm := huh.NewConfirm().
			Title("WARNING: This will drop ALL tables, delete ALL data, and recreate the schema. Continue?").
			Value(&wipeConfirm)
		if err := confirm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				utility.DefaultLogger.Info("Wipe-redeploy cancelled")
				return nil
			}
			return fmt.Errorf("confirmation form error: %w", err)
		}
		if !wipeConfirm {
			utility.DefaultLogger.Info("Wipe-redeploy cancelled")
			return nil
		}

		driver := db.ConfigDB(*cfg)
		if err := driver.DropAllTables(); err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
		utility.DefaultLogger.Info("All tables dropped")

		if err := driver.CreateAllTables(); err != nil {
			return fmt.Errorf("failed to recreate tables: %w", err)
		}
		if err := driver.CreateBootstrapData(); err != nil {
			return fmt.Errorf("failed to create bootstrap data: %w", err)
		}
		if err := driver.ValidateBootstrapData(); err != nil {
			return fmt.Errorf("failed to validate bootstrap data: %w", err)
		}

		utility.DefaultLogger.Info("Database wiped and redeployed successfully")
		return nil
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete the database file (SQLite only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		utility.DefaultLogger.Info("Resetting database", "path", cfg.Db_URL)
		if err := os.Remove(cfg.Db_URL); err != nil {
			return fmt.Errorf("error deleting database file: %w", err)
		}

		utility.DefaultLogger.Info("Database reset complete")
		return nil
	},
}

var dbExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Dump database SQL to file",
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

		driver := db.ConfigDB(*cfg)
		if err := driver.DumpSql(*cfg); err != nil {
			return fmt.Errorf("database export failed: %w", err)
		}

		utility.DefaultLogger.Info("Database export complete")
		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbInitCmd)
	dbCmd.AddCommand(dbWipeCmd)
	dbCmd.AddCommand(dbWipeRedeployCmd)
	dbCmd.AddCommand(dbResetCmd)
	dbCmd.AddCommand(dbExportCmd)
}
