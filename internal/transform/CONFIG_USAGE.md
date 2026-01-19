# Transform Config Usage Guide

**Created:** 2026-01-16
**Purpose:** How to configure and use output format transformers

---

## Configuration File

### config.json

```json
{
  "output_format": "contentful",
  "space_id": "modulacms",
  "client_site": "https://example.com",
  "db_driver": "postgres"
}
```

---

## Valid Output Formats

### Defined in `internal/config/config.go`

```go
const (
	FormatContentful config.OutputFormat = "contentful"
	FormatSanity     config.OutputFormat = "sanity"
	FormatStrapi     config.OutputFormat = "strapi"
	FormatWordPress  config.OutputFormat = "wordpress"
	FormatClean      config.OutputFormat = "clean"
	FormatRaw        config.OutputFormat = "raw"
	FormatDefault    config.OutputFormat = "" // Defaults to raw
)
```

---

## Usage in Handlers

### Method 1: Using Config Struct (Recommended)

```go
package handlers

import (
	"net/http"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/transform"
)

func GetContentHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query database
		root := queryDatabase(r)

		// Create transform config using typed format
		transformCfg := transform.NewTransformConfig(
			cfg.Output_Format,  // config.OutputFormat type
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		// Transform and write
		transformCfg.TransformAndWrite(w, root)
	}
}
```

---

### Method 2: Using String Format

```go
func GetContentHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		root := queryDatabase(r)

		// Get format from query parameter
		formatStr := r.URL.Query().Get("format")
		if formatStr == "" {
			formatStr = string(cfg.Output_Format)
		}

		// Validate format
		if !config.IsValidOutputFormat(formatStr) {
			http.Error(w, "Invalid output format", http.StatusBadRequest)
			return
		}

		// Create transform config from string
		transformCfg := transform.NewTransformConfigFromString(
			formatStr,
			cfg.Client_Site,
			cfg.Space_ID,
			nil,
		)

		transformCfg.TransformAndWrite(w, root)
	}
}
```

---

### Method 3: Using Constants

```go
func GetContentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		root := queryDatabase(r)

		// Use constant directly
		transformCfg := transform.NewTransformConfig(
			config.FormatContentful,  // Typed constant
			"https://example.com",
			"modulacms",
			nil,
		)

		transformCfg.TransformAndWrite(w, root)
	}
}
```

---

## Format Validation

### Check if format is valid

```go
import "github.com/hegner123/modulacms/internal/config"

func validateFormat(format string) error {
	if !config.IsValidOutputFormat(format) {
		validFormats := config.GetValidOutputFormats()
		return fmt.Errorf("invalid format '%s', valid options: %v", format, validFormats)
	}
	return nil
}
```

### Get all valid formats

```go
validFormats := config.GetValidOutputFormats()
// ["contentful", "sanity", "strapi", "wordpress", "clean", "raw"]
```

---

## Environment Variables

### Override config.json with environment variables

```bash
# Set output format via env var
export OUTPUT_FORMAT=contentful
export SPACE_ID=modulacms
export CLIENT_SITE=https://example.com
```

### Load in code

```go
func loadConfig() *config.Config {
	cfg := loadConfigFromFile()

	// Override with env vars
	if envFormat := os.Getenv("OUTPUT_FORMAT"); envFormat != "" {
		if config.IsValidOutputFormat(envFormat) {
			cfg.Output_Format = config.OutputFormat(envFormat)
		}
	}

	return cfg
}
```

---

## Query Parameter Override

### Allow format override per request

```go
func GetContentHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		root := queryDatabase(r)

		// Get format from query param or use default
		format := cfg.Output_Format
		if paramFormat := r.URL.Query().Get("format"); paramFormat != "" {
			if config.IsValidOutputFormat(paramFormat) {
				format = config.OutputFormat(paramFormat)
			} else {
				http.Error(w, "Invalid format parameter", http.StatusBadRequest)
				return
			}
		}

		transformCfg := transform.NewTransformConfig(
			format,
			cfg.Client_Site,
			cfg.Space_ID,
			nil,
		)

		transformCfg.TransformAndWrite(w, root)
	}
}
```

**Example requests:**
```bash
# Use default format from config
GET /content/42

# Override with query parameter
GET /content/42?format=contentful
GET /content/42?format=sanity
GET /content/42?format=clean
GET /content/42?format=raw
```

---

## Header-Based Format Selection

### Use Accept header for content negotiation

```go
func GetContentHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		root := queryDatabase(r)

		// Determine format from Accept header
		format := cfg.Output_Format
		accept := r.Header.Get("Accept")

		switch accept {
		case "application/vnd.contentful+json":
			format = config.FormatContentful
		case "application/vnd.sanity+json":
			format = config.FormatSanity
		case "application/vnd.strapi+json":
			format = config.FormatStrapi
		case "application/vnd.wordpress+json":
			format = config.FormatWordPress
		case "application/vnd.modulacms.clean+json":
			format = config.FormatClean
		case "application/json":
			// Use default
		}

		transformCfg := transform.NewTransformConfig(
			format,
			cfg.Client_Site,
			cfg.Space_ID,
			nil,
		)

		transformCfg.TransformAndWrite(w, root)
	}
}
```

**Example requests:**
```bash
curl -H "Accept: application/vnd.contentful+json" /content/42
curl -H "Accept: application/vnd.sanity+json" /content/42
curl -H "Accept: application/json" /content/42  # Uses default
```

---

## Per-Route Format Configuration

### Different formats for different routes

```go
func SetupRoutes(cfg *config.Config, driver db.DbDriver) {
	// Public API - Contentful format
	http.HandleFunc("/api/content/", func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		transformCfg := transform.NewTransformConfig(
			config.FormatContentful,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	})

	// Admin API - Clean format
	http.HandleFunc("/admin/api/content/", func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		transformCfg := transform.NewTransformConfig(
			config.FormatClean,
			cfg.Admin_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	})

	// Debug API - Raw format
	http.HandleFunc("/debug/api/content/", func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		transformCfg := transform.NewTransformConfig(
			config.FormatRaw,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	})
}
```

---

## Default Behavior

### When output_format is empty or missing

```go
// In config.json
{
  "output_format": ""  // Defaults to raw (no transformation)
}

// Or omitted entirely
{
  // No output_format field
}
```

**Behavior:**
- Empty string `""` → `FormatDefault` → treated as `FormatRaw`
- Returns original ModulaCMS JSON structure
- No transformation overhead

---

## Example Configurations

### Development (Raw format for debugging)

```json
{
  "environment": "development",
  "output_format": "raw",
  "space_id": "dev",
  "client_site": "http://localhost:3000"
}
```

### Staging (Clean format for testing)

```json
{
  "environment": "staging",
  "output_format": "clean",
  "space_id": "staging",
  "client_site": "https://staging.example.com"
}
```

### Production (Contentful format for compatibility)

```json
{
  "environment": "production",
  "output_format": "contentful",
  "space_id": "production",
  "client_site": "https://example.com"
}
```

### Migration from Contentful

```json
{
  "environment": "production",
  "output_format": "contentful",
  "space_id": "your-old-contentful-space-id",
  "client_site": "https://example.com",
  "db_driver": "postgres",
  "db_url": "postgres://..."
}
```

**Frontend changes:** ZERO
- Just point SDK at new API URL
- Same format, same responses
- Seamless migration

---

## Testing Different Formats

### Unit test with all formats

```go
func TestAllFormats(t *testing.T) {
	root := createTestData()

	formats := []config.OutputFormat{
		config.FormatContentful,
		config.FormatSanity,
		config.FormatStrapi,
		config.FormatWordPress,
		config.FormatClean,
		config.FormatRaw,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			cfg := transform.NewTransformConfig(
				format,
				"https://test.com",
				"test",
				nil,
			)

			transformer, err := cfg.GetTransformer()
			if err != nil {
				t.Fatal(err)
			}

			result, err := transformer.Transform(root)
			if err != nil {
				t.Fatal(err)
			}

			if result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}
```

---

## Error Handling

### Invalid format

```go
func CreateHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := r.URL.Query().Get("format")

		if format != "" && !config.IsValidOutputFormat(format) {
			validFormats := config.GetValidOutputFormats()
			http.Error(
				w,
				fmt.Sprintf("Invalid format '%s'. Valid options: %v", format, validFormats),
				http.StatusBadRequest,
			)
			return
		}

		// Process request...
	}
}
```

---

## Summary

### Configuration Methods

1. **config.json** - Static configuration
2. **Environment variables** - Runtime override
3. **Query parameters** - Per-request override
4. **Accept headers** - Content negotiation
5. **Route-based** - Different formats per route

### Valid Formats

- `contentful` - Contentful API format
- `sanity` - Sanity client format
- `strapi` - Strapi REST API format
- `wordpress` - WordPress REST API format
- `clean` - Clean ModulaCMS format
- `raw` - Original ModulaCMS format (no transformation)
- `""` - Empty defaults to `raw`

### Type Safety

- Use `config.OutputFormat` type for compile-time safety
- Use constants (`config.FormatContentful`, etc.) to avoid typos
- Validate user input with `config.IsValidOutputFormat()`

---

**Last Updated:** 2026-01-16
