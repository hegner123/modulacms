// Package deploy provides utilities for backing up and syncing ModulaCMS data
// between development and production environments.
package deploy

// IssueMakeBackup creates a backup on a remote server.
func IssueMakeBackup() {}

// DownloadBackup retrieves a backup file from a remote server.
func DownloadBackup() {}

/*
func SyncToDev()  {
    //backup.CreateBackup()
    backupZip := ""

}
func SyncToProd() {
    //backup.CreateBackup()
    backupZip := ""

}
*/

// SyncFromDev pulls a backup from the development environment and restores it locally.
func SyncFromDev() error {
	IssueMakeBackup()
	DownloadBackup()
	// TODO: Integrate with backup.RestoreFromBackup(cfg, path) when deploy is implemented
	return nil
}

// SyncFromProd pulls a backup from the production environment and restores it locally.
func SyncFromProd() error {
	IssueMakeBackup()
	DownloadBackup()
	// TODO: Integrate with backup.RestoreFromBackup(cfg, path) when deploy is implemented
	return nil
}
