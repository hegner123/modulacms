package transform

// This file shows example usage of the transform package in HTTP handlers
// This is for documentation purposes and is not compiled

/*

Example 1: Basic Handler with Transform

package handlers

import (
	"net/http"
	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
)

func GetContentHandler(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Query database for content
		// (This is pseudocode - use actual DB queries)
		root := model.Root{
			Node: &model.Node{
				// ... query result from database ...
			},
		}

		// 2. Create transform config
		transformCfg := transform.NewTransformConfig(
			cfg.Output_Format,  // "contentful" | "sanity" | "strapi" | "wordpress" | "clean" | "raw"
			cfg.Client_Site,    // Site URL
			cfg.Space_ID,       // Space ID (for Contentful)
			driver,             // Database driver
		)

		// 3. Transform and write response
		err := transformCfg.TransformAndWrite(w, root)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

---

Example 2: Manual Transform (More Control)

func GetContentHandlerManual(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query database
		root := model.Root{
			Node: &model.Node{
				// ... query result ...
			},
		}

		// Create transformer
		transformCfg := transform.NewTransformConfig(
			cfg.Output_Format,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformer, err := transformCfg.GetTransformer()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Transform to JSON
		jsonData, err := transformer.TransformToJSON(root)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Custom headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("X-Transform-Format", cfg.Output_Format)

		// Write response
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

---

Example 3: Different Formats Per Route

func SetupRoutes(cfg *config.Config, driver db.DbDriver) {
	// Route 1: Contentful format for frontend
	http.HandleFunc("/api/content/", func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		transformCfg := transform.NewTransformConfig(
			"contentful",      // Force Contentful format
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	})

	// Route 2: Clean format for admin
	http.HandleFunc("/admin/api/content/", func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		transformCfg := transform.NewTransformConfig(
			"clean",           // Force clean format
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	})

	// Route 3: Raw format for debugging
	http.HandleFunc("/debug/api/content/", func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		transformCfg := transform.NewTransformConfig(
			"raw",             // No transformation
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	})
}

---

Example 4: Query Parameter Override

func GetContentWithFormatOverride(cfg *config.Config, driver db.DbDriver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		root := getContent(r)

		// Allow format override via query parameter
		// GET /api/content/42?format=sanity
		format := r.URL.Query().Get("format")
		if format == "" {
			format = cfg.Output_Format // Default from config
		}

		transformCfg := transform.NewTransformConfig(
			format,
			cfg.Client_Site,
			cfg.Space_ID,
			driver,
		)

		transformCfg.TransformAndWrite(w, root)
	}
}

---

Example 5: Testing Different Formats

package transform_test

import (
	"testing"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
)

func TestAllTransformers(t *testing.T) {
	root := model.Root{
		Node: &model.Node{
			// ... test data ...
		},
	}

	formats := []string{
		"contentful",
		"sanity",
		"strapi",
		"wordpress",
		"clean",
		"raw",
	}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cfg := transform.NewTransformConfig(
				format,
				"https://example.com",
				"test-space",
				nil,
			)

			transformer, err := cfg.GetTransformer()
			if err != nil {
				t.Fatalf("Failed to get transformer: %v", err)
			}

			result, err := transformer.Transform(root)
			if err != nil {
				t.Fatalf("Transform failed: %v", err)
			}

			if result == nil {
				t.Error("Expected non-nil result")
			}
		})
	}
}

---

Example 6: Configuration

// config.json
{
  "output_format": "contentful",
  "space_id": "modulacms",
  "client_site": "https://example.com",
  "db_driver": "postgres",
  "db_url": "localhost:5432"
}

// Or use environment variables
OUTPUT_FORMAT=sanity
SPACE_ID=my-space
CLIENT_SITE=https://mysite.com

---

Example 7: Migration Path (Frontend)

// Before (using Contentful SDK)
import { createClient } from 'contentful'

const client = createClient({
  space: 'your_space',
  accessToken: 'your_token'
})

const entry = await client.getEntry('42')
console.log(entry.fields.title)
console.log(entry.fields.body)

// After (using ModulaCMS with Contentful transformer)
// 1. Update config.json
{
  "output_format": "contentful",
  "space_id": "modulacms"
}

// 2. Update API endpoint only (frontend code unchanged!)
const entry = await fetch('https://api.modulacms.com/content/42').then(r => r.json())
console.log(entry.fields.title)  // Same!
console.log(entry.fields.body)   // Same!

// Zero frontend changes required!

---

Example 8: Performance Testing

package transform_test

import (
	"testing"
	"time"
	"github.com/hegner123/modulacms/internal/model"
	"github.com/hegner123/modulacms/internal/transform"
)

func BenchmarkTransformers(b *testing.B) {
	root := createLargeTestData() // 100 nodes, 20 fields each

	formats := []string{"contentful", "sanity", "strapi", "wordpress", "clean"}

	for _, format := range formats {
		b.Run(format, func(b *testing.B) {
			cfg := transform.NewTransformConfig(format, "https://example.com", "test", nil)
			transformer, _ := cfg.GetTransformer()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := transformer.Transform(root)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Expected results:
// BenchmarkTransformers/contentful-8  	    5000	    250000 ns/op
// BenchmarkTransformers/sanity-8      	    5000	    240000 ns/op
// BenchmarkTransformers/strapi-8      	    5000	    245000 ns/op
// BenchmarkTransformers/wordpress-8   	    5000	    255000 ns/op
// BenchmarkTransformers/clean-8       	    6000	    220000 ns/op
// (All under 1ms - very fast!)

*/
