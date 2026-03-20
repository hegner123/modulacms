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
modula --gen-certs
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

## Media and S3 Storage

### dial tcp: lookup media.minio: no such host

```
upload original to S3: S3 upload: failed to upload to S3: RequestError: send request failed
caused by: Put "http://media.minio:9000/...": dial tcp: lookup media.minio: no such host
```

The S3 client is using **virtual-hosted-style** URLs, prepending the bucket name to the endpoint hostname (`media.minio:9000` instead of `minio:9000/media`). Docker DNS cannot resolve the bucket-prefixed hostname.

Set `bucket_force_path_style` to `true` in `modula.config.json`:

```json
{
  "bucket_force_path_style": true
}
```

This switches to **path-style** URLs (`endpoint/bucket/key`), which is required for MinIO and most S3-compatible services in Docker or local development. The project default is `true`, but an explicit `false` in your config file overrides it.

### Failed to ensure media bucket / Failed to ensure backup bucket

```
[WARN] Failed to ensure media bucket: create bucket "media": RequestError: send request failed
```

S3 is unreachable at startup. The server continues running, but uploads return 503.

1. Check that the S3 service (MinIO, AWS, etc.) is running and reachable from the CMS container
2. Verify `bucket_endpoint` matches the S3 hostname -- in Docker, use the container/service name (e.g., `minio:9000`), not `localhost`
3. Verify `bucket_access_key` and `bucket_secret_key` are correct
4. Verify `bucket_force_path_style` is `true` for MinIO or local S3-compatible services

### S3 storage must be configured for media uploads

Upload rejected because S3 credentials are missing. Set all required fields in `modula.config.json`:

```json
{
  "bucket_media": "media",
  "bucket_endpoint": "minio:9000",
  "bucket_access_key": "your-access-key",
  "bucket_secret_key": "your-secret-key",
  "bucket_force_path_style": true
}
```

Admin media uploads fall back to the shared bucket config. If you have separate admin media storage, set `bucket_admin_media`, `bucket_admin_endpoint`, `bucket_admin_access_key`, and `bucket_admin_secret_key`.

### File too large or invalid multipart form

The uploaded file exceeds `max_upload_size`. The default is 10 MB (10485760 bytes). Increase it:

```json
{
  "max_upload_size": 52428800
}
```

The value is in bytes. 52428800 = 50 MB.

### upload original to S3: S3 upload failed

S3 upload failed after the file was received and processed. Common causes:

- **Bucket doesn't exist:** The bucket named in `bucket_media` was not created. Start the server once to auto-create buckets, or create them manually in the MinIO console.
- **Permission denied:** The access key lacks `s3:PutObject` permission on the bucket.
- **Network timeout:** The S3 endpoint is slow or unreachable. Check container networking.
- **Disk full on S3:** The storage backend ran out of space.

Check server logs for the full wrapped error message.

### image width/height exceeds maximum

```
image width 25000 exceeds maximum 20000
```

The uploaded image dimensions exceed safety limits. These are hardcoded to prevent memory exhaustion during image processing:

| Limit | Value |
|-------|-------|
| Maximum width | 20,000 pixels |
| Maximum height | 20,000 pixels |
| Maximum total pixels | 500 megapixels |

Resize the image before uploading.

### unsupported file extension

Image optimization only supports `.png`, `.jpg`, `.jpeg`, `.webp`, and `.gif`. Other image formats (`.bmp`, `.tiff`, `.svg`) are stored as-is without generating dimension variants.

### cannot delete folder: contains N child folder(s) and M media item(s)

Folders must be empty before deletion. Either delete or move the contents first:

```bash
# Move media out of the folder (to root)
curl -X POST http://localhost:8080/api/v1/media/move \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"media_ids": ["01ABC..."], "folder_id": null}'

# Delete child folders first, then the parent
curl -X DELETE http://localhost:8080/api/v1/media-folders/CHILD_FOLDER_ID \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### creating this folder would exceed maximum folder depth of 10

Folders nest up to 10 levels deep. Flatten your folder hierarchy or move the parent folder to a shallower level before creating children.

### Orphaned S3 objects after failed uploads or deletions

When an upload fails partway through, or a media deletion cannot reach S3, objects may be left in the bucket without a matching database record.

Scan for orphans:

```bash
curl http://localhost:8080/api/v1/media/health \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

Clean up orphans:

```bash
curl -X DELETE http://localhost:8080/api/v1/media/cleanup \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

The cleanup endpoint deletes S3 objects that have no corresponding database record. It returns a count of deleted objects and any that failed to delete.

### Media URLs show Docker hostname instead of public URL

API responses contain URLs like `http://minio:9000/media/photo.jpg` that browsers cannot resolve.

Set `bucket_public_url` to the externally reachable address:

```json
{
  "bucket_endpoint": "minio:9000",
  "bucket_public_url": "http://localhost:9000"
}
```

`bucket_endpoint` is used for server-side S3 API calls (Docker-internal hostname is fine). `bucket_public_url` is used in API responses that reach the browser.

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
| lookup media.minio: no such host | Set `"bucket_force_path_style": true` in config |
| S3 storage must be configured | Set `bucket_access_key`, `bucket_secret_key`, `bucket_endpoint` in config |
| File too large | Increase `max_upload_size` in config (value in bytes) |
| Folder depth exceeded | Flatten folder hierarchy (max 10 levels) |
| Cannot delete folder with children | Delete or move folder contents first |
| Orphaned S3 objects | Run `GET /api/v1/media/health` then `DELETE /api/v1/media/cleanup` |
| Media URLs show Docker hostname | Set `bucket_public_url` to externally reachable address |

### Emergency Recovery Commands

```bash
# Kill a hung process
pkill -9 modulacms

# Reset the database (destructive -- removes all data)
rm modula.db
modula serve

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
