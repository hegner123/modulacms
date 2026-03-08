# Localization

ModulaCMS supports content in multiple languages through a locale-aware field system. Locales represent languages or regions, content fields can store per-locale values, and the delivery API resolves the correct value through a fallback chain. The system does not duplicate content nodes for each locale -- instead, it stores locale-specific variants at the field level.

## Locales

A locale represents a language/region combination configured in the CMS.

```go
type Locale struct {
    LocaleID     LocaleID `json:"locale_id"`
    Code         string   `json:"code"`
    Label        string   `json:"label"`
    IsDefault    bool     `json:"is_default"`
    IsEnabled    bool     `json:"is_enabled"`
    FallbackCode string   `json:"fallback_code"`
    SortOrder    int64    `json:"sort_order"`
    DateCreated  string   `json:"date_created"`
}
```

| Field | Purpose |
|-------|---------|
| `Code` | Language/region identifier following BCP 47 conventions (e.g., `en`, `en-US`, `fr-CA`) |
| `Label` | Human-readable name (e.g., `English (US)`, `French (Canada)`) |
| `IsDefault` | True for the primary content locale. Exactly one locale is the default. |
| `IsEnabled` | False disables the locale without deleting it. Disabled locales are excluded from content delivery. |
| `FallbackCode` | Locale code to fall back to when content is not available in this locale (e.g., `en-US` falls back to `en`) |
| `SortOrder` | Display ordering in locale selection UI |

### Managing Locales

```bash
# Create a locale
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

Setting `is_default: true` on a new locale clears the default flag on the previous default locale.

### Listing Enabled Locales

```go
// All locales (including disabled)
locales, err := client.Locales.List(ctx, nil)

// Only enabled locales
enabled, err := client.Locales.ListEnabled(ctx)
```

## Translatable Fields

Whether a field supports per-locale values is controlled by the `Translatable` field on the field definition. When `Translatable` is non-zero (typically `1`), the field can store different values for each locale.

```go
type Field struct {
    // ... other fields
    Translatable int64 `json:"translatable"`
}
```

Fields that are not translatable (e.g., a slug, a sort order, a media reference) store a single value that is the same across all locales. Translatable fields (e.g., a title, a body, a meta description) store separate values per locale.

## How Locale-Specific Values Are Stored

Content field values include a `Locale` field that identifies which locale the value belongs to:

```go
type ContentField struct {
    ContentFieldID ContentFieldID `json:"content_field_id"`
    ContentDataID  *ContentID     `json:"content_data_id"`
    FieldID        *FieldID       `json:"field_id"`
    FieldValue     string         `json:"field_value"`
    Locale         string         `json:"locale"`
    // ...
}
```

For a translatable field on a single content node, there is one `ContentField` row per locale. A non-translatable field has one row with the default locale code.

Example: A content node with a translatable `title` field and locales `en` and `fr`:

| ContentDataID | FieldID | Locale | FieldValue |
|--------------|---------|--------|------------|
| 01ABC... | 01DEF... | en | About Us |
| 01ABC... | 01DEF... | fr | A propos de nous |

## Creating Translations

The `CreateTranslation` endpoint copies all translatable field values from the default locale into a target locale as a starting point. You then update individual field values with the translated text.

```go
resp, err := client.Locales.CreateTranslation(ctx, contentDataID, modula.CreateTranslationRequest{
    Locale: "fr",
})
// resp.FieldsCreated == 5 (number of translatable fields copied)
```

This creates `ContentField` rows for the target locale with the default locale's values. If translations already exist for the target locale, they are updated rather than duplicated.

After creating the translation scaffold, update individual fields:

```bash
curl -X PUT http://localhost:8080/api/v1/contentfields/ \
  -H "Cookie: session=YOUR_SESSION_COOKIE" \
  -H "Content-Type: application/json" \
  -d '{
    "content_field_id": "01GHI...",
    "field_value": "A propos de nous",
    "content_data_id": "01ABC...",
    "field_id": "01DEF..."
  }'
```

## Fallback Chain

When content is requested in a specific locale, the delivery API resolves field values through a fallback chain:

1. **Requested locale** -- use the value for the exact locale code if it exists.
2. **Fallback locale** -- follow the `FallbackCode` chain (e.g., `fr-CA` falls back to `fr`, which falls back to `en`).
3. **Default locale** -- use the default locale's value if no fallback matches.
4. **First available** -- use whatever locale value exists if all above fail.

This chain is traversed per field. One field on a content node might resolve from `fr-CA` while another field on the same node falls back to `en`, depending on which translations are available.

## Locale-Aware Content Delivery

Request content in a specific locale with the `locale` query parameter:

```bash
curl "http://localhost:8080/api/v1/content/about?locale=fr"
```

When no `locale` parameter is specified, the default locale is used.

The delivered content tree includes the resolved field values for the requested locale. The frontend does not need to handle fallback logic -- the CMS resolves the correct value before delivery.

## Versioning and Locales

Content versions are locale-aware. The `ContentVersion.Locale` field records which locale the snapshot applies to. Publishing content can create locale-specific version snapshots, and restoring a version restores the field values for that specific locale.

```go
type ContentVersion struct {
    // ...
    Locale string `json:"locale"`
}
```

## API Endpoints

| Method | Path | Permission | Description |
|--------|------|------------|-------------|
| GET | `/api/v1/locales` | `locales:read` | List all locales |
| POST | `/api/v1/locales` | `locales:create` | Create a locale |
| GET | `/api/v1/locales/` | `locales:read` | Get a locale (`?q=ID`) |
| PUT | `/api/v1/locales/` | `locales:update` | Update a locale |
| DELETE | `/api/v1/locales/` | `locales:delete` | Delete a locale |
| POST | `/api/v1/admin/contentdata/{id}/translations` | `content:create` | Create translation field values for a content node |
