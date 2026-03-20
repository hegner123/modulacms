/**
 * Escape HTML special characters to prevent XSS in template output.
 */
export function escapeHtml(str: string): string {
  let result = "";
  for (let i = 0; i < str.length; i++) {
    const ch = str[i];
    switch (ch) {
      case "&":
        result += "&amp;";
        break;
      case "<":
        result += "&lt;";
        break;
      case ">":
        result += "&gt;";
        break;
      case '"':
        result += "&quot;";
        break;
      case "'":
        result += "&#39;";
        break;
      default:
        result += ch;
    }
  }
  return result;
}

/**
 * Tagged template literal for building HTML strings.
 * Interpolated values are escaped; raw template portions pass through unchanged.
 */
export function html(
  strings: TemplateStringsArray,
  ...values: unknown[]
): string {
  let result = strings[0] ?? "";
  for (let i = 0; i < values.length; i++) {
    const val = values[i];
    if (val !== null && val !== undefined) {
      result += escapeHtml(String(val));
    }
    result += strings[i + 1] ?? "";
  }
  return result;
}

/**
 * Attach an open shadow root to a host element, inject CSS and HTML content.
 */
export function createShadowRoot(
  host: HTMLElement,
  css: string,
  htmlContent: string,
): ShadowRoot {
  const shadow = host.attachShadow({ mode: "open" });
  shadow.innerHTML = `<style>${css}</style>${htmlContent}`;
  return shadow;
}

/** querySelector shorthand. */
export function $(
  root: ShadowRoot | HTMLElement,
  selector: string,
): Element | null {
  return root.querySelector(selector);
}

/** querySelectorAll as array. */
export function $$(
  root: ShadowRoot | HTMLElement,
  selector: string,
): Element[] {
  return Array.from(root.querySelectorAll(selector));
}
