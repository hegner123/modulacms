# Prioritized Recommendations

Ordered by impact. Things that will make the biggest difference for the least effort first.

## High Impact, Moderate Effort

### 1. Add Missing Database Indexes
Add indexes on `content_fields(content_data_id)` and `content_fields(field_id)`. This is a one-line SQL change per database backend that prevents performance degradation as content grows. Do this today.

### 2. Add Admin Panel Tests
Install Vitest + React Testing Library. Write tests for:
- `useBlockEditorState` (the most complex state logic)
- Auth flow (login, logout, session/API key switching)
- SDK proxy pattern
- Media upload hook

Start with the block editor. Its local edit tracking, dirty state, and save logic are the highest-risk untested code.

### 3. Eliminate TUI Global Context Variables
Move all 15+ global dialog context variables into the Model struct. This is mechanical work that immediately improves testability and eliminates potential race conditions.

### 4. Add Query Filtering to SDKs
Add filter/sort parameters to list methods across all three SDKs. Even basic support (`List(ctx, ListParams{Status: "published", Sort: "date_created", Order: "desc"})`) would dramatically improve SDK usefulness for frontend developers.

## High Impact, High Effort

### 5. Split DbDriver Interface
Decompose the 315-method interface into focused repository interfaces:
```go
type ContentRepository interface { /* 20 methods */ }
type UserRepository interface { /* 15 methods */ }
type MediaRepository interface { /* 12 methods */ }
// etc.

type DbDriver interface {
    ContentRepository
    UserRepository
    MediaRepository
    // ...
}
```
This enables targeted mocking, reduces per-test setup, and makes the codebase navigable. Large effort (touching many files) but high architectural payoff.

### 6. Generate Wrapper Layer Code
Build a code generator that produces mapper functions from sqlc output. This eliminates ~5,000 lines of hand-written, triplicated code and ensures all three backends stay in sync automatically.

## Medium Impact, Low Effort

### 7. Standardize TUI Error Handling
Define one error handling pattern and apply it everywhere. Suggested: all errors produce a user-visible status message + structured log entry. No silent failures.

### 8. Add CI Container Image Build
Add a GitHub Actions job that builds and pushes Docker images to ghcr.io on tag push. The Dockerfile already works; this just automates it.

### 9. Consolidate Docker Compose Files
Use Docker Compose profiles to merge the 6 compose files into 2 (dev + prod). Update Justfile recipes to pass `--profile` instead of `-f` flags.

### 10. Add Release Checksums
Generate sha256sum files for all binary artifacts in the release job. Low effort, meaningful security improvement.

## Medium Impact, Medium Effort

### 11. Split TUI Large Files
Break `update_dialog.go` (2,893 lines), `commands.go` (2,378 lines), `form_dialog.go` (2,302 lines), and `update_controls.go` (1,800 lines) into files under 500 lines each. Group by entity or page type.

### 12. Extract serve.go Concerns
Move HTTP server, HTTPS server, SSH server, and plugin initialization into separate functions or files. The main `RunE` function should orchestrate, not implement.

### 13. Make author_id NOT NULL
Add a migration to make `content_data.author_id` and `datatypes.author_id` NOT NULL. Backfill any existing NULL values with a system user ID. Remove defensive NULL checks from application code.

## Low Priority

### 14. Remove or Gate Dead TUI Pages
Remove DEVELOPMENT, DYNAMICPAGE stubs or gate them behind a developer flag.

### 15. Prune AI Documentation
Consolidate overlapping docs in `ai/`. Remove stale refactoring plans. Add an index file.

### 16. Add SAST to CI
Add golangci-lint and trivy scanning to the CI pipeline for ongoing code quality and security.
