package transform

import (
	"encoding/json"

	"github.com/hegner123/modulacms/internal/model"
)

// StrapiTransformer transforms ModulaCMS data to Strapi format
type StrapiTransformer struct {
	BaseTransformer
}

// StrapiResponse represents a Strapi API response
type StrapiResponse struct {
	Data any            `json:"data"`
	Meta map[string]any `json:"meta"`
}

type StrapiEntry struct {
	ID         string         `json:"id"`
	Attributes map[string]any `json:"attributes"`
}

type StrapiMedia struct {
	Data StrapiMediaData `json:"data"`
}

type StrapiMediaData struct {
	ID         int64                  `json:"id"`
	Attributes StrapiMediaAttributes  `json:"attributes"`
}

type StrapiMediaAttributes struct {
	URL             string `json:"url"`
	Name            string `json:"name"`
	AlternativeText string `json:"alternativeText,omitempty"`
}

type StrapiRelation struct {
	Data []StrapiEntry `json:"data"`
}

func (s *StrapiTransformer) Transform(root model.Root) (any, error) {
	if root.Node == nil {
		return StrapiResponse{
			Data: StrapiEntry{},
			Meta: map[string]any{},
		}, nil
	}

	return StrapiResponse{
		Data: s.transformNode(root.Node),
		Meta: map[string]any{},
	}, nil
}

func (s *StrapiTransformer) TransformToJSON(root model.Root) ([]byte, error) {
	result, err := s.Transform(root)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (s *StrapiTransformer) transformNode(node *model.Node) StrapiEntry {
	entry := StrapiEntry{
		ID:         node.Datatype.Content.ContentDataID,
		Attributes: make(map[string]any),
	}

	// Add timestamps
	entry.Attributes["createdAt"] = s.getDateCreated(node)
	entry.Attributes["updatedAt"] = s.getDateModified(node)

	// Transform fields
	for _, field := range node.Fields {
		key := fieldLabelToKey(field.Info.Label)
		value := s.transformField(field)
		entry.Attributes[key] = value
	}

	// Transform child nodes as relations
	if len(node.Nodes) > 0 {
		children := make([]StrapiEntry, 0, len(node.Nodes))
		for _, child := range node.Nodes {
			children = append(children, s.transformNode(child))
		}

		key := "children"
		if allChildrenSameType(node.Nodes) {
			key = pluralize(node.Nodes[0].Datatype.Info.Label)
		}
		entry.Attributes[key] = StrapiRelation{Data: children}
	}

	return entry
}

func (s *StrapiTransformer) transformField(field model.Field) any {
	fieldType := field.Info.Type
	value := field.Content.FieldValue

	if value == "" {
		return nil
	}

	// Handle media fields
	if fieldType == "image" || fieldType == "asset" {
		filename := extractFilename(value)

		return StrapiMedia{
			Data: StrapiMediaData{
				ID: int64(hashStringInt(value)),
				Attributes: StrapiMediaAttributes{
					URL:             value,
					Name:            filename,
					AlternativeText: field.Info.Label,
				},
			},
		}
	}

	// Parse other types
	return s.parseFieldValue(value, fieldType)
}

func (s *StrapiTransformer) parseFieldValue(value string, fieldType string) any {
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

func (s *StrapiTransformer) getDateCreated(node *model.Node) string {
	return node.Datatype.Content.DateCreated
}

func (s *StrapiTransformer) getDateModified(node *model.Node) string {
	if node.Datatype.Content.DateModified != "" {
		return node.Datatype.Content.DateModified
	}
	return s.getDateCreated(node)
}

// Helper functions
func extractFilename(url string) string {
	// Extract filename from URL
	filename := ""
	lastSlash := -1

	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash >= 0 && lastSlash < len(url)-1 {
		filename = url[lastSlash+1:]
	} else {
		filename = "file"
	}

	return filename
}

func hashStringInt(s string) int {
	hash := 0
	for i := 0; i < len(s); i++ {
		hash = ((hash << 5) - hash) + int(s[i])
		hash = hash & hash
	}

	if hash < 0 {
		return -hash
	}

	return hash
}
