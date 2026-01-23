package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

// Slug represents a URL-safe identifier (lowercase alphanumeric with hyphens)
type Slug string

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func (s Slug) Validate() error {
	if s == "" {
		return fmt.Errorf("Slug: cannot be empty")
	}
	if len(s) > 255 {
		return fmt.Errorf("Slug: too long (max 255 chars)")
	}
	if !slugRegex.MatchString(string(s)) {
		return fmt.Errorf("Slug: invalid format %q (must be lowercase alphanumeric with hyphens)", s)
	}
	return nil
}

func (s Slug) String() string { return string(s) }

func (s Slug) IsZero() bool { return s == "" }

// Slugify converts a string to a valid slug
func Slugify(input string) Slug {
	// Lowercase
	result := strings.ToLower(input)
	// Replace spaces and underscores with hyphens
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "_", "-")
	// Remove non-alphanumeric except hyphens
	var sb strings.Builder
	for _, r := range result {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			sb.WriteRune(r)
		}
	}
	result = sb.String()
	// Collapse multiple hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	// Trim hyphens from ends
	result = strings.Trim(result, "-")
	return Slug(result)
}

func (s Slug) Value() (driver.Value, error) {
	if s == "" {
		return nil, fmt.Errorf("Slug: cannot be empty")
	}
	return string(s), nil
}

func (s *Slug) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("Slug: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*s = Slug(v)
	case []byte:
		*s = Slug(string(v))
	default:
		return fmt.Errorf("Slug: cannot scan %T", value)
	}
	return s.Validate()
}

func (s Slug) MarshalJSON() ([]byte, error) { return json.Marshal(string(s)) }

func (s *Slug) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Slug: %w", err)
	}
	*s = Slug(str)
	return s.Validate()
}

// Email represents a validated email address
type Email string

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func (e Email) Validate() error {
	if e == "" {
		return fmt.Errorf("Email: cannot be empty")
	}
	if len(e) > 254 {
		return fmt.Errorf("Email: too long (max 254 chars)")
	}
	if !emailRegex.MatchString(string(e)) {
		return fmt.Errorf("Email: invalid format %q", e)
	}
	return nil
}

func (e Email) String() string { return string(e) }

func (e Email) IsZero() bool { return e == "" }

// Domain returns the domain part of the email address
func (e Email) Domain() string {
	parts := strings.Split(string(e), "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func (e Email) Value() (driver.Value, error) {
	if e == "" {
		return nil, fmt.Errorf("Email: cannot be empty")
	}
	return string(e), nil
}

func (e *Email) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("Email: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*e = Email(v)
	case []byte:
		*e = Email(string(v))
	default:
		return fmt.Errorf("Email: cannot scan %T", value)
	}
	return e.Validate()
}

func (e Email) MarshalJSON() ([]byte, error) { return json.Marshal(string(e)) }

func (e *Email) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Email: %w", err)
	}
	*e = Email(str)
	return e.Validate()
}

// URL represents a validated URL (must have scheme and host)
type URL string

func (u URL) Validate() error {
	if u == "" {
		return fmt.Errorf("URL: cannot be empty")
	}
	parsed, err := url.Parse(string(u))
	if err != nil {
		return fmt.Errorf("URL: invalid format %q: %w", u, err)
	}
	if parsed.Scheme == "" {
		return fmt.Errorf("URL: missing scheme in %q", u)
	}
	if parsed.Host == "" {
		return fmt.Errorf("URL: missing host in %q", u)
	}
	return nil
}

func (u URL) String() string { return string(u) }

func (u URL) IsZero() bool { return u == "" }

// Parse returns the URL as a parsed *url.URL
func (u URL) Parse() (*url.URL, error) {
	return url.Parse(string(u))
}

func (u URL) Value() (driver.Value, error) {
	if u == "" {
		return nil, fmt.Errorf("URL: cannot be empty")
	}
	return string(u), nil
}

func (u *URL) Scan(value any) error {
	if value == nil {
		return fmt.Errorf("URL: cannot be null")
	}
	switch v := value.(type) {
	case string:
		*u = URL(v)
	case []byte:
		*u = URL(string(v))
	default:
		return fmt.Errorf("URL: cannot scan %T", value)
	}
	return u.Validate()
}

func (u URL) MarshalJSON() ([]byte, error) { return json.Marshal(string(u)) }

func (u *URL) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("URL: %w", err)
	}
	*u = URL(str)
	return u.Validate()
}

// NullableSlug represents a nullable slug for optional fields
type NullableSlug struct {
	Slug  Slug
	Valid bool
}

func (n NullableSlug) Validate() error {
	if n.Valid {
		return n.Slug.Validate()
	}
	return nil
}

func (n NullableSlug) String() string {
	if !n.Valid {
		return "null"
	}
	return n.Slug.String()
}

func (n NullableSlug) IsZero() bool { return !n.Valid || n.Slug == "" }

func (n NullableSlug) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.Slug), nil
}

func (n *NullableSlug) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.Slug = ""
		return nil
	}
	n.Valid = true
	return n.Slug.Scan(value)
}

func (n NullableSlug) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Slug)
}

func (n *NullableSlug) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Slug = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Slug)
}

// NullableEmail represents a nullable email for optional fields
type NullableEmail struct {
	Email Email
	Valid bool
}

func (n NullableEmail) Validate() error {
	if n.Valid {
		return n.Email.Validate()
	}
	return nil
}

func (n NullableEmail) String() string {
	if !n.Valid {
		return "null"
	}
	return n.Email.String()
}

func (n NullableEmail) IsZero() bool { return !n.Valid || n.Email == "" }

// Domain returns the domain part of the email address, or empty string if null
func (n NullableEmail) Domain() string {
	if !n.Valid {
		return ""
	}
	return n.Email.Domain()
}

func (n NullableEmail) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.Email), nil
}

func (n *NullableEmail) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.Email = ""
		return nil
	}
	n.Valid = true
	return n.Email.Scan(value)
}

func (n NullableEmail) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Email)
}

func (n *NullableEmail) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.Email = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Email)
}

// NullableURL represents a nullable URL for optional fields
type NullableURL struct {
	URL   URL
	Valid bool
}

func (n NullableURL) Validate() error {
	if n.Valid {
		return n.URL.Validate()
	}
	return nil
}

func (n NullableURL) String() string {
	if !n.Valid {
		return "null"
	}
	return n.URL.String()
}

func (n NullableURL) IsZero() bool { return !n.Valid || n.URL == "" }

// Parse returns the URL as a parsed *url.URL, or nil if null
func (n NullableURL) Parse() (*url.URL, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.URL.Parse()
}

func (n NullableURL) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return string(n.URL), nil
}

func (n *NullableURL) Scan(value any) error {
	if value == nil {
		n.Valid = false
		n.URL = ""
		return nil
	}
	n.Valid = true
	return n.URL.Scan(value)
}

func (n NullableURL) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.URL)
}

func (n *NullableURL) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		n.URL = ""
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.URL)
}
