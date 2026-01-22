# QUICKSTART.md

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/QUICKSTART.md`

**Purpose:** This document provides a fast-track guide to getting ModulaCMS running on your local machine, creating your first content, and understanding the basic development workflow. Start here if you want to go from zero to working CMS in under 10 minutes.

**Last Updated:** 2026-01-12

---

## Prerequisites

Before starting, ensure you have:

- **Go 1.23.0 or later** (project uses toolchain go1.24.2)
- **Git** for cloning the repository
- **SQLite** (built into Go with CGO, no separate install needed for development)
- **Make** for running build commands
- **Text editor** (vim, emacs, nano, VS Code, whatever you prefer)

**Optional for production:**
- MySQL 5.7+ or PostgreSQL 12+
- S3-compatible storage credentials (AWS S3, Linode, DigitalOcean Spaces, etc.)
- SSH client for remote management

**Check Go version:**
```bash
go version
# Should show go1.23 or later
```

---

## Quick Start (5 Minutes)

### 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/hegner123/modulacms.git
cd modulacms

# Build for local development (creates modulacms-x86 binary)
make dev

# OR build for production (creates both x86 and AMD64 binaries)
make build
```

**What this does:**
- `make dev` - Quick local build using Go modules, outputs `modulacms-x86` in project root
- `make build` - Full production build, outputs to `out/bin/`, cross-compiles for AMD64 Linux

**Build time:** ~30 seconds on first build (downloads dependencies), ~5 seconds on subsequent builds

### 2. Run Installation Wizard

```bash
./modulacms-x86 --install
```

**What happens:**
1. Installation wizard asks where to create config file
2. Prompts for database settings (default: SQLite in current directory)
3. Prompts for port configuration (default: 4000)
4. Prompts for site domains (default: localhost:1234)
5. Creates `config.json` with your settings
6. Initializes database with schema migrations
7. Creates default admin user (if applicable)

**Example installation:**
```
Would you like to install ModulaCMS at
/Users/you/modulacms
? (y/N) y

Database driver? (sqlite/mysql/postgres) sqlite
Database path? ./modula.db
HTTP port? 4000
HTTPS port? 4000
Client site domain? localhost:1234
Admin site domain? admin.localhost:1234

✓ Configuration created at config.json
✓ Database initialized
✓ Ready to start!
```

### 3. Start the TUI

```bash
./modulacms-x86 --cli
```

**What you see:**
- Terminal UI with keyboard navigation
- Welcome screen with options
- Database status indicator
- Navigation menu

**First-time setup in TUI:**
1. Navigate to "Routes" - Create your first site/route
2. Navigate to "Datatypes" - Define content types (Page, Post, etc.)
3. Navigate to "Fields" - Add fields to datatypes (Title, Body, Image, etc.)
4. Navigate to "Content" - Create your first content

### 4. Create Your First Content

**Step-by-step in TUI:**

**A. Create a Route (Site Root)**
```
1. Press 'r' for Routes menu
2. Press 'n' to create new route
3. Fill in form:
   - Label: "Main Site"
   - Slug: "main"
   - Domain: "localhost:1234"
4. Press Enter to save
```

**B. Create a Datatype (Content Schema)**
```
1. Press 'd' for Datatypes menu
2. Press 'n' to create new datatype
3. Fill in form:
   - Label: "Page"
   - Type: "ROOT" (for top-level content)
4. Press Enter to save
```

**C. Add Fields to Datatype**
```
1. Press 'f' for Fields menu
2. Press 'n' to create new field
3. Add Title field:
   - Label: "Title"
   - Type: "text"
   - Parent: Select "Page" datatype
4. Press Enter to save
5. Repeat for Body field:
   - Label: "Body"
   - Type: "richtext"
   - Parent: Select "Page" datatype
```

**D. Create Content**
```
1. Press 'c' for Content menu
2. Press 'n' to create new content
3. Select route: "Main Site"
4. Select datatype: "Page"
5. Fill in fields:
   - Title: "Welcome to ModulaCMS"
   - Body: "Your first page content here"
6. Press Enter to save
```

**Result:** You now have a content tree with one page that can be served via HTTP.

---

## Running Tests

### Run All Tests

```bash
make test
```

**What this does:**
1. Cleans up test database files
2. Removes old backup files
3. Runs all tests in the project
4. Cleans up test artifacts after completion

**Expected output:**
```
touch testdb/create_tests.db
touch ./testdb/testing2348263.db
rm ./testdb/*.db
rm ./backups/*.zip
go test -v ./...
?       github.com/hegner123/modulacms/cmd    [no test files]
ok      github.com/hegner123/modulacms/internal/backup    0.234s
ok      github.com/hegner123/modulacms/internal/db-sqlite 0.521s
...
rm ./testdb/*.db
```

### Run Specific Package Tests

```bash
# Test database package only
go test -v ./internal/db-sqlite

# Test media package only
go test -v ./internal/media

# Test TUI/CLI package only
go test -v ./internal/cli
```

### Run Coverage Report

```bash
make coverage
```

**Output:** Shows test coverage percentage for each package.

### Run Development Tests Only

```bash
make test-development
```

**Use case:** Quick tests during active development without full test suite.

---

## Development Workflow

### Typical Day-to-Day Workflow

**1. Make Code Changes**
```bash
# Edit files in internal/, cmd/, or sql/
vim internal/model/content.go
```

**2. Update Database Queries (if needed)**
```bash
# Edit SQL files in sql/mysql/, sql/postgres/, or sql/schema/
vim sql/mysql/content.sql

# Regenerate Go code from SQL
make sqlc
```

**3. Rebuild and Test**
```bash
# Quick rebuild
make dev

# Run tests
make test

# Run the binary
./modulacms-x86 --cli
```

**4. Iterate**

Repeat steps 1-3 until feature is complete.

### Hot Reload for Active Development

```bash
# Clear debug log before starting
echo "" > debug.log

# Build and run
make run

# In another terminal, tail the debug log
tail -f debug.log
```

**Note:** ModulaCMS doesn't have built-in hot reload. Use `make run` to rebuild and restart manually.

---

## Common Commands Reference

### Build Commands

```bash
make dev              # Fast local build (x86)
make build            # Production build (x86 + AMD64)
make run              # Build and immediately run
make clean            # Remove build artifacts
```

### Test Commands

```bash
make test             # Run all tests
make test-development # Run development package tests only
make coverage         # Run tests with coverage report
make template-test    # Run template-specific tests
```

### Database Commands

```bash
make sqlc             # Generate Go code from SQL queries
make dump             # Dump SQLite database to SQL file
make docker-db        # Start database containers
```

### Code Quality Commands

```bash
make lint             # Run all linters
make lint-go          # Lint Go code only
make vendor           # Update vendor directory
```

### Running the Application

```bash
./modulacms-x86 --cli          # Start TUI in CLI mode
./modulacms-x86 --install      # Run installation wizard
./modulacms-x86 --version      # Show version
./modulacms-x86 --config=/path # Use custom config file
```

---

## Making Your First Code Change

Let's add a simple feature to understand the workflow.

### Example: Add a "Featured" Flag to Content

**Step 1: Update Database Schema**

Create new migration directory:
```bash
mkdir -p sql/schema/25_featured_content
```

Create schema file: `sql/schema/25_featured_content/schema.sql`
```sql
ALTER TABLE content_data ADD COLUMN featured INTEGER DEFAULT 0;
```

Update combined schemas:
```bash
# Append to all_schema.sql, all_schema_mysql.sql, all_schema_psql.sql
echo "ALTER TABLE content_data ADD COLUMN featured INTEGER DEFAULT 0;" >> sql/all_schema.sql
```

**Step 2: Add SQL Query**

Edit `sql/mysql/content.sql` (or your database-specific file):
```sql
-- name: SetContentFeatured :exec
UPDATE content_data
SET featured = ?
WHERE content_data_id = ?;

-- name: ListFeaturedContent :many
SELECT * FROM content_data
WHERE featured = 1 AND route_id = ?
ORDER BY date_modified DESC;
```

**Step 3: Generate Go Code**

```bash
make sqlc
```

**Result:** New methods added to database driver interfaces:
- `SetContentFeatured(featured int, id int64) error`
- `ListFeaturedContent(routeID int64) ([]ContentData, error)`

**Step 4: Use in Application Code**

Edit `internal/model/content.go`:
```go
func (c *Content) SetFeatured(featured bool) error {
    val := 0
    if featured {
        val = 1
    }
    return db.Driver.SetContentFeatured(val, c.ID)
}

func GetFeaturedContent(routeID int64) ([]Content, error) {
    rows, err := db.Driver.ListFeaturedContent(routeID)
    if err != nil {
        return nil, err
    }
    // Transform rows to Content structs
    return transformToContent(rows), nil
}
```

**Step 5: Add TUI Command**

Edit `internal/cli/update.go`:
```go
case tea.KeyMsg:
    switch msg.String() {
    case "f": // Toggle featured
        return m, ToggleFeaturedCmd{NodeID: m.SelectedNode.ID}
    }
```

**Step 6: Test**

```bash
# Run tests
make test

# Build and run
make run

# In TUI, press 'f' on selected content to toggle featured status
```

**Result:** You've added a database column, queries, application logic, and a TUI command. This is the typical workflow for new features.

---

## Production Deployment

### Quick Production Setup

**1. Build for Production**
```bash
make build
```

**Output:**
- `out/bin/modulacms-x86` - Local x86 binary
- `out/bin/modulacms-amd` - Linux AMD64 binary

**2. Configure for Production**

Create production config.json:
```json
{
  "environment": "production",
  "db_driver": "mysql",
  "db_url": "your-db-host:3306",
  "db_name": "modulacms_prod",
  "db_username": "modulacms_user",
  "db_password": "your-secure-password",
  "port": "80",
  "ssl_port": "443",
  "client_site": "yoursite.com",
  "admin_site": "admin.yoursite.com",
  "bucket_url": "us-east-1.linodeobjects.com",
  "bucket_endpoint": "yoursite.us-east-1.linodeobjects.com",
  "bucket_access_key": "YOUR_ACCESS_KEY",
  "bucket_secret_key": "YOUR_SECRET_KEY",
  "bucket_media": "media",
  "bucket_backup": "backups",
  "oauth_client_id": "your-oauth-client-id",
  "oauth_client_secret": "your-oauth-client-secret",
  "oauth_endpoint": {
    "oauth_auth_url": "https://your-oauth-provider.com/oauth/authorize",
    "oauth_token_url": "https://your-oauth-provider.com/oauth/token"
  },
  "cors_origins": ["https://yoursite.com", "https://admin.yoursite.com"],
  "cors_methods": ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
  "cors_headers": ["Content-Type", "Authorization"],
  "cors_credentials": true
}
```

**3. Deploy Binary**

```bash
# Upload binary to server
scp out/bin/modulacms-amd user@yourserver:/opt/modulacms/modulacms

# Upload config
scp config.json user@yourserver:/opt/modulacms/config.json

# SSH into server
ssh user@yourserver

# Run installation
cd /opt/modulacms
./modulacms --install --config=config.json
```

**4. Run as Service**

Create systemd service: `/etc/systemd/system/modulacms.service`
```ini
[Unit]
Description=ModulaCMS Server
After=network.target

[Service]
Type=simple
User=modulacms
WorkingDirectory=/opt/modulacms
ExecStart=/opt/modulacms/modulacms --config=/opt/modulacms/config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Start service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable modulacms
sudo systemctl start modulacms
sudo systemctl status modulacms
```

**5. Verify Deployment**

```bash
# Check HTTP server
curl http://yoursite.com

# Check HTTPS server (after Let's Encrypt provisions certs)
curl https://yoursite.com

# SSH into TUI
ssh modulacms@yoursite.com
```

---

## Troubleshooting Common Issues

### Build Errors

**Problem:** `cannot find package`
```
internal/db/db.go:5:2: cannot find package "github.com/mattn/go-sqlite3"
```

**Solution:** Install dependencies
```bash
go mod download
go mod vendor
make dev
```

---

**Problem:** `CGO_ENABLED required for SQLite`
```
error: undefined: sqlite3.SQLiteDriver
```

**Solution:** Ensure CGO is enabled
```bash
export CGO_ENABLED=1
make dev
```

---

### Runtime Errors

**Problem:** `failed to load configuration`
```
Failed to load configuration: open config.json: no such file or directory
```

**Solution:** Run installation wizard or specify config path
```bash
./modulacms-x86 --install
# OR
./modulacms-x86 --config=/path/to/config.json --cli
```

---

**Problem:** `database schema out of date`
```
Error: migration 23 not applied
```

**Solution:** Re-run installation to apply migrations
```bash
./modulacms-x86 --install
```

---

### Test Failures

**Problem:** Tests fail with database errors
```
Error: unable to open database file
```

**Solution:** Clean test artifacts and retry
```bash
rm -rf testdb/*.db backups/*.zip
make test
```

---

### TUI Issues

**Problem:** TUI doesn't render properly
```
Characters appear as boxes or garbled output
```

**Solution:** Check terminal compatibility
```bash
# Ensure TERM is set
echo $TERM

# Try different terminal emulator
# iTerm2, Alacritty, or Kitty recommended

# Check Bubbletea compatibility
export TERM=xterm-256color
./modulacms-x86 --cli
```

---

## What's Next?

After completing this quickstart, explore these resources:

**Core Concepts:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Understanding content hierarchy
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - How datatypes and fields work

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Adding new features end-to-end
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Adding database tables
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/CREATING_TUI_SCREENS.md` - Building new TUI screens

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI development guide
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database operations

**Domain Knowledge:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/ROUTES_AND_SITES.md` - Multi-site setup
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/DATATYPES_AND_FIELDS.md` - Content modeling
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/MEDIA_SYSTEM.md` - Media uploads and S3

**Development Guidelines:**
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Code style, conventions, and build commands

---

## Related Documentation

**Getting Started:**
- `/Users/home/Documents/Code/Go_dev/modulacms/README.md` - Project overview and philosophy
- `/Users/home/Documents/Code/Go_dev/modulacms/START.md` - Documentation index and onboarding

**Reference:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/PATTERNS.md` - Common code patterns
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/TROUBLESHOOTING.md` - Detailed troubleshooting
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/DEPENDENCIES.md` - Dependency rationale

**Testing:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Testing strategies and patterns
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEBUGGING.md` - Debugging guide

---

## Quick Reference

### Essential Commands

```bash
# Development
make dev              # Build for local development
make run              # Build and run immediately
make test             # Run all tests

# Database
make sqlc             # Regenerate Go code from SQL
make dump             # Export SQLite to SQL file

# Application
./modulacms-x86 --cli     # Start TUI
./modulacms-x86 --install # Run installation wizard
./modulacms-x86 --version # Show version
```

### Project Structure

```
modulacms/
├── cmd/main.go                  # Application entry point
├── internal/
│   ├── cli/                     # TUI implementation
│   ├── db/                      # Database interface
│   ├── db-sqlite/               # SQLite driver
│   ├── db-mysql/                # MySQL driver
│   ├── db-psql/                 # PostgreSQL driver
│   ├── model/                   # Business logic
│   ├── install/                 # Installation wizard
│   └── middleware/              # HTTP middleware
├── sql/
│   ├── schema/                  # Database migrations
│   ├── mysql/                   # MySQL queries
│   └── postgres/                # PostgreSQL queries
├── Makefile                     # Build commands
├── config.json                  # Configuration file
└── modula.db                    # SQLite database (dev)
```

### Configuration Keys

```json
{
  "db_driver": "sqlite|mysql|postgres",
  "db_url": "database connection string",
  "port": "HTTP port",
  "ssl_port": "HTTPS port",
  "client_site": "public site domain",
  "admin_site": "admin site domain",
  "bucket_endpoint": "S3 endpoint",
  "bucket_access_key": "S3 access key",
  "bucket_secret_key": "S3 secret key",
  "oauth_client_id": "OAuth client ID",
  "oauth_endpoint": {
    "oauth_auth_url": "OAuth authorization URL",
    "oauth_token_url": "OAuth token URL"
  }
}
```

### TUI Keyboard Shortcuts

```
Navigation:
  ↑/k     - Move up
  ↓/j     - Move down
  →/l     - Expand node / Enter section
  ←/h     - Collapse node / Go back

Actions:
  n       - Create new item
  e       - Edit selected item
  d       - Delete selected item
  f       - Toggle feature (context-dependent)

System:
  q       - Quit
  ?       - Help
  Ctrl+C  - Force quit
```

### Typical Development Timeline

**First time setup:** 5 minutes
**Add new database table:** 10 minutes
**Add new TUI screen:** 30 minutes
**Add complete feature (DB + logic + TUI):** 1-2 hours
**Deploy to production:** 10 minutes

---

**Last Updated:** 2026-01-12
**Document Version:** 1.0
**Status:** Complete
