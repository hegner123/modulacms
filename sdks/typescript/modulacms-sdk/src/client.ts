import { ModulaError } from "./errors.js";
import type { ContentFormat, ContentData, ContentField, Route, Media, MediaDimension, Datatype, Field } from "@modulacms/types";

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
  /** Optional runtime validator. When provided, the response is checked before returning. Throws {@link ModulaError} on validation failure. */
  validate?: Validator<T>;
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
   * @param slug - Route slug (e.g. "about", "blog"). Must not be empty or start with "/".
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
    if (!slug || slug.startsWith("/")) {
      throw new ModulaError(0, { error: `Invalid slug: "${slug}". Slugs should not be empty or start with "/"` });
    }
    const format = options?.format ?? this.defaultFormat;
    const params: Record<string, string> = {};
    if (format) {
      params.format = format;
    }
    return this.request<T>(`/${slug}`, params, options?.validate);
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
}
