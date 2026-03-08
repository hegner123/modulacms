# Localization

ModulaCMS supports multi-language content through a locale system. You define locales (languages or regional variants), mark fields as translatable, and create locale-specific field values for content nodes. Content delivery respects the requested locale, falling back to the default locale when a translation is missing.

## Concepts

**Locale** -- A language or regional variant identified by a code (e.g., `en`, `en-US`, `fr-FR`, `ja`). Locales are stored in the database and can be created, enabled, disabled, or deleted at runtime.

**Default locale** -- The primary content locale. Content created without a locale is assigned to the default locale. When a requested locale has no translation for a field, the default locale value is used as a fallback.

**Translatable field** -- A field definition with its `translatable` flag set to a non-zero value. Only translatable fields get per-locale values. Non-translatable fields (like a slug or sort order) share a single value across all locales.

**Translation** -- A set of locale-specific content field values for a content node. Creating a translation copies all translatable field values from the default locale into the target locale as a starting point.

## Managing Locales

### Creating a Locale

```bash
curl -X POST http://localhost:8080/api/v1/locales \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "fr-FR",
    "label": "French (France)",
    "is_default": false,
    "is_enabled": true,
    "fallback_code": "en",
    "sort_order": 2
  }'
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | string | Yes | Locale identifier (e.g., `en`, `fr-FR`, `ja`) |
| `label` | string | Yes | Human-readable name (e.g., "French (France)") |
| `is_default` | bool | No | Set as the default locale. Setting this to true clears the flag on the previous default. |
| `is_enabled` | bool | No | Whether this locale is active for content translation |
| `fallback_code` | string | No | Locale code to fall back to when a translation is missing |
| `sort_order` | int | No | Display ordering in locale pickers |

### Listing Locales

List all locales:

```bash
curl http://localhost:8080/api/v1/locales \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

List only enabled locales:

```bash
curl "http://localhost:8080/api/v1/locales?enabled=true" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

### Updating a Locale

```bash
curl -X PUT http://localhost:8080/api/v1/locales/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "locale_id": "01JNRWFQ8JOVY1U9S3B6D0M7HP",
    "code": "fr-FR",
    "label": "Francais (France)",
    "is_default": false,
    "is_enabled": true,
    "fallback_code": "en",
    "sort_order": 2
  }'
```

### Enabling and Disabling Locales

Disabling a locale prevents new translations from being created in that locale but does not delete existing translations. Re-enabling the locale restores access to existing translations.

```bash
# Disable a locale
curl -X PUT http://localhost:8080/api/v1/locales/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "locale_id": "01JNRWFQ8JOVY1U9S3B6D0M7HP",
    "code": "fr-FR",
    "label": "Francais (France)",
    "is_default": false,
    "is_enabled": false,
    "sort_order": 2
  }'
```

### Deleting a Locale

```bash
curl -X DELETE "http://localhost:8080/api/v1/locales/?q=01JNRWFQ8JOVY1U9S3B6D0M7HP" \
  -H "Cookie: session=YOUR_SESSION_COOKIE"
```

## Creating Translations

Create locale-specific field values for a content node. This copies all translatable field definitions from the default locale into the target locale as a starting point for translation.

```bash
curl -X POST http://localhost:8080/api/v1/admin/contentdata/01JNRWBM4FNRZ7R5N9X4C6K8DM/translations \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "locale": "fr-FR"
  }'
```

Response:

```json
{
  "locale": "fr-FR",
  "fields_created": 4
}
```

`fields_created` indicates how many translatable field values were copied. Non-translatable fields are not duplicated.

After creating the translation, update individual field values to the translated content:

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

## Content Delivery with Locale

Request content in a specific locale using the `locale` query parameter on the content delivery endpoint:

```bash
curl "http://localhost:8080/api/v1/content/homepage?locale=fr-FR"
```

When a locale is specified:
1. Translatable fields return the value for the requested locale if one exists.
2. If no translation exists for the requested locale, the fallback locale is checked.
3. If no fallback exists, the default locale value is returned.
4. Non-translatable fields always return the same value regardless of locale.

When no locale parameter is provided, the default locale is used.

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
    Code:      "fr-FR",
    Label:     "French (France)",
    IsDefault: false,
    IsEnabled: true,
    FallbackCode: "en",
    SortOrder: 2,
})

// List all locales
locales, err := client.Locales.List(ctx, nil)

// List only enabled locales
enabled, err := client.Locales.ListEnabled(ctx)

// Create a translation for a content node
translation, err := client.Locales.CreateTranslation(ctx,
    "01JNRWBM4FNRZ7R5N9X4C6K8DM",
    modula.CreateTranslationRequest{
        Locale: "fr-FR",
    },
)

// Fetch content in a specific locale using the query API
result, err := client.Query.Query(ctx, "blog-posts", &modula.QueryParams{
    Locale: "fr-FR",
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
  code: 'fr-FR',
  label: 'French (France)',
  is_default: false,
  is_enabled: true,
  fallback_code: 'en',
  sort_order: 2,
})

// List enabled locales
const enabled = await client.locales.listEnabled()

// Create a translation
const translation = await client.locales.createTranslation(
  '01JNRWBM4FNRZ7R5N9X4C6K8DM',
  { locale: 'fr-FR' }
)

// Query content in a specific locale
const result = await client.query('blog-posts', {
  locale: 'fr-FR',
  status: 'published',
})
```

## API Reference

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/locales` | `locales:read` | List all locales (supports `?enabled=true`) |
| POST | `/api/v1/locales` | `locales:create` | Create a locale |
| GET | `/api/v1/locales/` | `locales:read` | Get a single locale (`?q=LOCALE_ID`) |
| PUT | `/api/v1/locales/` | `locales:update` | Update a locale |
| DELETE | `/api/v1/locales/` | `locales:delete` | Delete a locale (`?q=LOCALE_ID`) |
| POST | `/api/v1/admin/contentdata/{id}/translations` | `content:create` | Create translation field values for a content node |

## Notes

- **Fallback chain.** When a translation is missing, the system checks the locale's `fallback_code`. If that locale also lacks a translation, the default locale is used. Chains longer than two levels are not supported -- fallback always terminates at the default locale.
- **Content field locale column.** Each content field record has a `locale` column. The default locale uses an empty string. Locale-specific values use the locale code. This allows multiple locale values per field on the same content node.
- **Translation as copy.** Creating a translation copies current field values as a starting point. After creation, the translated fields are independent -- changes to the default locale do not propagate to translations.
- **Disabled locales persist.** Disabling a locale hides it from `ListEnabled` queries but does not delete translation data. Content delivery still returns disabled locale content if explicitly requested.
- **Publishing per locale.** Publishing supports a `locale` parameter to publish locale-specific content independently. See [publishing.md](publishing.md) for details.
