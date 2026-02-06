package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Slug represents a URL path (starts with /, lowercase alphanumeric segments with hyphens)
// Examples: /, /about, /about/careers, /blog/2024/my-post
type Slug string

func (s Slug) Validate() error {
	if s == "" {
		return fmt.Errorf("Slug: cannot be empty")
	}
	if len(s) > 255 {
		return fmt.Errorf("Slug: too long (max 255 chars)")
	}
	str := string(s)

	// Must start with /
	if str[0] != '/' {
		return fmt.Errorf("Slug: must start with / (got %q)", s)
	}

	// Root path is valid
	if str == "/" {
		return nil
	}

	// Check each character and segment structure
	prevChar := byte('/')
	for i := 1; i < len(str); i++ {
		c := str[i]
		switch {
		case c >= 'a' && c <= 'z':
			// lowercase letter - always valid
		case c >= '0' && c <= '9':
			// digit - always valid
		case c == '-':
			// hyphen - not allowed after / or another -
			if prevChar == '/' || prevChar == '-' {
				return fmt.Errorf("Slug: invalid format %q (hyphen cannot follow / or another hyphen)", s)
			}
		case c == '/':
			// slash - not allowed after / or -
			if prevChar == '/' || prevChar == '-' {
				return fmt.Errorf("Slug: invalid format %q (/ cannot follow / or hyphen)", s)
			}
		default:
			return fmt.Errorf("Slug: invalid character %q in %q (allowed: a-z, 0-9, -, /)", string(c), s)
		}
		prevChar = c
	}

	// Cannot end with - or /
	lastChar := str[len(str)-1]
	if lastChar == '-' || lastChar == '/' {
		return fmt.Errorf("Slug: cannot end with %q", string(lastChar))
	}

	return nil
}

func (s Slug) String() string { return string(s) }

func (s Slug) IsZero() bool { return s == "" }

// Slugify converts a string to a valid slug path
// Input "Home" becomes "/home", "About Us" becomes "/about-us"
// Input "/about/careers" is preserved (with cleanup)
func Slugify(input string) Slug {
	// Lowercase
	result := strings.ToLower(input)
	// Replace spaces and underscores with hyphens
	result = strings.ReplaceAll(result, " ", "-")
	result = strings.ReplaceAll(result, "_", "-")
	// Keep only alphanumeric, hyphens, and slashes
	var sb strings.Builder
	for _, r := range result {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '/' {
			sb.WriteRune(r)
		}
	}
	result = sb.String()
	// Collapse multiple hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}
	// Collapse multiple slashes
	for strings.Contains(result, "//") {
		result = strings.ReplaceAll(result, "//", "/")
	}
	// Remove hyphen-slash and slash-hyphen combinations
	for strings.Contains(result, "-/") {
		result = strings.ReplaceAll(result, "-/", "/")
	}
	for strings.Contains(result, "/-") {
		result = strings.ReplaceAll(result, "/-", "/")
	}
	// Trim hyphens and slashes from end
	result = strings.TrimRight(result, "-/")
	// Ensure starts with /
	if !strings.HasPrefix(result, "/") {
		result = "/" + result
	}
	// Handle empty result (just /)
	if result == "" {
		result = "/"
	}
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
	// Skip validation for legacy data - validation is enforced on write
	return nil
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

func (e Email) Validate() error {
	if e == "" {
		return fmt.Errorf("Email: cannot be empty")
	}
	if len(e) > 254 {
		return fmt.Errorf("Email: too long (max 254 chars)")
	}

	str := string(e)

	// Find @ symbol - must have exactly one
	atIndex := strings.Index(str, "@")
	if atIndex == -1 {
		return fmt.Errorf("Email: missing @ in %q", e)
	}
	if strings.Count(str, "@") > 1 {
		return fmt.Errorf("Email: multiple @ symbols in %q", e)
	}

	local := str[:atIndex]
	domain := str[atIndex+1:]

	// Local part validation
	if len(local) == 0 {
		return fmt.Errorf("Email: empty local part in %q", e)
	}
	for _, c := range local {
		if !isEmailLocalChar(c) {
			return fmt.Errorf("Email: invalid character %q in local part of %q", string(c), e)
		}
	}

	// Domain validation
	if len(domain) == 0 {
		return fmt.Errorf("Email: empty domain in %q", e)
	}
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("Email: domain must contain a dot in %q", e)
	}

	// Check domain characters and structure
	lastDotIndex := strings.LastIndex(domain, ".")
	tld := domain[lastDotIndex+1:]
	if len(tld) < 2 {
		return fmt.Errorf("Email: TLD must be at least 2 characters in %q", e)
	}

	for _, c := range domain {
		if !isEmailDomainChar(c) {
			return fmt.Errorf("Email: invalid character %q in domain of %q", string(c), e)
		}
	}

	return nil
}

// isEmailLocalChar returns true if c is valid in email local part
func isEmailLocalChar(c rune) bool {
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	// Special characters allowed in local part: ._%+-
	return c == '.' || c == '_' || c == '%' || c == '+' || c == '-'
}

// isEmailDomainChar returns true if c is valid in email domain
func isEmailDomainChar(c rune) bool {
	if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
		return true
	}
	if c >= '0' && c <= '9' {
		return true
	}
	return c == '.' || c == '-'
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
