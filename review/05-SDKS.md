# SDK Review (TypeScript, Go, Swift)

## The SDK Strategy

Three SDKs targeting the three major platform ecosystems:
- **TypeScript**: Web frontends and Node.js backends
- **Go**: Backend services and CLIs
- **Swift**: iOS, macOS, tvOS, watchOS, visionOS

All three share the same patterns: branded ID types, generic CRUD resources, consistent error helpers, pagination support, and specialized resources for auth, media upload, admin tree, plugins, etc. This cross-language consistency is rare and valuable.

## TypeScript SDK

### @modulacms/types (Shared Foundation)
Branded ID types using `Brand<T, B>` pattern. Zero runtime cost, full compile-time safety. 26 ID types + 3 value types. Entity types, enums, and pagination primitives. This package is the foundation both SDKs build on.

**Verdict:** Excellent. Clean type design, zero dependencies, well-documented.

### @modulacms/sdk (Content Delivery)
Read-only SDK with a hero method: `getPage(slug, options?)`. Supports format negotiation (contentful, sanity, strapi, wordpress, clean, raw) and optional runtime validators for type narrowing. ~400 lines of focused code.

**What's good:** Single responsibility, zero dependencies, optional validation pattern is elegant.
**What's missing:** Pagination only for media (not content or datatypes). No filtering or sorting.

### @modulacms/admin-sdk (Admin CRUD)
Full admin SDK with 20+ resources. Factory function pattern: `createAdminClient(config)`. Generic CRUD resource with list, get, create, update, remove, listPaginated, count. Specialized resources for auth, media upload, admin tree, plugins, config, import.

**What's good:** Comprehensive coverage, proper error discrimination (`isDuplicateMedia`, `isFileTooLarge`), clean HTTP layer with signal merging, zero dependencies beyond @modulacms/types.

**What's impressive:** Real HTTP servers in tests (not just mocking). 13 test files with 150+ tests.

**Build quality:** ESM + CJS dual builds via tsup, strict TypeScript, source maps, tree-shaking enabled.

### TypeScript SDK Verdict: A

## Go SDK

~3,600 lines (source + tests). Generic `Resource[Entity, CreateParams, UpdateParams, ID ~string]` pattern using Go 1.18+ generics. Branded ID types as distinct string types. Context propagation on all I/O. Private httpClient with proper body draining.

**What's good:**
- Generic resource eliminates repetition across 14+ entity types
- All operations accept `context.Context`
- Error helpers use `errors.As()` idiomatically
- Optional HTTP client injection
- Zero external dependencies (stdlib only)

**What's missing:**
- No query filtering or sorting (only `RawList()` with manual url.Values)
- AdminTree and ContentDelivery return `json.RawMessage` (forces manual unmarshal)
- Test coverage gaps: Auth, AdminTree, ContentDelivery, Import, Plugins untested
- No per-request timeout override

### Go SDK Verdict: A-

## Swift SDK

~2,500 lines source + ~2,100 lines tests. Async/await throughout. Sendable compliance on all public types. `ResourceID` protocol with default implementations for 30+ branded ID types. Custom URLSession configuration (30s timeout, cookies disabled).

**What's good:**
- Modern Swift concurrency (async/await, Sendable)
- ResourceID protocol eliminates boilerplate for branded IDs
- ExpressibleByStringLiteral for convenient ID construction
- 8 test files with MockURLProtocol infrastructure
- Zero external dependencies (Foundation only)

**What's missing:**
- ContentDelivery and AdminTree return `Data` (raw bytes)
- No per-request timeout or cancellation override
- Auth resource untested
- Same filtering/sorting gap as Go SDK

### Swift SDK Verdict: A-

## Cross-SDK Consistency

| Feature | TypeScript | Go | Swift |
|---------|-----------|------|-------|
| Branded IDs | Brand<T,B> | Type alias | ResourceID protocol |
| Generic CRUD | Per-resource | Resource[E,C,U,ID] | Resource<E,C,U,ID> |
| Error helpers | isNotFound() | IsNotFound() | isNotFound() |
| Auth methods | 5 (login/logout/me/register/reset) | 5 | 5 |
| Media upload | multipart | multipart | multipart |
| Pagination | listPaginated() | ListPaginated() | listPaginated() |
| Plugin support | Full | Full | Full |
| Dependencies | 0 external | 0 external | 0 external |
| Test coverage | 16 files, 200+ tests | 5 files, ~100 tests | 8 files, ~80 tests |

The consistency is remarkable. Method names map 1:1 across languages (adjusted for naming conventions). Entity types match. Error handling patterns match. This makes documentation and migration between platforms straightforward.

## What Is Extra

The TypeScript delivery SDK (`@modulacms/sdk`) is very minimal - it could be just the `getPage()` method plus some list endpoints. The admin SDK covers everything else. This separation is clean and intentional, but the delivery SDK's limited pagination support means frontend developers may need the admin SDK for anything beyond basic page fetching.

## Recommendations

1. **Add query filtering/sorting** to all three SDKs (builder pattern or parameter objects)
2. **Type the tree/content responses** instead of returning raw JSON/Data
3. **Complete test coverage** for specialized resources (Auth, AdminTree, Plugins)
4. **Add per-request timeout** override in Go and Swift SDKs
5. **Consider a shared specification** (OpenAPI or similar) to keep SDKs in sync automatically
