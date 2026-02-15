# modulacms-sdk

[![CI](https://github.com/hegner123/modulacms-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/hegner123/modulacms-sdk/actions/workflows/ci.yml)

TypeScript SDK for the ModulaCMS content delivery API. Provides typed, read-only access to content trees, routes, media, and schema definitions.

- Universal: works in browsers and Node.js 18+
- Zero runtime dependencies
- Ships ESM + CJS with full type declarations
- Optional runtime validation via type predicates

## Install

```bash
npm install modulacms-sdk
```

## Quick Start

```ts
import { ModulaClient } from "modulacms-sdk";

const cms = new ModulaClient({
  baseUrl: "https://example.com",
  apiKey: "optional-bearer-token",
  defaultFormat: "clean",
});

// Fetch a rendered page by slug
const page = await cms.getPage("about");

// With a specific output format
const page = await cms.getPage("blog", { format: "contentful" });

// With runtime validation
import type { Validator } from "modulacms-sdk";

interface HomePage {
  hero: { title: string };
}

const isHomePage: Validator<HomePage> = (data): data is HomePage =>
  typeof data === "object" && data !== null && "hero" in data;

const page = await cms.getPage("home", { validate: isHomePage });
```

## API

### Constructor

```ts
new ModulaClient(config: ModulaClientConfig)
```

| Option | Type | Description |
|--------|------|-------------|
| `baseUrl` | `string` | Base URL of the ModulaCMS instance (required) |
| `apiKey` | `string` | Bearer token for API key auth |
| `defaultFormat` | `ContentFormat` | Default output format for `getPage` |
| `timeout` | `number` | Request timeout in milliseconds |
| `credentials` | `RequestCredentials` | Fetch credentials mode (e.g. `"include"` for cookie auth) |

### Content Delivery

| Method | Return Type | Description |
|--------|-------------|-------------|
| `getPage<T>(slug, options?)` | `Promise<T>` | Fetch rendered content tree by route slug |

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
| `listMediaDimensions()` | `Promise<MediaDimension[]>` | List all dimension presets |
| `getMediaDimension(id)` | `Promise<MediaDimension>` | Get a dimension preset by ULID |

### Schema

| Method | Return Type | Description |
|--------|-------------|-------------|
| `listDatatypes()` | `Promise<Datatype[]>` | List all datatype definitions |
| `getDatatype(id)` | `Promise<Datatype>` | Get a datatype by ULID |
| `listFields()` | `Promise<Field[]>` | List all field definitions |
| `getField(id)` | `Promise<Field>` | Get a field definition by ULID |

## Error Handling

All methods throw `ModulaError` on non-2xx responses:

```ts
import { ModulaError } from "modulacms-sdk";

try {
  const page = await cms.getPage("missing");
} catch (err) {
  if (err instanceof ModulaError) {
    console.log(err.status);       // 404
    console.log(err.errorMessage); // "Not found"
    console.log(err.body);         // raw response body
  }
}
```

## Content Formats

Available output formats for the `?format=` query parameter:

| Format | Description |
|--------|-------------|
| `"contentful"` | Contentful-compatible structure |
| `"sanity"` | Sanity.io-compatible structure |
| `"strapi"` | Strapi-compatible structure |
| `"wordpress"` | WordPress-compatible structure |
| `"clean"` | ModulaCMS native clean format |
| `"raw"` | Unprocessed content tree |

Use the `CONTENT_FORMATS` array for runtime validation:

```ts
import { CONTENT_FORMATS } from "modulacms-sdk";

if (CONTENT_FORMATS.includes(userInput as any)) {
  // valid format
}
```

## Exported Types

All types are importable for use in consumer code:

```ts
import type {
  ModulaClientConfig,
  GetPageOptions,
  Validator,
  ULID,
  Timestamp,
  ContentFormat,
  Route,
  ContentData,
  ContentField,
  Media,
  MediaDimension,
  Datatype,
  Field,
} from "modulacms-sdk";
```

## License

MIT
