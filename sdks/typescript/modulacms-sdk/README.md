# @modulacms/sdk

TypeScript SDK for the ModulaCMS content delivery API. Provides typed, read-only access to content trees, routes, media, and schema definitions.

- Universal: works in browsers and Node.js 18+
- Zero runtime dependencies (only `@modulacms/types` for shared type definitions)
- Ships ESM + CJS with full type declarations
- Tree-shakeable (`sideEffects: false`)
- Optional runtime validation via type predicates

## Install

```bash
npm install @modulacms/sdk
# or
pnpm add @modulacms/sdk
```

## Quick Start

```ts
import { ModulaClient } from "@modulacms/sdk";

const cms = new ModulaClient({
  baseUrl: "https://example.com",
  apiKey: "your-api-key",
  defaultFormat: "clean",
  timeout: 5000,
});

// Fetch a rendered page by slug
const page = await cms.getPage("about");

// With a specific output format
const blog = await cms.getPage("blog", { format: "contentful" });

// List all routes
const routes = await cms.listRoutes();
```

### Runtime Validation

Use a `Validator` type predicate for guaranteed type safety on `getPage` responses:

```ts
import { ModulaClient } from "@modulacms/sdk";
import type { Validator } from "@modulacms/sdk";

interface HomePage {
  hero: { title: string; subtitle: string };
  sections: Array<{ heading: string; body: string }>;
}

const isHomePage: Validator<HomePage> = (data): data is HomePage =>
  typeof data === "object" &&
  data !== null &&
  "hero" in data &&
  "sections" in data;

const cms = new ModulaClient({ baseUrl: "https://example.com" });
const page = await cms.getPage("home", { validate: isHomePage });
// page is typed as HomePage with runtime guarantee
```

### Browser Usage with Cookie Auth

```ts
const cms = new ModulaClient({
  baseUrl: "https://example.com",
  credentials: "include", // send cookies for session-based auth
});
```

## API

### Constructor

```ts
new ModulaClient(config: ModulaClientConfig)
```

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `baseUrl` | `string` | Yes | Base URL of the ModulaCMS instance. Path segments are preserved (e.g. `https://example.com/cms`). Trailing slashes are stripped. |
| `apiKey` | `string` | No | Bearer token sent as `Authorization: Bearer <apiKey>`. |
| `defaultFormat` | `ContentFormat` | No | Default output format applied to `getPage` when no per-call format is specified. |
| `timeout` | `number` | No | Request timeout in milliseconds. Uses `AbortSignal.timeout()` internally. |
| `credentials` | `RequestCredentials` | No | Fetch credentials mode. Set to `"include"` for browser cookie auth. |

Throws `ModulaError` with `status: 0` if `baseUrl` is not a valid URL.

### Content Delivery

| Method | Return Type | Description |
|--------|-------------|-------------|
| `getPage<T>(slug, options?)` | `Promise<T>` | Fetch rendered content tree by route slug |

- `slug` -- Route slug (e.g. `"about"`, `"blog/post-1"`, `"/"`). Leading slashes are stripped; use `"/"` for the root page.
- `options.format` -- Overrides `defaultFormat` for this request.
- `options.validate` -- Runtime type predicate. Throws `ModulaError` if the response fails validation.

### Routes

| Method | Return Type | Description |
|--------|-------------|-------------|
| `listRoutes()` | `Promise<Route[]>` | List all routes |
| `getRoute(id)` | `Promise<Route>` | Get a route by ULID |

### Content Data

| Method | Return Type | Description |
|--------|-------------|-------------|
| `listContentData()` | `Promise<ContentData[]>` | List all content nodes |
| `getContentData(id)` | `Promise<ContentData>` | Get a content node by ULID |

### Content Fields

| Method | Return Type | Description |
|--------|-------------|-------------|
| `listContentFields()` | `Promise<ContentField[]>` | List all content field values |
| `getContentField(id)` | `Promise<ContentField>` | Get a content field by ULID |

### Media

| Method | Return Type | Description |
|--------|-------------|-------------|
| `listMedia()` | `Promise<Media[]>` | List all media items |
| `getMedia(id)` | `Promise<Media>` | Get a media item by ULID |
| `listMediaPaginated(params)` | `Promise<PaginatedResponse<Media>>` | List media with pagination |
| `listMediaDimensions()` | `Promise<MediaDimension[]>` | List all dimension presets |
| `getMediaDimension(id)` | `Promise<MediaDimension>` | Get a dimension preset by ULID |

#### Pagination

```ts
const result = await cms.listMediaPaginated({ limit: 20, offset: 0 });

console.log(result.data);    // Media[]
console.log(result.total);   // total count across all pages
console.log(result.limit);   // 20
console.log(result.offset);  // 0
```

### Schema

| Method | Return Type | Description |
|--------|-------------|-------------|
| `listDatatypes()` | `Promise<Datatype[]>` | List all datatype definitions |
| `getDatatype(id)` | `Promise<Datatype>` | Get a datatype by ULID |
| `listFields()` | `Promise<Field[]>` | List all field definitions |
| `getField(id)` | `Promise<Field>` | Get a field definition by ULID |

## Error Handling

All methods throw `ModulaError` on non-2xx responses or client-side failures:

```ts
import { ModulaClient, ModulaError } from "@modulacms/sdk";

try {
  const page = await cms.getPage("missing-page");
} catch (err) {
  if (err instanceof ModulaError) {
    err.status;       // HTTP status code (e.g. 404), or 0 for client-side errors
    err.message;      // error string extracted from response, or fallback message
    err.errorMessage; // body.error as string, or undefined if not present
    err.body;         // raw response body (parsed JSON object or plain string)
  }
}
```

Client-side errors (invalid config, empty slug) use `status: 0`:

```ts
try {
  new ModulaClient({ baseUrl: "not-a-url" });
} catch (err) {
  if (err instanceof ModulaError) {
    err.status; // 0
  }
}
```

## Content Formats

The `format` parameter controls the shape of the response from `getPage`. Available formats:

| Format | Description |
|--------|-------------|
| `"contentful"` | Contentful-compatible structure |
| `"sanity"` | Sanity-compatible structure |
| `"strapi"` | Strapi-compatible structure |
| `"wordpress"` | WordPress-compatible structure |
| `"clean"` | ModulaCMS native clean format |
| `"raw"` | Unprocessed content tree |

The `CONTENT_FORMATS` constant is available for runtime validation:

```ts
import { CONTENT_FORMATS } from "@modulacms/sdk";
// ["contentful", "sanity", "strapi", "wordpress", "clean", "raw"]
```

## Exported Types

```ts
// Client
import { ModulaClient, ModulaError, CONTENT_FORMATS } from "@modulacms/sdk";

// Type-only imports
import type {
  ModulaClientConfig,
  GetPageOptions,
  Validator,
  ContentFormat,
  Route,
  ContentData,
  ContentField,
  Media,
  MediaDimension,
  Datatype,
  Field,
  PaginationParams,
  PaginatedResponse,
  ULID,
  Timestamp,
  NullableString,
  NullableNumber,
} from "@modulacms/sdk";
```

For branded ID types (`ContentID`, `RouteID`, `MediaID`, etc.), content tree types (`ContentTree`, `ContentNode`), and additional enums (`ContentStatus`, `FieldType`), import directly from `@modulacms/types`.

## License

MIT
