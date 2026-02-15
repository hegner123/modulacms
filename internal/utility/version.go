package utility

// Version, GitCommit, and BuildDate are set at build time and contain version information about the application.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// GetVersionInfo returns the application version string.
func GetVersionInfo() string {
	return Version
}

// GetFullVersionInfo returns a formatted string containing the version, git commit, and build date.
func GetFullVersionInfo() string {
	return "Version: " + Version + "\nCommit: " + GitCommit + "\nBuilt: " + BuildDate
}

// GetVersion returns a pointer to the full version information string.
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
