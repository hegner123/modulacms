package modula

// ContentStatus represents the publication status of a content item.
type ContentStatus string

const (
	ContentStatusDraft     ContentStatus = "draft"
	ContentStatusPublished ContentStatus = "published"
	ContentStatusArchived  ContentStatus = "archived"
	ContentStatusPending   ContentStatus = "pending"
)

// FieldType represents the data type of a field.
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeTextarea FieldType = "textarea"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeDatetime FieldType = "datetime"
	FieldTypeBoolean  FieldType = "boolean"
	FieldTypeSelect   FieldType = "select"
	FieldTypeMedia    FieldType = "media"
	FieldTypeRelation FieldType = "relation"
	FieldTypeJSON     FieldType = "json"
	FieldTypeRichtext FieldType = "richtext"
	FieldTypeSlug     FieldType = "slug"
	FieldTypeEmail    FieldType = "email"
	FieldTypeURL      FieldType = "url"
)

// RouteType represents the type of a route.
type RouteType string

const (
	RouteTypeStatic   RouteType = "static"
	RouteTypeDynamic  RouteType = "dynamic"
	RouteTypeAPI      RouteType = "api"
	RouteTypeRedirect RouteType = "redirect"
)
