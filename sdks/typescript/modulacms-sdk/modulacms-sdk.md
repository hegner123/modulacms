# modulacms-sdk

TypeScript SDK for the ModulaCMS content delivery API. Provides type-safe, read-only access to content trees, routes, media, and schema definitions.

## Overview

The ModulaCMS SDK is a client library for consuming content from ModulaCMS instances. It wraps the content delivery API with TypeScript types, runtime validation support, and error handling.

All API methods throw ModulaError on non-2xx responses. Types reflect expected API shapes but are not validated at runtime unless an explicit Validator is provided to getPage.

The SDK supports multiple content output formats including contentful, sanity, strapi, wordpress, clean, and raw. Format selection can be configured globally via defaultFormat or per-request via GetPageOptions.

## Installation

```bash
npm install modulacms-sdk
```

## Quick Start

```typescript
import { ModulaClient } from "modulacms-sdk";

const cms = new ModulaClient({
  baseUrl: "https://example.com",
  apiKey: "optional-bearer-token",
  defaultFormat: "clean"
});

const page = await cms.getPage("about");
const routes = await cms.listRoutes();
```

## class ModulaClient

Client for the ModulaCMS content delivery API. Create an instance with ModulaClientConfig and call methods to fetch content.

```typescript
const cms = new ModulaClient({
  baseUrl: "https://example.com",
  apiKey: "optional-key",
  timeout: 5000
});
```

All methods are async and return promises. Non-2xx responses throw ModulaError.

### constructor(config: ModulaClientConfig)

Create a new ModulaCMS client.

Throws ModulaError if baseUrl is not a valid URL. The baseUrl is normalized by stripping trailing slashes and preserving path segments.

```typescript
const cms = new ModulaClient({
  baseUrl: "https://example.com/cms",
  apiKey: "my-api-key",
  defaultFormat: "clean",
  timeout: 5000
});
```

#### async getPage

```typescript
async getPage<T = unknown>(slug: string, options?: GetPageOptions<T>): Promise<T>
```

Fetch a rendered content tree by route slug. This is the primary content delivery method. The API resolves the slug to a route, builds the full content tree, and returns it in the requested output format.

The slug must not be empty or start with a forward slash. Format can be overridden per-request via GetPageOptions.format, otherwise defaultFormat is used.

Throws ModulaError if the slug is invalid, the route is not found, or validation fails.

```typescript
const page = await cms.getPage("about");

const typed = await cms.getPage<HomePage>("home", {
  format: "contentful",
  validate: isHomePage
});
```

#### async listRoutes

```typescript
async listRoutes(): Promise<Route[]>
```

List all routes. Returns all routes registered in the CMS.

Throws ModulaError on non-2xx response.

```typescript
const routes = await cms.listRoutes();
```

#### async getRoute

```typescript
async getRoute(id: string): Promise<Route>
```

Get a single route by ID. The id parameter is a ULID.

Returns the matching route. Throws ModulaError on non-2xx response such as 404 if not found.

```typescript
const route = await cms.getRoute("01HXK4N2F8QZJV3K7M1Y9ABCDE");
```

#### async listContentData

```typescript
async listContentData(): Promise<ContentData[]>
```

List all content data nodes. Returns all content nodes in the CMS.

Throws ModulaError on non-2xx response.

```typescript
const nodes = await cms.listContentData();
```

#### async getContentData

```typescript
async getContentData(id: string): Promise<ContentData>
```

Get a single content data node by ID. The id parameter is a ULID.

Returns the matching content node. Throws ModulaError on non-2xx response such as 404 if not found.

```typescript
const node = await cms.getContentData("01HXK4N2F8QZJV3K7M1Y9ABCDE");
```

#### async listContentFields

```typescript
async listContentFields(): Promise<ContentField[]>
```

List all content field values. Returns all content field entries in the CMS.

Throws ModulaError on non-2xx response.

```typescript
const fields = await cms.listContentFields();
```

#### async getContentField

```typescript
async getContentField(id: string): Promise<ContentField>
```

Get a single content field value by ID. The id parameter is a ULID.

Returns the matching content field. Throws ModulaError on non-2xx response such as 404 if not found.

```typescript
const field = await cms.getContentField("01HXK4N2F8QZJV3K7M1Y9ABCDE");
```

#### async listMedia

```typescript
async listMedia(): Promise<Media[]>
```

List all media items. Returns all media records in the CMS.

Throws ModulaError on non-2xx response.

```typescript
const mediaItems = await cms.listMedia();
```

#### async getMedia

```typescript
async getMedia(id: string): Promise<Media>
```

Get a single media item by ID. The id parameter is a ULID.

Returns the matching media record. Throws ModulaError on non-2xx response such as 404 if not found.

```typescript
const media = await cms.getMedia("01HXK4N2F8QZJV3K7M1Y9ABCDE");
```

#### async listMediaDimensions

```typescript
async listMediaDimensions(): Promise<MediaDimension[]>
```

List all media dimension presets. Returns all dimension presets defined in the CMS.

Throws ModulaError on non-2xx response.

```typescript
const dimensions = await cms.listMediaDimensions();
```

#### async getMediaDimension

```typescript
async getMediaDimension(id: string): Promise<MediaDimension>
```

Get a single media dimension preset by ID. The id parameter is a ULID.

Returns the matching dimension preset. Throws ModulaError on non-2xx response such as 404 if not found.

```typescript
const dimension = await cms.getMediaDimension("01HXK4N2F8QZJV3K7M1Y9ABCDE");
```

#### async listDatatypes

```typescript
async listDatatypes(): Promise<Datatype[]>
```

List all datatype definitions. Returns all datatypes registered in the CMS. Datatypes are content schemas that define the structure of content nodes.

Throws ModulaError on non-2xx response.

```typescript
const datatypes = await cms.listDatatypes();
```

#### async getDatatype

```typescript
async getDatatype(id: string): Promise<Datatype>
```

Get a single datatype definition by ID. The id parameter is a ULID.

Returns the matching datatype. Throws ModulaError on non-2xx response such as 404 if not found.

```typescript
const datatype = await cms.getDatatype("01HXK4N2F8QZJV3K7M1Y9ABCDE");
```

#### async listFields

```typescript
async listFields(): Promise<Field[]>
```

List all field definitions. Returns all field definitions in the CMS. Fields are schema building blocks that define individual data points within datatypes.

Throws ModulaError on non-2xx response.

```typescript
const fields = await cms.listFields();
```

#### async getField

```typescript
async getField(id: string): Promise<Field>
```

Get a single field definition by ID. The id parameter is a ULID.

Returns the matching field definition. Throws ModulaError on non-2xx response such as 404 if not found.

```typescript
const field = await cms.getField("01HXK4N2F8QZJV3K7M1Y9ABCDE");
```

## class ModulaError

Error thrown by ModulaClient on non-2xx API responses or client-side validation failures.

```typescript
export class ModulaError extends Error {
  readonly status: number;
  readonly body: unknown;
  get errorMessage(): string | undefined;
}
```

The status field is the HTTP status code from the API response, or 0 for client-side errors such as invalid config. The body field contains the raw response body as either a parsed JSON object or plain text string.

```typescript
try {
  const page = await cms.getPage("about");
} catch (err) {
  if (err instanceof ModulaError) {
    console.log(err.status);
    console.log(err.errorMessage);
    console.log(err.body);
  }
}
```

### constructor(status: number, body: unknown)

Create a ModulaError with status code and response body. If the body contains an error property, that value is used as the error message. Otherwise the message defaults to a status-based string.

```typescript
const err = new ModulaError(404, { error: "Not found" });
console.log(err.message);
```

### get errorMessage

Convenience accessor that extracts the error string from the response body. Returns undefined if the body does not contain an error property.

```typescript
const err = new ModulaError(404, { error: "Not found" });
console.log(err.errorMessage);
```

## interface ModulaClientConfig

Configuration for creating a ModulaClient instance.

```typescript
export interface ModulaClientConfig {
  baseUrl: string;
  apiKey?: string;
  defaultFormat?: ContentFormat;
  timeout?: number;
  credentials?: RequestCredentials;
}
```

The baseUrl is required and must be a valid URL. It can include a path segment such as https://example.com/cms. The apiKey is optional and sent as Authorization Bearer header when provided.

The defaultFormat applies to all getPage calls unless overridden per-request. The timeout is in milliseconds and aborts requests that exceed the duration. The credentials mode controls fetch credential behavior such as include for browser cookie authentication.

```typescript
const config: ModulaClientConfig = {
  baseUrl: "https://example.com",
  apiKey: "my-api-key",
  defaultFormat: "clean",
  timeout: 5000
};
```

## interface GetPageOptions

Options for ModulaClient getPage method.

```typescript
export interface GetPageOptions<T = unknown> {
  format?: ContentFormat;
  validate?: Validator<T>;
}
```

The format field overrides the defaultFormat for this request. The validate field is an optional runtime validator that checks the response before returning. Throws ModulaError on validation failure.

```typescript
const page = await cms.getPage("about", {
  format: "contentful",
  validate: isAboutPage
});
```

## type Validator

```typescript
export type Validator<T> = (data: unknown) => data is T;
```

A type predicate function used to validate API response data at runtime. Pass to GetPageOptions.validate to get runtime-safe type narrowing.

```typescript
const isHomePage: Validator<HomePage> = (data): data is HomePage =>
  typeof data === "object" && data !== null && "hero" in data;

const page = await cms.getPage("home", { validate: isHomePage });
```

## interface Route

A URL route that maps a slug to a content tree. Routes are the top-level addressing mechanism. Each route slug is used with ModulaClient.getPage to fetch rendered content.

```typescript
export interface Route {
  route_id: ULID;
  slug: string;
  title: string;
  status: number;
  author_id: NullableString;
  date_created: Timestamp;
  date_modified: Timestamp;
}
```

The slug is URL-safe and used to address the route. The title is human-readable. The status is a numeric code defined by the API where 0 is inactive and 1 is active. Author and timestamp fields track creation and modification.

## interface ContentData

A content node in the CMS content tree. Content nodes are arranged in a tree structure using parent child and sibling pointers. Each node belongs to a route and may be associated with a datatype.

```typescript
export interface ContentData {
  content_data_id: ULID;
  parent_id: NullableString;
  first_child_id: NullableString;
  next_sibling_id: NullableString;
  prev_sibling_id: NullableString;
  route_id: NullableString;
  datatype_id: NullableString;
  author_id: NullableString;
  status: string;
  date_created: Timestamp;
  date_modified: Timestamp;
}
```

The parent_id is null for root nodes. The first_child_id next_sibling_id and prev_sibling_id fields support tree traversal. The route_id links this node to its route. The datatype_id defines the node schema. The status is an API-defined value such as draft or published.

## interface ContentField

A field value attached to a content node. Content fields store the actual data for each content node, linking a field definition to its value within a specific content context.

```typescript
export interface ContentField {
  content_field_id: ULID;
  route_id: NullableString;
  content_data_id: NullableString;
  field_id: NullableString;
  field_value: string;
  author_id: NullableString;
  date_created: Timestamp;
  date_modified: Timestamp;
}
```

The content_data_id links this field to its content node. The field_id references the field definition. The field_value contains the stored data. Author and timestamp fields track creation and modification.

## interface Media

A media item managed by the CMS. Media records store metadata about uploaded files including display properties, dimensions, and the public URL.

```typescript
export interface Media {
  media_id: ULID;
  name: NullableString;
  display_name: NullableString;
  alt: NullableString;
  caption: NullableString;
  description: NullableString;
  class: NullableString;
  mimetype: NullableString;
  dimensions: NullableString;
  url: string;
  srcset: NullableString;
  author_id: NullableString;
  date_created: Timestamp;
  date_modified: Timestamp;
}
```

The name is the internal filename. The display_name is human-readable. The alt caption and description fields support accessibility and metadata. The class field is a CSS class hint. The mimetype is detected automatically. The dimensions field is a string such as 1920x1080. The url is the public access URL. The srcset is a responsive image srcset string.

## interface MediaDimension

A named dimension preset for media assets. Used to define standard image sizes across the CMS such as thumbnail or hero.

```typescript
export interface MediaDimension {
  md_id: string;
  label: NullableString;
  width: NullableNumber;
  height: NullableNumber;
  aspect_ratio: NullableString;
}
```

The md_id field is abbreviated from the API schema. The label is human-readable such as Thumbnail or Hero Banner. The width and height are in pixels and may be null if unconstrained. The aspect_ratio is a string such as 16:9 and may be null if unconstrained.

## interface Datatype

A datatype defines a content schema. It is the structural blueprint for a category of content nodes such as Blog Post or Hero Section.

```typescript
export interface Datatype {
  datatype_id: ULID;
  parent_id: NullableString;
  label: string;
  type: string;
  author_id: NullableString;
  date_created: Timestamp;
  date_modified: Timestamp;
}
```

The parent_id supports hierarchical datatypes and is null for top-level datatypes. The label is human-readable. The type is a datatype category or kind as defined by the API. Author and timestamp fields track creation and modification.

## interface Field

A field definition within a datatype schema. Fields define the individual data points that content nodes of a given datatype can hold such as Title Body or Featured Image.

```typescript
export interface Field {
  field_id: ULID;
  parent_id: NullableString;
  label: string;
  data: string;
  validation: string;
  ui_config: string;
  type: string;
  author_id: NullableString;
  date_created: Timestamp;
  date_modified: Timestamp;
}
```

The parent_id is the datatype this field belongs to or null if standalone. The label is human-readable. The data validation and ui_config fields are JSON-serialized strings. Parse them with JSON.parse and validate at runtime. The type is a field type identifier such as text richtext image or relation.

## type ULID

```typescript
export type ULID = string;
```

A ULID is a Universally Unique Lexicographically Sortable Identifier string. All entity primary keys in ModulaCMS use this format.

Example: 01HXK4N2F8QZJV3K7M1Y9ABCDE

## type Timestamp

```typescript
export type Timestamp = string;
```

An ISO 8601 RFC 3339 UTC timestamp string as returned by the API.

Example: 2026-01-30T12:00:00Z

## type NullableString

```typescript
export type NullableString = string | null;
```

A string value that may be null when the API field is not set.

## type NullableNumber

```typescript
export type NullableNumber = number | null;
```

A numeric value that may be null when the API field is not set.

## type ContentFormat

```typescript
export type ContentFormat = (typeof CONTENT_FORMATS)[number];
```

A content output format accepted by the format query parameter. Derived from CONTENT_FORMATS to keep the runtime array and type in sync.

Valid values: contentful, sanity, strapi, wordpress, clean, raw

## const CONTENT_FORMATS

```typescript
export const CONTENT_FORMATS = ["contentful", "sanity", "strapi", "wordpress", "clean", "raw"] as const;
```

All supported content output format identifiers. Use this array for runtime validation of format values.

```typescript
if (CONTENT_FORMATS.includes(userInput)) {
  const format: ContentFormat = userInput;
}
```

## Authentication

The SDK supports two authentication modes. Bearer token authentication via apiKey in ModulaClientConfig sends Authorization Bearer header on all requests. Cookie authentication via credentials include in ModulaClientConfig uses browser cookies.

```typescript
const cms = new ModulaClient({
  baseUrl: "https://example.com",
  apiKey: "my-bearer-token"
});

const browserCms = new ModulaClient({
  baseUrl: "https://example.com",
  credentials: "include"
});
```

## Error Handling

All methods throw ModulaError on non-2xx responses. Catch the error to access status code, error message, and raw response body.

```typescript
try {
  const page = await cms.getPage("about");
} catch (err) {
  if (err instanceof ModulaError) {
    if (err.status === 404) {
      console.log("Page not found");
    } else {
      console.log(err.errorMessage);
    }
  }
}
```

## Runtime Type Safety

The SDK types reflect expected API shapes but are not validated at runtime by default. For guaranteed type safety, use the validate option with getPage.

```typescript
interface HomePage {
  hero: string;
  sections: Section[];
}

const isHomePage: Validator<HomePage> = (data): data is HomePage =>
  typeof data === "object" &&
  data !== null &&
  "hero" in data &&
  "sections" in data;

const page = await cms.getPage("home", { validate: isHomePage });
```

The validator is a type predicate function. If validation fails, getPage throws ModulaError with the response data in the body.

## Timeouts

Set a timeout in milliseconds via ModulaClientConfig to abort requests that exceed the duration. Uses AbortSignal timeout internally.

```typescript
const cms = new ModulaClient({
  baseUrl: "https://example.com",
  timeout: 5000
});
```

Requests exceeding the timeout are aborted and throw an error.

## Base URL Handling

The baseUrl is normalized during client construction by stripping trailing slashes and preserving path segments. This ensures consistent URL construction across all API calls.

```typescript
const cms = new ModulaClient({
  baseUrl: "https://example.com/cms///"
});

await cms.listRoutes();
```

The client strips trailing slashes and constructs the URL as https://example.com/cms/api/v1/routes.
