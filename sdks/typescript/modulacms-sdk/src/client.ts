import { ModulaError } from "./errors.js";
import type { ContentFormat, ContentData, ContentField, Route, Media, MediaDimension, Datatype, Field, PaginationParams, PaginatedResponse, QueryParams, QueryResult } from "@modulacms/types";

/**
 * A type predicate function used to validate API response data at runtime.
 * Pass to {@link GetPageOptions.validate} to get runtime-safe type narrowing.
 *
 * @typeParam T - The expected type after validation.
 *
 * @example
 * const isHomePage: Validator<HomePage> = (data): data is HomePage =>
 *   typeof data === "object" && data !== null && "hero" in data;
 *
 * const page = await cms.getPage("home", { validate: isHomePage });
 * // page is typed as HomePage with runtime guarantee
 */
export type Validator<T> = (data: unknown) => data is T;

/**
 * Configuration for creating a {@link ModulaClient} instance.
 *
 * @example
 * const config: ModulaClientConfig = {
 *   baseUrl: "https://example.com",
 *   apiKey: "my-api-key",
 *   defaultFormat: "clean",
 *   timeout: 5000,
 * };
 */
export interface ModulaClientConfig {
  /** Base URL of the ModulaCMS instance (e.g. "https://example.com" or "https://example.com/cms"). */
  baseUrl: string;
  /** Optional Bearer token for API key authentication. Sent as `Authorization: Bearer <apiKey>`. */
  apiKey?: string;
  /** Default content output format applied to {@link ModulaClient.getPage} when no format is specified per-call. */
  defaultFormat?: ContentFormat;
  /** Request timeout in milliseconds. When set, requests that exceed this duration are aborted. */
  timeout?: number;
  /** Credentials mode for fetch. Set to `"include"` for browser cookie authentication. */
  credentials?: RequestCredentials;
}

/**
 * Options for {@link ModulaClient.getPage}.
 *
 * @typeParam T - The expected page content type. Defaults to `unknown` when no validator is provided.
 */
export interface GetPageOptions<T = unknown> {
  /** Content output format override for this request. Takes precedence over {@link ModulaClientConfig.defaultFormat}. */
  format?: ContentFormat;
  /** Locale code to request content in a specific locale (e.g. `"en"`, `"fr"`). */
  locale?: string;
  /** Optional runtime validator. When provided, the response is checked before returning. Throws {@link ModulaError} on validation failure. */
  validate?: Validator<T>;
}

/**
 * Options for {@link ModulaClient.search}.
 * All fields are optional; zero values are omitted from the request.
 */
export interface SearchOptions {
  /** Filter results to a specific datatype name (e.g. "blog_post"). */
  type?: string;
  /** Filter results to a specific locale code (e.g. "en", "fr"). */
  locale?: string;
  /** Maximum number of results to return. Server default is 20. */
  limit?: number;
  /** Number of results to skip for pagination. */
  offset?: number;
  /** Enable prefix matching on the last query term for search-as-you-type. */
  prefix?: boolean;
}

/**
 * A single search hit with relevance score and snippet.
 */
export interface SearchResult {
  id: string;
  content_data_id: string;
  route_slug: string;
  route_title: string;
  datatype_name: string;
  datatype_label: string;
  section?: string;
  section_anchor?: string;
  score: number;
  snippet: string;
  published_at: string;
}

/**
 * The envelope returned by a search query.
 */
export interface SearchResponse {
  query: string;
  results: SearchResult[];
  total: number;
  limit: number;
  offset: number;
}

/**
 * Response from the health check endpoint.
 */
export interface HealthResponse {
  status: 'ok' | 'degraded';
  environment: string;
  checks: {
    database: boolean;
    storage: boolean;
    plugins: boolean;
  };
  details: {
    database?: string;
    storage?: string;
    plugins?: string;
  };
}

/**
 * Response from the environment endpoint.
 */
export interface EnvironmentResponse {
  environment: string;
  stage: string;
}

/**
 * A fully composed content data view with resolved fields and relations.
 */
export interface ContentDataFullView {
  [key: string]: unknown;
}

/**
 * A fully composed route view with resolved content data.
 */
export interface RouteFullView {
  [key: string]: unknown;
}

/**
 * A media item with full metadata including dimensions and references.
 */
export interface MediaFullItem {
  [key: string]: unknown;
}

/**
 * A fully composed datatype view with all field definitions.
 */
export interface DatatypeFullView {
  [key: string]: unknown;
}

/**
 * A datatype summary in the full list response.
 */
export interface DatatypeFullListItem {
  [key: string]: unknown;
}

/**
 * Client for the ModulaCMS content delivery API.
 * Provides read-only access to content trees, routes, media, and schema definitions.
 *
 * All methods throw {@link ModulaError} on non-2xx responses or validation failures.
 * Types on returned data reflect the expected API shape but are not validated at runtime
 * unless a {@link Validator} is provided.
 *
 * @example
 * const cms = new ModulaClient({
 *   baseUrl: "https://example.com",
 *   apiKey: "optional-bearer-token",
 *   defaultFormat: "clean",
 *   timeout: 5000,
 * });
 *
 * const page = await cms.getPage("about");
 * const routes = await cms.listRoutes();
 */
export class ModulaClient {
  private baseUrl: string;
  private apiKey: string | undefined;
  private defaultFormat: ContentFormat | undefined;
  private timeout: number | undefined;
  private credentials: RequestCredentials | undefined;

  /**
   * Create a new ModulaCMS client.
   *
   * @param config - Client configuration. See {@link ModulaClientConfig}.
   * @throws {@link ModulaError} If `baseUrl` is not a valid URL.
   */
  constructor(config: ModulaClientConfig) {
    try {
      const parsed = new URL(config.baseUrl);
      // Strip trailing slashes for consistent URL construction
      this.baseUrl = parsed.origin + parsed.pathname.replace(/\/+$/, "");
    } catch {
      throw new ModulaError(0, { error: `Invalid baseUrl: ${config.baseUrl}` });
    }
    this.apiKey = config.apiKey;
    this.defaultFormat = config.defaultFormat;
    this.timeout = config.timeout;
    this.credentials = config.credentials;
  }

  private async request<T>(
    path: string,
    params?: Record<string, string>,
    validate?: Validator<T>,
  ): Promise<T> {
    const url = new URL(this.baseUrl + path);
    if (params) {
      for (const [key, value] of Object.entries(params)) {
        url.searchParams.set(key, value);
      }
    }

    const headers: Record<string, string> = {
      "Accept": "application/json",
    };
    if (this.apiKey) {
      headers["Authorization"] = `Bearer ${this.apiKey}`;
    }

    const fetchOptions: RequestInit = { headers };

    if (this.timeout !== undefined) {
      fetchOptions.signal = AbortSignal.timeout(this.timeout);
    }

    if (this.credentials !== undefined) {
      fetchOptions.credentials = this.credentials;
    }

    const response = await fetch(url.toString(), fetchOptions);

    if (!response.ok) {
      let body: unknown;
      try {
        body = await response.json();
      } catch {
        body = await response.text();
      }
      throw new ModulaError(response.status, body);
    }

    const data: unknown = await response.json();

    if (validate && !validate(data)) {
      throw new ModulaError(response.status, {
        error: "Response failed validation",
        data,
      });
    }

    // Assertion: either validated above via type predicate, or caller accepts
    // responsibility for the shape (same trade-off as every HTTP client SDK).
    // Typed methods (listRoutes, getRoute, etc.) reflect the expected API shape
    // but are not validated at runtime. Use getPage() with a validator for
    // guaranteed type safety.
    return data as T;
  }

  /**
   * Fetch a rendered content tree by route slug.
   *
   * This is the primary content delivery method. The API resolves the slug to a route,
   * builds the full content tree, and returns it in the requested output format.
   *
   * @typeParam T - Expected shape of the page content. Defaults to `unknown`.
   * @param slug - Route slug (e.g. "about", "blog", "/"). Leading slashes are accepted; use "/" for the root page.
   * @param options - Optional format override and/or runtime validator.
   * @returns The rendered content tree.
   * @throws {@link ModulaError} If the slug is invalid, the route is not found (404), or validation fails.
   *
   * @example
   * // Untyped (returns unknown):
   * const page = await cms.getPage("about");
   *
   * // With type parameter (no runtime validation):
   * const page = await cms.getPage<MyPageType>("about");
   *
   * // With format override:
   * const page = await cms.getPage("blog", { format: "contentful" });
   *
   * // With runtime validator (guaranteed type safety):
   * const page = await cms.getPage("about", { validate: isMyPage });
   */
  async getPage<T = unknown>(slug: string, options?: GetPageOptions<T>): Promise<T> {
    if (!slug) {
      throw new ModulaError(0, { error: `Invalid slug: "${slug}". Slugs should not be empty.` });
    }
    const format = options?.format ?? this.defaultFormat;
    const params: Record<string, string> = {};
    if (format) {
      params.format = format;
    }
    if (options?.locale) {
      params.locale = options.locale;
    }
    const trimmed = slug.startsWith("/") ? slug.slice(1) : slug;
    const path = `/api/v1/content/${trimmed}`;
    return this.request<T>(path, params, options?.validate);
  }

  /**
   * List all routes.
   *
   * @returns All routes registered in the CMS.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listRoutes(): Promise<Route[]> {
    return this.request<Route[]>("/api/v1/routes");
  }

  /**
   * Get a single route by ID.
   *
   * @param id - ULID of the route to fetch.
   * @returns The matching route.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getRoute(id: string): Promise<Route> {
    return this.request<Route>("/api/v1/routes/", { q: id });
  }

  /**
   * List all content data nodes.
   *
   * @returns All content nodes in the CMS.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listContentData(): Promise<ContentData[]> {
    return this.request<ContentData[]>("/api/v1/contentdata");
  }

  /**
   * Get a single content data node by ID.
   *
   * @param id - ULID of the content node to fetch.
   * @returns The matching content node.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getContentData(id: string): Promise<ContentData> {
    return this.request<ContentData>("/api/v1/contentdata/", { q: id });
  }

  /**
   * List all content field values.
   *
   * @returns All content field entries in the CMS.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listContentFields(): Promise<ContentField[]> {
    return this.request<ContentField[]>("/api/v1/contentfields");
  }

  /**
   * Get a single content field value by ID.
   *
   * @param id - ULID of the content field entry to fetch.
   * @returns The matching content field.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getContentField(id: string): Promise<ContentField> {
    return this.request<ContentField>("/api/v1/contentfields/", { q: id });
  }

  /**
   * List all media items.
   *
   * @returns All media records in the CMS.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listMedia(): Promise<Media[]> {
    return this.request<Media[]>("/api/v1/media");
  }

  /**
   * Get a single media item by ID.
   *
   * @param id - ULID of the media item to fetch.
   * @returns The matching media record.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getMedia(id: string): Promise<Media> {
    return this.request<Media>("/api/v1/media/", { q: id });
  }

  /**
   * List media items with pagination.
   *
   * @param params - Pagination parameters (limit and offset).
   * @returns A paginated envelope containing the data page and total count.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listMediaPaginated(params: PaginationParams): Promise<PaginatedResponse<Media>> {
    return this.request<PaginatedResponse<Media>>("/api/v1/media", {
      limit: String(params.limit),
      offset: String(params.offset),
    });
  }

  /**
   * List all media dimension presets.
   *
   * @returns All dimension presets defined in the CMS.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listMediaDimensions(): Promise<MediaDimension[]> {
    return this.request<MediaDimension[]>("/api/v1/mediadimensions");
  }

  /**
   * Get a single media dimension preset by ID.
   *
   * @param id - ULID of the dimension preset to fetch.
   * @returns The matching dimension preset.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getMediaDimension(id: string): Promise<MediaDimension> {
    return this.request<MediaDimension>("/api/v1/mediadimensions/", { q: id });
  }

  /**
   * List all datatype definitions.
   *
   * @returns All datatypes (content schemas) registered in the CMS.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listDatatypes(): Promise<Datatype[]> {
    return this.request<Datatype[]>("/api/v1/datatype");
  }

  /**
   * Get a single datatype definition by ID.
   *
   * @param id - ULID of the datatype to fetch.
   * @returns The matching datatype.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getDatatype(id: string): Promise<Datatype> {
    return this.request<Datatype>("/api/v1/datatype/", { q: id });
  }

  /**
   * List all field definitions.
   *
   * @returns All field definitions (schema building blocks) in the CMS.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listFields(): Promise<Field[]> {
    return this.request<Field[]>("/api/v1/fields");
  }

  /**
   * Get a single field definition by ID.
   *
   * @param id - ULID of the field definition to fetch.
   * @returns The matching field definition.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getField(id: string): Promise<Field> {
    return this.request<Field>("/api/v1/fields/", { q: id });
  }

  /**
   * Query content items by datatype name with optional filtering, sorting, and pagination.
   *
   * @param datatype - The datatype name to query (e.g. "blog_post").
   * @param params - Optional query parameters (filters, sort, limit, offset, locale, status).
   * @returns A paginated query result envelope.
   * @throws {@link ModulaError} On non-2xx response.
   *
   * @example
   * const result = await cms.queryContent("blog_post", {
   *   filters: { "category": "engineering" },
   *   sort: "-published_at",
   *   limit: 10,
   * });
   */
  /**
   * Fetch all published global content trees.
   *
   * Global content items are root nodes typed as `_global` whose published trees
   * are available site-wide (e.g. navigation, footer, site settings).
   *
   * @returns The raw globals response. Shape depends on datatype schema configuration.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async getGlobals(): Promise<unknown> {
    return this.request<unknown>("/api/v1/globals");
  }

  async queryContent(datatype: string, params?: QueryParams): Promise<QueryResult> {
    const p: Record<string, string> = {};
    if (params?.sort) p.sort = params.sort;
    if (params?.limit !== undefined) p.limit = String(params.limit);
    if (params?.offset !== undefined) p.offset = String(params.offset);
    if (params?.locale) p.locale = params.locale;
    if (params?.status) p.status = params.status;
    if (params?.filters) {
      for (const [key, value] of Object.entries(params.filters)) {
        p[key] = value;
      }
    }
    return this.request<QueryResult>(`/api/v1/query/${encodeURIComponent(datatype)}`, p);
  }

  /**
   * Execute a full-text search against published content.
   *
   * This is a public endpoint that requires no authentication. The search index
   * covers published content only.
   *
   * @param query - The search query string. Required.
   * @param options - Optional filtering and pagination parameters.
   * @returns The search response with results, total count, and pagination info.
   * @throws {@link ModulaError} On non-2xx response.
   *
   * @example
   * const results = await cms.search("installation guide");
   *
   * // With options:
   * const results = await cms.search("guide", {
   *   type: "doc_page",
   *   limit: 10,
   *   prefix: true,
   * });
   */
  async search(query: string, options?: SearchOptions): Promise<SearchResponse> {
    const params: Record<string, string> = { q: query };
    if (options?.type) params.type = options.type;
    if (options?.locale) params.locale = options.locale;
    if (options?.limit !== undefined) params.limit = String(options.limit);
    if (options?.offset !== undefined) params.offset = String(options.offset);
    if (options?.prefix !== undefined) params.prefix = String(options.prefix);
    return this.request<SearchResponse>("/api/v1/search", params);
  }

  /**
   * Check the health status of the CMS instance.
   *
   * @returns Health status including database, storage, and plugin checks.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async health(): Promise<HealthResponse> {
    return this.request<HealthResponse>("/api/v1/health");
  }

  /**
   * Get the current environment and stage information.
   *
   * @returns Environment name and stage.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async environment(): Promise<EnvironmentResponse> {
    return this.request<EnvironmentResponse>("/api/v1/environment");
  }

  /**
   * Get a fully composed content data node by ID, including resolved fields and relations.
   *
   * @param id - ULID of the content node to fetch.
   * @returns The fully composed content data view.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getContentDataFull(id: string): Promise<ContentDataFullView> {
    return this.request<ContentDataFullView>("/api/v1/contentdata/full", { q: id });
  }

  /**
   * Get a fully composed route by ID, including resolved content data.
   *
   * @param id - ULID of the route to fetch.
   * @returns The fully composed route view.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getRouteFull(id: string): Promise<RouteFullView> {
    return this.request<RouteFullView>("/api/v1/routes/full", { q: id });
  }

  /**
   * List all media items with full metadata.
   *
   * @returns All media items with full metadata.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listMediaFull(): Promise<MediaFullItem[]> {
    return this.request<MediaFullItem[]>("/api/v1/media/full");
  }

  /**
   * Get a fully composed datatype by ID, including all field definitions.
   *
   * @param id - ULID of the datatype to fetch.
   * @returns The fully composed datatype view.
   * @throws {@link ModulaError} On non-2xx response (e.g. 404 if not found).
   */
  async getDatatypeFull(id: string): Promise<DatatypeFullView> {
    return this.request<DatatypeFullView>("/api/v1/datatype/full", { q: id });
  }

  /**
   * List all datatypes with full field definitions.
   *
   * @returns All datatypes with their full field definitions.
   * @throws {@link ModulaError} On non-2xx response.
   */
  async listDatatypesFull(): Promise<DatatypeFullListItem[]> {
    return this.request<DatatypeFullListItem[]>("/api/v1/datatype/full/list");
  }
}
