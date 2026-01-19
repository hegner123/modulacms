package utility

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func GetVersionInfo() string {
	return Version
}

func GetFullVersionInfo() string {
	return "Version: " + Version + "\nCommit: " + GitCommit + "\nBuilt: " + BuildDate
}

func GetVersion() (*string, error) {
	version := GetFullVersionInfo()
	return &version, nil
}

// GetCurrentVersion returns just the version string (e.g., "0.1.0")
func GetCurrentVersion() string {
	return Version
}

// IsDevBuild checks if running a development build
func IsDevBuild() bool {
	return Version == "dev" || GitCommit == "unknown"
}
