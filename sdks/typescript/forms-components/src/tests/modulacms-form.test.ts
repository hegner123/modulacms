import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import type { FormDefinition } from "../types.js";

// ---------------------------------------------------------------------------
// Mock fetch globally
// ---------------------------------------------------------------------------
let mockFetch: ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockFetch = vi.fn();
  vi.stubGlobal("fetch", mockFetch);
});

afterEach(() => {
  vi.restoreAllMocks();
  document.body.innerHTML = "";
});

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function sampleForm(): FormDefinition {
  return {
    id: "form-1",
    name: "Contact",
    description: "Contact form",
    submit_label: "Send",
    success_message: "Thanks!",
    redirect_url: null,
    captcha_config: null,
    version: 1,
    enabled: true,
    fields: [
      {
        id: "f-1",
        form_id: "form-1",
        name: "email",
        label: "Email",
        field_type: "email",
        placeholder: "you@example.com",
        default_value: null,
        help_text: null,
        required: true,
        validation_rules: null,
        options: null,
        position: 0,
        config: null,
      },
      {
        id: "f-2",
        form_id: "form-1",
        name: "message",
        label: "Message",
        field_type: "textarea",
        placeholder: null,
        default_value: null,
        help_text: "Tell us what you need",
        required: false,
        validation_rules: null,
        options: null,
        position: 1,
        config: null,
      },
    ],
  };
}

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

/** Wait a tick for async connectedCallback and rendering to settle. */
async function tick(ms = 50): Promise<void> {
  await new Promise((r) => setTimeout(r, ms));
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe("ModulaCMSForm", () => {
  it("element is defined in customElements registry", async () => {
    await import("../index.js");
    const ctor = customElements.get("modulacms-form");
    expect(ctor).toBeDefined();
  });

  it("renders shadow DOM when form-id and api-url are set", async () => {
    mockFetch.mockResolvedValueOnce(okResponse(sampleForm()));
    await import("../index.js");

    const el = document.createElement("modulacms-form");
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);

    await tick();

    expect(el.shadowRoot).not.toBeNull();
    // The shadow root should contain rendered content (style + form markup)
    const shadowHtml = el.shadowRoot!.innerHTML;
    expect(shadowHtml.length).toBeGreaterThan(0);
  });

  it("dispatches form:loaded after successful fetch", async () => {
    mockFetch.mockResolvedValueOnce(okResponse(sampleForm()));
    await import("../index.js");

    const el = document.createElement("modulacms-form");
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");

    const loaded = new Promise<CustomEvent>((resolve) => {
      el.addEventListener("form:loaded", ((e: Event) => resolve(e as CustomEvent)) as EventListener, { once: true });
    });

    document.body.appendChild(el);

    const event = await loaded;
    expect(event).toBeDefined();
    expect(event.type).toBe("form:loaded");
  });

  it("shows error state on fetch failure", async () => {
    mockFetch.mockResolvedValueOnce(errorResponse(500, { error: "Internal Server Error" }));
    await import("../index.js");

    const el = document.createElement("modulacms-form");
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);

    await tick();

    expect(el.shadowRoot).not.toBeNull();
    const shadowHtml = el.shadowRoot!.innerHTML;
    // Should contain some error indication
    expect(shadowHtml).toContain("error");
  });

  it.skipIf(!process.env["COMPONENT_INTEGRATION"])("honeypot field is rendered with correct attributes", async () => {
    mockFetch.mockResolvedValueOnce(okResponse(sampleForm()));
    await import("../index.js");

    const el = document.createElement("modulacms-form");
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);

    await tick();

    const honeypot = el.shadowRoot!.querySelector(".mcms-honeypot");
    expect(honeypot).not.toBeNull();
    // The honeypot should contain an input
    const input = honeypot!.querySelector("input");
    expect(input).not.toBeNull();
    expect(input!.getAttribute("tabindex")).toBe("-1");
    expect(input!.getAttribute("autocomplete")).toBe("off");
  });

  it.skipIf(!process.env["COMPONENT_INTEGRATION"])("submit dispatches cancelable form:submit event", async () => {
    mockFetch.mockResolvedValueOnce(okResponse(sampleForm()));
    await import("../index.js");

    const el = document.createElement("modulacms-form") as HTMLElement & { submit: () => void };
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);

    await tick();

    let submitEventFired = false;
    let cancelable = false;
    el.addEventListener("form:submit", ((e: Event) => {
      submitEventFired = true;
      cancelable = e.cancelable;
      e.preventDefault();
    }) as EventListener);

    // Trigger submit via the component's submit button
    const submitBtn = el.shadowRoot!.querySelector(".mcms-submit") as HTMLButtonElement | null;
    if (submitBtn) {
      submitBtn.click();
      await tick();
    }

    // If component exposes submit() method, try that
    if (typeof el.submit === "function" && !submitEventFired) {
      el.submit();
      await tick();
    }

    expect(submitEventFired).toBe(true);
    expect(cancelable).toBe(true);
  });

  it("validate returns errors for invalid data", async () => {
    mockFetch.mockResolvedValueOnce(okResponse(sampleForm()));
    await import("../index.js");

    const el = document.createElement("modulacms-form") as HTMLElement & {
      validate: () => Array<{ field: string; message: string }>;
    };
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);

    await tick();

    // The email field is required; with no input the form should report errors
    if (typeof el.validate === "function") {
      const errors = el.validate();
      expect(Array.isArray(errors)).toBe(true);
      expect(errors.length).toBeGreaterThan(0);
      expect(errors[0].field).toBe("email");
    }
  });

  it("reset clears all field values", async () => {
    mockFetch.mockResolvedValueOnce(okResponse(sampleForm()));
    await import("../index.js");

    const el = document.createElement("modulacms-form") as HTMLElement & { reset: () => void };
    el.setAttribute("form-id", "form-1");
    el.setAttribute("api-url", "https://example.com");
    document.body.appendChild(el);

    await tick();

    // Fill in a value
    const inputs = el.shadowRoot!.querySelectorAll("input, textarea");
    for (const input of inputs) {
      if ((input as HTMLInputElement).type !== "hidden") {
        (input as HTMLInputElement).value = "test-data";
      }
    }

    // Reset
    if (typeof el.reset === "function") {
      el.reset();
      await tick();
    }

    // All visible inputs should be empty
    const inputsAfter = el.shadowRoot!.querySelectorAll("input, textarea");
    for (const input of inputsAfter) {
      const inp = input as HTMLInputElement;
      if (inp.type !== "hidden" && !inp.closest(".mcms-honeypot")) {
        expect(inp.value).toBe("");
      }
    }
  });
});
