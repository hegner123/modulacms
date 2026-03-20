# Forms Web Components

Render forms, display submissions, and build form editors using three embeddable web components.

The `@modulacms/forms` package provides `<modulacms-form>`, `<modulacms-entries>`, and `<modulacms-form-builder>` as custom HTML elements. They work in any framework (React, Vue, Svelte, plain HTML) and have zero runtime dependencies.

## Installation

### npm / pnpm

```bash
pnpm add @modulacms/forms
```

```typescript
import "@modulacms/forms";
```

Importing the package registers all three custom elements automatically.

### CDN / Script Tag

```html
<script src="https://unpkg.com/@modulacms/forms/dist/index.global.js"></script>
```

The IIFE bundle registers the elements on load. No import needed.

## Form Renderer

`<modulacms-form>` fetches a form definition and renders an accessible, styled form that submits to your ModulaCMS instance.

### Basic Usage

```html
<modulacms-form
  form-id="01HXYZ..."
  api-url="https://cms.example.com"
></modulacms-form>
```

That is all you need. The component fetches the form definition from the public API, renders all fields with labels and validation, and handles submission.

### Attributes

| Attribute | Required | Description |
|-----------|----------|-------------|
| `form-id` | Yes | The form ID from ModulaCMS |
| `api-url` | Yes | Your ModulaCMS server URL |
| `api-key` | No | API key for authenticated endpoints |
| `submit-label` | No | Override the submit button text |
| `success-message` | No | Override the success message |
| `redirect-url` | No | Redirect after successful submission |
| `loading-text` | No | Text shown while the form loads |

### Events

Listen for form lifecycle events on the element:

```javascript
const form = document.querySelector("modulacms-form");

form.addEventListener("form:loaded", (e) => {
  console.log("Form definition:", e.detail.form);
});

form.addEventListener("form:submit", (e) => {
  console.log("Submitting:", e.detail.data);
  // Call e.preventDefault() to cancel the submission
});

form.addEventListener("form:success", (e) => {
  console.log("Submitted:", e.detail.response.id);
});

form.addEventListener("form:error", (e) => {
  console.error("Error:", e.detail.error);
});

form.addEventListener("field:change", (e) => {
  console.log(e.detail.name, "=", e.detail.value);
});
```

| Event | Cancelable | Detail |
|-------|-----------|--------|
| `form:loaded` | No | `{ form: FormDefinition }` |
| `form:submit` | Yes | `{ data: Record<string, string> }` |
| `form:success` | No | `{ response: { id, message, redirect_url } }` |
| `form:error` | No | `{ error: string }` |
| `field:change` | No | `{ name: string, value: string }` |
| `field:validate` | No | `{ name: string, error: string \| null }` |

### Methods

```javascript
const form = document.querySelector("modulacms-form");

// Validate without submitting
const errors = form.validate();
// Returns: [{ field: "email", message: "Email is required" }]

// Submit programmatically
await form.submit();

// Reset all fields
form.reset();

// Read/write field values
form.setFieldValue("email", "user@example.com");
const email = form.getFieldValue("email");
```

### Styling

The component uses Shadow DOM with CSS custom properties and `::part()` selectors for full style control.

**Custom properties** (set on the element or any ancestor):

```css
modulacms-form {
  --modulacms-font-family: "Inter", sans-serif;
  --modulacms-primary-color: #2563eb;
  --modulacms-error-color: #dc2626;
  --modulacms-border-color: #d1d5db;
  --modulacms-border-radius: 0.5rem;
  --modulacms-field-gap: 1.5rem;
  --modulacms-input-padding: 0.75rem 1rem;
  --modulacms-button-bg: #1d4ed8;
  --modulacms-button-color: white;
  --modulacms-button-padding: 0.75rem 2rem;
  --modulacms-button-radius: 0.5rem;
}
```

**Part selectors** for targeted styling:

```css
modulacms-form::part(form) {
  max-width: 600px;
  margin: 0 auto;
}

modulacms-form::part(label) {
  font-weight: 600;
  font-size: 0.875rem;
}

modulacms-form::part(input) {
  border: 2px solid var(--modulacms-border-color);
  transition: border-color 0.2s;
}

modulacms-form::part(input):focus {
  border-color: var(--modulacms-primary-color);
  outline: none;
}

modulacms-form::part(submit) {
  width: 100%;
  font-size: 1rem;
}

modulacms-form::part(error) {
  font-size: 0.75rem;
}

modulacms-form::part(success) {
  background: #f0fdf4;
  padding: 1.5rem;
  border-radius: 0.5rem;
}
```

Available parts: `form`, `field`, `label`, `input`, `help-text`, `error`, `submit`, `loading`, `success`, `error-state`.

### Framework Examples

**React:**

```tsx
function ContactForm() {
  return (
    <modulacms-form
      form-id="01HXYZ..."
      api-url="https://cms.example.com"
      onform:success={(e) => router.push("/thank-you")}
    />
  );
}
```

> **Good to know**: React requires a ref or `addEventListener` for custom element events. The `on*` syntax works in some frameworks but not all React versions. Use a ref for reliable event handling.

**Vue:**

```vue
<template>
  <modulacms-form
    form-id="01HXYZ..."
    api-url="https://cms.example.com"
    @form:success="onSuccess"
    @form:error="onError"
  />
</template>
```

**Svelte:**

```svelte
<modulacms-form
  form-id="01HXYZ..."
  api-url="https://cms.example.com"
  on:form:success={handleSuccess}
/>
```

**Plain HTML:**

```html
<modulacms-form
  id="contact"
  form-id="01HXYZ..."
  api-url="https://cms.example.com"
></modulacms-form>

<script>
  document.getElementById("contact")
    .addEventListener("form:success", (e) => {
      window.location.href = "/thank-you";
    });
</script>
```

## Entries Viewer

`<modulacms-entries>` displays form submissions in a paginated, sortable table. Use this in admin panels to review and manage entries.

### Basic Usage

```html
<modulacms-entries
  form-id="01HXYZ..."
  api-url="https://cms.example.com"
  api-key="your-admin-key"
></modulacms-entries>
```

### Attributes

| Attribute | Required | Description |
|-----------|----------|-------------|
| `form-id` | Yes | The form ID |
| `api-url` | Yes | Your Modula server URL |
| `api-key` | Yes | Admin API key (entries require authentication) |
| `page-size` | No | Entries per page (default 20) |
| `sortable` | No | Enable column sorting (presence attribute) |
| `filterable` | No | Enable column filtering (presence attribute) |
| `export-enabled` | No | Show export button (presence attribute) |

### Full-Featured Example

```html
<modulacms-entries
  form-id="01HXYZ..."
  api-url="https://cms.example.com"
  api-key="your-admin-key"
  page-size="50"
  sortable
  filterable
  export-enabled
></modulacms-entries>
```

### Events

| Event | Detail |
|-------|--------|
| `entries:loaded` | `{ entries: FormEntry[], total: number }` |
| `entry:select` | `{ entry: FormEntry }` |
| `entries:page-change` | `{ page: number }` |
| `entries:export` | `{ entries: FormEntry[] }` |

### Methods

```javascript
const viewer = document.querySelector("modulacms-entries");

// Refresh the current page
viewer.refresh();

// Navigate to a specific page
viewer.goToPage(3);

// Export all entries as a JSON download
await viewer.exportEntries();

// Filter and sort programmatically
viewer.setFilter({ status: "submitted" });
viewer.clearFilters();
viewer.setSort("created_at", "desc");
```

### Styling

Same custom property system as the form renderer. Additional parts for table elements:

```css
modulacms-entries::part(table) {
  font-size: 0.875rem;
}

modulacms-entries::part(th) {
  background: #f9fafb;
  text-transform: uppercase;
  font-size: 0.75rem;
  letter-spacing: 0.05em;
}

modulacms-entries::part(tr):hover {
  background: #f3f4f6;
  cursor: pointer;
}

modulacms-entries::part(pagination) {
  justify-content: center;
  padding: 1rem 0;
}

modulacms-entries::part(export-button) {
  background: transparent;
  border: 1px solid var(--modulacms-border-color);
}
```

Available parts: `table`, `thead`, `th`, `tbody`, `tr`, `td`, `pagination`, `page-button`, `export-button`, `filter-input`, `empty-state`, `loading`, `error-state`.

## Form Builder

`<modulacms-form-builder>` provides a drag-and-drop form editor for admin panels. Users can add fields from a palette, configure field properties, reorder with drag-and-drop, and save changes to the server.

### Basic Usage

```html
<modulacms-form-builder
  form-id="01HXYZ..."
  api-url="https://cms.example.com"
  api-key="your-admin-key"
></modulacms-form-builder>
```

### Attributes

| Attribute | Required | Description |
|-----------|----------|-------------|
| `form-id` | Yes | The form ID to edit |
| `api-url` | Yes | Your ModulaCMS server URL |
| `api-key` | Yes | Admin API key |
| `auto-save` | No | Automatically save changes after 2 seconds of inactivity (presence attribute) |

### Events

| Event | Detail |
|-------|--------|
| `builder:loaded` | `{ form: FormDefinition, fields: FormFieldDefinition[] }` |
| `builder:save` | `{ fields: FormFieldDefinition[] }` |
| `builder:change` | `{ fields: FormFieldDefinition[] }` |
| `field:add` | `{ field: FormFieldDefinition }` |
| `field:remove` | `{ index: number }` |
| `field:reorder` | `{ fromIndex: number, toIndex: number }` |

### Methods

```javascript
const builder = document.querySelector("modulacms-form-builder");

// Save current state to server
await builder.save();

// Add a new field
builder.addField("email");

// Remove a field by position
builder.removeField(2);

// Move a field from one position to another
builder.moveField(0, 3);

// Get/set the full field definition list
const fields = builder.getDefinition();
builder.setDefinition(modifiedFields);
```

### Layout

The builder renders two panels:

- **Left**: Field palette with buttons for each field type (text, email, number, etc.). Click or press Enter to add a field.
- **Right**: Canvas showing the current field list. Each field is a card with a drag handle, label, type badge, and delete button. Click a card to expand its configuration panel.

Drag-and-drop uses native Pointer Events for reliable cross-browser behavior. The implementation is ported from the ModulaCMS block editor.

### Styling

```css
modulacms-form-builder::part(builder) {
  min-height: 500px;
}

modulacms-form-builder::part(field-palette) {
  background: #f9fafb;
  padding: 1rem;
}

modulacms-form-builder::part(field-type-button) {
  font-size: 0.75rem;
  border-radius: 0.375rem;
}

modulacms-form-builder::part(canvas) {
  background: white;
  min-height: 300px;
}

modulacms-form-builder::part(field-item) {
  border: 1px solid var(--modulacms-border-color);
  border-radius: 0.5rem;
  padding: 0.75rem;
}

modulacms-form-builder::part(save-button) {
  background: var(--modulacms-primary-color);
  color: white;
}
```

Available parts: `builder`, `toolbar`, `field-palette`, `field-type-button`, `canvas`, `field-item`, `field-handle`, `field-config`, `config-input`, `preview`, `save-button`, `loading`, `error-state`.

## TypeScript API Client

The package also exports `FormsApiClient` for programmatic access without the web components:

```typescript
import { FormsApiClient } from "@modulacms/forms";

const api = new FormsApiClient("https://cms.example.com", "your-api-key");

// Public endpoints (no API key needed)
const form = await api.getPublicForm("01HXYZ...");
const result = await api.submitForm("01HXYZ...", { email: "user@example.com" });

// Admin CRUD
const forms = await api.listForms({ limit: 10 });
const created = await api.createForm({ name: "Survey", submit_label: "Submit" });
const updated = await api.updateForm("01HXYZ...", { name: "Feedback", version: 2 });

// Field operations
const fields = await api.listFields("01HXYZ...");
await api.reorderFields("01HXYZ...", ["field-c", "field-a", "field-b"], 3);

// Entries
const entries = await api.listEntries("01HXYZ...", { limit: 50 });
const exported = await api.exportEntries("01HXYZ...");

// Webhooks
await api.createWebhook("01HXYZ...", {
  url: "https://hooks.slack.com/services/...",
  events: "entry.created",
  method: "POST",
});
const queue = await api.getQueueInfo("01HXYZ...");
```

The client throws on non-2xx responses with the error message from the server.

## Validation Utilities

Client-side validation functions are exported for use outside the components:

```typescript
import { validateField, validateForm } from "@modulacms/forms";

// Validate a single field
const error = validateField(fieldDef, "not-an-email");
// Returns: "Invalid email address" or null

// Validate all fields at once
const errors = validateForm(fields, formData);
// Returns: [{ field: "email", message: "Email is required" }, ...]
```

These use the same rules as the server-side Lua validation. Both must agree for a submission to succeed.

## Accessibility

All three components follow WAI-ARIA patterns:

- Labels are linked to inputs via `for`/`id` attributes
- Required fields have the `required` attribute and `aria-required="true"`
- Validation errors use `role="alert"` with `aria-live="polite"`
- Inputs with errors have `aria-invalid="true"` and `aria-describedby` pointing to the error message
- The entries table uses `role="grid"` with keyboard navigation (arrow keys, Home, End, Enter to select)
- The form builder supports keyboard reordering and palette navigation
- The honeypot field uses `aria-hidden="true"` and `tabindex="-1"` to prevent screen reader and keyboard interaction

## Browser Support

The components use Shadow DOM and Custom Elements v1, supported in all modern browsers (Chrome 54+, Firefox 63+, Safari 10.1+, Edge 79+). No polyfills are included. For older browsers, add the [@webcomponents/webcomponentsjs](https://github.com/webcomponents/polyfills) polyfill.

## Next Steps

- [Forms Plugin reference](../extending/forms-plugin.md) for the full REST API
- [Plugin SDK](../sdks/overview.md) for building custom plugin UIs
- [Webhooks](webhooks.md) for processing the delivery queue
