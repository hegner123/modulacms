import type {
  FormDefinition,
  FormFieldDefinition,
  FieldType,
} from "../types.js";
import { FormsApiClient } from "../api-client.js";
import { createShadowRoot, html, escapeHtml, $, $$ } from "../utils/dom.js";
import { dispatch } from "../utils/events.js";
import { initDragAndDrop } from "../utils/drag.js";
import { CSS_CUSTOM_PROPERTIES } from "../utils/styles.js";
import { baseStyles } from "../styles/base.css.js";
import { builderStyles } from "../styles/builder.css.js";

// ---------------------------------------------------------------------------
// Field type metadata
// ---------------------------------------------------------------------------

interface FieldTypeMeta {
  label: string;
  icon: string;
}

const FIELD_TYPES: ReadonlyMap<FieldType, FieldTypeMeta> = new Map<FieldType, FieldTypeMeta>([
  ["text", { label: "Text", icon: "T" }],
  ["textarea", { label: "Textarea", icon: "\u00B6" }],
  ["email", { label: "Email", icon: "@" }],
  ["number", { label: "Number", icon: "#" }],
  ["tel", { label: "Phone", icon: "\u260E" }],
  ["url", { label: "URL", icon: "\u29C9" }],
  ["date", { label: "Date", icon: "\u25A3" }],
  ["time", { label: "Time", icon: "\u25F7" }],
  ["datetime", { label: "Date/Time", icon: "\u25A3" }],
  ["select", { label: "Select", icon: "\u25BE" }],
  ["radio", { label: "Radio", icon: "\u25C9" }],
  ["checkbox", { label: "Checkbox", icon: "\u2611" }],
  ["hidden", { label: "Hidden", icon: "\u2205" }],
  ["file", { label: "File", icon: "\u2191" }],
]);

const FIELD_TYPE_LIST: FieldType[] = [
  "text", "textarea", "email", "number", "tel", "url",
  "date", "time", "datetime", "select", "radio", "checkbox",
  "hidden", "file",
];

/** Debounce helper. Returns a wrapped function and a cancel handle. */
function debounce(fn: () => void, ms: number): { call: () => void; cancel: () => void } {
  let timer: ReturnType<typeof setTimeout> | null = null;
  return {
    call(): void {
      if (timer !== null) {
        clearTimeout(timer);
      }
      timer = setTimeout(() => {
        timer = null;
        fn();
      }, ms);
    },
    cancel(): void {
      if (timer !== null) {
        clearTimeout(timer);
        timer = null;
      }
    },
  };
}

// ---------------------------------------------------------------------------
// Tracked field: internal representation with dirty/new/deleted flags
// ---------------------------------------------------------------------------

interface TrackedField {
  definition: FormFieldDefinition;
  isNew: boolean;
  isDirty: boolean;
  isDeleted: boolean;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export class ModulaCMSFormBuilder extends HTMLElement {
  static observedAttributes = ["form-id", "api-url", "api-key", "auto-save"];

  // --- State ---
  private fields: TrackedField[] = [];
  private formDefinition: FormDefinition | null = null;
  private formVersion = 0;
  private selectedIndex = -1;
  private fieldCounter = 0;
  private loading = false;
  private errorMessage: string | null = null;
  private saving = false;

  // --- Infrastructure ---
  private shadow: ShadowRoot | null = null;
  private client: FormsApiClient | null = null;
  private destroyDrag: (() => void) | null = null;
  private autoSaveDebounce: { call: () => void; cancel: () => void } | null = null;

  // =========================================================================
  // Lifecycle
  // =========================================================================

  connectedCallback(): void {
    this.shadow = createShadowRoot(
      this,
      CSS_CUSTOM_PROPERTIES + baseStyles + builderStyles,
      "",
    );
    this.initClient();

    const formId = this.getAttribute("form-id");
    if (formId) {
      this.loadForm(formId);
    } else {
      this.render();
    }
  }

  disconnectedCallback(): void {
    this.teardownDrag();
    if (this.autoSaveDebounce) {
      this.autoSaveDebounce.cancel();
    }
  }

  attributeChangedCallback(name: string, oldValue: string | null, newValue: string | null): void {
    if (oldValue === newValue) {
      return;
    }

    if (name === "api-url" || name === "api-key") {
      this.initClient();
    }

    if (name === "form-id" && newValue && this.shadow) {
      this.loadForm(newValue);
    }

    if (name === "auto-save") {
      this.setupAutoSave();
    }
  }

  // =========================================================================
  // Public API
  // =========================================================================

  async save(): Promise<void> {
    if (!this.client || !this.formDefinition) {
      return;
    }

    this.saving = true;
    this.renderSaveButton();

    try {
      const formId = this.formDefinition.id;

      // Process deletions
      const deleted = this.fields.filter((t) => t.isDeleted && !t.isNew);
      for (const tracked of deleted) {
        await this.client.deleteField(tracked.definition.id, this.formVersion);
      }

      // Process new fields
      const created = this.fields.filter((t) => t.isNew && !t.isDeleted);
      for (const tracked of created) {
        const result = await this.client.createField(formId, {
          ...tracked.definition,
          version: this.formVersion,
        });
        tracked.definition.id = result.id;
        tracked.isNew = false;
        tracked.isDirty = false;
      }

      // Process updated fields
      const updated = this.fields.filter((t) => t.isDirty && !t.isNew && !t.isDeleted);
      for (const tracked of updated) {
        await this.client.updateField(tracked.definition.id, {
          ...tracked.definition,
          version: this.formVersion,
        });
        tracked.isDirty = false;
      }

      // Reorder (always send current order after creates/updates)
      const liveFields = this.fields.filter((t) => !t.isDeleted);
      if (liveFields.length > 0) {
        const ids = liveFields.map((t) => t.definition.id);
        const reorderResult = await this.client.reorderFields(formId, ids, this.formVersion);
        this.formVersion = reorderResult.version;
      }

      // Remove deleted entries from local state
      this.fields = this.fields.filter((t) => !t.isDeleted);

      dispatch(this, "builder:save", { fields: this.getDefinition() });
    } finally {
      this.saving = false;
      this.renderSaveButton();
    }
  }

  addField(type: FieldType): void {
    this.fieldCounter++;
    const position = this.fields.filter((t) => !t.isDeleted).length;

    const field: FormFieldDefinition = {
      id: "",
      form_id: this.formDefinition?.id ?? "",
      name: `field_${this.fieldCounter}`,
      label: capitalize(type),
      field_type: type,
      placeholder: null,
      default_value: null,
      help_text: null,
      required: false,
      validation_rules: null,
      options: hasOptions(type) ? ["Option 1"] : null,
      position,
      config: null,
    };

    const tracked: TrackedField = {
      definition: field,
      isNew: true,
      isDirty: false,
      isDeleted: false,
    };

    this.fields.push(tracked);
    this.selectedIndex = this.liveFields().length - 1;

    dispatch(this, "field:add", { field });
    this.onChange();
    this.render();
  }

  removeField(index: number): void {
    const live = this.liveFields();
    if (index < 0 || index >= live.length) {
      return;
    }

    const tracked = live[index] as TrackedField | undefined;
    if (!tracked) {
      return;
    }
    tracked.isDeleted = true;

    // Adjust selection
    if (this.selectedIndex === index) {
      this.selectedIndex = -1;
    } else if (this.selectedIndex > index) {
      this.selectedIndex--;
    }

    dispatch(this, "field:remove", { index });
    this.onChange();
    this.render();
  }

  moveField(from: number, to: number): void {
    const live = this.liveFields();
    if (from < 0 || from >= live.length || to < 0 || to >= live.length) {
      return;
    }
    if (from === to) {
      return;
    }

    // Reorder the underlying fields array.
    // Find the TrackedField references for live items, reorder them,
    // then rebuild the fields array preserving deleted items.
    const reordered = [...live];
    const spliced = reordered.splice(from, 1);
    const moved = spliced[0];
    if (!moved) {
      return;
    }
    reordered.splice(to, 0, moved);

    // Rebuild: keep deleted items, replace live ordering
    const deletedItems = this.fields.filter((t) => t.isDeleted);
    this.fields = [...reordered, ...deletedItems];

    // Update positions
    reordered.forEach((t, i) => {
      t.definition.position = i;
      if (!t.isNew) {
        t.isDirty = true;
      }
    });

    // Track selection
    if (this.selectedIndex === from) {
      this.selectedIndex = to;
    } else if (from < this.selectedIndex && to >= this.selectedIndex) {
      this.selectedIndex--;
    } else if (from > this.selectedIndex && to <= this.selectedIndex) {
      this.selectedIndex++;
    }

    dispatch(this, "field:reorder", { fromIndex: from, toIndex: to });
    this.onChange();
    this.render();
  }

  getDefinition(): FormFieldDefinition[] {
    return this.liveFields().map((t) => ({ ...t.definition }));
  }

  setDefinition(fields: FormFieldDefinition[]): void {
    this.fields = fields.map((f, i) => ({
      definition: { ...f, position: i },
      isNew: !f.id,
      isDirty: false,
      isDeleted: false,
    }));
    this.fieldCounter = fields.length;
    this.selectedIndex = -1;
    this.render();
  }

  // =========================================================================
  // Internal helpers
  // =========================================================================

  private liveFields(): TrackedField[] {
    return this.fields.filter((t) => !t.isDeleted);
  }

  private initClient(): void {
    const apiUrl = this.getAttribute("api-url");
    if (!apiUrl) {
      this.client = null;
      return;
    }
    const apiKey = this.getAttribute("api-key") ?? undefined;
    this.client = new FormsApiClient(apiUrl, apiKey);
  }

  private setupAutoSave(): void {
    if (this.autoSaveDebounce) {
      this.autoSaveDebounce.cancel();
      this.autoSaveDebounce = null;
    }

    if (this.hasAttribute("auto-save")) {
      this.autoSaveDebounce = debounce(() => {
        this.save();
      }, 2000);
    }
  }

  private async loadForm(formId: string): Promise<void> {
    if (!this.client) {
      this.errorMessage = "API URL is required to load a form";
      this.render();
      return;
    }

    this.loading = true;
    this.errorMessage = null;
    this.render();

    try {
      const result = await this.client.getForm(formId);
      this.formDefinition = result.form;
      this.formVersion = result.form.version;
      this.fields = result.fields
        .sort((a, b) => a.position - b.position)
        .map((f) => ({
          definition: f,
          isNew: false,
          isDirty: false,
          isDeleted: false,
        }));
      this.fieldCounter = this.fields.length;
      this.selectedIndex = -1;

      dispatch(this, "builder:loaded", {
        form: result.form,
        fields: result.fields,
      });

      this.setupAutoSave();
    } catch (err) {
      this.errorMessage = err instanceof Error ? err.message : "Failed to load form";
    } finally {
      this.loading = false;
      this.render();
    }
  }

  private onChange(): void {
    dispatch(this, "builder:change", { fields: this.getDefinition() });
    if (this.autoSaveDebounce) {
      this.autoSaveDebounce.call();
    }
  }

  private teardownDrag(): void {
    if (this.destroyDrag) {
      this.destroyDrag();
      this.destroyDrag = null;
    }
  }

  // =========================================================================
  // Rendering
  // =========================================================================

  private render(): void {
    if (!this.shadow) {
      return;
    }

    // Teardown previous drag listeners before replacing DOM
    this.teardownDrag();

    // Keep the <style> intact, replace the content wrapper
    let wrapper = this.shadow.querySelector(".mcms-builder-root");
    if (!wrapper) {
      wrapper = document.createElement("div");
      wrapper.classList.add("mcms-builder-root");
      this.shadow.appendChild(wrapper);
    }

    if (this.loading) {
      wrapper.innerHTML = this.renderLoading();
      return;
    }

    if (this.errorMessage) {
      wrapper.innerHTML = this.renderError();
      this.bindErrorEvents(wrapper as HTMLElement);
      return;
    }

    wrapper.innerHTML = this.renderBuilder();
    this.bindBuilderEvents(wrapper as HTMLElement);
    this.initDrag(wrapper as HTMLElement);
  }

  private renderLoading(): string {
    return `<div class="mcms-builder" part="builder">
      <div class="loading-state" part="loading">Loading form...</div>
    </div>`;
  }

  private renderError(): string {
    return `<div class="mcms-builder" part="builder">
      <div class="error-state" part="error-state">
        <div class="error-message">${escapeHtml(this.errorMessage ?? "Unknown error")}</div>
        <button class="retry-btn" type="button">Retry</button>
      </div>
    </div>`;
  }

  private renderBuilder(): string {
    const live = this.liveFields();

    return `<div class="mcms-builder" part="builder">
      <div class="mcms-builder-toolbar" part="toolbar">
        <span class="mcms-builder-toolbar-title">${escapeHtml(this.formDefinition?.name ?? "Form Builder")}</span>
        <div class="mcms-builder-toolbar-actions">
          <button
            class="mcms-builder-btn mcms-builder-btn--primary js-save-btn"
            type="button"
            part="save-button"
            ${this.saving ? "disabled" : ""}
          >${this.saving ? "Saving..." : "Save"}</button>
        </div>
      </div>
      <div class="mcms-builder-body" style="display:grid;grid-template-columns:14rem 1fr;overflow:hidden;">
        ${this.renderPalette()}
        ${live.length === 0 ? this.renderEmptyCanvas() : this.renderCanvas(live)}
      </div>
    </div>`;
  }

  private renderPalette(): string {
    const items = FIELD_TYPE_LIST.map((type) => {
      const meta = FIELD_TYPES.get(type);
      if (!meta) {
        return "";
      }
      return `<button
        class="mcms-palette-item"
        type="button"
        part="field-type-button"
        data-field-type="${type}"
        aria-label="Add ${meta.label} field"
      >
        <span class="mcms-palette-icon" aria-hidden="true">${escapeHtml(meta.icon)}</span>
        ${escapeHtml(meta.label)}
      </button>`;
    }).join("");

    return `<div class="mcms-palette" part="field-palette" role="toolbar" aria-label="Field types" aria-orientation="vertical">
      <p class="mcms-palette-title">Field Types</p>
      <div class="mcms-palette-grid" role="group">
        ${items}
      </div>
    </div>`;
  }

  private renderEmptyCanvas(): string {
    return `<div class="mcms-canvas mcms-canvas--empty" part="canvas">
      <span>Click a field type to add it</span>
    </div>`;
  }

  private renderCanvas(live: TrackedField[]): string {
    const cards = live.map((tracked, index) => this.renderFieldCard(tracked, index)).join("");
    return `<div class="mcms-canvas js-canvas" part="canvas">
      ${cards}
    </div>`;
  }

  private renderFieldCard(tracked: TrackedField, index: number): string {
    const f = tracked.definition;
    const selected = index === this.selectedIndex;
    const meta = FIELD_TYPES.get(f.field_type);
    const typeName = meta?.label ?? f.field_type;

    const handleDots = `<span class="mcms-drag-handle-row" aria-hidden="true">
      <span class="mcms-drag-handle-dot"></span><span class="mcms-drag-handle-dot"></span>
    </span>`.repeat(3);

    return `<div
      class="mcms-field-card${selected ? " mcms-field-card--selected" : ""}"
      part="field-item"
      data-index="${index}"
      role="listitem"
    >
      <div class="mcms-drag-handle js-drag-handle" part="field-handle" aria-label="Drag to reorder" tabindex="0">
        ${handleDots}
      </div>
      <div class="mcms-field-card-body js-field-toggle" data-index="${index}">
        <div class="mcms-field-card-label">${escapeHtml(f.label || f.name)}</div>
        <div class="mcms-field-card-type">
          <span>${escapeHtml(typeName)}</span>${f.required ? ` <span style="color:var(--modulacms-error-color,#dc2626);font-weight:600;">*</span>` : ""}
        </div>
      </div>
      <div class="mcms-field-card-actions">
        <button
          class="mcms-field-card-btn mcms-field-card-btn--delete js-delete-field"
          type="button"
          data-index="${index}"
          aria-label="Remove ${escapeHtml(f.label || f.name)} field"
        >&times;</button>
      </div>
      ${selected ? this.renderFieldConfig(tracked, index) : ""}
    </div>`;
  }

  private renderFieldConfig(tracked: TrackedField, index: number): string {
    const f = tracked.definition;
    const showOptions = hasOptions(f.field_type);

    let optionsHtml = "";
    if (showOptions) {
      const opts = f.options ?? [];
      const rows = opts.map((opt, oi) => `<div class="option-row">
        <input
          class="option-input js-option-input"
          type="text"
          value="${escapeHtml(opt)}"
          data-field-index="${index}"
          data-option-index="${oi}"
          aria-label="Option ${oi + 1}"
        />
        <button
          class="option-remove-btn js-option-remove"
          type="button"
          data-field-index="${index}"
          data-option-index="${oi}"
          aria-label="Remove option ${oi + 1}"
        >&times;</button>
      </div>`).join("");

      optionsHtml = `<div class="mcms-field-config-row" style="grid-column:1/-1;">
        <span class="mcms-field-config-label">Options</span>
        <div class="options-editor">
          ${rows}
          <button
            class="option-add-btn js-option-add"
            type="button"
            data-field-index="${index}"
          >+ Add option</button>
        </div>
      </div>`;
    }

    const rules = f.validation_rules;
    const minLen = rules?.min_length ?? "";
    const maxLen = rules?.max_length ?? "";

    return `<div class="mcms-field-config" part="field-config" style="display:flex;flex-direction:column;gap:0.75rem;grid-column:1/-1;padding-top:0.75rem;margin-top:0.5rem;border-top:1px solid #e5e7eb;">
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem;">
        <div class="mcms-field-config-row">
          <label class="mcms-field-config-label">Name</label>
          <input
            class="mcms-field-config-input js-config-input"
            type="text"
            value="${escapeHtml(f.name)}"
            data-field-index="${index}"
            data-config-key="name"
            part="config-input"
          />
        </div>
        <div class="mcms-field-config-row">
          <label class="mcms-field-config-label">Label</label>
          <input
            class="mcms-field-config-input js-config-input"
            type="text"
            value="${escapeHtml(f.label)}"
            data-field-index="${index}"
            data-config-key="label"
            part="config-input"
          />
        </div>
        <div class="mcms-field-config-row">
          <label class="mcms-field-config-label">Type</label>
          <select
            class="mcms-field-config-input js-config-input"
            data-field-index="${index}"
            data-config-key="field_type"
            part="config-input"
          >
            ${FIELD_TYPE_LIST.map((t) => {
              const m = FIELD_TYPES.get(t);
              const label = m?.label ?? t;
              const sel = t === f.field_type ? " selected" : "";
              return `<option value="${t}"${sel}>${escapeHtml(label)}</option>`;
            }).join("")}
          </select>
        </div>
        <div class="mcms-field-config-row">
          <label class="mcms-field-config-label">Placeholder</label>
          <input
            class="mcms-field-config-input js-config-input"
            type="text"
            value="${escapeHtml(f.placeholder ?? "")}"
            data-field-index="${index}"
            data-config-key="placeholder"
            part="config-input"
          />
        </div>
        <div class="mcms-field-config-row" style="grid-column:1/-1;">
          <label class="mcms-field-config-label">Help text</label>
          <input
            class="mcms-field-config-input js-config-input"
            type="text"
            value="${escapeHtml(f.help_text ?? "")}"
            data-field-index="${index}"
            data-config-key="help_text"
            part="config-input"
          />
        </div>
        <div class="mcms-field-config-row">
          <div class="mcms-field-config-checkbox">
            <input
              type="checkbox"
              class="js-config-checkbox"
              data-field-index="${index}"
              data-config-key="required"
              ${f.required ? "checked" : ""}
              id="req-${index}"
            />
            <label for="req-${index}" class="mcms-field-config-label" style="margin:0;">Required</label>
          </div>
        </div>
      </div>
      ${optionsHtml}
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:0.5rem;padding-top:0.5rem;border-top:1px solid #f3f4f6;">
        <div class="mcms-field-config-row">
          <label class="mcms-field-config-label">Min length</label>
          <input
            class="mcms-field-config-input js-validation-input"
            type="number"
            min="0"
            value="${minLen}"
            data-field-index="${index}"
            data-rule-key="min_length"
            part="config-input"
          />
        </div>
        <div class="mcms-field-config-row">
          <label class="mcms-field-config-label">Max length</label>
          <input
            class="mcms-field-config-input js-validation-input"
            type="number"
            min="0"
            value="${maxLen}"
            data-field-index="${index}"
            data-rule-key="max_length"
            part="config-input"
          />
        </div>
      </div>
    </div>`;
  }

  // =========================================================================
  // Event binding
  // =========================================================================

  private bindErrorEvents(root: HTMLElement): void {
    const retryBtn = root.querySelector(".retry-btn");
    if (retryBtn) {
      retryBtn.addEventListener("click", () => {
        const formId = this.getAttribute("form-id");
        if (formId) {
          this.loadForm(formId);
        }
      });
    }
  }

  private bindBuilderEvents(root: HTMLElement): void {
    // Save button
    const saveBtn = root.querySelector(".js-save-btn");
    if (saveBtn) {
      saveBtn.addEventListener("click", () => {
        this.save();
      });
    }

    // Palette: add field on click, Enter, or Space
    for (const btn of Array.from(root.querySelectorAll(".mcms-palette-item"))) {
      btn.addEventListener("click", () => {
        const type = (btn as HTMLElement).dataset.fieldType as FieldType;
        if (type) {
          this.addField(type);
        }
      });
      btn.addEventListener("keydown", (e: Event) => {
        const ke = e as KeyboardEvent;
        if (ke.key === "Enter" || ke.key === " ") {
          ke.preventDefault();
          const type = (btn as HTMLElement).dataset.fieldType as FieldType;
          if (type) {
            this.addField(type);
          }
        }
      });
    }

    // Palette: arrow key navigation
    this.bindPaletteArrowKeys(root);

    // Field toggle (expand/collapse config)
    for (const toggle of Array.from(root.querySelectorAll(".js-field-toggle"))) {
      toggle.addEventListener("click", () => {
        const idx = parseInt((toggle as HTMLElement).dataset.index ?? "-1", 10);
        this.toggleFieldSelection(idx);
      });
    }

    // Delete field
    for (const btn of Array.from(root.querySelectorAll(".js-delete-field"))) {
      btn.addEventListener("click", (e: Event) => {
        e.stopPropagation();
        const idx = parseInt((btn as HTMLElement).dataset.index ?? "-1", 10);
        this.removeField(idx);
      });
    }

    // Config inputs (text, select)
    for (const input of Array.from(root.querySelectorAll(".js-config-input"))) {
      input.addEventListener("input", () => {
        this.handleConfigInput(input as HTMLInputElement | HTMLSelectElement);
      });
      input.addEventListener("change", () => {
        this.handleConfigInput(input as HTMLInputElement | HTMLSelectElement);
      });
    }

    // Config checkboxes
    for (const cb of Array.from(root.querySelectorAll(".js-config-checkbox"))) {
      cb.addEventListener("change", () => {
        this.handleConfigCheckbox(cb as HTMLInputElement);
      });
    }

    // Validation inputs
    for (const input of Array.from(root.querySelectorAll(".js-validation-input"))) {
      input.addEventListener("input", () => {
        this.handleValidationInput(input as HTMLInputElement);
      });
    }

    // Option inputs
    for (const input of Array.from(root.querySelectorAll(".js-option-input"))) {
      input.addEventListener("input", () => {
        this.handleOptionInput(input as HTMLInputElement);
      });
    }

    // Option remove buttons
    for (const btn of Array.from(root.querySelectorAll(".js-option-remove"))) {
      btn.addEventListener("click", () => {
        this.handleOptionRemove(btn as HTMLElement);
      });
    }

    // Option add buttons
    for (const btn of Array.from(root.querySelectorAll(".js-option-add"))) {
      btn.addEventListener("click", () => {
        this.handleOptionAdd(btn as HTMLElement);
      });
    }
  }

  private bindPaletteArrowKeys(root: HTMLElement): void {
    const grid = root.querySelector(".mcms-palette-grid");
    if (!grid) {
      return;
    }

    grid.addEventListener("keydown", (e: Event) => {
      const ke = e as KeyboardEvent;
      const items = Array.from(grid.querySelectorAll(".mcms-palette-item")) as HTMLElement[];
      const current = (ke.target as HTMLElement).closest(".mcms-palette-item") as HTMLElement | null;
      if (!current) {
        return;
      }

      const currentIndex = items.indexOf(current);
      if (currentIndex < 0) {
        return;
      }

      // Grid is 2 columns
      const cols = 2;
      let nextIndex = -1;

      switch (ke.key) {
        case "ArrowRight":
          nextIndex = currentIndex + 1;
          break;
        case "ArrowLeft":
          nextIndex = currentIndex - 1;
          break;
        case "ArrowDown":
          nextIndex = currentIndex + cols;
          break;
        case "ArrowUp":
          nextIndex = currentIndex - cols;
          break;
        default:
          return;
      }

      if (nextIndex >= 0 && nextIndex < items.length) {
        ke.preventDefault();
        const nextItem = items[nextIndex] as HTMLElement | undefined;
        if (nextItem) {
          nextItem.focus();
        }
      }
    });
  }

  private initDrag(root: HTMLElement): void {
    const canvas = root.querySelector(".js-canvas") as HTMLElement | null;
    if (!canvas) {
      return;
    }

    this.destroyDrag = initDragAndDrop(canvas, {
      handleSelector: ".js-drag-handle",
      onReorder: (from, to) => {
        this.moveField(from, to);
      },
    });
  }

  // =========================================================================
  // Config change handlers
  // =========================================================================

  private toggleFieldSelection(index: number): void {
    if (this.selectedIndex === index) {
      this.selectedIndex = -1;
    } else {
      this.selectedIndex = index;
    }
    this.render();
  }

  private handleConfigInput(input: HTMLInputElement | HTMLSelectElement): void {
    const fieldIndex = parseInt(input.dataset.fieldIndex ?? "-1", 10);
    const key = input.dataset.configKey;
    if (fieldIndex < 0 || !key) {
      return;
    }

    const live = this.liveFields();
    if (fieldIndex >= live.length) {
      return;
    }

    const tracked = live[fieldIndex] as TrackedField | undefined;
    if (!tracked) {
      return;
    }
    const def = tracked.definition;
    const value = input.value;

    switch (key) {
      case "name":
        def.name = value;
        break;
      case "label":
        def.label = value;
        break;
      case "field_type":
        def.field_type = value as FieldType;
        // Add default options when switching to select/radio
        if (hasOptions(def.field_type) && (!def.options || def.options.length === 0)) {
          def.options = ["Option 1"];
        }
        break;
      case "placeholder":
        def.placeholder = value || null;
        break;
      case "help_text":
        def.help_text = value || null;
        break;
    }

    tracked.isDirty = true;
    this.onChange();

    // Re-render for field_type changes (shows/hides options panel)
    if (key === "field_type") {
      this.render();
    }
  }

  private handleConfigCheckbox(input: HTMLInputElement): void {
    const fieldIndex = parseInt(input.dataset.fieldIndex ?? "-1", 10);
    const key = input.dataset.configKey;
    if (fieldIndex < 0 || !key) {
      return;
    }

    const live = this.liveFields();
    if (fieldIndex >= live.length) {
      return;
    }

    const tracked = live[fieldIndex] as TrackedField | undefined;
    if (!tracked) {
      return;
    }
    if (key === "required") {
      tracked.definition.required = input.checked;
    }
    tracked.isDirty = true;
    this.onChange();
  }

  private handleValidationInput(input: HTMLInputElement): void {
    const fieldIndex = parseInt(input.dataset.fieldIndex ?? "-1", 10);
    const ruleKey = input.dataset.ruleKey;
    if (fieldIndex < 0 || !ruleKey) {
      return;
    }

    const live = this.liveFields();
    if (fieldIndex >= live.length) {
      return;
    }

    const tracked = live[fieldIndex] as TrackedField | undefined;
    if (!tracked) {
      return;
    }
    const def = tracked.definition;

    if (!def.validation_rules) {
      def.validation_rules = {};
    }

    const numValue = input.value ? parseInt(input.value, 10) : undefined;

    if (ruleKey === "min_length") {
      if (numValue !== undefined && Number.isFinite(numValue)) {
        def.validation_rules.min_length = numValue;
      } else {
        delete def.validation_rules.min_length;
      }
    } else if (ruleKey === "max_length") {
      if (numValue !== undefined && Number.isFinite(numValue)) {
        def.validation_rules.max_length = numValue;
      } else {
        delete def.validation_rules.max_length;
      }
    }

    // Clear validation_rules if empty
    if (
      def.validation_rules.min_length === undefined &&
      def.validation_rules.max_length === undefined &&
      def.validation_rules.max_file_size === undefined
    ) {
      def.validation_rules = null;
    }

    tracked.isDirty = true;
    this.onChange();
  }

  private handleOptionInput(input: HTMLInputElement): void {
    const fieldIndex = parseInt(input.dataset.fieldIndex ?? "-1", 10);
    const optionIndex = parseInt(input.dataset.optionIndex ?? "-1", 10);
    if (fieldIndex < 0 || optionIndex < 0) {
      return;
    }

    const live = this.liveFields();
    if (fieldIndex >= live.length) {
      return;
    }

    const tracked = live[fieldIndex] as TrackedField | undefined;
    if (!tracked) {
      return;
    }
    if (!tracked.definition.options) {
      return;
    }

    tracked.definition.options[optionIndex] = input.value;
    tracked.isDirty = true;
    this.onChange();
  }

  private handleOptionRemove(btn: HTMLElement): void {
    const fieldIndex = parseInt(btn.dataset.fieldIndex ?? "-1", 10);
    const optionIndex = parseInt(btn.dataset.optionIndex ?? "-1", 10);
    if (fieldIndex < 0 || optionIndex < 0) {
      return;
    }

    const live = this.liveFields();
    if (fieldIndex >= live.length) {
      return;
    }

    const tracked = live[fieldIndex] as TrackedField | undefined;
    if (!tracked) {
      return;
    }
    if (!tracked.definition.options) {
      return;
    }

    tracked.definition.options.splice(optionIndex, 1);
    tracked.isDirty = true;
    this.onChange();
    this.render();
  }

  private handleOptionAdd(btn: HTMLElement): void {
    const fieldIndex = parseInt(btn.dataset.fieldIndex ?? "-1", 10);
    if (fieldIndex < 0) {
      return;
    }

    const live = this.liveFields();
    if (fieldIndex >= live.length) {
      return;
    }

    const tracked = live[fieldIndex] as TrackedField | undefined;
    if (!tracked) {
      return;
    }
    if (!tracked.definition.options) {
      tracked.definition.options = [];
    }

    const nextNum = tracked.definition.options.length + 1;
    tracked.definition.options.push(`Option ${nextNum}`);
    tracked.isDirty = true;
    this.onChange();
    this.render();
  }

  private renderSaveButton(): void {
    if (!this.shadow) {
      return;
    }
    const btn = this.shadow.querySelector(".js-save-btn") as HTMLButtonElement | null;
    if (!btn) {
      return;
    }
    btn.disabled = this.saving;
    btn.textContent = this.saving ? "Saving..." : "Save";
  }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function capitalize(str: string): string {
  if (str.length === 0) {
    return str;
  }
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function hasOptions(type: FieldType): boolean {
  return type === "select" || type === "radio";
}
