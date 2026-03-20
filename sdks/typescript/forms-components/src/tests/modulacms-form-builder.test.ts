import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

let mockFetch: ReturnType<typeof vi.fn>;

beforeEach(() => {
  mockFetch = vi.fn();
  vi.stubGlobal("fetch", mockFetch);
});

afterEach(() => {
  vi.restoreAllMocks();
  document.body.innerHTML = "";
});

async function tick(ms = 50): Promise<void> {
  await new Promise((r) => setTimeout(r, ms));
}

interface BuilderElement extends HTMLElement {
  addField: (type: string) => void;
  removeField: (index: number) => void;
  moveField: (from: number, to: number) => void;
  getDefinition: () => Array<{ name: string; field_type: string }>;
}

describe("ModulaCMSFormBuilder", () => {
  it("element is defined in customElements registry", async () => {
    await import("../index.js");
    const ctor = customElements.get("modulacms-form-builder");
    expect(ctor).toBeDefined();
  });

  it("addField adds a field to the definition", async () => {
    await import("../index.js");
    const el = document.createElement("modulacms-form-builder") as BuilderElement;
    document.body.appendChild(el);
    await tick();

    const before = el.getDefinition().length;
    el.addField("email");
    await tick();
    const after = el.getDefinition().length;
    expect(after).toBe(before + 1);

    const last = el.getDefinition()[after - 1];
    expect(last?.field_type).toBe("email");
  });

  it("removeField removes a field by index", async () => {
    await import("../index.js");
    const el = document.createElement("modulacms-form-builder") as BuilderElement;
    document.body.appendChild(el);
    await tick();

    el.addField("text");
    el.addField("email");
    el.addField("number");
    await tick();
    expect(el.getDefinition().length).toBe(3);

    el.removeField(1);
    await tick();
    expect(el.getDefinition().length).toBe(2);
    expect(el.getDefinition()[0]?.field_type).toBe("text");
    expect(el.getDefinition()[1]?.field_type).toBe("number");
  });

  it("moveField reorders fields correctly", async () => {
    await import("../index.js");
    const el = document.createElement("modulacms-form-builder") as BuilderElement;
    document.body.appendChild(el);
    await tick();

    el.addField("text");
    el.addField("email");
    el.addField("number");
    await tick();

    el.moveField(2, 0);
    await tick();

    const defs = el.getDefinition();
    expect(defs[0]?.field_type).toBe("number");
    expect(defs[1]?.field_type).toBe("text");
    expect(defs[2]?.field_type).toBe("email");
  });

  it("dispatches builder:change on modifications", async () => {
    await import("../index.js");
    const el = document.createElement("modulacms-form-builder") as BuilderElement;
    document.body.appendChild(el);
    await tick();

    const changes: CustomEvent[] = [];
    el.addEventListener("builder:change", ((e: Event) => {
      changes.push(e as CustomEvent);
    }) as EventListener);

    el.addField("text");
    await tick();

    expect(changes.length).toBeGreaterThanOrEqual(1);
    expect(changes[0]!.type).toBe("builder:change");
  });
});
