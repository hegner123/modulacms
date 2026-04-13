# Localization

Serve content in multiple languages with per-field translations and automatic fallback resolution.

## Concepts

**Locale** -- A language or regional variant identified by a code following BCP 47 conventions (e.g., `en`, `en-US`, `fr-CA`, `ja`). Locales are managed at runtime and can be created, enabled, disabled, or deleted through the API.

**Default locale** -- The primary content locale. Content created without specifying a locale is assigned to the default locale. Exactly one locale is the default at any time.

**Translatable field** -- A field definition with its `translatable` flag enabled. Translatable fields store one value per locale. Non-translatable fields (like a slug or sort order) store a single value shared across all locales.

**Translation** -- A set of locale-specific field values for a content node. Creating a translation copies all translatable field values from the default locale into the target locale as a starting point.

## Create Locales

Define the languages your content supports:

```bash
curl -X POST http://localhost:8080/api/v1/locales \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "fr",
    "label": "French",
    "is_default": false,
    "is_enabled": true,
    "fallback_code": "en",
    "sort_order": 2
  }'
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | string | Yes | Locale identifier (e.g., `en`, `fr-FR`, `ja`) |
| `label` | string | Yes | Human-readable name (e.g., "French", "Japanese") |
| `is_default` | bool | No | Set as the default locale |
| `is_enabled` | bool | No | Whether this locale is active for content translation |
| `fallback_code` | string | No | Locale code to fall back to when a translation is missing |
| `sort_order` | int | No | Display ordering in locale pickers |

> **Good to know**: Setting `is_default: true` on a new locale automatically clears the default flag on the previous default locale.

### List Locales

```bash
# All locales
curl http://localhost:8080/api/v1/locales \
  -H "Cookie: session=YOUR_SESSION_COOKIE"

# Only enabled locales
curl "http://localhost:8080/api/v1/locales?enabled=true" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Disable a Locale

Disabling a locale prevents new translations from being created in that locale but does not delete existing translations. Re-enabling the locale restores access.

```bash
curl -X PUT http://localhost:8080/api/v1/locales/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "locale_id": "01JNRWFQ8JOVY1U9S3B6D0M7HP",
    "code": "fr",
    "label": "French",
    "is_default": false,
    "is_enabled": false,
    "sort_order": 2
  }'
```

## Mark Fields as Translatable

When you create a field definition for a datatype, set the `translatable` flag to indicate that the field should store per-locale values. Fields like title, body text, and meta descriptions are typically translatable. Fields like slug, sort order, and media references are typically not.

ModulaCMS does not duplicate content nodes for each locale. A single content node serves all languages, with locale-specific values stored at the field level.

## Translate Content

### Create a Translation Scaffold

The translation endpoint copies all translatable field values from the default locale into a target locale as a starting point:

```bash
curl -X POST http://localhost:8080/api/v1/admin/contentdata/01JNRWBM4FNRZ7R5N9X4C6K8DM/translations \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"locale": "fr"}'
```

```json
{
  "locale": "fr",
  "fields_created": 4
}
```

`fields_created` indicates how many translatable field values were copied. Non-translatable fields are not duplicated.

> **Good to know**: If translations already exist for the target locale, they are updated rather than duplicated. After creation, translated fields are independent -- changes to the default locale do not propagate to translations.

### Update Translated Fields

After creating the scaffold, update individual fields with translated content:

```bash
curl -X PUT http://localhost:8080/api/v1/contentfields/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_field_id": "01JNRWGR9KPWZ2V0T4C7E1N8IQ",
    "content_data_id": "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    "field_id": "01JNRW7K8CNQZ5P3R9W6TJ4MAS",
    "field_value": "Mon Premier Article de Blog"
  }'
```

## Fallback Chain

When you request content in a specific locale, ModulaCMS resolves each field value through a fallback chain:

1. **Requested locale** -- use the value for the exact locale code if it exists.
2. **Fallback locale** -- follow the `fallback_code` chain (e.g., `fr-CA` falls back to `fr`, which falls back to `en`).
3. **Default locale** -- use the default locale's value if no fallback matches.
4. **First available** -- use whatever locale value exists if all of the above fail.

This chain is traversed per field. One field on a content node might resolve from `fr-CA` while another field on the same node falls back to `en`, depending on which translations exist.

> **Good to know**: Your frontend does not need to handle fallback logic. ModulaCMS resolves the correct value before delivery.

### Example Fallback Setup

A site supporting English, French, and Canadian French:

```bash
# English (default)
curl -X POST http://localhost:8080/api/v1/locales \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"code": "en", "label": "English", "is_default": true, "is_enabled": true, "sort_order": 1}'

# French (falls back to English)
curl -X POST http://localhost:8080/api/v1/locales \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"code": "fr", "label": "French", "is_default": false, "is_enabled": true, "fallback_code": "en", "sort_order": 2}'

# Canadian French (falls back to French, which falls back to English)
curl -X POST http://localhost:8080/api/v1/locales \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{"code": "fr-CA", "label": "French (Canada)", "is_default": false, "is_enabled": true, "fallback_code": "fr", "sort_order": 3}'
```

## Deliver Content in a Locale

Pass the `locale` query parameter to the content delivery endpoint:

```bash
curl "http://localhost:8080/api/v1/content/about?locale=fr"
```

When no `locale` parameter is specified, ModulaCMS uses the default locale.

## Publish Per Locale

Publishing supports a `locale` parameter to publish locale-specific content independently. You can publish the English version of a page while the French translation is still in draft.

## Versioning and Locales

Content versions are locale-aware. Each version snapshot records which locale it applies to. Restoring a version restores the field values for that specific locale only, leaving other locale translations unchanged.

## SDK Examples

### Go

```go
import modula "github.com/hegner123/modulacms/sdks/go"

client, _ := modula.NewClient(modula.ClientConfig{
    BaseURL: "http://localhost:8080",
    APIKey:  "mcms_YOUR_API_KEY",
})

// Create a locale
locale, err := client.Locales.Create(ctx, modula.CreateLocaleRequest{
    Code:         "fr",
    Label:        "French",
    IsDefault:    false,
    IsEnabled:    true,
    FallbackCode: "en",
    SortOrder:    2,
})

// Create a translation scaffold for a content node
translation, err := client.Locales.CreateTranslation(ctx,
    "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    modula.CreateTranslationRequest{Locale: "fr"},
)

// Query content in a specific locale
result, err := client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
    Locale: "fr",
    Status: "published",
})
```

### TypeScript

```typescript
import { ModulaCMSAdmin } from '@modulacms/admin-sdk'

const client = new ModulaCMSAdmin({
  baseUrl: 'http://localhost:8080',
  apiKey: 'mcms_YOUR_API_KEY',
})

// Create a locale
const locale = await client.locales.create({
  code: 'fr',
  label: 'French',
  is_default: false,
  is_enabled: true,
  fallback_code: 'en',
  sort_order: 2,
})

// Create a translation scaffold
const translation = await client.locales.createTranslation(
  '01JNRWBM4FNRZ7R5N9X4C6K8DM',
  { locale: 'fr' }
)

// Query content in a specific locale
const result = await client.query('blog-posts', {
  locale: 'fr',
  status: 'published',
})
```

## API Reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/locales` | `locales:read` | List all locales (supports `?enabled=true`) |
| POST | `/api/v1/locales` | `locales:create` | Create a locale |
| GET | `/api/v1/locales/` | `locales:read` | Get a locale (`?q=LOCALE_ID`) |
| PUT | `/api/v1/locales/` | `locales:update` | Update a locale |
| DELETE | `/api/v1/locales/` | `locales:delete` | Delete a locale (`?q=LOCALE_ID`) |
| POST | `/api/v1/admin/contentdata/{id}/translations` | `content:create` | Create translation field values |

## Next Steps

- [Content modeling guide](/docs/building-content/content-modeling) -- define datatypes and fields (including the translatable flag)
- [Deploy sync](/docs/integrations/deploy-sync) -- export and import translated content between environments
- [Configuration reference](/docs/getting-started/configuration) -- all config fields
