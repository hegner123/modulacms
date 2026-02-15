package types

import (
	"encoding/json"
	"testing"
)

// ============================================================
// Slug
// ============================================================

func TestSlug_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		slug    Slug
		wantErr bool
	}{
		// --- valid ---
		{name: "root path", slug: "/", wantErr: false},
		{name: "simple path", slug: "/about", wantErr: false},
		{name: "nested path", slug: "/about/careers", wantErr: false},
		{name: "with digits", slug: "/blog/2024/my-post", wantErr: false},
		{name: "single char segment", slug: "/a", wantErr: false},
		{name: "deep nesting", slug: "/a/b/c/d/e", wantErr: false},
		{name: "hyphenated", slug: "/my-great-post", wantErr: false},
		{name: "digits only segment", slug: "/123", wantErr: false},

		// --- invalid: empty ---
		{name: "empty", slug: "", wantErr: true},

		// --- invalid: must start with / ---
		{name: "no leading slash", slug: "about", wantErr: true},
		{name: "starts with letter", slug: "a/b", wantErr: true},

		// --- invalid: trailing characters ---
		{name: "trailing slash", slug: "/about/", wantErr: true},
		{name: "trailing hyphen", slug: "/about-", wantErr: true},

		// --- invalid: double chars ---
		{name: "double slash", slug: "/about//careers", wantErr: true},
		{name: "double hyphen", slug: "/about--us", wantErr: true},

		// --- invalid: hyphen/slash adjacency ---
		{name: "hyphen before slash", slug: "/about-/careers", wantErr: true},
		{name: "slash then hyphen", slug: "/about/-careers", wantErr: true},

		// --- invalid: bad characters ---
		{name: "uppercase", slug: "/About", wantErr: true},
		{name: "space", slug: "/about us", wantErr: true},
		{name: "underscore", slug: "/about_us", wantErr: true},
		{name: "special char", slug: "/about@us", wantErr: true},
		{name: "unicode", slug: "/caf\u00e9", wantErr: true},

		// --- invalid: too long ---
		{name: "256 chars", slug: Slug("/" + longString('a', 255)), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.slug.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Slug(%q).Validate() error = %v, wantErr %v", tt.slug, err, tt.wantErr)
			}
		})
	}
}

func TestSlug_StringAndIsZero(t *testing.T) {
	t.Parallel()
	s := Slug("/about")
	if s.String() != "/about" {
		t.Errorf("Slug.String() = %q, want %q", s.String(), "/about")
	}
	if s.IsZero() {
		t.Error("Slug(\"/about\").IsZero() = true, want false")
	}
	var zero Slug
	if !zero.IsZero() {
		t.Error("zero Slug.IsZero() = false, want true")
	}
}

func TestSlug_Value(t *testing.T) {
	t.Parallel()
	s := Slug("/about")
	v, err := s.Value()
	if err != nil {
		t.Fatalf("Slug.Value() error = %v", err)
	}
	if v != "/about" {
		t.Errorf("Slug.Value() = %v, want %q", v, "/about")
	}

	empty := Slug("")
	_, err = empty.Value()
	if err == nil {
		t.Error("Slug(\"\").Value() expected error")
	}
}

func TestSlug_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		wantErr bool
		want    Slug
	}{
		{name: "string", input: "/about", wantErr: false, want: "/about"},
		{name: "bytes", input: []byte("/about"), wantErr: false, want: "/about"},
		{name: "nil", input: nil, wantErr: true},
		{name: "int", input: 42, wantErr: true},
		// Scan skips validation for legacy data
		{name: "legacy invalid uppercase", input: "/About", wantErr: false, want: "/About"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Slug
			err := s.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Slug.Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && s != tt.want {
				t.Errorf("Slug.Scan(%v) = %q, want %q", tt.input, s, tt.want)
			}
		})
	}
}

func TestSlug_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	s := Slug("/blog/2024/my-post")
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Slug.MarshalJSON error = %v", err)
	}
	var got Slug
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Slug.UnmarshalJSON error = %v", err)
	}
	if got != s {
		t.Errorf("JSON round-trip: got %q, want %q", got, s)
	}
}

func TestSlug_UnmarshalJSON_Invalid(t *testing.T) {
	t.Parallel()
	// UnmarshalJSON validates, so invalid slug should fail
	var s Slug
	if err := json.Unmarshal([]byte(`"About"`), &s); err == nil {
		t.Error("Slug.UnmarshalJSON(\"About\") expected error, got nil")
	}
}

// ============================================================
// Slugify
// ============================================================

func TestSlugify(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  Slug
	}{
		{name: "simple word", input: "Home", want: "/home"},
		{name: "two words", input: "About Us", want: "/about-us"},
		{name: "with leading slash", input: "/about/careers", want: "/about/careers"},
		{name: "underscores", input: "my_great_post", want: "/my-great-post"},
		{name: "uppercase", input: "BLOG", want: "/blog"},
		{name: "mixed junk", input: "  Hello!! World?? ", want: "/-hello-world"},
		{name: "multiple hyphens collapse", input: "a---b", want: "/a-b"},
		{name: "multiple slashes collapse", input: "//a///b//", want: "/a/b"},
		{name: "hyphen before slash removed", input: "a-/b", want: "/a/b"},
		{name: "slash before hyphen removed", input: "a/-b", want: "/a/b"},
		{name: "empty input", input: "", want: "/"},
		{name: "only special chars", input: "!!@@##", want: "/"},
		{name: "digits preserved", input: "blog/2024", want: "/blog/2024"},
		{name: "trailing hyphens trimmed", input: "about-", want: "/about"},
		{name: "trailing slashes trimmed", input: "about/", want: "/about"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// Email
// ============================================================

func TestEmail_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		email   Email
		wantErr bool
	}{
		// --- valid ---
		{name: "simple", email: "user@example.com", wantErr: false},
		{name: "with dots in local", email: "first.last@example.com", wantErr: false},
		{name: "with plus", email: "user+tag@example.com", wantErr: false},
		{name: "with underscore", email: "user_name@example.com", wantErr: false},
		{name: "with percent", email: "user%name@example.com", wantErr: false},
		{name: "with hyphen in domain", email: "user@my-domain.com", wantErr: false},
		{name: "subdomain", email: "user@mail.example.co.uk", wantErr: false},
		{name: "numeric domain", email: "user@123.456.com", wantErr: false},
		{name: "uppercase local", email: "USER@example.com", wantErr: false},
		{name: "uppercase domain", email: "user@EXAMPLE.COM", wantErr: false},

		// --- invalid ---
		{name: "empty", email: "", wantErr: true},
		{name: "no at", email: "userexample.com", wantErr: true},
		{name: "multiple at", email: "user@@example.com", wantErr: true},
		{name: "no local", email: "@example.com", wantErr: true},
		{name: "no domain", email: "user@", wantErr: true},
		{name: "no dot in domain", email: "user@localhost", wantErr: true},
		{name: "short tld", email: "user@example.c", wantErr: true},
		{name: "space in local", email: "us er@example.com", wantErr: true},
		{name: "space in domain", email: "user@exam ple.com", wantErr: true},
		{name: "too long", email: Email(longString('a', 65) + "@" + longString('b', 185) + ".com"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.email.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Email(%q).Validate() error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestEmail_Domain(t *testing.T) {
	t.Parallel()
	e := Email("user@example.com")
	if d := e.Domain(); d != "example.com" {
		t.Errorf("Email.Domain() = %q, want %q", d, "example.com")
	}

	bad := Email("no-at-sign")
	if d := bad.Domain(); d != "" {
		t.Errorf("Email(no @).Domain() = %q, want empty", d)
	}
}

func TestEmail_StringAndIsZero(t *testing.T) {
	t.Parallel()
	e := Email("user@example.com")
	if e.String() != "user@example.com" {
		t.Errorf("Email.String() = %q", e.String())
	}
	if e.IsZero() {
		t.Error("Email.IsZero() = true, want false")
	}
	var zero Email
	if !zero.IsZero() {
		t.Error("zero Email.IsZero() = false, want true")
	}
}

func TestEmail_Value(t *testing.T) {
	t.Parallel()
	e := Email("user@example.com")
	v, err := e.Value()
	if err != nil {
		t.Fatalf("Email.Value() error = %v", err)
	}
	if v != "user@example.com" {
		t.Errorf("Email.Value() = %v", v)
	}

	empty := Email("")
	_, err = empty.Value()
	if err == nil {
		t.Error("Email(\"\").Value() expected error")
	}
}

func TestEmail_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		wantErr bool
		want    Email
	}{
		{name: "string", input: "user@example.com", wantErr: false, want: "user@example.com"},
		{name: "bytes", input: []byte("user@example.com"), wantErr: false, want: "user@example.com"},
		{name: "nil", input: nil, wantErr: true},
		{name: "int", input: 42, wantErr: true},
		// Scan validates, so invalid email should fail
		{name: "invalid email", input: "not-an-email", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var e Email
			err := e.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Email.Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && e != tt.want {
				t.Errorf("Email.Scan(%v) = %q, want %q", tt.input, e, tt.want)
			}
		})
	}
}

func TestEmail_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	e := Email("user@example.com")
	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("Email.MarshalJSON error = %v", err)
	}
	var got Email
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Email.UnmarshalJSON error = %v", err)
	}
	if got != e {
		t.Errorf("JSON round-trip: got %q, want %q", got, e)
	}
}

// ============================================================
// URL
// ============================================================

func TestURL_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		url     URL
		wantErr bool
	}{
		{name: "https", url: "https://example.com", wantErr: false},
		{name: "http", url: "http://example.com", wantErr: false},
		{name: "with path", url: "https://example.com/path", wantErr: false},
		{name: "with query", url: "https://example.com?q=1", wantErr: false},
		{name: "with port", url: "https://example.com:8080", wantErr: false},
		{name: "ftp scheme", url: "ftp://files.example.com", wantErr: false},

		{name: "empty", url: "", wantErr: true},
		{name: "no scheme", url: "example.com", wantErr: true},
		{name: "no host", url: "https://", wantErr: true},
		{name: "relative path", url: "/about", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.url.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("URL(%q).Validate() error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestURL_Parse(t *testing.T) {
	t.Parallel()
	u := URL("https://example.com/path?q=1")
	parsed, err := u.Parse()
	if err != nil {
		t.Fatalf("URL.Parse() error = %v", err)
	}
	if parsed.Host != "example.com" {
		t.Errorf("URL.Parse().Host = %q, want %q", parsed.Host, "example.com")
	}
	if parsed.Path != "/path" {
		t.Errorf("URL.Parse().Path = %q, want %q", parsed.Path, "/path")
	}
}

func TestURL_StringAndIsZero(t *testing.T) {
	t.Parallel()
	u := URL("https://example.com")
	if u.String() != "https://example.com" {
		t.Errorf("URL.String() = %q", u.String())
	}
	if u.IsZero() {
		t.Error("URL.IsZero() = true, want false")
	}
	var zero URL
	if !zero.IsZero() {
		t.Error("zero URL.IsZero() = false, want true")
	}
}

func TestURL_Value(t *testing.T) {
	t.Parallel()
	u := URL("https://example.com")
	v, err := u.Value()
	if err != nil {
		t.Fatalf("URL.Value() error = %v", err)
	}
	if v != "https://example.com" {
		t.Errorf("URL.Value() = %v", v)
	}

	empty := URL("")
	_, err = empty.Value()
	if err == nil {
		t.Error("URL(\"\").Value() expected error")
	}
}

func TestURL_Scan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   any
		wantErr bool
		want    URL
	}{
		{name: "string", input: "https://example.com", wantErr: false, want: "https://example.com"},
		{name: "bytes", input: []byte("https://example.com"), wantErr: false, want: "https://example.com"},
		{name: "nil", input: nil, wantErr: true},
		{name: "int", input: 42, wantErr: true},
		{name: "no scheme", input: "example.com", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var u URL
			err := u.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("URL.Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && u != tt.want {
				t.Errorf("URL.Scan(%v) = %q, want %q", tt.input, u, tt.want)
			}
		})
	}
}

func TestURL_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	u := URL("https://example.com/api")
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("URL.MarshalJSON error = %v", err)
	}
	var got URL
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("URL.UnmarshalJSON error = %v", err)
	}
	if got != u {
		t.Errorf("JSON round-trip: got %q, want %q", got, u)
	}
}

// ============================================================
// NullableSlug
// ============================================================

func TestNullableSlug_NullLifecycle(t *testing.T) {
	t.Parallel()
	var ns NullableSlug

	// zero value is null
	if ns.Valid {
		t.Error("zero NullableSlug.Valid = true")
	}
	if !ns.IsZero() {
		t.Error("zero NullableSlug.IsZero() = false")
	}
	if ns.String() != "null" {
		t.Errorf("null NullableSlug.String() = %q, want %q", ns.String(), "null")
	}

	// Validate null is ok
	if err := ns.Validate(); err != nil {
		t.Errorf("null NullableSlug.Validate() = %v", err)
	}

	// Value returns nil
	v, err := ns.Value()
	if err != nil {
		t.Errorf("null NullableSlug.Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null NullableSlug.Value() = %v, want nil", v)
	}
}

func TestNullableSlug_ValidLifecycle(t *testing.T) {
	t.Parallel()
	ns := NullableSlug{Slug: "/about", Valid: true}

	if ns.IsZero() {
		t.Error("valid NullableSlug.IsZero() = true")
	}
	if ns.String() != "/about" {
		t.Errorf("NullableSlug.String() = %q", ns.String())
	}
	if err := ns.Validate(); err != nil {
		t.Errorf("NullableSlug.Validate() = %v", err)
	}

	v, err := ns.Value()
	if err != nil {
		t.Errorf("NullableSlug.Value() error = %v", err)
	}
	if v != "/about" {
		t.Errorf("NullableSlug.Value() = %v", v)
	}
}

func TestNullableSlug_Scan(t *testing.T) {
	t.Parallel()
	// Scan nil -> null
	var ns NullableSlug
	if err := ns.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if ns.Valid {
		t.Error("Scan(nil) -> Valid = true")
	}

	// Scan string -> valid
	if err := ns.Scan("/about"); err != nil {
		t.Fatalf("Scan(\"/about\") error = %v", err)
	}
	if !ns.Valid || ns.Slug != "/about" {
		t.Errorf("Scan(\"/about\") -> Valid=%v Slug=%q", ns.Valid, ns.Slug)
	}
}

func TestNullableSlug_JSON_Null(t *testing.T) {
	t.Parallel()
	ns := NullableSlug{Valid: false}
	data, err := ns.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	if string(data) != "null" {
		t.Errorf("MarshalJSON = %s, want null", data)
	}

	var got NullableSlug
	if err := got.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if got.Valid {
		t.Error("UnmarshalJSON(null) -> Valid = true")
	}
}

func TestNullableSlug_JSON_Valid(t *testing.T) {
	t.Parallel()
	ns := NullableSlug{Slug: "/about", Valid: true}
	data, err := json.Marshal(ns)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}

	var got NullableSlug
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid || got.Slug != "/about" {
		t.Errorf("JSON round-trip: got Valid=%v Slug=%q", got.Valid, got.Slug)
	}
}

// ============================================================
// NullableEmail
// ============================================================

func TestNullableEmail_NullLifecycle(t *testing.T) {
	t.Parallel()
	var ne NullableEmail
	if ne.Valid {
		t.Error("zero NullableEmail.Valid = true")
	}
	if !ne.IsZero() {
		t.Error("zero NullableEmail.IsZero() = false")
	}
	if ne.String() != "null" {
		t.Errorf("null NullableEmail.String() = %q", ne.String())
	}
	if err := ne.Validate(); err != nil {
		t.Errorf("null NullableEmail.Validate() = %v", err)
	}
	if d := ne.Domain(); d != "" {
		t.Errorf("null NullableEmail.Domain() = %q", d)
	}
	v, err := ne.Value()
	if err != nil {
		t.Errorf("null NullableEmail.Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null NullableEmail.Value() = %v", v)
	}
}

func TestNullableEmail_ValidLifecycle(t *testing.T) {
	t.Parallel()
	ne := NullableEmail{Email: "user@example.com", Valid: true}
	if ne.IsZero() {
		t.Error("valid NullableEmail.IsZero() = true")
	}
	if ne.Domain() != "example.com" {
		t.Errorf("NullableEmail.Domain() = %q", ne.Domain())
	}
	if err := ne.Validate(); err != nil {
		t.Errorf("NullableEmail.Validate() = %v", err)
	}
}

func TestNullableEmail_Scan(t *testing.T) {
	t.Parallel()
	var ne NullableEmail
	if err := ne.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if ne.Valid {
		t.Error("Scan(nil) -> Valid = true")
	}

	if err := ne.Scan("user@example.com"); err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	if !ne.Valid || ne.Email != "user@example.com" {
		t.Errorf("after Scan: Valid=%v Email=%q", ne.Valid, ne.Email)
	}
}

func TestNullableEmail_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	ne := NullableEmail{Email: "user@example.com", Valid: true}
	data, err := json.Marshal(ne)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	var got NullableEmail
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid || got.Email != "user@example.com" {
		t.Errorf("JSON round-trip: Valid=%v Email=%q", got.Valid, got.Email)
	}

	// null round-trip
	nullData, _ := json.Marshal(NullableEmail{Valid: false})
	var gotNull NullableEmail
	if err := json.Unmarshal(nullData, &gotNull); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if gotNull.Valid {
		t.Error("null JSON round-trip: Valid = true")
	}
}

// ============================================================
// NullableURL
// ============================================================

func TestNullableURL_NullLifecycle(t *testing.T) {
	t.Parallel()
	var nu NullableURL
	if nu.Valid {
		t.Error("zero NullableURL.Valid = true")
	}
	if !nu.IsZero() {
		t.Error("zero NullableURL.IsZero() = false")
	}
	if nu.String() != "null" {
		t.Errorf("null NullableURL.String() = %q", nu.String())
	}
	if err := nu.Validate(); err != nil {
		t.Errorf("null NullableURL.Validate() = %v", err)
	}

	parsed, err := nu.Parse()
	if err != nil {
		t.Errorf("null NullableURL.Parse() error = %v", err)
	}
	if parsed != nil {
		t.Errorf("null NullableURL.Parse() = %v, want nil", parsed)
	}

	v, err := nu.Value()
	if err != nil {
		t.Errorf("null NullableURL.Value() error = %v", err)
	}
	if v != nil {
		t.Errorf("null NullableURL.Value() = %v", v)
	}
}

func TestNullableURL_ValidLifecycle(t *testing.T) {
	t.Parallel()
	nu := NullableURL{URL: "https://example.com", Valid: true}
	if nu.IsZero() {
		t.Error("valid NullableURL.IsZero() = true")
	}
	if err := nu.Validate(); err != nil {
		t.Errorf("NullableURL.Validate() = %v", err)
	}
	parsed, err := nu.Parse()
	if err != nil {
		t.Fatalf("NullableURL.Parse() error = %v", err)
	}
	if parsed.Host != "example.com" {
		t.Errorf("NullableURL.Parse().Host = %q", parsed.Host)
	}
}

func TestNullableURL_Scan(t *testing.T) {
	t.Parallel()
	var nu NullableURL
	if err := nu.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) error = %v", err)
	}
	if nu.Valid {
		t.Error("Scan(nil) -> Valid = true")
	}

	if err := nu.Scan("https://example.com"); err != nil {
		t.Fatalf("Scan error = %v", err)
	}
	if !nu.Valid || nu.URL != "https://example.com" {
		t.Errorf("after Scan: Valid=%v URL=%q", nu.Valid, nu.URL)
	}
}

func TestNullableURL_JSON_RoundTrip(t *testing.T) {
	t.Parallel()
	nu := NullableURL{URL: "https://example.com/api", Valid: true}
	data, err := json.Marshal(nu)
	if err != nil {
		t.Fatalf("MarshalJSON error = %v", err)
	}
	var got NullableURL
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("UnmarshalJSON error = %v", err)
	}
	if !got.Valid || got.URL != "https://example.com/api" {
		t.Errorf("JSON round-trip: Valid=%v URL=%q", got.Valid, got.URL)
	}

	// null round-trip
	nullData, _ := json.Marshal(NullableURL{Valid: false})
	var gotNull NullableURL
	if err := json.Unmarshal(nullData, &gotNull); err != nil {
		t.Fatalf("UnmarshalJSON(null) error = %v", err)
	}
	if gotNull.Valid {
		t.Error("null JSON round-trip: Valid = true")
	}
}

// ============================================================
// helpers
// ============================================================

// longString generates a string of length n by repeating character c.
func longString(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}
