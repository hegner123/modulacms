package utility

import (
	"strings"
	"testing"
)

// ============================================================
// GetVersionInfo
// ============================================================

func TestGetVersionInfo(t *testing.T) {
	t.Parallel()
	got := GetVersionInfo()
	if got != Version {
		t.Errorf("GetVersionInfo() = %q, want %q (same as Version var)", got, Version)
	}
}

// ============================================================
// GetFullVersionInfo
// ============================================================

func TestGetFullVersionInfo(t *testing.T) {
	t.Parallel()
	got := GetFullVersionInfo()

	if !strings.Contains(got, "Version: "+Version) {
		t.Errorf("GetFullVersionInfo() missing version line, got:\n%s", got)
	}
	if !strings.Contains(got, "Commit: "+GitCommit) {
		t.Errorf("GetFullVersionInfo() missing commit line, got:\n%s", got)
	}
	if !strings.Contains(got, "Built: "+BuildDate) {
		t.Errorf("GetFullVersionInfo() missing build date line, got:\n%s", got)
	}
}

// ============================================================
// GetVersion
// ============================================================

func TestGetVersion(t *testing.T) {
	t.Parallel()
	got, err := GetVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("GetVersion() returned nil pointer")
	}
	if !strings.Contains(*got, Version) {
		t.Errorf("GetVersion() = %q, expected it to contain %q", *got, Version)
	}
}

// ============================================================
// GetCurrentVersion
// ============================================================

func TestGetCurrentVersion(t *testing.T) {
	t.Parallel()
	got := GetCurrentVersion()
	if got != Version {
		t.Errorf("GetCurrentVersion() = %q, want %q", got, Version)
	}
}

// ============================================================
// IsDevBuild
// ============================================================

func TestIsDevBuild(t *testing.T) {
	// NOT parallel: this test mutates package-level Version and GitCommit vars,
	// which are read by other version tests running in the parallel group.

	// Save originals and restore after test
	origVersion := Version
	origCommit := GitCommit
	t.Cleanup(func() {
		Version = origVersion
		GitCommit = origCommit
	})

	tests := []struct {
		name      string
		version   string
		gitCommit string
		want      bool
	}{
		{name: "default dev values", version: "dev", gitCommit: "unknown", want: true},
		{name: "version is dev", version: "dev", gitCommit: "abc123", want: true},
		{name: "commit is unknown", version: "1.0.0", gitCommit: "unknown", want: true},
		{name: "both set to real values", version: "1.0.0", gitCommit: "abc123", want: false},
		{name: "empty version is not dev", version: "", gitCommit: "abc123", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Not parallel because we modify package-level vars
			Version = tt.version
			GitCommit = tt.gitCommit
			got := IsDevBuild()
			if got != tt.want {
				t.Errorf("IsDevBuild() with Version=%q, GitCommit=%q = %v, want %v",
					tt.version, tt.gitCommit, got, tt.want)
			}
		})
	}
}
