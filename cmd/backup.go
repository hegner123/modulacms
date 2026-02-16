package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup root command for creating, restoring, and listing backups.
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore commands",
}

// backupCreateCmd represents the backup create subcommand that creates a full backup of the database and configured paths.
var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a full backup of the database and configured paths",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		mgr, _, err := loadConfigAndDB()
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

		driver := db.ConfigDB(*cfg)
		backupID := types.NewBackupID()
		startTime := time.Now().UTC()

		_, recordErr := driver.CreateBackup(db.CreateBackupParams{
			BackupID:    backupID,
			NodeID:      types.NodeID(cfg.Node_ID),
			BackupType:  types.BackupTypeFull,
			Status:      types.BackupStatusInProgress,
			StartedAt:   types.NewTimestamp(startTime),
			StoragePath: "",
			TriggeredBy: types.NullableString{String: "cli", Valid: true},
			Metadata:    types.JSONData{Valid: false},
		})
		if recordErr != nil {
			utility.DefaultLogger.Warn("Failed to record backup start (continuing anyway)", recordErr)
		}

		utility.DefaultLogger.Info("Creating backup...")

		path, sizeBytes, err := backup.CreateFullBackup(*cfg)
		if err != nil {
			if recordErr == nil {
				updateErr := driver.UpdateBackupStatus(db.UpdateBackupStatusParams{
					BackupID:     backupID,
					Status:       types.BackupStatusFailed,
					CompletedAt:  types.NewTimestamp(time.Now().UTC()),
					DurationMs:   types.NullableInt64{Int64: time.Since(startTime).Milliseconds(), Valid: true},
					ErrorMessage: types.NullableString{String: err.Error(), Valid: true},
				})
				if updateErr != nil {
					utility.DefaultLogger.Warn("Failed to record backup failure", updateErr)
				}
			}
			return fmt.Errorf("backup failed: %w", err)
		}

		if recordErr == nil {
			updateErr := driver.UpdateBackupStatus(db.UpdateBackupStatusParams{
				BackupID:    backupID,
				Status:      types.BackupStatusCompleted,
				CompletedAt: types.NewTimestamp(time.Now().UTC()),
				DurationMs:  types.NullableInt64{Int64: time.Since(startTime).Milliseconds(), Valid: true},
				SizeBytes:   types.NullableInt64{Int64: sizeBytes, Valid: true},
			})
			if updateErr != nil {
				utility.DefaultLogger.Warn("Failed to record backup completion", updateErr)
			}
		}

		utility.DefaultLogger.Info("Backup created successfully",
			"path", path,
			"size", formatBytes(sizeBytes),
		)

		return nil
	},
}

// backupRestoreCmd represents the backup restore subcommand that restores the database from a backup archive.
var backupRestoreCmd = &cobra.Command{
	Use:   "restore <path>",
	Short: "Restore from a backup archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()
		backupPath := args[0]

		cfg, err := loadConfigPtr()
		if err != nil {
			return fmt.Errorf("loading configuration: %w", err)
		}

		// Read and display manifest
		manifest, err := backup.ReadManifest(backupPath)
		if err != nil {
			return fmt.Errorf("failed to read backup: %w", err)
		}

		utility.DefaultLogger.Info("Backup details",
			"driver", manifest.Driver,
			"timestamp", manifest.Timestamp,
			"version", manifest.Version,
			"node_id", manifest.NodeID,
			"db_name", manifest.DbName,
		)

		// Confirm with user
		restoreConfirm := false
		confirm := huh.NewConfirm().
			Title("WARNING: This will replace the current database. Continue?").
			Value(&restoreConfirm)
		if err := confirm.Run(); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				utility.DefaultLogger.Info("Restore cancelled")
				return nil
			}
			return fmt.Errorf("confirmation form error: %w", err)
		}
		if !restoreConfirm {
			utility.DefaultLogger.Info("Restore cancelled")
			return nil
		}

		utility.DefaultLogger.Info("Restoring backup...")

		if err := backup.RestoreFromBackup(*cfg, backupPath); err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}

		utility.DefaultLogger.Info("Backup restored successfully")
		return nil
	},
}

// backupListCmd represents the backup list subcommand that displays backup history from the database.
var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup history from the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		configureLogger()

		mgr, _, err := loadConfigAndDB()
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

		driver := db.ConfigDB(*cfg)
		backups, err := driver.ListBackups(db.ListBackupsParams{Limit: 50, Offset: 0})
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if backups == nil || len(*backups) == 0 {
			utility.DefaultLogger.Info("No backups found")
			return nil
		}

		fmt.Printf("%-28s %-12s %-12s %-22s %-12s %s\n",
			"ID", "Type", "Status", "Started", "Size", "Path")
		fmt.Println("----------------------------+------------+------------+----------------------+------------+--------------------")

		for _, b := range *backups {
			size := ""
			if b.SizeBytes.Valid {
				size = formatBytes(b.SizeBytes.Int64)
			}
			fmt.Printf("%-28s %-12s %-12s %-22s %-12s %s\n",
				b.BackupID,
				b.BackupType,
				b.Status,
				b.StartedAt,
				size,
				b.StoragePath,
			)
		}

		return nil
	},
}

// formatBytes converts a byte count to a human-readable format with appropriate units (B, KB, MB, etc.).
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func init() {
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	backupCmd.AddCommand(backupListCmd)
}
