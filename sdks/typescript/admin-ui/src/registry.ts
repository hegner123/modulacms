import type { FieldType } from "@modulacms/types";

/**
 * Context passed to every field renderer when it needs to render.
 * Contains the field definition, current value, and a callback
 * the renderer calls when the user changes the value.
 */
export interface FieldRenderContext {
  /** Machine-readable field name (used as form key). */
  name: string;
  /** Human-readable label. */
  label: string;
  /** The field type key this renderer was matched on. */
  type: string;
  /** Current field value, or empty string if unset. */
  value: string;
  /** Whether the field is required. */
  required: boolean;
  /** Whether the field is read-only. */
  disabled: boolean;
  /** Parsed validation rules from the field definition. */
  validation: Record<string, unknown>;
  /** Parsed UI config from the field definition. */
  uiConfig: Record<string, unknown>;
  /** Parsed extra data from the field definition. */
  data: Record<string, unknown>;
  /** Call this when the user changes the value. */
  onChange: (value: string) => void;
}

/**
 * A field renderer is a custom element constructor (class) that
 * extends HTMLElement. When the field-renderer component encounters
 * a field type, it looks up the registered renderer, instantiates it,
 * and calls setContext() to pass in the field data.
 *
 * Renderers own their own DOM — they can use shadow DOM, light DOM,
 * or anything else. The only contract is the setContext method.
 */
export interface McmsFieldRenderer extends HTMLElement {
  /** Called by the field-renderer host to pass field data. */
  setContext(ctx: FieldRenderContext): void;
}

/**
 * Constructor type for field renderer custom elements.
 */
export interface McmsFieldRendererConstructor {
  new (): McmsFieldRenderer;
}

/**
 * Registration entry — either a custom element tag name (string)
 * or a constructor that will be auto-registered.
 */
export type FieldRendererEntry = string | McmsFieldRendererConstructor;

const renderers = new Map<string, FieldRendererEntry>();
let tagCounter = 0;

/**
 * Register a custom field renderer for a given type key.
 *
 * Accepts either:
 * - A string tag name of an already-registered custom element
 * - A class constructor that extends HTMLElement (auto-registered)
 *
 * @example
 * // Register by tag name (element already defined elsewhere)
 * registerFieldRenderer('color', 'my-color-picker');
 *
 * @example
 * // Register by class (auto-registers the custom element)
 * class MyColorPicker extends HTMLElement {
 *   setContext(ctx: FieldRenderContext) { ... }
 * }
 * registerFieldRenderer('color', MyColorPicker);
 *
 * @example
 * // Override a built-in renderer
 * registerFieldRenderer('richtext', MyBetterRichtext);
 */
export function registerFieldRenderer(
  type: FieldType | (string & {}),
  renderer: FieldRendererEntry,
): void {
  if (typeof renderer === "function" && !isRegisteredElement(renderer)) {
    const tag = `mcms-field-${type}-${++tagCounter}`;
    customElements.define(tag, renderer);
    renderers.set(type, tag);
    return;
  }
  renderers.set(type, renderer);
}

/**
 * Register multiple field renderers at once.
 *
 * @example
 * // declare-style registration (like Jotai atoms)
 * registerFieldRenderers({
 *   'color': MyColorPicker,
 *   'rating': 'my-rating-stars',
 *   'richtext': MyBetterRichtext,  // overrides built-in
 * });
 */
export function registerFieldRenderers(
  entries: Record<string, FieldRendererEntry>,
): void {
  for (const [type, renderer] of Object.entries(entries)) {
    registerFieldRenderer(type, renderer);
  }
}

/**
 * Look up the tag name for a registered field renderer.
 * Returns undefined if no renderer is registered for this type.
 */
export function getFieldRenderer(type: string): string | undefined {
  const entry = renderers.get(type);
  if (entry === undefined) return undefined;
  if (typeof entry === "string") return entry;
  // Constructor was registered — its tag should already be in the map
  // as a string from registerFieldRenderer, but handle the edge case.
  return undefined;
}

/**
 * Check if a type has a registered renderer (built-in or custom).
 */
export function hasFieldRenderer(type: string): boolean {
  return renderers.has(type);
}

/**
 * Remove a registered renderer. Useful for testing or
 * replacing renderers at runtime.
 */
export function unregisterFieldRenderer(type: string): boolean {
  return renderers.delete(type);
}

/**
 * List all registered field type keys.
 */
export function registeredFieldTypes(): string[] {
  return Array.from(renderers.keys());
}

function isRegisteredElement(ctor: McmsFieldRendererConstructor): boolean {
  for (const entry of renderers.values()) {
    if (typeof entry === "string") {
      const existing = customElements.get(entry);
      if (existing === ctor) return true;
    }
  }
  return false;
}
