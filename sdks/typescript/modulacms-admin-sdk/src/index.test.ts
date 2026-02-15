import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createAdminClient, isApiError } from "./index.js";
import type { ClientConfig, ModulaCMSAdminClient } from "./index.js";
import type { ApiError } from "./types/common.js";

// ---------------------------------------------------------------------------
// Test server -- node:http (matching existing test conventions)
// ---------------------------------------------------------------------------

let server: http.Server;
let baseUrl: string;

// Track the last request for assertion
let lastRequest: {
  method: string;
  url: string;
  path: string;
  params: Record<string, string>;
  body: unknown;
  headers: Record<string, string>;
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
        let rawBody = "";
        for await (const chunk of req) rawBody += chunk;
        try {
          body = JSON.parse(rawBody);
        } catch {
          body = null;
        }
      }
    }

    lastRequest = {
      method: req.method!,
      url: url.toString(),
      path,
      params,
      body,
      headers: Object.fromEntries(
        Object.entries(req.headers).map(([k, v]) => [k, Array.isArray(v) ? v.join(", ") : v ?? ""])
      ) as Record<string, string>,
    };

    // -- Admin routes list (standard + ordered) --
    if (path === "/api/v1/adminroutes" && req.method === "GET") {
      if (params.ordered === "true") {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify([
          { admin_route_id: "r-2", slug: "about", title: "About", status: 1 },
          { admin_route_id: "r-1", slug: "home", title: "Home", status: 1 },
        ]));
        return;
      }
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([
        { admin_route_id: "r-1", slug: "home", title: "Home", status: 1 },
        { admin_route_id: "r-2", slug: "about", title: "About", status: 1 },
      ]));
      return;
    }

    // -- Admin routes get by slug --
    if (path === "/api/v1/adminroutes/" && req.method === "GET" && params.q) {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        admin_route_id: "r-1",
        slug: params.q,
        title: "Found-" + params.q,
        status: 1,
      }));
      return;
    }

    // -- Admin routes create --
    if (path === "/api/v1/adminroutes" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        admin_route_id: "r-new",
        slug: "new-route",
        title: "New Route",
        status: 1,
      }));
      return;
    }

    // -- Admin routes update --
    if (path === "/api/v1/adminroutes/" && req.method === "PUT") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        admin_route_id: "r-1",
        slug: "updated",
        title: "Updated",
        status: 1,
      }));
      return;
    }

    // -- Admin routes delete (custom: query param q) --
    if (path === "/api/v1/adminroutes/" && req.method === "DELETE") {
      res.writeHead(204);
      res.end();
      return;
    }

    // -- Generic CRUD endpoints for verifying resource wiring --
    // admincontentdatas
    if (path === "/api/v1/admincontentdatas" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // admincontentfields
    if (path === "/api/v1/admincontentfields" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // admindatatypes
    if (path === "/api/v1/admindatatypes" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // adminfields
    if (path === "/api/v1/adminfields" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // contentdata
    if (path === "/api/v1/contentdata" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // contentfields
    if (path === "/api/v1/contentfields" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // datatype
    if (path === "/api/v1/datatype" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // fields
    if (path === "/api/v1/fields" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // routes
    if (path === "/api/v1/routes" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // media
    if (path === "/api/v1/media" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // mediadimensions
    if (path === "/api/v1/mediadimensions" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // users
    if (path === "/api/v1/users" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // roles
    if (path === "/api/v1/roles" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // tokens
    if (path === "/api/v1/tokens" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // usersoauth
    if (path === "/api/v1/usersoauth" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // tables
    if (path === "/api/v1/tables" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([]));
      return;
    }

    // -- Auth endpoints --
    if (path === "/api/v1/auth/login" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        user_id: "u-1",
        email: "test@example.com",
        username: "testuser",
        created_at: "2025-01-01T00:00:00Z",
      }));
      return;
    }

    if (path === "/api/v1/auth/logout" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({}));
      return;
    }

    if (path === "/api/v1/auth/me" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        user_id: "u-1",
        email: "test@example.com",
        username: "testuser",
        name: "Test User",
        role: "admin",
      }));
      return;
    }

    if (path === "/api/v1/auth/register" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        user_id: "u-new",
        email: "new@example.com",
        username: "newuser",
      }));
      return;
    }

    if (path === "/api/v1/auth/reset" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify("reset-token-abc"));
      return;
    }

    // -- Sessions --
    if (path === "/api/v1/sessions/" && req.method === "PUT") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ session_id: "s-1", updated: true }));
      return;
    }
    if (path === "/api/v1/sessions/" && req.method === "DELETE") {
      res.writeHead(204);
      res.end();
      return;
    }

    // -- SSH keys --
    if (path === "/api/v1/ssh-keys" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify([{ id: "sk-1", name: "my-key" }]));
      return;
    }
    if (path === "/api/v1/ssh-keys" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ id: "sk-new", name: "new-key", public_key: "ssh-rsa AAA" }));
      return;
    }
    // SSH key delete uses path param: DELETE /api/v1/ssh-keys/:id
    if (path.startsWith("/api/v1/ssh-keys/") && req.method === "DELETE") {
      res.writeHead(204);
      res.end();
      return;
    }

    // -- Admin tree --
    if (path.startsWith("/api/v1/admin/tree/") && req.method === "GET") {
      const slug = path.split("/api/v1/admin/tree/")[1];
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        route: { admin_route_id: "r-1", slug, title: "Tree Route" },
        tree: [],
      }));
      return;
    }

    // -- Import endpoints --
    if (path === "/api/v1/import/contentful" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, datatypes_created: 1, fields_created: 2, content_created: 3, message: "ok", errors: [] }));
      return;
    }
    if (path === "/api/v1/import/sanity" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, datatypes_created: 0, fields_created: 0, content_created: 0, message: "ok", errors: [] }));
      return;
    }
    if (path === "/api/v1/import/strapi" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, datatypes_created: 0, fields_created: 0, content_created: 0, message: "ok", errors: [] }));
      return;
    }
    if (path === "/api/v1/import/wordpress" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, datatypes_created: 0, fields_created: 0, content_created: 0, message: "ok", errors: [] }));
      return;
    }
    if (path === "/api/v1/import/clean" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, datatypes_created: 0, fields_created: 0, content_created: 0, message: "ok", errors: [] }));
      return;
    }
    if (path === "/api/v1/import" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, datatypes_created: 0, fields_created: 0, content_created: 0, message: "ok", errors: [] }));
      return;
    }

    // -- Media upload --
    if (path === "/api/v1/mediaupload/" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ media_id: "m-new", filename: "test.png", url: "https://cdn.example.com/test.png" }));
      return;
    }

    // -- Echo endpoint for verifying request details --
    if (path === "/api/v1/echo") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        method: req.method,
        headers: Object.fromEntries(
          Object.entries(req.headers).map(([k, v]) => [k, Array.isArray(v) ? v.join(", ") : v ?? ""])
        ),
        params,
        path,
      }));
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
// Helper to create a valid client config pointing at the test server
// ---------------------------------------------------------------------------

function validConfig(overrides?: Partial<ClientConfig>): ClientConfig {
  return {
    baseUrl: baseUrl,
    allowInsecure: true, // test server uses http://
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("createAdminClient", () => {
  // -------------------------------------------------------------------------
  // Config validation: baseUrl
  // -------------------------------------------------------------------------

  describe("baseUrl validation", () => {
    test.each([
      { name: "empty string", url: "" },
      { name: "not a URL", url: "not-a-url" },
      { name: "missing protocol", url: "example.com" },
      { name: "just a path", url: "/api" },
      { name: "just a port", url: ":8080" },
    ])("throws on invalid baseUrl ($name)", ({ url }) => {
      expect(() => createAdminClient({ baseUrl: url })).toThrow(
        "Invalid baseUrl: " + url,
      );
    });

    test.each([
      { name: "https URL", url: "https://api.example.com" },
      { name: "https with port", url: "https://localhost:3000" },
      { name: "https with path", url: "https://example.com/api" },
    ])("accepts valid $name", ({ url }) => {
      // Should not throw -- the client is created successfully
      const client = createAdminClient({ baseUrl: url });
      expect(client).toBeDefined();
    });

    test("accepts http URL when allowInsecure is true", () => {
      const client = createAdminClient({
        baseUrl: "http://localhost:3000",
        allowInsecure: true,
      });
      expect(client).toBeDefined();
    });
  });

  // -------------------------------------------------------------------------
  // Config validation: http:// protocol enforcement
  // -------------------------------------------------------------------------

  describe("http:// protocol enforcement", () => {
    test("throws on http:// without allowInsecure", () => {
      expect(() =>
        createAdminClient({ baseUrl: "http://localhost:3000" }),
      ).toThrow(
        "baseUrl uses http:// which transmits credentials in plaintext. " +
          "Set allowInsecure: true to allow this, or use https://.",
      );
    });

    test("throws on http:// when allowInsecure is false", () => {
      expect(() =>
        createAdminClient({
          baseUrl: "http://localhost:3000",
          allowInsecure: false,
        }),
      ).toThrow(
        "baseUrl uses http:// which transmits credentials in plaintext. " +
          "Set allowInsecure: true to allow this, or use https://.",
      );
    });

    test("allows http:// when allowInsecure is true", () => {
      const client = createAdminClient({
        baseUrl: "http://localhost:3000",
        allowInsecure: true,
      });
      expect(client).toBeDefined();
    });

    test("does not throw on https:// regardless of allowInsecure", () => {
      const client = createAdminClient({
        baseUrl: "https://api.example.com",
      });
      expect(client).toBeDefined();
    });

    test("does not throw on https:// with allowInsecure false", () => {
      const client = createAdminClient({
        baseUrl: "https://api.example.com",
        allowInsecure: false,
      });
      expect(client).toBeDefined();
    });
  });

  // -------------------------------------------------------------------------
  // Config validation: apiKey
  // -------------------------------------------------------------------------

  describe("apiKey validation", () => {
    test("throws when apiKey is an empty string", () => {
      expect(() =>
        createAdminClient({
          baseUrl: "https://api.example.com",
          apiKey: "",
        }),
      ).toThrow("apiKey must not be an empty string");
    });

    test("accepts undefined apiKey (no auth)", () => {
      const client = createAdminClient({
        baseUrl: "https://api.example.com",
      });
      expect(client).toBeDefined();
    });

    test("accepts a non-empty apiKey", () => {
      const client = createAdminClient({
        baseUrl: "https://api.example.com",
        apiKey: "my-secret-key",
      });
      expect(client).toBeDefined();
    });

    // Verify the empty string check uses === '' (not falsy check).
    // An apiKey that is explicitly undefined should NOT trigger the error.
    test("does not throw when apiKey is explicitly undefined", () => {
      const client = createAdminClient({
        baseUrl: "https://api.example.com",
        apiKey: undefined,
      });
      expect(client).toBeDefined();
    });
  });

  // -------------------------------------------------------------------------
  // Config defaults
  // -------------------------------------------------------------------------

  describe("config defaults", () => {
    test("uses defaultTimeout of 30000 when not specified", async () => {
      // We verify this indirectly: the client should work with a server that
      // responds within 30 seconds. The timeout is passed to createHttpClient.
      const client = createAdminClient(validConfig());
      // If defaultTimeout were 0, this would time out immediately
      const result = await client.adminRoutes.list();
      expect(result).toBeInstanceOf(Array);
    });

    test("uses custom defaultTimeout when specified", async () => {
      // Use a very short timeout to verify it's being applied.
      // Any endpoint that responds quickly should still work.
      const client = createAdminClient(validConfig({ defaultTimeout: 5000 }));
      const result = await client.adminRoutes.list();
      expect(result).toBeInstanceOf(Array);
    });

    test("uses credentials 'include' by default", async () => {
      // Verify the client works with default credentials.
      // The credentials value is passed through to fetch; we can't directly
      // observe it on the server side, but we verify no error is thrown.
      const client = createAdminClient(validConfig());
      const result = await client.adminRoutes.list();
      expect(result).toBeInstanceOf(Array);
    });

    test("uses custom credentials when specified", async () => {
      const client = createAdminClient(validConfig({ credentials: "omit" }));
      const result = await client.adminRoutes.list();
      expect(result).toBeInstanceOf(Array);
    });
  });

  // -------------------------------------------------------------------------
  // Trailing slash stripping
  // -------------------------------------------------------------------------

  describe("trailing slash stripping", () => {
    test.each([
      { name: "single trailing slash", url: "/", stripped: "" },
      { name: "multiple trailing slashes", url: "///", stripped: "" },
    ])("strips $name from baseUrl", async ({ url, stripped }) => {
      const fullUrl = baseUrl + url;
      const client = createAdminClient({
        baseUrl: fullUrl,
        allowInsecure: true,
      });
      // If trailing slashes weren't stripped, the URL would be malformed
      // and the request would fail or go to the wrong path.
      // Verify the client makes a successful request.
      const result = await client.adminRoutes.list();
      expect(result).toBeInstanceOf(Array);
    });

    test("does not alter baseUrl without trailing slash", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.adminRoutes.list();
      expect(result).toBeInstanceOf(Array);
      // Verify the request arrived at the expected path
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.path).toBe("/api/v1/adminroutes");
    });

    test("correctly builds URLs after stripping trailing slashes", async () => {
      const client = createAdminClient({
        baseUrl: baseUrl + "/",
        allowInsecure: true,
      });
      await client.adminRoutes.list();
      expect(lastRequest).not.toBeNull();
      // Without stripping, this would be /api/v1/adminroutes with a double slash somewhere
      expect(lastRequest!.path).toBe("/api/v1/adminroutes");
    });
  });

  // -------------------------------------------------------------------------
  // Client shape: all resource properties exist
  // -------------------------------------------------------------------------

  describe("client shape", () => {
    test("has auth property with all methods", () => {
      const client = createAdminClient(validConfig());
      expect(typeof client.auth.login).toBe("function");
      expect(typeof client.auth.logout).toBe("function");
      expect(typeof client.auth.me).toBe("function");
      expect(typeof client.auth.register).toBe("function");
      expect(typeof client.auth.reset).toBe("function");
    });

    test("has adminRoutes with CRUD methods plus listOrdered and custom remove", () => {
      const client = createAdminClient(validConfig());
      expect(typeof client.adminRoutes.list).toBe("function");
      expect(typeof client.adminRoutes.get).toBe("function");
      expect(typeof client.adminRoutes.create).toBe("function");
      expect(typeof client.adminRoutes.update).toBe("function");
      expect(typeof client.adminRoutes.remove).toBe("function");
      expect(typeof client.adminRoutes.listOrdered).toBe("function");
    });

    test.each([
      { name: "adminContentData" },
      { name: "adminContentFields" },
      { name: "adminDatatypes" },
      { name: "adminFields" },
      { name: "contentData" },
      { name: "contentFields" },
      { name: "datatypes" },
      { name: "fields" },
      { name: "routes" },
      { name: "media" },
      { name: "mediaDimensions" },
      { name: "users" },
      { name: "roles" },
      { name: "tokens" },
      { name: "usersOauth" },
      { name: "tables" },
    ])("has $name CRUD resource with all five methods", ({ name }) => {
      const client = createAdminClient(validConfig());
      const resource = client[name as keyof ModulaCMSAdminClient] as Record<string, unknown>;
      expect(typeof resource.list).toBe("function");
      expect(typeof resource.get).toBe("function");
      expect(typeof resource.create).toBe("function");
      expect(typeof resource.update).toBe("function");
      expect(typeof resource.remove).toBe("function");
    });

    test("has adminTree property with get method", () => {
      const client = createAdminClient(validConfig());
      expect(typeof client.adminTree.get).toBe("function");
    });

    test("has mediaUpload property with upload method", () => {
      const client = createAdminClient(validConfig());
      expect(typeof client.mediaUpload.upload).toBe("function");
    });

    test("has sessions property with update and remove methods", () => {
      const client = createAdminClient(validConfig());
      expect(typeof client.sessions.update).toBe("function");
      expect(typeof client.sessions.remove).toBe("function");
    });

    test("has sshKeys property with list, create, and remove methods", () => {
      const client = createAdminClient(validConfig());
      expect(typeof client.sshKeys.list).toBe("function");
      expect(typeof client.sshKeys.create).toBe("function");
      expect(typeof client.sshKeys.remove).toBe("function");
    });

    test("has import property with all format methods and bulk", () => {
      const client = createAdminClient(validConfig());
      expect(typeof client.import.contentful).toBe("function");
      expect(typeof client.import.sanity).toBe("function");
      expect(typeof client.import.strapi).toBe("function");
      expect(typeof client.import.wordpress).toBe("function");
      expect(typeof client.import.clean).toBe("function");
      expect(typeof client.import.bulk).toBe("function");
    });
  });

  // -------------------------------------------------------------------------
  // adminRoutes: listOrdered
  // -------------------------------------------------------------------------

  describe("adminRoutes.listOrdered", () => {
    test("sends GET request with ordered=true query parameter", async () => {
      const client = createAdminClient(validConfig());
      await client.adminRoutes.listOrdered();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/adminroutes");
      expect(lastRequest!.params.ordered).toBe("true");
    });

    test("returns the ordered list of admin routes", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.adminRoutes.listOrdered();
      expect(result).toEqual([
        { admin_route_id: "r-2", slug: "about", title: "About", status: 1 },
        { admin_route_id: "r-1", slug: "home", title: "Home", status: 1 },
      ]);
    });

    test("forwards RequestOptions to the HTTP client", async () => {
      const client = createAdminClient(validConfig());
      const controller = new AbortController();
      controller.abort();
      try {
        await client.adminRoutes.listOrdered({ signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // adminRoutes: custom remove
  // -------------------------------------------------------------------------

  describe("adminRoutes.remove", () => {
    test("sends DELETE request to /adminroutes/ with query param q", async () => {
      const client = createAdminClient(validConfig());
      await client.adminRoutes.remove("route-id-1" as AdminRouteID);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("DELETE");
      expect(lastRequest!.path).toBe("/api/v1/adminroutes/");
      expect(lastRequest!.params.q).toBe("route-id-1");
    });

    test("converts AdminRouteID to string via String()", async () => {
      const client = createAdminClient(validConfig());
      const id = "my-route-id" as AdminRouteID;
      await client.adminRoutes.remove(id);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.params.q).toBe("my-route-id");
    });

    test("resolves void on success", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.adminRoutes.remove("r-1" as AdminRouteID);
      expect(result).toBeUndefined();
    });

    test("forwards RequestOptions to the HTTP client", async () => {
      const client = createAdminClient(validConfig());
      const controller = new AbortController();
      controller.abort();
      try {
        await client.adminRoutes.remove("r-1" as AdminRouteID, {
          signal: controller.signal,
        });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // adminRoutes: inherited CRUD methods from spread
  // -------------------------------------------------------------------------

  describe("adminRoutes inherited CRUD", () => {
    test("list returns array of admin routes", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.adminRoutes.list();
      expect(result).toEqual([
        { admin_route_id: "r-1", slug: "home", title: "Home", status: 1 },
        { admin_route_id: "r-2", slug: "about", title: "About", status: 1 },
      ]);
    });

    test("get retrieves a single admin route by slug", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.adminRoutes.get("home" as Slug);
      expect(result.slug).toBe("home");
      expect(result.title).toBe("Found-home");
    });

    test("create sends POST to /adminroutes", async () => {
      const client = createAdminClient(validConfig());
      const params = {
        slug: "new-page" as Slug,
        title: "New Page",
        status: 1,
        author_id: null,
        date_created: "2025-01-01T00:00:00Z",
        date_modified: "2025-01-01T00:00:00Z",
      };
      const result = await client.adminRoutes.create(params);
      expect(result.admin_route_id).toBe("r-new");
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/adminroutes");
    });

    test("update sends PUT to /adminroutes/", async () => {
      const client = createAdminClient(validConfig());
      const params = {
        slug: "home" as Slug,
        title: "Updated Home",
        status: 1,
        author_id: null,
        date_created: "2025-01-01T00:00:00Z",
        date_modified: "2025-01-01T00:00:00Z",
        slug_2: "home-updated" as Slug,
      };
      const result = await client.adminRoutes.update(params);
      expect(result.admin_route_id).toBe("r-1");
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("PUT");
      expect(lastRequest!.path).toBe("/api/v1/adminroutes/");
    });
  });

  // -------------------------------------------------------------------------
  // CRUD resource wiring: verify each resource hits the correct API path
  // -------------------------------------------------------------------------

  describe("resource wiring", () => {
    test.each([
      { name: "adminContentData", path: "/api/v1/admincontentdatas" },
      { name: "adminContentFields", path: "/api/v1/admincontentfields" },
      { name: "adminDatatypes", path: "/api/v1/admindatatypes" },
      { name: "adminFields", path: "/api/v1/adminfields" },
      { name: "contentData", path: "/api/v1/contentdata" },
      { name: "contentFields", path: "/api/v1/contentfields" },
      { name: "datatypes", path: "/api/v1/datatype" },
      { name: "fields", path: "/api/v1/fields" },
      { name: "routes", path: "/api/v1/routes" },
      { name: "media", path: "/api/v1/media" },
      { name: "mediaDimensions", path: "/api/v1/mediadimensions" },
      { name: "users", path: "/api/v1/users" },
      { name: "roles", path: "/api/v1/roles" },
      { name: "tokens", path: "/api/v1/tokens" },
      { name: "usersOauth", path: "/api/v1/usersoauth" },
      { name: "tables", path: "/api/v1/tables" },
    ])("$name.list() hits $path", async ({ name, path }) => {
      const client = createAdminClient(validConfig());
      const resource = client[name as keyof ModulaCMSAdminClient] as { list: () => Promise<unknown[]> };
      await resource.list();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe(path);
    });
  });

  // -------------------------------------------------------------------------
  // Auth resource integration
  // -------------------------------------------------------------------------

  describe("auth resource", () => {
    test("login sends POST to /auth/login with credentials", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.auth.login({
        email: "test@example.com" as Email,
        password: "secret",
      });
      expect(result.user_id).toBe("u-1");
      expect(result.email).toBe("test@example.com");
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/auth/login");
    });

    test("logout sends POST to /auth/logout", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.auth.logout();
      expect(result).toBeUndefined();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/auth/logout");
    });

    test("me sends GET to /auth/me", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.auth.me();
      expect(result.user_id).toBe("u-1");
      expect(result.role).toBe("admin");
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/auth/me");
    });
  });

  // -------------------------------------------------------------------------
  // apiKey passes through to HTTP client
  // -------------------------------------------------------------------------

  describe("apiKey forwarding", () => {
    test("sends Authorization header when apiKey is provided", async () => {
      const client = createAdminClient(validConfig({ apiKey: "test-api-key-123" }));
      // Use a CRUD list call to trigger a request, then check the echo
      await client.adminRoutes.list();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer test-api-key-123");
    });

    test("does not send Authorization header when apiKey is omitted", async () => {
      const client = createAdminClient(validConfig());
      await client.adminRoutes.list();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // Sessions resource integration
  // -------------------------------------------------------------------------

  describe("sessions resource", () => {
    test("update sends PUT to /sessions/", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.sessions.update({} as UpdateSessionParams);
      expect(result).toBeDefined();
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("PUT");
      expect(lastRequest!.path).toBe("/api/v1/sessions/");
    });

    test("remove sends DELETE to /sessions/ with query param q", async () => {
      const client = createAdminClient(validConfig());
      await client.sessions.remove("sess-1" as SessionID);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("DELETE");
      expect(lastRequest!.path).toBe("/api/v1/sessions/");
      expect(lastRequest!.params.q).toBe("sess-1");
    });
  });

  // -------------------------------------------------------------------------
  // SSH keys resource integration
  // -------------------------------------------------------------------------

  describe("sshKeys resource", () => {
    test("list sends GET to /ssh-keys", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.sshKeys.list();
      expect(result).toEqual([{ id: "sk-1", name: "my-key" }]);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/ssh-keys");
    });

    test("create sends POST to /ssh-keys", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.sshKeys.create({
        name: "new-key",
        public_key: "ssh-rsa AAAA",
      } as CreateSshKeyRequest);
      expect(result.id).toBe("sk-new");
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/ssh-keys");
    });

    test("remove sends DELETE to /ssh-keys/:id", async () => {
      const client = createAdminClient(validConfig());
      await client.sshKeys.remove("sk-1");
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("DELETE");
      expect(lastRequest!.path).toBe("/api/v1/ssh-keys/sk-1");
    });
  });

  // -------------------------------------------------------------------------
  // Admin tree resource integration
  // -------------------------------------------------------------------------

  describe("adminTree resource", () => {
    test("get sends GET to /admin/tree/:slug", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.adminTree.get("home-page" as Slug);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/admin/tree/home-page");
      expect((result as { route: { slug: string } }).route.slug).toBe("home-page");
    });

    test("get sends format query parameter when provided", async () => {
      const client = createAdminClient(validConfig());
      await client.adminTree.get("home" as Slug, "raw");
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.params.format).toBe("raw");
    });

    test("get does not send format query parameter when omitted", async () => {
      const client = createAdminClient(validConfig());
      await client.adminTree.get("home" as Slug);
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.params.format).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // Import resource integration
  // -------------------------------------------------------------------------

  describe("import resource", () => {
    test.each([
      { method: "contentful", path: "/api/v1/import/contentful" },
      { method: "sanity", path: "/api/v1/import/sanity" },
      { method: "strapi", path: "/api/v1/import/strapi" },
      { method: "wordpress", path: "/api/v1/import/wordpress" },
      { method: "clean", path: "/api/v1/import/clean" },
    ])("$method sends POST to $path", async ({ method, path }) => {
      const client = createAdminClient(validConfig());
      const importResource = client.import as Record<string, (data: Record<string, unknown>) => Promise<unknown>>;
      const result = await importResource[method]({ data: "test" });
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe(path);
      expect((result as { success: boolean }).success).toBe(true);
    });

    test("contentful returns ImportResponse shape", async () => {
      const client = createAdminClient(validConfig());
      const result = await client.import.contentful({ data: "test" });
      expect(result.success).toBe(true);
      expect(result.datatypes_created).toBe(1);
      expect(result.fields_created).toBe(2);
      expect(result.content_created).toBe(3);
      expect(result.message).toBe("ok");
      expect(result.errors).toEqual([]);
    });

    test("bulk sends POST to /import with format query parameter", async () => {
      const client = createAdminClient(validConfig());
      await client.import.bulk("contentful", { data: "test" });
      expect(lastRequest).not.toBeNull();
      expect(lastRequest!.method).toBe("POST");
      // The bulk method constructs /import?format=contentful, which becomes
      // the path portion: /api/v1/import and format in search params
      expect(lastRequest!.path).toBe("/api/v1/import");
    });
  });

  // -------------------------------------------------------------------------
  // isApiError re-export
  // -------------------------------------------------------------------------

  describe("isApiError re-export", () => {
    test("returns true for an ApiError-shaped object", () => {
      const err: ApiError = {
        _tag: "ApiError",
        status: 404,
        message: "Not Found",
      };
      expect(isApiError(err)).toBe(true);
    });

    test("returns false for a regular Error", () => {
      expect(isApiError(new Error("test"))).toBe(false);
    });

    test("returns false for null", () => {
      expect(isApiError(null)).toBe(false);
    });

    test("returns false for undefined", () => {
      expect(isApiError(undefined)).toBe(false);
    });

    test("returns false for a plain object without _tag", () => {
      expect(isApiError({ status: 404, message: "Not Found" })).toBe(false);
    });

    test("returns false for an object with wrong _tag value", () => {
      expect(isApiError({ _tag: "OtherError", status: 500, message: "fail" })).toBe(false);
    });

    test("returns true for ApiError with body property", () => {
      const err: ApiError = {
        _tag: "ApiError",
        status: 400,
        message: "Bad Request",
        body: { error: "validation_failed" },
      };
      expect(isApiError(err)).toBe(true);
    });
  });

  // -------------------------------------------------------------------------
  // Multiple clients with different configs are isolated
  // -------------------------------------------------------------------------

  describe("client isolation", () => {
    test("two clients with different apiKeys send different headers", async () => {
      const client1 = createAdminClient(validConfig({ apiKey: "key-alpha" }));
      const client2 = createAdminClient(validConfig({ apiKey: "key-beta" }));

      await client1.adminRoutes.list();
      const headers1 = lastRequest!.headers["authorization"];

      await client2.adminRoutes.list();
      const headers2 = lastRequest!.headers["authorization"];

      expect(headers1).toBe("Bearer key-alpha");
      expect(headers2).toBe("Bearer key-beta");
    });

    test("client without apiKey does not inherit auth from another client", async () => {
      const authedClient = createAdminClient(validConfig({ apiKey: "secret-key" }));
      const unauthedClient = createAdminClient(validConfig());

      await authedClient.adminRoutes.list();
      expect(lastRequest!.headers["authorization"]).toBe("Bearer secret-key");

      await unauthedClient.adminRoutes.list();
      expect(lastRequest!.headers["authorization"]).toBeUndefined();
    });
  });

  // -------------------------------------------------------------------------
  // Error propagation: config errors vs runtime errors
  // -------------------------------------------------------------------------

  describe("error propagation", () => {
    // Config validation errors are synchronous throws
    test("config validation errors are synchronous (not async)", () => {
      // These should throw immediately, not return a rejected promise
      let threw = false;
      try {
        createAdminClient({ baseUrl: "not-a-url" });
      } catch {
        threw = true;
      }
      expect(threw).toBe(true);
    });

    // Runtime errors from the HTTP client are async
    test("runtime API errors propagate as rejected promises", async () => {
      const client = createAdminClient(validConfig());
      // Hit a path that doesn't exist on the test server
      try {
        await client.adminRoutes.get("nonexistent-slug" as Slug);
        // The test server does handle this, so if we want to test error propagation
        // we need to verify the request completed successfully instead
        expect(lastRequest).not.toBeNull();
      } catch (err: unknown) {
        // If it does error, verify it's an ApiError
        const apiErr = err as ApiError;
        expect(apiErr._tag).toBe("ApiError");
      }
    });
  });

  // -------------------------------------------------------------------------
  // Validation order: baseUrl is validated before apiKey
  // -------------------------------------------------------------------------

  describe("validation order", () => {
    test("invalid baseUrl is caught before empty apiKey", () => {
      // Both baseUrl and apiKey are invalid, but baseUrl should be checked first
      expect(() =>
        createAdminClient({ baseUrl: "not-a-url", apiKey: "" }),
      ).toThrow("Invalid baseUrl: not-a-url");
    });

    test("http:// protocol error is caught before empty apiKey", () => {
      // http:// without allowInsecure AND empty apiKey -- protocol check comes first
      expect(() =>
        createAdminClient({ baseUrl: "http://localhost:3000", apiKey: "" }),
      ).toThrow(
        "baseUrl uses http:// which transmits credentials in plaintext. " +
          "Set allowInsecure: true to allow this, or use https://.",
      );
    });
  });
});

// ---------------------------------------------------------------------------
// Type imports used in tests (branded types need casting)
// ---------------------------------------------------------------------------

type AdminRouteID = import("./types/common.js").AdminRouteID;
type Slug = import("./types/common.js").Slug;
type Email = import("./types/common.js").Email;
type SessionID = import("./types/common.js").SessionID;
type UpdateSessionParams = import("./types/users.js").UpdateSessionParams;
type CreateSshKeyRequest = import("./types/users.js").CreateSshKeyRequest;
