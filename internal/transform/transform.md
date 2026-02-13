# transform

The transform package provides bidirectional conversion between ModulaCMS internal format and external CMS formats like Contentful, Sanity, Strapi, WordPress, and Clean JSON.

## Overview

This package implements the Transformer interface for converting ModulaCMS content trees to and from various CMS formats. It supports both outbound transformation for API responses and inbound parsing for accepting content from external systems.

## Constants

Package constants define supported output formats by aliasing config package formats.

### FormatContentful

```go
const FormatContentful = config.FormatContentful
```

Contentful JSON API format. Includes sys metadata and nested fields structure.

### FormatSanity

```go
const FormatSanity = config.FormatSanity
```

Sanity.io document format with underscore-prefixed system fields and portable text blocks.

### FormatStrapi

```go
const FormatStrapi = config.FormatStrapi
```

Strapi v4 JSON API format with data and attributes structure.

### FormatWordPress

```go
const FormatWordPress = config.FormatWordPress
```

WordPress REST API format with rendered fields and ACF custom fields.

### FormatClean

```go
const FormatClean = config.FormatClean
```

Clean flat JSON format without CMS-specific metadata.

### FormatRaw

```go
const FormatRaw = config.FormatRaw
```

Raw ModulaCMS internal format without transformation.

## Types

### OutputFormat

```go
type OutputFormat = config.OutputFormat
```

OutputFormat is a type alias for config.OutputFormat. Specifies the target CMS format for transformation.

### Transformer

```go
type Transformer interface {
    Transform(root model.Root) (any, error)
    TransformToJSON(root model.Root) ([]byte, error)
    Parse(data []byte) (model.Root, error)
    ParseToNode(data []byte) (*model.Node, error)
}
```

Transformer converts ModulaCMS data to different CMS formats. Transform and TransformToJSON handle outbound conversion. Parse and ParseToNode handle inbound conversion from external formats.

### TransformConfig

```go
type TransformConfig struct {
    Format  OutputFormat
    SiteURL string
    SpaceID string
    Driver  db.DbDriver
}
```

TransformConfig holds configuration for transformation including format selection, site URL for link generation, space ID for multi-tenant systems, and database driver.

### BaseTransformer

```go
type BaseTransformer struct {
    SiteURL string
    SpaceID string
}
```

BaseTransformer provides common functionality for all transformers. Embedded in concrete transformer implementations.

### RawTransformer

```go
type RawTransformer struct{}
```

RawTransformer returns ModulaCMS data without transformation. Used when format is empty or explicitly set to raw.

### ContentfulTransformer

```go
type ContentfulTransformer struct {
    BaseTransformer
}
```

ContentfulTransformer transforms ModulaCMS data to Contentful format with sys metadata, content types, and asset references.

### ContentfulEntry

```go
type ContentfulEntry struct {
    Sys    ContentfulSys
    Fields map[string]any
}
```

ContentfulEntry represents a Contentful entry with system metadata and field data.

### ContentfulSys

```go
type ContentfulSys struct {
    ID          string
    Type        string
    ContentType ContentfulContentType
    Space       *ContentfulSpace
    CreatedAt   string
    UpdatedAt   string
    Revision    int
}
```

ContentfulSys contains Contentful system metadata including ID, type, content type reference, space reference, timestamps, and revision number.

### ContentfulContentType

```go
type ContentfulContentType struct {
    Sys ContentfulLink
}
```

ContentfulContentType wraps a link to the content type definition.

### ContentfulLink

```go
type ContentfulLink struct {
    ID       string
    Type     string
    LinkType string
}
```

ContentfulLink represents a Contentful reference with ID, type, and link type.

### ContentfulSpace

```go
type ContentfulSpace struct {
    Sys ContentfulLink
}
```

ContentfulSpace wraps a link to the Contentful space.

### ContentfulAsset

```go
type ContentfulAsset struct {
    Sys ContentfulLink
}
```

ContentfulAsset wraps a link to a Contentful asset like images or files.

### SanityTransformer

```go
type SanityTransformer struct {
    BaseTransformer
}
```

SanityTransformer transforms ModulaCMS data to Sanity format with underscore-prefixed system fields, portable text blocks, and image references.

### SanityDocument

```go
type SanityDocument struct {
    ID        string
    Type      string
    CreatedAt string
    UpdatedAt string
    Rev       string
    Fields    map[string]any
}
```

SanityDocument represents a Sanity document with system fields and flattened field structure. Fields are merged into the document root.

### SanitySlug

```go
type SanitySlug struct {
    Current string
    Type    string
}
```

SanitySlug wraps a slug value with Sanity type metadata.

### SanityImage

```go
type SanityImage struct {
    Type  string
    Asset SanityReference
}
```

SanityImage represents a Sanity image field with asset reference.

### SanityReference

```go
type SanityReference struct {
    Ref  string
    Type string
}
```

SanityReference points to another Sanity document or asset by ID.

### SanityBlock

```go
type SanityBlock struct {
    Type     string
    Children []SanityBlockChild
}
```

SanityBlock represents a portable text block with children. Used for rich text content.

### SanityBlockChild

```go
type SanityBlockChild struct {
    Type string
    Text string
}
```

SanityBlockChild represents a span of text within a portable text block.

### StrapiTransformer

```go
type StrapiTransformer struct {
    BaseTransformer
}
```

StrapiTransformer transforms ModulaCMS data to Strapi v4 format with data wrapper, attributes, and relation structure.

### StrapiResponse

```go
type StrapiResponse struct {
    Data any
    Meta map[string]any
}
```

StrapiResponse wraps Strapi API response with data and meta sections.

### StrapiEntry

```go
type StrapiEntry struct {
    ID         string
    Attributes map[string]any
}
```

StrapiEntry represents a Strapi content entry with ID and attributes.

### StrapiMedia

```go
type StrapiMedia struct {
    Data StrapiMediaData
}
```

StrapiMedia wraps media data in Strapi format.

### StrapiMediaData

```go
type StrapiMediaData struct {
    ID         int64
    Attributes StrapiMediaAttributes
}
```

StrapiMediaData contains media ID and attributes.

### StrapiMediaAttributes

```go
type StrapiMediaAttributes struct {
    URL             string
    Name            string
    AlternativeText string
}
```

StrapiMediaAttributes holds media metadata including URL, filename, and alt text.

### StrapiRelation

```go
type StrapiRelation struct {
    Data []StrapiEntry
}
```

StrapiRelation wraps an array of related Strapi entries.

### WordPressTransformer

```go
type WordPressTransformer struct {
    BaseTransformer
}
```

WordPressTransformer transforms ModulaCMS data to WordPress REST API format with rendered fields, ACF custom fields, and post metadata.

### WordPressPost

```go
type WordPressPost struct {
    ID            string
    Date          string
    DateGMT       string
    Modified      string
    ModifiedGMT   string
    Slug          string
    Status        string
    Type          string
    Link          string
    Title         WordPressRendered
    Content       WordPressContent
    Excerpt       WordPressContent
    Author        string
    FeaturedMedia int64
    CommentStatus string
    PingStatus    string
    Meta          map[string]any
    ACF           map[string]any
}
```

WordPressPost represents a WordPress post or page with all standard fields including title, content, excerpt, author, featured image, and custom ACF fields.

### WordPressRendered

```go
type WordPressRendered struct {
    Rendered string
}
```

WordPressRendered wraps rendered HTML content for title fields.

### WordPressContent

```go
type WordPressContent struct {
    Rendered  string
    Protected bool
}
```

WordPressContent wraps rendered HTML content with protection status for content and excerpt fields.

### CleanTransformer

```go
type CleanTransformer struct {
    BaseTransformer
}
```

CleanTransformer transforms ModulaCMS data to clean flat format without CMS-specific metadata. Fields are flattened to root level with meta prefix for system fields.

### CleanDocument

```go
type CleanDocument struct {
    ID   string
    Type string
    Meta CleanMeta
    Data map[string]any
}
```

CleanDocument represents a clean ModulaCMS document with ID, type, metadata, and flattened field data.

### CleanMeta

```go
type CleanMeta struct {
    AuthorID     string
    RouteID      string
    DateCreated  string
    DateModified string
}
```

CleanMeta holds system metadata separate from content fields.

## Functions

### NewTransformConfig

```go
func NewTransformConfig(format OutputFormat, siteURL string, spaceID string, driver db.DbDriver) *TransformConfig
```

NewTransformConfig creates a new TransformConfig from typed format value and configuration parameters.

### NewTransformConfigFromString

```go
func NewTransformConfigFromString(format string, siteURL string, spaceID string, driver db.DbDriver) *TransformConfig
```

NewTransformConfigFromString creates a new TransformConfig from string format value. Converts format string to OutputFormat type.

### GetTransformer

```go
func (tc *TransformConfig) GetTransformer() (Transformer, error)
```

GetTransformer returns the appropriate transformer implementation based on config format. Returns error for unknown formats.

### TransformAndWrite

```go
func (tc *TransformConfig) TransformAndWrite(w http.ResponseWriter, root model.Root) error
```

TransformAndWrite transforms data and writes JSON response for outbound API responses. Sets Content-Type header and writes marshaled JSON to response writer.

### ParseRequest

```go
func (tc *TransformConfig) ParseRequest(r *http.Request) (model.Root, error)
```

ParseRequest parses incoming request body from CMS format to ModulaCMS for inbound content ingestion. Reads request body and delegates to transformer Parse method.

### Transform (RawTransformer)

```go
func (r *RawTransformer) Transform(root model.Root) (any, error)
```

Transform returns ModulaCMS root without modification. No transformation applied for raw format.

### TransformToJSON (RawTransformer)

```go
func (r *RawTransformer) TransformToJSON(root model.Root) ([]byte, error)
```

TransformToJSON marshals ModulaCMS root directly to JSON without transformation.

### Parse (RawTransformer)

```go
func (r *RawTransformer) Parse(data []byte) (model.Root, error)
```

Parse unmarshals JSON directly to ModulaCMS root without transformation.

### ParseToNode (RawTransformer)

```go
func (r *RawTransformer) ParseToNode(data []byte) (*model.Node, error)
```

ParseToNode unmarshals JSON to ModulaCMS root and returns the node. Used when only node data is needed.

### Transform (ContentfulTransformer)

```go
func (c *ContentfulTransformer) Transform(root model.Root) (any, error)
```

Transform converts ModulaCMS root to Contentful entry format with sys metadata and fields.

### TransformToJSON (ContentfulTransformer)

```go
func (c *ContentfulTransformer) TransformToJSON(root model.Root) ([]byte, error)
```

TransformToJSON converts ModulaCMS root to Contentful format and marshals to JSON.

### Parse (ContentfulTransformer)

```go
func (c *ContentfulTransformer) Parse(data []byte) (model.Root, error)
```

Parse converts Contentful entry JSON to ModulaCMS root. Extracts sys metadata and converts fields back to ModulaCMS format.

### ParseToNode (ContentfulTransformer)

```go
func (c *ContentfulTransformer) ParseToNode(data []byte) (*model.Node, error)
```

ParseToNode converts Contentful entry JSON to ModulaCMS node. Returns only the node portion of parsed root.

### Transform (SanityTransformer)

```go
func (s *SanityTransformer) Transform(root model.Root) (any, error)
```

Transform converts ModulaCMS root to Sanity document format with underscore-prefixed system fields.

### TransformToJSON (SanityTransformer)

```go
func (s *SanityTransformer) TransformToJSON(root model.Root) ([]byte, error)
```

TransformToJSON converts ModulaCMS root to Sanity format and marshals to JSON.

### Parse (SanityTransformer)

```go
func (s *SanityTransformer) Parse(data []byte) (model.Root, error)
```

Parse stub returns error. Sanity inbound parsing not yet implemented. Use format clean or raw for POST and PUT requests.

### ParseToNode (SanityTransformer)

```go
func (s *SanityTransformer) ParseToNode(data []byte) (*model.Node, error)
```

ParseToNode stub returns error. Sanity inbound parsing not yet implemented. Use format clean or raw for POST and PUT requests.

### Transform (StrapiTransformer)

```go
func (s *StrapiTransformer) Transform(root model.Root) (any, error)
```

Transform converts ModulaCMS root to Strapi v4 response format with data wrapper and attributes.

### TransformToJSON (StrapiTransformer)

```go
func (s *StrapiTransformer) TransformToJSON(root model.Root) ([]byte, error)
```

TransformToJSON converts ModulaCMS root to Strapi format and marshals to JSON.

### Parse (StrapiTransformer)

```go
func (s *StrapiTransformer) Parse(data []byte) (model.Root, error)
```

Parse stub returns error. Strapi inbound parsing not yet implemented. Use format clean or raw for POST and PUT requests.

### ParseToNode (StrapiTransformer)

```go
func (s *StrapiTransformer) ParseToNode(data []byte) (*model.Node, error)
```

ParseToNode stub returns error. Strapi inbound parsing not yet implemented. Use format clean or raw for POST and PUT requests.

### Transform (WordPressTransformer)

```go
func (w *WordPressTransformer) Transform(root model.Root) (any, error)
```

Transform converts ModulaCMS root to WordPress post format with rendered fields and ACF custom fields.

### TransformToJSON (WordPressTransformer)

```go
func (w *WordPressTransformer) TransformToJSON(root model.Root) ([]byte, error)
```

TransformToJSON converts ModulaCMS root to WordPress format and marshals to JSON.

### Parse (WordPressTransformer)

```go
func (w *WordPressTransformer) Parse(data []byte) (model.Root, error)
```

Parse stub returns error. WordPress inbound parsing not yet implemented. Use format clean or raw for POST and PUT requests.

### ParseToNode (WordPressTransformer)

```go
func (w *WordPressTransformer) ParseToNode(data []byte) (*model.Node, error)
```

ParseToNode stub returns error. WordPress inbound parsing not yet implemented. Use format clean or raw for POST and PUT requests.

### Transform (CleanTransformer)

```go
func (c *CleanTransformer) Transform(root model.Root) (any, error)
```

Transform converts ModulaCMS root to clean flat JSON format with metadata prefix and flattened fields.

### TransformToJSON (CleanTransformer)

```go
func (c *CleanTransformer) TransformToJSON(root model.Root) ([]byte, error)
```

TransformToJSON converts ModulaCMS root to clean format and marshals to JSON.

### Parse (CleanTransformer)

```go
func (c *CleanTransformer) Parse(data []byte) (model.Root, error)
```

Parse converts clean JSON format to ModulaCMS root. Extracts ID, type, metadata, and converts fields back to ModulaCMS format.

### ParseToNode (CleanTransformer)

```go
func (c *CleanTransformer) ParseToNode(data []byte) (*model.Node, error)
```

ParseToNode converts clean JSON to ModulaCMS node. Returns only the node portion of parsed root.

### TransformToJSON (BaseTransformer)

```go
func (b *BaseTransformer) TransformToJSON(transformer Transformer, root model.Root) ([]byte, error)
```

TransformToJSON helper transforms root using provided transformer and marshals result to JSON. Used by concrete transformer implementations.

### fieldLabelToKey

```go
func fieldLabelToKey(label string) string
```

fieldLabelToKey converts field labels to camelCase keys. Title becomes title, Featured Image becomes featuredImage, SEO Meta Description becomes seoMetaDescription.

### splitWords

```go
func splitWords(s string) []string
```

splitWords splits a string into words by space, underscore, and hyphen delimiters.

### toLowerCase

```go
func toLowerCase(s string) string
```

toLowerCase converts string to lowercase by adding 32 to uppercase ASCII characters.

### toTitleCase

```go
func toTitleCase(s string) string
```

toTitleCase converts string to TitleCase with first character uppercase and remaining lowercase.

### pluralize

```go
func pluralize(word string) string
```

pluralize converts singular word to plural using simplified rules. Already plural words unchanged, words ending in y converted to ies, others get s appended.

### allChildrenSameType

```go
func allChildrenSameType(nodes []*model.Node) bool
```

allChildrenSameType returns true if every child node has the same datatype label. Used to determine if children should be grouped by type or as generic children array.
