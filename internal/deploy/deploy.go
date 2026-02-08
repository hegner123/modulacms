package deploy

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
	// TODO: Integrate with backup.RestoreFromBackup(cfg, path) when deploy is implemented
	return nil
}
func SyncFromProd() error {
	IssueMakeBackup()
	DownloadBackup()
	// TODO: Integrate with backup.RestoreFromBackup(cfg, path) when deploy is implemented
	return nil
}
