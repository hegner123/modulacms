# Troubleshooting

Find solutions to common errors by searching for the error message you're seeing.

## Build and Compilation Errors

### cgo: C compiler "gcc" not found

```
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH
```

The SQLite driver requires CGO, which needs a C compiler.

**macOS:**

```bash
xcode-select --install
```

**Linux (Debian/Ubuntu):**

```bash
sudo apt-get install build-essential
```

**Linux (RHEL/CentOS):**

```bash
sudo yum groupinstall "Development Tools"
```

### cannot find package "github.com/mattn/go-sqlite3"

The vendor directory is out of sync or missing. Restore it:

```bash
go mod vendor
go build ./cmd
```

If the vendor directory is corrupted, delete and recreate it:

```bash
rm -rf vendor/
go mod tidy
go mod vendor
go build ./cmd
```

### Cross-compilation fails with CGO

Cross-compiling CGO code from macOS to Linux requires a cross-compiler:

```bash
brew install FiloSottile/musl-cross/musl-cross
```

Then build with:

```bash
CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ \
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
go build -mod vendor -o modulacms-amd ./cmd
```

Alternatively, build directly on the target Linux server.

### pattern ./...: directory prefix . does not contain main module

You are running the command from the wrong directory. Ensure you are in the project root where `go.mod` lives:

```bash
cd /path/to/modulacms
ls go.mod  # should exist
go test ./...
```

## Database Connection Errors

### failed to open SQLite database: unable to open database file

The database file path is invalid, the directory doesn't exist, or permissions are wrong.

```bash
# Check that the directory exists
mkdir -p $(dirname /path/to/modula.db)

# Check permissions
chmod 755 $(dirname /path/to/modula.db)

# Verify the path in modula.config.json
cat modula.config.json | jq '.db_url'
```

If `db_url` is not set, the default is `./modula.db` in the current working directory.

### failed to connect to MySQL: connection refused

MySQL is not running or the host/port is incorrect.

```bash
# Check if MySQL is running
mysql -u root -p -h localhost -P 3306

# Start MySQL
brew services start mysql   # macOS
sudo systemctl start mysql  # Linux

# Verify config
cat modula.config.json | jq '.db_url, .db_name, .db_username'
```

### Access denied for user 'root'@'localhost'

Incorrect username or password in `modula.config.json`. Verify the credentials:

```json
{
  "db_driver": "mysql",
  "db_url": "localhost:3306",
  "db_name": "modulacms",
  "db_username": "your_user",
  "db_password": "your_password"
}
```

### failed to connect to PostgreSQL: pq: SSL is not enabled

PostgreSQL requires SSL but the connection string has it disabled (or vice versa).

For development, append `?sslmode=disable` to your connection string or configure it in the database initialization code. For production, configure SSL certificates.

## sqlc Generation Errors

### sqlc: unknown driver "sqlite"

Your sqlc version is outdated. Upgrade to 1.20.0+:

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc version
```

### query parameter references column that is not in the query

The SQL query is missing a parameter that the sqlc annotation expects. Add a WHERE clause or other parameter usage:

```sql
-- Wrong: no parameter
-- name: GetContentData :one
SELECT * FROM content_data;

-- Correct: includes parameter
-- name: GetContentData :one
SELECT * FROM content_data WHERE content_data_id = ?;
```

### column reference is ambiguous

Multiple tables in a JOIN have the same column name without a table prefix:

```sql
-- Wrong
SELECT content_data_id FROM content_data
JOIN content_fields ON content_fields.content_data_id = content_data.content_data_id;

-- Correct
SELECT cd.content_data_id FROM content_data cd
JOIN content_fields cf ON cf.content_data_id = cd.content_data_id;
```

### sqlc.yml not found

sqlc must be run from the `sql/` directory:

```bash
cd sql
sqlc generate
```

Or use the just target which handles the directory change:

```bash
just sqlc
```

## Foreign Key Constraint Errors

### FOREIGN KEY constraint failed

You're inserting or updating a record that references a parent record that doesn't exist.

Check that the referenced records exist before the operation:

```sql
SELECT datatype_id FROM datatypes WHERE datatype_id = 5;
SELECT route_id FROM routes WHERE route_id = 1;
```

Create parent records before child records. When creating content, the route, datatype, and author must already exist.

### Cannot delete or update a parent row

You're trying to delete a record that has child records referencing it.

Delete child records first, or ensure the schema uses `ON DELETE CASCADE`:

```sql
-- Delete children first
DELETE FROM content_fields WHERE content_data_id = 123;
-- Then delete parent
DELETE FROM content_data WHERE content_data_id = 123;
```

## Server Startup Errors

### bind: address already in use

Another process is using the port.

```bash
# Find the process
lsof -i :8080

# Kill it
kill <PID>
```

Or change the port in `modula.config.json`:

```json
{
  "port": ":8081",
  "ssl_port": ":4444",
  "ssh_port": "2234"
}
```

### Certificate Directory path is invalid

The certificate directory specified in `cert_dir` doesn't exist.

```bash
mkdir -p ./certs
chmod 755 ./certs
```

For local development, generate self-signed certificates:

```bash
./modulacms-x86 --gen-certs
```

### ssh: no host key available

The SSH host key doesn't exist. Generate one:

```bash
mkdir -p .ssh
ssh-keygen -t ed25519 -f .ssh/id_ed25519 -N ""
```

## TUI Errors

### Screen is blank or frozen

Possible causes:
- No active PTY (non-interactive SSH session)
- Terminal dimensions are zero
- Panic in View() or Update()

Check the log output for panic messages. Ensure you are connecting with an interactive SSH session that allocates a PTY.

### Garbled output or visible escape codes

The terminal doesn't support ANSI colors. Check your TERM environment variable:

```bash
echo $TERM
# Should be xterm-256color or similar
export TERM=xterm-256color
```

### TUI stops responding to input

The Update function returns nil for `tea.Cmd` when it should return a command. Every asynchronous operation must return a command that eventually produces a message. Check the Update function for code paths that return `nil` as the command.

## Tree Loading Errors

### Circular reference detected

A content node references itself or one of its own descendants as its parent. Find and fix the cycle:

```sql
-- Break the cycle by clearing the parent
UPDATE content_data SET parent_id = NULL WHERE content_data_id = 123;
```

### Orphaned nodes detected

A node's parent references a record that doesn't exist. Find orphaned nodes:

```sql
SELECT cd.content_data_id, cd.parent_id
FROM content_data cd
LEFT JOIN content_data parent ON cd.parent_id = parent.content_data_id
WHERE cd.parent_id IS NOT NULL
  AND parent.content_data_id IS NULL;
```

Fix by correcting the parent or setting it to NULL.

### Tree loads slowly

Large trees with thousands of nodes can be slow. Add database indexes on frequently queried columns:

```sql
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_parent ON content_data(parent_id);
```

## Authentication and OAuth Errors

### Missing code parameter (OAuth callback)

The OAuth provider didn't return an authorization code. The user may have cancelled the flow, or the provider configuration is incorrect.

Verify the OAuth callback URL registered with your provider matches exactly: `http://localhost:8080/api/v1/auth/oauth/callback` (including scheme and port).

### Error exchanging token

Invalid client credentials or incorrect token endpoint URL. Verify `oauth_client_id`, `oauth_client_secret`, and `oauth_endpoint` fields in `modula.config.json`.

### sessions don't match

The session cookie is invalid, corrupted, or expired. Clear the browser cookie and log in again.

To check session expiration in the database:

```sql
SELECT session_id, user_id, expires_at FROM sessions WHERE user_id = 123;
```

## Runtime Errors

### nil pointer dereference

Accessing a field or method on a nil pointer. Common in tree operations where a node's parent, child, or sibling may be nil. Check for nil before dereferencing.

### too many open files

The file descriptor limit is too low. Increase it:

```bash
ulimit -n 4096
```

For a permanent fix on Linux, edit `/etc/security/limits.conf`. Close all database connections and HTTP response bodies with `defer`.

### too many SQL variables

SQLite limits queries to 999 variables. Batch large operations into groups of 900 or fewer.

## Quick Reference Table

| Error | Fix |
|-------|-----|
| C compiler not found | Install gcc: `xcode-select --install` (macOS) or `apt-get install build-essential` (Linux) |
| Database file not found | Check `db_url` path in `modula.config.json` and create the directory |
| Foreign key constraint failed | Verify parent record exists before insert |
| Port already in use | `lsof -i :8080` then `kill <PID>`, or change port in config |
| Session expired | Clear browser cookies and log in again |
| sqlc generation fails | Run `just sqlc` from the project root |
| Nil pointer panic | Add nil check before accessing the value |

### Emergency Recovery Commands

```bash
# Kill a hung process
pkill -9 modulacms

# Reset the database (destructive -- removes all data)
rm modula.db
./modulacms-x86 serve

# Clear test artifacts
rm -rf testdb/*.db backups/*.zip

# Regenerate all sqlc code
just sqlc

# Full clean rebuild
just clean
just dev

# View application logs
sudo journalctl -u modulacms -f
```
