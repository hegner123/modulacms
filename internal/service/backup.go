package service

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/backup"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
)

// BackupService wraps backup creation and restoration with config injection.
type BackupService struct {
	mgr    *config.Manager
	driver db.DbDriver
}

// NewBackupService creates a BackupService.
func NewBackupService(mgr *config.Manager, driver db.DbDriver) *BackupService {
	return &BackupService{mgr: mgr, driver: driver}
}

// CreateFullBackup creates a zip archive containing the database and any
// configured backup paths. It returns the file path, size in bytes, or an error.
func (s *BackupService) CreateFullBackup() (path string, sizeBytes int64, err error) {
	cfg, err := s.mgr.Config()
	if err != nil {
		return "", 0, fmt.Errorf("load config: %w", err)
	}
	return backup.CreateFullBackup(*cfg, s.driver)
}

// ReadManifest reads and returns the backup manifest from a backup archive.
func (s *BackupService) ReadManifest(backupPath string) (*backup.BackupManifest, error) {
	return backup.ReadManifest(backupPath)
}
