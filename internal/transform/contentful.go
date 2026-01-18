package transform

import (
	"database/sql"
	"encoding/json"
	"hash/fnv"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

// ContentfulTransformer transforms ModulaCMS data to Contentful format
type ContentfulTransformer struct {
	BaseTransformer
}

// ContentfulEntry represents a Contentful entry
type ContentfulEntry struct {
	Sys    ContentfulSys      `json:"sys"`
	Fields map[string]any     `json:"fields"`
}

type ContentfulSys struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	ContentType ContentfulContentType  `json:"contentType"`
	Space       *ContentfulSpace       `json:"space,omitempty"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
	Revision    int                    `json:"revision,omitempty"`
}

type ContentfulContentType struct {
	Sys ContentfulLink `json:"sys"`
}

type ContentfulLink struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	LinkType string `json:"linkType"`
}

type ContentfulSpace struct {
	Sys ContentfulLink `json:"sys"`
}

type ContentfulAsset struct {
	Sys ContentfulLink `json:"sys"`
}

func (c *ContentfulTransformer) Transform(root model.Root) (any, error) {
	if root.Node == nil {
		return ContentfulEntry{}, nil
	}

	return c.transformNode(root.Node), nil
}

func (c *ContentfulTransformer) TransformToJSON(root model.Root) ([]byte, error) {
	result, err := c.Transform(root)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (c *ContentfulTransformer) transformNode(node *model.Node) ContentfulEntry {
	entry := ContentfulEntry{
		Sys: ContentfulSys{
			ID:   int64ToString(node.Datatype.Content.ContentDataID),
			Type: "Entry",
			ContentType: ContentfulContentType{
				Sys: ContentfulLink{
					ID:       c.toContentfulID(node.Datatype.Info.Label),
					Type:     "Link",
					LinkType: "ContentType",
				},
			},
			CreatedAt: c.getDateCreated(node),
			UpdatedAt: c.getDateModified(node),
			Revision:  1,
		},
		Fields: make(map[string]any),
	}

	// Add space if configured
	if c.SpaceID != "" {
		entry.Sys.Space = &ContentfulSpace{
			Sys: ContentfulLink{
				ID:       c.SpaceID,
				Type:     "Link",
				LinkType: "Space",
			},
		}
	}

	// Transform fields
	for _, field := range node.Fields {
		key := fieldLabelToKey(field.Info.Label.(string))
		value := c.transformField(field)
		entry.Fields[key] = value
	}

	// Transform child nodes
	if len(node.Nodes) > 0 {
		childType := pluralize(node.Nodes[0].Datatype.Info.Label)
		children := make([]ContentfulEntry, 0, len(node.Nodes))

		for _, child := range node.Nodes {
			children = append(children, c.transformNode(child))
		}

		entry.Fields[childType] = children
	}

	return entry
}

func (c *ContentfulTransformer) transformField(field model.Field) any {
	fieldType := field.Info.Type
	value := field.Content.FieldValue

	// Handle asset/image fields
	if fieldType == "image" || fieldType == "asset" {
		return ContentfulAsset{
			Sys: ContentfulLink{
				ID:       c.extractAssetID(value),
				Type:     "Link",
				LinkType: "Asset",
			},
		}
	}

	// Parse other field types
	return c.parseFieldValue(value, fieldType)
}

func (c *ContentfulTransformer) parseFieldValue(value string, fieldType string) any {
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

func (c *ContentfulTransformer) toContentfulID(label string) string {
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
		// Skip spaces and special chars
	}
	return result
}

func (c *ContentfulTransformer) extractAssetID(url string) string {
	// Try to extract ID from URL, otherwise hash it
	hash := fnv.New32a()
	hash.Write([]byte(url))
	return int64ToString(int64(hash.Sum32()))
}

func (c *ContentfulTransformer) getDateCreated(node *model.Node) string {
	if node.Datatype.Content.DateCreated.Valid {
		return node.Datatype.Content.DateCreated.String
	}
	return ""
}

func (c *ContentfulTransformer) getDateModified(node *model.Node) string {
	if node.Datatype.Content.DateModified.Valid {
		return node.Datatype.Content.DateModified.String
	}
	return c.getDateCreated(node)
}

// Helper functions
func int64ToString(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

func parseFloat(s string) float64 {
	var result float64
	var decimal int
	var afterDecimal bool

	for _, char := range s {
		if char >= '0' && char <= '9' {
			if afterDecimal {
				decimal = decimal*10 + int(char-'0')
				result = result + float64(decimal)/float64(pow(10, len(s[len(s)-decimal:])))
			} else {
				result = result*10 + float64(char-'0')
			}
		} else if char == '.' {
			afterDecimal = true
		}
	}

	return result
}

func parseInt(s string) int64 {
	var result int64

	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int64(char-'0')
		}
	}

	return result
}

func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// Parse converts Contentful format to ModulaCMS format (INBOUND)
func (c *ContentfulTransformer) Parse(data []byte) (model.Root, error) {
	var entry ContentfulEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return model.Root{}, err
	}

	node := c.parseEntry(entry)
	return model.Root{Node: node}, nil
}

// ParseToNode converts Contentful entry to ModulaCMS Node (INBOUND)
func (c *ContentfulTransformer) ParseToNode(data []byte) (*model.Node, error) {
	root, err := c.Parse(data)
	if err != nil {
		return nil, err
	}
	return root.Node, nil
}

func (c *ContentfulTransformer) parseEntry(entry ContentfulEntry) *model.Node {
	node := &model.Node{
		Datatype: model.Datatype{
			Info: db.DatatypeJSON{
				DatatypeID: stringToInt64(entry.Sys.ContentType.Sys.ID),
				Label:      entry.Sys.ContentType.Sys.ID,
				Type:       entry.Sys.Type,
			},
			Content: db.ContentDataJSON{
				ContentDataID: stringToInt64(entry.Sys.ID),
				DateCreated:   db.NullString{NullString: sql.NullString{String: entry.Sys.CreatedAt, Valid: entry.Sys.CreatedAt != ""}},
				DateModified:  db.NullString{NullString: sql.NullString{String: entry.Sys.UpdatedAt, Valid: entry.Sys.UpdatedAt != ""}},
			},
		},
		Fields: []model.Field{},
		Nodes:  []*model.Node{},
	}

	// Convert fields
	fieldID := int64(1)
	for key, value := range entry.Fields {
		// Check if it's an asset
		if assetVal, ok := value.(map[string]any); ok {
			if sys, hasSys := assetVal["sys"]; hasSys {
				if sysMap, ok := sys.(map[string]any); ok {
					if linkType, hasType := sysMap["linkType"]; hasType && linkType == "Asset" {
						// Extract asset URL from ID
						assetID := ""
						if id, hasID := sysMap["id"]; hasID {
							assetID = toString(id)
						}
						value = assetID // Store asset ID as value
					}
				}
			}
		}

		// Check if it's a nested entry array (child nodes)
		if arr, isArray := value.([]any); isArray && len(arr) > 0 {
			if firstItem, ok := arr[0].(map[string]any); ok {
				if _, hasSys := firstItem["sys"]; hasSys {
					// This is an array of entries - parse as child nodes
					for _, item := range arr {
						if itemMap, ok := item.(map[string]any); ok {
							itemBytes, _ := json.Marshal(itemMap)
							var childEntry ContentfulEntry
							if json.Unmarshal(itemBytes, &childEntry) == nil {
								node.Nodes = append(node.Nodes, c.parseEntry(childEntry))
							}
						}
					}
					continue
				}
			}
		}

		field := model.Field{
			Info: db.FieldsJSON{
				FieldID: fieldID,
				Label:   c.camelCaseToLabel(key),
				Type:    c.detectFieldType(value),
			},
			Content: db.ContentFieldsJSON{
				ContentFieldID: fieldID,
				FieldID:        fieldID,
				FieldValue:     c.valueToString(value),
			},
		}

		node.Fields = append(node.Fields, field)
		fieldID++
	}

	return node
}

func (c *ContentfulTransformer) camelCaseToLabel(camelCase string) string {
	if camelCase == "" {
		return ""
	}

	result := ""
	for i, char := range camelCase {
		if i == 0 {
			// First character uppercase
			if char >= 'a' && char <= 'z' {
				result += string(char - 32)
			} else {
				result += string(char)
			}
		} else if char >= 'A' && char <= 'Z' {
			// Uppercase letter, add space before it
			result += " " + string(char)
		} else {
			result += string(char)
		}
	}

	return result
}

func (c *ContentfulTransformer) detectFieldType(value any) string {
	switch v := value.(type) {
	case bool:
		return "boolean"
	case float64:
		return "number"
	case int, int64:
		return "integer"
	case map[string]any:
		return "json"
	case []any:
		return "json"
	default:
		_ = v
		return "text"
	}
}

func (c *ContentfulTransformer) valueToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return floatToString(v)
	case int:
		return int64ToString(int64(v))
	case int64:
		return int64ToString(v)
	case map[string]any, []any:
		bytes, _ := json.Marshal(v)
		return string(bytes)
	default:
		return ""
	}
}

func stringToInt64(s string) int64 {
	var result int64
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int64(char-'0')
		}
	}
	return result
}

func floatToString(f float64) string {
	return toString(f)
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return int64ToString(int64(val))
	case int64:
		return int64ToString(val)
	case float64:
		// Simple float to string
		return int64ToString(int64(val))
	default:
		return ""
	}
}
