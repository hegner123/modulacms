package modula

// ContentStatus represents the publication lifecycle state of a content item.
// Content begins as draft and transitions to published when approved for
// public delivery. The API returns and accepts these as lowercase JSON strings.
type ContentStatus string

const (
	// ContentStatusDraft indicates content that is being edited and is not
	// yet visible through the public content delivery API.
	ContentStatusDraft ContentStatus = "draft"

	// ContentStatusPublished indicates content that has been approved and
	// is served by the public content delivery API.
	ContentStatusPublished ContentStatus = "published"
)

// FieldType identifies the data type and input behavior of a field within
// a datatype schema. Each field type determines validation rules, the UI
// widget used in the admin panel/TUI, and how the value is stored and
// serialized in API responses.
type FieldType string

const (
	// FieldTypeText is a single-line plain text input.
	FieldTypeText FieldType = "text"

	// FieldTypeTextarea is a multi-line plain text input.
	FieldTypeTextarea FieldType = "textarea"

	// FieldTypeNumber stores numeric values (integers or decimals).
	FieldTypeNumber FieldType = "number"

	// FieldTypeDate stores a calendar date without a time component (YYYY-MM-DD).
	FieldTypeDate FieldType = "date"

	// FieldTypeDatetime stores a date with time, serialized as RFC 3339 UTC.
	FieldTypeDatetime FieldType = "datetime"

	// FieldTypeBoolean stores a true/false value.
	FieldTypeBoolean FieldType = "boolean"

	// FieldTypeSelect stores a value chosen from a predefined list of options
	// configured on the field definition.
	FieldTypeSelect FieldType = "select"

	// FieldTypeMedia stores a reference to a media item ([MediaID]).
	FieldTypeMedia FieldType = "media"

	// FieldTypeID stores a content data ID (ULID). On _reference datatype
	// nodes, the composition engine resolves this value to fetch and attach
	// referenced subtrees at delivery time.
	FieldTypeID FieldType = "_id"

	// FieldTypeJSON stores arbitrary JSON data. The value is preserved as-is
	// without schema validation.
	FieldTypeJSON FieldType = "json"

	// FieldTypeRichtext stores formatted text with markup (HTML or a structured
	// rich text format).
	FieldTypeRichtext FieldType = "richtext"

	// FieldTypeSlug stores a URL-safe slug, typically auto-generated from a
	// title field. Validated to contain only lowercase letters, numbers, and hyphens.
	FieldTypeSlug FieldType = "slug"

	// FieldTypeEmail stores an email address with format validation.
	FieldTypeEmail FieldType = "email"

	// FieldTypeURL stores a URL string with format validation.
	FieldTypeURL FieldType = "url"
)

// RouteType classifies how a route maps incoming URL patterns to content or
// behavior. Routes are the bridge between URLs that frontend clients request
// and the content items that serve them.
type RouteType string

const (
	// RouteTypeStatic maps a fixed URL path to a single content item.
	// Example: "/about" always resolves to the About page.
	RouteTypeStatic RouteType = "static"

	// RouteTypeDynamic maps a URL pattern with parameters to content items
	// resolved at request time. Example: "/blog/:slug" resolves based on
	// the slug parameter.
	RouteTypeDynamic RouteType = "dynamic"

	// RouteTypeAPI maps a URL path to a custom API endpoint, typically
	// registered by a plugin.
	RouteTypeAPI RouteType = "api"

	// RouteTypeRedirect maps a URL path to a redirect target. The API
	// returns redirect metadata (target URL, status code) rather than
	// issuing an HTTP redirect directly.
	RouteTypeRedirect RouteType = "redirect"
)
