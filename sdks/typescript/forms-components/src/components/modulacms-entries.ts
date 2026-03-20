import type {
  FormDefinition,
  FormFieldDefinition,
  FormEntry,
  PaginatedResponse,
} from "../types.js";
import { FormsApiClient } from "../api-client.js";
import { createShadowRoot, html, escapeHtml, $ } from "../utils/dom.js";
import { dispatch } from "../utils/events.js";
import { CSS_CUSTOM_PROPERTIES } from "../utils/styles.js";
import { baseStyles } from "../styles/base.css.js";
import { entriesStyles } from "../styles/entries.css.js";

// ---------------------------------------------------------------------------
// Internal state types
// ---------------------------------------------------------------------------

interface EntriesState {
  loading: boolean;
  error: string | null;
  form: FormDefinition | null;
  fields: FormFieldDefinition[];
  entries: FormEntry[];
  total: number;
  page: number;
  pageSize: number;
  sortField: string | null;
  sortDir: "asc" | "desc";
  filters: Record<string, string>;
  selectedEntryId: string | null;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export class ModulaCMSEntries extends HTMLElement {
  static get observedAttributes(): string[] {
    return [
      "form-id",
      "api-url",
      "api-key",
      "page-size",
      "sortable",
      "filterable",
      "export-enabled",
    ];
  }

  private shadow!: ShadowRoot;
  private client: FormsApiClient | null = null;
  private state: EntriesState = {
    loading: false,
    error: null,
    form: null,
    fields: [],
    entries: [],
    total: 0,
    page: 1,
    pageSize: 20,
    sortField: null,
    sortDir: "asc",
    filters: {},
    selectedEntryId: null,
  };
  private abortController: AbortController | null = null;
  private filterDebounceTimer: ReturnType<typeof setTimeout> | null = null;

  // -----------------------------------------------------------------------
  // Lifecycle
  // -----------------------------------------------------------------------

  connectedCallback(): void {
    this.shadow = createShadowRoot(
      this,
      CSS_CUSTOM_PROPERTIES + baseStyles + entriesStyles,
      "",
    );

    const pageSizeAttr = this.getAttribute("page-size");
    if (pageSizeAttr !== null) {
      const parsed = parseInt(pageSizeAttr, 10);
      if (parsed > 0) {
        this.state.pageSize = parsed;
      }
    }

    this.initClient();
    if (this.client && this.getAttribute("form-id")) {
      this.loadFormAndEntries();
    } else {
      this.render();
    }
  }

  disconnectedCallback(): void {
    this.cancelPendingFetch();
    if (this.filterDebounceTimer !== null) {
      clearTimeout(this.filterDebounceTimer);
      this.filterDebounceTimer = null;
    }
  }

  attributeChangedCallback(
    name: string,
    oldValue: string | null,
    newValue: string | null,
  ): void {
    if (oldValue === newValue) return;

    if (name === "page-size" && newValue !== null) {
      const parsed = parseInt(newValue, 10);
      if (parsed > 0) {
        this.state.pageSize = parsed;
        this.state.page = 1;
      }
    }

    if (name === "api-url" || name === "api-key") {
      this.initClient();
    }

    // Re-fetch when form-id or api-url changes after connect
    if (!this.shadow) return;
    if (name === "form-id" || name === "api-url") {
      this.state.page = 1;
      this.state.filters = {};
      this.state.sortField = null;
      this.state.sortDir = "asc";
      this.state.selectedEntryId = null;
      if (this.client && this.getAttribute("form-id")) {
        this.loadFormAndEntries();
      } else {
        this.state.form = null;
        this.state.fields = [];
        this.state.entries = [];
        this.state.total = 0;
        this.render();
      }
    }

    if (
      name === "sortable" ||
      name === "filterable" ||
      name === "export-enabled"
    ) {
      this.render();
    }
  }

  // -----------------------------------------------------------------------
  // Public API
  // -----------------------------------------------------------------------

  refresh(): void {
    if (this.client && this.getAttribute("form-id")) {
      this.fetchEntries();
    }
  }

  goToPage(n: number): void {
    const totalPages = this.totalPages();
    if (n < 1 || n > totalPages) return;
    this.state.page = n;
    this.fetchEntries();
    dispatch(this, "entries:page-change", { page: n });
  }

  async exportEntries(_format?: string): Promise<void> {
    const formId = this.getAttribute("form-id");
    if (!this.client || !formId) return;

    const allEntries: FormEntry[] = [];
    let after: string | undefined;
    let hasMore = true;

    while (hasMore) {
      const batch = await this.client.exportEntries(formId, after);
      allEntries.push(...batch.items);
      hasMore = batch.has_more;
      if (batch.after) {
        after = batch.after;
      } else {
        hasMore = false;
      }
    }

    const json = JSON.stringify(allEntries, null, 2);
    const blob = new Blob([json], { type: "application/json" });
    const url = URL.createObjectURL(blob);

    const link = document.createElement("a");
    link.href = url;
    link.download = "entries-" + formId + ".json";
    link.style.display = "none";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);

    dispatch(this, "entries:export", { entries: allEntries });
  }

  setFilter(filter: Record<string, string>): void {
    this.state.filters = { ...filter };
    this.state.page = 1;
    this.fetchEntries();
  }

  clearFilters(): void {
    this.state.filters = {};
    this.state.page = 1;
    this.fetchEntries();
  }

  setSort(field: string, dir: "asc" | "desc"): void {
    this.state.sortField = field;
    this.state.sortDir = dir;
    this.state.page = 1;
    this.fetchEntries();
  }

  // -----------------------------------------------------------------------
  // Data fetching
  // -----------------------------------------------------------------------

  private initClient(): void {
    const apiUrl = this.getAttribute("api-url");
    const apiKey = this.getAttribute("api-key") ?? undefined;
    if (apiUrl) {
      this.client = new FormsApiClient(apiUrl, apiKey);
    }
  }

  private cancelPendingFetch(): void {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
  }

  private async loadFormAndEntries(): Promise<void> {
    const formId = this.getAttribute("form-id");
    if (!this.client || !formId) return;

    this.state.loading = true;
    this.state.error = null;
    this.render();

    try {
      const result = await this.client.getForm(formId);
      this.state.form = result.form;
      this.state.fields = result.fields;
      await this.fetchEntriesInternal();
    } catch (err) {
      this.state.loading = false;
      this.state.error =
        err instanceof Error ? err.message : "Failed to load form";
      this.render();
    }
  }

  private async fetchEntries(): Promise<void> {
    this.state.loading = true;
    this.state.error = null;
    this.render();

    try {
      await this.fetchEntriesInternal();
    } catch (err) {
      this.state.loading = false;
      this.state.error =
        err instanceof Error ? err.message : "Failed to load entries";
      this.render();
    }
  }

  private async fetchEntriesInternal(): Promise<void> {
    const formId = this.getAttribute("form-id");
    if (!this.client || !formId) return;

    this.cancelPendingFetch();
    this.abortController = new AbortController();

    const offset = (this.state.page - 1) * this.state.pageSize;
    const result: PaginatedResponse<FormEntry> =
      await this.client.listEntries(formId, {
        limit: this.state.pageSize,
        offset,
      });

    let entries = result.items;

    // Client-side filtering (the API may not support field-level filters)
    entries = this.applyFilters(entries);

    // Client-side sorting
    entries = this.applySorting(entries);

    this.state.entries = entries;
    this.state.total = result.total;
    this.state.loading = false;
    this.state.error = null;
    this.abortController = null;

    this.render();

    dispatch(this, "entries:loaded", {
      entries: this.state.entries,
      total: this.state.total,
    });
  }

  private applyFilters(entries: FormEntry[]): FormEntry[] {
    const filterKeys = Object.keys(this.state.filters);
    if (filterKeys.length === 0) return entries;

    return entries.filter((entry) => {
      for (const key of filterKeys) {
        const filterValue = (this.state.filters[key] ?? "").toLowerCase();
        if (!filterValue) continue;

        if (key === "status") {
          if (!entry.status.toLowerCase().includes(filterValue)) return false;
          continue;
        }

        const dataValue = entry.data[key];
        if (dataValue === null || dataValue === undefined) return false;
        if (!String(dataValue).toLowerCase().includes(filterValue))
          return false;
      }
      return true;
    });
  }

  private applySorting(entries: FormEntry[]): FormEntry[] {
    if (!this.state.sortField) return entries;

    const field = this.state.sortField;
    const dir = this.state.sortDir === "asc" ? 1 : -1;

    return [...entries].sort((a, b) => {
      let valA: unknown;
      let valB: unknown;

      if (field === "created_at") {
        valA = a.created_at;
        valB = b.created_at;
      } else if (field === "status") {
        valA = a.status;
        valB = b.status;
      } else {
        valA = a.data[field];
        valB = b.data[field];
      }

      if (valA === null || valA === undefined) return dir;
      if (valB === null || valB === undefined) return -dir;

      const strA = String(valA);
      const strB = String(valB);

      // Attempt numeric comparison
      const numA = Number(strA);
      const numB = Number(strB);
      if (!isNaN(numA) && !isNaN(numB)) {
        return (numA - numB) * dir;
      }

      return strA.localeCompare(strB) * dir;
    });
  }

  // -----------------------------------------------------------------------
  // Rendering
  // -----------------------------------------------------------------------

  private totalPages(): number {
    if (this.state.total <= 0) return 1;
    return Math.ceil(this.state.total / this.state.pageSize);
  }

  private render(): void {
    if (!this.shadow) return;

    // Preserve the <style> tag, replace everything after it
    const style = $(this.shadow, "style");
    const content = this.buildContent();

    // Remove existing content nodes (keep style)
    const children = Array.from(this.shadow.childNodes);
    for (const child of children) {
      if (child !== style) {
        this.shadow.removeChild(child);
      }
    }

    const wrapper = document.createElement("div");
    wrapper.innerHTML = content;
    while (wrapper.firstChild) {
      this.shadow.appendChild(wrapper.firstChild);
    }

    this.attachEventListeners();
  }

  private buildContent(): string {
    if (this.state.loading) {
      return this.renderLoading();
    }

    if (this.state.error) {
      return this.renderError();
    }

    if (!this.state.form) {
      return this.renderEmpty("No form selected");
    }

    const parts: string[] = [];
    parts.push('<div class="mcms-entries" role="region" aria-label="Form entries">');

    // Toolbar (filters + export)
    if (this.hasAttribute("filterable") || this.hasAttribute("export-enabled")) {
      parts.push(this.renderToolbar());
    }

    // Table or empty state
    if (this.state.entries.length === 0) {
      parts.push(this.renderEmpty("No entries yet"));
    } else {
      parts.push(this.renderTable());
    }

    // Pagination
    if (this.state.total > this.state.pageSize) {
      parts.push(this.renderPagination());
    }

    parts.push("</div>");
    return parts.join("");
  }

  private renderLoading(): string {
    return [
      '<div class="mcms-entries-loading" part="loading" role="status" aria-live="polite">',
      '  <span class="mcms-spinner" aria-hidden="true"></span>',
      "  Loading entries...",
      "</div>",
    ].join("");
  }

  private renderError(): string {
    return [
      '<div part="error-state" role="alert" style="',
      "display:flex;flex-direction:column;align-items:center;justify-content:center;",
      "padding:2rem 1rem;color:var(--modulacms-error-color,#dc2626);",
      'font-size:0.875rem;text-align:center;">',
      '  <div style="margin-bottom:0.5rem;">Failed to load entries</div>',
      "  <div>" + escapeHtml(this.state.error ?? "") + "</div>",
      '  <button class="mcms-retry-btn" style="',
      "margin-top:0.75rem;padding:0.375rem 0.75rem;",
      "border:1px solid var(--modulacms-error-color,#dc2626);",
      "border-radius:var(--modulacms-border-radius,0.375rem);",
      "background:transparent;color:var(--modulacms-error-color,#dc2626);",
      'font-family:inherit;font-size:0.8125rem;cursor:pointer;">',
      "    Retry",
      "  </button>",
      "</div>",
    ].join("");
  }

  private renderEmpty(message: string): string {
    return [
      '<div class="mcms-entries-empty" part="empty-state" role="status">',
      "  " + escapeHtml(message),
      "</div>",
    ].join("");
  }

  private renderToolbar(): string {
    const parts: string[] = [];
    parts.push('<div class="mcms-entries-toolbar">');

    if (this.hasAttribute("filterable")) {
      for (const field of this.state.fields) {
        const currentValue = this.state.filters[field.name] ?? "";
        parts.push(
          '<input class="mcms-filter-input" part="filter-input"' +
            ' type="text"' +
            ' data-filter-field="' +
            escapeHtml(field.name) +
            '"' +
            ' placeholder="Filter ' +
            escapeHtml(field.label) +
            '..."' +
            ' value="' +
            escapeHtml(currentValue) +
            '"' +
            ' aria-label="Filter by ' +
            escapeHtml(field.label) +
            '"' +
            " />",
        );
      }
    }

    if (this.hasAttribute("export-enabled")) {
      parts.push(
        '<button class="mcms-export-btn" part="export-button" type="button" aria-label="Export entries as JSON">',
        "  Export JSON",
        "</button>",
      );
    }

    parts.push("</div>");
    return parts.join("");
  }

  private renderTable(): string {
    const fields = this.state.fields;
    const sortable = this.hasAttribute("sortable");
    const parts: string[] = [];

    parts.push('<div class="mcms-entries-table-wrap">');
    parts.push(
      '<table class="mcms-entries-table" part="table" role="grid" aria-label="Entries">',
    );

    // thead
    parts.push('<thead part="thead"><tr role="row">');
    for (const field of fields) {
      const isSorted =
        this.state.sortField === field.name;
      const sortIndicator = isSorted
        ? this.state.sortDir === "asc"
          ? " &#9650;"
          : " &#9660;"
        : "";
      const ariaSort = isSorted
        ? ' aria-sort="' + this.state.sortDir + "ending" + '"'
        : "";

      if (sortable) {
        parts.push(
          "<th part=\"th\" role=\"columnheader\" data-sortable data-field=\"" +
            escapeHtml(field.name) +
            '"' +
            ariaSort +
            ' tabindex="0"' +
            ">" +
            escapeHtml(field.label) +
            sortIndicator +
            "</th>",
        );
      } else {
        parts.push(
          '<th part="th" role="columnheader">' +
            escapeHtml(field.label) +
            "</th>",
        );
      }
    }

    // Submitted column
    const submittedSorted = this.state.sortField === "created_at";
    const submittedIndicator = submittedSorted
      ? this.state.sortDir === "asc"
        ? " &#9650;"
        : " &#9660;"
      : "";
    const submittedAriaSort = submittedSorted
      ? ' aria-sort="' + this.state.sortDir + "ending" + '"'
      : "";

    if (sortable) {
      parts.push(
        '<th part="th" role="columnheader" data-sortable data-field="created_at"' +
          submittedAriaSort +
          ' tabindex="0">Submitted' +
          submittedIndicator +
          "</th>",
      );
    } else {
      parts.push('<th part="th" role="columnheader">Submitted</th>');
    }

    // Status column
    const statusSorted = this.state.sortField === "status";
    const statusIndicator = statusSorted
      ? this.state.sortDir === "asc"
        ? " &#9650;"
        : " &#9660;"
      : "";
    const statusAriaSort = statusSorted
      ? ' aria-sort="' + this.state.sortDir + "ending" + '"'
      : "";

    if (sortable) {
      parts.push(
        '<th part="th" role="columnheader" data-sortable data-field="status"' +
          statusAriaSort +
          ' tabindex="0">Status' +
          statusIndicator +
          "</th>",
      );
    } else {
      parts.push('<th part="th" role="columnheader">Status</th>');
    }

    parts.push("</tr></thead>");

    // tbody
    parts.push('<tbody part="tbody">');
    for (let i = 0; i < this.state.entries.length; i++) {
      const entry = this.state.entries[i]!;
      const isSelected = entry.id === this.state.selectedEntryId;
      parts.push(
        '<tr part="tr" role="row" tabindex="0"' +
          ' data-entry-index="' +
          i +
          '"' +
          ' data-entry-id="' +
          escapeHtml(entry.id) +
          '"' +
          (isSelected ? ' class="selected" aria-selected="true"' : "") +
          ">",
      );

      for (const field of fields) {
        const value = entry.data[field.name];
        const display =
          value === null || value === undefined ? "" : String(value);
        parts.push(
          '<td part="td" title="' +
            escapeHtml(display) +
            '">' +
            escapeHtml(display) +
            "</td>",
        );
      }

      // Submitted
      parts.push('<td part="td">' + escapeHtml(this.formatDate(entry.created_at)) + "</td>");

      // Status
      const statusClass = this.statusClassName(entry.status);
      parts.push(
        '<td part="td"><span class="' +
          statusClass +
          '">' +
          escapeHtml(entry.status) +
          "</span></td>",
      );

      parts.push("</tr>");
    }
    parts.push("</tbody>");

    parts.push("</table>");
    parts.push("</div>");
    return parts.join("");
  }

  private renderPagination(): string {
    const totalPages = this.totalPages();
    const page = this.state.page;
    const start = (page - 1) * this.state.pageSize + 1;
    const end = Math.min(page * this.state.pageSize, this.state.total);

    const parts: string[] = [];
    parts.push(
      '<nav class="mcms-pagination" part="pagination" role="navigation" aria-label="Entries pagination">',
    );

    // Info
    parts.push(
      '<span class="mcms-pagination-info">' +
        start +
        " - " +
        end +
        " of " +
        this.state.total +
        "</span>",
    );

    // Controls
    parts.push('<div class="mcms-pagination-controls">');

    // Previous
    parts.push(
      '<button class="mcms-page-btn" part="page-button" data-page="prev"' +
        (page <= 1 ? " disabled" : "") +
        ' aria-label="Previous page"' +
        '>&lsaquo;</button>',
    );

    // Page number buttons (windowed)
    const pageNumbers = this.computePageWindow(page, totalPages);
    let lastRendered = 0;
    for (const pn of pageNumbers) {
      if (lastRendered > 0 && pn > lastRendered + 1) {
        parts.push(
          '<span class="mcms-pagination-ellipsis" aria-hidden="true">...</span>',
        );
      }
      const isCurrent = pn === page;
      parts.push(
        '<button class="mcms-page-btn" part="page-button" data-page="' +
          pn +
          '"' +
          (isCurrent ? ' aria-current="page"' : "") +
          ' aria-label="Page ' +
          pn +
          '"' +
          ">" +
          pn +
          "</button>",
      );
      lastRendered = pn;
    }

    // Next
    parts.push(
      '<button class="mcms-page-btn" part="page-button" data-page="next"' +
        (page >= totalPages ? " disabled" : "") +
        ' aria-label="Next page"' +
        '>&rsaquo;</button>',
    );

    parts.push("</div>"); // controls
    parts.push("</nav>");
    return parts.join("");
  }

  /**
   * Build a compact page number window: always show first, last, and up to
   * 2 neighbours on each side of the current page.
   */
  private computePageWindow(current: number, total: number): number[] {
    if (total <= 7) {
      const result: number[] = [];
      for (let i = 1; i <= total; i++) {
        result.push(i);
      }
      return result;
    }

    const pages = new Set<number>();
    pages.add(1);
    pages.add(total);
    for (
      let i = Math.max(2, current - 2);
      i <= Math.min(total - 1, current + 2);
      i++
    ) {
      pages.add(i);
    }

    const sorted = Array.from(pages);
    sorted.sort((a, b) => a - b);
    return sorted;
  }

  private statusClassName(status: string): string {
    const lower = status.toLowerCase();
    if (lower === "pending") return "mcms-status-badge mcms-status-pending";
    if (lower === "reviewed") return "mcms-status-badge mcms-status-reviewed";
    if (lower === "spam") return "mcms-status-badge mcms-status-spam";
    return "mcms-status-badge mcms-status-default";
  }

  private formatDate(iso: string): string {
    try {
      const d = new Date(iso);
      return d.toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return iso;
    }
  }

  // -----------------------------------------------------------------------
  // Event binding
  // -----------------------------------------------------------------------

  private attachEventListeners(): void {
    // Row clicks + keyboard
    const rows = Array.from(this.shadow.querySelectorAll("tbody tr[data-entry-id]"));
    for (const row of rows) {
      (row as HTMLElement).addEventListener("click", this.handleRowClick);
      (row as HTMLElement).addEventListener("keydown", this.handleRowKeydown);
    }

    // Sort header clicks + keyboard
    const sortHeaders = Array.from(this.shadow.querySelectorAll("th[data-sortable]"));
    for (const th of sortHeaders) {
      (th as HTMLElement).addEventListener("click", this.handleSortClick);
      (th as HTMLElement).addEventListener("keydown", this.handleSortKeydown);
    }

    // Pagination buttons
    const pageButtons = Array.from(this.shadow.querySelectorAll(
      "button[data-page]",
    ));
    for (const btn of pageButtons) {
      (btn as HTMLElement).addEventListener("click", this.handlePageClick);
    }

    // Filter inputs
    const filterInputs = Array.from(this.shadow.querySelectorAll(
      "input[data-filter-field]",
    ));
    for (const input of filterInputs) {
      (input as HTMLElement).addEventListener("input", this.handleFilterInput);
    }

    // Export button
    const exportBtn = this.shadow.querySelector(".mcms-export-btn");
    if (exportBtn) {
      (exportBtn as HTMLElement).addEventListener(
        "click",
        this.handleExportClick,
      );
    }

    // Retry button
    const retryBtn = this.shadow.querySelector(".mcms-retry-btn");
    if (retryBtn) {
      (retryBtn as HTMLElement).addEventListener(
        "click",
        this.handleRetryClick,
      );
    }
  }

  // Arrow functions for stable `this` binding in addEventListener/removeEventListener
  private handleRowClick = (e: Event): void => {
    const tr = (e.currentTarget as HTMLElement).closest(
      "tr[data-entry-id]",
    ) as HTMLElement | null;
    if (!tr) return;
    this.selectRow(tr);
  };

  private handleRowKeydown = (e: Event): void => {
    const ke = e as KeyboardEvent;
    const tr = ke.currentTarget as HTMLElement;

    if (ke.key === "Enter" || ke.key === " ") {
      ke.preventDefault();
      this.selectRow(tr);
      return;
    }

    if (ke.key === "ArrowDown") {
      ke.preventDefault();
      const next = tr.nextElementSibling as HTMLElement | null;
      if (next && next.hasAttribute("data-entry-id")) {
        next.focus();
      }
      return;
    }

    if (ke.key === "ArrowUp") {
      ke.preventDefault();
      const prev = tr.previousElementSibling as HTMLElement | null;
      if (prev && prev.hasAttribute("data-entry-id")) {
        prev.focus();
      }
      return;
    }

    if (ke.key === "Home") {
      ke.preventDefault();
      const tbody = tr.parentElement;
      if (tbody) {
        const first = tbody.querySelector(
          "tr[data-entry-id]",
        ) as HTMLElement | null;
        if (first) first.focus();
      }
      return;
    }

    if (ke.key === "End") {
      ke.preventDefault();
      const tbody = tr.parentElement;
      if (tbody) {
        const all = tbody.querySelectorAll("tr[data-entry-id]");
        if (all.length > 0) {
          (all[all.length - 1] as HTMLElement).focus();
        }
      }
    }
  };

  private selectRow(tr: HTMLElement): void {
    const entryId = tr.getAttribute("data-entry-id");
    if (!entryId) return;

    const indexStr = tr.getAttribute("data-entry-index");
    const index = indexStr !== null ? parseInt(indexStr, 10) : -1;
    if (index < 0 || index >= this.state.entries.length) return;

    // Update selection visual state
    const prev = this.shadow.querySelector('tr[aria-selected="true"]');
    if (prev) {
      prev.removeAttribute("aria-selected");
      prev.classList.remove("selected");
    }
    tr.setAttribute("aria-selected", "true");
    tr.classList.add("selected");
    this.state.selectedEntryId = entryId;

    dispatch(this, "entry:select", { entry: this.state.entries[index] });
  }

  private handleSortClick = (e: Event): void => {
    const th = (e.currentTarget as HTMLElement).closest(
      "th[data-field]",
    ) as HTMLElement | null;
    if (!th) return;
    this.toggleSort(th);
  };

  private handleSortKeydown = (e: Event): void => {
    const ke = e as KeyboardEvent;
    if (ke.key === "Enter" || ke.key === " ") {
      ke.preventDefault();
      const th = ke.currentTarget as HTMLElement;
      this.toggleSort(th);
    }
  };

  private toggleSort(th: HTMLElement): void {
    const field = th.getAttribute("data-field");
    if (!field) return;

    if (this.state.sortField === field) {
      this.state.sortDir = this.state.sortDir === "asc" ? "desc" : "asc";
    } else {
      this.state.sortField = field;
      this.state.sortDir = "asc";
    }

    this.state.page = 1;
    this.fetchEntries();
  }

  private handlePageClick = (e: Event): void => {
    const btn = e.currentTarget as HTMLElement;
    const pageAttr = btn.getAttribute("data-page");
    if (!pageAttr) return;

    if (pageAttr === "prev") {
      this.goToPage(this.state.page - 1);
    } else if (pageAttr === "next") {
      this.goToPage(this.state.page + 1);
    } else {
      const n = parseInt(pageAttr, 10);
      if (!isNaN(n)) {
        this.goToPage(n);
      }
    }
  };

  private handleFilterInput = (e: Event): void => {
    const input = e.currentTarget as HTMLInputElement;
    const fieldName = input.getAttribute("data-filter-field");
    if (!fieldName) return;

    const value = input.value.trim();
    if (value) {
      this.state.filters[fieldName] = value;
    } else {
      delete this.state.filters[fieldName];
    }

    // Debounce filter application
    if (this.filterDebounceTimer !== null) {
      clearTimeout(this.filterDebounceTimer);
    }
    this.filterDebounceTimer = setTimeout(() => {
      this.filterDebounceTimer = null;
      this.state.page = 1;
      this.fetchEntries();
    }, 300);
  };

  private handleExportClick = (): void => {
    this.exportEntries();
  };

  private handleRetryClick = (): void => {
    if (this.state.form) {
      this.fetchEntries();
    } else {
      this.loadFormAndEntries();
    }
  };
}
