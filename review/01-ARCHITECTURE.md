# Architecture Review

## The Core Idea

A single Go binary runs three concurrent servers: HTTP (public API), HTTPS (autocert with Let's Encrypt), and SSH (Bubbletea TUI). Content flows through a middleware chain into stdlib ServeMux handlers, through a DbDriver interface, into one of three database backends. This is a clean, defensible architecture.

## What Solves a Real Problem

### Tri-Database Support
Supporting SQLite, MySQL, and PostgreSQL from one binary is genuinely useful. SQLite for development and small deployments, PostgreSQL for production at scale. The implementation via sqlc code generation into three separate packages, wrapped by a unified interface, actually works. The sqlc.yml configuration (75KB, 250+ type overrides) is thorough and correct.

### Single Binary Distribution
Everything compiles into one binary: the Go server, the admin panel (embedded via `//go:embed`), the TUI, the plugin runtime. This dramatically simplifies deployment. Run `./modulacms serve` and you have a CMS. No Node.js runtime, no separate frontend build, no process manager.

### Typed ID System (ULID-based)
30 distinct ID types (ContentID, UserID, DatatypeID, etc.) all backed by ULIDs. Each implements `driver.Valuer`, `sql.Scanner`, `json.Marshaler`. You cannot pass a UserID where a ContentID is expected - the compiler catches it. This prevents an entire class of bugs that plague systems using raw strings for IDs.

### Audit Trail
Every mutation atomically records a change_event with operation type, old/new JSON values, user ID, request ID, and IP address. Uses hybrid logical clocks for distributed ordering. This is real compliance-grade auditing, not a logging afterthought.

### Permission System
Role-based access control with `resource:operation` granular permissions. In-memory cache with build-then-swap pattern (readers never block during DB refresh). Periodic 60-second refresh. Admin bypass via boolean flag, not wildcard. Fail-closed: missing permissions in context returns 403.

## What Is Good But Has Growing Pains

### Middleware Chain
Clean functional composition: `corsMiddleware(authLimiter.Middleware(handler))`. But every route in mux.go wraps 3-6 middleware layers inline. 370 lines of route registration, each nearly identical in structure. A route registration helper could reduce this to a table of `{method, path, permission, handler}` entries.

### Configuration System
Hot-reloading with build-then-swap, provider abstraction, environment variable interpolation. Good. But the Config struct has 100+ fields spanning database, OAuth, CORS, email, plugins, observability, SSL. It's approaching the point where feature flags or sub-configs would help maintainability.

### Serve Command
The serve command orchestrates HTTP, HTTPS, SSH, plugin initialization, email service, permission cache, config watching. At ~530 lines, it does too much. Each server type could be extracted into its own start function.

## What Is Concerning

### DbDriver Interface Size
315 methods in a single interface. This is the largest structural problem in the codebase. It means:
- You cannot mock it for testing without implementing all 315 methods
- Adding a new entity requires adding methods to the interface and all three wrapper structs
- The interface file itself is 3,199 lines

This should be decomposed into focused interfaces: `ContentRepository`, `UserRepository`, `MediaRepository`, etc. The wrappers already group methods by entity - the interface just hasn't caught up.

### Three Copies of Everything
Each database entity has three mapper functions (SQLite, MySQL, PostgreSQL), three CRUD implementations, three sets of NULL type conversions. The wrapper layer is hand-written, not generated. Adding a field to the users table means updating 6+ mapper functions manually. A code generator for the wrapper layer would eliminate ~5,000 lines of mechanical duplication.

### Global State in Server Initialization
`cfgPath` and `verbose` are package-level variables set by CLI flags. The DefaultLogger is a singleton. These make the server startup logic hard to test in isolation. Dependency injection via constructor functions would improve testability.

## Architecture Diagram

```
                    +-----------+
                    |   Client  |
                    +-----+-----+
                          |
          +---------------+---------------+
          |               |               |
     HTTP:8080       HTTPS:8443      SSH:2222
          |               |               |
          v               v               v
    +-----+-----+  +-----+-----+  +------+------+
    | Middleware |  | Middleware |  | Wish/SSH    |
    | Chain      |  | + autocert|  | + Bubbletea |
    +-----+-----+  +-----+-----+  +------+------+
          |               |               |
          v               v               v
    +-----+---------------+---------------+------+
    |              stdlib ServeMux                 |
    |     (Go 1.22+ pattern routing)              |
    +-----+-------+-------+-------+---------+----+
          |       |       |       |         |
          v       v       v       v         v
       Auth    Content  Media  Schema    Plugins
       Handler Handler Handler Handler  Bridge
          |       |       |       |         |
          v       v       v       v         v
    +-----+-------+-------+-------+---------+----+
    |           DbDriver Interface                 |
    |            (315 methods)                     |
    +-----+---------------+---------------+-------+
          |               |               |
          v               v               v
     +----+----+    +-----+-----+   +-----+-----+
     | SQLite  |    |   MySQL   |   | PostgreSQL|
     | (sqlc)  |    |   (sqlc)  |   |   (sqlc)  |
     +---------+    +-----------+   +-----------+
```

## Recommendations

1. **Split DbDriver** into 8-10 resource-specific interfaces composed into a main driver
2. **Generate wrapper code** instead of hand-writing mappers for three backends
3. **Extract serve.go** concerns into server-specific initialization functions
4. **Add route registration table** to reduce mux.go boilerplate
5. **Sub-config the Config struct** by feature domain
