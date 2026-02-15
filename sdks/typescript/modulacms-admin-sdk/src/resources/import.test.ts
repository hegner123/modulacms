import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createImportResource } from "./import.js";
import type { ImportResource } from "./import.js";
import { createHttpClient } from "../http.js";
import type { HttpClient } from "../http.js";
import type { ApiError } from "../types/common.js";
import type { ImportFormat, ImportResponse } from "../types/import.js";

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const fakeImportResponse: ImportResponse = {
  success: true,
  datatypes_created: 3,
  fields_created: 12,
  content_created: 47,
  message: "Import completed successfully",
  errors: [],
};

const fakeImportData: Record<string, unknown> = {
  entries: [
    { title: "First post", body: "Hello world" },
    { title: "Second post", body: "Goodbye world" },
  ],
  content_types: ["blog", "page"],
};

// ---------------------------------------------------------------------------
// Test server -- real node:http server (matching project convention)
// ---------------------------------------------------------------------------

let server: http.Server;
let baseUrl: string;

let lastRequest: {
  method: string;
  path: string;
  query: string;
  body: unknown;
  headers: Record<string, string>;
} | null = null;

beforeAll(async () => {
  server = http.createServer(async (req, res) => {
    const port = (server.address() as import("net").AddressInfo).port;
    const url = new URL(req.url!, `http://localhost:${port}`);
    const path = url.pathname;
    const query = url.search;

    let body: unknown = null;
    if (req.method === "POST") {
      const contentType = req.headers["content-type"];
      if (contentType && contentType.includes("application/json")) {
        try {
          let rawBody = "";
          for await (const chunk of req) rawBody += chunk;
          body = JSON.parse(rawBody);
        } catch {
          body = null;
        }
      }
    }

    lastRequest = {
      method: req.method!,
      path,
      query,
      body,
      headers: Object.fromEntries(
        Object.entries(req.headers).map(([k, v]) => [k, Array.isArray(v) ? v.join(", ") : v ?? ""])
      ) as Record<string, string>,
    };

    // -- format-specific import endpoints --
    if (path === "/api/v1/import/contentful" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeImportResponse));
      return;
    }
    if (path === "/api/v1/import/sanity" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeImportResponse));
      return;
    }
    if (path === "/api/v1/import/strapi" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeImportResponse));
      return;
    }
    if (path === "/api/v1/import/wordpress" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeImportResponse));
      return;
    }
    if (path === "/api/v1/import/clean" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeImportResponse));
      return;
    }

    // -- bulk import endpoint (query string has format param) --
    if (path === "/api/v1/import" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeImportResponse));
      return;
    }

    // -- slow endpoint for abort/timeout tests --
    if (path === "/api/v1/import/slow" && req.method === "POST") {
      setTimeout(() => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(fakeImportResponse));
      }, 5000);
      return;
    }

    // Fallback
    res.writeHead(404);
    res.end("Not Found");
  });
  await new Promise<void>(resolve => server.listen(0, resolve));
  const port = (server.address() as import("net").AddressInfo).port;
  baseUrl = `http://localhost:${port}`;
});

afterAll(async () => {
  await new Promise<void>(resolve => server.close(() => resolve()));
});

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeClient(overrides?: {
  apiKey?: string;
  defaultTimeout?: number;
}): HttpClient {
  return createHttpClient({
    baseUrl,
    apiKey: overrides?.apiKey,
    defaultTimeout: overrides?.defaultTimeout ?? 10000,
    credentials: "omit",
  });
}

function makeImport(overrides?: {
  apiKey?: string;
  defaultTimeout?: number;
}): ImportResource {
  return createImportResource(makeClient(overrides));
}

// ---------------------------------------------------------------------------
// Inline server helper
// ---------------------------------------------------------------------------

async function createInlineServer(
  handler: (req: http.IncomingMessage, res: http.ServerResponse) => void,
): Promise<{ server: http.Server; port: number }> {
  const srv = http.createServer(handler);
  await new Promise<void>(resolve => srv.listen(0, resolve));
  const port = (srv.address() as import("net").AddressInfo).port;
  return { server: srv, port };
}

async function closeServer(srv: http.Server): Promise<void> {
  await new Promise<void>(resolve => srv.close(() => resolve()));
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("createImportResource", () => {
  // -------------------------------------------------------------------------
  // Factory
  // -------------------------------------------------------------------------

  describe("factory", () => {
    test("returns an object with all six import methods", () => {
      const imp = makeImport();
      expect(typeof imp.contentful).toBe("function");
      expect(typeof imp.sanity).toBe("function");
      expect(typeof imp.strapi).toBe("function");
      expect(typeof imp.wordpress).toBe("function");
      expect(typeof imp.clean).toBe("function");
      expect(typeof imp.bulk).toBe("function");
    });

    test("all methods return promises", () => {
      const imp = makeImport();

      const contentfulResult = imp.contentful(fakeImportData);
      expect(typeof contentfulResult.then).toBe("function");
      contentfulResult.catch(() => {});

      const sanityResult = imp.sanity(fakeImportData);
      expect(typeof sanityResult.then).toBe("function");
      sanityResult.catch(() => {});

      const strapiResult = imp.strapi(fakeImportData);
      expect(typeof strapiResult.then).toBe("function");
      strapiResult.catch(() => {});

      const wordpressResult = imp.wordpress(fakeImportData);
      expect(typeof wordpressResult.then).toBe("function");
      wordpressResult.catch(() => {});

      const cleanResult = imp.clean(fakeImportData);
      expect(typeof cleanResult.then).toBe("function");
      cleanResult.catch(() => {});

      const bulkResult = imp.bulk("contentful", fakeImportData);
      expect(typeof bulkResult.then).toBe("function");
      bulkResult.catch(() => {});
    });
  });

  // -------------------------------------------------------------------------
  // HTTP method and URL path correctness (parameterized)
  // -------------------------------------------------------------------------

  describe("HTTP method and URL path correctness", () => {
    test.each([
      { name: "contentful", expectedPath: "/api/v1/import/contentful" },
      { name: "sanity", expectedPath: "/api/v1/import/sanity" },
      { name: "strapi", expectedPath: "/api/v1/import/strapi" },
      { name: "wordpress", expectedPath: "/api/v1/import/wordpress" },
      { name: "clean", expectedPath: "/api/v1/import/clean" },
    ])("$name sends POST to $expectedPath", async ({ name, expectedPath }) => {
      const imp = makeImport();

      switch (name) {
        case "contentful":
          await imp.contentful(fakeImportData);
          break;
        case "sanity":
          await imp.sanity(fakeImportData);
          break;
        case "strapi":
          await imp.strapi(fakeImportData);
          break;
        case "wordpress":
          await imp.wordpress(fakeImportData);
          break;
        case "clean":
          await imp.clean(fakeImportData);
          break;
      }

      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe(expectedPath);
    });
  });

  // -------------------------------------------------------------------------
  // Body forwarding (parameterized)
  // -------------------------------------------------------------------------

  describe("body forwarding", () => {
    test.each([
      { name: "contentful" },
      { name: "sanity" },
      { name: "strapi" },
      { name: "wordpress" },
      { name: "clean" },
    ])("$name forwards data as JSON body", async ({ name }) => {
      const imp = makeImport();
      const testData = { source: name, items: [1, 2, 3] };

      switch (name) {
        case "contentful":
          await imp.contentful(testData);
          break;
        case "sanity":
          await imp.sanity(testData);
          break;
        case "strapi":
          await imp.strapi(testData);
          break;
        case "wordpress":
          await imp.wordpress(testData);
          break;
        case "clean":
          await imp.clean(testData);
          break;
      }

      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({ source: name, items: [1, 2, 3] });
    });

    test("bulk forwards data as JSON body", async () => {
      const imp = makeImport();
      const testData = { entries: [{ id: "abc" }], mapping: true };
      await imp.bulk("sanity", testData);

      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({ entries: [{ id: "abc" }], mapping: true });
    });
  });

  // -------------------------------------------------------------------------
  // ImportResponse return shape
  // -------------------------------------------------------------------------

  describe("ImportResponse return shape", () => {
    test.each([
      { name: "contentful", action: (imp: ImportResource) => imp.contentful(fakeImportData) },
      { name: "sanity", action: (imp: ImportResource) => imp.sanity(fakeImportData) },
      { name: "strapi", action: (imp: ImportResource) => imp.strapi(fakeImportData) },
      { name: "wordpress", action: (imp: ImportResource) => imp.wordpress(fakeImportData) },
      { name: "clean", action: (imp: ImportResource) => imp.clean(fakeImportData) },
      { name: "bulk", action: (imp: ImportResource) => imp.bulk("contentful", fakeImportData) },
    ])("$name returns ImportResponse with all expected fields", async ({ action }) => {
      const imp = makeImport();
      const result = await action(imp);

      expect(result).toEqual(fakeImportResponse);
      expect(result.success).toBe(true);
      expect(result.datatypes_created).toBe(3);
      expect(result.fields_created).toBe(12);
      expect(result.content_created).toBe(47);
      expect(result.message).toBe("Import completed successfully");
      expect(result.errors).toEqual([]);
    });
  });

  // -------------------------------------------------------------------------
  // Bulk: query string format param
  // -------------------------------------------------------------------------

  describe("bulk query string", () => {
    test.each([
      { format: "contentful" as ImportFormat, expectedQuery: "?format=contentful" },
      { format: "sanity" as ImportFormat, expectedQuery: "?format=sanity" },
      { format: "strapi" as ImportFormat, expectedQuery: "?format=strapi" },
      { format: "wordpress" as ImportFormat, expectedQuery: "?format=wordpress" },
      { format: "clean" as ImportFormat, expectedQuery: "?format=clean" },
    ])("bulk($format) sends POST to /api/v1/import with query $expectedQuery", async ({ format, expectedQuery }) => {
      const imp = makeImport();
      await imp.bulk(format, fakeImportData);

      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/import");
      expect(lastRequest!.query).toBe(expectedQuery);
    });

    test("bulk sends format as query parameter, not in the path", async () => {
      const imp = makeImport();
      await imp.bulk("wordpress", fakeImportData);

      expect(lastRequest).not.toBeNull();
      // Path should be /api/v1/import (no format in path)
      expect(lastRequest!.path).toBe("/api/v1/import");
      // Query should contain format=wordpress
      expect(lastRequest!.query).toContain("format=wordpress");
    });
  });

  // -------------------------------------------------------------------------
  // Error propagation: JSON error responses
  // -------------------------------------------------------------------------

  describe("error propagation: JSON error response", () => {
    test.each([
      { name: "contentful", action: (imp: ImportResource) => imp.contentful(fakeImportData) },
      { name: "sanity", action: (imp: ImportResource) => imp.sanity(fakeImportData) },
      { name: "strapi", action: (imp: ImportResource) => imp.strapi(fakeImportData) },
      { name: "wordpress", action: (imp: ImportResource) => imp.wordpress(fakeImportData) },
      { name: "clean", action: (imp: ImportResource) => imp.clean(fakeImportData) },
      { name: "bulk", action: (imp: ImportResource) => imp.bulk("contentful", fakeImportData) },
    ])("$name throws ApiError with JSON body on 403 response", async ({ action }) => {
      const { server: errorServer, port: errorPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        res.writeHead(403, "Forbidden", { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "forbidden", detail: "access denied" }));
      });

      try {
        const errHttp = createHttpClient({
          baseUrl: `http://localhost:${errorPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const imp = createImportResource(errHttp);

        try {
          await action(imp);
          expect(true).toBe(false);
        } catch (err: unknown) {
          const apiErr = err as ApiError;
          expect(apiErr._tag).toBe("ApiError");
          expect(apiErr.status).toBe(403);
          expect(apiErr.message).toBe("Forbidden");
          expect(apiErr.body).toEqual({ error: "forbidden", detail: "access denied" });
        }
      } finally {
        await closeServer(errorServer);
      }
    });
  });

  // -------------------------------------------------------------------------
  // Error propagation: non-JSON error responses
  // -------------------------------------------------------------------------

  describe("error propagation: non-JSON error response", () => {
    test.each([
      { name: "contentful", action: (imp: ImportResource) => imp.contentful(fakeImportData) },
      { name: "sanity", action: (imp: ImportResource) => imp.sanity(fakeImportData) },
      { name: "strapi", action: (imp: ImportResource) => imp.strapi(fakeImportData) },
      { name: "wordpress", action: (imp: ImportResource) => imp.wordpress(fakeImportData) },
      { name: "clean", action: (imp: ImportResource) => imp.clean(fakeImportData) },
      { name: "bulk", action: (imp: ImportResource) => imp.bulk("sanity", fakeImportData) },
    ])("$name throws ApiError with undefined body on 500 text response", async ({ action }) => {
      const { server: errorServer, port: errorPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        res.writeHead(500, "Internal Server Error", { "Content-Type": "text/plain" });
        res.end("Internal Server Error");
      });

      try {
        const errHttp = createHttpClient({
          baseUrl: `http://localhost:${errorPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const imp = createImportResource(errHttp);

        try {
          await action(imp);
          expect(true).toBe(false);
        } catch (err: unknown) {
          const apiErr = err as ApiError;
          expect(apiErr._tag).toBe("ApiError");
          expect(apiErr.status).toBe(500);
          expect(apiErr.message).toBe("Internal Server Error");
          expect(apiErr.body).toBeUndefined();
        }
      } finally {
        await closeServer(errorServer);
      }
    });
  });

  // -------------------------------------------------------------------------
  // Error propagation: various HTTP status codes
  // -------------------------------------------------------------------------

  describe("error propagation: status codes", () => {
    test.each([
      { status: 400, statusText: "Bad Request" },
      { status: 401, statusText: "Unauthorized" },
      { status: 404, statusText: "Not Found" },
      { status: 422, statusText: "Unprocessable Entity" },
      { status: 429, statusText: "Too Many Requests" },
      { status: 503, statusText: "Service Unavailable" },
    ])("contentful propagates $status $statusText", async ({ status, statusText }) => {
      const { server: errorServer, port: errorPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        res.writeHead(status, statusText, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "test_error" }));
      });

      try {
        const errHttp = createHttpClient({
          baseUrl: `http://localhost:${errorPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const imp = createImportResource(errHttp);

        try {
          await imp.contentful(fakeImportData);
          expect(true).toBe(false);
        } catch (err: unknown) {
          const apiErr = err as ApiError;
          expect(apiErr._tag).toBe("ApiError");
          expect(apiErr.status).toBe(status);
          expect(apiErr.message).toBe(statusText);
        }
      } finally {
        await closeServer(errorServer);
      }
    });
  });

  // -------------------------------------------------------------------------
  // Signal forwarding: pre-aborted signal
  // -------------------------------------------------------------------------

  describe("signal forwarding: pre-aborted", () => {
    test.each([
      { name: "contentful", action: (imp: ImportResource, opts: { signal: AbortSignal }) => imp.contentful(fakeImportData, opts) },
      { name: "sanity", action: (imp: ImportResource, opts: { signal: AbortSignal }) => imp.sanity(fakeImportData, opts) },
      { name: "strapi", action: (imp: ImportResource, opts: { signal: AbortSignal }) => imp.strapi(fakeImportData, opts) },
      { name: "wordpress", action: (imp: ImportResource, opts: { signal: AbortSignal }) => imp.wordpress(fakeImportData, opts) },
      { name: "clean", action: (imp: ImportResource, opts: { signal: AbortSignal }) => imp.clean(fakeImportData, opts) },
      { name: "bulk", action: (imp: ImportResource, opts: { signal: AbortSignal }) => imp.bulk("contentful", fakeImportData, opts) },
    ])("$name aborts immediately with pre-aborted signal", async ({ action }) => {
      const imp = makeImport();
      const controller = new AbortController();
      controller.abort();

      try {
        await action(imp, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // Signal forwarding: user aborts during in-flight request
  // -------------------------------------------------------------------------

  describe("signal forwarding: user aborts during request", () => {
    test("contentful aborts when user signal fires mid-request", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        setTimeout(() => {
          res.writeHead(200, { "Content-Type": "application/json" });
          res.end(JSON.stringify(fakeImportResponse));
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const imp = createImportResource(slowHttp);
        const controller = new AbortController();
        setTimeout(() => controller.abort(), 50);

        try {
          await imp.contentful(fakeImportData, { signal: controller.signal });
          expect(true).toBe(false);
        } catch (err: unknown) {
          expect(err).toBeInstanceOf(Error);
          const error = err as Error;
          expect(error.name).toBe("AbortError");
        }
      } finally {
        await closeServer(slowServer);
      }
    });

    test("bulk aborts when user signal fires mid-request", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        setTimeout(() => {
          res.writeHead(200, { "Content-Type": "application/json" });
          res.end(JSON.stringify(fakeImportResponse));
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const imp = createImportResource(slowHttp);
        const controller = new AbortController();
        setTimeout(() => controller.abort(), 50);

        try {
          await imp.bulk("strapi", fakeImportData, { signal: controller.signal });
          expect(true).toBe(false);
        } catch (err: unknown) {
          expect(err).toBeInstanceOf(Error);
          const error = err as Error;
          expect(error.name).toBe("AbortError");
        }
      } finally {
        await closeServer(slowServer);
      }
    });
  });

  // -------------------------------------------------------------------------
  // Timeout via short defaultTimeout
  // -------------------------------------------------------------------------

  describe("timeout behavior", () => {
    test("times out when server is too slow and no user signal provided", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        setTimeout(() => {
          res.writeHead(200, { "Content-Type": "application/json" });
          res.end(JSON.stringify(fakeImportResponse));
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 50,
          credentials: "omit",
        });
        const imp = createImportResource(slowHttp);

        try {
          await imp.contentful(fakeImportData);
          expect(true).toBe(false);
        } catch (err: unknown) {
          expect(err).toBeInstanceOf(Error);
          const error = err as Error;
          expect(error.name).toBe("TimeoutError");
        }
      } finally {
        await closeServer(slowServer);
      }
    });
  });

  // -------------------------------------------------------------------------
  // Authorization header passthrough
  // -------------------------------------------------------------------------

  describe("authorization header", () => {
    test("includes Bearer token when apiKey is provided", async () => {
      const imp = makeImport({ apiKey: "import-api-key-123" });
      await imp.contentful(fakeImportData);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer import-api-key-123");
    });

    test("does not include Authorization header when apiKey is absent", async () => {
      const imp = makeImport();
      await imp.contentful(fakeImportData);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });

    test.each([
      { name: "contentful", action: (imp: ImportResource) => imp.contentful(fakeImportData) },
      { name: "sanity", action: (imp: ImportResource) => imp.sanity(fakeImportData) },
      { name: "strapi", action: (imp: ImportResource) => imp.strapi(fakeImportData) },
      { name: "wordpress", action: (imp: ImportResource) => imp.wordpress(fakeImportData) },
      { name: "clean", action: (imp: ImportResource) => imp.clean(fakeImportData) },
      { name: "bulk", action: (imp: ImportResource) => imp.bulk("contentful", fakeImportData) },
    ])("$name sends Bearer token when apiKey is set", async ({ action }) => {
      const imp = makeImport({ apiKey: "shared-import-key-456" });
      await action(imp);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer shared-import-key-456");
    });
  });

  // -------------------------------------------------------------------------
  // Content-Type header
  // -------------------------------------------------------------------------

  describe("content-type header", () => {
    test.each([
      { name: "contentful", action: (imp: ImportResource) => imp.contentful(fakeImportData) },
      { name: "sanity", action: (imp: ImportResource) => imp.sanity(fakeImportData) },
      { name: "strapi", action: (imp: ImportResource) => imp.strapi(fakeImportData) },
      { name: "wordpress", action: (imp: ImportResource) => imp.wordpress(fakeImportData) },
      { name: "clean", action: (imp: ImportResource) => imp.clean(fakeImportData) },
      { name: "bulk", action: (imp: ImportResource) => imp.bulk("wordpress", fakeImportData) },
    ])("$name sends Content-Type: application/json", async ({ action }) => {
      const imp = makeImport();
      await action(imp);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["content-type"]).toBe("application/json");
    });
  });

  // -------------------------------------------------------------------------
  // Resource isolation
  // -------------------------------------------------------------------------

  describe("resource isolation", () => {
    test("two import resources from different HttpClients are independent", async () => {
      const imp1 = makeImport({ apiKey: "key-alpha" });
      const imp2 = makeImport({ apiKey: "key-beta" });

      await imp1.contentful(fakeImportData);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer key-alpha");

      await imp2.contentful(fakeImportData);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer key-beta");
    });

    test("import resource without apiKey does not leak auth from another", async () => {
      const authed = makeImport({ apiKey: "secret-key" });
      const unauthed = makeImport();

      await authed.sanity(fakeImportData);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer secret-key");

      await unauthed.sanity(fakeImportData);
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // Response shape validation (individual field checks)
  // -------------------------------------------------------------------------

  describe("response shape validation", () => {
    test("contentful response has all ImportResponse fields", async () => {
      const imp = makeImport();
      const result = await imp.contentful(fakeImportData);
      expect(result.success).toBe(true);
      expect(typeof result.datatypes_created).toBe("number");
      expect(typeof result.fields_created).toBe("number");
      expect(typeof result.content_created).toBe("number");
      expect(typeof result.message).toBe("string");
      expect(Array.isArray(result.errors)).toBe(true);
    });

    test("bulk response has all ImportResponse fields", async () => {
      const imp = makeImport();
      const result = await imp.bulk("clean", fakeImportData);
      expect(result.success).toBe(true);
      expect(typeof result.datatypes_created).toBe("number");
      expect(typeof result.fields_created).toBe("number");
      expect(typeof result.content_created).toBe("number");
      expect(typeof result.message).toBe("string");
      expect(Array.isArray(result.errors)).toBe(true);
    });

    test("import response with errors array populated", async () => {
      const responseWithErrors: ImportResponse = {
        success: false,
        datatypes_created: 1,
        fields_created: 2,
        content_created: 0,
        message: "Partial import",
        errors: ["Field type mismatch on entry 3", "Missing required field on entry 7"],
      };

      const { server: errorResponseServer, port: errorResponsePort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(responseWithErrors));
      });

      try {
        const httpClient = createHttpClient({
          baseUrl: `http://localhost:${errorResponsePort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const imp = createImportResource(httpClient);
        const result = await imp.strapi(fakeImportData);

        expect(result.success).toBe(false);
        expect(result.datatypes_created).toBe(1);
        expect(result.fields_created).toBe(2);
        expect(result.content_created).toBe(0);
        expect(result.message).toBe("Partial import");
        expect(result.errors).toEqual([
          "Field type mismatch on entry 3",
          "Missing required field on entry 7",
        ]);
      } finally {
        await closeServer(errorResponseServer);
      }
    });
  });

  // -------------------------------------------------------------------------
  // Empty data body
  // -------------------------------------------------------------------------

  describe("empty data body", () => {
    test("contentful sends empty object body when given empty object", async () => {
      const imp = makeImport();
      await imp.contentful({});
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({});
    });

    test("bulk sends empty object body when given empty object", async () => {
      const imp = makeImport();
      await imp.bulk("clean", {});
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({});
    });
  });
});
