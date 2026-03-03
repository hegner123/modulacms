import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { ModulaClient } from "../src/client.js";
import { ModulaError } from "../src/errors.js";

function mockFetch(body: unknown, init?: { status?: number; ok?: boolean }) {
  const status = init?.status ?? 200;
  const ok = init?.ok ?? (status >= 200 && status < 300);
  return vi.fn().mockResolvedValue({
    ok,
    status,
    json: () => Promise.resolve(body),
    text: () => Promise.resolve(JSON.stringify(body)),
  });
}

describe("ModulaClient", () => {
  const originalFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = originalFetch;
    vi.restoreAllMocks();
  });

  describe("constructor", () => {
    it("accepts a valid URL", () => {
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      expect(client).toBeInstanceOf(ModulaClient);
    });

    it("strips trailing slashes from baseUrl", () => {
      globalThis.fetch = mockFetch([]);
      const client = new ModulaClient({ baseUrl: "https://example.com///" });

      client.listRoutes();

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/routes");
    });

    it("preserves base path segments", () => {
      globalThis.fetch = mockFetch([]);
      const client = new ModulaClient({ baseUrl: "https://example.com/cms" });

      client.listRoutes();

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/cms/api/v1/routes");
    });

    it("throws ModulaError for invalid URL", () => {
      expect(() => new ModulaClient({ baseUrl: "not-a-url" })).toThrow(ModulaError);

      try {
        new ModulaClient({ baseUrl: "not-a-url" });
      } catch (err) {
        expect(err).toBeInstanceOf(ModulaError);
        expect((err as ModulaError).status).toBe(0);
      }
    });
  });

  describe("request headers", () => {
    it("sends Accept: application/json", async () => {
      globalThis.fetch = mockFetch([]);
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await client.listRoutes();

      const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      const headers = fetchCall[1].headers as Record<string, string>;
      expect(headers["Accept"]).toBe("application/json");
    });

    it("sends Authorization header when apiKey is provided", async () => {
      globalThis.fetch = mockFetch([]);
      const client = new ModulaClient({
        baseUrl: "https://example.com",
        apiKey: "test-key",
      });

      await client.listRoutes();

      const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      const headers = fetchCall[1].headers as Record<string, string>;
      expect(headers["Authorization"]).toBe("Bearer test-key");
    });

    it("does not send Authorization header when no apiKey", async () => {
      globalThis.fetch = mockFetch([]);
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await client.listRoutes();

      const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      const headers = fetchCall[1].headers as Record<string, string>;
      expect(headers["Authorization"]).toBeUndefined();
    });

    it("passes credentials option to fetch", async () => {
      globalThis.fetch = mockFetch([]);
      const client = new ModulaClient({
        baseUrl: "https://example.com",
        credentials: "include",
      });

      await client.listRoutes();

      const fetchCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0];
      expect(fetchCall[1].credentials).toBe("include");
    });
  });

  describe("error handling", () => {
    it("throws ModulaError on non-2xx response with JSON body", async () => {
      globalThis.fetch = mockFetch({ error: "Not found" }, { status: 404, ok: false });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await expect(client.listRoutes()).rejects.toThrow(ModulaError);

      try {
        await client.listRoutes();
      } catch (err) {
        expect(err).toBeInstanceOf(ModulaError);
        expect((err as ModulaError).status).toBe(404);
        expect((err as ModulaError).message).toBe("Not found");
      }
    });

    it("falls back to text body when JSON parsing fails", async () => {
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
        json: () => Promise.reject(new Error("invalid json")),
        text: () => Promise.resolve("Internal Server Error"),
      });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      try {
        await client.listRoutes();
      } catch (err) {
        expect(err).toBeInstanceOf(ModulaError);
        expect((err as ModulaError).status).toBe(500);
        expect((err as ModulaError).body).toBe("Internal Server Error");
      }
    });
  });

  describe("getPage", () => {
    it("requests the correct URL for a slug", async () => {
      globalThis.fetch = mockFetch({ title: "About" });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await client.getPage("about");

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/content/about");
    });

    it("appends format query parameter", async () => {
      globalThis.fetch = mockFetch({ title: "About" });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await client.getPage("about", { format: "contentful" });

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/content/about?format=contentful");
    });

    it("uses defaultFormat when no per-call format is specified", async () => {
      globalThis.fetch = mockFetch({ title: "About" });
      const client = new ModulaClient({
        baseUrl: "https://example.com",
        defaultFormat: "clean",
      });

      await client.getPage("about");

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/content/about?format=clean");
    });

    it("per-call format overrides defaultFormat", async () => {
      globalThis.fetch = mockFetch({ title: "About" });
      const client = new ModulaClient({
        baseUrl: "https://example.com",
        defaultFormat: "clean",
      });

      await client.getPage("about", { format: "raw" });

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/content/about?format=raw");
    });

    it("throws on empty slug", async () => {
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await expect(client.getPage("")).rejects.toThrow(ModulaError);
    });

    it("accepts slug starting with /", async () => {
      globalThis.fetch = mockFetch({ title: "About" });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await client.getPage("/about");

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/content/about");
    });

    it("requests root path for / slug", async () => {
      globalThis.fetch = mockFetch({ title: "Home" });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      await client.getPage("/");

      const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/content/");
    });

    it("runs validator and returns data when valid", async () => {
      globalThis.fetch = mockFetch({ title: "About" });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      interface Page {
        title: string;
      }
      const isPage = (data: unknown): data is Page =>
        typeof data === "object" && data !== null && "title" in data;

      const result = await client.getPage("about", { validate: isPage });

      expect(result.title).toBe("About");
    });

    it("throws when validator rejects the response", async () => {
      globalThis.fetch = mockFetch({ wrong: "shape" });
      const client = new ModulaClient({ baseUrl: "https://example.com" });

      const alwaysFalse = (_data: unknown): _data is never => false;

      await expect(
        client.getPage("about", { validate: alwaysFalse }),
      ).rejects.toThrow(ModulaError);
    });
  });

  describe("API methods build correct URLs", () => {
    let client: ModulaClient;

    beforeEach(() => {
      globalThis.fetch = mockFetch([]);
      client = new ModulaClient({ baseUrl: "https://example.com" });
    });

    const cases: Array<{ method: keyof ModulaClient; args: string[]; expectedUrl: string }> = [
      { method: "listRoutes", args: [], expectedUrl: "https://example.com/api/v1/routes" },
      { method: "getRoute", args: ["01ABC"], expectedUrl: "https://example.com/api/v1/routes/?q=01ABC" },
      { method: "listContentData", args: [], expectedUrl: "https://example.com/api/v1/contentdata" },
      { method: "getContentData", args: ["01ABC"], expectedUrl: "https://example.com/api/v1/contentdata/?q=01ABC" },
      { method: "listContentFields", args: [], expectedUrl: "https://example.com/api/v1/contentfields" },
      { method: "getContentField", args: ["01ABC"], expectedUrl: "https://example.com/api/v1/contentfields/?q=01ABC" },
      { method: "listMedia", args: [], expectedUrl: "https://example.com/api/v1/media" },
      { method: "getMedia", args: ["01ABC"], expectedUrl: "https://example.com/api/v1/media/?q=01ABC" },
      { method: "listMediaDimensions", args: [], expectedUrl: "https://example.com/api/v1/mediadimensions" },
      { method: "getMediaDimension", args: ["01ABC"], expectedUrl: "https://example.com/api/v1/mediadimensions/?q=01ABC" },
      { method: "listDatatypes", args: [], expectedUrl: "https://example.com/api/v1/datatype" },
      { method: "getDatatype", args: ["01ABC"], expectedUrl: "https://example.com/api/v1/datatype/?q=01ABC" },
      { method: "listFields", args: [], expectedUrl: "https://example.com/api/v1/fields" },
      { method: "getField", args: ["01ABC"], expectedUrl: "https://example.com/api/v1/fields/?q=01ABC" },
    ];

    for (const { method, args, expectedUrl } of cases) {
      it(`${method}(${args.join(", ")}) -> ${expectedUrl}`, async () => {
        const fn = client[method] as (...a: string[]) => Promise<unknown>;
        await fn.call(client, ...args);

        const calledUrl = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
        expect(calledUrl).toBe(expectedUrl);
      });
    }
  });
});
