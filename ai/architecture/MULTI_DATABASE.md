# MULTI_DATABASE.md

Multi-Database Support Architecture for ModulaCMS

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/MULTI_DATABASE.md`
**Created:** 2026-01-12
**Purpose:** Explains why ModulaCMS supports three database engines, how driver selection works, database-specific considerations, and migration strategies.

---

## Overview

ModulaCMS supports three database engines through a unified interface:

1. **SQLite** - Development and small deployments
2. **MySQL** - Production deployments
3. **PostgreSQL** - Enterprise deployments

This multi-database architecture allows developers to use SQLite locally for fast iteration, then deploy to MySQL or PostgreSQL in production without code changes. The abstraction is achieved through the `DbDriver` interface combined with sqlc-generated type-safe code.

---

## Why Three Database Engines?

### SQLite: Development and Small Deployments

**When to use:**
- Local development
- Testing environments
- Small sites (< 10,000 content nodes)
- Single-server deployments
- Prototyping and demos

**Advantages:**
- Zero configuration (just a file path)
- No separate database server needed
- Fast for development (no network latency)
- Perfect for CI/CD testing (in-memory mode available)
- Simple backups (copy the .db file)

**Limitations:**
- No concurrent writes (database locked during write operations)
- Single server only (no horizontal scaling)
- Limited to filesystem I/O performance
- Not suitable for high-traffic production sites

**Configuration example:**
```json
{
  "db_driver": "sqlite",
  "db_url": "./modula.db"
}
```

### MySQL: Production Deployments

**When to use:**
- Production websites
- Multi-server deployments
- High-traffic sites (10,000+ req/sec)
- Environments with established MySQL infrastructure
- Cost-optimized hosting (MySQL available everywhere)

**Advantages:**
- Concurrent writes supported
- Horizontal scaling with replication
- Mature ecosystem and tooling
- Wide hosting availability
- Excellent performance for CMS workloads
- Strong community support

**Considerations:**
- Requires separate MySQL server
- Connection pooling configuration important at scale
- Requires credentials management

**Configuration example:**
```json
{
  "db_driver": "mysql",
  "db_url": "localhost:3306",
  "db_name": "modulacms",
  "db_username": "cms_user",
  "db_password": "secure_password"
}
```

### PostgreSQL: Enterprise Deployments

**When to use:**
- Enterprise environments
- Complex data integrity requirements
- Advanced query needs
- Environments with established PostgreSQL infrastructure
- Compliance-heavy industries

**Advantages:**
- Superior data integrity enforcement
- Advanced indexing strategies
- Complex query optimization
- ACID compliance strictness
- JSON support for flexible schemas
- Enterprise-grade features

**Considerations:**
- Slightly more complex administration
- Less ubiquitous than MySQL in hosting environments
- Requires separate PostgreSQL server

**Configuration example:**
```json
{
  "db_driver": "postgres",
  "db_url": "localhost:5432",
  "db_name": "modulacms",
  "db_username": "cms_user",
  "db_password": "secure_password"
}
```

---

## Driver Selection Architecture

### The DbDriver Interface

All database operations go through the `DbDriver` interface defined in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go`:

```go
type DbDriver interface {
    // Connection management
    GetConnection() (*sql.DB, context.Context, error)
    Ping() error
    InitDB(v *bool) error
    CreateAllTables() error

    // Count operations (14 methods)
    CountContentData() (*int64, error)
    CountRoutes() (*int64, error)
    // ... more count methods

    // Create operations (21 methods)
    CreateContentData(CreateContentDataParams) ContentData
    CreateRoute(CreateRouteParams) Routes
    // ... more create methods

    // Get operations (26 methods)
    GetContentData(int64) (*ContentData, error)
    GetRoute(int64) (*Routes, error)
    // ... more get methods

    // List operations (24 methods)
    ListContentData() (*[]ContentData, error)
    ListRoutes() (*[]Routes, error)
    // ... more list methods

    // Update operations (21 methods)
    UpdateContentData(UpdateContentDataParams) (*string, error)
    UpdateRoute(UpdateRouteParams) (*string, error)
    // ... more update methods

    // Delete operations (21 methods)
    DeleteContentData(int64) error
    DeleteRoute(int64) error
    // ... more delete methods
}
```

**Total interface methods:** 150+ methods covering all database operations.

### Three Implementations

Each database has its own struct implementing `DbDriver`:

**SQLite:**
```go
type Database struct {
    Src            string
    Status         DbStatus
    Connection     *sql.DB
    LastConnection string
    Err            error
    Config         config.Config
    Context        context.Context
}
```

**MySQL:**
```go
type MysqlDatabase struct {
    Src            string
    Status         DbStatus
    Connection     *sql.DB
    LastConnection string
    Err            error
    Config         config.Config
    Context        context.Context
}
```

**PostgreSQL:**
```go
type PsqlDatabase struct {
    Src            string
    Status         DbStatus
    Connection     *sql.DB
    LastConnection string
    Err            error
    Config         config.Config
    Context        context.Context
}
```

### Runtime Driver Selection

The `ConfigDB` function in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go` switches between drivers based on configuration:

```go
func ConfigDB(env config.Config) DbDriver {
    verbose := false
    switch env.Db_Driver {
    case config.Sqlite:
        d := Database{Src: env.Db_URL, Config: env}
        dbc := d.GetDb(&verbose)
        return dbc
    case config.Mysql:
        d := MysqlDatabase{Src: env.Db_Name, Config: env}
        dbc := d.GetDb(&verbose)
        return dbc
    case config.Psql:
        d := PsqlDatabase{Src: env.Db_Name, Config: env}
        dbc := d.GetDb(&verbose)
        return dbc
    }
    return nil
}
```

**Key points:**
- Single entry point for all database initialization
- Returns `DbDriver` interface (not concrete type)
- Driver selection happens once at startup
- All application code uses `DbDriver` interface methods

---

## Database-Specific Connection Logic

### SQLite Connection

Location: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`

```go
func (d Database) GetDb(verbose *bool) DbDriver {
    if *verbose {
        utility.DefaultLogger.Info("Connecting to SQLite database...")
    }
    ctx := context.Background()

    // Use default path if not specified
    if d.Src == "" {
        d.Src = "./modula.db"
        utility.DefaultLogger.Info("Using default database path", "path", d.Src)
    }

    // Open database connection
    db, err := sql.Open("sqlite3", d.Src)
    if err != nil {
        errWithContext := fmt.Errorf("failed to open SQLite database: %w", err)
        utility.DefaultLogger.Error("Database connection error", errWithContext, "path", d.Src)
        d.Err = errWithContext
        return d
    }

    // Enable foreign keys (CRITICAL for SQLite)
    _, err = db.Exec("PRAGMA foreign_keys=1;")
    if err != nil {
        errWithContext := fmt.Errorf("failed to enable foreign keys: %w", err)
        utility.DefaultLogger.Error("Database configuration error", errWithContext)
        d.Err = errWithContext
        return d
    }

    d.Connection = db
    d.Context = ctx
    d.Err = nil
    return d
}
```

**SQLite-specific considerations:**
- File path connection string (not network DSN)
- **MUST enable foreign keys with PRAGMA** (disabled by default in SQLite)
- No authentication needed
- Connection is essentially a file handle

### MySQL Connection

Location: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`

```go
func (d MysqlDatabase) GetDb(verbose *bool) DbDriver {
    if *verbose {
        utility.DefaultLogger.Info("Connecting to MySQL database...")
    }
    ctx := context.Background()

    // Create connection string (DSN format)
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s",
        d.Config.Db_User,
        d.Config.Db_Password,
        d.Config.Db_URL,
        d.Config.Db_Name)

    // Hide password in logs
    sanitizedDsn := fmt.Sprintf("%s:****@tcp(%s)/%s",
        d.Config.Db_User,
        d.Config.Db_URL,
        d.Config.Db_Name)
    utility.DefaultLogger.Info("Preparing MySQL connection", "dsn", sanitizedDsn)

    // Open database connection
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        errWithContext := fmt.Errorf("failed to open MySQL database: %w", err)
        utility.DefaultLogger.Error("Database connection error", errWithContext, "host", d.Config.Db_URL)
        d.Err = errWithContext
        return d
    }

    // Test the connection
    err = db.Ping()
    if err != nil {
        errWithContext := fmt.Errorf("failed to connect to MySQL database: %w", err)
        utility.DefaultLogger.Error("Database ping error", errWithContext, "host", d.Config.Db_URL)
        d.Err = errWithContext
        return d
    }

    utility.DefaultLogger.Info("MySQL database connected successfully", "database", d.Config.Db_Name)
    d.Connection = db
    d.Context = ctx
    d.Err = nil
    return d
}
```

**MySQL-specific considerations:**
- DSN format: `user:password@tcp(host:port)/database`
- Requires explicit connection test with `Ping()`
- Password sanitization in logs for security
- Network connection (TCP)
- Authentication required

### PostgreSQL Connection

Location: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`

```go
func (d PsqlDatabase) GetDb(verbose *bool) DbDriver {
    if *verbose {
        utility.DefaultLogger.Info("Connecting to PostgreSQL database...")
    }
    ctx := context.Background()

    // Create connection string (PostgreSQL URI format)
    connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
        d.Config.Db_User,
        d.Config.Db_Password,
        d.Config.Db_URL,
        d.Config.Db_Name)

    // Hide password in logs
    sanitizedConnStr := fmt.Sprintf("postgres://%s:****@%s/%s?sslmode=disable",
        d.Config.Db_User,
        d.Config.Db_URL,
        d.Config.Db_Name)
    utility.DefaultLogger.Info("Preparing PostgreSQL connection", "connection", sanitizedConnStr)

    // Open database connection
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        errWithContext := fmt.Errorf("failed to open PostgreSQL database: %w", err)
        utility.DefaultLogger.Error("Database connection error", errWithContext, "host", d.Config.Db_URL)
        d.Err = errWithContext
        return d
    }

    // Test the connection
    err = db.Ping()
    if err != nil {
        errWithContext := fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
        utility.DefaultLogger.Error("Database ping error", errWithContext, "host", d.Config.Db_URL)
        d.Err = errWithContext
        return d
    }

    utility.DefaultLogger.Info("PostgreSQL database connected successfully", "database", d.Config.Db_Name)
    d.Connection = db
    d.Context = ctx
    d.Err = nil
    return d
}
```

**PostgreSQL-specific considerations:**
- URI format: `postgres://user:password@host:port/database?options`
- Currently uses `sslmode=disable` (should be configurable for production)
- Requires explicit connection test with `Ping()`
- Password sanitization in logs
- Network connection
- Authentication required

---

## Connection Management

### Connection Pooling

ModulaCMS uses Go's standard `database/sql` package, which provides built-in connection pooling with sensible defaults:

**Default connection pool settings:**
- `MaxOpenConns`: Unlimited (0)
- `MaxIdleConns`: 2
- `ConnMaxLifetime`: Unlimited
- `ConnMaxIdleTime`: Unlimited

**Current implementation:**
No custom connection pool configuration is set. The application relies on Go's defaults, which work well for most CMS workloads.

**Future optimization opportunities:**
For high-traffic production deployments, consider configuring:
```go
db.SetMaxOpenConns(25)           // Limit concurrent connections
db.SetMaxIdleConns(10)           // Keep connections ready
db.SetConnMaxLifetime(5 * time.Minute)  // Recycle old connections
```

### Context Usage

All database operations receive a `context.Context` for cancellation and timeout control. This allows:
- Request-scoped database operations
- Graceful shutdown during long queries
- Timeout enforcement for slow operations

Example from application code:
```go
databaseConnection, ctx, _ := db.ConfigDB(*configuration).GetConnection()
defer utility.HandleConnectionCloseDeferErr(databaseConnection)
```

---

## Database-Specific Schema Differences

Each database engine requires slightly different SQL syntax. ModulaCMS maintains three versions of each schema migration:

### Schema File Organization

Location: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/`

```
sql/schema/
├── 1_permissions/
│   ├── schema.sql          # SQLite
│   ├── schema_mysql.sql    # MySQL
│   └── schema_psql.sql     # PostgreSQL
├── 2_roles/
│   ├── schema.sql
│   ├── schema_mysql.sql
│   └── schema_psql.sql
├── 16_content_data/
│   ├── schema.sql
│   ├── schema_mysql.sql
│   └── schema_psql.sql
└── ... (22 total migration directories)
```

### Key Schema Differences

#### Auto-Increment Primary Keys

**SQLite:**
```sql
CREATE TABLE permissions (
    permission_id INTEGER PRIMARY KEY,
    -- INTEGER PRIMARY KEY is auto-incrementing in SQLite
)
```

**MySQL:**
```sql
CREATE TABLE permissions (
    permission_id INT AUTO_INCREMENT PRIMARY KEY,
    -- Explicit AUTO_INCREMENT keyword
)
```

**PostgreSQL:**
```sql
CREATE TABLE permissions (
    permission_id SERIAL PRIMARY KEY,
    -- SERIAL is PostgreSQL's auto-increment type
)
```

#### Text vs VARCHAR

**SQLite:**
```sql
label TEXT NOT NULL
```

**MySQL:**
```sql
label VARCHAR(255) NOT NULL
-- MySQL requires length specification
```

**PostgreSQL:**
```sql
label TEXT NOT NULL
-- PostgreSQL TEXT is preferred over VARCHAR
```

#### Foreign Key Constraints

**SQLite (inline references):**
```sql
CREATE TABLE content_data (
    content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    route_id INTEGER NOT NULL REFERENCES routes ON DELETE CASCADE,

    -- Also explicit foreign key declarations
    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE RESTRICT
);
```

**MySQL (named constraints):**
```sql
CREATE TABLE content_data (
    content_data_id INT AUTO_INCREMENT PRIMARY KEY,
    parent_id INT NULL,
    route_id INT NULL,

    -- Named constraints with ON UPDATE CASCADE support
    CONSTRAINT fk_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    CONSTRAINT fk_content_data_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);
```

**PostgreSQL (similar to MySQL):**
```sql
CREATE TABLE content_data (
    content_data_id SERIAL PRIMARY KEY,
    parent_id INTEGER NULL,
    route_id INTEGER NULL,

    CONSTRAINT fk_content_data_parent_id
        FOREIGN KEY (parent_id) REFERENCES content_data (content_data_id)
            ON DELETE SET NULL,
    CONSTRAINT fk_content_data_route_id
        FOREIGN KEY (route_id) REFERENCES routes (route_id)
            ON DELETE RESTRICT
);
```

**Key differences:**
- SQLite uses inline references AND explicit FOREIGN KEY declarations
- MySQL and PostgreSQL use named CONSTRAINT syntax
- MySQL supports `ON UPDATE CASCADE` (SQLite does not)
- PostgreSQL has stricter foreign key enforcement by default

#### Timestamp Handling

**SQLite:**
```sql
date_created TEXT DEFAULT CURRENT_TIMESTAMP,
date_modified TEXT DEFAULT CURRENT_TIMESTAMP
-- Stores timestamps as TEXT
```

**MySQL:**
```sql
date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP
-- Native TIMESTAMP type with automatic update on row modification
```

**PostgreSQL:**
```sql
date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- Native TIMESTAMP type
```

**Key differences:**
- SQLite stores timestamps as TEXT (ISO 8601 format)
- MySQL has `ON UPDATE CURRENT_TIMESTAMP` for automatic modification tracking
- PostgreSQL uses proper TIMESTAMP type but requires triggers for auto-update behavior

---

## sqlc Integration

### How sqlc Works with Multiple Databases

ModulaCMS uses [sqlc](https://sqlc.dev) to generate type-safe Go code from SQL queries. The configuration is in `/Users/home/Documents/Code/Go_dev/modulacms/sql/sqlc.yml`:

```yaml
version: "2"
sql:
  - engine: "sqlite"
    queries: "sql/"
    schema: "sql/schema/"
    gen:
      go:
        package: "mdb"
        out: "internal/db-sqlite"

  - engine: "mysql"
    queries: "sql/mysql/"
    schema: "sql/schema/"
    gen:
      go:
        package: "mdbm"
        out: "internal/db-mysql"

  - engine: "postgresql"
    queries: "sql/postgres/"
    schema: "sql/schema/"
    gen:
      go:
        package: "mdbp"
        out: "internal/db-psql"
```

### Generated Code Structure

**SQLite package:** `internal/db-sqlite/` (package `mdb`)
**MySQL package:** `internal/db-mysql/` (package `mdbm`)
**PostgreSQL package:** `internal/db-psql/` (package `mdbp`)

Each package contains:
- `db.go` - sqlc-generated database interface
- `models.go` - Type definitions for database rows
- `queries.sql.go` - Generated query functions

### Query Files Organization

**SQLite queries:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/` (shared with base)
**MySQL queries:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/mysql/`
**PostgreSQL queries:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/postgres/`

Example query in `sql/mysql/content.sql`:
```sql
-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = ?;

-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY date_created DESC;

-- name: CreateContentData :exec
INSERT INTO content_data (
    parent_id, route_id, datatype_id, author_id
) VALUES (?, ?, ?, ?);
```

sqlc generates:
```go
func (q *Queries) GetContentData(ctx context.Context, contentDataID int64) (ContentData, error)
func (q *Queries) ListContentData(ctx context.Context) ([]ContentData, error)
func (q *Queries) CreateContentData(ctx context.Context, arg CreateContentDataParams) error
```

### Regenerating Code After Schema Changes

After modifying schema files or queries:

```bash
cd sql
just sqlc
```

This runs sqlc for all three database engines and regenerates type-safe Go code.

**Important:** The `DbDriver` interface in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go` must be updated manually when adding new operations.

---

## Migration Strategy Between Databases

### Development to Production Migration

**Workflow:**
1. Develop locally with SQLite (`db_driver: "sqlite"`)
2. Test all features
3. Update config for production database:
   - Change `db_driver` to `"mysql"` or `"postgres"`
   - Add connection credentials
4. Deploy binary
5. Run schema initialization (binary handles this automatically)

### Database Export/Import

ModulaCMS includes database-specific dump functionality in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go`:

**SQLite dump:**
```go
func (d Database) DumpSql(c config.Config) error {
    script, err := sqlFiles.ReadFile("sql/dump_sql.sh")
    // Executes: sqlite3 database.db .dump > backup.sql
    cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "sqlite"+t+".sql")
    output, err := cmd.CombinedOutput()
    return nil
}
```

**MySQL dump:**
```go
func (d MysqlDatabase) DumpSql(c config.Config) error {
    script, err := sqlFiles.ReadFile("sql/dump_mysql.sh")
    // Executes: mysqldump -u user -p database > backup.sql
    cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_User, c.Db_Password, c.Db_Name, "mysql"+t+".sql")
    output, err := cmd.CombinedOutput()
    return nil
}
```

**PostgreSQL dump:**
```go
func (d PsqlDatabase) DumpSql(c config.Config) error {
    script, err := sqlFiles.ReadFile("sql/dump_psql.sh")
    // Executes: pg_dump database > backup.sql
    cmd := exec.Command("/bin/bash", tmpFile.Name(), c.Db_Name, "psql"+t+".sql")
    output, err := cmd.CombinedOutput()
    return nil
}
```

### Cross-Database Migration

**SQLite → MySQL/PostgreSQL:**
1. Dump SQLite database to SQL file
2. Manually edit SQL dump to fix syntax differences:
   - Change `INTEGER PRIMARY KEY` to `INT AUTO_INCREMENT` or `SERIAL`
   - Change `TEXT` to `VARCHAR(255)` for MySQL
   - Update foreign key syntax
3. Import into target database
4. Update config
5. Restart application

**Note:** A proper migration tool would automate these transformations. This is a planned enhancement.

---

## Schema Initialization

### Embedded Schema Files

All schema files are embedded in the binary using Go's `//go:embed` directive in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go`:

```go
//go:embed sql
var sqlFiles embed.FS
```

### InitDB Process

Location: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`

```go
func (d Database) InitDB(v *bool) error {
    tables, err := ReadSchemaFiles(v)
    if err != nil {
        return err
    }
    if _, err := d.Connection.ExecContext(d.Context, tables); err != nil {
        return err
    }
    return nil
}
```

The `ReadSchemaFiles` function walks the embedded filesystem:

```go
func ReadSchemaFiles(verbose *bool) (string, error) {
    var result []string

    err := fs.WalkDir(SqlFiles, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() && filepath.Base(path) == "schema.sql" {
            data, err := SqlFiles.ReadFile(path)
            if err != nil {
                return fmt.Errorf("failed to read file %s: %w", path, err)
            }
            result = append(result, string(data))
        }
        return nil
    })
    if err != nil {
        return "", err
    }
    if *verbose {
        fmt.Println(strings.Join(result, "\n"))
    }
    return strings.Join(result, "\n"), nil
}
```

**Process:**
1. Walk embedded filesystem looking for `schema.sql` files
2. Read all schema files into a string slice
3. Join with newlines
4. Execute combined SQL against the database

**Note:** This currently reads the base `schema.sql` files (SQLite). For MySQL/PostgreSQL deployments, this should read `schema_mysql.sql` or `schema_psql.sql` instead. This is a known limitation.

---

## Adding Support for a New Database

**Steps to add a fourth database engine (e.g., CockroachDB):**

1. **Add driver constant** in `/Users/home/Documents/Code/Go_dev/modulacms/internal/config/config.go`:
   ```go
   const (
       Sqlite DbDriver = "sqlite"
       Mysql  DbDriver = "mysql"
       Psql   DbDriver = "postgres"
       CockroachDB DbDriver = "cockroachdb"  // New
   )
   ```

2. **Create driver struct** in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go`:
   ```go
   type CockroachDatabase struct {
       Src            string
       Status         DbStatus
       Connection     *sql.DB
       LastConnection string
       Err            error
       Config         config.Config
       Context        context.Context
   }
   ```

3. **Implement DbDriver interface methods** (150+ methods)

4. **Add GetDb method** in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`

5. **Update ConfigDB switch** in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/init.go`:
   ```go
   case config.CockroachDB:
       d := CockroachDatabase{Src: env.Db_Name, Config: env}
       dbc := d.GetDb(&verbose)
       return dbc
   ```

6. **Create schema files** for all migrations:
   ```
   sql/schema/1_permissions/schema_cockroachdb.sql
   sql/schema/2_roles/schema_cockroachdb.sql
   ...
   ```

7. **Add sqlc configuration** in `/Users/home/Documents/Code/Go_dev/modulacms/sql/sqlc.yml`

8. **Create query directory:** `sql/cockroachdb/`

9. **Run sqlc generation:** `just sqlc`

10. **Test all operations** with the new driver

---

## Performance Considerations

### Query Performance by Database

**SQLite:**
- Excellent for reads (in-memory page cache)
- Single-writer bottleneck for writes
- No query planner tuning needed
- Good for < 100 concurrent connections

**MySQL:**
- InnoDB buffer pool critical for performance
- Query cache (deprecated in 8.0+)
- Excellent for read-heavy workloads
- Handles 1000+ concurrent connections

**PostgreSQL:**
- Sophisticated query planner
- Excellent for complex queries
- Superior for write-heavy workloads
- Handles 1000+ concurrent connections

### Index Strategies

ModulaCMS schema includes indexes on:
- Primary keys (automatic)
- Foreign keys (recommended for performance)
- Tree traversal columns (`parent_id`, `first_child_id`, `next_sibling_id`, `prev_sibling_id`)
- Lookup columns (`route_id`, `datatype_id`)

**Database-specific index considerations:**

**SQLite:**
```sql
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_parent ON content_data(parent_id);
```

**MySQL:**
```sql
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_parent ON content_data(parent_id);
-- InnoDB automatically indexes foreign keys
```

**PostgreSQL:**
```sql
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_parent ON content_data(parent_id);
-- Consider CONCURRENTLY for live systems:
CREATE INDEX CONCURRENTLY idx_name ON table(column);
```

---

## Configuration Best Practices

### Development Configuration

**Use SQLite for development:**
```json
{
  "environment": "development",
  "db_driver": "sqlite",
  "db_url": "./modula.db"
}
```

**Benefits:**
- Fast iteration (no network overhead)
- Simple setup (no separate database server)
- Easy to reset (delete .db file)
- Version control friendly (small file size)

### Production Configuration

**Use MySQL or PostgreSQL for production:**

**MySQL example:**
```json
{
  "environment": "production",
  "db_driver": "mysql",
  "db_url": "db.example.com:3306",
  "db_name": "modulacms_prod",
  "db_username": "cms_app",
  "db_password": "${DB_PASSWORD}"
}
```

**PostgreSQL example:**
```json
{
  "environment": "production",
  "db_driver": "postgres",
  "db_url": "db.example.com:5432",
  "db_name": "modulacms_prod",
  "db_username": "cms_app",
  "db_password": "${DB_PASSWORD}"
}
```

**Security recommendations:**
- Use environment variables for passwords
- Use database user with minimal permissions
- Enable SSL/TLS for database connections (update connection strings)
- Separate read and write database users if needed

### Staging/Testing Configuration

**Option 1: Use separate database on same server:**
```json
{
  "environment": "staging",
  "db_driver": "mysql",
  "db_url": "db.example.com:3306",
  "db_name": "modulacms_staging"
}
```

**Option 2: Use SQLite for fast test cycles:**
```json
{
  "environment": "staging",
  "db_driver": "sqlite",
  "db_url": "./staging.db"
}
```

---

## Troubleshooting

### Common Issues

#### SQLite: Foreign Key Constraints Not Enforced

**Symptom:** Can delete parent records that should be protected by foreign keys.

**Cause:** Foreign keys are disabled by default in SQLite.

**Solution:** The `GetDb` method for SQLite automatically runs `PRAGMA foreign_keys=1;`. Verify this is executing:
```go
_, err = db.Exec("PRAGMA foreign_keys=1;")
```

#### MySQL: Connection Refused

**Symptom:** `failed to connect to MySQL database: dial tcp: connection refused`

**Causes:**
- MySQL server not running
- Wrong host or port in config
- Firewall blocking connection

**Solution:**
```bash
# Test MySQL connection manually
mysql -h localhost -P 3306 -u cms_user -p

# Check MySQL is running
systemctl status mysql

# Check port
netstat -tlnp | grep 3306
```

#### PostgreSQL: Password Authentication Failed

**Symptom:** `failed to connect to PostgreSQL database: password authentication failed`

**Causes:**
- Wrong username or password
- PostgreSQL pg_hba.conf not configured for password auth
- User doesn't have permissions on database

**Solution:**
```bash
# Test PostgreSQL connection manually
psql -h localhost -p 5432 -U cms_user -d modulacms_prod

# Check pg_hba.conf for authentication method
sudo nano /etc/postgresql/14/main/pg_hba.conf

# Grant permissions
GRANT ALL PRIVILEGES ON DATABASE modulacms_prod TO cms_user;
```

#### Connection Pool Exhaustion

**Symptom:** Operations hang or timeout under high load.

**Cause:** Too many concurrent database operations, not enough connections in pool.

**Solution:** Configure connection pool limits:
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
```

#### Schema Mismatch Between Databases

**Symptom:** Application works with SQLite but fails with MySQL/PostgreSQL.

**Cause:** Schema files not properly synced between database engines.

**Solution:**
- Verify all three schema files exist for each migration
- Check for syntax differences (AUTO_INCREMENT vs SERIAL)
- Regenerate sqlc code: `just sqlc`
- Test with all three databases in CI/CD

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - DbDriver interface philosophy
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Database schema relationships

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - SQL file organization
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - sqlc usage and configuration
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database package structure

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Creating new database tables
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Database testing strategies

---

## Quick Reference

### Switching Database Engines

**1. Stop the application**
```bash
# Send SIGTERM to gracefully shut down
kill -SIGTERM <pid>
```

**2. Update configuration**
```json
{
  "db_driver": "mysql",  // Change from "sqlite"
  "db_url": "localhost:3306",
  "db_name": "modulacms",
  "db_username": "cms_user",
  "db_password": "secure_password"
}
```

**3. Initialize new database**
```bash
# Create database in MySQL
mysql -u root -p
CREATE DATABASE modulacms;
GRANT ALL PRIVILEGES ON modulacms.* TO 'cms_user'@'localhost';
```

**4. Restart application**
```bash
./modulacms
```

Application will automatically run schema initialization on first connection.

### Database Driver Import Paths

**Go database drivers used:**
```go
import (
    _ "github.com/mattn/go-sqlite3"      // SQLite (CGO required)
    _ "github.com/go-sql-driver/mysql"   // MySQL
    _ "github.com/lib/pq"                // PostgreSQL
)
```

**Blank imports required:** These drivers register themselves with `database/sql` when imported.

### Configuration Keys

| Key | SQLite | MySQL | PostgreSQL |
|-----|--------|-------|------------|
| `db_driver` | `"sqlite"` | `"mysql"` | `"postgres"` |
| `db_url` | File path | `host:port` | `host:port` |
| `db_name` | *(unused)* | Database name | Database name |
| `db_username` | *(unused)* | Username | Username |
| `db_password` | *(unused)* | Password | Password |

### Performance Comparison

| Operation | SQLite | MySQL | PostgreSQL |
|-----------|--------|-------|------------|
| Simple SELECT | ⚡ Excellent | ⚡ Excellent | ⚡ Excellent |
| Complex JOIN | ⚡ Good | ⚡⚡ Excellent | ⚡⚡⚡ Outstanding |
| Concurrent Reads | ⚡ Good | ⚡⚡ Excellent | ⚡⚡ Excellent |
| Concurrent Writes | ⚠️ Limited | ⚡⚡ Excellent | ⚡⚡⚡ Outstanding |
| Bulk INSERT | ⚡ Good | ⚡⚡ Excellent | ⚡⚡ Excellent |
| Tree Traversal | ⚡ Excellent | ⚡ Excellent | ⚡ Excellent |

**Recommendation:** SQLite for development and small deployments (< 10,000 nodes). MySQL or PostgreSQL for production (> 10,000 nodes or high concurrency).

---

**Last Updated:** 2026-01-12
