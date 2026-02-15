# ModulaCMS TypeScript SDK — Implementation Plan

## Context

ModulaCMS is a headless CMS with a Go backend. Client sites (SPAs, SSR apps, static generators) need a typed TypeScript SDK to fetch content from the API. The SDK covers **content delivery only** — read-only access to public content trees, routes, media, schema definitions, and content data. No admin, auth, user management, or write operations.

## Architecture

Universal (browser + Node.js 18+) SDK using native `fetch`. No external runtime dependencies. Ships ESM + CJS via tsup.

### Directory Structure

```
modulacms-sdk/
  package.json
  tsconfig.json
  tsup.config.ts
  src/
    index.ts              # Public API re-exports
    client.ts             # ModulaClient class
    errors.ts             # ModulaError
    types/
      index.ts            # Type re-exports
      common.ts           # ULID, Timestamp, ContentFormat, NullableString
      content.ts          # ContentData, ContentField
      routes.ts           # Route
      media.ts            # Media, MediaDimension
      schema.ts           # Datatype, Field
```

### Client API

```ts
const cms = new ModulaClient({
  baseUrl: "https://example.com",
  apiKey: "optional-bearer-token",   // for gated content
  defaultFormat: "clean",            // optional, default output format
})

// Primary: fetch rendered content tree by slug
const page = await cms.getPage("about")
const page = await cms.getPage("blog", { format: "contentful" })

// Routes
const routes = await cms.listRoutes()
const route  = await cms.getRoute(ulid)

// Content data (raw nodes)
const items = await cms.listContentData()
const item  = await cms.getContentData(ulid)

// Content fields
const fields = await cms.listContentFields()
const field  = await cms.getContentField(ulid)

// Media
const media     = await cms.listMedia()
const mediaItem = await cms.getMedia(ulid)
const dims      = await cms.listMediaDimensions()
const dim       = await cms.getMediaDimension(ulid)

// Schema
const datatypes = await cms.listDatatypes()
const datatype  = await cms.getDatatype(ulid)
const fields    = await cms.listFields()
const field     = await cms.getField(ulid)
```

### TypeScript Types (from STRUCT_REFERENCE.md → SDK types)

Go nullable types map to TypeScript as:

| Go Type | TS Type |
|---------|---------|
| `string` | `string` |
| `sql.NullString` / `types.NullableXxxID` | `string \| null` |
| `types.Timestamp` | `string` (ISO 8601) |
| `int64` | `number` |
| `sql.NullInt64` | `number \| null` |
| `types.FieldType` | `string` |
| `types.ContentStatus` | `string` |
| `types.Slug` | `string` |

### Error Handling

```ts
class ModulaError extends Error {
  status: number
  body: unknown
}
```

Thrown on non-2xx responses. Callers use try/catch.

### Content Formats

Type-safe union for the `?format=` query parameter:

```ts
type ContentFormat = "contentful" | "sanity" | "strapi" | "wordpress" | "clean" | "raw"
```

## Implementation Steps

### Step 1: Project scaffolding
- `pnpm init`, install dev deps: `typescript`, `tsup`
- Create `tsconfig.json` (strict, ESNext target, declaration emit)
- Create `tsup.config.ts` (ESM + CJS, dts generation)
- Add scripts: `build`, `dev`, `typecheck`

### Step 2: Types (`src/types/`)
- `common.ts` — ULID, Timestamp, ContentFormat, NullableString
- `content.ts` — ContentData, ContentField (from STRUCT_REFERENCE)
- `routes.ts` — Route
- `media.ts` — Media, MediaDimension
- `schema.ts` — Datatype, Field
- `index.ts` — re-export all

### Step 3: Error class (`src/errors.ts`)
- `ModulaError` extending `Error` with `status` and `body`

### Step 4: Client (`src/client.ts`)
- `ModulaClientConfig` type (baseUrl, apiKey?, defaultFormat?)
- Internal `request<T>(path, params?)` method handling:
  - URL construction
  - Authorization header (if apiKey provided)
  - JSON parsing
  - Error wrapping
- Public methods: `getPage`, `listRoutes`, `getRoute`, `listContentData`, `getContentData`, `listContentFields`, `getContentField`, `listMedia`, `getMedia`, `listMediaDimensions`, `getMediaDimension`, `listDatatypes`, `getDatatype`, `listFields`, `getField`

### Step 5: Barrel export (`src/index.ts`)
- Export `ModulaClient`, all types, `ModulaError`

### Step 6: Build verification
- Run `pnpm build`, confirm ESM + CJS + .d.ts output
- Verify types compile cleanly

## Key Files

| File | Purpose |
|------|---------|
| `src/client.ts` | Core SDK logic — all API methods |
| `src/types/*.ts` | TypeScript interfaces derived from STRUCT_REFERENCE.md |
| `src/errors.ts` | Error class |
| `src/index.ts` | Public API surface |
| `tsup.config.ts` | Build config |

## Verification

1. `pnpm build` succeeds with no type errors
2. `dist/` contains `.js`, `.mjs`, `.d.ts` files
3. Exports are correct: `import { ModulaClient } from "modulacms-sdk"` resolves types
4. Manual review: every public content delivery endpoint from API_CONTRACT.md has a corresponding SDK method
