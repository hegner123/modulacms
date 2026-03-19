# SDK Overview

ModulaCMS provides official SDKs for Go, TypeScript, and Swift that share the same design patterns and cover the same API surface.

## SDK Comparison

| Feature | Go | TypeScript | Swift |
|---------|:--:|:----------:|:-----:|
| Content delivery (read-only) | Yes | Yes | Yes |
| Admin CRUD (full read/write) | Yes | Yes | Yes |
| Branded ID types | Yes | Yes | Yes |
| Generic CRUD resources | Yes | Yes | Yes |
| Typed error handling | Yes | Yes | Yes |
| Pagination support | Yes | Yes | Yes |
| Media upload | Yes | Yes | Yes |
| Content tree operations | Yes | Yes | Yes |
| Publishing / versioning | Yes | Yes | Yes |
| Plugin management | Yes | Yes | Yes |
| Deploy operations | Yes | Yes | Yes |
| Webhook management | Yes | Yes | Yes |
| Locale / i18n | Yes | Yes | Yes |
| Media folder management | Yes | Yes | Yes |

| Property | Go | TypeScript | Swift |
|----------|:--:|:----------:|:-----:|
| Install method | `go get` | `npm install` / `pnpm add` | Swift Package Manager |
| Min runtime | Go 1.25+ | Node 22+ / any modern browser | iOS 16+, macOS 13+, tvOS 16+, watchOS 9+ |
| Dependencies | Zero | Zero | Zero |
| Build output | Compiled binary | ESM + CJS dual build | Framework |
| Package name | `modulacms` | `@modulacms/types`, `@modulacms/tree`, `@modulacms/sdk`, `@modulacms/admin-sdk`, `@modulacms/plugin-sdk`, `@modulacms/admin-ui` | `ModulaCMS` |

## When to Use Each SDK

**Go** -- Server-side applications, CLI tools, backend services, data pipelines, and any Go codebase that needs to read or write CMS content. The Go SDK is a single package with zero dependencies, making it straightforward to vendor or embed in existing projects.

**TypeScript** -- Web applications, server-side rendering (Next.js, Nuxt, SvelteKit), and frontend SPAs. The TypeScript SDK ships as six packages: `@modulacms/types` for shared entity types and branded IDs, `@modulacms/tree` for content tree utilities, `@modulacms/sdk` for read-only content delivery, `@modulacms/admin-sdk` for full admin CRUD, `@modulacms/plugin-sdk` for building plugin UIs with Web Components, and `@modulacms/admin-ui` for admin panel TypeScript (block editor state). The content SDKs share types from `@modulacms/types`; the plugin SDK has zero dependencies.

**Swift** -- Native Apple platform applications (iOS, macOS, tvOS, watchOS). The Swift SDK uses URLSession with async/await and requires no third-party dependencies.

## Shared Design Patterns

All three SDKs follow these conventions:

**Branded ID types.** Entity IDs are distinct types (e.g., `ContentID`, `UserID`, `DatatypeID`), not raw strings. The compiler catches mistakes like passing a `UserID` where a `ContentID` belongs.

**Generic CRUD resources.** Most API resources expose the same set of methods -- List, Get, Create, Update, Delete, ListPaginated, Count -- through a generic resource type parameterized by entity, create params, update params, and ID type.

**Consistent error types.** API errors carry the HTTP status code, a server message, and the raw response body. Helper functions classify common error conditions: not found, unauthorized, duplicate media, invalid media path.

**Single client instance.** Each SDK creates one client from a configuration struct (base URL, API key, optional HTTP client). All resources hang off that client instance. The client is safe for concurrent use.

## Getting Started

- [Go SDK](go/getting-started.md)
- [TypeScript SDK](typescript/getting-started.md)
- [Swift SDK](swift/getting-started.md)
