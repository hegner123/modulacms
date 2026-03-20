# Adding Features

Add a feature to ModulaCMS by following the standard flow from schema design through deployment.

## Determine the Scope

Before writing code, determine the scope:

- **Feature stores new data in an existing table** -- Add a column to the existing schema, update queries, regenerate sqlc code.
- **Feature introduces a new domain concept** -- Create a new table with the full workflow described in [Adding Tables](adding-tables.md).
- **Feature uses existing data in a new way** -- Skip directly to business logic. No schema or sqlc changes needed.

## Follow the Development Flow

```
Schema Design (if needed)
    |
SQL Files (schema + queries)
    |
sqlc Code Generation
    |
DbDriver Interface Update
    |
Driver Implementations (SQLite, MySQL, PostgreSQL)
    |
Business Logic
    |
TUI Interface (if needed)
    |
HTTP/API Endpoints (if needed)
    |
Admin Panel Pages (if needed)
    |
Testing
    |
Deployment
```

Not every feature touches every layer. A read-only export feature skips the schema and sqlc steps. A background job skips the TUI and API steps. Follow the flow and skip layers that do not apply.

## Step 1: Write Schema and Queries

If the feature requires database changes, follow the [Adding Tables](adding-tables.md) guide for new tables, or add ALTER TABLE statements to a new migration directory for column additions.

Key points:
- Create migrations for all three databases (SQLite, MySQL, PostgreSQL)
- Write sqlc-annotated queries for all three databases
- Update the combined schema files
- Run `just sqlc` to generate Go code

## Step 2: Update the DbDriver Interface

If new queries were added in Step 1, add the new methods to the `DbDriver` interface in `internal/db/db.go`. Then implement those methods on all three driver structs: `Database` (SQLite), `MysqlDatabase`, and `PsqlDatabase`.

Each implementation maps between sqlc-generated types and application-level types, handling NULL conversions and type width differences between database engines.

## Step 3: Implement Business Logic

Place domain logic in the appropriate location:

- **Simple CRUD** -- Handled by the driver implementations from Step 2.
- **Domain rules and validation** -- `internal/model/`
- **HTTP request handling** -- `internal/router/`
- **TUI interaction** -- `internal/tui/`

Use structured logging at decision points and error paths.

## Step 4: Build the TUI Interface

If the feature needs user interaction in the SSH TUI:

1. **Define message types** for the feature's events.
2. **Add keyboard commands** in the Update function.
3. **Create command functions** that perform async operations and return messages.
4. **Update the View** to render the new feature's state.

For entirely new screens, create a new model file in `internal/tui/`. For additions to existing screens, modify the relevant model's Update and View functions.

## Step 5: Add HTTP/API Endpoints

If the feature needs REST API access:

1. **Create handler functions** in `internal/router/`.
2. **Register routes** in `internal/router/mux.go` with appropriate permission middleware.
3. **Add permission labels** if the feature needs new permissions -- add them to the bootstrap data.

> **Good to know**: All admin endpoints must be wrapped with `RequireResourcePermission` or `RequirePermission` middleware.

## Step 6: Create Admin Panel Pages

If the feature needs a web admin interface:

1. **Create templ templates** in `internal/admin/pages/` and `internal/admin/partials/`.
2. **Create handlers** in `internal/admin/handlers/`.
3. **Register routes** in the `registerAdminRoutes()` function.
4. **Regenerate templates** with `just admin generate`.

## Step 7: Write Tests

Every feature needs tests. At minimum:

- **Unit tests** for business logic and validation functions
- **Database tests** for new CRUD operations (using SQLite test databases)
- **Manual testing** of TUI commands via SSH
- **Manual testing** of API endpoints with curl

```bash
# Run all tests
just test

# Run specific package
go test -v ./internal/db -run TestCommentCRUD

# Run with coverage
just coverage

# Run with race detector
go test -race ./...
```

### Testing Checklist

- Unit tests for all business logic
- Database CRUD operations tested
- Error cases handled and tested (invalid input, missing records, constraint violations)
- TUI commands tested manually
- API endpoints tested with curl
- `just test` passes
- `just lint` passes

## Step 8: Deploy the Feature

Build and verify locally before deploying:

```bash
# Build for local testing
just dev
modula

# Run the full test suite
just test

# Build for production
just build
```

Test the feature locally by connecting to the TUI via SSH and hitting the API endpoints. Then deploy following the standard deployment process (push to dev branch for CI, or manual deploy with `just build`).

## Common Patterns

### Adding a Column to an Existing Table

1. Create migration directory: `sql/schema/N_feature_name/`
2. Write ALTER TABLE statements for all three databases
3. Add or update queries
4. Run `just sqlc`
5. Update DbDriver interface if new queries were added
6. Implement on all three drivers
7. Update business logic and interfaces
8. Test and deploy

### Creating a New Table

Follow [Adding Tables](adding-tables.md) for the full schema-to-code workflow, then continue with business logic, interfaces, and testing from this guide.

### Read-Only Feature (No Database Changes)

1. Implement business logic
2. Add TUI interface or HTTP endpoints
3. Test and deploy

### Background Job

1. Implement the job logic
2. Add configuration fields to `modula.config.json` if needed
3. Register the job in the server startup flow
4. Add logging for monitoring
5. Test and deploy

## Avoid Common Pitfalls

**Forgetting to implement on all three database drivers.** The feature works in SQLite during development but fails with MySQL or PostgreSQL in production. Implement DbDriver methods on all three structs.

**Not updating combined schema files.** Fresh installations use `all_schema*.sql`. Missing tables cause failures on new deployments.

**SQL dialect differences between databases.** MySQL uses `?` placeholders, PostgreSQL uses `$1, $2, $3`. MySQL does not support `RETURNING`. Test queries against all backends.

**Breaking the TUI message loop.** Returning `nil` for `tea.Cmd` when a command is expected causes the TUI to stop responding. Return the appropriate command from Update.

**Missing permission guards on API endpoints.** Every admin endpoint must be wrapped with permission middleware. Unguarded endpoints bypass RBAC.
