// Black-box tests for the update package.
//
// Pure functions (CompareVersions, GetDownloadURL, VerifyBinary) are tested
// exhaustively with table-driven tests. Functions that perform HTTP calls
// (DownloadUpdate, CheckForUpdates) use httptest.Server. Filesystem operations
// (ApplyUpdateTo, RollbackUpdateTo) are tested with temp directory simulations
// covering direct binary, symlink, read-only, and rollback scenarios.
package update_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/update"
)

// ---------------------------------------------------------------------------
// CompareVersions
// ---------------------------------------------------------------------------

func TestCompareVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		// Equal versions
		{name: "equal simple", current: "1.0.0", latest: "1.0.0", want: 0},
		{name: "equal with v prefix on both", current: "v1.2.3", latest: "v1.2.3", want: 0},
		{name: "equal mixed v prefix", current: "v1.2.3", latest: "1.2.3", want: 0},
		{name: "equal two parts", current: "1.2", latest: "1.2", want: 0},
		{name: "equal single part", current: "5", latest: "5", want: 0},

		// Current is older (returns -1)
		{name: "older major", current: "1.0.0", latest: "2.0.0", want: -1},
		{name: "older minor", current: "1.1.0", latest: "1.2.0", want: -1},
		{name: "older patch", current: "1.2.3", latest: "1.2.4", want: -1},
		{name: "older with v prefix", current: "v0.9.0", latest: "v1.0.0", want: -1},
		{name: "older fewer parts", current: "1.0", latest: "1.0.1", want: -1},

		// Current is newer (returns 1)
		{name: "newer major", current: "3.0.0", latest: "2.0.0", want: 1},
		{name: "newer minor", current: "1.5.0", latest: "1.4.0", want: 1},
		{name: "newer patch", current: "1.2.4", latest: "1.2.3", want: 1},
		{name: "newer with v prefix", current: "v2.0.0", latest: "v1.9.9", want: 1},

		// Special versions
		{name: "dev is always old", current: "dev", latest: "0.0.1", want: -1},
		{name: "unknown is always old", current: "unknown", latest: "0.0.1", want: -1},
		{name: "dev vs high version", current: "dev", latest: "99.99.99", want: -1},

		// Pre-release tags (hyphen-separated suffix stripped to numeric prefix)
		{name: "prerelease equal numeric", current: "1.2.3-beta", latest: "1.2.3-rc1", want: 0},
		{name: "prerelease older", current: "1.2.3-alpha", latest: "1.2.4-beta", want: -1},
		{name: "prerelease newer", current: "1.3.0-rc1", latest: "1.2.9-beta", want: 1},

		// Asymmetric part counts
		{name: "shorter current equal prefix", current: "1.2", latest: "1.2.0", want: 0},
		{name: "shorter latest equal prefix", current: "1.2.0", latest: "1.2", want: 0},
		{name: "shorter current less", current: "1", latest: "1.0.1", want: -1},
		{name: "shorter latest less", current: "1.0.1", latest: "1", want: 1},

		// Zero values
		{name: "both zero", current: "0.0.0", latest: "0.0.0", want: 0},
		{name: "zero vs nonzero", current: "0.0.0", latest: "0.0.1", want: -1},

		// Large version numbers
		{name: "large numbers", current: "100.200.300", latest: "100.200.301", want: -1},

		// Non-numeric parts default to 0 from Atoi failure
		{name: "non-numeric part treated as zero", current: "1.abc.3", latest: "1.0.3", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := update.CompareVersions(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GetDownloadURL
// ---------------------------------------------------------------------------

func TestGetDownloadURL(t *testing.T) {
	t.Parallel()

	release := &update.ReleaseInfo{
		TagName: "v1.0.0",
		Assets: []update.Asset{
			{Name: "modulacms-linux-amd64", BrowserDownloadURL: "https://example.com/linux-amd64", Size: 10000000},
			{Name: "modulacms-darwin-arm64", BrowserDownloadURL: "https://example.com/darwin-arm64", Size: 12000000},
			{Name: "modulacms-darwin-amd64", BrowserDownloadURL: "https://example.com/darwin-amd64", Size: 11000000},
			{Name: "modulacms-windows-amd64", BrowserDownloadURL: "https://example.com/windows-amd64", Size: 13000000},
		},
	}

	tests := []struct {
		name    string
		release *update.ReleaseInfo
		goos    string
		goarch  string
		wantURL string
		wantErr string
	}{
		{
			name:    "linux amd64 match",
			release: release,
			goos:    "linux",
			goarch:  "amd64",
			wantURL: "https://example.com/linux-amd64",
		},
		{
			name:    "darwin arm64 match",
			release: release,
			goos:    "darwin",
			goarch:  "arm64",
			wantURL: "https://example.com/darwin-arm64",
		},
		{
			name:    "darwin amd64 match",
			release: release,
			goos:    "darwin",
			goarch:  "amd64",
			wantURL: "https://example.com/darwin-amd64",
		},
		{
			name:    "windows amd64 match",
			release: release,
			goos:    "windows",
			goarch:  "amd64",
			wantURL: "https://example.com/windows-amd64",
		},
		{
			name:    "no match for platform",
			release: release,
			goos:    "freebsd",
			goarch:  "arm",
			wantErr: "no compatible binary found for freebsd/arm",
		},
		{
			name:    "nil release",
			release: nil,
			goos:    "linux",
			goarch:  "amd64",
			wantErr: "release info is nil",
		},
		{
			name: "empty assets",
			release: &update.ReleaseInfo{
				TagName: "v1.0.0",
				Assets:  []update.Asset{},
			},
			goos:    "linux",
			goarch:  "amd64",
			wantErr: "no compatible binary found for linux/amd64",
		},
		{
			name: "nil assets slice",
			release: &update.ReleaseInfo{
				TagName: "v1.0.0",
				Assets:  nil,
			},
			goos:    "linux",
			goarch:  "amd64",
			wantErr: "no compatible binary found for linux/amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := update.GetDownloadURL(tt.release, tt.goos, tt.goarch)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantURL {
				t.Errorf("got URL %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestGetDownloadURLForCurrentPlatform(t *testing.T) {
	t.Parallel()

	// Build a release with an asset matching the current platform so we get a match.
	// We use runtime.GOOS/GOARCH directly since GetDownloadURLForCurrentPlatform
	// delegates to GetDownloadURL with those exact values.
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	expectedURL := fmt.Sprintf("https://example.com/%s-%s", goos, goarch)
	release := &update.ReleaseInfo{
		TagName: "v2.0.0",
		Assets: []update.Asset{
			{
				Name:               fmt.Sprintf("modulacms-%s-%s", goos, goarch),
				BrowserDownloadURL: expectedURL,
				Size:               10000000,
			},
			// Include a non-matching asset to verify correct selection
			{
				Name:               "modulacms-plan9-mips",
				BrowserDownloadURL: "https://example.com/should-not-match",
				Size:               10000000,
			},
		},
	}

	got, err := update.GetDownloadURLForCurrentPlatform(release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != expectedURL {
		t.Errorf("got URL %q, want %q", got, expectedURL)
	}
}

func TestGetDownloadURLForCurrentPlatform_NoMatch(t *testing.T) {
	t.Parallel()

	// Release with no assets matching any real platform
	release := &update.ReleaseInfo{
		TagName: "v2.0.0",
		Assets: []update.Asset{
			{Name: "modulacms-plan9-mips", BrowserDownloadURL: "https://example.com/plan9", Size: 10000000},
		},
	}

	_, err := update.GetDownloadURLForCurrentPlatform(release)
	if err == nil {
		t.Fatal("expected error for non-matching platform, got nil")
	}
	if !strings.Contains(err.Error(), "no compatible binary found") {
		t.Errorf("expected error about no compatible binary, got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// VerifyBinary
// ---------------------------------------------------------------------------

func TestVerifyBinary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns path to file
		wantErr string
	}{
		{
			name: "valid binary above 1MB with exec permissions",
			setup: func(t *testing.T) string {
				t.Helper()
				return createTempFile(t, 2*1024*1024, 0755) // 2MB, executable
			},
		},
		{
			name: "file too small",
			setup: func(t *testing.T) string {
				t.Helper()
				return createTempFile(t, 100, 0755) // 100 bytes
			},
			wantErr: "binary too small",
		},
		{
			name: "exactly 1MB minus 1 byte is too small",
			setup: func(t *testing.T) string {
				t.Helper()
				return createTempFile(t, 1024*1024-1, 0755)
			},
			wantErr: "binary too small",
		},
		{
			name: "exactly 1MB is valid",
			setup: func(t *testing.T) string {
				t.Helper()
				return createTempFile(t, 1024*1024, 0755)
			},
		},
		{
			name: "not executable",
			setup: func(t *testing.T) string {
				t.Helper()
				return createTempFile(t, 2*1024*1024, 0644) // no exec bits
			},
			wantErr: "binary is not executable",
		},
		{
			name: "nonexistent file",
			setup: func(t *testing.T) string {
				t.Helper()
				return filepath.Join(t.TempDir(), "does-not-exist")
			},
			wantErr: "failed to stat binary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.setup(t)
			err := update.VerifyBinary(path)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// We skip the "binary too large" test because creating a 500MB+ file in
// tests is wasteful. The boundary is well-defined and the logic is
// straightforward (single comparison).

// ---------------------------------------------------------------------------
// DownloadUpdate
// ---------------------------------------------------------------------------

func TestDownloadUpdate_HTTPErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "server returns 404",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: "bad status downloading update",
		},
		{
			name: "server returns 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: "bad status downloading update",
		},
		{
			name: "server returns empty 200 body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				// Empty body means zero-size file, which will fail VerifyBinary
			},
			// File will be created but verification will fail because it's too small
			wantErr: "binary verification failed",
		},
		{
			name: "server returns small body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not a real binary"))
			},
			// Small file fails VerifyBinary size check
			wantErr: "binary verification failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)

			_, err := update.DownloadUpdate(srv.URL)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestDownloadUpdate_InvalidURL(t *testing.T) {
	t.Parallel()

	_, err := update.DownloadUpdate("http://localhost:0/nonexistent")
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
	if !strings.Contains(err.Error(), "failed to download update") {
		t.Errorf("expected 'failed to download update' error, got %q", err.Error())
	}
}

func TestDownloadUpdate_SuccessfulDownload(t *testing.T) {
	t.Parallel()

	// Serve a file that is >= 1MB and will pass VerifyBinary.
	// We use a 1MB buffer of zeros plus a small header to exceed the minimum.
	binaryData := make([]byte, 1*1024*1024+1) // 1MB + 1 byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(binaryData)
	}))
	t.Cleanup(srv.Close)

	tmpPath, err := update.DownloadUpdate(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Clean up the downloaded temp file
	t.Cleanup(func() { os.Remove(tmpPath) })

	// Verify the temp file exists and has the right properties
	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("temp file does not exist: %v", err)
	}

	if info.Size() != int64(len(binaryData)) {
		t.Errorf("downloaded file size = %d, want %d", info.Size(), len(binaryData))
	}

	// Check executable permissions
	if info.Mode()&0111 == 0 {
		t.Error("downloaded file is not executable")
	}
}

// ---------------------------------------------------------------------------
// ApplyUpdateTo — tests the filesystem logic for each deployment scenario
// ---------------------------------------------------------------------------

func TestApplyUpdateTo_DirectBinary(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	execPath := filepath.Join(dir, "modula")
	tempPath := filepath.Join(dir, "modula-new")

	os.WriteFile(execPath, []byte("old-binary"), 0755)
	os.WriteFile(tempPath, []byte("new-binary"), 0755)

	if err := update.ApplyUpdateTo(tempPath, execPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the binary was replaced
	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("failed to read updated binary: %v", err)
	}
	if string(got) != "new-binary" {
		t.Errorf("binary content = %q, want %q", got, "new-binary")
	}

	// Verify backup was cleaned up
	if _, err := os.Stat(execPath + ".bak"); !os.IsNotExist(err) {
		t.Error("expected .bak to be removed after successful update")
	}

	// Verify temp file was consumed (moved away)
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("expected temp file to be removed after move")
	}
}

func TestApplyUpdateTo_Symlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	realPath := filepath.Join(dir, "bin", "modula")
	linkPath := filepath.Join(dir, "usr-local-bin", "modula")
	tempPath := filepath.Join(dir, "modula-new")

	// Create directory structure simulating /opt/modula/modula symlinked from /usr/local/bin/modula
	os.MkdirAll(filepath.Join(dir, "bin"), 0755)
	os.MkdirAll(filepath.Join(dir, "usr-local-bin"), 0755)
	os.WriteFile(realPath, []byte("old-binary"), 0755)
	os.Symlink(realPath, linkPath)
	os.WriteFile(tempPath, []byte("new-binary"), 0755)

	// Resolve the symlink (as ApplyUpdate would do)
	resolved, err := filepath.EvalSymlinks(linkPath)
	if err != nil {
		t.Fatalf("failed to resolve symlink: %v", err)
	}

	if err := update.ApplyUpdateTo(tempPath, resolved); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the real file was updated
	got, err := os.ReadFile(realPath)
	if err != nil {
		t.Fatalf("failed to read real path: %v", err)
	}
	if string(got) != "new-binary" {
		t.Errorf("real binary content = %q, want %q", got, "new-binary")
	}

	// Verify the symlink still works and points to the updated content
	gotViaLink, err := os.ReadFile(linkPath)
	if err != nil {
		t.Fatalf("failed to read via symlink: %v", err)
	}
	if string(gotViaLink) != "new-binary" {
		t.Errorf("symlink content = %q, want %q", gotViaLink, "new-binary")
	}
}

func TestApplyUpdateTo_RollbackOnMissingTemp(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	execPath := filepath.Join(dir, "modula")
	os.WriteFile(execPath, []byte("old-binary"), 0755)

	// Point at a temp file that doesn't exist — the move will fail
	// and the original binary should be restored via rollback.
	missingTemp := filepath.Join(dir, "does-not-exist")
	err := update.ApplyUpdateTo(missingTemp, execPath)
	if err == nil {
		t.Fatal("expected error for missing temp file, got nil")
	}
	if !strings.Contains(err.Error(), "rollback successful") {
		t.Errorf("expected rollback message, got %q", err.Error())
	}

	// Verify original binary was restored
	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("original binary should be restored: %v", err)
	}
	if string(got) != "old-binary" {
		t.Errorf("restored binary content = %q, want %q", got, "old-binary")
	}
}

func TestApplyUpdateTo_ReadOnlyDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	execPath := filepath.Join(dir, "modula")
	tempPath := filepath.Join(t.TempDir(), "modula-new")

	os.WriteFile(execPath, []byte("old-binary"), 0755)
	os.WriteFile(tempPath, []byte("new-binary"), 0755)

	// Make the directory read-only so the backup rename fails
	os.Chmod(dir, 0555)
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	err := update.ApplyUpdateTo(tempPath, execPath)
	if err == nil {
		t.Fatal("expected error for read-only directory, got nil")
	}
	if !strings.Contains(err.Error(), "failed to backup") {
		t.Errorf("expected backup error, got %q", err.Error())
	}
}

func TestApplyUpdateTo_PreservesPermissions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	execPath := filepath.Join(dir, "modula")
	tempPath := filepath.Join(dir, "modula-new")

	os.WriteFile(execPath, []byte("old-binary"), 0755)
	os.WriteFile(tempPath, []byte("new-binary"), 0755)

	if err := update.ApplyUpdateTo(tempPath, execPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(execPath)
	if err != nil {
		t.Fatalf("failed to stat updated binary: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("updated binary lost executable permissions")
	}
}

// ---------------------------------------------------------------------------
// RollbackUpdateTo
// ---------------------------------------------------------------------------

func TestRollbackUpdateTo_Success(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	execPath := filepath.Join(dir, "modula")
	backupPath := execPath + ".bak"

	// Simulate state after a failed update: .bak exists, current is broken
	os.WriteFile(backupPath, []byte("good-binary"), 0755)
	os.WriteFile(execPath, []byte("broken-binary"), 0755)

	if err := update.RollbackUpdateTo(execPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatalf("failed to read restored binary: %v", err)
	}
	if string(got) != "good-binary" {
		t.Errorf("restored content = %q, want %q", got, "good-binary")
	}

	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Error("expected .bak to be consumed by rollback")
	}
}

func TestRollbackUpdateTo_NoBackup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	execPath := filepath.Join(dir, "modula")
	os.WriteFile(execPath, []byte("current-binary"), 0755)

	err := update.RollbackUpdateTo(execPath)
	if err == nil {
		t.Fatal("expected error when no backup exists, got nil")
	}
	if !strings.Contains(err.Error(), "no backup file found") {
		t.Errorf("expected 'no backup file found' error, got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// CheckForUpdates — end-to-end tests using httptest servers via GitHubAPIURL
// ---------------------------------------------------------------------------

// setTestAPIURL points GitHubAPIURL at the given httptest server and returns
// a cleanup function that restores the original value.
func setTestAPIURL(t *testing.T, url string) {
	t.Helper()
	original := update.GitHubAPIURL
	update.GitHubAPIURL = url
	t.Cleanup(func() { update.GitHubAPIURL = original })
}

func TestCheckForUpdates_NetworkError(t *testing.T) {
	// Point at a URL that will refuse connections.
	setTestAPIURL(t, "http://127.0.0.1:1/releases/latest")

	_, available, err := update.CheckForUpdates("1.0.0", "stable")
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
	if available {
		t.Error("expected available=false on error")
	}
	if !strings.Contains(err.Error(), "could not reach GitHub") {
		t.Errorf("expected network error message, got %q", err.Error())
	}
}

func TestCheckForUpdates_NoReleases(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	release, available, err := update.CheckForUpdates("1.0.0", "stable")
	if err != nil {
		t.Fatalf("expected no error for 404, got %v", err)
	}
	if available {
		t.Error("expected available=false when no releases exist")
	}
	if release != nil {
		t.Error("expected nil release when no releases exist")
	}
}

func TestCheckForUpdates_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	_, _, err := update.CheckForUpdates("1.0.0", "stable")
	if err == nil {
		t.Fatal("expected error for 403, got nil")
	}
	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected rate limit error, got %q", err.Error())
	}
}

func TestCheckForUpdates_UnexpectedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	_, _, err := update.CheckForUpdates("1.0.0", "stable")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Errorf("expected unexpected status error, got %q", err.Error())
	}
}

func TestCheckForUpdates_UpdateAvailable(t *testing.T) {
	release := update.ReleaseInfo{
		TagName:     "v2.0.0",
		Name:        "Release 2.0.0",
		PublishedAt: "2026-01-01T00:00:00Z",
		Assets: []update.Asset{
			{Name: "modulacms-linux-amd64", BrowserDownloadURL: "https://example.com/linux-amd64", Size: 10000000},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	got, available, err := update.CheckForUpdates("1.0.0", "stable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected available=true when current version is older")
	}
	if got == nil {
		t.Fatal("expected non-nil release")
	}
	if got.TagName != "v2.0.0" {
		t.Errorf("got tag %q, want %q", got.TagName, "v2.0.0")
	}
}

func TestCheckForUpdates_AlreadyUpToDate(t *testing.T) {
	release := update.ReleaseInfo{
		TagName:     "v1.0.0",
		Name:        "Release 1.0.0",
		PublishedAt: "2026-01-01T00:00:00Z",
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	_, available, err := update.CheckForUpdates("1.0.0", "stable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected available=false when versions are equal")
	}
}

func TestCheckForUpdates_DraftSkipped(t *testing.T) {
	release := update.ReleaseInfo{
		TagName: "v2.0.0",
		Draft:   true,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	got, available, err := update.CheckForUpdates("1.0.0", "stable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected available=false for draft release")
	}
	if got != nil {
		t.Error("expected nil release for draft")
	}
}

func TestCheckForUpdates_StableChannelSkipsPrerelease(t *testing.T) {
	release := update.ReleaseInfo{
		TagName:    "v2.0.0-beta.1",
		Prerelease: true,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	got, available, err := update.CheckForUpdates("1.0.0", "stable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected available=false for prerelease on stable channel")
	}
	if got != nil {
		t.Error("expected nil release for filtered prerelease")
	}
}

func TestCheckForUpdates_PrereleaseChannelIncludesPrerelease(t *testing.T) {
	release := update.ReleaseInfo{
		TagName:    "v2.0.0-beta.1",
		Prerelease: true,
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	}))
	t.Cleanup(srv.Close)
	setTestAPIURL(t, srv.URL)

	got, available, err := update.CheckForUpdates("1.0.0", "beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected available=true for prerelease on non-stable channel")
	}
	if got == nil {
		t.Fatal("expected non-nil release")
	}
}

// ---------------------------------------------------------------------------
// Fetch (legacy function)
// ---------------------------------------------------------------------------

func TestFetch_InvalidURL(t *testing.T) {
	t.Parallel()

	err := update.Fetch("http://localhost:0/nonexistent")
	if err == nil {
		t.Fatal("expected error for unreachable URL, got nil")
	}
	if !strings.Contains(err.Error(), "failed to download update") {
		t.Errorf("expected 'failed to download update' error, got %q", err.Error())
	}
}

func TestFetch_ServerReturnsError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	err := update.Fetch(srv.URL)
	if err == nil {
		t.Fatal("expected error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "bad status downloading update") {
		t.Errorf("expected 'bad status downloading update' error, got %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// ReleaseInfo and Asset JSON marshaling
// ---------------------------------------------------------------------------

func TestReleaseInfo_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := update.ReleaseInfo{
		TagName:     "v1.5.0",
		Name:        "Release 1.5.0",
		Body:        "Bug fixes and improvements",
		Draft:       false,
		Prerelease:  true,
		PublishedAt: "2025-01-15T10:00:00Z",
		Assets: []update.Asset{
			{
				Name:               "modulacms-linux-amd64",
				BrowserDownloadURL: "https://github.com/example/download",
				Size:               15000000,
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal ReleaseInfo: %v", err)
	}

	var decoded update.ReleaseInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal ReleaseInfo: %v", err)
	}

	// Verify all fields survive the round trip
	if decoded.TagName != original.TagName {
		t.Errorf("TagName = %q, want %q", decoded.TagName, original.TagName)
	}
	if decoded.Name != original.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, original.Name)
	}
	if decoded.Body != original.Body {
		t.Errorf("Body = %q, want %q", decoded.Body, original.Body)
	}
	if decoded.Draft != original.Draft {
		t.Errorf("Draft = %v, want %v", decoded.Draft, original.Draft)
	}
	if decoded.Prerelease != original.Prerelease {
		t.Errorf("Prerelease = %v, want %v", decoded.Prerelease, original.Prerelease)
	}
	if decoded.PublishedAt != original.PublishedAt {
		t.Errorf("PublishedAt = %q, want %q", decoded.PublishedAt, original.PublishedAt)
	}
	if len(decoded.Assets) != len(original.Assets) {
		t.Fatalf("Assets length = %d, want %d", len(decoded.Assets), len(original.Assets))
	}
	if decoded.Assets[0].Name != original.Assets[0].Name {
		t.Errorf("Asset.Name = %q, want %q", decoded.Assets[0].Name, original.Assets[0].Name)
	}
	if decoded.Assets[0].BrowserDownloadURL != original.Assets[0].BrowserDownloadURL {
		t.Errorf("Asset.BrowserDownloadURL = %q, want %q", decoded.Assets[0].BrowserDownloadURL, original.Assets[0].BrowserDownloadURL)
	}
	if decoded.Assets[0].Size != original.Assets[0].Size {
		t.Errorf("Asset.Size = %d, want %d", decoded.Assets[0].Size, original.Assets[0].Size)
	}
}

func TestReleaseInfo_JSONFieldNames(t *testing.T) {
	t.Parallel()

	// Verify that JSON tags produce the expected GitHub API field names.
	release := update.ReleaseInfo{
		TagName:     "v1.0.0",
		Prerelease:  true,
		PublishedAt: "2025-01-01T00:00:00Z",
		Assets: []update.Asset{
			{
				Name:               "test",
				BrowserDownloadURL: "https://example.com",
				Size:               100,
			},
		},
	}

	data, err := json.Marshal(release)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(data)

	// These field names must match the GitHub API response format
	expectedFields := []string{
		`"tag_name"`,
		`"name"`,
		`"body"`,
		`"draft"`,
		`"prerelease"`,
		`"published_at"`,
		`"assets"`,
		`"browser_download_url"`,
		`"size"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON output missing expected field %s in: %s", field, jsonStr)
		}
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// createTempFile creates a file of the given size with the given permissions
// in a test-owned temp directory. Returns the file path.
func createTempFile(t *testing.T, size int64, perm os.FileMode) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "testbinary")

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Write zeros to reach desired size
	if size > 0 {
		if err := f.Truncate(size); err != nil {
			f.Close()
			t.Fatalf("failed to truncate file to %d bytes: %v", size, err)
		}
	}

	if err := f.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	if err := os.Chmod(path, perm); err != nil {
		t.Fatalf("failed to chmod temp file: %v", err)
	}

	return path
}
