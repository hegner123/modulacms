// Package update provides functionality for checking and downloading ModulaCMS updates
// from the GitHub release API, including version comparison and platform-specific binary selection.
package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ReleaseInfo represents a GitHub release
type ReleaseInfo struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	Body        string  `json:"body"`
	Draft       bool    `json:"draft"`
	Prerelease  bool    `json:"prerelease"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
}

// Asset represents a release asset (binary file)
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

const (
	githubAPIURL = "https://api.github.com/repos/hegner123/modulacms/releases/latest"
	userAgent    = "ModulaCMS"
)

// CheckForUpdates queries GitHub API for the latest release and compares with current version
// Returns: (latestRelease, updateAvailable, error)
func CheckForUpdates(currentVersion string, channel string) (*ReleaseInfo, bool, error) {
	// Query GitHub API
	release, err := fetchLatestRelease()
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch release info: %w", err)
	}

	// Skip draft releases
	if release.Draft {
		return nil, false, nil
	}

	// Check release channel (stable vs prerelease)
	if channel == "stable" && release.Prerelease {
		return nil, false, nil
	}

	// Compare versions
	comparison := CompareVersions(currentVersion, release.TagName)
	if comparison < 0 {
		// Current version is older than latest release
		return release, true, nil
	}

	return release, false, nil
}

// fetchLatestRelease queries the GitHub API
func fetchLatestRelease() (*ReleaseInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &release, nil
}

// CompareVersions uses semantic versioning comparison
// Returns: -1 (current < latest), 0 (equal), 1 (current > latest)
func CompareVersions(current, latest string) int {
	// Strip "v" prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	// Handle "dev" builds as very old version
	if current == "dev" || current == "unknown" {
		return -1
	}

	// Split by dots
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Compare each part
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	for i := 0; i < maxLen; i++ {
		currentVal := 0
		latestVal := 0

		if i < len(currentParts) {
			currentVal, _ = strconv.Atoi(strings.Split(currentParts[i], "-")[0]) // Handle pre-release tags
		}

		if i < len(latestParts) {
			latestVal, _ = strconv.Atoi(strings.Split(latestParts[i], "-")[0])
		}

		if currentVal < latestVal {
			return -1
		}
		if currentVal > latestVal {
			return 1
		}
	}

	return 0
}

// GetDownloadURL determines the correct binary URL based on OS/Arch
func GetDownloadURL(release *ReleaseInfo, goos, goarch string) (string, error) {
	if release == nil {
		return "", fmt.Errorf("release info is nil")
	}

	// Build expected binary name
	// Format: modulacms-{goos}-{goarch}
	expectedName := fmt.Sprintf("modulacms-%s-%s", goos, goarch)

	// Search through assets
	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no compatible binary found for %s/%s", goos, goarch)
}

// GetDownloadURLForCurrentPlatform returns the download URL for the current OS/Arch
func GetDownloadURLForCurrentPlatform(release *ReleaseInfo) (string, error) {
	return GetDownloadURL(release, runtime.GOOS, runtime.GOARCH)
}
