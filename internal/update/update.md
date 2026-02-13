# update

Package update provides self-update functionality for ModulaCMS binaries. It handles checking for new releases on GitHub, downloading platform-specific binaries, and applying updates with automatic rollback on failure.

## Overview

The update package interacts with the GitHub Releases API to detect and install new versions. It supports stable and prerelease channels, semantic version comparison, and platform-aware binary selection. Update operations include verification checks to ensure downloaded binaries are valid before replacement.

## Constants

Package-level constants define GitHub API endpoints and HTTP client configuration.

#### githubAPIURL

The GitHub API endpoint for fetching the latest ModulaCMS release. Points to the latest release endpoint for the hegner123/modulacms repository.

#### userAgent

HTTP User-Agent header value set to ModulaCMS for GitHub API requests.

## Types

### ReleaseInfo

ReleaseInfo represents a GitHub release with metadata and downloadable assets.

Fields include TagName for the version tag, Name for the release title, Body for release notes, Draft and Prerelease boolean flags, PublishedAt timestamp string, and Assets slice containing downloadable binaries.

### Asset

Asset represents a single downloadable file attached to a GitHub release.

Fields include Name for the filename, BrowserDownloadURL for the direct download link, and Size for the file size in bytes.

## Functions

### CheckForUpdates

```go
func CheckForUpdates(currentVersion string, channel string) (*ReleaseInfo, bool, error)
```

CheckForUpdates queries the GitHub API for the latest release and compares it with the current version. Returns a ReleaseInfo pointer, a boolean indicating whether an update is available, and an error.

The channel parameter controls which releases are considered: stable channel ignores prereleases, while other values include all non-draft releases. Draft releases are always skipped.

Version comparison uses semantic versioning. Returns true if the latest release is newer than currentVersion.

### CompareVersions

```go
func CompareVersions(current, latest string) int
```

CompareVersions compares two semantic version strings. Returns -1 if current is older, 0 if equal, and 1 if current is newer.

Both version strings have any leading v prefix stripped. Special handling treats dev and unknown as very old versions. Comparison proceeds part-by-part splitting on dots, with each numeric component compared individually. Pre-release tags are handled by splitting on hyphens and comparing only the numeric prefix.

### GetDownloadURL

```go
func GetDownloadURL(release *ReleaseInfo, goos, goarch string) (string, error)
```

GetDownloadURL finds the download URL for a binary matching the specified operating system and architecture. Returns the browser download URL from the matching asset or an error if no compatible binary exists.

Binary naming convention is modulacms-GOOS-GOARCH. Searches through all assets in the release to find an exact name match.

### GetDownloadURLForCurrentPlatform

```go
func GetDownloadURLForCurrentPlatform(release *ReleaseInfo) (string, error)
```

GetDownloadURLForCurrentPlatform returns the download URL for the current runtime platform. Calls GetDownloadURL with runtime.GOOS and runtime.GOARCH.

### DownloadUpdate

```go
func DownloadUpdate(url string) (string, error)
```

DownloadUpdate downloads a binary from the provided URL to a temporary file. Returns the path to the temporary file and an error.

Creates a temporary file with the update prefix, downloads the binary via HTTP GET, sets executable permissions to 0755, and verifies the binary before returning. If any step fails, the temporary file is removed and an error is returned.

Verification ensures the downloaded file is a valid executable before allowing further processing.

### ApplyUpdate

```go
func ApplyUpdate(tempPath string) error
```

ApplyUpdate replaces the currently running binary with the downloaded version at tempPath. Returns an error if the update fails.

Determines the current executable path by resolving symlinks, creates a backup with a .bak extension, renames the new binary into place, and removes the backup on success.

If renaming the new binary fails, automatically rolls back by restoring the backup. Returns an error describing both the update failure and any rollback failure if rollback also fails.

### VerifyBinary

```go
func VerifyBinary(path string) error
```

VerifyBinary performs sanity checks on a binary file. Returns an error if the file fails validation.

Checks that file size is between 1MB and 500MB to detect corrupt or suspicious binaries. Verifies executable permissions are set. Does not perform cryptographic signature validation.

### RollbackUpdate

```go
func RollbackUpdate() error
```

RollbackUpdate restores the previous binary version from a backup file. Returns an error if no backup exists or restoration fails.

Looks for a .bak file adjacent to the current executable and renames it back to the original executable path. This function is useful for manual recovery after a failed update.

### Fetch

```go
func Fetch(url string) error
```

Fetch is a legacy function combining DownloadUpdate and ApplyUpdate into a single call. Deprecated in favor of calling the functions separately for better error handling and control.

Downloads the binary from url, applies the update, and cleans up the temporary file. Returns an error if either download or apply fails.
