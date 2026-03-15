# Tailwind UI Migration Plan

Maps every existing admin panel component to its Tailwind UI replacement. Migration is incremental — pages migrate when touched, new pages use Tailwind UI patterns from the start.

## Phase 1: Application Shell

The shell wraps every page. Migrating it first gives immediate visual lift everywhere.

### Sidebar + Topbar + Content Area

| Current | Tailwind UI Source | Notes |
|---------|-------------------|-------|
| `layouts/admin.templ` (sidebar + topbar + main) | `application-shells/sidebar-layouts-dark-sidebar-with-header.html` | Dark sidebar with header bar matches current layout. Mobile drawer via `el-dialog` replaces current overlay toggle. |
| `components/sidebar.templ` (`.sidebar` + `.sidebar-nav`) | `navigation/sidebar-navigation-with-expandable-sections.html` | Permission-gated nav items, active link styling. Use `bg-white/5 text-white` active pattern from sample. |
| `components/topbar.templ` (`.topbar`) | Top header bar from `application-shells/sidebar-layouts-dark-sidebar-with-header.html` | Search, theme toggle, user menu. Profile dropdown from `elements/dropdowns-with-simple-header.html`. |
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
| `tokens.css` | 162 | Keep during migration (reset + tokens). Remove reset once Tailwind preflight is sole reset. Theme tokens move to `input.css @theme`. | Last |
| `block-editor.css` | 833 | Independent migration | Last or never (self-contained) |

---

## Migration Order (Recommended)

1. **Phase 1: Shell** — biggest visual impact, wraps everything
2. **Phase 6: Elements** — buttons + badges used everywhere, small per-file changes
3. **Phase 2: Page Headers** — standardizes every page top section
4. **Phase 3: Tables** — covers 15+ pages
5. **Phase 4: Forms** — covers all CRUD pages
6. **Phase 7: Feedback** — toasts, alerts, empty states
7. **Phase 5: Cards/Details** — detail page layouts
8. **Phase 8: Specialized Pages** — dashboard, settings, media, content edit

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

When adapting a Tailwind UI sample, replace `indigo-*` with `primary`/`primary-hover` and `gray-*` with the semantic equivalents. The `dark:` variants in samples become the default (dark-first), and `html.light` overrides become the light variants.

---

## @tailwindplus/elements

Several Tailwind UI samples use `<el-dialog>`, `<el-dropdown>`, `<el-menu>` from `@tailwindplus/elements`. These are lightweight web components for interactive behavior (open/close, transitions).

**Decision needed:** Use `@tailwindplus/elements` alongside existing `mcms-*` components, or rewrite `mcms-dialog`/`mcms-confirm` to match the `el-*` API.

**Recommendation:** Use `@tailwindplus/elements` for new components. Existing `mcms-*` components keep working (Light DOM, utility classes applied directly). Migrate `mcms-dialog` to `el-dialog` pattern when it's touched.
