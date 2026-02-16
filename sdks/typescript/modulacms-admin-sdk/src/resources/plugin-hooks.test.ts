import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createPluginHooksResource } from "./plugin-hooks.js";
import type { PluginHooksResource } from "./plugin-hooks.js";
import { createHttpClient } from "../http.js";
import type { HttpClient } from "../http.js";
import type { ApiError } from "../types/common.js";
import type { HookApprovalItem } from "../types/plugins.js";

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const fakeHooks = [
  { plugin_name: "analytics", event: "after_create", table: "content_data", priority: 10, approved: true, is_wildcard: false },
  { plugin_name: "seo", event: "before_update", table: "*", priority: 5, approved: false, is_wildcard: true },
];

const approvalItems: HookApprovalItem[] = [
  { plugin: "seo", event: "before_update", table: "*" },
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

    // -- list hooks --
    if (path === "/api/v1/admin/plugins/hooks" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ hooks: fakeHooks }));
      return;
    }

    // -- approve hooks --
    if (path === "/api/v1/admin/plugins/hooks/approve" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ ok: true }));
      return;
    }

    // -- revoke hooks --
    if (path === "/api/v1/admin/plugins/hooks/revoke" && req.method === "POST") {
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

function makePluginHooks(): PluginHooksResource {
  return createPluginHooksResource(makeClient());
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

describe("createPluginHooksResource", () => {
  describe("factory", () => {
    test("returns an object with list, approve, revoke", () => {
      const hooks = makePluginHooks();
      expect(typeof hooks.list).toBe("function");
      expect(typeof hooks.approve).toBe("function");
      expect(typeof hooks.revoke).toBe("function");
    });
  });

  describe("list", () => {
    test("returns unwrapped hook list", async () => {
      const hooks = makePluginHooks();
      const result = await hooks.list();
      expect(result).toEqual(fakeHooks);
    });

    test("sends GET to /admin/plugins/hooks", async () => {
      const hooks = makePluginHooks();
      await hooks.list();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/hooks");
    });
  });

  describe("approve", () => {
    test("resolves void on success", async () => {
      const hooks = makePluginHooks();
      const result = await hooks.approve(approvalItems);
      expect(result).toBeUndefined();
    });

    test("sends POST to /admin/plugins/hooks/approve with body", async () => {
      const hooks = makePluginHooks();
      await hooks.approve(approvalItems);
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/hooks/approve");
      expect(lastRequest!.body).toEqual({ hooks: approvalItems });
    });
  });

  describe("revoke", () => {
    test("resolves void on success", async () => {
      const hooks = makePluginHooks();
      const result = await hooks.revoke(approvalItems);
      expect(result).toBeUndefined();
    });

    test("sends POST to /admin/plugins/hooks/revoke with body", async () => {
      const hooks = makePluginHooks();
      await hooks.revoke(approvalItems);
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/hooks/revoke");
      expect(lastRequest!.body).toEqual({ hooks: approvalItems });
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
        const hooks = createPluginHooksResource(errHttp);

        try {
          await hooks.list();
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
      const hooks = makePluginHooks();
      const controller = new AbortController();
      controller.abort();

      try {
        await hooks.list({ signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });
});
