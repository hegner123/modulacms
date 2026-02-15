import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createSessionsResource } from "./sessions.js";
import type { SessionsResource } from "./sessions.js";
import { createHttpClient } from "../http.js";
import type { HttpClient } from "../http.js";
import type { ApiError, SessionID, UserID } from "../types/common.js";
import type { UpdateSessionParams, Session } from "../types/users.js";

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const fakeSession: Session = {
  session_id: "sess-001" as SessionID,
  user_id: "u-42" as UserID,
  created_at: "2025-07-01T10:00:00Z",
  expires_at: "2025-07-02T10:00:00Z",
  last_access: "2025-07-01T12:30:00Z",
  ip_address: "192.168.1.100",
  user_agent: "TestAgent/1.0",
  session_data: '{"theme":"dark"}',
};

const updateParams: UpdateSessionParams = {
  session_id: "sess-001" as SessionID,
  user_id: "u-42" as UserID,
  created_at: "2025-07-01T10:00:00Z",
  expires_at: "2025-07-02T10:00:00Z",
  last_access: "2025-07-01T12:30:00Z",
  ip_address: "192.168.1.100",
  user_agent: "TestAgent/1.0",
  session_data: '{"theme":"dark"}',
};

const removeId = "sess-remove-99" as SessionID;

// ---------------------------------------------------------------------------
// Test server -- real node:http server (matching project convention)
// ---------------------------------------------------------------------------

let server: http.Server;
let baseUrl: string;

let lastRequest: {
  method: string;
  path: string;
  body: unknown;
  headers: Record<string, string>;
  params: Record<string, string>;
} | null = null;

beforeAll(async () => {
  server = http.createServer(async (req, res) => {
    const port = (server.address() as import("net").AddressInfo).port;
    const url = new URL(req.url!, `http://localhost:${port}`);
    const path = url.pathname;

    const params: Record<string, string> = {};
    url.searchParams.forEach((v, k) => {
      params[k] = v;
    });

    let body: unknown = null;
    if (req.method === "POST" || req.method === "PUT") {
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
      body,
      headers: Object.fromEntries(
        Object.entries(req.headers).map(([k, v]) => [k, Array.isArray(v) ? v.join(", ") : v ?? ""])
      ) as Record<string, string>,
      params,
    };

    // -- update (PUT /api/v1/sessions/) --
    if (path === "/api/v1/sessions/" && req.method === "PUT") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeSession));
      return;
    }

    // -- remove (DELETE /api/v1/sessions/) --
    if (path === "/api/v1/sessions/" && req.method === "DELETE") {
      res.writeHead(204);
      res.end();
      return;
    }

    // -- slow endpoint for abort/timeout tests --
    if (path === "/api/v1/sessions/slow") {
      setTimeout(() => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(fakeSession));
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

function makeSessions(overrides?: {
  apiKey?: string;
  defaultTimeout?: number;
}): SessionsResource {
  return createSessionsResource(makeClient(overrides));
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

describe("createSessionsResource", () => {
  // -------------------------------------------------------------------------
  // Factory
  // -------------------------------------------------------------------------

  describe("factory", () => {
    test("returns an object with update and remove methods", () => {
      const sessions = makeSessions();
      expect(typeof sessions.update).toBe("function");
      expect(typeof sessions.remove).toBe("function");
    });

    test("both methods return promises", () => {
      const sessions = makeSessions();

      const updateResult = sessions.update(updateParams);
      expect(typeof updateResult.then).toBe("function");
      updateResult.catch(() => {});

      const removeResult = sessions.remove(removeId);
      expect(typeof removeResult.then).toBe("function");
      removeResult.catch(() => {});
    });
  });

  // -------------------------------------------------------------------------
  // update
  // -------------------------------------------------------------------------

  describe("update", () => {
    test("returns parsed Session on success", async () => {
      const sessions = makeSessions();
      const result = await sessions.update(updateParams);
      expect(result).toEqual(fakeSession);
    });

    test("sends PUT to /sessions/", async () => {
      const sessions = makeSessions();
      await sessions.update(updateParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("PUT");
      expect(lastRequest!.path).toBe("/api/v1/sessions/");
    });

    test("sends UpdateSessionParams as JSON body", async () => {
      const sessions = makeSessions();
      await sessions.update(updateParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({
        session_id: "sess-001",
        user_id: "u-42",
        created_at: "2025-07-01T10:00:00Z",
        expires_at: "2025-07-02T10:00:00Z",
        last_access: "2025-07-01T12:30:00Z",
        ip_address: "192.168.1.100",
        user_agent: "TestAgent/1.0",
        session_data: '{"theme":"dark"}',
      });
    });

    test("response has all Session fields", async () => {
      const sessions = makeSessions();
      const result = await sessions.update(updateParams);
      expect(result.session_id).toBe("sess-001");
      expect(result.user_id).toBe("u-42");
      expect(result.created_at).toBe("2025-07-01T10:00:00Z");
      expect(result.expires_at).toBe("2025-07-02T10:00:00Z");
      expect(result.last_access).toBe("2025-07-01T12:30:00Z");
      expect(result.ip_address).toBe("192.168.1.100");
      expect(result.user_agent).toBe("TestAgent/1.0");
      expect(result.session_data).toBe('{"theme":"dark"}');
    });

    test("forwards RequestOptions signal", async () => {
      const sessions = makeSessions();
      const controller = new AbortController();
      controller.abort();
      try {
        await sessions.update(updateParams, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // remove
  // -------------------------------------------------------------------------

  describe("remove", () => {
    test("resolves void on success", async () => {
      const sessions = makeSessions();
      const result = await sessions.remove(removeId);
      expect(result).toBeUndefined();
    });

    test("sends DELETE to /sessions/", async () => {
      const sessions = makeSessions();
      await sessions.remove(removeId);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("DELETE");
      expect(lastRequest!.path).toBe("/api/v1/sessions/");
    });

    test("sends id as query param q", async () => {
      const sessions = makeSessions();
      await sessions.remove(removeId);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.params["q"]).toBe("sess-remove-99");
    });

    test("converts id to string via String(id)", async () => {
      const sessions = makeSessions();
      const numericId = "12345" as SessionID;
      await sessions.remove(numericId);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.params["q"]).toBe("12345");
    });

    test("forwards RequestOptions signal", async () => {
      const sessions = makeSessions();
      const controller = new AbortController();
      controller.abort();
      try {
        await sessions.remove(removeId, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // HTTP method correctness (parameterized)
  // -------------------------------------------------------------------------

  describe("HTTP method correctness", () => {
    test.each([
      { name: "update", expectedMethod: "PUT", expectedPath: "/api/v1/sessions/" },
      { name: "remove", expectedMethod: "DELETE", expectedPath: "/api/v1/sessions/" },
    ])("$name sends $expectedMethod to $expectedPath", async ({ name, expectedMethod, expectedPath }) => {
      const sessions = makeSessions();

      switch (name) {
        case "update":
          await sessions.update(updateParams);
          break;
        case "remove":
          await sessions.remove(removeId);
          break;
      }

      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe(expectedMethod);
      expect(lastRequest!.path).toBe(expectedPath);
    });
  });

  // -------------------------------------------------------------------------
  // Error propagation -- JSON error responses
  // -------------------------------------------------------------------------

  describe("error propagation: JSON error response", () => {
    test.each([
      { name: "update", action: (s: SessionsResource) => s.update(updateParams) },
      { name: "remove", action: (s: SessionsResource) => s.remove(removeId) },
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
        const sessions = createSessionsResource(errHttp);

        try {
          await action(sessions);
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
  // Error propagation -- non-JSON error responses
  // -------------------------------------------------------------------------

  describe("error propagation: non-JSON error response", () => {
    test.each([
      { name: "update", action: (s: SessionsResource) => s.update(updateParams) },
      { name: "remove", action: (s: SessionsResource) => s.remove(removeId) },
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
        const sessions = createSessionsResource(errHttp);

        try {
          await action(sessions);
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
  // Error propagation -- various HTTP status codes
  // -------------------------------------------------------------------------

  describe("error propagation: status codes", () => {
    test.each([
      { status: 400, statusText: "Bad Request" },
      { status: 401, statusText: "Unauthorized" },
      { status: 404, statusText: "Not Found" },
      { status: 422, statusText: "Unprocessable Entity" },
      { status: 429, statusText: "Too Many Requests" },
      { status: 503, statusText: "Service Unavailable" },
    ])("update propagates $status $statusText", async ({ status, statusText }) => {
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
        const sessions = createSessionsResource(errHttp);

        try {
          await sessions.update(updateParams);
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

    test.each([
      { status: 400, statusText: "Bad Request" },
      { status: 401, statusText: "Unauthorized" },
      { status: 404, statusText: "Not Found" },
      { status: 422, statusText: "Unprocessable Entity" },
      { status: 429, statusText: "Too Many Requests" },
      { status: 503, statusText: "Service Unavailable" },
    ])("remove propagates $status $statusText", async ({ status, statusText }) => {
      const { server: errorServer, port: errorPort } = await createInlineServer((_req, res) => {
        res.writeHead(status, statusText, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "test_error" }));
      });

      try {
        const errHttp = createHttpClient({
          baseUrl: `http://localhost:${errorPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const sessions = createSessionsResource(errHttp);

        try {
          await sessions.remove(removeId);
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
  // Abort signal: pre-aborted signal for each method
  // -------------------------------------------------------------------------

  describe("abort signal: pre-aborted", () => {
    test.each([
      { name: "update", action: (s: SessionsResource, opts: { signal: AbortSignal }) => s.update(updateParams, opts) },
      { name: "remove", action: (s: SessionsResource, opts: { signal: AbortSignal }) => s.remove(removeId, opts) },
    ])("$name aborts immediately with pre-aborted signal", async ({ action }) => {
      const sessions = makeSessions();
      const controller = new AbortController();
      controller.abort();

      try {
        await action(sessions, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // Abort signal: user aborts during in-flight request
  // -------------------------------------------------------------------------

  describe("abort signal: user aborts during request", () => {
    test("update aborts when user signal fires mid-request", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        setTimeout(() => {
          res.writeHead(200, { "Content-Type": "application/json" });
          res.end(JSON.stringify(fakeSession));
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const sessions = createSessionsResource(slowHttp);
        const controller = new AbortController();
        setTimeout(() => controller.abort(), 50);

        try {
          await sessions.update(updateParams, { signal: controller.signal });
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

    test("remove aborts when user signal fires mid-request", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer((_req, res) => {
        setTimeout(() => {
          res.writeHead(204);
          res.end();
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const sessions = createSessionsResource(slowHttp);
        const controller = new AbortController();
        setTimeout(() => controller.abort(), 50);

        try {
          await sessions.remove(removeId, { signal: controller.signal });
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
    test("update times out when server is too slow", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        setTimeout(() => {
          res.writeHead(200, { "Content-Type": "application/json" });
          res.end(JSON.stringify(fakeSession));
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 50,
          credentials: "omit",
        });
        const sessions = createSessionsResource(slowHttp);

        try {
          await sessions.update(updateParams);
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

    test("remove times out when server is too slow", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer((_req, res) => {
        setTimeout(() => {
          res.writeHead(204);
          res.end();
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 50,
          credentials: "omit",
        });
        const sessions = createSessionsResource(slowHttp);

        try {
          await sessions.remove(removeId);
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
      const sessions = makeSessions({ apiKey: "session-api-key-123" });
      await sessions.update(updateParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer session-api-key-123");
    });

    test("does not include Authorization header when apiKey is absent", async () => {
      const sessions = makeSessions();
      await sessions.update(updateParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });

    test.each([
      { name: "update", action: (s: SessionsResource) => s.update(updateParams) },
      { name: "remove", action: (s: SessionsResource) => s.remove(removeId) },
    ])("$name sends Bearer token when apiKey is set", async ({ action }) => {
      const sessions = makeSessions({ apiKey: "shared-session-key" });
      await action(sessions);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer shared-session-key");
    });
  });

  // -------------------------------------------------------------------------
  // Content-Type header
  // -------------------------------------------------------------------------

  describe("content-type header", () => {
    test.each([
      { name: "update", action: (s: SessionsResource) => s.update(updateParams) },
      { name: "remove", action: (s: SessionsResource) => s.remove(removeId) },
    ])("$name sends Content-Type: application/json", async ({ action }) => {
      const sessions = makeSessions();
      await action(sessions);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["content-type"]).toBe("application/json");
    });
  });

  // -------------------------------------------------------------------------
  // Resource isolation
  // -------------------------------------------------------------------------

  describe("resource isolation", () => {
    test("two sessions resources from different HttpClients are independent", async () => {
      const sessions1 = makeSessions({ apiKey: "key-alpha" });
      const sessions2 = makeSessions({ apiKey: "key-beta" });

      await sessions1.update(updateParams);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer key-alpha");

      await sessions2.update(updateParams);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer key-beta");
    });

    test("sessions resource without apiKey does not leak auth from another", async () => {
      const authed = makeSessions({ apiKey: "secret-key" });
      const unauthed = makeSessions();

      await authed.update(updateParams);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer secret-key");

      await unauthed.update(updateParams);
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // remove: void return semantics
  // -------------------------------------------------------------------------

  describe("remove void return", () => {
    test("remove resolves undefined (void) on 204 No Content", async () => {
      const sessions = makeSessions();
      const result = await sessions.remove(removeId);
      expect(result).toBeUndefined();
    });

    test("remove does not reject on successful 204 response", async () => {
      const sessions = makeSessions();
      await expect(sessions.remove(removeId)).resolves.toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // update: params with nullable fields
  // -------------------------------------------------------------------------

  describe("update with nullable fields", () => {
    test("sends null fields correctly in JSON body", async () => {
      const sessions = makeSessions();
      const nullableParams: UpdateSessionParams = {
        session_id: "sess-null-test" as SessionID,
        user_id: null,
        created_at: "2025-07-01T00:00:00Z",
        expires_at: "2025-07-02T00:00:00Z",
        last_access: null,
        ip_address: null,
        user_agent: null,
        session_data: null,
      };
      await sessions.update(nullableParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({
        session_id: "sess-null-test",
        user_id: null,
        created_at: "2025-07-01T00:00:00Z",
        expires_at: "2025-07-02T00:00:00Z",
        last_access: null,
        ip_address: null,
        user_agent: null,
        session_data: null,
      });
    });
  });

  // -------------------------------------------------------------------------
  // remove: query param formatting
  // -------------------------------------------------------------------------

  describe("remove query param formatting", () => {
    test.each([
      { name: "simple id", id: "sess-001" as SessionID, expected: "sess-001" },
      { name: "numeric-looking id", id: "99999" as SessionID, expected: "99999" },
      { name: "uuid-style id", id: "550e8400-e29b-41d4-a716-446655440000" as SessionID, expected: "550e8400-e29b-41d4-a716-446655440000" },
    ])("sends $name as q param value: $expected", async ({ id, expected }) => {
      const sessions = makeSessions();
      await sessions.remove(id);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.params["q"]).toBe(expected);
    });
  });
});
