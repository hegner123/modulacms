package search

import (
	"strings"
	"unicode"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/publishing"
)

// SplitByHeadings splits HTML content by heading tags (h1-h6) into sections.
// Each section contains the heading text, a URL-safe anchor slug, and the
// body text (HTML-stripped) until the next heading or end of string.
func SplitByHeadings(html string) []Section {
	var sections []Section
	runes := []rune(html)
	n := len(runes)

	i := 0
	var currentHeading string
	var currentAnchor string
	var bodyBuilder strings.Builder
	inSection := false

	for i < n {
		if runes[i] == '<' && i+2 < n && isHeadingStart(runes, i, n) {
			// Flush previous section if any
			if inSection {
				body := StripHTML(bodyBuilder.String())
				sections = append(sections, Section{
					Heading: currentHeading,
					Anchor:  currentAnchor,
					Body:    body,
				})
				bodyBuilder.Reset()
			}

			// Extract heading level digit position (already validated by isHeadingStart)
			// Find the closing '>' of the opening tag
			tagEnd := i
			for tagEnd < n && runes[tagEnd] != '>' {
				tagEnd++
			}
			if tagEnd >= n {
				// Malformed: no closing '>', treat rest as body text
				bodyBuilder.WriteString(string(runes[i:]))
				break
			}
			tagEnd++ // move past '>'

			// Find the closing heading tag </hN>
			headingContentStart := tagEnd
			headingContentEnd := findClosingHeadingTag(runes, headingContentStart, n)

			// Extract heading inner HTML and strip tags from it
			var headingHTML string
			if headingContentEnd < n {
				headingHTML = string(runes[headingContentStart:headingContentEnd])
			} else {
				headingHTML = string(runes[headingContentStart:])
			}
			currentHeading = StripHTML(headingHTML)
			currentHeading = strings.TrimSpace(currentHeading)
			currentAnchor = slugify(currentHeading)
			inSection = true

			// Skip past the closing tag
			if headingContentEnd < n {
				closeEnd := headingContentEnd
				for closeEnd < n && runes[closeEnd] != '>' {
					closeEnd++
				}
				if closeEnd < n {
					closeEnd++ // past '>'
				}
				i = closeEnd
			} else {
				i = n
			}
			continue
		}

		if inSection {
			bodyBuilder.WriteRune(runes[i])
		}
		i++
	}

	// Flush last section
	if inSection {
		body := StripHTML(bodyBuilder.String())
		sections = append(sections, Section{
			Heading: currentHeading,
			Anchor:  currentAnchor,
			Body:    body,
		})
	}

	return sections
}

// isHeadingStart checks if position i in runes starts an opening heading tag
// like <h1, <h2, ..., <h6 (case-insensitive). Expects runes[i] == '<'.
func isHeadingStart(runes []rune, i, n int) bool {
	// Need at least '<', 'h'/'H', digit
	if i+2 >= n {
		return false
	}
	ch := runes[i+1]
	if ch != 'h' && ch != 'H' {
		return false
	}
	digit := runes[i+2]
	return digit >= '1' && digit <= '6'
}

// findClosingHeadingTag scans from start looking for </h followed by a digit
// and >. Returns the index of '<' in the closing tag, or n if not found.
func findClosingHeadingTag(runes []rune, start, n int) int {
	for j := start; j < n; j++ {
		if runes[j] != '<' {
			continue
		}
		// Check for </h or </H followed by digit 1-6
		if j+3 < n && runes[j+1] == '/' {
			ch := runes[j+2]
			if (ch == 'h' || ch == 'H') && j+3 < n {
				digit := runes[j+3]
				if digit >= '1' && digit <= '6' {
					return j
				}
			}
		}
	}
	return n
}

// slugify converts text to a URL-safe anchor: lowercase, non-alphanumeric
// characters replaced with hyphens, consecutive hyphens collapsed, leading
// and trailing hyphens trimmed.
func slugify(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	prevHyphen := false

	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevHyphen = false
		} else {
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}

	result := b.String()
	result = strings.TrimRight(result, "-")
	return result
}

// BuildDocuments converts a published snapshot and its version metadata into
// a flat list of SearchDocument values suitable for indexing. Each root content
// node produces one document; richtext fields with headings produce additional
// per-section documents.
func BuildDocuments(snapshot *publishing.Snapshot, version db.ContentVersion, cfg SearchConfig) []SearchDocument {
	if snapshot == nil || len(snapshot.ContentData) == 0 {
		return nil
	}

	routeSlug := snapshot.Route.Slug
	routeTitle := snapshot.Route.Title
	locale := version.Locale
	publishedAt := version.DateCreated.String()

	// Build field lookup: FieldID → FieldsJSON
	fieldsByID := make(map[string]db.FieldsJSON, len(snapshot.Fields))
	for _, f := range snapshot.Fields {
		fieldsByID[f.FieldID] = f
	}

	// Build datatype lookup: DatatypeID → DatatypeJSON
	datatypesByID := make(map[string]db.DatatypeJSON, len(snapshot.Datatypes))
	for _, dt := range snapshot.Datatypes {
		datatypesByID[dt.DatatypeID] = dt
	}

	// Root content node is always first by construction
	root := snapshot.ContentData[0]
	rootDT := datatypesByID[root.DatatypeID]

	// Collect text-bearing field values for the root node
	fields := make(map[string]string)
	fields["_title"] = routeTitle

	// Track which fields are richtext for section splitting
	var richtextFields []struct {
		name  string
		value string
	}

	for _, cf := range snapshot.ContentFields {
		if cf.ContentDataID != root.ContentDataID {
			continue
		}
		fieldDef, ok := fieldsByID[cf.FieldID]
		if !ok {
			continue
		}
		if !cfg.IndexableFieldTypes[fieldDef.Type] {
			continue
		}
		fields[fieldDef.Name] = cf.FieldValue
		if fieldDef.Type == "richtext" {
			richtextFields = append(richtextFields, struct {
				name  string
				value string
			}{name: fieldDef.Name, value: cf.FieldValue})
		}
	}

	// Add child content node fields to the root document with qualified keys
	for _, child := range snapshot.ContentData[1:] {
		childDT := datatypesByID[child.DatatypeID]
		for _, cf := range snapshot.ContentFields {
			if cf.ContentDataID != child.ContentDataID {
				continue
			}
			fieldDef, ok := fieldsByID[cf.FieldID]
			if !ok {
				continue
			}
			if !cfg.IndexableFieldTypes[fieldDef.Type] {
				continue
			}
			qualifiedKey := childDT.Name + "." + fieldDef.Name
			fields[qualifiedKey] = cf.FieldValue
		}
	}

	// Build root document
	rootDoc := SearchDocument{
		ID:            root.ContentDataID,
		ContentDataID: root.ContentDataID,
		RouteSlug:     routeSlug,
		RouteTitle:    routeTitle,
		DatatypeName:  rootDT.Name,
		DatatypeLabel: rootDT.Label,
		Locale:        locale,
		Fields:        fields,
		PublishedAt:   publishedAt,
	}

	docs := []SearchDocument{rootDoc}

	// Split richtext fields by headings into section documents
	for _, rtf := range richtextFields {
		sections := SplitByHeadings(rtf.value)
		for _, section := range sections {
			sectionDoc := SearchDocument{
				ID:            root.ContentDataID + "#" + section.Anchor,
				ContentDataID: root.ContentDataID,
				RouteSlug:     routeSlug,
				RouteTitle:    routeTitle,
				DatatypeName:  rootDT.Name,
				DatatypeLabel: rootDT.Label,
				Locale:        locale,
				Section:       section.Heading,
				SectionAnchor: section.Anchor,
				Fields: map[string]string{
					"_section_heading": section.Heading,
					"_section_body":    section.Body,
				},
				PublishedAt: publishedAt,
			}
			docs = append(docs, sectionDoc)
		}
	}

	return docs
}
