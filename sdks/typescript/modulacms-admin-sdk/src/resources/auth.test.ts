import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createAuthResource } from "./auth.js";
import type { AuthResource } from "./auth.js";
import { createHttpClient } from "../http.js";
import type { HttpClient } from "../http.js";
import type { ApiError } from "../types/common.js";
import type { LoginRequest } from "../types/auth.js";
import type { CreateUserParams, UpdateUserParams } from "../types/users.js";
import type { Email, UserID } from "../types/common.js";

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const fakeLoginResponse = {
  user_id: "u-login-1" as UserID,
  email: "alice@example.com" as Email,
  username: "alice",
  created_at: "2025-06-01T12:00:00Z",
};

const fakeMeResponse = {
  user_id: "u-me-1" as UserID,
  email: "bob@example.com" as Email,
  username: "bob",
  name: "Bob Smith",
  role: "editor",
};

const fakeUser = {
  user_id: "u-reg-1" as UserID,
  username: "carol",
  name: "Carol Danvers",
  email: "carol@example.com" as Email,
  hash: "hashed-pw",
  role: "viewer",
  date_created: "2025-06-01T00:00:00Z",
  date_modified: "2025-06-01T00:00:00Z",
};

const fakeResetToken = "reset-token-xyz-789";

const loginParams: LoginRequest = {
  email: "alice@example.com" as Email,
  password: "s3cret!",
};

const registerParams: CreateUserParams = {
  username: "carol",
  name: "Carol Danvers",
  email: "carol@example.com" as Email,
  hash: "hashed-pw",
  role: "viewer",
  date_created: "2025-06-01T00:00:00Z",
  date_modified: "2025-06-01T00:00:00Z",
};

const resetParams: UpdateUserParams = {
  user_id: "u-reset-1" as UserID,
  username: "dave",
  name: "Dave Grohl",
  email: "dave@example.com" as Email,
  hash: "new-hashed-pw",
  role: "admin",
  date_created: "2025-01-01T00:00:00Z",
  date_modified: "2025-06-15T00:00:00Z",
};

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
} | null = null;

beforeAll(async () => {
  server = http.createServer(async (req, res) => {
    const port = (server.address() as import("net").AddressInfo).port;
    const url = new URL(req.url!, `http://localhost:${port}`);
    const path = url.pathname;

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
    };

    // -- login --
    if (path === "/api/v1/auth/login" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeLoginResponse));
      return;
    }

    // -- logout --
    if (path === "/api/v1/auth/logout" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({}));
      return;
    }

    // -- me --
    if (path === "/api/v1/auth/me" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeMeResponse));
      return;
    }

    // -- register --
    if (path === "/api/v1/auth/register" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeUser));
      return;
    }

    // -- reset --
    if (path === "/api/v1/auth/reset" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeResetToken));
      return;
    }

    // -- slow endpoint for abort tests --
    if (path === "/api/v1/auth/slow" && req.method === "POST") {
      setTimeout(() => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify(fakeLoginResponse));
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

function makeAuth(overrides?: {
  apiKey?: string;
  defaultTimeout?: number;
}): AuthResource {
  return createAuthResource(makeClient(overrides));
}

// ---------------------------------------------------------------------------
// Inline server helper -- creates a node:http server and returns it + port
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

describe("createAuthResource", () => {
  // -------------------------------------------------------------------------
  // Factory
  // -------------------------------------------------------------------------

  describe("factory", () => {
    test("returns an object with all five auth methods", () => {
      const auth = makeAuth();
      expect(typeof auth.login).toBe("function");
      expect(typeof auth.logout).toBe("function");
      expect(typeof auth.me).toBe("function");
      expect(typeof auth.register).toBe("function");
      expect(typeof auth.reset).toBe("function");
    });

    test("all methods return promises", () => {
      const auth = makeAuth();

      const loginResult = auth.login(loginParams);
      expect(typeof loginResult.then).toBe("function");
      loginResult.catch(() => {});

      const logoutResult = auth.logout();
      expect(typeof logoutResult.then).toBe("function");
      logoutResult.catch(() => {});

      const meResult = auth.me();
      expect(typeof meResult.then).toBe("function");
      meResult.catch(() => {});

      const registerResult = auth.register(registerParams);
      expect(typeof registerResult.then).toBe("function");
      registerResult.catch(() => {});

      const resetResult = auth.reset(resetParams);
      expect(typeof resetResult.then).toBe("function");
      resetResult.catch(() => {});
    });
  });

  // -------------------------------------------------------------------------
  // login
  // -------------------------------------------------------------------------

  describe("login", () => {
    test("returns parsed LoginResponse on success", async () => {
      const auth = makeAuth();
      const result = await auth.login(loginParams);
      expect(result).toEqual(fakeLoginResponse);
    });

    test("sends POST to /auth/login", async () => {
      const auth = makeAuth();
      await auth.login(loginParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/auth/login");
    });

    test("sends login params as JSON body", async () => {
      const auth = makeAuth();
      await auth.login(loginParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({
        email: "alice@example.com",
        password: "s3cret!",
      });
    });

    test("forwards RequestOptions signal", async () => {
      const auth = makeAuth();
      const controller = new AbortController();
      controller.abort();
      try {
        await auth.login(loginParams, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // logout
  // -------------------------------------------------------------------------

  describe("logout", () => {
    test("resolves void on success", async () => {
      const auth = makeAuth();
      const result = await auth.logout();
      expect(result).toBeUndefined();
    });

    test("sends POST to /auth/logout", async () => {
      const auth = makeAuth();
      await auth.logout();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/auth/logout");
    });

    test("sends no body (undefined)", async () => {
      const auth = makeAuth();
      await auth.logout();
      expect(lastRequest).not.toBeNull();
      // http.post receives undefined as body; JSON.stringify(undefined) is not called,
      // so the server receives no JSON body. The body parse attempt yields null.
      expect(lastRequest!.body).toBeNull();
    });

    test("forwards RequestOptions signal", async () => {
      const auth = makeAuth();
      const controller = new AbortController();
      controller.abort();
      try {
        await auth.logout({ signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // me
  // -------------------------------------------------------------------------

  describe("me", () => {
    test("returns parsed MeResponse on success", async () => {
      const auth = makeAuth();
      const result = await auth.me();
      expect(result).toEqual(fakeMeResponse);
    });

    test("sends GET to /auth/me", async () => {
      const auth = makeAuth();
      await auth.me();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/auth/me");
    });

    test("forwards RequestOptions signal", async () => {
      const auth = makeAuth();
      const controller = new AbortController();
      controller.abort();
      try {
        await auth.me({ signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // register
  // -------------------------------------------------------------------------

  describe("register", () => {
    test("returns parsed User on success", async () => {
      const auth = makeAuth();
      const result = await auth.register(registerParams);
      expect(result).toEqual(fakeUser);
    });

    test("sends POST to /auth/register", async () => {
      const auth = makeAuth();
      await auth.register(registerParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/auth/register");
    });

    test("sends CreateUserParams as JSON body", async () => {
      const auth = makeAuth();
      await auth.register(registerParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({
        username: "carol",
        name: "Carol Danvers",
        email: "carol@example.com",
        hash: "hashed-pw",
        role: "viewer",
        date_created: "2025-06-01T00:00:00Z",
        date_modified: "2025-06-01T00:00:00Z",
      });
    });

    test("forwards RequestOptions signal", async () => {
      const auth = makeAuth();
      const controller = new AbortController();
      controller.abort();
      try {
        await auth.register(registerParams, { signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // reset
  // -------------------------------------------------------------------------

  describe("reset", () => {
    test("returns parsed string on success", async () => {
      const auth = makeAuth();
      const result = await auth.reset(resetParams);
      expect(result).toBe(fakeResetToken);
    });

    test("sends POST to /auth/reset", async () => {
      const auth = makeAuth();
      await auth.reset(resetParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/auth/reset");
    });

    test("sends UpdateUserParams as JSON body", async () => {
      const auth = makeAuth();
      await auth.reset(resetParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.body).toEqual({
        user_id: "u-reset-1",
        username: "dave",
        name: "Dave Grohl",
        email: "dave@example.com",
        hash: "new-hashed-pw",
        role: "admin",
        date_created: "2025-01-01T00:00:00Z",
        date_modified: "2025-06-15T00:00:00Z",
      });
    });

    test("forwards RequestOptions signal", async () => {
      const auth = makeAuth();
      const controller = new AbortController();
      controller.abort();
      try {
        await auth.reset(resetParams, { signal: controller.signal });
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
      { name: "login", expectedMethod: "POST", expectedPath: "/api/v1/auth/login" },
      { name: "logout", expectedMethod: "POST", expectedPath: "/api/v1/auth/logout" },
      { name: "me", expectedMethod: "GET", expectedPath: "/api/v1/auth/me" },
      { name: "register", expectedMethod: "POST", expectedPath: "/api/v1/auth/register" },
      { name: "reset", expectedMethod: "POST", expectedPath: "/api/v1/auth/reset" },
    ])("$name sends $expectedMethod to $expectedPath", async ({ name, expectedMethod, expectedPath }) => {
      const auth = makeAuth();

      switch (name) {
        case "login":
          await auth.login(loginParams);
          break;
        case "logout":
          await auth.logout();
          break;
        case "me":
          await auth.me();
          break;
        case "register":
          await auth.register(registerParams);
          break;
        case "reset":
          await auth.reset(resetParams);
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
      { name: "login", action: (auth: AuthResource) => auth.login(loginParams) },
      { name: "logout", action: (auth: AuthResource) => auth.logout() },
      { name: "me", action: (auth: AuthResource) => auth.me() },
      { name: "register", action: (auth: AuthResource) => auth.register(registerParams) },
      { name: "reset", action: (auth: AuthResource) => auth.reset(resetParams) },
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
        const auth = createAuthResource(errHttp);

        try {
          await action(auth);
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
      { name: "login", action: (auth: AuthResource) => auth.login(loginParams) },
      { name: "logout", action: (auth: AuthResource) => auth.logout() },
      { name: "me", action: (auth: AuthResource) => auth.me() },
      { name: "register", action: (auth: AuthResource) => auth.register(registerParams) },
      { name: "reset", action: (auth: AuthResource) => auth.reset(resetParams) },
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
        const auth = createAuthResource(errHttp);

        try {
          await action(auth);
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
    ])("login propagates $status $statusText", async ({ status, statusText }) => {
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
        const auth = createAuthResource(errHttp);

        try {
          await auth.login(loginParams);
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
      { name: "login", action: (auth: AuthResource, opts: { signal: AbortSignal }) => auth.login(loginParams, opts) },
      { name: "logout", action: (auth: AuthResource, opts: { signal: AbortSignal }) => auth.logout(opts) },
      { name: "me", action: (auth: AuthResource, opts: { signal: AbortSignal }) => auth.me(opts) },
      { name: "register", action: (auth: AuthResource, opts: { signal: AbortSignal }) => auth.register(registerParams, opts) },
      { name: "reset", action: (auth: AuthResource, opts: { signal: AbortSignal }) => auth.reset(resetParams, opts) },
    ])("$name aborts immediately with pre-aborted signal", async ({ action }) => {
      const auth = makeAuth();
      const controller = new AbortController();
      controller.abort();

      try {
        await action(auth, { signal: controller.signal });
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
    test("login aborts when user signal fires mid-request", async () => {
      const { server: slowServer, port: slowPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        setTimeout(() => {
          res.writeHead(200, { "Content-Type": "application/json" });
          res.end(JSON.stringify(fakeLoginResponse));
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const auth = createAuthResource(slowHttp);
        const controller = new AbortController();
        setTimeout(() => controller.abort(), 50);

        try {
          await auth.login(loginParams, { signal: controller.signal });
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
          res.end(JSON.stringify(fakeLoginResponse));
        }, 5000);
      });

      try {
        const slowHttp = createHttpClient({
          baseUrl: `http://localhost:${slowPort}`,
          defaultTimeout: 50,
          credentials: "omit",
        });
        const auth = createAuthResource(slowHttp);

        try {
          await auth.login(loginParams);
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
      const auth = makeAuth({ apiKey: "my-api-key-123" });
      await auth.login(loginParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer my-api-key-123");
    });

    test("does not include Authorization header when apiKey is absent", async () => {
      const auth = makeAuth();
      await auth.login(loginParams);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });

    test.each([
      { name: "login", action: (auth: AuthResource) => auth.login(loginParams) },
      { name: "logout", action: (auth: AuthResource) => auth.logout() },
      { name: "me", action: (auth: AuthResource) => auth.me() },
      { name: "register", action: (auth: AuthResource) => auth.register(registerParams) },
      { name: "reset", action: (auth: AuthResource) => auth.reset(resetParams) },
    ])("$name sends Bearer token when apiKey is set", async ({ action }) => {
      const auth = makeAuth({ apiKey: "shared-key-456" });
      await action(auth);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer shared-key-456");
    });
  });

  // -------------------------------------------------------------------------
  // Resource isolation
  // -------------------------------------------------------------------------

  describe("resource isolation", () => {
    test("two auth resources from different HttpClients are independent", async () => {
      const auth1 = makeAuth({ apiKey: "key-alpha" });
      const auth2 = makeAuth({ apiKey: "key-beta" });

      await auth1.login(loginParams);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer key-alpha");

      await auth2.login(loginParams);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer key-beta");
    });

    test("auth resource without apiKey does not leak auth from another", async () => {
      const authed = makeAuth({ apiKey: "secret-key" });
      const unauthed = makeAuth();

      await authed.login(loginParams);
      expect(lastRequest!.headers["authorization"]).toBe("Bearer secret-key");

      await unauthed.login(loginParams);
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // logout: void return semantics
  // -------------------------------------------------------------------------

  describe("logout void return", () => {
    test("logout awaits http.post and discards the response body", async () => {
      // The server returns JSON {} but logout should return undefined
      const auth = makeAuth();
      const result = await auth.logout();
      expect(result).toBeUndefined();
    });

    test("logout does not reject when server returns empty JSON object", async () => {
      const auth = makeAuth();
      // This should resolve without error -- the {} is consumed and discarded
      await expect(auth.logout()).resolves.toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // Successful response shape validation
  // -------------------------------------------------------------------------

  describe("response shape validation", () => {
    test("login response has user_id, email, username, created_at", async () => {
      const auth = makeAuth();
      const result = await auth.login(loginParams);
      expect(result.user_id).toBe("u-login-1");
      expect(result.email).toBe("alice@example.com");
      expect(result.username).toBe("alice");
      expect(result.created_at).toBe("2025-06-01T12:00:00Z");
    });

    test("me response has user_id, email, username, name, role", async () => {
      const auth = makeAuth();
      const result = await auth.me();
      expect(result.user_id).toBe("u-me-1");
      expect(result.email).toBe("bob@example.com");
      expect(result.username).toBe("bob");
      expect(result.name).toBe("Bob Smith");
      expect(result.role).toBe("editor");
    });

    test("register response has all User fields", async () => {
      const auth = makeAuth();
      const result = await auth.register(registerParams);
      expect(result.user_id).toBe("u-reg-1");
      expect(result.username).toBe("carol");
      expect(result.name).toBe("Carol Danvers");
      expect(result.email).toBe("carol@example.com");
      expect(result.hash).toBe("hashed-pw");
      expect(result.role).toBe("viewer");
      expect(result.date_created).toBe("2025-06-01T00:00:00Z");
      expect(result.date_modified).toBe("2025-06-01T00:00:00Z");
    });

    test("reset response is a string token", async () => {
      const auth = makeAuth();
      const result = await auth.reset(resetParams);
      expect(result).toBe("reset-token-xyz-789");
      expect(typeof result).toBe("string");
    });
  });

  // -------------------------------------------------------------------------
  // Content-Type header is set correctly (application/json)
  // -------------------------------------------------------------------------

  describe("content-type header", () => {
    test.each([
      { name: "login", action: (auth: AuthResource) => auth.login(loginParams) },
      { name: "logout", action: (auth: AuthResource) => auth.logout() },
      { name: "me", action: (auth: AuthResource) => auth.me() },
      { name: "register", action: (auth: AuthResource) => auth.register(registerParams) },
      { name: "reset", action: (auth: AuthResource) => auth.reset(resetParams) },
    ])("$name sends Content-Type: application/json", async ({ action }) => {
      const auth = makeAuth();
      await action(auth);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["content-type"]).toBe("application/json");
    });
  });
});
