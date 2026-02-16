import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createPluginRoutesResource } from "./plugin-routes.js";
import type { PluginRoutesResource } from "./plugin-routes.js";
import { createHttpClient } from "../http.js";
import type { HttpClient } from "../http.js";
import type { ApiError } from "../types/common.js";
import type { RouteApprovalItem } from "../types/plugins.js";

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const fakeRoutes = [
  { plugin: "analytics", method: "GET", path: "/analytics/dashboard", public: false, approved: true },
  { plugin: "seo", method: "POST", path: "/seo/audit", public: true, approved: false },
];

const approvalItems: RouteApprovalItem[] = [
  { plugin: "seo", method: "POST", path: "/seo/audit" },
];

// ---------------------------------------------------------------------------
// Test server
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

    // -- list routes --
    if (path === "/api/v1/admin/plugins/routes" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ routes: fakeRoutes }));
      return;
    }

    // -- approve routes --
    if (path === "/api/v1/admin/plugins/routes/approve" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ ok: true }));
      return;
    }

    // -- revoke routes --
    if (path === "/api/v1/admin/plugins/routes/revoke" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ ok: true }));
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

function makeClient(): HttpClient {
  return createHttpClient({
    baseUrl,
    defaultTimeout: 10000,
    credentials: "omit",
  });
}

function makePluginRoutes(): PluginRoutesResource {
  return createPluginRoutesResource(makeClient());
}

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

describe("createPluginRoutesResource", () => {
  describe("factory", () => {
    test("returns an object with list, approve, revoke", () => {
      const routes = makePluginRoutes();
      expect(typeof routes.list).toBe("function");
      expect(typeof routes.approve).toBe("function");
      expect(typeof routes.revoke).toBe("function");
    });
  });

  describe("list", () => {
    test("returns unwrapped route list", async () => {
      const routes = makePluginRoutes();
      const result = await routes.list();
      expect(result).toEqual(fakeRoutes);
    });

    test("sends GET to /admin/plugins/routes", async () => {
      const routes = makePluginRoutes();
      await routes.list();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/routes");
    });
  });

  describe("approve", () => {
    test("resolves void on success", async () => {
      const routes = makePluginRoutes();
      const result = await routes.approve(approvalItems);
      expect(result).toBeUndefined();
    });

    test("sends POST to /admin/plugins/routes/approve with body", async () => {
      const routes = makePluginRoutes();
      await routes.approve(approvalItems);
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/routes/approve");
      expect(lastRequest!.body).toEqual({ routes: approvalItems });
    });
  });

  describe("revoke", () => {
    test("resolves void on success", async () => {
      const routes = makePluginRoutes();
      const result = await routes.revoke(approvalItems);
      expect(result).toBeUndefined();
    });

    test("sends POST to /admin/plugins/routes/revoke with body", async () => {
      const routes = makePluginRoutes();
      await routes.revoke(approvalItems);
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/routes/revoke");
      expect(lastRequest!.body).toEqual({ routes: approvalItems });
    });
  });

  describe("error propagation", () => {
    test("throws ApiError on 403 response", async () => {
      const { server: errorServer, port: errorPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        res.writeHead(403, "Forbidden", { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "forbidden" }));
      });

      try {
        const errHttp = createHttpClient({
          baseUrl: `http://localhost:${errorPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const routes = createPluginRoutesResource(errHttp);

        try {
          await routes.list();
          expect(true).toBe(false);
        } catch (err: unknown) {
          const apiErr = err as ApiError;
          expect(apiErr._tag).toBe("ApiError");
          expect(apiErr.status).toBe(403);
          expect(apiErr.body).toEqual({ error: "forbidden" });
        }
      } finally {
        await closeServer(errorServer);
      }
    });
  });

  describe("abort signal", () => {
    test("aborts immediately with pre-aborted signal", async () => {
      const routes = makePluginRoutes();
      const controller = new AbortController();
      controller.abort();

      try {
        await routes.list({ signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });
});
