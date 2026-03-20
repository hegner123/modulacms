// Polyfill CSS.escape for jsdom (not available in jsdom)
if (typeof globalThis.CSS === "undefined") {
  (globalThis as unknown as Record<string, unknown>).CSS = {
    escape: (s: string) =>
      s.replace(/([^\w-])/g, "\\$1"),
    supports: () => false,
  };
} else if (typeof CSS.escape !== "function") {
  CSS.escape = (s: string) => s.replace(/([^\w-])/g, "\\$1");
}
