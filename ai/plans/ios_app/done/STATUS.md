# iOS App Status

Last updated: 2026-02-15

## Architecture

- **Platform**: iOS 16+, SwiftUI, `@Observable` pattern
- **SDK**: `ModulaCMS` Swift SDK (SPM package at `sdks/swift/`)
- **Auth**: Bearer token via `Authorization` header (API key from Settings)
- **Storage**: UserDefaults for config persistence
- **Theme**: Dark mode enforced globally

## Navigation

Three-tab layout (`RootView.swift`):

| Tab | Icon | Root View | Purpose |
|-----|------|-----------|---------|
| Content | `list.bullet` | DatatypeListView | Browse/edit content by type |
| Routes | `globe` | RouteListView | Manage routes directly |
| Settings | `gear` | SettingsView | Backend connection config |

First-launch welcome sheet guides user to Settings.

## Screens

### Content Tab Flow

1. **DatatypeListView** — Lists root-type datatypes (e.g. Page, Post)
2. **ContentListView** — Lists content items for a datatype, sorted by modification date
3. **ContentEditorView** — View/edit/delete a content item (route + fields)
4. **ContentCreateView** — Create new content (route + status + fields)

### Routes Tab Flow

1. **RouteListView** — Lists all routes with active/inactive status
2. **RouteCreateView** — Create a standalone route
3. Tap a route resolves its content and navigates to ContentEditorView

### Settings

- **SettingsView** — Base URL, API key, author ID, connection test

## Implemented Features

| Feature | Notes |
|---------|-------|
| Content CRUD | Create, read, update, delete with full field support |
| Field type rendering | text, textarea, richtext, number, boolean, select, date, datetime |
| Route list/create/delete | Full CRUD minus inline route editing |
| Auto-slug generation | Title lowercased, spaces to hyphens, filtered to alphanumeric |
| Slug `/` prefix enforcement | All slug writes prepend `/` if missing |
| Content status management | Draft, published, archived, pending — picker in create/edit |
| Route status | Active (1) / Inactive (0) picker |
| Author tracking | `AppConfig.authorID` passed as required `UserID` on all mutations |
| Pull-to-refresh | All list views |
| Swipe-to-delete | Content list and route list |
| Delete confirmation | Dialog before destructive actions |
| Error banners | Inline red banner in forms for save/load errors |
| Connection testing | Tests API with route count display |
| Config persistence | Base URL, API key, author ID saved in UserDefaults |
| Client reconfiguration | `CMSClient.reconfigure()` after settings save |

## Shared UI Components

| Component | Purpose |
|-----------|---------|
| `LoadStateView<T>` | Generic async state handler (idle/loading/loaded/error/empty) |
| `StatusBadge` | Color-coded capsule (green=published, orange=draft, gray=archived, blue=pending) |
| `ErrorBanner` | Red inline error display |
| `ConfirmDeleteButton` | Destructive button with confirmation dialog |
| `FieldRenderer` | Dispatches field type to appropriate read-only or editable control |

## Field Type Editors

| Field Type | Editor | Notes |
|------------|--------|-------|
| text | `TextFieldView` | Single-line input |
| textarea | `TextAreaFieldView` | Multi-line (3-8 lines) |
| richtext | `TextAreaFieldView` | Same as textarea (no rich editor) |
| number | `NumberFieldView` | Decimal pad keyboard |
| boolean | `BooleanFieldView` | Toggle, stores "true"/"false" |
| select | `SelectFieldView` | Picker, options parsed from field.data JSON |
| date | `DateFieldView` | DatePicker, ISO8601 format |
| datetime | `DateTimeFieldView` | DatePicker with time, ISO8601 format |

## SDK Resources Used

```
client.datatypes.list()
client.contentData.list() / .get() / .create() / .update() / .delete()
client.contentFields.list() / .create() / .update()
client.fields.list()
client.datatypeFields.list()
client.routes.list() / .create() / .update() / .delete()
```

## Not Implemented

| Feature | Priority | Notes |
|---------|----------|-------|
| Media upload | High | No file picker or image upload UI |
| Search/filter | Medium | No search box in lists |
| Pagination | Medium | Loads all items; fine for small datasets |
| Batch operations | Low | Only single-item swipe-to-delete |
| Relation fields | Low | No UI for linking content to other content |
| Rich text editor | Low | Falls back to plain textarea |
| Publish scheduling | Low | Status field exists but no date-based scheduling |
| Offline support | Low | No caching or offline mode |

## File Map

```
ios/ModulaCMS Mobile/ModulaCMS Mobile/
  ModulaCMS_MobileApp.swift          — App entry point
  Config/
    AppConfig.swift                  — UserDefaults-backed settings
  Services/
    CMSClient.swift                  — SDK singleton wrapper
  Models/
    LoadState.swift                  — Generic async state enum
    ContentItem.swift                — Content + route display wrapper
    FieldConfig.swift                — Select option parser
  ViewModels/
    DatatypeListViewModel.swift      — Datatype listing
    ContentListViewModel.swift       — Content listing + delete
    ContentCreateViewModel.swift     — Content + route + fields creation
    ContentEditorViewModel.swift     — Content viewing/editing/deleting
    RouteListViewModel.swift         — Route listing + content resolution
    RouteCreateViewModel.swift       — Route creation
    SettingsViewModel.swift          — Settings + connection test
  Views/
    RootView.swift                   — Tab bar + welcome sheet
    Content/
      DatatypeListView.swift         — Datatype list screen
      ContentListView.swift          — Content list screen
      ContentCreateView.swift        — Content creation form
      ContentEditorView.swift        — Content detail/edit screen
      ContentFieldsSection.swift     — Read-only field display
      ContentFieldEditView.swift     — Editable field display
    Fields/
      FieldRenderer.swift            — Field type dispatcher
      TextFieldView.swift            — Text input
      TextAreaFieldView.swift        — Multiline input
      NumberFieldView.swift          — Number input
      BooleanFieldView.swift         — Toggle
      SelectFieldView.swift          — Dropdown picker
      DateFieldView.swift            — Date picker
      DateTimeFieldView.swift        — Date+time picker
    Routes/
      RouteListView.swift            — Route list screen
      RouteCreateView.swift          — Route creation form
    Settings/
      SettingsView.swift             — Settings form
    Shared/
      LoadStateView.swift            — Generic loading state view
      StatusBadge.swift              — Status capsule badge
      ErrorBanner.swift              — Error message banner
      ConfirmDeleteButton.swift      — Delete with confirmation
```
