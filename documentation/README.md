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

- [Installation](/docs/getting-started/installation) -- build options, Docker, database backends
- [Your First Project](/docs/getting-started/first-project) -- init, serve, and connect
- [Configuration](/docs/getting-started/configuration) -- all `modula.config.json` fields

### Building Content

Define your content model, organize it into trees, and serve it to your frontend.

- [Content Modeling](/docs/building-content/content-modeling) -- design datatypes and fields
- [Content Trees](/docs/building-content/creating-content) -- create, move, and reorder hierarchical content
- [Routing](/docs/building-content/routing) -- map URL slugs to content
- [Media Management](/docs/building-content/media) -- upload files and serve responsive images

### Custom Admin

Build your own admin experience using the CMS to manage your admin interface.

- [Custom Admin Overview](/docs/custom-admin/overview)

### Extending

Add custom functionality with Lua plugins.

- [Overview](/docs/extending/overview) -- what plugins can do
- [Tutorial](/docs/extending/tutorial) -- build a bookmarks plugin from scratch
- [Testing](/docs/extending/testing) -- write and run automated plugin tests
- [Lua API Reference](/docs/extending/lua-api) -- every function and parameter
- [Examples](/docs/extending/examples) -- complete working plugins

## Concepts

Deep dives into how ModulaCMS models data.

- [Content Model](/docs/building-content/content-modeling)
- [Tree Structure](/docs/building-content/creating-content)
- [Publishing Lifecycle](/docs/building-content/publishing)
- [RBAC](/docs/custom-admin/authentication)
- [Media Pipeline](/docs/building-content/media)

## SDKs

Client libraries for consuming the ModulaCMS API. See the [SDK overview](/docs/sdks/overview) for a comparison.

- [Go SDK](/docs/sdks/go/getting-started) -- `import modulacms "github.com/hegner123/modulacms/sdks/go"`
- [TypeScript SDK](/docs/sdks/typescript/getting-started) -- `@modulacms/sdk` and `@modulacms/admin-sdk`
- [Swift SDK](/docs/sdks/swift/getting-started) -- SPM package for iOS 16+, macOS 13+

## API Reference

- [REST API](/docs/api/rest-api) -- endpoints, authentication, pagination, output formats

## Deployment

- [Local Development](/docs/deployment/local-development)
- [Docker](/docs/deployment/docker)
- [Production](/docs/deployment/production)

## Reference

- [Glossary](/docs/reference/glossary)
- [Troubleshooting](/docs/reference/troubleshooting)
- [Philosophy](/docs/PHILOSOPHY) -- performance, flexibility, transparency

## Contributing

- [Adding Features](/docs/contributing/adding-features)
- [Adding Tables](/docs/contributing/adding-tables)
- [Testing](/docs/contributing/testing)
