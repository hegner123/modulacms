# GLOSSARY.md

Comprehensive glossary of ModulaCMS-specific terminology and concepts.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/GLOSSARY.md`
**Purpose:** Define project-specific terms, acronyms, and concepts used throughout the ModulaCMS codebase and documentation.
**Last Updated:** 2026-01-12

---

## Table of Contents

- [Core Concepts](#core-concepts)
- [Database & Schema](#database--schema)
- [Content Management](#content-management)
- [Tree Structure](#tree-structure)
- [Authentication & Security](#authentication--security)
- [Architecture & Frameworks](#architecture--frameworks)
- [Tools & Technologies](#tools--technologies)
- [HTTP & Networking](#http--networking)
- [Development](#development)

---

## Core Concepts

### Admin Site
The administrative interface domain for managing ModulaCMS content. Configured separately from the client site to allow distinct domains or subdomains (e.g., `admin.example.com`). Handles all content management operations through the TUI and HTTP API.

**Related:** Client Site, Routes, Admin Routes

---

### Client Site
The public-facing website domain that serves content to end users. Configured in the application settings and can be completely separate from the admin site. Multiple client sites can be managed from a single ModulaCMS instance.

**Related:** Admin Site, Routes, Multi-Site Architecture

---

### Datatype
A content schema definition that specifies the structure and fields for a particular type of content. Similar to WordPress "post types" or content types in other CMSs. Examples include "Page", "Post", "Product", or custom types defined by users.

**Database Table:** `datatypes` (schema 7_datatypes)

**Example:** A "Blog Post" datatype might have fields for Title, Body, Author, Published Date, and Featured Image.

**Related:** Field, Content Data, Datatype-Field Junction

---

### Datatype-Field Junction
The many-to-many relationship between datatypes and fields, implemented through the `datatypes_fields` junction table. Defines which fields belong to which datatypes and their ordering within that datatype.

**Database Table:** `datatypes_fields` (schema 20_datatypes_fields)

**Related:** Datatype, Field

---

### DbDriver
The database abstraction interface that provides a unified API for database operations across SQLite, MySQL, and PostgreSQL. Implements ~150 methods for all database operations, allowing the application to switch database engines through configuration alone.

**Code Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/`

**Implementations:**
- SQLite: `internal/db-sqlite/`
- MySQL: `internal/db-mysql/`
- PostgreSQL: `internal/db-psql/`

**Related:** sqlc, Schema Migrations

---

### Field
A schema property definition that describes a single piece of data within a datatype. Fields have types (text, richtext, image, number, etc.) and metadata. They define what data can be stored but not the actual content values.

**Database Table:** `fields` (schema 8_fields)

**Example:** A "Title" field might be type "text" with max length 255, while a "Body" field might be type "richtext" with no length limit.

**Related:** Datatype, Content Field, Field Types

---

### Field Types
The different data types that fields can store:
- **text**: Short text strings (titles, names)
- **richtext**: HTML content with formatting
- **image**: Image references (media IDs)
- **number**: Numeric values (integers, decimals)
- **boolean**: True/false values
- **date**: Date and time values
- **relation**: References to other content items

**Related:** Field, Content Field

---

### Media Dimensions
Predefined image size presets used for automatic image optimization. When images are uploaded, ModulaCMS generates multiple versions at these dimensions for responsive design and performance optimization.

**Database Table:** `media_dimensions` (schema 3_media_dimension)

**Example:** Large (1200x800), Medium (800x600), Thumbnail (400x300)

**Related:** Media, S3 Storage, Image Optimization

---

### Multi-Site Architecture
ModulaCMS's ability to manage multiple independent websites from a single installation. Each route represents a separate site with its own content tree, and content is isolated between routes.

**Implementation:** Routes serve as content tree roots, with domain routing determining which route handles each request.

**Related:** Routes, Content Tree, Domain Routing

---

### Route
A content tree root representing a separate website or section. Each route has its own independent content hierarchy and can be mapped to a specific domain or subdomain. Routes enable multi-site management from a single ModulaCMS instance.

**Database Table:** `routes` (schema 6_routes)

**Example:** Route 1 might be "example.com" while Route 2 is "blog.example.com", each with completely separate content structures.

**Related:** Admin Route, Content Tree, Multi-Site Architecture

---

### Admin Route
Similar to routes but specifically for the admin panel interface. Admin routes define the structure and navigation of the administrative interface rather than public-facing content.

**Database Table:** `admin_routes` (schema 5_admin_routes)

**Related:** Route, Admin Site

---

## Database & Schema

### Content Data
Individual content items stored in the tree structure. Each content_data record represents one node in the content tree and is an instance of a specific datatype. Contains the tree structure pointers (parent, children, siblings) and references to the datatype it instantiates.

**Database Table:** `content_data` (schema 16_content_data)

**Related:** Content Field, Datatype, Tree Structure, Sibling Pointers

---

### Content Field
The actual content values for fields defined in a datatype. While fields define the schema (what data CAN be stored), content_fields contain the actual data (what data IS stored) for a specific content_data instance.

**Database Table:** `content_fields` (schema 17_content_fields)

**Example:** If content_data #42 is a "Blog Post", its content_fields might include: Title="Hello World", Body="Content here...", Author="John Doe"

**Related:** Content Data, Field, Datatype

---

### Junction Table
A database table that implements a many-to-many relationship between two entities. ModulaCMS uses junction tables to link datatypes with fields, allowing flexible schema definitions where datatypes can have multiple fields and fields can belong to multiple datatypes.

**Examples:**
- `datatypes_fields`: Links datatypes to fields
- `admin_datatypes_fields`: Links admin datatypes to admin fields

**Related:** Datatype, Field, Normalization

---

### Migration
A numbered database schema change tracked through the schema directory structure. Migrations are numbered (1-22) to ensure proper execution order based on foreign key dependencies.

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/`

**Example:** `16_content_data` creates the content_data table after tables it depends on (routes, datatypes) are created in earlier migrations.

**Related:** Schema, DbDriver

---

### Schema
The database structure definition including tables, columns, indexes, and foreign keys. ModulaCMS uses separate schema files for SQLite, MySQL, and PostgreSQL to accommodate database-specific syntax.

**Files:**
- `sql/all_schema.sql` (SQLite)
- `sql/all_schema_mysql.sql` (MySQL)
- `sql/all_schema_psql.sql` (PostgreSQL)

**Related:** Migration, sqlc, DbDriver

---

### Session
An authentication session linking a user to a temporary session identifier stored in a cookie. Sessions have expiration times and are validated on each request to authenticate users.

**Database Table:** `sessions` (schema 15_sessions)

**Related:** Cookie, Authentication, Middleware

---

### sqlc
A SQL compiler that generates type-safe Go code from SQL queries. ModulaCMS uses sqlc to convert `.sql` files with annotations into strongly-typed Go functions, eliminating runtime SQL string construction and reducing errors.

**Configuration:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/sqlc.yml`

**Command:** `just sqlc`

**Annotations:**
- `-- name: GetUser :one` (returns single row)
- `-- name: ListUsers :many` (returns multiple rows)
- `-- name: CreateUser :exec` (executes without returning data)

**Related:** DbDriver, Database Layer

---

## Content Management

### Content Hierarchy
The tree-based organization of content items where each item can have a parent and multiple children, forming a nested structure. Used to represent website navigation, document structures, and organizational taxonomies.

**Implementation:** Sibling-pointer tree in `content_data` table

**Related:** Content Tree, Tree Structure, Sibling Pointers

---

### Content Tree
The complete hierarchical structure of content items for a specific route. Each route has its own independent content tree, with the route serving as the root node.

**Structure:** Sibling-pointer tree with O(1) navigation operations

**Related:** Route, Content Data, TreeRoot, TreeNode

---

### Lazy Loading
A performance optimization where child nodes in the content tree are not loaded until explicitly requested. Prevents loading the entire content tree into memory when only a subset is needed.

**Implementation:** TreeNode's children are loaded on-demand when accessed

**Related:** Content Tree, TreeNode, Performance

---

### NodeIndex
A map data structure that provides O(1) lookup of tree nodes by content_data_id. Maps `content_data.id` → `*TreeNode` for instant access to any node in the tree without traversing.

**Type:** `map[int64]*TreeNode`

**Related:** Tree Structure, TreeNode, Performance

---

### Orphan
A content item in the database that has lost its connection to the content tree due to a missing or invalid parent reference. Orphan resolution is performed during tree loading to detect and handle these disconnected nodes.

**Detection:** During tree loading, nodes with parent_id references that don't exist in NodeIndex

**Related:** Content Tree, Tree Loading, Circular References

---

### TreeNode
An in-memory representation of a content_data record within the content tree. Contains the content data, metadata, and pointers to parent, children, and siblings for tree navigation.

**Structure:**
```go
type TreeNode struct {
    ID          int64
    ParentID    *int64
    FirstChild  *TreeNode
    NextSibling *TreeNode
    PrevSibling *TreeNode
    Children    []*TreeNode
    Data        *ContentData
}
```

**Related:** TreeRoot, Content Tree, NodeIndex

---

### TreeRoot
The top-level container for a content tree, representing a complete site or route. Contains the root node, NodeIndex map, and metadata about the tree.

**Structure:**
```go
type TreeRoot struct {
    Route     *Route
    RootNode  *TreeNode
    NodeIndex map[int64]*TreeNode
    LoadedAt  time.Time
}
```

**Related:** Route, TreeNode, Content Tree

---

## Tree Structure

### Circular Reference
An invalid tree state where a node is its own ancestor, creating a loop in the tree structure. Detected during orphan resolution to prevent infinite loops during tree traversal.

**Example:** Node A → Node B → Node C → Node A (circular)

**Detection:** During orphan resolution, checking if a node already exists in the ancestor chain

**Related:** Orphan, Tree Loading, Tree Validation

---

### First Child
The leftmost child node in a parent's children list. Stored in `content_data.first_child_id` to provide O(1) access to a parent's children without querying.

**Database Field:** `content_data.first_child_id`

**Related:** Sibling Pointers, Parent-Child Relationship, Tree Structure

---

### Next Sibling
The sibling node immediately to the right in the parent's children list. Forms a linked list of siblings for efficient traversal without array storage.

**Database Field:** `content_data.next_sibling_id`

**Related:** Previous Sibling, Sibling Pointers, Sibling Order

---

### Parent-Child Relationship
The hierarchical relationship between a parent node and its children in the content tree. Each child has exactly one parent (or none for root), and parents can have multiple children.

**Database Field:** `content_data.parent_id`

**Related:** First Child, Sibling Pointers, Tree Hierarchy

---

### Previous Sibling
The sibling node immediately to the left in the parent's children list. Forms a doubly-linked list with next_sibling for bidirectional traversal.

**Database Field:** `content_data.prev_sibling_id`

**Related:** Next Sibling, Sibling Pointers, Sibling Order

---

### Sibling Order
The left-to-right ordering of children under a common parent. Maintained through sibling pointers and can be modified without updating all siblings (O(1) reordering).

**Implementation:** Doubly-linked list via next_sibling_id and prev_sibling_id

**Related:** Sibling Pointers, Next Sibling, Previous Sibling

---

### Sibling Pointers
The tree structure technique used by ModulaCMS where each node stores references to its parent, first child, next sibling, and previous sibling. Enables O(1) tree navigation operations without join queries.

**Database Fields:**
- `parent_id`: Parent node reference
- `first_child_id`: First child node reference
- `next_sibling_id`: Next sibling node reference
- `prev_sibling_id`: Previous sibling node reference

**Advantages:**
- O(1) navigation between related nodes
- O(1) insertion/deletion/reordering
- No recursive queries needed
- Efficient tree traversal

**Alternative:** Adjacency List (stores only parent_id, requires recursive queries)

**Related:** Tree Structure, NodeIndex, Content Tree

---

### Three-Phase Loading
The algorithm ModulaCMS uses to load content trees from the database:

**Phase 1:** Create all TreeNode instances and add to NodeIndex
**Phase 2:** Assign hierarchy (connect parent-child and sibling relationships)
**Phase 3:** Resolve orphans (detect circular references and disconnected nodes)

**Purpose:** Separating creation from relationship assignment prevents null pointer issues and enables circular reference detection

**Related:** Tree Loading, NodeIndex, Orphan Resolution

---

## Authentication & Security

### Authentication Context
A Go context value that stores the authenticated user for a request. Injected by middleware after successful session validation, allowing handlers to access the current user.

**Type:** `authcontext` (string type in middleware package)

**Related:** Middleware, Session, Cookie

---

### Cookie
An HTTP cookie used for session-based authentication. ModulaCMS cookies store base64-encoded JSON containing session ID and user ID.

**Format:**
```json
{
  "session": "abc123def456",
  "userId": 42
}
```

**Configuration:** `config.Cookie_Name`, `config.Cookie_Duration`

**Related:** Session, Authentication, Middleware

---

### OAuth
Open Authentication standard for third-party login. ModulaCMS supports OAuth providers like GitHub, Google, Azure AD, and custom OAuth endpoints.

**Configuration Fields:**
- `Oauth_Client_Id`
- `Oauth_Client_Secret`
- `Oauth_Scopes`
- `Oauth_Endpoint`

**Related:** Authentication, User OAuth, Token

---

### Permission
An authorization rule that defines what actions can be performed. Permissions are assigned to roles, and roles are assigned to users, implementing role-based access control (RBAC).

**Database Table:** `permissions` (schema 1_permissions)

**Related:** Role, User, RBAC

---

### Role
A collection of permissions assigned to users. Roles define authorization levels (e.g., Admin, Editor, Viewer) and determine what actions users can perform.

**Database Table:** `roles` (schema 2_roles)

**Related:** Permission, User, RBAC

---

### Session Data
The session identifier string stored in the database and compared against the cookie value for authentication. Must match for a session to be considered valid.

**Database Field:** `sessions.session_data`

**Related:** Session, Cookie, Authentication

---

### Session Expiration
The timestamp when a session becomes invalid. Checked on every request to ensure sessions are still active.

**Database Field:** `sessions.expires_at`

**Related:** Session, Authentication, Cookie Duration

---

### Token
An authentication token used for API access or OAuth integration. Tokens can be session tokens, API tokens, or OAuth access/refresh tokens.

**Database Table:** `tokens` (schema 11_tokens)

**Related:** OAuth, Authentication, API

---

### User OAuth
The association between a ModulaCMS user and their OAuth provider account (GitHub, Google, etc.). Enables third-party authentication while maintaining local user records.

**Database Table:** `user_oauth` (schema 12_user_oauth)

**Related:** OAuth, User, Authentication

---

## Architecture & Frameworks

### Bubbletea
A Charmbracelet framework for building terminal user interfaces based on the Elm Architecture. ModulaCMS uses Bubbletea for its SSH-based TUI.

**Package:** `github.com/charmbracelet/bubbletea`

**Pattern:** Model-Update-View

**Related:** TUI, Elm Architecture, CLI

---

### Elm Architecture
A functional programming pattern for user interfaces with three core concepts:

**Model:** Application state
**Update:** Pure function that handles messages and returns new state
**View:** Pure function that renders state to UI

**Implementation in ModulaCMS:**
```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
}
```

**Related:** Bubbletea, TUI, Message, Command

---

### Command (Elm Architecture)
A side effect to be performed after an update, such as database queries, file I/O, or async operations. Commands return messages to the update loop when they complete.

**Type:** `tea.Cmd`

**Example:** Loading content from database returns a message when complete

**Related:** Elm Architecture, Message, Update Function

---

### Message (Elm Architecture)
An event or data that triggers a state change in the Elm Architecture. Messages are passed to the Update function to produce a new Model and optional Command.

**Type:** `tea.Msg` (interface)

**Examples:**
- User input (keypresses, mouse clicks)
- Completed async operations (database query results)
- Timer events
- Navigation commands

**Related:** Elm Architecture, Update Function, Command

---

### Model (Elm Architecture)
The complete application state at a point in time. In ModulaCMS TUI, the model contains the current screen, selected content, form state, and all UI state.

**Interface:**
```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
}
```

**Related:** Elm Architecture, Update Function, View Function

---

### TUI
Terminal User Interface - a text-based interactive interface running in a terminal emulator. ModulaCMS's TUI runs over SSH for remote content management.

**Framework:** Charmbracelet Bubbletea

**Architecture:** Elm Architecture (Model-Update-View)

**Access:** SSH connection to configured SSH_Host:SSH_Port

**Related:** CLI, Bubbletea, Elm Architecture, SSH

---

### Update Function
A pure function in the Elm Architecture that receives the current model and a message, then returns a new model and optional command. All state changes flow through the update function.

**Signature:** `Update(msg tea.Msg) (tea.Model, tea.Cmd)`

**Related:** Elm Architecture, Model, Message, Command

---

### View Function
A pure function in the Elm Architecture that renders the current model to a string for display in the terminal. Called after every update to reflect state changes.

**Signature:** `View() string`

**Related:** Elm Architecture, Model, Lipgloss

---

### Wish
A Charmbracelet library for building SSH servers. ModulaCMS uses Wish to provide SSH access to the TUI.

**Package:** `github.com/charmbracelet/wish`

**Related:** SSH, TUI, Bubbletea

---

## Tools & Technologies

### autocert
Go package for automatic TLS certificate management with Let's Encrypt. ModulaCMS uses autocert for automatic HTTPS configuration.

**Package:** `golang.org/x/crypto/acme/autocert`

**Related:** HTTPS, TLS, Let's Encrypt

---

### Huh
A Charmbracelet library for building terminal forms. Used in ModulaCMS TUI for data entry and user input.

**Package:** `github.com/charmbracelet/huh`

**Related:** TUI, Forms, Bubbletea

---

### Let's Encrypt
A free certificate authority that provides TLS certificates for HTTPS. ModulaCMS has built-in Let's Encrypt support via autocert.

**Configuration:** `manager := autocert.Manager{...}`

**Related:** HTTPS, TLS, autocert

---

### Lipgloss
A Charmbracelet library for styling terminal output with colors, borders, padding, and alignment. Used throughout ModulaCMS TUI for visual presentation.

**Package:** `github.com/charmbracelet/lipgloss`

**Related:** TUI, View Function, Styling

---

### Lua
A lightweight scripting language used for ModulaCMS plugins. Allows extending functionality without recompiling the binary.

**Package:** `github.com/yuin/gopher-lua`

**Use Cases:**
- Custom output adapters (WordPress JSON compatibility)
- Import scripts for data migration
- Custom business logic

**Related:** Plugin, gopher-lua

---

### gopher-lua
A Lua VM implementation in pure Go. ModulaCMS uses gopher-lua to embed Lua scripting for plugins.

**Package:** `github.com/yuin/gopher-lua`

**Related:** Lua, Plugin, Plugin System

---

### S3
Simple Storage Service - an object storage protocol originally from AWS, now supported by many providers. ModulaCMS uses S3-compatible storage for media files.

**Supported Providers:**
- AWS S3
- Linode Object Storage
- DigitalOcean Spaces
- Backblaze B2
- Any S3-compatible service

**Related:** Media, Bucket, Object Storage

---

## HTTP & Networking

### API Endpoint
An HTTP URL path that provides programmatic access to ModulaCMS functionality. All API endpoints are prefixed with `/api/v1/`.

**Examples:**
- `/api/v1/contentdata` - Content management
- `/api/v1/media` - Media operations
- `/api/v1/users` - User management

**Protection:** All `/api/*` routes require authentication

**Related:** REST, HTTP, Router

---

### CORS
Cross-Origin Resource Sharing - HTTP headers that allow browsers to make requests from different domains. ModulaCMS CORS is configurable per deployment.

**Headers:**
- `Access-Control-Allow-Origin`
- `Access-Control-Allow-Methods`
- `Access-Control-Allow-Headers`
- `Access-Control-Allow-Credentials`

**Configuration:** `config.Cors_Origins`, `config.Cors_Methods`, `config.Cors_Headers`

**Related:** HTTP, Middleware, Preflight

---

### HTTP Server
The server that handles standard HTTP requests on the configured port. Serves as fallback if HTTPS fails.

**Configuration:** `config.Port`

**Default:** `:80` or `:8080`

**Related:** HTTPS Server, Server Trio

---

### HTTPS Server
The server that handles encrypted HTTPS requests with TLS. Automatically configured with Let's Encrypt certificates.

**Configuration:** `config.SSL_Port`

**Default:** `:443`

**Related:** HTTP Server, Let's Encrypt, TLS

---

### Middleware
HTTP request pre-processing functions that run before reaching route handlers. ModulaCMS uses middleware for CORS, authentication, and session validation.

**Implementation:** Single `Serve()` function that wraps the router

**Operations:**
1. CORS header setting
2. Authentication (cookie validation)
3. API route protection
4. Context injection

**Related:** CORS, Authentication, HTTP

---

### Preflight Request
An HTTP OPTIONS request sent by browsers before actual requests to check CORS permissions. ModulaCMS middleware responds with 204 No Content and appropriate CORS headers.

**Method:** `OPTIONS`

**Related:** CORS, Middleware, HTTP

---

### Router
The HTTP request router that maps URL paths to handler functions. ModulaCMS uses Go's standard `http.ServeMux`.

**Implementation:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/router/mux.go`

**Pattern:** Standard library routing (no third-party frameworks)

**Related:** HTTP, API Endpoint, Middleware

---

### SSH
Secure Shell protocol used for remote terminal access. ModulaCMS runs an SSH server that provides access to the TUI.

**Configuration:** `config.SSH_Host`, `config.SSH_Port`

**Framework:** Charmbracelet Wish

**Related:** TUI, Wish, Server Trio

---

### SSH Server
The server that handles SSH connections and provides TUI access. Runs alongside HTTP/HTTPS servers.

**Configuration:** `config.SSH_Port`

**Default:** `:2222`

**Related:** SSH, TUI, Server Trio

---

### Server Trio
The three servers that run simultaneously in ModulaCMS:
1. **HTTP Server** - Standard web traffic (fallback)
2. **HTTPS Server** - Encrypted web traffic (primary)
3. **SSH Server** - Terminal access for TUI

**Implementation:** Three goroutines started from `main.go`

**Related:** HTTP Server, HTTPS Server, SSH Server

---

## Development

### CGO
C code interoperability in Go programs. Required by ModulaCMS for SQLite driver (github.com/mattn/go-sqlite3).

**Impact:** Increases build time, requires C compiler

**Related:** SQLite, Build

---

### Makefile
Build automation file defining common development tasks.

**Commands:**
- `just build` - Build production binaries
- `just dev` - Build local development binary
- `just test` - Run all tests
- `just sqlc` - Generate sqlc code
- `just lint` - Run linters

**Related:** Build, Development, Testing

---

### Test Database
A temporary SQLite database created during test runs. Located in `testdb/` directory and cleaned before/after tests.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/testdb/`

**Lifecycle:** Created for tests, deleted after completion

**Related:** Testing, SQLite

---

### Vendor
A Go module pattern where dependencies are copied into the project repository. Useful for reproducible builds and offline development.

**Command:** `just vendor`

**Related:** Go Modules, Dependencies

---

## Quick Reference

### Acronyms

- **API**: Application Programming Interface
- **CGO**: C-Go (C interoperability)
- **CLI**: Command Line Interface
- **CMS**: Content Management System
- **CORS**: Cross-Origin Resource Sharing
- **CRUD**: Create, Read, Update, Delete
- **CSS**: Cascading Style Sheets
- **CSV**: Comma-Separated Values
- **HTTP**: Hypertext Transfer Protocol
- **HTTPS**: HTTP Secure
- **JSON**: JavaScript Object Notation
- **JWT**: JSON Web Token
- **OAuth**: Open Authentication
- **RBAC**: Role-Based Access Control
- **REST**: Representational State Transfer
- **S3**: Simple Storage Service
- **SQL**: Structured Query Language
- **SSH**: Secure Shell
- **TLS**: Transport Layer Security
- **TUI**: Terminal User Interface
- **UI**: User Interface
- **URL**: Uniform Resource Locator
- **UUID**: Universally Unique Identifier
- **VM**: Virtual Machine

---

### Key File Paths

```
/Users/home/Documents/Code/Go_dev/modulacms/
├── cmd/main.go                    # Application entry point
├── internal/
│   ├── cli/                       # TUI implementation
│   ├── db/                        # Database interface
│   ├── middleware/                # HTTP middleware
│   ├── model/                     # Business logic
│   └── router/                    # HTTP router
├── sql/
│   ├── schema/                    # Database migrations
│   ├── mysql/                     # MySQL queries
│   └── postgres/                  # PostgreSQL queries
└── ai/
    ├── START.md                   # Documentation index
    ├── architecture/              # Architecture docs
    ├── workflows/                 # How-to guides
    ├── packages/                  # Package docs
    ├── domain/                    # Domain docs
    └── reference/                 # Reference docs (this file)
```

---

### Database Tables (Alphabetical)

- `admin_content_data` - Admin content instances
- `admin_content_fields` - Admin content values
- `admin_datatypes` - Admin content schemas
- `admin_datatypes_fields` - Admin datatype-field junction
- `admin_fields` - Admin field definitions
- `admin_routes` - Admin panel routes
- `content_data` - Content tree nodes
- `content_fields` - Content field values
- `datatypes` - Content type definitions
- `datatypes_fields` - Datatype-field junction
- `fields` - Field schema definitions
- `media` - Media file tracking
- `media_dimensions` - Image size presets
- `permissions` - Permission definitions
- `roles` - Role definitions
- `routes` - Site/route definitions
- `sessions` - User sessions
- `tables` - Table metadata
- `tokens` - Authentication tokens
- `user_oauth` - OAuth provider links
- `users` - User accounts

---

**Last Updated:** 2026-01-12
**Maintained By:** ModulaCMS Core Team
**Total Terms:** 80+
