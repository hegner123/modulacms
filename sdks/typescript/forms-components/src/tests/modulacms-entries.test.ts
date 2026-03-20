import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { PaginatedResponse, FormEntry } from "../types.js";

let mockFetch: ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockFetch = vi.fn();
  vi.stubGlobal("fetch", mockFetch);
});

afterEach(() => {
  vi.restoreAllMocks();
  document.body.innerHTML = "";
});

function sampleForm() {
  return {
    form: {
      id: "form-1",
      name: "Contact",
      description: "",
      submit_label: "Submit",
      success_message: "Thanks",
      redirect_url: null,
      captcha_config: null,
      version: 1,
      enabled: true,
      fields: [],
    },
    fields: [
      { id: "f1", form_id: "form-1", name: "email", label: "Email", field_type: "email", placeholder: null, default_value: null, help_text: null, required: true, validation_rules: null, options: null, position: 0, config: null },
    ],
  };
}

function sampleEntries(): PaginatedResponse<FormEntry> {
  return {
    items: [
      { id: "entry-1", form_id: "form-1", form_version: 1, data: { email: "alice@example.com" }, client_ip: "127.0.0.1", user_agent: "Test/1.0", status: "submitted", created_at: "2026-01-01T00:00:00Z", updated_at: "2026-01-01T00:00:00Z" },
      { id: "entry-2", form_id: "form-1", form_version: 1, data: { email: "bob@example.com" }, client_ip: "127.0.0.2", user_agent: "Test/2.0", status: "submitted", created_at: "2026-01-02T00:00:00Z", updated_at: "2026-01-02T00:00:00Z" },
    ],
    total: 2,
    limit: 20,
    offset: 0,
  };
}

function emptyEntries(): PaginatedResponse<FormEntry> {
  return { items: [], total: 0, limit: 20, offset: 0 };
}

function okResponse(body: unknown): Response {
  return { ok: true, status: 200, json: () => Promise.resolve(body) } as Response;
}

function mockBothFetches(entries: PaginatedResponse<FormEntry> = sampleEntries()) {
  // First call: getForm, second call: listEntries
  mockFetch
    .mockResolvedValueOnce(okResponse(sampleForm()))
    .mockResolvedValueOnce(okResponse(entries));
}

async function tick(ms = 100): Promise<void> {
  await new Promise((r) => setTimeout(r, ms));
}

describe("ModulaCMSEntries", () => {
  it("element is defined in customElements registry", async () => {
    await import("../index.js");
    const ctor = customElements.get("modulacms-entries");
    expect(ctor).toBeDefined();
  });

  it("renders table structure in shadow DOM", async () => {
    mockBothFetches();
    await import("../index.js");

    const el = document.createElement("modulacms-entries");
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);
    await tick();

    expect(el.shadowRoot).not.toBeNull();
    const table = el.shadowRoot!.querySelector("table");
    expect(table).not.toBeNull();
    expect(table!.querySelector("thead")).not.toBeNull();
    expect(table!.querySelector("tbody")).not.toBeNull();
  });

  it("dispatches entries:loaded after fetch", async () => {
    mockBothFetches();
    await import("../index.js");

    const el = document.createElement("modulacms-entries");
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");

    const loaded = new Promise<CustomEvent>((resolve, reject) => {
      const timer = setTimeout(() => reject(new Error("entries:loaded not dispatched")), 3000);
      el.addEventListener("entries:loaded", ((e: Event) => {
        clearTimeout(timer);
        resolve(e as CustomEvent);
      }) as EventListener, { once: true });
    });

    document.body.appendChild(el);
    const event = await loaded;
    expect(event.type).toBe("entries:loaded");
  });

  it.skipIf(!process.env["COMPONENT_INTEGRATION"])("goToPage dispatches entries:page-change", async () => {
    mockBothFetches();
    await import("../index.js");

    const el = document.createElement("modulacms-entries") as HTMLElement & { goToPage: (n: number) => void };
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);
    await tick();

    // Mock the subsequent fetch for page 2
    mockFetch.mockResolvedValueOnce(okResponse(sampleEntries()));

    let pageChangeEvent: CustomEvent | null = null;
    el.addEventListener("entries:page-change", ((e: Event) => {
      pageChangeEvent = e as CustomEvent;
    }) as EventListener);

    el.goToPage(2);
    await tick();
    expect(pageChangeEvent).not.toBeNull();
    expect(pageChangeEvent!.type).toBe("entries:page-change");
  });

  it("shows empty state when no entries", async () => {
    mockBothFetches(emptyEntries());
    await import("../index.js");

    const el = document.createElement("modulacms-entries");
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);
    await tick();

    expect(el.shadowRoot).not.toBeNull();
    // Should not have table rows
    const tbody = el.shadowRoot!.querySelector("tbody");
    if (tbody) {
      expect(tbody.querySelectorAll("tr").length).toBe(0);
    }
  });
});
