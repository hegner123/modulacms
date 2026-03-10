# ModulaCMS Documentation

ModulaCMS is a headless CMS written in Go that ships as a single binary running three concurrent servers: HTTP (REST API and admin panel), HTTPS (with automatic Let's Encrypt certificates), and SSH (a terminal-based management TUI via Charmbracelet Wish). Content is managed through the SSH TUI, the web admin panel, or the REST API, and delivered to frontend clients over HTTP/HTTPS.

## Getting Started

Set up ModulaCMS from source and configure it for your environment.

- [Installation](getting-started/installation.md) -- build from source, prerequisites, platform requirements
- [Quickstart](getting-started/quickstart.md) -- clone to running in five minutes
- [Configuration](getting-started/configuration.md) -- `modula.config.json` reference, database drivers, S3, OAuth, CORS

## Concepts

Core data model and system design. Read these to understand how ModulaCMS structures content.

- [Content Model](concepts/content-model.md) -- datatypes, fields, content data, content fields, and the admin content system
- [Tree Structure](concepts/tree-structure.md) -- sibling-pointer trees, O(1) operations, tree assembly algorithm
- [Publishing Lifecycle](concepts/publishing-lifecycle.md) -- draft/published states, version snapshots, scheduling, restore
- [RBAC](concepts/rbac.md) -- roles, permissions, the `resource:operation` label format, permission cache
- [Media Pipeline](concepts/media-pipeline.md) -- upload, image optimization, dimension presets, S3 storage, focal point cropping
- [Localization](concepts/localization.md) -- locales, translatable fields, fallback chains, locale-aware delivery

## Guides

Step-by-step instructions for common tasks.

- [Content Modeling](guides/content-modeling.md) -- design datatypes and fields for your content
- [Content Trees](guides/content-trees.md) -- create, move, reorder, and deliver tree-structured content
- [Authentication](guides/authentication.md) -- password login, OAuth providers, sessions, tokens, SSH keys
- [Routing](guides/routing.md) -- map URL slugs to content, route types, output formats
- [Media Management](guides/media-management.md) -- upload files, manage dimensions, focal points, srcset
- [Admin Panel](guides/admin-panel.md) -- HTMX-based web admin interface
- [Managing Plugins](guides/plugins.md) -- install, configure, approve, monitor, and troubleshoot plugins

## Plugin Development

Documentation for building plugins. If you are a CMS administrator, the [Managing Plugins](guides/plugins.md) guide covers installation, configuration, and approval.

- [Overview](plugins/overview.md) -- architecture, VM pool design, lifecycle states, how plugins interact with the CMS
- [Tutorial](plugins/tutorial.md) -- build a bookmarks plugin from scratch, step by step
- [Lua API Reference](plugins/lua-api.md) -- every function, parameter, and return value for db, http, hooks, log, and require
- [Configuration](plugins/configuration.md) -- all modula.config.json fields with tuning guidance and example configurations
- [Security](plugins/security.md) -- sandbox, database isolation, circuit breakers, rate limiting, operation budgets
- [Approval Workflow](plugins/approval.md) -- route and hook approval via CLI, API, and TUI
- [Examples](plugins/examples.md) -- complete example plugins: task tracker, content validator, webhook relay, analytics logger, API gateway

## SDKs

Client libraries for consuming the ModulaCMS API.

- Go SDK -- `import modulacms "github.com/hegner123/modulacms/sdks/go"` ([source](../sdks/go/))
- TypeScript SDK -- `@modulacms/sdk` for content delivery, `@modulacms/admin-sdk` for admin CRUD ([source](../sdks/typescript/))
- Swift SDK -- `ModulaCMS` SPM package for Apple platforms ([source](../sdks/swift/))

## API Reference

- [REST API](api/rest-api.md) -- endpoint reference, authentication, pagination, output formats

## Deployment

- [Local Development](deployment/local-development.md) -- run locally with SQLite
- [Docker](deployment/docker.md) -- containerized stacks for SQLite, MySQL, PostgreSQL
- [Production](deployment/production.md) -- deploy, health checks, rollback, TLS

## Reference

- [Glossary](reference/glossary.md) -- terminology used throughout the documentation
- [Troubleshooting](reference/troubleshooting.md) -- common issues and their solutions

## Contributing

- [Adding Tables](contributing/adding-tables.md) -- add a new database table across all three backends
- [Adding Features](contributing/adding-features.md) -- end-to-end feature development flow
- [Testing](contributing/testing.md) -- testing strategies, integration tests, Docker infrastructure
- [Debugging](contributing/debugging.md) -- logging, TUI debugging, database inspection
