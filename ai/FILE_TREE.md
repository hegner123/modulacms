# FILE_TREE.md

Complete directory structure of ModulaCMS with full absolute paths and detailed descriptions for AI agent consumption.

## Project Root
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms`
**Description:** Root directory of ModulaCMS project. Contains Go modules, Makefile, configuration files, and all source code.

---

## Top-Level Directories

### AI & Automation
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai`
**Description:** AI-related files, workflows, and automation scripts for the project.

### Binary Output
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/bin`
**Description:** Directory for compiled binary executables. Generated during build process.

### Main Application
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd`
**Description:** Main application entry point. Contains `main.go` which starts HTTP/HTTPS/SSH servers and handles command-line flags.
**Key Files:** `main.go`, `main_test.go`

### Documentation
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/docs`
**Description:** Project documentation including CLI guides, plugin documentation, and page structure documentation.
**Key Files:** `cli.md`, `plugins.md`, `page.md`, `table_list.md`

### Test Database
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/testdb`
**Description:** Temporary directory for test database files. Cleaned before and after test runs. Contains `.db` files created during testing.

---

## Internal Package Structure

### Internal Root
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal`
**Description:** All internal Go packages that are not intended for external import. Contains core application logic organized by domain.

---

## Authentication & Security

### Authentication
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/auth`
**Description:** OAuth authentication implementation. Handles third-party authentication providers (GitHub, Google, Azure AD, custom OAuth providers).
**Purpose:** User authentication, OAuth flows, token management.

### Middleware
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/middleware`
**Description:** HTTP middleware implementations for request processing.
**Purpose:** CORS handling, cookie management, session management, request/response interception.
**Key Files:** `cors.go`, `cookies.go`, `session.go`, `middleware.go`

---

## Database Layer

### Database Abstraction
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db`
**Description:** Database abstraction layer defining the DbDriver interface. Provides common interface for all database operations regardless of underlying database engine.
**Purpose:** Database interface definitions, query abstractions, common database utilities.
**Key Files:** Interface definitions for 150+ database methods.

### Legacy Database Code
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/old`
**Description:** Legacy database implementation code. Kept for reference during refactoring.
**Purpose:** Historical reference, migration assistance.

### Database SQL Definitions
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/sql`
**Description:** SQL query definitions used by the database abstraction layer.
**Purpose:** Shared SQL query definitions.

### MySQL SQL Queries
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/sql/mysql`
**Description:** MySQL-specific SQL query implementations.
**Purpose:** MySQL dialect queries, MySQL-specific operations.

### PostgreSQL SQL Queries
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/sql/psql`
**Description:** PostgreSQL-specific SQL query implementations.
**Purpose:** PostgreSQL dialect queries, PostgreSQL-specific operations.

### SQLite SQL Queries
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/sql/sqlite`
**Description:** SQLite-specific SQL query implementations.
**Purpose:** SQLite dialect queries, SQLite-specific operations.

### SQLite Driver Implementation
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-sqlite`
**Description:** SQLite database driver implementation. Implements DbDriver interface for SQLite databases.
**Purpose:** SQLite connection management, query execution, transaction handling.
**Database:** Best for development, small deployments, single-server setups.

### MySQL Driver Implementation
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-mysql`
**Description:** MySQL database driver implementation. Implements DbDriver interface for MySQL databases.
**Purpose:** MySQL connection management, query execution, transaction handling.
**Database:** Production-ready, suitable for medium to large deployments.

### PostgreSQL Driver Implementation
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-psql`
**Description:** PostgreSQL database driver implementation. Implements DbDriver interface for PostgreSQL databases.
**Purpose:** PostgreSQL connection management, query execution, transaction handling.
**Database:** Production-ready, suitable for enterprise deployments with advanced features.

---

## CLI & Terminal User Interface

### CLI Application
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli`
**Description:** Terminal User Interface (TUI) application logic using Charmbracelet Bubbletea framework. Implements Elm Architecture (Model-Update-View) for SSH-based content management.
**Purpose:** TUI screens, interactive forms, tree navigation, content editing, SSH interface.
**Architecture:** Elm Architecture with messages, models, update functions, view rendering.

### CLI Assets
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/assets`
**Description:** Static assets used by the CLI application including ASCII art, templates, and resource files.
**Purpose:** CLI visual assets, templates, static resources.

### CLI Title Screens
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/titles`
**Description:** Title screens and splash screens displayed in the TUI.
**Purpose:** Welcome screens, branding, navigation headers.

---

## Core Features

### Backup System
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/backup`
**Description:** Backup and restore functionality. Creates zip archives containing SQL dumps and media files from S3 storage.
**Purpose:** Database backup, media backup, restore operations, environment syncing.
**Format:** Timestamped zip files with SQL dumps and media.

### Object Storage
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/bucket`
**Description:** S3-compatible object storage integration for media files. Supports AWS S3, Linode Object Storage, DigitalOcean Spaces, Backblaze B2, and any S3-compatible API.
**Purpose:** File uploads, media storage, CDN integration, backup storage.
**Configuration:** Bucket URL, endpoint, access keys, region settings.

### Bucket Test Files
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/bucket/testfiles`
**Description:** Test files for S3 bucket operations including sample images and upload test data.
**Purpose:** Unit testing, integration testing, bucket connectivity testing.

### Configuration Management
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/config`
**Description:** Application configuration management. Handles loading and parsing of configuration from JSON files and environment variables.
**Purpose:** Config loading, environment detection, settings validation.
**Config Items:** Database credentials, OAuth settings, S3 buckets, CORS origins, SSL ports, domain names.

### Deployment Utilities
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/deploy`
**Description:** Deployment utilities and helpers for production deployment.
**Purpose:** Deployment automation, server setup, production helpers.

### File Operations
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/file`
**Description:** File system operations including saving files and extracting zip archives.
**Purpose:** File I/O, zip extraction, file manipulation.
**Key Files:** `save.go`, `unzip.go`

### Command-Line Flags
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/flags`
**Description:** Command-line argument parsing and flag definitions.
**Purpose:** CLI argument handling, flag definitions for --cli, --install, --version, --config, etc.
**Key Files:** `flags.go`

### Installation Wizard
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/install`
**Description:** Interactive installation wizard for first-time setup. Creates initial database, configures settings, sets up admin user.
**Purpose:** First-run setup, database initialization, initial configuration.
**Key Files:** `install_main.go`, `install_form.go`, `install_create.go`, `install_checks.go`

### Media Processing
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/media`
**Description:** Media upload, image optimization, and dimension preset management. Generates multiple image sizes from uploads (like WordPress).
**Purpose:** Image cropping, scaling, format conversion (PNG, JPEG, GIF, WebP), dimension presets.
**Features:** Center cropping, automatic scaling, S3 upload integration.

### Data Models
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model`
**Description:** Data models and business logic. Contains struct definitions for content trees, nodes, datatypes, fields, and core data structures.
**Purpose:** Business logic, data structures, type definitions, tree operations.
**Key Concepts:** TreeRoot, TreeNode, NodeIndex (O(1) lookups), sibling pointers.

### Update System
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/update`
**Description:** Application update system for checking and applying updates.
**Purpose:** Version checking, update downloads, binary replacement.
**Key Files:** `update.go`

### Utilities
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility`
**Description:** Shared utility functions and logging infrastructure.
**Purpose:** Logger configuration, helper functions, common utilities, error handling.
**Key Features:** DefaultLogger for structured logging throughout application.

---

## Plugin System

### Plugin Core
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/plugin`
**Description:** Lua plugin system using gopher-lua for embedded scripting. Allows extending CMS without recompiling binary.
**Purpose:** Plugin loading, Lua VM management, plugin API, output adapters, import adapters.
**Use Cases:** Custom output formats (WordPress JSON compatibility), database migrations, custom business logic.

### Blog Plugin
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/plugin/blog`
**Description:** Example blog plugin implementation demonstrating plugin architecture.
**Purpose:** Reference implementation, blog functionality, plugin pattern example.

### Blog Plugin Assets
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/plugin/blog/assets`
**Description:** Static assets for the blog plugin including templates and resources.
**Purpose:** Blog templates, CSS, images, static resources.

### Blog Plugin JavaScript
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/plugin/blog/assets/js`
**Description:** JavaScript files for blog plugin frontend functionality.
**Purpose:** Client-side blog functionality, interactive features.

---

## SQL Schema & Queries

### SQL Root
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql`
**Description:** Root directory for all SQL queries and database schema definitions. Contains sqlc configuration and generated code output.
**Purpose:** SQL query organization, schema definitions, sqlc configuration.
**Key Files:** `sqlc.yml`, `all_schema.sql`, `all_schema_mysql.sql`, `all_schema_psql.sql`, `docker-compose.yml`, `permissions.json`

### MySQL Queries
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/mysql`
**Description:** MySQL-specific query files used by sqlc to generate type-safe Go code.
**Purpose:** MySQL query definitions with sqlc annotations.
**Format:** SQL files with `-- name: QueryName :many` annotations.

### PostgreSQL Queries
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/postgres`
**Description:** PostgreSQL-specific query files used by sqlc to generate type-safe Go code.
**Purpose:** PostgreSQL query definitions with sqlc annotations.
**Format:** SQL files with `-- name: QueryName :many` annotations.

### Schema Migrations Root
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema`
**Description:** Database schema organized in numbered directories for migration ordering. Each directory contains CREATE TABLE statements and related schema definitions.
**Purpose:** Schema migrations, table definitions, database structure versioning.
**Migration Order:** Numbered 1-22 for dependency ordering.
**Key Files:** `instructions.md`, `list.md`

---

## Schema Migration Directories (Execution Order)

### 1. Permissions Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/1_permissions`
**Description:** Permissions table schema. Defines permission types and authorization rules.
**Purpose:** Permission definitions, role-based access control foundation.
**Creates:** `permissions` table.

### 2. Roles Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/2_roles`
**Description:** Roles table schema. Defines user roles and their associated permissions.
**Purpose:** Role definitions, permission assignments, user authorization levels.
**Creates:** `roles` table.
**Dependencies:** Requires `permissions` table.

### 3. Media Dimensions Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/3_media_dimension`
**Description:** Media dimension presets for image optimization. Defines target dimensions for automatic image resizing.
**Purpose:** Image size presets (e.g., 1200x800, 800x600, 400x300), responsive image generation.
**Creates:** `media_dimensions` table.

### 4. Users Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/4_users`
**Description:** Users table schema. Core user account information.
**Purpose:** User accounts, authentication credentials, profile data.
**Creates:** `users` table.
**Dependencies:** Requires `roles` table.

### 5. Admin Routes Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/5_admin_routes`
**Description:** Admin panel routes configuration. Defines admin interface routing.
**Purpose:** Admin URL structure, admin panel navigation, admin-specific routes.
**Creates:** `admin_routes` table.

### 6. Routes Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/6_routes`
**Description:** Public routes/sites configuration. Each route represents a separate content tree (like a website or section).
**Purpose:** Multi-site support, content tree roots, URL routing configuration.
**Creates:** `routes` table.
**Concept:** Routes are top-level containers for content hierarchies.

### 7. Datatypes Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/7_datatypes`
**Description:** Datatypes table schema. Defines content types (Page, Post, Product, etc.) and their structure.
**Purpose:** Content type definitions, schema for different content types, field organization.
**Creates:** `datatypes` table.
**Concept:** Datatypes define the schema/structure of content (like WordPress post types).

### 8. Fields Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/8_fields`
**Description:** Fields table schema. Defines individual fields that belong to datatypes (Title, Body, Image, etc.).
**Purpose:** Field definitions, field types (text, richtext, image, number), field metadata.
**Creates:** `fields` table.
**Dependencies:** Requires `datatypes` table.

### 9. Admin Datatypes Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/9_admin_datatypes`
**Description:** Admin-specific datatypes configuration. Datatypes used in admin panel.
**Purpose:** Admin content types, admin interface structure.
**Creates:** `admin_datatypes` table.

### 10. Admin Fields Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/10_admin_fields`
**Description:** Admin-specific fields configuration. Fields used in admin panel forms.
**Purpose:** Admin field definitions, admin interface customization.
**Creates:** `admin_fields` table.
**Dependencies:** Requires `admin_datatypes` table.

### 11. Tokens Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/11_tokens`
**Description:** Authentication tokens table. Stores API tokens, session tokens, OAuth tokens.
**Purpose:** Token management, API authentication, session handling.
**Creates:** `tokens` table.
**Dependencies:** Requires `users` table.

### 12. User OAuth Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/12_user_oauth`
**Description:** User OAuth provider associations. Links users to OAuth providers (GitHub, Google, Azure AD, etc.).
**Purpose:** OAuth integration, third-party authentication, provider linking.
**Creates:** `user_oauth` table.
**Dependencies:** Requires `users` table.

### 13. Tables Metadata Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/13_tables`
**Description:** Tables metadata tracking. Stores information about database tables and their structure.
**Purpose:** Schema versioning, table metadata, migration tracking.
**Creates:** `tables` table.

### 14. Media Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/14_media`
**Description:** Media files table. Tracks uploaded media in S3 storage with metadata.
**Purpose:** Media tracking, S3 key storage, image metadata (dimensions, file size, mime type).
**Creates:** `media` table.
**Fields:** S3 bucket, key, filename, dimensions, mime type, file size.

### 15. Sessions Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/15_sessions`
**Description:** User sessions table. Manages active user sessions and session state.
**Purpose:** Session management, user authentication state, session expiration.
**Creates:** `sessions` table.
**Dependencies:** Requires `users` table.

### 16. Content Data Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/16_content_data`
**Description:** Content data table. Stores individual content items in the tree structure using sibling pointers.
**Purpose:** Content storage, tree hierarchy (parent_id, first_child_id, next_sibling_id, prev_sibling_id).
**Creates:** `content_data` table.
**Structure:** Sibling-pointer tree for O(1) operations.
**Dependencies:** Requires `datatypes`, `routes` tables.

### 17. Content Fields Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/17_content_fields`
**Description:** Content field values. Stores actual content values for fields defined in datatypes.
**Purpose:** Field value storage, content data population.
**Creates:** `content_fields` table.
**Dependencies:** Requires `content_data`, `fields` tables.

### 18. Admin Content Data Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/18_admin_content_data`
**Description:** Admin-specific content data. Content items for admin panel.
**Purpose:** Admin content storage, admin panel data.
**Creates:** `admin_content_data` table.
**Dependencies:** Requires `admin_datatypes`, `admin_routes` tables.

### 19. Admin Content Fields Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/19_admin_content_fields`
**Description:** Admin-specific content field values.
**Purpose:** Admin field value storage, admin content population.
**Creates:** `admin_content_fields` table.
**Dependencies:** Requires `admin_content_data`, `admin_fields` tables.

### 20. Datatypes Fields Junction Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/20_datatypes_fields`
**Description:** Junction table linking datatypes to their fields. Defines which fields belong to which datatypes.
**Purpose:** Many-to-many relationship between datatypes and fields, field ordering.
**Creates:** `datatypes_fields` table.
**Dependencies:** Requires `datatypes`, `fields` tables.

### 21. Admin Datatypes Fields Junction Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/21_admin_datatypes_fields`
**Description:** Junction table linking admin datatypes to admin fields.
**Purpose:** Many-to-many relationship for admin types and fields, admin field ordering.
**Creates:** `admin_datatypes_fields` table.
**Dependencies:** Requires `admin_datatypes`, `admin_fields` tables.

### 22. Joins Schema
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/22_joins`
**Description:** Complex join definitions and view creation for common query patterns.
**Purpose:** Optimized query patterns, materialized views, complex joins for content trees.
**Creates:** Views and indexes for performance optimization.

### Schema Utilities
**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/utility`
**Description:** Utility scripts for schema management, migration helpers, and database maintenance.
**Purpose:** Schema tools, migration utilities, database administration scripts.

---

## Directory Summary

**Total Directories:** 63
**Core Packages:** 17 internal packages
**Database Drivers:** 3 implementations (SQLite, MySQL, PostgreSQL)
**Schema Migrations:** 22 ordered migration directories
**Plugin System:** 1 core + 1 example plugin

## Key Architectural Notes

### Tree Structure
Content hierarchy uses sibling pointers (parent_id, first_child_id, next_sibling_id, prev_sibling_id) stored in `content_data` table. This enables O(1) operations via NodeIndex map.

### Database Abstraction
DbDriver interface (defined in `internal/db/`) provides 150+ methods implemented by three database drivers. Switch between SQLite/MySQL/PostgreSQL via configuration only.

### Schema Loading
Schema migrations are embedded in binary using `//go:embed` directives. Migrations run in numbered order (1-22) on application startup.

### sqlc Integration
SQL queries in `sql/mysql/`, `sql/postgres/` directories are processed by sqlc to generate type-safe Go functions. Run `make sqlc` to regenerate.

### TUI Architecture
CLI uses Elm Architecture (Model-Update-View) via Charmbracelet Bubbletea. All state lives in Model, changes via Update function, rendering via View function.
