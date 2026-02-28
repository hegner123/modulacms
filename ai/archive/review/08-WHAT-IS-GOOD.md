# What Is Good

A consolidated list of things this project gets right.

## Architectural Decisions

**Single binary distribution.** The entire CMS - server, admin panel, TUI, plugin runtime - compiles into one binary. This is the right call for a CMS that targets small-to-medium deployments. No runtime dependencies, no process managers, no separate frontend builds.

**Three concurrent servers from one process.** HTTP, HTTPS (with Let's Encrypt autocert), and SSH (with Bubbletea TUI) all start from `serve` and shut down together. Graceful shutdown with signal handling (first SIGINT triggers drain, second forces exit).

**Stdlib-first philosophy.** Go 1.22+ stdlib ServeMux for routing, database/sql for database access, net/http for the server, crypto/bcrypt for passwords. External dependencies are used only when stdlib doesn't cover the need (Bubbletea for TUI, sqlc for code generation, Cobra for CLI).

**Content tree with sibling pointers.** `parent_id`, `first_child_id`, `next_sibling_id`, `prev_sibling_id` enables O(1) insertion, deletion, and reordering. No sort_order renumbering. This is a better data structure than most CMS systems use.

## Type Safety

**30 branded ID types.** DatatypeID, ContentID, UserID, FieldID, MediaID, RoleID, etc. Each is a distinct Go type that implements database and JSON interfaces. The compiler prevents passing a UserID where a ContentID is expected. This is replicated consistently across all three SDKs.

**ULID-based IDs.** Sortable by creation time, globally unique without coordination, 26-character string representation. Thread-safe generation with mutex-protected monotonic entropy.

**Enum validation.** ContentStatus, FieldType, RouteType all validate on assignment and database scan. Invalid values are caught at the boundary, not deep in business logic.

## Security

**RBAC permission system.** Granular `resource:operation` permissions (47 permissions across 3 bootstrap roles). In-memory cache with build-then-swap for zero-blocking reads. 60-second periodic refresh. Fail-closed: missing context returns 403.

**Plugin sandboxing.** Restricted Lua stdlib, frozen API modules, operation budgets, namespace isolation, route approval gates, circuit breakers. This is serious about preventing plugins from damaging the host.

**Rate limiting on auth endpoints.** 10 requests per minute per IP. Not configurable to dangerous values.

**bcrypt with cost 12.** Standard, well-tested, appropriate cost factor.

**Non-root Docker container.** uid 1000 with `/bin/false` shell. Stripped binary. Minimal runtime image.

## Testing

**99,000 lines of Go test code.** Nearly 1:1 with production code. The database layer has 79 test files. Table-driven tests throughout. SQLite test databases created and cleaned per run.

**SDK tests use real HTTP servers.** The TypeScript admin SDK and Go SDK both spin up actual HTTP test servers instead of only mocking. This catches real integration issues.

**Swift SDK has MockURLProtocol.** Properly handles both httpBody and httpBodyStream, which is a common source of test failures in URLSession testing.

## Developer Experience

**80+ Justfile recipes.** One command for any common operation: `just dev`, `just test`, `just sdk-build`, `just docker-up`, `just deploy`, `just lint`.

**sqlc code generation.** SQL is the source of truth. Write queries in SQL, get type-safe Go code. The 75KB sqlc.yml configuration with 250+ type overrides demonstrates deep commitment to this approach.

**Hot-reloading.** Config changes apply without restart (build-then-swap). Plugin changes detected via file polling with debounce. Admin panel uses Vite HMR in development.

**70+ documentation files.** Architecture guides, domain docs, workflow procedures, sqlc reference, refactoring plans. This project is self-documenting at a level few open-source projects achieve.

## SDK Design

**Zero external dependencies in all three SDKs.** TypeScript uses fetch, Go uses net/http, Swift uses URLSession. No axios, no alamofire, no third-party HTTP clients.

**Consistent API surface across three languages.** Method names, resource types, error helpers, and pagination all map 1:1 (adjusted for language conventions).

**Generic CRUD pattern.** One generic implementation serves 14+ entity types in Go and Swift. No copy-paste CRUD boilerplate.

## Admin Panel

**Modern, well-chosen stack.** React 19, TypeScript strict, TanStack Router/Query, shadcn/ui, Tailwind. No unnecessary abstractions, no heavy state management library.

**Block editor with local state management.** `useBlockEditorState()` tracks per-block dirty state, enables individual block saves, and handles drag-drop reordering. This is the core UX feature and it's well-implemented.

**SDK proxy pattern for auth switching.** Clean runtime switching between session and API key auth without component rewrites.

## Plugin System

**Complete lifecycle management.** Discovery, validation, loading, running, reload (blue-green), shutdown. Hot reload with checksumming, debounce, and cooldown.

**Approval workflow.** Routes and hooks must be approved by admin before activation. Version changes reset approvals. This prevents plugins from silently gaining access.

**Metrics instrumentation.** Request counts, durations, hook execution stats, error rates, circuit breaker state. Ready for production monitoring.
