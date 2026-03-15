# Tailwind UI Migration Plan

Maps every existing admin panel component to its Tailwind UI replacement. Migration is incremental — pages migrate when touched, new pages use Tailwind UI patterns from the start.

## Resolved Decisions

- **Theme:** Dark mode only. No light mode support during or after migration. Remove the `html.light` overrides in `tokens.css` during Phase 0. The `@theme` block in `input.css` defines one palette.
- **Web components:** Keep all `mcms-*` web components. Do not adopt `@tailwindplus/elements`. Restyle `mcms-*` internals with Tailwind utility classes. Tailwind UI samples that use `<el-dialog>`, `<el-dropdown>`, `<el-menu>` must be adapted to use the existing `mcms-*` equivalents or native `<dialog>`.
- **Icons:** Keep Lucide icons. Do not switch to Heroicons. When adapting Tailwind UI samples, replace Heroicon SVGs with Lucide equivalents via `<i data-lucide="icon-name">`.
- **JS-CSS decoupling:** Web components must use `data-*` attributes for DOM targeting, not CSS class names. This allows CSS classes to change freely during migration without breaking JS behavior.

---

## Phase 0: Foundation

Complete before any visual migration. No user-visible changes.

### 0a: Token Unification

Eliminate the dual token system. Make `input.css @theme` the single source of truth.

**Critical constraint:** Tailwind v4 `@theme` variables are tree-shaken — they are only emitted as CSS custom properties if a Tailwind utility class references them. Hand-written CSS using `var(--color-*)` will break if the token is only defined in `@theme` and no utility class uses it. To solve this, keep a `:root` block in `input.css` (outside `@theme`) for tokens consumed by hand-written CSS during the migration.

**Token naming:** The existing `tokens.css` uses names like `--color-bg`, `--color-text`, `--color-text-muted`. The `@theme` block already uses different names (`--color-page`, `--color-muted`, `--color-dim`). During migration, keep the `tokens.css` variable names in the `:root` block so that all 400+ existing CSS variable references continue to resolve. The `@theme` semantic names (`--color-page`, `--color-surface`, etc.) are used by Tailwind utility classes only.

1. Add a `:root` block in `input.css` (outside `@theme`) containing ALL tokens from `tokens.css` — including shade variants (`--color-primary-50` through `--color-primary-300`, `--color-danger-*`, `--color-success-*`, `--color-warning-*`, `--color-neutral-100`, `--color-overlay`), z-index values, transition values, focus ring values, semantic aliases, container sizes, and layout values (`--sidebar-width`, `--topbar-height`)
2. Keep the `@theme` block for Tailwind utility class consumption (using the semantic names already established)
3. Remove the `html.light` override block from `tokens.css` (dark mode only)
4. Remove the theme toggle button from `components/topbar.templ`, the `toggleTheme()` and `initTheme()` functions from `admin.js`, and the inline theme script from `layouts/base.templ`
5. Remove the `.theme-toggle` and `.theme-toggle:hover` CSS rules from `utilities.css` (dead code after step 4)
6. Move the `tokens.css` reset (`*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0 }`) into `input.css` (outside `@theme`, before the `:root` block). Keep it until Phase 1 verifies Tailwind preflight is sufficient, then remove it.
7. Remove the `<link rel="stylesheet" href=...tokens.css...>` line from `internal/admin/layouts/base.templ`
8. Delete `tokens.css` entirely — `input.css` now contains all token definitions
9. Run `just admin generate` to rebuild `tailwind.css` and regenerate `base_templ.go`
10. Run `just admin bundle` to verify block editor CSS token references still resolve
11. Run `just check` to verify compilation (catches any `go:embed` issues from deleted file)
12. HUMAN CHECKPOINT: Visually verify no color/spacing regressions across all pages before proceeding to Phase 0b

As each phase migrates hand-written CSS to Tailwind utilities, the corresponding `:root` tokens become unused and can be removed. The `:root` block shrinks to zero when migration is complete.

### 0b: Decouple JS from CSS Classes

Audit all 12 web components + `admin.js`. Replace CSS class **query selectors** with `data-*` attribute selectors so that CSS classes can be freely changed in subsequent phases.

**Classify each class-name reference into one of three categories:**

1. **CLASS-BASED SELECTORS** (convert to `data-*` attributes):
   `querySelector('.dialog-title')`, `querySelectorAll('.media-picker-item.selected')`, `closest('.tree-node')`, `classList.contains()` used in conditional logic

2. **STATE TOGGLES** (convert to `data-*` attributes):
   `classList.add('open')`, `classList.add('active')`, `classList.add('dragging')`, `classList.add('selected')` when also used with `querySelector('.foo.selected')`. When a class name serves dual purpose (JS state detection AND CSS styling), add the `data-*` attribute for JS, keep the class name for CSS styling during migration, and add a comment `// DUAL: data-{name} + class` so later phases know to update the CSS selector.

3. **STYLING ASSIGNMENTS** (leave as class names — will become Tailwind classes in later phases):
   `element.className = 'dialog-backdrop'`, all `createElement().className` assignments. Do NOT convert these to `data-*` attributes.

**Rule:** If a class name appears in BOTH a `querySelector`/`classList.contains` AND a `className` assignment, add the `data-*` attribute AND keep the class name.

**Pattern:**
```js
// Before (coupled to CSS class names)
this.querySelector('.dialog-title')
this.classList.add('dialog-open')

// After (coupled to data attributes)
this.querySelector('[data-dialog-title]')
this.dataset.state = 'open'
```

**For each web component:**
1. List every `querySelector('.')`, `querySelectorAll('.')`, `classList.add/remove/toggle/contains`, and `className` reference in the JS
2. Replace class-based selectors with `data-*` attribute selectors
3. Add the corresponding `data-*` attributes to the templ files
4. Run `just check` to verify compilation. Run `just test` to verify unit tests pass. Run `just admin bundle` if block editor files were modified.

**Files to audit:**
- `mcms-dialog.js` — `dialog-title`, `dialog-confirm-btn`, `dialog-cancel-btn`, `dialog-backdrop`, `dialog-panel`, `dialog-body`, `dialog-actions`
- `mcms-data-table.js` — sort/filter class toggles
- `mcms-field-renderer.js` — field type class references
- `mcms-media-picker.js` — picker item selection classes
- `mcms-media-tree.js` — tree node expansion classes
- `mcms-tree-nav.js` — node toggle/active classes
- `mcms-search.js` — suggestion list classes
- `mcms-file-input.js` — dropzone state classes
- `mcms-toast.js` — toast variant classes
- `mcms-confirm.js` — confirm button classes
- `mcms-validation-wizard.js` — step state classes
- `mcms-scroll.js` — trigger element classes
- `admin.js` — sidebar toggle (`document.querySelector('.sidebar')` in `toggleSidebar()`, `document.querySelector('.sidebar-overlay')` in `toggleSidebar()`, `sidebar.classList.toggle('open')` in `toggleSidebar()`), active link management (`link.classList.add('active')` / `link.classList.remove('active')` in `updateSidebarActive()`), clickable rows (`e.target.closest('tr.clickable-row[data-href]')`)
- `block-editor-src/*.js` — the block editor source files contain class-coupled DOM references across `index.js`, `dom-patches.js`, `picker.js`, and `drag.js`. After decoupling, run `just admin bundle` to rebuild `block-editor.js` from source.

### 0c: HTMX Compatibility Rule

Partials and their parent page templates must be migrated in the same phase. An HTMX partial swapped into a page must use the same styling system as the page it lands in. Do not migrate a page header without also migrating the partials that swap into that page's content area.

**Shared partials** (`pagination.templ`, `empty_state.templ`, `toast.templ`, `delete_confirm.templ`) are used across many pages. Migrate these with the FIRST page that uses them. All subsequent pages must accept the migrated partial's styling.

**Before starting Phase 1:** Generate a page-to-partial dependency map by scanning all `.templ` files for HTMX `hx-get`/`hx-post`/`hx-target` attributes that reference partial endpoints. Store this map in `ai/plans/admin/PAGE_PARTIAL_MAP.md` for use in Phases 1-8.

---

## Phase 1: Application Shell

The shell wraps every page. Migrating it first gives immediate visual lift everywhere.

### Sidebar + Topbar + Content Area

| Current | Tailwind UI Source | Notes |
|---------|-------------------|-------|
| `layouts/admin.templ` (sidebar + topbar + main) | `application-shells/sidebar-layouts-dark-sidebar-with-header.html` | Dark sidebar with header bar matches current layout. Mobile drawer uses existing sidebar overlay toggle (no `el-dialog`). |
| `components/sidebar.templ` (`.sidebar` + `.sidebar-nav`) | `navigation/sidebar-navigation-with-expandable-sections.html` | Permission-gated nav items, active link styling. Use `bg-white/5 text-white` active pattern from sample. |
| `components/topbar.templ` (`.topbar`) | Top header bar from `application-shells/sidebar-layouts-dark-sidebar-with-header.html` | Currently has: sidebar toggle, brand link, username, logout button. Theme toggle is removed in Phase 0. Consider adding: search input, profile dropdown (from `elements/dropdowns-with-simple-header.html`) as new features during migration. |
| `layouts/auth.templ` (centered auth card) | `forms/sign-in-and-register-simple-card.html` | Login page layout. |

### Current CSS to remove after Phase 1
- `layout.css` (210 lines) — entirely replaced

---

## Phase 2: Page Headers

Every page uses `.page-header`. Standardize to Tailwind UI heading patterns.

| Current Pattern | Pages Using It | Tailwind UI Source |
|-----------------|---------------|-------------------|
| Simple header: h1 + single action button | Users, Fields, Tokens, Sessions, Webhooks, Plugins, Locales, Field Types | `headings/page-headings-with-actions.html` |
| Header with back button + h1 + badge + save | Datatype Detail, Field Detail, Media Detail, Plugin Detail | `headings/page-headings-with-actions-and-breadcrumbs.html` |
| Header with filters (search + select + sort) | Content List, Datatypes List | `headings/page-headings-with-filter-and-action.html` |
| Header with breadcrumb trail | Media List (folder breadcrumbs) | `navigation/breadcrumbs-simple-with-chevrons.html` |
| Section headings inside pages | Settings fieldsets, Role detail sections | `headings/section-headings-with-action.html` or `section-headings-simple.html` |

### Card Headings (for detail sections)

| Current | Tailwind UI Source |
|---------|-------------------|
| `.detail-section` header (h2 + description) | `card-headings/card-headings-with-description.html` |
| `.detail-section` header with action button | `card-headings/card-headings-with-action.html` |

### Current CSS to remove after Phase 2
- Page header rules from `pages.css` (~50 lines)
- Page header responsive rules from `utilities.css` (~20 lines)

---

## Phase 3: Data Tables

18 table variants in the Tailwind UI library. The admin panel uses tables on 15+ pages.

| Current Pattern | Tailwind UI Source |
|-----------------|-------------------|
| Standard data table (`.data-table`) | `lists/tables-simple.html` |
| Table with clickable rows (datatypes, fields) | `lists/tables-simple.html` + add `cursor-pointer hover:bg-gray-50 dark:hover:bg-white/5` |
| Table with action buttons column | `lists/tables-full-width.html` (has Edit link in last column) |
| Table with badges (status, type) | Use badge elements inline. See Phase 6. |
| Table with checkboxes (role permissions) | `lists/tables-with-checkboxes.html` |
| Table in card container | `lists/tables-simple-in-card.html` |
| Empty state (no rows) | `feedback/empty-states-simple.html` |
| Pagination | `navigation/pagination-card-footer-with-page-buttons.html` |

### mcms-data-table web component
Keep the JS behavior (sort, filter). Replace the CSS styling to use Tailwind classes. The web component renders to Light DOM so utility classes work directly.

### Current CSS to remove after Phase 3
- Table rules from `components.css` (~100 lines)
- Table rules from `web-components.css` (~50 lines)
- Pagination rules from `components.css` (~60 lines)

---

## Phase 4: Forms

117 form samples cover every input pattern the admin panel uses.

| Current Pattern | Pages Using It | Tailwind UI Source |
|-----------------|---------------|-------------------|
| Stacked form layout (label above input) | All create/edit forms | `forms/form-layouts-stacked.html` |
| Two-column form in cards | Settings page | `forms/form-layouts-two-column-with-cards.html` |
| Text input with label + error | Everywhere | `forms/input-groups-input-with-label-and-help-text.html` + `input-groups-input-with-validation-error.html` |
| Select dropdown | Field type, content status, etc. | `forms/select-menus-simple-native.html` |
| Textarea | Validation rules, JSON config | `forms/textareas-simple.html` |
| Checkbox with label | Settings toggles, permissions | `forms/checkboxes-list-with-description.html` |
| Toggle switches | Settings boolean values | `forms/toggles-with-left-label-and-description.html` |
| File upload dropzone | Media upload | `forms/input-groups-input-with-label.html` (keep mcms-file-input JS, restyle) |
| Settings fieldset groups | Settings page | `forms/form-layouts-two-column-with-cards.html` (each fieldset = a card) |
| Action panels (save/cancel bar) | Bottom of forms | `forms/action-panels-with-button-on-right.html` |

### Dialog Forms

| Current Pattern | Tailwind UI Source |
|-----------------|-------------------|
| Create dialog (modal + form) | `overlays/modal-dialogs-simple-with-gray-footer.html` |
| Upload dialog | `overlays/modal-dialogs-simple-with-gray-footer.html` + mcms-file-input inside |
| Confirmation dialog | `overlays/modal-dialogs-simple-alert.html` |

### Current CSS to remove after Phase 4
- Form rules from `components.css` (~200 lines)
- Settings form rules from `pages.css` (~100 lines)
- Dialog rules from `web-components.css` (~100 lines)

---

## Phase 5: Cards and Detail Layouts

| Current Pattern | Pages Using It | Tailwind UI Source |
|-----------------|---------------|-------------------|
| `.detail-grid` (2fr + 1fr) | Datatype Detail, Field Detail | `layout/cards-with-header.html` inside a `grid grid-cols-1 lg:grid-cols-3 lg:gap-6` (main = col-span-2) |
| `.detail-section` (card container) | Detail pages | `layout/cards-with-header.html` or `layout/cards-basic-card.html` |
| `.settings-section` (fieldset card) | Settings | `layout/cards-with-header.html` |
| `.detail-meta` (metadata list) | Detail pages | `data-display/description-lists-left-aligned-in-card.html` |
| `.media-detail-layout` (preview + sidebar) | Media Detail | `page-examples/detail-screens-sidebar.html` for overall layout |
| Role detail panel | Roles page | `layout/cards-with-header-and-footer.html` (header = role name, footer = save/delete) |
| `.content-layout` (tree sidebar + editor) | Content Edit | `application-shells/multi-column-layouts-full-width-secondary-column-on-right.html` |

### Description Lists (metadata displays)

| Current Pattern | Tailwind UI Source |
|-----------------|-------------------|
| Media detail info (file size, dimensions, type) | `data-display/description-lists-left-aligned-in-card.html` |
| Datatype/field metadata (ID, created, modified) | `data-display/description-lists-left-aligned.html` |

### Current CSS to remove after Phase 5
- Detail/settings rules from `pages.css` (~150 lines)
- Card rules from `components.css` (~50 lines)

---

## Phase 6: Elements

| Current Pattern | Tailwind UI Source |
|-----------------|-------------------|
| `.btn` / `.btn-primary` | `elements/buttons-primary-buttons.html` |
| `.btn-ghost` (secondary) | `elements/buttons-secondary-buttons.html` or `elements/buttons-soft-buttons.html` |
| `.btn-danger` | Primary button with danger colors |
| `.btn-sm` | Size variants in all button samples |
| `.btn` with icon | `elements/buttons-with-leading-icon.html` |
| `.btn-split` (split dropdown) | `elements/button-groups-with-dropdown.html` |
| `.badge` / `.badge-{status}` | `elements/badges-flat-pill.html` or `elements/badges-flat-pill-with-dot.html` |
| `.text-muted` | `text-gray-500 dark:text-gray-400` (built-in Tailwind) |
| Avatars (user list, topbar) | `elements/avatars-circular-avatars-with-placeholder-initials.html` |
| Dropdown menus | `elements/dropdowns-with-icons.html` |

### Current CSS to remove after Phase 6
- Button rules from `components.css` (~100 lines)
- Badge rules from `components.css` (~80 lines)

---

## Phase 7: Feedback and Navigation

| Current Pattern | Tailwind UI Source |
|-----------------|-------------------|
| Toast notifications (mcms-toast) | `overlays/notifications-simple.html` or `overlays/notifications-with-actions-below.html` |
| Alert messages (login errors) | `feedback/alerts-with-description.html` |
| Empty states | `feedback/empty-states-simple.html` |
| Delete confirmation | `overlays/modal-dialogs-simple-alert.html` |
| Sidebar navigation | `navigation/sidebar-navigation-dark.html` |
| Breadcrumbs | `navigation/breadcrumbs-simple-with-chevrons.html` |
| Tabs (content edit locales) | `navigation/tabs-with-underline.html` |
| Pagination | `navigation/pagination-card-footer-with-page-buttons.html` |

### Current CSS to remove after Phase 7
- Alert/toast rules from `components.css` (~80 lines)
- Empty state rules from `components.css` (~30 lines)

---

## Phase 8: Specialized Pages

These pages have unique layouts that map to full page examples.

| Page | Tailwind UI Source |
|------|-------------------|
| Dashboard | `page-examples/home-screens-sidebar.html` + `data-display/stats-simple-in-cards.html` |
| Settings | `page-examples/settings-screens-sidebar.html` |
| Roles (two-panel) | `page-examples/detail-screens-sidebar.html` (sidebar = role list, main = role detail) |
| Media List (folder tree + grid) | Multi-column layout. Sidebar = `navigation/sidebar-navigation-with-expandable-sections.html`, Grid = `lists/grid-lists-images-with-details.html` |
| Content Edit (tree + editor + fields) | `application-shells/multi-column-layouts-full-width-three-column.html` |
| Login / Register | `forms/sign-in-and-register-simple-card.html` |
| Audit Log | Standard table page (Phase 3 pattern) |

---

## What Stays Custom

These components have behavior too specific to replace with Tailwind UI templates. Restyle with Tailwind utility classes but keep the JS/structure.

| Component | Reason |
|-----------|--------|
| `mcms-data-table.js` | Sort/filter/column toggle behavior. Restyle only. |
| `mcms-field-renderer.js` | Dynamic field type rendering. Restyle inputs per Phase 4. |
| `mcms-validation-wizard.js` | Complex multi-step rule builder. Restyle only. |
| `mcms-media-picker.js` | Modal media selection grid. Restyle modal per Phase 4, grid per Phase 8. |
| `mcms-media-tree.js` | Drag-drop folder tree. Restyle nodes with Tailwind utilities. |
| `mcms-tree-nav.js` | Content tree navigation. Restyle only. |
| `mcms-search.js` | Debounced search with suggestions. Restyle input per Phase 4. |
| `mcms-scroll.js` | Infinite scroll trigger. No visual component. |
| `mcms-file-input.js` | Drag-drop file upload. Restyle dropzone. |
| `mcms-toast.js` | Toast container. Restyle per Phase 7 notification patterns. |
| `mcms-confirm.js` | Confirmation dialog. Restyle per Phase 4 modal patterns. |
| `mcms-dialog.js` | Generic modal. Restyle per Phase 4 modal patterns. |
| Block editor (`block-editor.css` + `block-editor.js`) | Self-contained. Migrate independently or last. 833 lines of scoped CSS. |

---

## CSS File Retirement Schedule

As phases complete, CSS files shrink. Target: remove entirely once all rules are migrated.

| File | Lines | Phases That Replace It | When to Delete |
|------|-------|----------------------|----------------|
| `layout.css` | 210 | Phase 1 | After shell migration |
| `components.css` | 950 | Phases 3, 4, 5, 6, 7 | After all component phases |
| `pages.css` | 752 | Phases 2, 5, 8 | After all page phases |
| `utilities.css` | 305 | Phases 1, 2 (responsive rules become Tailwind responsive) | After shell + headers |
| `web-components.css` | 1,062 | Phases 3, 4, 7 + custom restyle | After web component restyle |
| `tokens.css` | 162 | Phase 0 moves all tokens to `input.css` (`:root` block for hand-written CSS consumption + `@theme` for Tailwind utilities). Reset block kept in `input.css` until Phase 1 verifies Tailwind preflight is sufficient. Light mode overrides and theme toggle removed. File deleted in Phase 0. | Phase 0 |
| `block-editor.css` | 833 | Independent migration | Last or never (self-contained) |

---

## Migration Order (Recommended)

1. **Phase 0: Foundation** — token unification, JS decoupling, HTMX rule. No visual changes.
2. **Phase 1: Shell** — biggest visual impact, wraps everything
3. **Phase 2: Page Headers** — standardizes every page top section
4. **Phase 3: Tables** — covers 15+ pages
5. **Phase 4: Forms** — covers all CRUD pages
6. **Phase 6: Elements** — buttons + badges (many already migrated inline during phases 3-4)
7. **Phase 7: Feedback** — toasts, alerts, empty states
8. **Phase 5: Cards/Details** — detail page layouts
9. **Phase 8: Specialized Pages** — dashboard, settings, media, content edit

---

## Color Mapping

Tailwind UI samples use `indigo-600` as primary. Map to project palette:

| Tailwind UI Default | ModulaCMS Equivalent | Tailwind Class |
|--------------------|---------------------|----------------|
| `indigo-600` / `indigo-500` | `--color-primary` (#3b82f6) | `bg-primary` / `text-primary` (via @theme) |
| `red-600` | `--color-danger` (#ef4444) | `bg-danger` |
| `green-600` | `--color-success` (#22c55e) | `bg-success` |
| `gray-900` (dark bg) | `--color-page` (#0a0a0a) | `bg-page` |
| `gray-800` (surfaces) | `--color-surface` (#141414) | `bg-surface` |
| `gray-700` (hover) | `--color-surface-hover` (#1a1a1a) | `bg-surface-hover` |
| `gray-400` (muted text) | `--color-muted` (#a1a1aa) | `text-muted` |
| `gray-500` (dim text) | `--color-dim` (#71717a) | `text-dim` |
| `white/10` (borders) | `--color-border` (#262626) | `border-border` |

When adapting a Tailwind UI sample:
- Replace `indigo-*` with `primary`/`primary-hover` and `gray-*` with the semantic equivalents
- Use the `dark:` variant classes as the default styles (the admin is dark-only)
- Strip all `dark:` prefixes and remove any light-mode-only classes
- Replace Heroicon SVGs with Lucide equivalents: `<i data-lucide="icon-name">` (see [Lucide icons](https://lucide.dev/icons/) for mappings)
- Replace `<el-dialog>`, `<el-dropdown>`, `<el-menu>` with `mcms-dialog`, `mcms-confirm`, or native `<dialog>` elements

---

## Adapting Tailwind UI Samples

Tailwind UI samples are light-first with `dark:` overrides. Since the admin panel is dark-only, apply this transformation when porting a sample:

1. If a Tailwind UI sample sets a property only for light mode (no `dark:` counterpart), do not use that value. Find the semantic equivalent from the project's `@theme` tokens (e.g., light `text-gray-900` with no dark variant maps to `text-white` or `text-muted` depending on context).
2. Take the `dark:` variant value as the default class (e.g., `dark:bg-gray-900` becomes `bg-gray-900`)
3. Remove the light-mode class entirely (e.g., drop `bg-white`)
4. Remove the `dark:` prefix from all remaining dark classes
5. Replace `<el-*>` interactive components with `mcms-*` web components or native `<dialog>`
6. Replace Heroicon `<svg>` elements with `<i data-lucide="equivalent-name">`
