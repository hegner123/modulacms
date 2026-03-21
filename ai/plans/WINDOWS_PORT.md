# Windows Port Plan

**Status:** Implemented (all 7 phases complete, pending test verification)
**Estimated effort:** 7-8 files, ~200 lines changed, 7 phases
**Dependencies:** Phases 1-5 are independent (parallelizable). Phase 6 requires 1-5 merged. Phase 7 anytime.

## Revised Assessment

Several items from the initial analysis are **not actually issues**:

- `syscall.SIGTERM` IS defined on Windows (invented constant) -- `signal.Notify` works fine
- SSH host key path `.ssh/id_ed25519` works because Go accepts `/` on Windows
- Registry path `~/.modula/` works via `os.UserHomeDir()` + `filepath.Join`
- File permissions (`0755`, `0600`) are no-ops on Windows but don't break anything
- Bubbletea and Wish both have Windows support
- Relative test paths work (Go's `os` package accepts both separators)

## Phase 1: Self-Signaling Fix

**Files:** `cmd/connect.go`, `cmd/tui.go`
**Change:** ~20 lines

`process.Signal(syscall.SIGTERM)` is documented as "not implemented" on Windows. Returns error at runtime.

Replace the self-signal pattern in both files with `os.Exit(1)`. These commands have no background servers to drain (unlike `serve`), so direct exit is the correct cross-platform equivalent.

- `cmd/connect.go:124-132` -- replace with `os.Exit(1)`
- `cmd/tui.go:86-95` -- replace with `os.Exit(1)`

Both files can then drop the `"syscall"` import.

**Verify:** `just check`, manual test of `tui` and `connect` exit behavior.

## Phase 2: DumpSql Rewrite

**Files:** `internal/db/db.go`
**Change:** ~120 lines

Three `DumpSql` methods shell out to `/bin/bash` with embedded scripts. These are broken on ALL platforms (embed ReadFile paths reference non-existent paths). Rewrite as direct `exec.Command` calls to database CLI tools:

- **SQLite:** `exec.Command("sqlite3", dbName, ".dump")` with stdout redirected to file
- **MySQL:** `exec.Command("mysqldump", "-u", user, database)` with `MYSQL_PWD` env var
- **PostgreSQL:** `exec.Command("pg_dump", "-U", user, "-d", database, "-f", outFile)` with `PGPASSWORD` env var

No temp files, no chmod, no bash. Aligns with existing direction (CLAUDE.md notes DumpSql is superseded by PLUGIN_DB_EXPORTER plan).

**Verify:** `just check`. Test each variant against its respective database if available.

## Phase 3: Editor Fallback

**Files:** `internal/tui/uiconfig_form_dialog.go`
**Change:** ~5 lines

The `editorCommand()` function falls back to `"vi"` which doesn't exist on Windows.

Add a `runtime.GOOS` check:

```go
if runtime.GOOS == "windows" {
    return "notepad"
}
return "vi"
```

No build tags needed.

**Verify:** `just check`, `just test`.

## Phase 4: WebP Encoder CGO Directive

**Files:** `vendor/github.com/kolesa-team/go-webp/encoder/encoder.go`, `vendor/github.com/kolesa-team/go-webp/encoder/options.go`
**Change:** 2 lines (one per file)

The vendored go-webp encoder only has `#cgo linux` and `#cgo darwin` LDFLAGS. Add:

```c
#cgo windows LDFLAGS: -lwebp
```

MSYS2 is already needed for MinGW (SQLite CGO), and libwebp is a standard MSYS2 package (`pacman -S mingw-w64-x86_64-libwebp`).

**Verify:** Cross-compile with MinGW if available, otherwise CI validates.

## Phase 5: VerifyBinary Executable Check

**Files:** `internal/update/update.go`, `internal/update/checker.go`
**Change:** ~5 lines

- Skip `mode&0111 == 0` check on Windows (files always report executable bits)
- Handle `.exe` suffix in `GetDownloadURL` asset name matching

```go
if runtime.GOOS != "windows" && mode&0111 == 0 {
    return fmt.Errorf("binary is not executable")
}
```

**Verify:** Unit tests in `internal/update/`.

## Phase 6: CI -- GitHub Actions Windows Runner

**Files:** `.github/workflows/go.yml`

Add Windows matrix entry to the build job:

```yaml
- os: windows-latest
  goos: windows
  goarch: amd64
  cc: gcc
```

Windows-specific steps:
- Install MSYS2 with MinGW-w64 toolchain (`msys2/setup-msys2@v2`)
- Install libwebp (`pacman -S mingw-w64-x86_64-libwebp`)
- Set `CC=gcc`, `CGO_ENABLED=1`
- Build output: `modula-windows-amd64.exe`

Optional: Windows test job (reduced scope, running in MSYS2 shell).

**Verify:** Push to branch, check GitHub Actions.

## Phase 7: Justfile Windows Recipes

**Files:** `justfile`

Use `just` built-in `os_family()` for cross-platform conditionals:

```just
check:
    go build -mod vendor -o {{ if os_family() == "windows" { "NUL" } else { "/dev/null" } }} ./cmd
```

The `test` recipe's `touch`/`rm` commands need PowerShell equivalents or wrapping in `os_family()` conditionals.

**Verify:** `just check` and `just test` on both platforms.

## Not Addressed (Acceptable)

| Item | Why it's fine |
|------|--------------|
| File permissions (0755/0600) | No-op on Windows, doesn't break |
| SSH host key `.ssh/` dir | Works with dot-prefix on Windows |
| `~/.modula/` config path | `os.UserHomeDir()` returns `%USERPROFILE%` |
| `syscall.EXDEV` cross-drive rename | Edge case for self-update; error propagates gracefully |
| Private key permission hardening | Would need Windows ACLs; security nice-to-have, not functional |
