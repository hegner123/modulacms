package transform

import (
	"strings"
	"testing"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

// ---------------------------------------------------------------------------
// Test helpers: build model objects concisely
// ---------------------------------------------------------------------------

func makeNode(id, label, typ string) *model.Node {
	return &model.Node{
		Datatype: model.Datatype{
			Info: db.DatatypeJSON{
				Name:  typ,
				Label: label,
				Type:  typ,
			},
			Content: db.ContentDataJSON{
				ContentDataID: id,
			},
		},
		Fields: []model.Field{},
		Nodes:  []*model.Node{},
	}
}

func makeNodeWithDates(id, label, typ, authorID, routeID, created, modified string) *model.Node {
	return &model.Node{
		Datatype: model.Datatype{
			Info: db.DatatypeJSON{
				Name:  typ,
				Label: label,
				Type:  typ,
			},
			Content: db.ContentDataJSON{
				ContentDataID: id,
				AuthorID:      authorID,
				RouteID:       routeID,
				DateCreated:   created,
				DateModified:  modified,
			},
		},
		Fields: []model.Field{},
		Nodes:  []*model.Node{},
	}
}

func makeField(label, typ, value string) model.Field {
	return model.Field{
		Info: db.FieldsJSON{
			Label: label,
			Type:  typ,
		},
		Content: db.ContentFieldsJSON{
			FieldValue: value,
		},
	}
}

// ---------------------------------------------------------------------------
// splitWords
// ---------------------------------------------------------------------------

func TestSplitWords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty string", input: "", want: []string{}},
		{name: "single word", input: "Title", want: []string{"Title"}},
		{name: "two words space", input: "Featured Image", want: []string{"Featured", "Image"}},
		{name: "three words space", input: "SEO Meta Description", want: []string{"SEO", "Meta", "Description"}},
		{name: "underscore separator", input: "featured_image", want: []string{"featured", "image"}},
		{name: "hyphen separator", input: "featured-image", want: []string{"featured", "image"}},
		{name: "mixed separators", input: "my-cool_field name", want: []string{"my", "cool", "field", "name"}},
		{name: "leading separator", input: " hello", want: []string{"hello"}},
		{name: "trailing separator", input: "hello ", want: []string{"hello"}},
		{name: "multiple consecutive separators", input: "hello   world", want: []string{"hello", "world"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitWords(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitWords(%q): got %d words %v, want %d words %v", tt.input, len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitWords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// toLowerCase
// ---------------------------------------------------------------------------

func TestToLowerCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "already lowercase", input: "hello", want: "hello"},
		{name: "all uppercase", input: "HELLO", want: "hello"},
		{name: "mixed case", input: "HeLLo", want: "hello"},
		{name: "with numbers", input: "Hello123", want: "hello123"},
		{name: "with spaces", input: "Hello World", want: "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := toLowerCase(tt.input)
			if got != tt.want {
				t.Errorf("toLowerCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// toTitleCase
// ---------------------------------------------------------------------------

func TestToTitleCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "lowercase word", input: "hello", want: "Hello"},
		{name: "already titlecase", input: "Hello", want: "Hello"},
		{name: "all uppercase", input: "HELLO", want: "Hello"},
		{name: "single char lowercase", input: "h", want: "H"},
		{name: "single char uppercase", input: "H", want: "H"},
		{name: "mixed case word", input: "hELLO", want: "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := toTitleCase(tt.input)
			if got != tt.want {
				t.Errorf("toTitleCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// fieldLabelToKey
// ---------------------------------------------------------------------------

func TestFieldLabelToKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "single word titlecase", input: "Title", want: "title"},
		{name: "two words", input: "Featured Image", want: "featuredImage"},
		{name: "three words", input: "SEO Meta Description", want: "seoMetaDescription"},
		{name: "already camelCase lowercase", input: "title", want: "title"},
		{name: "all uppercase words", input: "ALL CAPS FIELD", want: "allCapsField"},
		{name: "underscore separated", input: "my_field_name", want: "myFieldName"},
		{name: "hyphen separated", input: "my-field-name", want: "myFieldName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fieldLabelToKey(tt.input)
			if got != tt.want {
				t.Errorf("fieldLabelToKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// pluralize
// ---------------------------------------------------------------------------

func TestPluralize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "regular word", input: "Post", want: "posts"},
		{name: "already plural (ends in s)", input: "Posts", want: "posts"},
		{name: "ends in y", input: "Category", want: "categories"},
		{name: "single letter", input: "x", want: "xs"},
		{name: "word ending in s lowercase", input: "items", want: "items"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := pluralize(tt.input)
			if got != tt.want {
				t.Errorf("pluralize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// allChildrenSameType
// ---------------------------------------------------------------------------

func TestAllChildrenSameType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		nodes []*model.Node
		want  bool
	}{
		{name: "nil slice", nodes: nil, want: true},
		{name: "empty slice", nodes: []*model.Node{}, want: true},
		{name: "single node", nodes: []*model.Node{makeNode("1", "Post", "post")}, want: true},
		{
			name: "two same type",
			nodes: []*model.Node{
				makeNode("1", "Post", "post"),
				makeNode("2", "Post", "post"),
			},
			want: true,
		},
		{
			name: "two different types",
			nodes: []*model.Node{
				makeNode("1", "Post", "post"),
				makeNode("2", "Page", "page"),
			},
			want: false,
		},
		{
			name: "three nodes mixed",
			nodes: []*model.Node{
				makeNode("1", "Post", "post"),
				makeNode("2", "Post", "post"),
				makeNode("3", "Page", "page"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := allChildrenSameType(tt.nodes)
			if got != tt.want {
				t.Errorf("allChildrenSameType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// containsIgnoreCase (from sanity.go)
// ---------------------------------------------------------------------------

func TestContainsIgnoreCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		str    string
		substr string
		want   bool
	}{
		{name: "exact match", str: "hello", substr: "hello", want: true},
		{name: "case mismatch", str: "Hello", substr: "hello", want: true},
		{name: "substring present", str: "Blog Post", substr: "post", want: true},
		{name: "substring absent", str: "Blog Post", substr: "page", want: false},
		{name: "empty substr", str: "anything", substr: "", want: true},
		{name: "substr longer than str", str: "hi", substr: "hello", want: false},
		{name: "both empty", str: "", substr: "", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := containsIgnoreCase(tt.str, tt.substr)
			if got != tt.want {
				t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.str, tt.substr, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// splitByNewlines (from sanity.go)
// ---------------------------------------------------------------------------

func TestSplitByNewlines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty", input: "", want: []string{}},
		{name: "single line", input: "hello world", want: []string{"hello world"}},
		{name: "two paragraphs", input: "first\n\nsecond", want: []string{"first", "second"}},
		{name: "single newline merges", input: "line one\nline two", want: []string{"line oneline two"}},
		{name: "multiple blank lines", input: "a\n\n\n\nb", want: []string{"a", "b"}},
		{name: "trailing newlines", input: "hello\n\n", want: []string{"hello"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitByNewlines(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitByNewlines(%q): got %d items %v, want %d items %v", tt.input, len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitByNewlines(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// trimString (from wordpress.go)
// ---------------------------------------------------------------------------

func TestTrimString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "no whitespace", input: "hello", want: "hello"},
		{name: "leading spaces", input: "   hello", want: "hello"},
		{name: "trailing spaces", input: "hello   ", want: "hello"},
		{name: "both sides", input: "  hello  ", want: "hello"},
		{name: "tabs and newlines", input: "\t\nhello\r\n", want: "hello"},
		{name: "all whitespace", input: "   \t\n  ", want: ""},
		{name: "internal spaces preserved", input: "  hello world  ", want: "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := trimString(tt.input)
			if got != tt.want {
				t.Errorf("trimString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractFilename (from strapi.go)
// ---------------------------------------------------------------------------

func TestExtractFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "full URL", input: "https://example.com/images/photo.jpg", want: "photo.jpg"},
		{name: "no path", input: "photo.jpg", want: "file"},
		{name: "trailing slash", input: "https://example.com/", want: "file"},
		{name: "single segment path", input: "/photo.jpg", want: "photo.jpg"},
		{name: "empty string", input: "", want: "file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractFilename(tt.input)
			if got != tt.want {
				t.Errorf("extractFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// int64ToString (from contentful.go)
// ---------------------------------------------------------------------------

func TestInt64ToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int64
		want  string
	}{
		{name: "zero", input: 0, want: "0"},
		{name: "positive", input: 42, want: "42"},
		{name: "large positive", input: 1234567890, want: "1234567890"},
		{name: "negative", input: -42, want: "-42"},
		{name: "one", input: 1, want: "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := int64ToString(tt.input)
			if got != tt.want {
				t.Errorf("int64ToString(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseFloat: documents a known bug where multi-digit decimals panic
// ---------------------------------------------------------------------------

func TestParseFloat_WholeNumbers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{name: "zero", input: "0", want: 0},
		{name: "simple integer string", input: "42", want: 42},
		{name: "large number", input: "1234", want: 1234},
		{name: "empty string", input: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseFloat(tt.input)
			if got != tt.want {
				t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseFloat_Bug documents a bug in the parseFloat implementation.
// The function uses the accumulated decimal digit value as a string index
// (line 234: `len(s[len(s)-decimal:])`), which panics when the accumulated
// value exceeds the string length. For example, "3.14" causes decimal=14,
// then `s[len(s)-14:]` panics with "slice bounds out of range".
// FIX: Track decimal digit count separately from the accumulated value.
func TestParseFloat_Bug(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected parseFloat(\"3.14\") to panic due to known bug, but it did not")
		}
	}()

	parseFloat("3.14")
}

// ---------------------------------------------------------------------------
// parseInt (from contentful.go)
// ---------------------------------------------------------------------------

func TestParseInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "simple number", input: "42", want: 42},
		{name: "zero", input: "0", want: 0},
		{name: "large number", input: "1234567890", want: 1234567890},
		{name: "empty string", input: "", want: 0},
		{name: "non-numeric", input: "abc", want: 0},
		{name: "mixed numeric and alpha", input: "12abc34", want: 1234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseInt(tt.input)
			if got != tt.want {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// intToBase36 (from sanity.go)
// ---------------------------------------------------------------------------

func TestIntToBase36(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input int
		want  string
	}{
		{name: "zero", input: 0, want: "0"},
		{name: "small number", input: 10, want: "a"},
		{name: "36", input: 36, want: "10"},
		{name: "35", input: 35, want: "z"},
		{name: "one", input: 1, want: "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := intToBase36(tt.input)
			if got != tt.want {
				t.Errorf("intToBase36(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// hashString - determinism check
// ---------------------------------------------------------------------------

func TestHashString_Deterministic(t *testing.T) {
	t.Parallel()

	// Same input should always produce same output
	input := "https://example.com/images/photo.jpg"
	first := hashString(input)
	second := hashString(input)

	if first != second {
		t.Errorf("hashString is not deterministic: %q vs %q for input %q", first, second, input)
	}

	if first == "" {
		t.Error("hashString returned empty string")
	}
}

// ---------------------------------------------------------------------------
// hashStringInt - determinism and non-negative check
// ---------------------------------------------------------------------------

func TestHashStringInt_Deterministic(t *testing.T) {
	t.Parallel()

	input := "https://example.com/images/photo.jpg"
	first := hashStringInt(input)
	second := hashStringInt(input)

	if first != second {
		t.Errorf("hashStringInt is not deterministic: %d vs %d", first, second)
	}

	if first < 0 {
		t.Errorf("hashStringInt returned negative value: %d", first)
	}
}

func TestHashStringInt_DifferentInputs(t *testing.T) {
	t.Parallel()

	a := hashStringInt("hello")
	b := hashStringInt("world")

	if a == b {
		t.Errorf("hashStringInt produced same hash for different inputs: %d", a)
	}
}

// ---------------------------------------------------------------------------
// GetTransformer factory
// ---------------------------------------------------------------------------

func TestGetTransformer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		format   OutputFormat
		wantType string // type name to check via sprint
		wantErr  string
	}{
		{name: "contentful", format: FormatContentful, wantType: "*transform.ContentfulTransformer"},
		{name: "sanity", format: FormatSanity, wantType: "*transform.SanityTransformer"},
		{name: "strapi", format: FormatStrapi, wantType: "*transform.StrapiTransformer"},
		{name: "wordpress", format: FormatWordPress, wantType: "*transform.WordPressTransformer"},
		{name: "clean", format: FormatClean, wantType: "*transform.CleanTransformer"},
		{name: "raw", format: FormatRaw, wantType: "*transform.RawTransformer"},
		{name: "empty string defaults to raw", format: "", wantType: "*transform.RawTransformer"},
		{name: "unknown format", format: "nonexistent", wantErr: "unknown output format: nonexistent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tc := &TransformConfig{
				Format:  tt.format,
				SiteURL: "https://example.com",
				SpaceID: "space123",
			}

			transformer, err := tc.GetTransformer()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			gotType := typeNameOf(transformer)
			if gotType != tt.wantType {
				t.Errorf("GetTransformer() returned %s, want %s", gotType, tt.wantType)
			}
		})
	}
}

// typeNameOf returns the fully qualified type name using fmt.Sprintf
func typeNameOf(v any) string {
	return strings.TrimPrefix(strings.Replace(
		strings.Replace(
			func() string {
				return typeNameReflectFree(v)
			}(), " ", "", -1),
		"\n", "", -1),
		"")
}

// typeNameReflectFree produces a type assertion based type name
func typeNameReflectFree(v any) string {
	switch v.(type) {
	case *ContentfulTransformer:
		return "*transform.ContentfulTransformer"
	case *SanityTransformer:
		return "*transform.SanityTransformer"
	case *StrapiTransformer:
		return "*transform.StrapiTransformer"
	case *WordPressTransformer:
		return "*transform.WordPressTransformer"
	case *CleanTransformer:
		return "*transform.CleanTransformer"
	case *RawTransformer:
		return "*transform.RawTransformer"
	default:
		return "unknown"
	}
}

// ---------------------------------------------------------------------------
// NewTransformConfig / NewTransformConfigFromString
// ---------------------------------------------------------------------------

func TestNewTransformConfig(t *testing.T) {
	t.Parallel()

	tc := NewTransformConfig(FormatClean, "https://example.com", "space1", nil)

	if tc.Format != FormatClean {
		t.Errorf("Format = %q, want %q", tc.Format, FormatClean)
	}
	if tc.SiteURL != "https://example.com" {
		t.Errorf("SiteURL = %q, want %q", tc.SiteURL, "https://example.com")
	}
	if tc.SpaceID != "space1" {
		t.Errorf("SpaceID = %q, want %q", tc.SpaceID, "space1")
	}
}

func TestNewTransformConfigFromString(t *testing.T) {
	t.Parallel()

	tc := NewTransformConfigFromString("contentful", "https://example.com", "space1", nil)

	if tc.Format != FormatContentful {
		t.Errorf("Format = %q, want %q", tc.Format, FormatContentful)
	}
}

// ---------------------------------------------------------------------------
// GetTransformer propagates SiteURL and SpaceID to BaseTransformer
// ---------------------------------------------------------------------------

func TestGetTransformer_PropagatesConfig(t *testing.T) {
	t.Parallel()

	tc := &TransformConfig{
		Format:  FormatContentful,
		SiteURL: "https://mysite.com",
		SpaceID: "myspace",
	}

	transformer, err := tc.GetTransformer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ct, ok := transformer.(*ContentfulTransformer)
	if !ok {
		t.Fatalf("expected *ContentfulTransformer, got %T", transformer)
	}

	if ct.SiteURL != "https://mysite.com" {
		t.Errorf("SiteURL = %q, want %q", ct.SiteURL, "https://mysite.com")
	}
	if ct.SpaceID != "myspace" {
		t.Errorf("SpaceID = %q, want %q", ct.SpaceID, "myspace")
	}
}
