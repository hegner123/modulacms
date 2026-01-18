package transform

import (
	"encoding/json"

	"github.com/hegner123/modulacms/internal/model"
)

// WordPressTransformer transforms ModulaCMS data to WordPress REST API format
type WordPressTransformer struct {
	BaseTransformer
}

// WordPressPost represents a WordPress post/page
type WordPressPost struct {
	ID            int64                  `json:"id"`
	Date          string                 `json:"date"`
	DateGMT       string                 `json:"date_gmt"`
	Modified      string                 `json:"modified"`
	ModifiedGMT   string                 `json:"modified_gmt"`
	Slug          string                 `json:"slug"`
	Status        string                 `json:"status"`
	Type          string                 `json:"type"`
	Link          string                 `json:"link"`
	Title         WordPressRendered      `json:"title"`
	Content       WordPressContent       `json:"content"`
	Excerpt       WordPressContent       `json:"excerpt"`
	Author        int64                  `json:"author"`
	FeaturedMedia int64                  `json:"featured_media"`
	CommentStatus string                 `json:"comment_status"`
	PingStatus    string                 `json:"ping_status"`
	Meta          map[string]any         `json:"meta"`
	ACF           map[string]any         `json:"acf,omitempty"`
}

type WordPressRendered struct {
	Rendered string `json:"rendered"`
}

type WordPressContent struct {
	Rendered  string `json:"rendered"`
	Protected bool   `json:"protected"`
}

func (w *WordPressTransformer) Transform(root model.Root) (any, error) {
	if root.Node == nil {
		return WordPressPost{}, nil
	}

	return w.transformNode(root.Node), nil
}

func (w *WordPressTransformer) TransformToJSON(root model.Root) ([]byte, error) {
	result, err := w.Transform(root)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (w *WordPressTransformer) transformNode(node *model.Node) WordPressPost {
	fields := w.extractFields(node)

	// Extract standard fields
	title := w.getFieldString(fields, "title")
	slug := w.getFieldString(fields, "slug")
	body := w.getFieldString(fields, "body")
	content := w.getFieldString(fields, "content")
	excerpt := w.getFieldString(fields, "excerpt")
	published := w.getFieldBool(fields, "published")
	featuredImage := w.getFieldString(fields, "featuredImage")

	// Generate slug if not provided
	if slug == "" && title != "" {
		slug = w.generateSlug(title)
	}

	// Use body or content
	mainContent := body
	if mainContent == "" {
		mainContent = content
	}

	// Generate excerpt if not provided
	if excerpt == "" && mainContent != "" {
		excerpt = w.generateExcerpt(mainContent, 150)
	}

	// Determine status
	status := "draft"
	if published {
		status = "publish"
	}

	// Determine type
	postType := w.toWordPressType(node.Datatype.Info.Label)

	// Build post
	post := WordPressPost{
		ID:            node.Datatype.Content.ContentDataID,
		Date:          w.getDateCreated(node),
		DateGMT:       w.getDateCreated(node),
		Modified:      w.getDateModified(node),
		ModifiedGMT:   w.getDateModified(node),
		Slug:          slug,
		Status:        status,
		Type:          postType,
		Link:          w.generateLink(postType, slug),
		Title:         WordPressRendered{Rendered: title},
		Content:       WordPressContent{Rendered: w.renderContent(mainContent), Protected: false},
		Excerpt:       WordPressContent{Rendered: w.renderExcerpt(excerpt), Protected: false},
		Author:        node.Datatype.Content.AuthorID,
		FeaturedMedia: w.hashMedia(featuredImage),
		CommentStatus: "open",
		PingStatus:    "open",
		Meta:          map[string]any{},
		ACF:           w.extractACF(fields),
	}

	return post
}

func (w *WordPressTransformer) extractFields(node *model.Node) map[string]any {
	fields := make(map[string]any)

	for _, field := range node.Fields {
		key := fieldLabelToKey(field.Info.Label.(string))
		value := w.parseFieldValue(field.Content.FieldValue, field.Info.Type)
		fields[key] = value
	}

	return fields
}

func (w *WordPressTransformer) extractACF(fields map[string]any) map[string]any {
	acf := make(map[string]any)

	// Standard WordPress fields
	standardFields := map[string]bool{
		"title": true, "slug": true, "body": true,
		"content": true, "excerpt": true, "published": true,
		"featuredImage": true,
	}

	// Put custom fields in ACF
	for key, value := range fields {
		if !standardFields[key] {
			acf[key] = value
		}
	}

	return acf
}

func (w *WordPressTransformer) parseFieldValue(value string, fieldType string) any {
	switch fieldType {
	case "boolean":
		return value == "true"
	case "number", "decimal", "float":
		return parseFloat(value)
	case "integer", "int":
		return parseInt(value)
	case "json":
		var jsonData any
		if err := json.Unmarshal([]byte(value), &jsonData); err == nil {
			return jsonData
		}
		return value
	default:
		return value
	}
}

func (w *WordPressTransformer) getFieldString(fields map[string]any, key string) string {
	if val, ok := fields[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (w *WordPressTransformer) getFieldBool(fields map[string]any, key string) bool {
	if val, ok := fields[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (w *WordPressTransformer) generateSlug(title string) string {
	slug := ""

	for _, char := range toLowerCase(title) {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			slug += string(char)
		} else if char == ' ' || char == '_' {
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug += "-"
			}
		}
	}

	// Remove trailing dash
	if len(slug) > 0 && slug[len(slug)-1] == '-' {
		slug = slug[:len(slug)-1]
	}

	return slug
}

func (w *WordPressTransformer) generateExcerpt(content string, length int) string {
	// Strip markdown symbols
	text := ""
	for _, char := range content {
		if char != '#' && char != '*' {
			text += string(char)
		}
	}

	// Trim
	text = trimString(text)

	// Truncate
	if len(text) > length {
		return text[:length] + "..."
	}

	return text
}

func (w *WordPressTransformer) renderContent(content string) string {
	// Simple markdown to HTML (basic conversion)
	// In production, use proper markdown parser
	if content == "" {
		return ""
	}

	// Very basic: wrap in paragraph
	return "<p>" + content + "</p>"
}

func (w *WordPressTransformer) renderExcerpt(excerpt string) string {
	if excerpt == "" {
		return ""
	}
	return "<p>" + excerpt + "</p>"
}

func (w *WordPressTransformer) toWordPressType(label string) string {
	lower := toLowerCase(label)

	if containsIgnoreCase(lower, "post") || containsIgnoreCase(lower, "article") {
		return "post"
	}

	if containsIgnoreCase(lower, "page") {
		return "page"
	}

	// Custom post type: convert to underscore
	result := ""
	for _, char := range lower {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			result += string(char)
		} else if char == ' ' || char == '-' {
			result += "_"
		}
	}

	return result
}

func (w *WordPressTransformer) generateLink(postType, slug string) string {
	siteURL := w.SiteURL
	if siteURL == "" {
		siteURL = "https://example.com"
	}

	return siteURL + "/" + postType + "/" + slug
}

func (w *WordPressTransformer) hashMedia(url string) int64 {
	if url == "" {
		return 0
	}

	return int64(hashStringInt(url))
}

func (w *WordPressTransformer) getDateCreated(node *model.Node) string {
	if node.Datatype.Content.DateCreated.Valid {
		return node.Datatype.Content.DateCreated.String
	}
	return ""
}

func (w *WordPressTransformer) getDateModified(node *model.Node) string {
	if node.Datatype.Content.DateModified.Valid {
		return node.Datatype.Content.DateModified.String
	}
	return w.getDateCreated(node)
}

// Helper functions
func trimString(s string) string {
	start := 0
	end := len(s)

	// Trim leading whitespace
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	// Trim trailing whitespace
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	return s[start:end]
}
