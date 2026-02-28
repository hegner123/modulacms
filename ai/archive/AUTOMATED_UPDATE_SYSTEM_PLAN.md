# Implementation Plan: Automated Update System for ModulaCMS

**Date:** 2026-01-16
**Status:** In Progress (Phases 1-3 Complete)

## Overview

Add a fully automated update system that:
1. Checks GitHub Releases API for new versions
2. Compares semantic versions
3. Shows user confirmation dialog OR auto-updates based on config
4. Downloads and applies updates seamlessly
5. Handles rollback on failure

## Configuration Options (New Fields)

Add to `internal/config/config.go`:
```go
Update_Auto_Enabled    bool   `json:"update_auto_enabled"`     // Auto-update without confirmation (default: false)
Update_Check_Interval  string `json:"update_check_interval"`   // "startup", "daily", "weekly", "never" (default: "startup")
Update_Channel         string `json:"update_channel"`          // "stable", "prerelease" (default: "stable")
Update_Notify_Only     bool   `json:"update_notify_only"`      // Show notification but don't install (default: false)
```

## Implementation Phases

### Phase 1: GitHub Release Checker ✅ COMPLETE

**New File: `internal/update/checker.go`**

Core functions:
- `CheckForUpdates(currentVersion string) (*ReleaseInfo, bool, error)`
  - Queries `https://api.github.com/repos/hegner123/modulacms/releases/latest`
  - Returns release info and whether update is available
- `CompareVersions(current, latest string) int`
  - Semantic version comparison (strip "v", split by ".", compare as integers)
- `GetDownloadURL(release *ReleaseInfo, goos, goarch string) (string, error)`
  - Selects correct binary based on OS/Arch from release assets

Structs:
```go
type ReleaseInfo struct {
    TagName     string  `json:"tag_name"`
    Name        string  `json:"name"`
    Body        string  `json:"body"`
    Assets      []Asset `json:"assets"`
}

type Asset struct {
    Name               string `json:"name"`
    BrowserDownloadURL string `json:"browser_download_url"`
}
```

**Enhance: `internal/utility/version.go`** ✅ COMPLETE
- Add `GetCurrentVersion() string` - returns just Version
- Add `IsDevBuild() bool` - checks if Version == "dev"

### Phase 2: Configuration Support ✅ COMPLETE

**Modify: `internal/config/config.go`** ✅ COMPLETE
- Add 4 update fields to Config struct (see above)

**Modify: `internal/config/default.go`** ✅ COMPLETE
- Set defaults in `DefaultConfig()` function

**Modify: `internal/config/env_provider.go`** ✅ COMPLETE
- Add environment variable parsing for update settings

### Phase 3: Update Download & Application ✅ COMPLETE

**Refactor: `internal/update/update.go`** ✅ COMPLETE

Split existing `Fetch()` into:
- `DownloadUpdate(url string) (tempPath string, error)`
  - Downloads binary to temp file
  - Validates size and permissions
- `ApplyUpdate(tempPath string) error`
  - Creates .bak backup
  - Replaces current binary atomically
  - Rolls back on failure
- `VerifyBinary(path string) error`
  - Checks file is executable and reasonable size

### Phase 4: CLI/TUI Integration ⏸️ BLOCKED (Waiting on CMS Content Creation)

**Add: `internal/cli/message_types.go`**

New message types:
- `UpdateCheckStartMsg` - Trigger update check
- `UpdateCheckCompleteMsg` - Check result with ReleaseInfo
- `UpdateConfirmMsg` - User/auto confirmed update
- `UpdateDeclineMsg` - User declined
- `UpdateDownloadCompleteMsg` - Download finished
- `UpdateApplyCompleteMsg` - Update applied

**Add: `internal/cli/constructors.go`**

Constructors for all update messages following existing patterns (tea.Cmd closures)

**New File: `internal/cli/update_version.go`**

Handler function:
- `UpdateVersion(msg tea.Msg) (Model, tea.Cmd)`
  - Handles all update messages
  - Orchestrates async flow: check → confirm → download → apply
  - Respects config settings (auto vs manual, notify-only)
  - Shows dialogs or auto-updates based on config

Helper command functions:
- `checkForUpdateCmd(cfg *config.Config) tea.Cmd` - Async GitHub API call
- `downloadUpdateCmd(release *ReleaseInfo) tea.Cmd` - Async download
- `applyUpdateCmd(tempPath string) tea.Cmd` - Apply binary update

**Modify: `internal/cli/dialog.go`**

Extend `DialogAction` enum:
```go
const (
    DIALOGUPDATE  DialogAction = "update"
    DIALOGRESTART DialogAction = "restart"
)
```

Add helper functions:
- `ShowUpdateDialogCmd(release *ReleaseInfo) tea.Cmd`
- `ShowRestartDialogCmd() tea.Cmd`

**Modify: `internal/cli/update_dialog.go`**

Handle new dialog actions:
- `DIALOGUPDATE`: User confirmed → trigger `UpdateConfirmMsg`
- `DIALOGRESTART`: User acknowledged → `tea.Quit`

**Modify: `internal/cli/model.go`**

Add fields to Model struct:
```go
updateReleaseInfo      *update.ReleaseInfo
updateDownloadProgress float64
lastUpdateCheck        time.Time
```

**Modify: `internal/cli/update.go`**

Add to update chain:
```go
if m, cmd := m.UpdateVersion(msg); cmd != nil {
    return m, cmd
}
```

**Modify: `internal/cli/model.go` (InitialModel function)**

Trigger update check on startup if configured:
```go
if c.Update_Check_Interval == "startup" && !utility.IsDevBuild() {
    updateCmd = UpdateCheckStartCmd()
}
return m, tea.Batch(GetTablesCMD(m.Config), updateCmd, ...)
```

### Phase 5: Command-Line Flag Enhancement ✅ COMPLETE

**Modify: `cmd/main.go` (HandleFlagUpdate function)** ✅ COMPLETE

Enhanced `--update` flag behavior:
1. Check GitHub for latest version
2. Compare with current version
3. Download correct binary for OS/Arch
4. Apply update
5. Exit with success/error message

The flag now uses the GitHub Releases API instead of hardcoded URLs.

### Phase 6: Optional - Periodic Checks ⏳ PENDING

Add ticker-based periodic checks (daily/weekly):
- New message type: `UpdateCheckTickMsg`
- Ticker command: `tickUpdateCheck(interval time.Duration)`
- Start ticker in `InitialModel` if interval is "daily" or "weekly"
- Check condition every hour, only trigger update check if enough time passed

## Update Flow Diagrams

### Startup Flow (Manual Confirmation)
```
Application Start
  → Config.Update_Check_Interval == "startup"
  → UpdateCheckStartMsg
  → checkForUpdateCmd() (async GitHub API)
  → UpdateCheckCompleteMsg (update available)
  → Config.Update_Auto_Enabled == false
  → ShowUpdateDialogCmd()
  → User sees dialog with version info
  → User accepts → DialogAcceptMsg
  → UpdateConfirmMsg
  → downloadUpdateCmd() (async)
  → UpdateDownloadCompleteMsg
  → applyUpdateCmd()
  → UpdateApplyCompleteMsg
  → ShowRestartDialogCmd()
  → User acknowledges → tea.Quit
```

### Startup Flow (Auto-Update)
```
Application Start
  → UpdateCheckStartMsg
  → UpdateCheckCompleteMsg (update available)
  → Config.Update_Auto_Enabled == true
  → UpdateConfirmMsg (no dialog)
  → Download → Apply → Restart prompt
```

### Startup Flow (Notify-Only)
```
Application Start
  → UpdateCheckStartMsg
  → UpdateCheckCompleteMsg (update available)
  → Config.Update_Notify_Only == true
  → Show status bar notification
  → User can manually trigger --update flag later
```

## Critical Files to Modify

1. ✅ **`internal/update/update.go`** - Refactor Fetch() into DownloadUpdate() + ApplyUpdate()
2. ✅ **`internal/config/config.go`** - Add 4 update config fields
3. ✅ **`internal/config/default.go`** - Set default values
4. ✅ **`internal/config/env_provider.go`** - Parse environment variables
5. ⏳ **`internal/cli/message_types.go`** - Add 6+ update message types
6. ⏳ **`internal/cli/constructors.go`** - Add message constructors
7. ⏳ **`internal/cli/dialog.go`** - Extend DialogAction enum, add helpers
8. ⏳ **`internal/cli/update_dialog.go`** - Handle update dialog actions
9. ⏳ **`internal/cli/model.go`** - Add fields, trigger check on startup
10. ⏳ **`internal/cli/update.go`** - Add UpdateVersion to handler chain
11. ⏳ **`cmd/main.go`** - Enhance HandleFlagUpdate()

## New Files to Create

1. ✅ **`internal/update/checker.go`** - GitHub API integration, version comparison
2. ⏳ **`internal/cli/update_version.go`** - Update message handler
3. ✅ **`internal/utility/version.go`** - Helper functions added (GetCurrentVersion, IsDevBuild)

## Verification Steps

### 1. Unit Tests
- Test version comparison: `CompareVersions("0.0.1", "0.1.0")` returns -1
- Test GitHub API parsing with mock JSON response
- Test binary URL selection for different OS/Arch combinations

### 2. Manual Testing Scenarios

**Test 1: No Update Available**
- Set version to future version (e.g., v99.0.0)
- Run with `Update_Check_Interval: "startup"`
- Should start normally with no dialog

**Test 2: Update Available - Manual Confirmation**
```json
{
  "update_auto_enabled": false,
  "update_check_interval": "startup",
  "update_notify_only": false
}
```
- Run application
- Should show update dialog with version info
- Accept → should download and apply
- Should prompt to restart

**Test 3: Update Available - Auto-Update**
```json
{
  "update_auto_enabled": true,
  "update_check_interval": "startup"
}
```
- Run application
- Should automatically download and apply without dialog
- Should prompt to restart

**Test 4: Notify-Only Mode**
```json
{
  "update_notify_only": true,
  "update_check_interval": "startup"
}
```
- Run application
- Should show notification but NOT download
- User can manually run `--update` flag

**Test 5: Never Check**
```json
{
  "update_check_interval": "never"
}
```
- Run application
- Should NOT check for updates at all

**Test 6: Manual Update Flag**
```bash
./modulacms-x86 --update
```
- Should check GitHub, download, apply, exit
- Should work even if `update_check_interval: "never"`

**Test 7: Dev Build Skip**
- Run without injecting version (Version == "dev")
- Should NOT check for updates (dev builds don't update)

**Test 8: Network Failure**
- Block access to api.github.com
- Should log error but not interrupt startup
- Application should continue normally

**Test 9: Download Failure**
- Mock GitHub API to return invalid download URL
- Should show error dialog
- Should NOT damage current binary

**Test 10: Rollback**
- Simulate update failure (chmod to remove write permissions)
- Should rollback to .bak file
- Should show error message with rollback confirmation

### 3. End-to-End Testing

1. Create GitHub release v0.1.1 with test binaries
2. Set current version to v0.1.0
3. Run through all update flows above
4. Verify downloaded binary runs correctly after update
5. Verify .bak file is created and cleaned up

## Security Considerations

1. **GitHub API**: Use HTTPS, validate JSON responses
2. **Binary Verification**: Check file size matches asset size before applying
3. **Atomic Operations**: Use os.Rename for atomic replacement
4. **Permissions**: Verify write permissions before attempting update
5. **Rollback**: Keep .bak file until next successful update

## Future Enhancements (Out of Scope)

- Checksum/signature verification for binaries
- Plugin update system
- Differential/delta updates
- Update scheduling
- Full release notes display in TUI

## Success Criteria

✅ Updates check on startup (configurable)
✅ Dialog appears when update available (unless auto-update enabled)
✅ Auto-update works when configured
✅ Manual `--update` flag works
✅ Rollback mechanism works on failure
✅ No disruption to running servers during check
✅ Clear error messages on all failure modes
✅ Dev builds skip update checks

## Implementation Progress

### Completed (2026-01-16)
- ✅ Phase 1: GitHub Release Checker
  - Created `checker.go` with GitHub API integration
  - Version comparison logic
  - Binary URL selection by platform
- ✅ Phase 2: Configuration Support
  - Added 4 config fields
  - Set sensible defaults
  - Environment variable support
- ✅ Phase 3: Update Download & Application
  - Refactored `update.go`
  - Separated download and apply steps
  - Added binary verification
  - Added rollback functionality
- ✅ Phase 5: Command-Line Flag Enhancement
  - Enhanced `HandleFlagUpdate()` in `cmd/main.go`
  - Now uses GitHub Releases API
  - Automatic platform detection
  - Clear progress logging
  - Exit codes for success/failure

### Blocked
- ⏸️ Phase 4: CLI/TUI Integration
  - **Blocked on**: CMS Content Creation implementation
  - **Reason**: Requires message system patterns from active refactor work
  - **Dependencies**: Phase 1 Core CMS Content Creation (90% complete)
  - Will implement after current refactor completes

### Remaining Work
- Phase 6: Periodic checks (optional)
- Testing and verification
- Create test GitHub release for validation

## Notes

- Implementation follows ModulaCMS patterns (Bubbletea architecture, message-based async)
- Backward compatible: legacy `Fetch()` function preserved
- Security-focused: binary verification, rollback on failure
- User-friendly: configurable behavior, clear error messages
