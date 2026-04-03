# ModulaCMS Documentation

ModulaCMS is a headless CMS that ships as a single Go binary with a REST API, web admin panel, and SSH terminal UI.

## Quick Start

```bash
modula init        # Create a project in the current directory
modula serve       # Start the server
```

Open [http://localhost:8080/admin/](http://localhost:8080/admin/) for the admin panel, or `ssh localhost -p 2233` for the terminal UI.

## The Journey

### Getting Started

Install ModulaCMS and create your first project.

- [Installation](getting-started/installation.md) -- build options, Docker, database backends
- [Your First Project](getting-started/first-project.md) -- init, serve, and connect
- [Configuration](getting-started/configuration.md) -- all `modula.config.json` fields

### Building Content

Define your content model, organize it into trees, and serve it to your frontend.

- [Content Modeling](building-content/content-modeling.md) -- design datatypes and fields
- [Content Trees](building-content/creating-content.md) -- create, move, and reorder hierarchical content
- [Routing](building-content/routing.md) -- map URL slugs to content
- [Media Management](building-content/media.md) -- upload files and serve responsive images

### Custom Admin

Build your own admin experience using the CMS to manage your admin interface.

- [Custom Admin Overview](custom-admin/overview.md)

### Extending

Add custom functionality with Lua plugins.

- [Overview](extending/overview.md) -- what plugins can do
- [Tutorial](extending/tutorial.md) -- build a bookmarks plugin from scratch
- [Testing](extending/testing.md) -- write and run automated plugin tests
- [Lua API Reference](extending/lua-api.md) -- every function and parameter
- [Examples](extending/examples.md) -- complete working plugins

## Concepts

Deep dives into how ModulaCMS models data.

- [Content Model](building-content/content-modeling.md)
- [Tree Structure](building-content/creating-content.md)
- [Publishing Lifecycle](building-content/publishing.md)
- [RBAC](custom-admin/authentication.md)
- [Media Pipeline](building-content/media.md)

## SDKs

Client libraries for consuming the ModulaCMS API. See the [SDK overview](sdks/overview.md) for a comparison.

- [Go SDK](sdks/go/getting-started.md) -- `import modulacms "github.com/hegner123/modulacms/sdks/go"`
- [TypeScript SDK](sdks/typescript/getting-started.md) -- `@modulacms/sdk` and `@modulacms/admin-sdk`
- [Swift SDK](sdks/swift/getting-started.md) -- SPM package for iOS 16+, macOS 13+

## API Reference

- [REST API](api/rest-api.md) -- endpoints, authentication, pagination, output formats

## Deployment

- [Local Development](deployment/local-development.md)
- [Docker](deployment/docker.md)
- [Production](deployment/production.md)

## Reference

- [Glossary](reference/glossary.md)
- [Troubleshooting](reference/troubleshooting.md)
- [Philosophy](PHILOSOPHY.md) -- performance, flexibility, transparency

## Contributing

- [Adding Features](contributing/adding-features.md)
- [Adding Tables](contributing/adding-tables.md)
- [Testing](contributing/testing.md)
