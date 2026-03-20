import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { FormsApiClient } from "../api-client.js";

describe("FormsApiClient", () => {
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    vi.stubGlobal("fetch", mockFetch);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function okResponse(body: unknown): Response {
    return {
      ok: true,
      status: 200,
      json: () => Promise.resolve(body),
    } as Response;
  }

  function errorResponse(status: number, body: unknown): Response {
    return {
      ok: false,
      status,
      json: () => Promise.resolve(body),
    } as Response;
  }

  // ---------------------------------------------------------------------------
  // URL construction
  // ---------------------------------------------------------------------------

  describe("URL construction", () => {
    it("strips trailing slash and appends base path", async () => {
      mockFetch.mockResolvedValueOnce(okResponse({ items: [], total: 0, limit: 20, offset: 0 }));
      const client = new FormsApiClient("https://example.com/", "key-1");

      await client.listForms();

      const calledUrl = mockFetch.mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/plugins/forms/forms");
    });

    it("works without trailing slash", async () => {
      mockFetch.mockResolvedValueOnce(okResponse({ items: [], total: 0, limit: 20, offset: 0 }));
      const client = new FormsApiClient("https://example.com", "key-1");

      await client.listForms();

      const calledUrl = mockFetch.mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/plugins/forms/forms");
    });
  });

  // ---------------------------------------------------------------------------
  // Headers
  // ---------------------------------------------------------------------------

  describe("headers", () => {
    it("sets Content-Type and X-API-Key headers", async () => {
      mockFetch.mockResolvedValueOnce(okResponse({ items: [], total: 0, limit: 20, offset: 0 }));
      const client = new FormsApiClient("https://example.com", "secret-key");

      await client.listForms();

      const init = mockFetch.mock.calls[0][1] as RequestInit;
      const headers = init.headers as Record<string, string>;
      expect(headers["Content-Type"]).toBe("application/json");
      expect(headers["X-API-Key"]).toBe("secret-key");
    });

    it("omits X-API-Key when no apiKey provided", async () => {
      mockFetch.mockResolvedValueOnce(okResponse({ items: [], total: 0, limit: 20, offset: 0 }));
      const client = new FormsApiClient("https://example.com");

      await client.listForms();

      const init = mockFetch.mock.calls[0][1] as RequestInit;
      const headers = init.headers as Record<string, string>;
      expect(headers["Content-Type"]).toBe("application/json");
      expect(headers["X-API-Key"]).toBeUndefined();
    });
  });

  // ---------------------------------------------------------------------------
  // getPublicForm
  // ---------------------------------------------------------------------------

  describe("getPublicForm", () => {
    it("calls GET /public/{id}", async () => {
      const form = { id: "form-1", name: "Contact", fields: [] };
      mockFetch.mockResolvedValueOnce(okResponse(form));
      const client = new FormsApiClient("https://example.com", "key");

      const result = await client.getPublicForm("form-1");

      const calledUrl = mockFetch.mock.calls[0][0] as string;
      const init = mockFetch.mock.calls[0][1] as RequestInit;
      expect(calledUrl).toBe("https://example.com/api/v1/plugins/forms/public/form-1");
      expect(init.method).toBe("GET");
      expect(result).toEqual(form);
    });
  });

  // ---------------------------------------------------------------------------
  // submitForm
  // ---------------------------------------------------------------------------

  describe("submitForm", () => {
    it("calls POST /public/{id}/submit with body", async () => {
      const submitResult = { id: "entry-1", message: "ok", redirect_url: null };
      mockFetch.mockResolvedValueOnce(okResponse(submitResult));
      const client = new FormsApiClient("https://example.com", "key");
      const data = { name: "Alice", email: "alice@example.com" };

      const result = await client.submitForm("form-1", data);

      const calledUrl = mockFetch.mock.calls[0][0] as string;
      const init = mockFetch.mock.calls[0][1] as RequestInit;
      expect(calledUrl).toBe("https://example.com/api/v1/plugins/forms/public/form-1/submit");
      expect(init.method).toBe("POST");
      expect(init.body).toBe(JSON.stringify(data));
      expect(result).toEqual(submitResult);
    });
  });

  // ---------------------------------------------------------------------------
  // listForms
  // ---------------------------------------------------------------------------

  describe("listForms", () => {
    it("calls GET /forms with query params", async () => {
      const page = { items: [], total: 0, limit: 10, offset: 5 };
      mockFetch.mockResolvedValueOnce(okResponse(page));
      const client = new FormsApiClient("https://example.com", "key");

      const result = await client.listForms({ limit: 10, offset: 5 });

      const calledUrl = mockFetch.mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/plugins/forms/forms?limit=10&offset=5");
      expect(result).toEqual(page);
    });

    it("calls GET /forms without query params when none provided", async () => {
      const page = { items: [], total: 0, limit: 20, offset: 0 };
      mockFetch.mockResolvedValueOnce(okResponse(page));
      const client = new FormsApiClient("https://example.com", "key");

      await client.listForms();

      const calledUrl = mockFetch.mock.calls[0][0] as string;
      expect(calledUrl).toBe("https://example.com/api/v1/plugins/forms/forms");
    });
  });

  // ---------------------------------------------------------------------------
  // Error handling
  // ---------------------------------------------------------------------------

  describe("error handling", () => {
    it("throws with error message from response body", async () => {
      mockFetch.mockResolvedValueOnce(errorResponse(404, { error: "Form not found" }));
      const client = new FormsApiClient("https://example.com", "key");

      await expect(client.getPublicForm("missing")).rejects.toThrow("Form not found");
    });

    it("throws with HTTP status when body has no error field", async () => {
      mockFetch.mockResolvedValueOnce(errorResponse(500, {}));
      const client = new FormsApiClient("https://example.com", "key");

      await expect(client.listForms()).rejects.toThrow("HTTP 500");
    });

    it("throws with HTTP status when body is not JSON", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 502,
        json: () => Promise.reject(new Error("invalid json")),
      } as unknown as Response);
      const client = new FormsApiClient("https://example.com", "key");

      await expect(client.listForms()).rejects.toThrow("HTTP 502");
    });
  });
});
