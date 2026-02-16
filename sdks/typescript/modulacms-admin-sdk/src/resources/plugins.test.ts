import { describe, test, expect, beforeAll, afterAll } from "vitest";
import http from "node:http";
import { createPluginsResource } from "./plugins.js";
import type { PluginsResource } from "./plugins.js";
import { createHttpClient } from "../http.js";
import type { HttpClient } from "../http.js";
import type { ApiError } from "../types/common.js";

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

const fakePluginList = [
  { name: "analytics", version: "1.0.0", description: "Analytics plugin", state: "active" },
  { name: "seo", version: "2.1.0", description: "SEO toolkit", state: "disabled", circuit_breaker_state: "closed" },
];

const fakePluginInfo = {
  name: "analytics",
  version: "1.0.0",
  description: "Analytics plugin",
  author: "ModulaCMS",
  state: "active",
  vms_available: 3,
  vms_total: 5,
  dependencies: ["core"],
};

const fakeReloadResponse = { ok: true as const, plugin: "analytics" };
const fakeEnableResponse = { ok: true as const, plugin: "analytics", state: "active" };
const fakeDisableResponse = { ok: true as const, plugin: "analytics", state: "disabled" };

const fakeCleanupDryRun = {
  orphaned_tables: ["plg_old_table"],
  count: 1,
  action: "dry_run" as const,
};

const fakeCleanupDrop = {
  dropped: ["plg_old_table"],
  count: 1,
};

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

    // -- list plugins --
    if (path === "/api/v1/admin/plugins" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ plugins: fakePluginList }));
      return;
    }

    // -- get plugin --
    if (path === "/api/v1/admin/plugins/analytics" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakePluginInfo));
      return;
    }

    // -- reload --
    if (path === "/api/v1/admin/plugins/analytics/reload" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeReloadResponse));
      return;
    }

    // -- enable --
    if (path === "/api/v1/admin/plugins/analytics/enable" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeEnableResponse));
      return;
    }

    // -- disable --
    if (path === "/api/v1/admin/plugins/analytics/disable" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeDisableResponse));
      return;
    }

    // -- cleanup dry-run --
    if (path === "/api/v1/admin/plugins/cleanup" && req.method === "GET") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeCleanupDryRun));
      return;
    }

    // -- cleanup drop --
    if (path === "/api/v1/admin/plugins/cleanup" && req.method === "POST") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify(fakeCleanupDrop));
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

function makePlugins(): PluginsResource {
  return createPluginsResource(makeClient());
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

describe("createPluginsResource", () => {
  describe("factory", () => {
    test("returns an object with all plugin methods", () => {
      const plugins = makePlugins();
      expect(typeof plugins.list).toBe("function");
      expect(typeof plugins.get).toBe("function");
      expect(typeof plugins.reload).toBe("function");
      expect(typeof plugins.enable).toBe("function");
      expect(typeof plugins.disable).toBe("function");
      expect(typeof plugins.cleanupDryRun).toBe("function");
      expect(typeof plugins.cleanupDrop).toBe("function");
    });
  });

  describe("list", () => {
    test("returns unwrapped plugin list", async () => {
      const plugins = makePlugins();
      const result = await plugins.list();
      expect(result).toEqual(fakePluginList);
    });

    test("sends GET to /admin/plugins", async () => {
      const plugins = makePlugins();
      await plugins.list();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins");
    });
  });

  describe("get", () => {
    test("returns plugin info", async () => {
      const plugins = makePlugins();
      const result = await plugins.get("analytics");
      expect(result).toEqual(fakePluginInfo);
    });

    test("sends GET to /admin/plugins/{name}", async () => {
      const plugins = makePlugins();
      await plugins.get("analytics");
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/analytics");
    });
  });

  describe("reload", () => {
    test("returns action response", async () => {
      const plugins = makePlugins();
      const result = await plugins.reload("analytics");
      expect(result).toEqual(fakeReloadResponse);
    });

    test("sends POST to /admin/plugins/{name}/reload", async () => {
      const plugins = makePlugins();
      await plugins.reload("analytics");
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/analytics/reload");
    });

    test("sends no body", async () => {
      const plugins = makePlugins();
      await plugins.reload("analytics");
      expect(lastRequest!.body).toBeNull();
    });
  });

  describe("enable", () => {
    test("returns state response", async () => {
      const plugins = makePlugins();
      const result = await plugins.enable("analytics");
      expect(result).toEqual(fakeEnableResponse);
    });

    test("sends POST to /admin/plugins/{name}/enable", async () => {
      const plugins = makePlugins();
      await plugins.enable("analytics");
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/analytics/enable");
    });
  });

  describe("disable", () => {
    test("returns state response", async () => {
      const plugins = makePlugins();
      const result = await plugins.disable("analytics");
      expect(result).toEqual(fakeDisableResponse);
    });

    test("sends POST to /admin/plugins/{name}/disable", async () => {
      const plugins = makePlugins();
      await plugins.disable("analytics");
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/analytics/disable");
    });
  });

  describe("cleanupDryRun", () => {
    test("returns dry-run response", async () => {
      const plugins = makePlugins();
      const result = await plugins.cleanupDryRun();
      expect(result).toEqual(fakeCleanupDryRun);
    });

    test("sends GET to /admin/plugins/cleanup", async () => {
      const plugins = makePlugins();
      await plugins.cleanupDryRun();
      expect(lastRequest!.method).toBe("GET");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/cleanup");
    });
  });

  describe("cleanupDrop", () => {
    test("returns drop response", async () => {
      const plugins = makePlugins();
      const result = await plugins.cleanupDrop({ confirm: true, tables: ["plg_old_table"] });
      expect(result).toEqual(fakeCleanupDrop);
    });

    test("sends POST to /admin/plugins/cleanup with body", async () => {
      const plugins = makePlugins();
      await plugins.cleanupDrop({ confirm: true, tables: ["plg_old_table"] });
      expect(lastRequest!.method).toBe("POST");
      expect(lastRequest!.path).toBe("/api/v1/admin/plugins/cleanup");
      expect(lastRequest!.body).toEqual({ confirm: true, tables: ["plg_old_table"] });
    });
  });

  describe("error propagation", () => {
    test("throws ApiError on 500 response", async () => {
      const { server: errorServer, port: errorPort } = await createInlineServer(async (req, res) => {
        let rawBody = ""; for await (const chunk of req) rawBody += chunk;
        res.writeHead(500, "Internal Server Error", { "Content-Type": "application/json" });
        res.end(JSON.stringify({ error: "plugin_error" }));
      });

      try {
        const errHttp = createHttpClient({
          baseUrl: `http://localhost:${errorPort}`,
          defaultTimeout: 10000,
          credentials: "omit",
        });
        const plugins = createPluginsResource(errHttp);

        try {
          await plugins.list();
          expect(true).toBe(false);
        } catch (err: unknown) {
          const apiErr = err as ApiError;
          expect(apiErr._tag).toBe("ApiError");
          expect(apiErr.status).toBe(500);
          expect(apiErr.body).toEqual({ error: "plugin_error" });
        }
      } finally {
        await closeServer(errorServer);
      }
    });
  });

  describe("abort signal", () => {
    test("aborts immediately with pre-aborted signal", async () => {
      const plugins = makePlugins();
      const controller = new AbortController();
      controller.abort();

      try {
        await plugins.list({ signal: controller.signal });
        expect(true).toBe(false);
      } catch (err: unknown) {
        expect(err).toBeInstanceOf(Error);
        const error = err as Error;
        expect(error.name).toBe("AbortError");
      }
    });
  });
});
