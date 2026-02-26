package transform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

// ===========================================================================
// ContentfulTransformer
// ===========================================================================

func TestContentfulTransformer_Transform_NilNode(t *testing.T) {
	t.Parallel()

	ct := &ContentfulTransformer{}
	result, err := ct.Transform(model.Root{Node: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry, ok := result.(ContentfulEntry)
	if !ok {
		t.Fatalf("expected ContentfulEntry, got %T", result)
	}

	// Empty entry for nil node
	if entry.Sys.ID != "" {
		t.Errorf("Sys.ID = %q, want empty", entry.Sys.ID)
	}
}

func TestContentfulTransformer_Transform_SimpleNode(t *testing.T) {
	t.Parallel()

	node := makeNodeWithDates("id-1", "Blog Post", "blogpost", "author1", "route1", "2024-01-01T00:00:00Z", "2024-06-15T00:00:00Z")
	node.Fields = append(node.Fields,
		makeField("Title", "text", "Hello World"),
		makeField("Published", "boolean", "true"),
		makeField("View Count", "integer", "42"),
	)

	ct := &ContentfulTransformer{BaseTransformer: BaseTransformer{SpaceID: "test-space"}}
	result, err := ct.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := result.(ContentfulEntry)

	// Check sys fields
	if entry.Sys.ID != "id-1" {
		t.Errorf("Sys.ID = %q, want %q", entry.Sys.ID, "id-1")
	}
	if entry.Sys.Type != "Entry" {
		t.Errorf("Sys.Type = %q, want %q", entry.Sys.Type, "Entry")
	}
	if entry.Sys.ContentType.Sys.ID != "blogpost" {
		t.Errorf("ContentType.Sys.ID = %q, want %q", entry.Sys.ContentType.Sys.ID, "blogpost")
	}
	if entry.Sys.ContentType.Sys.LinkType != "ContentType" {
		t.Errorf("ContentType.Sys.LinkType = %q, want %q", entry.Sys.ContentType.Sys.LinkType, "ContentType")
	}
	if entry.Sys.CreatedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %q, want %q", entry.Sys.CreatedAt, "2024-01-01T00:00:00Z")
	}
	if entry.Sys.UpdatedAt != "2024-06-15T00:00:00Z" {
		t.Errorf("UpdatedAt = %q, want %q", entry.Sys.UpdatedAt, "2024-06-15T00:00:00Z")
	}
	if entry.Sys.Revision != 1 {
		t.Errorf("Revision = %d, want 1", entry.Sys.Revision)
	}

	// Check space
	if entry.Sys.Space == nil {
		t.Fatal("Space is nil, want non-nil when SpaceID is set")
	}
	if entry.Sys.Space.Sys.ID != "test-space" {
		t.Errorf("Space.Sys.ID = %q, want %q", entry.Sys.Space.Sys.ID, "test-space")
	}

	// Check fields
	if entry.Fields["title"] != "Hello World" {
		t.Errorf("title = %v, want %q", entry.Fields["title"], "Hello World")
	}
	if entry.Fields["published"] != true {
		t.Errorf("published = %v, want true", entry.Fields["published"])
	}
	if entry.Fields["viewCount"] != int64(42) {
		t.Errorf("viewCount = %v (%T), want int64(42)", entry.Fields["viewCount"], entry.Fields["viewCount"])
	}
}

func TestContentfulTransformer_Transform_NoSpaceID(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	ct := &ContentfulTransformer{} // No SpaceID

	result, err := ct.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := result.(ContentfulEntry)
	if entry.Sys.Space != nil {
		t.Errorf("Space should be nil when SpaceID is empty, got %+v", entry.Sys.Space)
	}
}

func TestContentfulTransformer_Transform_AssetField(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Featured Image", "image", "https://example.com/photo.jpg"))

	ct := &ContentfulTransformer{}
	result, err := ct.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := result.(ContentfulEntry)
	asset, ok := entry.Fields["featuredImage"].(ContentfulAsset)
	if !ok {
		t.Fatalf("featuredImage is not ContentfulAsset, got %T", entry.Fields["featuredImage"])
	}

	if asset.Sys.Type != "Link" {
		t.Errorf("asset Sys.Type = %q, want %q", asset.Sys.Type, "Link")
	}
	if asset.Sys.LinkType != "Asset" {
		t.Errorf("asset Sys.LinkType = %q, want %q", asset.Sys.LinkType, "Asset")
	}
	// Asset ID should be non-empty (hash of URL)
	if asset.Sys.ID == "" {
		t.Error("asset Sys.ID should not be empty")
	}
}

func TestContentfulTransformer_Transform_EmptyFieldValue(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Title", "text", ""))

	ct := &ContentfulTransformer{}
	result, err := ct.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := result.(ContentfulEntry)
	if entry.Fields["title"] != nil {
		t.Errorf("empty field should produce nil, got %v", entry.Fields["title"])
	}
}

func TestContentfulTransformer_Transform_ChildNodes(t *testing.T) {
	t.Parallel()

	parent := makeNode("parent-1", "Blog", "blog")
	child1 := makeNode("child-1", "Post", "post")
	child2 := makeNode("child-2", "Post", "post")
	parent.Nodes = []*model.Node{child1, child2}

	ct := &ContentfulTransformer{}
	result, err := ct.Transform(model.Root{Node: parent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := result.(ContentfulEntry)

	// Same type children grouped under pluralized label
	posts, ok := entry.Fields["posts"]
	if !ok {
		t.Fatal("expected 'posts' key for same-type children")
	}

	postSlice, ok := posts.([]ContentfulEntry)
	if !ok {
		t.Fatalf("posts is not []ContentfulEntry, got %T", posts)
	}
	if len(postSlice) != 2 {
		t.Errorf("got %d posts, want 2", len(postSlice))
	}
}

func TestContentfulTransformer_Transform_MixedChildNodes(t *testing.T) {
	t.Parallel()

	parent := makeNode("parent-1", "Page", "page")
	child1 := makeNode("child-1", "Post", "post")
	child2 := makeNode("child-2", "Widget", "widget")
	parent.Nodes = []*model.Node{child1, child2}

	ct := &ContentfulTransformer{}
	result, err := ct.Transform(model.Root{Node: parent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entry := result.(ContentfulEntry)

	children, ok := entry.Fields["children"]
	if !ok {
		t.Fatal("expected 'children' key for mixed-type children")
	}

	childSlice, ok := children.([]ContentfulEntry)
	if !ok {
		t.Fatalf("children is not []ContentfulEntry, got %T", children)
	}
	if len(childSlice) != 2 {
		t.Errorf("got %d children, want 2", len(childSlice))
	}
}

func TestContentfulTransformer_TransformToJSON(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Title", "text", "Hello"))

	ct := &ContentfulTransformer{}
	data, err := ct.TransformToJSON(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	sys, ok := parsed["sys"].(map[string]any)
	if !ok {
		t.Fatal("missing sys object in JSON output")
	}
	if sys["id"] != "id-1" {
		t.Errorf("sys.id = %v, want %q", sys["id"], "id-1")
	}
}

func TestContentfulTransformer_ToContentfulID(t *testing.T) {
	t.Parallel()

	ct := &ContentfulTransformer{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple label", input: "Post", want: "post"},
		{name: "two word label", input: "Blog Post", want: "blogpost"},
		{name: "with numbers", input: "Post123", want: "post123"},
		{name: "with special chars", input: "My-Cool Post!", want: "mycoolpost"},
		{name: "empty", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ct.toContentfulID(tt.input)
			if got != tt.want {
				t.Errorf("toContentfulID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestContentfulTransformer_GetDateModified_Fallback(t *testing.T) {
	t.Parallel()

	ct := &ContentfulTransformer{}

	t.Run("has modified date", func(t *testing.T) {
		t.Parallel()
		node := makeNodeWithDates("id", "Post", "post", "", "", "2024-01-01", "2024-06-15")
		got := ct.getDateModified(node)
		if got != "2024-06-15" {
			t.Errorf("getDateModified = %q, want %q", got, "2024-06-15")
		}
	})

	t.Run("no modified date falls back to created", func(t *testing.T) {
		t.Parallel()
		node := makeNodeWithDates("id", "Post", "post", "", "", "2024-01-01", "")
		got := ct.getDateModified(node)
		if got != "2024-01-01" {
			t.Errorf("getDateModified = %q, want %q (fallback to created)", got, "2024-01-01")
		}
	})
}

// ---------------------------------------------------------------------------
// ContentfulTransformer: Parse (inbound)
// ---------------------------------------------------------------------------

func TestContentfulTransformer_Parse(t *testing.T) {
	t.Parallel()

	ct := &ContentfulTransformer{}

	t.Run("simple entry", func(t *testing.T) {
		t.Parallel()

		input := `{
			"sys": {
				"id": "entry-1",
				"type": "Entry",
				"contentType": {
					"sys": {"id": "blogpost", "type": "Link", "linkType": "ContentType"}
				},
				"createdAt": "2024-01-01T00:00:00Z",
				"updatedAt": "2024-06-15T00:00:00Z"
			},
			"fields": {
				"title": "Hello World",
				"published": true
			}
		}`

		root, err := ct.Parse([]byte(input))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		node := root.Node
		if node == nil {
			t.Fatal("Parse returned nil node")
		}

		if node.Datatype.Content.ContentDataID != "entry-1" {
			t.Errorf("ContentDataID = %q, want %q", node.Datatype.Content.ContentDataID, "entry-1")
		}
		if node.Datatype.Info.Label != "blogpost" {
			t.Errorf("Label = %q, want %q", node.Datatype.Info.Label, "blogpost")
		}
		if node.Datatype.Content.DateCreated != "2024-01-01T00:00:00Z" {
			t.Errorf("DateCreated = %q, want %q", node.Datatype.Content.DateCreated, "2024-01-01T00:00:00Z")
		}

		// Should have parsed fields
		if len(node.Fields) < 1 {
			t.Fatalf("expected at least 1 field, got %d", len(node.Fields))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()
		_, err := ct.Parse([]byte("not json"))
		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})
}

func TestContentfulTransformer_ParseToNode(t *testing.T) {
	t.Parallel()

	ct := &ContentfulTransformer{}
	input := `{
		"sys": {"id": "e1", "type": "Entry", "contentType": {"sys": {"id": "post", "type": "Link", "linkType": "ContentType"}}, "createdAt": "", "updatedAt": ""},
		"fields": {"title": "Test"}
	}`

	node, err := ct.ParseToNode([]byte(input))
	if err != nil {
		t.Fatalf("ParseToNode failed: %v", err)
	}
	if node == nil {
		t.Fatal("ParseToNode returned nil")
	}
	if node.Datatype.Content.ContentDataID != "e1" {
		t.Errorf("ContentDataID = %q, want %q", node.Datatype.Content.ContentDataID, "e1")
	}
}

func TestContentfulTransformer_CamelCaseToLabel(t *testing.T) {
	t.Parallel()

	ct := &ContentfulTransformer{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "simple camelCase", input: "featuredImage", want: "Featured Image"},
		{name: "lowercase word", input: "title", want: "Title"},
		{name: "already uppercase first", input: "Title", want: "Title"},
		{name: "multi capital", input: "seoMetaDescription", want: "Seo Meta Description"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ct.camelCaseToLabel(tt.input)
			if got != tt.want {
				t.Errorf("camelCaseToLabel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ===========================================================================
// SanityTransformer
// ===========================================================================

func TestSanityTransformer_Transform_NilNode(t *testing.T) {
	t.Parallel()

	st := &SanityTransformer{}
	result, err := st.Transform(model.Root{Node: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if len(doc) != 0 {
		t.Errorf("expected empty map for nil node, got %v", doc)
	}
}

func TestSanityTransformer_Transform_SimpleNode(t *testing.T) {
	t.Parallel()

	node := makeNodeWithDates("id-1", "Blog Post", "blogpost", "author1", "route1", "2024-01-01T00:00:00Z", "2024-06-15T00:00:00Z")
	node.Fields = append(node.Fields,
		makeField("Title", "text", "Hello Sanity"),
		makeField("Published", "boolean", "true"),
	)

	st := &SanityTransformer{}
	result, err := st.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc := result.(map[string]any)

	if doc["_id"] != "id-1" {
		t.Errorf("_id = %v, want %q", doc["_id"], "id-1")
	}
	if doc["_type"] != "blogpost" {
		t.Errorf("_type = %v, want %q", doc["_type"], "blogpost")
	}
	if doc["_createdAt"] != "2024-01-01T00:00:00Z" {
		t.Errorf("_createdAt = %v, want %q", doc["_createdAt"], "2024-01-01T00:00:00Z")
	}
	if doc["_updatedAt"] != "2024-06-15T00:00:00Z" {
		t.Errorf("_updatedAt = %v, want %q", doc["_updatedAt"], "2024-06-15T00:00:00Z")
	}
	if doc["_rev"] != "v1" {
		t.Errorf("_rev = %v, want %q", doc["_rev"], "v1")
	}
	if doc["title"] != "Hello Sanity" {
		t.Errorf("title = %v, want %q", doc["title"], "Hello Sanity")
	}
	if doc["published"] != true {
		t.Errorf("published = %v, want true", doc["published"])
	}
}

func TestSanityTransformer_Transform_SlugField(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Slug", "text", "hello-world"))

	st := &SanityTransformer{}
	result, err := st.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc := result.(map[string]any)

	// Slug fields get wrapped in SanitySlug struct
	slug, ok := doc["slug"].(SanitySlug)
	if !ok {
		t.Fatalf("slug is not SanitySlug, got %T: %v", doc["slug"], doc["slug"])
	}
	if slug.Current != "hello-world" {
		t.Errorf("slug.Current = %q, want %q", slug.Current, "hello-world")
	}
	if slug.Type != "slug" {
		t.Errorf("slug.Type = %q, want %q", slug.Type, "slug")
	}
}

func TestSanityTransformer_Transform_ImageField(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Cover", "image", "https://example.com/photo.jpg"))

	st := &SanityTransformer{}
	result, err := st.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc := result.(map[string]any)
	img, ok := doc["cover"].(SanityImage)
	if !ok {
		t.Fatalf("cover is not SanityImage, got %T", doc["cover"])
	}

	if img.Type != "image" {
		t.Errorf("image Type = %q, want %q", img.Type, "image")
	}
	if img.Asset.Type != "reference" {
		t.Errorf("asset Type = %q, want %q", img.Asset.Type, "reference")
	}
	if !strings.HasPrefix(img.Asset.Ref, "image-") {
		t.Errorf("asset Ref = %q, want prefix %q", img.Asset.Ref, "image-")
	}
}

func TestSanityTransformer_Transform_MarkdownField(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Body", "markdown", "# Hello\n\nWorld"))

	st := &SanityTransformer{}
	result, err := st.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	doc := result.(map[string]any)
	blocks, ok := doc["body"].([]SanityBlock)
	if !ok {
		t.Fatalf("body is not []SanityBlock, got %T", doc["body"])
	}

	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block from markdown")
	}

	// First block should have stripped the markdown header
	if blocks[0].Type != "block" {
		t.Errorf("block Type = %q, want %q", blocks[0].Type, "block")
	}
	if len(blocks[0].Children) == 0 {
		t.Fatal("expected at least 1 child in block")
	}
	if blocks[0].Children[0].Type != "span" {
		t.Errorf("child Type = %q, want %q", blocks[0].Children[0].Type, "span")
	}
}

func TestSanityTransformer_ToSanityType(t *testing.T) {
	t.Parallel()

	st := &SanityTransformer{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple", input: "Post", want: "post"},
		{name: "two words", input: "Blog Post", want: "blogpost"},
		{name: "with numbers", input: "Type2", want: "type2"},
		{name: "special chars removed", input: "My-Type!", want: "mytype"},
		{name: "empty", input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := st.toSanityType(tt.input)
			if got != tt.want {
				t.Errorf("toSanityType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSanityTransformer_TransformToJSON(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Title", "text", "Hello"))

	st := &SanityTransformer{}
	data, err := st.TransformToJSON(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed["_id"] != "id-1" {
		t.Errorf("_id = %v, want %q", parsed["_id"], "id-1")
	}
}

func TestSanityTransformer_GetDateModified_Fallback(t *testing.T) {
	t.Parallel()

	st := &SanityTransformer{}

	node := makeNodeWithDates("id", "Post", "post", "", "", "2024-01-01", "")
	got := st.getDateModified(node)
	if got != "2024-01-01" {
		t.Errorf("getDateModified = %q, want %q (fallback to created)", got, "2024-01-01")
	}
}

// ===========================================================================
// StrapiTransformer
// ===========================================================================

func TestStrapiTransformer_Transform_NilNode(t *testing.T) {
	t.Parallel()

	st := &StrapiTransformer{}
	result, err := st.Transform(model.Root{Node: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, ok := result.(StrapiResponse)
	if !ok {
		t.Fatalf("expected StrapiResponse, got %T", result)
	}

	// For nil node, data is an empty StrapiEntry
	entry, ok := resp.Data.(StrapiEntry)
	if !ok {
		t.Fatalf("Data is not StrapiEntry, got %T", resp.Data)
	}
	if entry.ID != "" {
		t.Errorf("entry.ID = %q, want empty", entry.ID)
	}
}

func TestStrapiTransformer_Transform_SimpleNode(t *testing.T) {
	t.Parallel()

	node := makeNodeWithDates("id-1", "Post", "post", "author1", "route1", "2024-01-01", "2024-06-15")
	node.Fields = append(node.Fields,
		makeField("Title", "text", "Hello Strapi"),
		makeField("Active", "boolean", "false"),
	)

	st := &StrapiTransformer{}
	result, err := st.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := result.(StrapiResponse)
	entry := resp.Data.(StrapiEntry)

	if entry.ID != "id-1" {
		t.Errorf("ID = %q, want %q", entry.ID, "id-1")
	}
	if entry.Attributes["createdAt"] != "2024-01-01" {
		t.Errorf("createdAt = %v, want %q", entry.Attributes["createdAt"], "2024-01-01")
	}
	if entry.Attributes["updatedAt"] != "2024-06-15" {
		t.Errorf("updatedAt = %v, want %q", entry.Attributes["updatedAt"], "2024-06-15")
	}
	if entry.Attributes["title"] != "Hello Strapi" {
		t.Errorf("title = %v, want %q", entry.Attributes["title"], "Hello Strapi")
	}
	if entry.Attributes["active"] != false {
		t.Errorf("active = %v, want false", entry.Attributes["active"])
	}
}

func TestStrapiTransformer_Transform_MediaField(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, model.Field{
		Info: db.FieldsJSON{
			Label: "Cover Image",
			Type:  "image",
		},
		Content: db.ContentFieldsJSON{
			FieldValue: "https://example.com/images/photo.jpg",
		},
	})

	st := &StrapiTransformer{}
	result, err := st.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := result.(StrapiResponse)
	entry := resp.Data.(StrapiEntry)

	media, ok := entry.Attributes["coverImage"].(StrapiMedia)
	if !ok {
		t.Fatalf("coverImage is not StrapiMedia, got %T", entry.Attributes["coverImage"])
	}

	if media.Data.Attributes.URL != "https://example.com/images/photo.jpg" {
		t.Errorf("URL = %q, want original URL", media.Data.Attributes.URL)
	}
	if media.Data.Attributes.Name != "photo.jpg" {
		t.Errorf("Name = %q, want %q", media.Data.Attributes.Name, "photo.jpg")
	}
	if media.Data.Attributes.AlternativeText != "Cover Image" {
		t.Errorf("AlternativeText = %q, want %q", media.Data.Attributes.AlternativeText, "Cover Image")
	}
}

func TestStrapiTransformer_Transform_ChildNodes_AsRelation(t *testing.T) {
	t.Parallel()

	parent := makeNode("parent-1", "Blog", "blog")
	child1 := makeNode("child-1", "Post", "post")
	child2 := makeNode("child-2", "Post", "post")
	parent.Nodes = []*model.Node{child1, child2}

	st := &StrapiTransformer{}
	result, err := st.Transform(model.Root{Node: parent})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp := result.(StrapiResponse)
	entry := resp.Data.(StrapiEntry)

	// Same-type children as relation
	rel, ok := entry.Attributes["posts"].(StrapiRelation)
	if !ok {
		t.Fatalf("posts is not StrapiRelation, got %T", entry.Attributes["posts"])
	}
	if len(rel.Data) != 2 {
		t.Errorf("got %d relation entries, want 2", len(rel.Data))
	}
}

func TestStrapiTransformer_TransformToJSON(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	st := &StrapiTransformer{}
	data, err := st.TransformToJSON(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if _, ok := parsed["data"]; !ok {
		t.Error("missing 'data' key in Strapi JSON response")
	}
	if _, ok := parsed["meta"]; !ok {
		t.Error("missing 'meta' key in Strapi JSON response")
	}
}

func TestStrapiTransformer_GetDateModified_Fallback(t *testing.T) {
	t.Parallel()

	st := &StrapiTransformer{}
	node := makeNodeWithDates("id", "Post", "post", "", "", "2024-01-01", "")
	got := st.getDateModified(node)
	if got != "2024-01-01" {
		t.Errorf("getDateModified = %q, want %q (fallback to created)", got, "2024-01-01")
	}
}

// ===========================================================================
// WordPressTransformer
// ===========================================================================

func TestWordPressTransformer_Transform_NilNode(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}
	result, err := wt.Transform(model.Root{Node: nil})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post, ok := result.(WordPressPost)
	if !ok {
		t.Fatalf("expected WordPressPost, got %T", result)
	}
	if post.ID != "" {
		t.Errorf("ID = %q, want empty", post.ID)
	}
}

func TestWordPressTransformer_Transform_SimpleNode(t *testing.T) {
	t.Parallel()

	node := makeNodeWithDates("id-1", "Blog Post", "blogpost", "author1", "route1", "2024-01-01", "2024-06-15")
	node.Fields = append(node.Fields,
		makeField("Title", "text", "Hello WordPress"),
		makeField("Body", "text", "Some content here"),
		makeField("Published", "boolean", "true"),
		makeField("Slug", "text", "hello-wordpress"),
	)

	wt := &WordPressTransformer{BaseTransformer: BaseTransformer{SiteURL: "https://mysite.com"}}
	result, err := wt.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post := result.(WordPressPost)

	if post.ID != "id-1" {
		t.Errorf("ID = %q, want %q", post.ID, "id-1")
	}
	if post.Title.Rendered != "Hello WordPress" {
		t.Errorf("Title.Rendered = %q, want %q", post.Title.Rendered, "Hello WordPress")
	}
	if post.Status != "publish" {
		t.Errorf("Status = %q, want %q", post.Status, "publish")
	}
	if post.Slug != "hello-wordpress" {
		t.Errorf("Slug = %q, want %q", post.Slug, "hello-wordpress")
	}
	if post.Content.Rendered != "<p>Some content here</p>" {
		t.Errorf("Content.Rendered = %q, want %q", post.Content.Rendered, "<p>Some content here</p>")
	}
	if post.Author != "author1" {
		t.Errorf("Author = %q, want %q", post.Author, "author1")
	}
	if post.Date != "2024-01-01" {
		t.Errorf("Date = %q, want %q", post.Date, "2024-01-01")
	}
	if post.Modified != "2024-06-15" {
		t.Errorf("Modified = %q, want %q", post.Modified, "2024-06-15")
	}
	if post.Type != "post" {
		t.Errorf("Type = %q, want %q (label contains 'post')", post.Type, "post")
	}

	// Link should include site URL
	if !strings.HasPrefix(post.Link, "https://mysite.com/") {
		t.Errorf("Link = %q, want prefix %q", post.Link, "https://mysite.com/")
	}
}

func TestWordPressTransformer_Transform_DraftStatus(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Published", "boolean", "false"))

	wt := &WordPressTransformer{}
	result, err := wt.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post := result.(WordPressPost)
	if post.Status != "draft" {
		t.Errorf("Status = %q, want %q", post.Status, "draft")
	}
}

func TestWordPressTransformer_GenerateSlug(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}

	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "simple title", title: "Hello World", want: "hello-world"},
		{name: "with special chars", title: "Hello, World!", want: "hello-world"},
		{name: "with numbers", title: "Top 10 Posts", want: "top-10-posts"},
		{name: "trailing space", title: "Hello ", want: "hello"},
		{name: "multiple spaces", title: "Hello   World", want: "hello-world"},
		{name: "underscores", title: "hello_world", want: "hello-world"},
		{name: "empty", title: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wt.generateSlug(tt.title)
			if got != tt.want {
				t.Errorf("generateSlug(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestWordPressTransformer_GenerateSlug_FallbackFromTitle(t *testing.T) {
	t.Parallel()

	// When slug field is empty, slug is generated from title
	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Title", "text", "My Cool Post"))

	wt := &WordPressTransformer{}
	result, err := wt.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post := result.(WordPressPost)
	if post.Slug != "my-cool-post" {
		t.Errorf("auto-generated slug = %q, want %q", post.Slug, "my-cool-post")
	}
}

func TestWordPressTransformer_GenerateExcerpt(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}

	t.Run("short text unchanged", func(t *testing.T) {
		t.Parallel()
		got := wt.generateExcerpt("Hello world", 150)
		if got != "Hello world" {
			t.Errorf("generateExcerpt = %q, want %q", got, "Hello world")
		}
	})

	t.Run("long text truncated with ellipsis", func(t *testing.T) {
		t.Parallel()
		long := strings.Repeat("a", 200)
		got := wt.generateExcerpt(long, 150)
		if len(got) != 153 { // 150 + "..."
			t.Errorf("len(generateExcerpt) = %d, want 153", len(got))
		}
		if !strings.HasSuffix(got, "...") {
			t.Errorf("excerpt should end with ..., got %q", got[len(got)-10:])
		}
	})

	t.Run("strips markdown symbols", func(t *testing.T) {
		t.Parallel()
		got := wt.generateExcerpt("## Hello **world**", 150)
		if strings.Contains(got, "#") || strings.Contains(got, "*") {
			t.Errorf("excerpt should strip markdown, got %q", got)
		}
	})
}

func TestWordPressTransformer_GenerateExcerpt_FromBody(t *testing.T) {
	t.Parallel()

	// When excerpt field is empty, it generates from body content
	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Body", "text", "This is the body content"))

	wt := &WordPressTransformer{}
	result, err := wt.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post := result.(WordPressPost)
	if post.Excerpt.Rendered == "" {
		t.Error("excerpt should be auto-generated from body, got empty")
	}
}

func TestWordPressTransformer_ToWordPressType(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}

	tests := []struct {
		name  string
		label string
		want  string
	}{
		{name: "post label", label: "Blog Post", want: "post"},
		{name: "article label", label: "News Article", want: "post"},
		{name: "page label", label: "Landing Page", want: "page"},
		{name: "custom type", label: "Product", want: "product"},
		{name: "custom with space", label: "Event Type", want: "event_type"},
		{name: "custom with hyphen", label: "my-type", want: "my_type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := wt.toWordPressType(tt.label)
			if got != tt.want {
				t.Errorf("toWordPressType(%q) = %q, want %q", tt.label, got, tt.want)
			}
		})
	}
}

func TestWordPressTransformer_GenerateLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		siteURL  string
		postType string
		slug     string
		want     string
	}{
		{
			name:     "with site URL",
			siteURL:  "https://mysite.com",
			postType: "post",
			slug:     "hello-world",
			want:     "https://mysite.com/post/hello-world",
		},
		{
			name:     "empty site URL uses default",
			siteURL:  "",
			postType: "page",
			slug:     "about",
			want:     "https://example.com/page/about",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wt := &WordPressTransformer{BaseTransformer: BaseTransformer{SiteURL: tt.siteURL}}
			got := wt.generateLink(tt.postType, tt.slug)
			if got != tt.want {
				t.Errorf("generateLink(%q, %q) = %q, want %q", tt.postType, tt.slug, got, tt.want)
			}
		})
	}
}

func TestWordPressTransformer_HashMedia(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}

	t.Run("empty URL returns 0", func(t *testing.T) {
		t.Parallel()
		got := wt.hashMedia("")
		if got != 0 {
			t.Errorf("hashMedia('') = %d, want 0", got)
		}
	})

	t.Run("non-empty URL returns positive int", func(t *testing.T) {
		t.Parallel()
		got := wt.hashMedia("https://example.com/photo.jpg")
		if got <= 0 {
			t.Errorf("hashMedia returned %d, want positive int", got)
		}
	})

	t.Run("deterministic", func(t *testing.T) {
		t.Parallel()
		url := "https://example.com/photo.jpg"
		a := wt.hashMedia(url)
		b := wt.hashMedia(url)
		if a != b {
			t.Errorf("hashMedia not deterministic: %d vs %d", a, b)
		}
	})
}

func TestWordPressTransformer_ExtractACF(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}

	fields := map[string]any{
		"title":         "Hello",
		"slug":          "hello",
		"body":          "content",
		"content":       "more content",
		"excerpt":       "short",
		"published":     true,
		"featuredImage": "url",
		"customField":   "custom value",
		"anotherField":  42,
	}

	acf := wt.extractACF(fields)

	// Standard fields should NOT be in ACF
	standardFields := []string{"title", "slug", "body", "content", "excerpt", "published", "featuredImage"}
	for _, sf := range standardFields {
		if _, exists := acf[sf]; exists {
			t.Errorf("standard field %q should not be in ACF", sf)
		}
	}

	// Custom fields should be in ACF
	if acf["customField"] != "custom value" {
		t.Errorf("customField = %v, want %q", acf["customField"], "custom value")
	}
	if acf["anotherField"] != 42 {
		t.Errorf("anotherField = %v, want 42", acf["anotherField"])
	}
}

func TestWordPressTransformer_RenderContent(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}

	t.Run("non-empty wraps in p tag", func(t *testing.T) {
		t.Parallel()
		got := wt.renderContent("Hello world")
		if got != "<p>Hello world</p>" {
			t.Errorf("renderContent = %q, want %q", got, "<p>Hello world</p>")
		}
	})

	t.Run("empty returns empty", func(t *testing.T) {
		t.Parallel()
		got := wt.renderContent("")
		if got != "" {
			t.Errorf("renderContent('') = %q, want empty", got)
		}
	})
}

func TestWordPressTransformer_RenderExcerpt(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}

	t.Run("non-empty wraps in p tag", func(t *testing.T) {
		t.Parallel()
		got := wt.renderExcerpt("Short text")
		if got != "<p>Short text</p>" {
			t.Errorf("renderExcerpt = %q, want %q", got, "<p>Short text</p>")
		}
	})

	t.Run("empty returns empty", func(t *testing.T) {
		t.Parallel()
		got := wt.renderExcerpt("")
		if got != "" {
			t.Errorf("renderExcerpt('') = %q, want empty", got)
		}
	})
}

func TestWordPressTransformer_TransformToJSON(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Title", "text", "Hello"))

	wt := &WordPressTransformer{}
	data, err := wt.TransformToJSON(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed["id"] != "id-1" {
		t.Errorf("id = %v, want %q", parsed["id"], "id-1")
	}
	if _, ok := parsed["title"]; !ok {
		t.Error("missing 'title' in WordPress JSON output")
	}
}

func TestWordPressTransformer_ContentFallback(t *testing.T) {
	t.Parallel()

	// When "body" field is empty, it should fall back to "content" field
	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Content", "text", "From content field"))

	wt := &WordPressTransformer{}
	result, err := wt.Transform(model.Root{Node: node})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	post := result.(WordPressPost)
	if post.Content.Rendered != "<p>From content field</p>" {
		t.Errorf("Content.Rendered = %q, want %q", post.Content.Rendered, "<p>From content field</p>")
	}
}

func TestWordPressTransformer_GetDateModified_Fallback(t *testing.T) {
	t.Parallel()

	wt := &WordPressTransformer{}
	node := makeNodeWithDates("id", "Post", "post", "", "", "2024-01-01", "")
	got := wt.getDateModified(node)
	if got != "2024-01-01" {
		t.Errorf("getDateModified = %q, want %q (fallback to created)", got, "2024-01-01")
	}
}

// ===========================================================================
// TransformAndWrite (HTTP integration)
// ===========================================================================

func TestTransformAndWrite(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	node.Fields = append(node.Fields, makeField("Title", "text", "Hello"))
	root := model.Root{Node: node}

	t.Run("raw format writes JSON response", func(t *testing.T) {
		t.Parallel()

		tc := &TransformConfig{Format: FormatRaw}
		w := httptest.NewRecorder()

		err := tc.TransformAndWrite(w, root)
		if err != nil {
			t.Fatalf("TransformAndWrite failed: %v", err)
		}

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}

		// Body should be valid JSON
		var parsed map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			t.Fatalf("response body is not valid JSON: %v", err)
		}
	})

	t.Run("clean format writes JSON response", func(t *testing.T) {
		t.Parallel()

		tc := &TransformConfig{Format: FormatClean}
		w := httptest.NewRecorder()

		err := tc.TransformAndWrite(w, root)
		if err != nil {
			t.Fatalf("TransformAndWrite failed: %v", err)
		}

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var parsed map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			t.Fatalf("response body is not valid JSON: %v", err)
		}

		if parsed["id"] != "id-1" {
			t.Errorf("id = %v, want %q", parsed["id"], "id-1")
		}
	})

	t.Run("contentful format writes JSON response", func(t *testing.T) {
		t.Parallel()

		tc := &TransformConfig{Format: FormatContentful}
		w := httptest.NewRecorder()

		err := tc.TransformAndWrite(w, root)
		if err != nil {
			t.Fatalf("TransformAndWrite failed: %v", err)
		}

		var parsed map[string]any
		if err := json.NewDecoder(w.Result().Body).Decode(&parsed); err != nil {
			t.Fatalf("response body is not valid JSON: %v", err)
		}

		if _, ok := parsed["sys"]; !ok {
			t.Error("Contentful response missing 'sys' object")
		}
	})

	t.Run("unknown format returns error", func(t *testing.T) {
		t.Parallel()

		tc := &TransformConfig{Format: "invalid"}
		w := httptest.NewRecorder()

		err := tc.TransformAndWrite(w, root)
		if err == nil {
			t.Fatal("expected error for unknown format, got nil")
		}
		if !strings.Contains(err.Error(), "unknown output format") {
			t.Errorf("error = %q, want 'unknown output format'", err.Error())
		}
	})
}

// ===========================================================================
// BaseTransformer.TransformToJSON
// ===========================================================================

func TestBaseTransformer_TransformToJSON(t *testing.T) {
	t.Parallel()

	node := makeNode("id-1", "Post", "post")
	root := model.Root{Node: node}

	base := &BaseTransformer{SiteURL: "https://example.com"}
	raw := &RawTransformer{}

	data, err := base.TransformToJSON(raw, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

// ===========================================================================
// Sanity: markdownToPortableText
// ===========================================================================

func TestSanityTransformer_MarkdownToPortableText(t *testing.T) {
	t.Parallel()

	st := &SanityTransformer{}

	t.Run("empty markdown", func(t *testing.T) {
		t.Parallel()
		blocks := st.markdownToPortableText("")
		if len(blocks) != 0 {
			t.Errorf("expected 0 blocks from empty markdown, got %d", len(blocks))
		}
	})

	t.Run("single paragraph", func(t *testing.T) {
		t.Parallel()
		blocks := st.markdownToPortableText("Hello world")
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		if blocks[0].Children[0].Text != "Hello world" {
			t.Errorf("text = %q, want %q", blocks[0].Children[0].Text, "Hello world")
		}
	})

	t.Run("with headers stripped", func(t *testing.T) {
		t.Parallel()
		blocks := st.markdownToPortableText("# Hello")
		if len(blocks) != 1 {
			t.Fatalf("expected 1 block, got %d", len(blocks))
		}
		// Header markers should be stripped
		text := blocks[0].Children[0].Text
		if strings.HasPrefix(text, "#") {
			t.Errorf("text should not start with #, got %q", text)
		}
	})

	t.Run("multiple paragraphs", func(t *testing.T) {
		t.Parallel()
		blocks := st.markdownToPortableText("First paragraph\n\nSecond paragraph")
		if len(blocks) != 2 {
			t.Fatalf("expected 2 blocks, got %d", len(blocks))
		}
	})
}

// ===========================================================================
// Contentful: parseFieldValue
// ===========================================================================

func TestContentfulTransformer_ParseFieldValue(t *testing.T) {
	t.Parallel()

	ct := &ContentfulTransformer{}

	tests := []struct {
		name      string
		value     string
		fieldType string
		wantNil   bool
		wantType  string // "string", "bool", "float64", "int64"
	}{
		{name: "empty", value: "", fieldType: "text", wantNil: true},
		{name: "text", value: "hello", fieldType: "text", wantType: "string"},
		{name: "boolean true", value: "true", fieldType: "boolean", wantType: "bool"},
		{name: "boolean false", value: "false", fieldType: "boolean", wantType: "bool"},
		{name: "number", value: "42", fieldType: "number", wantType: "float64"},
		{name: "integer", value: "10", fieldType: "integer", wantType: "int64"},
		{name: "json", value: `{"k":"v"}`, fieldType: "json", wantType: "map"},
		{name: "unknown type", value: "val", fieldType: "custom", wantType: "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ct.parseFieldValue(tt.value, tt.fieldType)

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %v (%T)", got, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil result")
			}

			switch tt.wantType {
			case "string":
				if _, ok := got.(string); !ok {
					t.Errorf("expected string, got %T", got)
				}
			case "bool":
				if _, ok := got.(bool); !ok {
					t.Errorf("expected bool, got %T", got)
				}
			case "float64":
				if _, ok := got.(float64); !ok {
					t.Errorf("expected float64, got %T", got)
				}
			case "int64":
				if _, ok := got.(int64); !ok {
					t.Errorf("expected int64, got %T", got)
				}
			case "map":
				if _, ok := got.(map[string]any); !ok {
					t.Errorf("expected map[string]any, got %T", got)
				}
			}
		})
	}
}

// ===========================================================================
// toString / floatToString / stringToInt64
// ===========================================================================

func TestToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{name: "string", input: "hello", want: "hello"},
		{name: "int", input: 42, want: "42"},
		{name: "int64", input: int64(100), want: "100"},
		{name: "float64", input: 3.14, want: "3"}, // toString truncates float to int
		{name: "nil", input: nil, want: ""},
		{name: "bool (unsupported)", input: true, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := toString(tt.input)
			if got != tt.want {
				t.Errorf("toString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStringToInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "simple", input: "42", want: 42},
		{name: "empty", input: "", want: 0},
		{name: "mixed", input: "12abc", want: 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stringToInt64(tt.input)
			if got != tt.want {
				t.Errorf("stringToInt64(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ===========================================================================
// pow (from contentful.go)
// ===========================================================================

func TestPow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		base int
		exp  int
		want int
	}{
		{name: "10^0", base: 10, exp: 0, want: 1},
		{name: "10^1", base: 10, exp: 1, want: 10},
		{name: "10^3", base: 10, exp: 3, want: 1000},
		{name: "2^10", base: 2, exp: 10, want: 1024},
		{name: "0^5", base: 0, exp: 5, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := pow(tt.base, tt.exp)
			if got != tt.want {
				t.Errorf("pow(%d, %d) = %d, want %d", tt.base, tt.exp, got, tt.want)
			}
		})
	}
}
