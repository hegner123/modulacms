package deploy

import backup "github.com/hegner123/modulacms/internal/Backup"

func IssueMakeBackup() {}

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

func SyncFromDev() error {
	IssueMakeBackup()
	DownloadBackup()
	err := backup.RestoreBackup("", "", "")
	if err != nil {
		return err
	}
	return nil
}
func SyncFromProd() error {
	IssueMakeBackup()
	DownloadBackup()
	err := backup.RestoreBackup("", "", "")
	if err != nil {
		return err
	}
	return nil

}
