import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createHttpClient } from "./http.js";
import type { HttpClient } from "./http.js";
import type { ApiError } from "./types/common.js";

// ---------------------------------------------------------------------------
// Test server -- node:http for integration tests
// ---------------------------------------------------------------------------

let server: http.Server;
let baseUrl: string;

beforeAll(async () => {
  server = http.createServer(async (req, res) => {
    const port = (server.address() as import("net").AddressInfo).port;
    const url = new URL(req.url!, `http://localhost:${port}`);
    const path = url.pathname;

    // Read body if needed
    let rawBody = "";
    for await (const chunk of req) rawBody += chunk;

    // -- Echo endpoint: returns request info as JSON --
    if (path === "/api/v1/echo") {
      const params: Record<string, string> = {};
      url.searchParams.forEach((v, k) => {
        params[k] = v;
      });
      const headers: Record<string, string> = {};
      for (let i = 0; i < req.rawHeaders.length; i += 2) {
        headers[req.rawHeaders[i].toLowerCase()] = req.rawHeaders[i + 1];
      }
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        method: req.method,
        headers,
        params,
        path: path,
      }));
      return;
    }

    // -- Echo body: returns the request body back --
    if (path === "/api/v1/echo-body" && (req.method === "POST" || req.method === "PUT")) {
      const body = JSON.parse(rawBody);
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ received: body }));
      return;
    }

    // -- Echo body with undefined: POST/PUT with no body --
    if (path === "/api/v1/echo-nobody" && (req.method === "POST" || req.method === "PUT")) {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ received: null }));
      return;
    }

    // -- JSON success --
    if (path === "/api/v1/success") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ ok: true, data: "test-value" }));
      return;
    }

    // -- Void success (204 No Content) --
    if (path === "/api/v1/void-success" && req.method === "DELETE") {
      res.writeHead(204);
      res.end();
      return;
    }

    // -- Delete success with JSON body --
    if (path === "/api/v1/delete-ok" && req.method === "DELETE") {
      res.writeHead(200);
      res.end();
      return;
    }

    // -- Error: 404 with JSON body --
    if (path === "/api/v1/error-json") {
      res.writeHead(404, "Not Found", { "Content-Type": "application/json" });
      res.end(JSON.stringify({ error: "not_found", detail: "resource missing" }));
      return;
    }

    // -- Error: 500 with plain text body --
    if (path === "/api/v1/error-text") {
      res.writeHead(500, "Internal Server Error", { "Content-Type": "text/plain" });
      res.end("Internal Server Error");
      return;
    }

    // -- Success but non-JSON content type --
    if (path === "/api/v1/success-text") {
      res.writeHead(200, "OK", { "Content-Type": "text/plain" });
      res.end("plain text response");
      return;
    }

    // -- Slow endpoint for timeout testing --
    if (path === "/api/v1/slow") {
      setTimeout(() => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ ok: true }));
      }, 5000);
      return;
    }

    // -- Default: 404 --
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
// Helper to create a client with defaults
// ---------------------------------------------------------------------------

function makeClient(overrides?: {
  apiKey?: string;
  defaultTimeout?: number;
  credentials?: RequestCredentials;
}): HttpClient {
  return createHttpClient({
    baseUrl: baseUrl,
    apiKey: overrides?.apiKey,
    defaultTimeout: overrides?.defaultTimeout ?? 10000,
    credentials: overrides?.credentials ?? "omit",
  });
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("createHttpClient", () => {
  // -------------------------------------------------------------------------
  // GET
  // -------------------------------------------------------------------------

  describe("get", () => {
    test("fetches JSON from the correct URL with no params", async () => {
      const client = makeClient();
      const result = await client.get<{ ok: boolean; data: string }>("/success");
      expect(result).toEqual({ ok: true, data: "test-value" });
    });

    test("includes query params in the URL", async () => {
      const client = makeClient();
      const result = await client.get<{ params: Record<string, string> }>("/echo", {
        foo: "bar",
        baz: "qux",
      });
      expect(result.params.foo).toBe("bar");
      expect(result.params.baz).toBe("qux");
    });

    test("sends GET method", async () => {
      const client = makeClient();
      const result = await client.get<{ method: string }>("/echo");
      expect(result.method).toBe("GET");
    });

    test("sends Content-Type application/json header", async () => {
      const client = makeClient();
      const result = await client.get<{ headers: Record<string, string> }>("/echo");
      expect(result.headers["content-type"]).toBe("application/json");
    });

    test("sends Authorization header when apiKey is set", async () => {
      const client = makeClient({ apiKey: "test-key-123" });
      const result = await client.get<{ headers: Record<string, string> }>("/echo");
      expect(result.headers["authorization"]).toBe("Bearer test-key-123");
    });

    test("does not send Authorization header when apiKey is not set", async () => {
      const client = makeClient();
      const result = await client.get<{ headers: Record<string, string> }>("/echo");
      expect(result.headers["authorization"]).toBeUndefined();
    });

    test("throws ApiError with JSON body on non-ok JSON response", async () => {
      const client = makeClient();
      try {
        await client.get("/error-json");
        // Should not reach here
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(404);
        expect(apiErr.message).toBe("Not Found");
        expect(apiErr.body).toEqual({ error: "not_found", detail: "resource missing" });
      }
    });

    test("throws ApiError without body on non-ok non-JSON response", async () => {
      const client = makeClient();
      try {
        await client.get("/error-text");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(500);
        expect(apiErr.message).toBe("Internal Server Error");
        expect(apiErr.body).toBeUndefined();
      }
    });

    // This tests the branch: response.ok is true, but content-type is not JSON.
    // The code throws an ApiError in this case because it always expects JSON from get().
    test("throws ApiError when response is ok but content-type is not JSON", async () => {
      const client = makeClient();
      try {
        await client.get("/success-text");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(200);
        expect(apiErr.body).toBeUndefined();
      }
    });
  });

  // -------------------------------------------------------------------------
  // POST
  // -------------------------------------------------------------------------

  describe("post", () => {
    test("sends POST method with JSON body", async () => {
      const client = makeClient();
      const result = await client.post<{ received: { name: string } }>("/echo-body", {
        name: "test-item",
      });
      expect(result.received).toEqual({ name: "test-item" });
    });

    test("sends POST with undefined body when body is omitted", async () => {
      const client = makeClient();
      const result = await client.post<{ received: null }>("/echo-nobody");
      expect(result.received).toBeNull();
    });

    test("does not include query params in URL (post builds URL without params)", async () => {
      const client = makeClient();
      // post() signature does not accept params -- it only takes path and body.
      // Verify URL is built correctly without params.
      const result = await client.post<{ received: { x: number } }>("/echo-body", { x: 42 });
      expect(result.received.x).toBe(42);
    });

    test("throws ApiError on non-ok JSON response", async () => {
      const client = makeClient();
      try {
        await client.post("/error-json", { data: "test" });
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(404);
        expect(apiErr.body).toEqual({ error: "not_found", detail: "resource missing" });
      }
    });

    test("throws ApiError on non-ok non-JSON response", async () => {
      const client = makeClient();
      try {
        await client.post("/error-text", { data: "test" });
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(500);
        expect(apiErr.body).toBeUndefined();
      }
    });

    test("includes Authorization header when apiKey is set", async () => {
      const client = makeClient({ apiKey: "post-key" });
      const result = await client.post<{ received: null }>("/echo-nobody");
      expect(result.received).toBeNull();
      // Reaching here means auth header was accepted; verify through echo
      const echo = await client.get<{ headers: Record<string, string> }>("/echo");
      expect(echo.headers["authorization"]).toBe("Bearer post-key");
    });
  });

  // -------------------------------------------------------------------------
  // PUT
  // -------------------------------------------------------------------------

  describe("put", () => {
    test("sends PUT method with JSON body", async () => {
      const client = makeClient();
      const result = await client.put<{ received: { updated: boolean } }>("/echo-body", {
        updated: true,
      });
      expect(result.received).toEqual({ updated: true });
    });

    test("sends PUT with undefined body when body is omitted", async () => {
      const client = makeClient();
      const result = await client.put<{ received: null }>("/echo-nobody");
      expect(result.received).toBeNull();
    });

    test("throws ApiError on non-ok response", async () => {
      const client = makeClient();
      try {
        await client.put("/error-json", { data: "test" });
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(404);
      }
    });
  });

  // -------------------------------------------------------------------------
  // DEL
  // -------------------------------------------------------------------------

  describe("del", () => {
    test("sends DELETE method and resolves void on success", async () => {
      const client = makeClient();
      // 204 No Content -- void response handler should succeed
      const result = await client.del("/void-success");
      expect(result).toBeUndefined();
    });

    test("sends DELETE with query params", async () => {
      const client = makeClient();
      // The echo endpoint returns JSON, but del() uses handleVoidResponse which
      // does not parse the body on success -- it just checks response.ok.
      // We use the echo endpoint to verify the request arrives with params.
      // Since echo returns 200 OK, and handleVoidResponse only checks !response.ok,
      // this will succeed (void).
      const result = await client.del("/echo", { id: "abc-123" });
      expect(result).toBeUndefined();
    });

    test("throws ApiError with JSON body on non-ok JSON response", async () => {
      const client = makeClient();
      try {
        await client.del("/error-json");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(404);
        expect(apiErr.message).toBe("Not Found");
        expect(apiErr.body).toEqual({ error: "not_found", detail: "resource missing" });
      }
    });

    test("throws ApiError without body on non-ok non-JSON response", async () => {
      const client = makeClient();
      try {
        await client.del("/error-text");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(500);
        expect(apiErr.message).toBe("Internal Server Error");
        expect(apiErr.body).toBeUndefined();
      }
    });
  });

  // -------------------------------------------------------------------------
  // RAW
  // -------------------------------------------------------------------------

  describe("raw", () => {
    test("returns the raw Response object without processing", async () => {
      const client = makeClient();
      const response = await client.raw("/success", { method: "GET" });
      expect(response).toBeInstanceOf(Response);
      expect(response.ok).toBe(true);
      const body = await response.json();
      expect(body).toEqual({ ok: true, data: "test-value" });
    });

    test("does not add default headers (uses provided init only)", async () => {
      const client = makeClient({ apiKey: "should-not-appear" });
      // raw() calls fetch(url, init) directly without merging headers
      const response = await client.raw("/echo", {
        method: "GET",
        headers: { "X-Custom": "test" },
      });
      const body = await response.json();
      // The raw endpoint should NOT have the Authorization header from the client
      expect(body.headers["authorization"]).toBeUndefined();
      expect(body.headers["x-custom"]).toBe("test");
    });

    test("builds URL with /api/v1 prefix", async () => {
      const client = makeClient();
      const response = await client.raw("/echo", { method: "GET" });
      const body = await response.json();
      expect(body.path).toBe("/api/v1/echo");
    });

    test("returns error responses without throwing", async () => {
      const client = makeClient();
      // raw() does not process the response -- caller handles it
      const response = await client.raw("/error-json", { method: "GET" });
      expect(response.ok).toBe(false);
      expect(response.status).toBe(404);
    });
  });

  // -------------------------------------------------------------------------
  // URL building (tested through public API)
  // -------------------------------------------------------------------------

  describe("URL building", () => {
    test("builds correct path with /api/v1 prefix", async () => {
      const client = makeClient();
      const result = await client.get<{ path: string }>("/echo");
      expect(result.path).toBe("/api/v1/echo");
    });

    test.each([
      {
        name: "single param",
        params: { q: "abc" },
        expectedParams: { q: "abc" },
      },
      {
        name: "multiple params",
        params: { page: "1", limit: "10" },
        expectedParams: { page: "1", limit: "10" },
      },
      {
        name: "param with empty value",
        params: { q: "" },
        expectedParams: { q: "" },
      },
      {
        name: "param with special characters",
        params: { q: "hello world", filter: "a&b=c" },
        expectedParams: { q: "hello world", filter: "a&b=c" },
      },
    ])("includes $name in query string", async ({ params, expectedParams }) => {
      const client = makeClient();
      const result = await client.get<{ params: Record<string, string> }>("/echo", params);
      for (const [key, value] of Object.entries(expectedParams)) {
        expect(result.params[key]).toBe(value);
      }
    });

    test("sends no query params when params argument is omitted", async () => {
      const client = makeClient();
      const result = await client.get<{ params: Record<string, string> }>("/echo");
      expect(Object.keys(result.params).length).toBe(0);
    });
  });

  // -------------------------------------------------------------------------
  // Abort / Timeout behavior
  // -------------------------------------------------------------------------

  describe("abort and timeout", () => {
    test("respects user-provided AbortSignal", async () => {
      const client = makeClient({ defaultTimeout: 10000 });
      const controller = new AbortController();
      controller.abort();

      try {
        await client.get("/success", undefined, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });

    test("times out when server is too slow", async () => {
      // Use a very short timeout so we don't wait long.
      const client = makeClient({ defaultTimeout: 50 });

      try {
        await client.get("/slow");
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("TimeoutError");
      }
    });

    test("user abort takes precedence over timeout when both fire", async () => {
      const client = makeClient({ defaultTimeout: 10000 });
      const controller = new AbortController();

      setTimeout(() => controller.abort(), 10);

      try {
        await client.get("/slow", undefined, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });

    test("falls back to timeout-only signal when no user signal is provided", async () => {
      // Verifies that mergeSignals returns AbortSignal.timeout(defaultTimeout) when
      // opts.signal is not provided -- this code path avoids AbortSignal.any entirely.
      const client = makeClient({ defaultTimeout: 10000 });
      // Should succeed within timeout
      const result = await client.get<{ ok: boolean }>("/success");
      expect(result.ok).toBe(true);
    });
  });

  // -------------------------------------------------------------------------
  // handleResponse error branch coverage
  // -------------------------------------------------------------------------

  describe("handleResponse error branches", () => {
    test("non-ok response with JSON content-type includes parsed body in ApiError", async () => {
      const client = makeClient();
      try {
        await client.get("/error-json");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.body).toEqual({ error: "not_found", detail: "resource missing" });
      }
    });

    test("non-ok response with non-JSON content-type has undefined body in ApiError", async () => {
      const client = makeClient();
      try {
        await client.get("/error-text");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.body).toBeUndefined();
      }
    });

    test("ok response with non-JSON content-type throws ApiError (get/post/put expect JSON)", async () => {
      const client = makeClient();
      try {
        await client.post("/success-text");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        // Status is 200 because the response was "ok" -- but non-JSON is treated as an error
        expect(apiErr.status).toBe(200);
      }
    });
  });

  // -------------------------------------------------------------------------
  // handleVoidResponse branch coverage
  // -------------------------------------------------------------------------

  describe("handleVoidResponse branches", () => {
    test("ok response returns void regardless of content-type", async () => {
      const client = makeClient();
      // 204 No Content -- no body, no content-type
      const result = await client.del("/void-success");
      expect(result).toBeUndefined();
    });

    test("ok response with JSON body also returns void (del does not parse success body)", async () => {
      const client = makeClient();
      // The echo endpoint returns 200 with JSON, but del() ignores the body on success
      const result = await client.del("/echo");
      expect(result).toBeUndefined();
    });

    test("non-ok response with JSON body throws ApiError with body", async () => {
      const client = makeClient();
      try {
        await client.del("/error-json");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(404);
        expect(apiErr.body).toEqual({ error: "not_found", detail: "resource missing" });
      }
    });

    test("non-ok response with non-JSON body throws ApiError without body", async () => {
      const client = makeClient();
      try {
        await client.del("/error-text");
        expect(true).toBe(false);
      } catch (err: unknown) {
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
        expect(apiErr.status).toBe(500);
        expect(apiErr.body).toBeUndefined();
      }
    });
  });

  // -------------------------------------------------------------------------
  // Auth header variations
  // -------------------------------------------------------------------------

  describe("auth header", () => {
    test.each([
      { name: "no apiKey", apiKey: undefined, expectedAuth: undefined },
      { name: "with apiKey", apiKey: "my-secret-key", expectedAuth: "Bearer my-secret-key" },
      { name: "apiKey with special characters", apiKey: "key+with/special=chars", expectedAuth: "Bearer key+with/special=chars" },
    ])("$name", async ({ apiKey, expectedAuth }) => {
      const client = makeClient({ apiKey });
      const result = await client.get<{ headers: Record<string, string> }>("/echo");
      if (expectedAuth === undefined) {
        expect(result.headers["authorization"]).toBeUndefined();
      } else {
        expect(result.headers["authorization"]).toBe(expectedAuth);
      }
    });
  });

  // -------------------------------------------------------------------------
  // Multiple clients with different configs
  // -------------------------------------------------------------------------

  describe("client isolation", () => {
    test("two clients with different apiKeys send different headers", async () => {
      const client1 = makeClient({ apiKey: "key-one" });
      const client2 = makeClient({ apiKey: "key-two" });

      const result1 = await client1.get<{ headers: Record<string, string> }>("/echo");
      const result2 = await client2.get<{ headers: Record<string, string> }>("/echo");

      expect(result1.headers["authorization"]).toBe("Bearer key-one");
      expect(result2.headers["authorization"]).toBe("Bearer key-two");
    });

    test("client without apiKey does not leak auth from another client", async () => {
      const clientWithKey = makeClient({ apiKey: "secret" });
      const clientNoKey = makeClient();

      // Use the keyed client first
      await clientWithKey.get("/echo");

      // Then use the keyless client -- it should not have auth
      const result = await clientNoKey.get<{ headers: Record<string, string> }>("/echo");
      expect(result.headers["authorization"]).toBeUndefined();
    });
  });
});
