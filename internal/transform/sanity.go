package transform

import (
	"encoding/json"

	"github.com/hegner123/modulacms/internal/model"
)

// SanityTransformer transforms ModulaCMS data to Sanity format
type SanityTransformer struct {
	BaseTransformer
}

// SanityDocument represents a Sanity document
type SanityDocument struct {
	ID         string         `json:"_id"`
	Type       string         `json:"_type"`
	CreatedAt  string         `json:"_createdAt"`
	UpdatedAt  string         `json:"_updatedAt"`
	Rev        string         `json:"_rev,omitempty"`
	Fields     map[string]any `json:"-"` // Flattened into document
}

type SanitySlug struct {
	Current string `json:"current"`
	Type    string `json:"_type"`
}

type SanityImage struct {
	Type  string           `json:"_type"`
	Asset SanityReference  `json:"asset"`
}

type SanityReference struct {
	Ref  string `json:"_ref"`
	Type string `json:"_type"`
}

type SanityBlock struct {
	Type     string              `json:"_type"`
	Children []SanityBlockChild  `json:"children"`
}

type SanityBlockChild struct {
	Type string `json:"_type"`
	Text string `json:"text"`
}

func (s *SanityTransformer) Transform(root model.Root) (any, error) {
	if root.Node == nil {
		return map[string]any{}, nil
	}

	return s.transformNode(root.Node), nil
}

func (s *SanityTransformer) TransformToJSON(root model.Root) ([]byte, error) {
	result, err := s.Transform(root)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (s *SanityTransformer) transformNode(node *model.Node) map[string]any {
	doc := make(map[string]any)

	// System fields
	doc["_id"] = node.Datatype.Content.ContentDataID
	doc["_type"] = s.toSanityType(node.Datatype.Info.Label)
	doc["_createdAt"] = s.getDateCreated(node)
	doc["_updatedAt"] = s.getDateModified(node)
	doc["_rev"] = "v1"

	// Transform fields
	for _, field := range node.Fields {
		key := fieldLabelToKey(field.Info.Label)
		value := s.transformField(field)
		doc[key] = value
	}

	// Transform child nodes
	if len(node.Nodes) > 0 {
		children := make([]map[string]any, 0, len(node.Nodes))
		for _, child := range node.Nodes {
			children = append(children, s.transformNode(child))
		}

		if allChildrenSameType(node.Nodes) {
			doc[pluralize(node.Nodes[0].Datatype.Info.Label)] = children
		} else {
			doc["children"] = children
		}
	}

	return doc
}

func (s *SanityTransformer) transformField(field model.Field) any {
	label := field.Info.Label
	fieldType := field.Info.Type
	value := field.Content.FieldValue

	if value == "" {
		return nil
	}

	// Handle slug fields
	if containsIgnoreCase(label, "slug") {
		return SanitySlug{
			Current: value,
			Type:    "slug",
		}
	}

	// Handle image/asset fields
	if fieldType == "image" || fieldType == "asset" {
		return SanityImage{
			Type: "image",
			Asset: SanityReference{
				Ref:  s.extractAssetRef(value),
				Type: "reference",
			},
		}
	}

	// Handle markdown/richtext as portable text
	if fieldType == "markdown" || fieldType == "richtext" {
		return s.markdownToPortableText(value)
	}

	// Parse other types
	return s.parseFieldValue(value, fieldType)
}

func (s *SanityTransformer) parseFieldValue(value string, fieldType string) any {
	if value == "" {
		return nil
	}
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

func (s *SanityTransformer) markdownToPortableText(markdown string) []SanityBlock {
	// Simplified portable text conversion
	// In production, use proper markdown parser
	paragraphs := splitByNewlines(markdown)
	blocks := make([]SanityBlock, 0, len(paragraphs))

	for _, para := range paragraphs {
		if para == "" {
			continue
		}

		// Remove markdown headers
		text := para
		if len(text) > 0 && text[0] == '#' {
			for i, char := range text {
				if char != '#' && char != ' ' {
					text = text[i:]
					break
				}
			}
		}

		blocks = append(blocks, SanityBlock{
			Type: "block",
			Children: []SanityBlockChild{
				{
					Type: "span",
					Text: text,
				},
			},
		})
	}

	return blocks
}

func (s *SanityTransformer) toSanityType(label string) string {
	// "Blog Post" → "blogpost"
	result := ""
	for _, char := range label {
		if char >= 'A' && char <= 'Z' {
			result += string(char + 32)
		} else if char >= 'a' && char <= 'z' {
			result += string(char)
		} else if char >= '0' && char <= '9' {
			result += string(char)
		}
		// Skip spaces
	}
	return result
}

func (s *SanityTransformer) extractAssetRef(url string) string {
	// Extract filename or hash URL
	return "image-" + hashString(url)
}

func (s *SanityTransformer) getDateCreated(node *model.Node) string {
	return node.Datatype.Content.DateCreated
}

func (s *SanityTransformer) getDateModified(node *model.Node) string {
	if node.Datatype.Content.DateModified != "" {
		return node.Datatype.Content.DateModified
	}
	return s.getDateCreated(node)
}

// Helper functions
func containsIgnoreCase(str, substr string) bool {
	str = toLowerCase(str)
	substr = toLowerCase(substr)

	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

func splitByNewlines(s string) []string {
	paragraphs := []string{}
	current := ""

	for i := 0; i < len(s); i++ {
		if i < len(s)-1 && s[i] == '\n' && s[i+1] == '\n' {
			if current != "" {
				paragraphs = append(paragraphs, current)
				current = ""
			}
			i++ // Skip second newline
		} else if s[i] != '\n' {
			current += string(s[i])
		}
	}

	if current != "" {
		paragraphs = append(paragraphs, current)
	}

	return paragraphs
}

func hashString(s string) string {
	hash := 0
	for i := 0; i < len(s); i++ {
		hash = ((hash << 5) - hash) + int(s[i])
		hash = hash & hash
	}

	if hash < 0 {
		hash = -hash
	}

	return intToBase36(hash)
}

func intToBase36(n int) string {
	if n == 0 {
		return "0"
	}

	const base36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	result := ""

	for n > 0 {
		result = string(base36[n%36]) + result
		n /= 36
	}

	return result
}
