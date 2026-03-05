package tui

// OpenFilePickerForRestoreMsg requests opening the file picker for backup restoration.
type OpenFilePickerForRestoreMsg struct{}

// RestoreBackupFromPathMsg requests restoring a backup from a file path.
type RestoreBackupFromPathMsg struct{ Path string }

// BackupRestoreCompleteMsg signals successful backup restoration.
type BackupRestoreCompleteMsg struct{ Path string }
