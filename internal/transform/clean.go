package transform

import (
	"encoding/json"
	"strconv"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

// CleanTransformer transforms ModulaCMS data to clean, flat format
type CleanTransformer struct {
	BaseTransformer
}

// CleanDocument represents a clean ModulaCMS document
type CleanDocument struct {
	ID   string                 `json:"id"`
	Type string                 `json:"type"`
	Meta CleanMeta              `json:"_meta"`
	Data map[string]any         `json:"-"` // Flattened fields
}

type CleanMeta struct {
	AuthorID     string `json:"authorId"`
	RouteID      string `json:"routeId"`
	DateCreated  string `json:"dateCreated,omitempty"`
	DateModified string `json:"dateModified,omitempty"`
}

func (c *CleanTransformer) Transform(root model.Root) (any, error) {
	if root.Node == nil {
		return map[string]any{}, nil
	}

	return c.transformNode(root.Node), nil
}

func (c *CleanTransformer) TransformToJSON(root model.Root) ([]byte, error) {
	result, err := c.Transform(root)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

func (c *CleanTransformer) transformNode(node *model.Node) map[string]any {
	doc := make(map[string]any)

	// Add ID and type
	doc["id"] = node.Datatype.Content.ContentDataID
	doc["type"] = node.Datatype.Info.Label

	// Add metadata
	meta := CleanMeta{
		AuthorID: node.Datatype.Content.AuthorID,
		RouteID:  node.Datatype.Content.RouteID,
	}

	if node.Datatype.Content.DateCreated != "" {
		meta.DateCreated = node.Datatype.Content.DateCreated
	}
	if node.Datatype.Content.DateModified != "" {
		meta.DateModified = node.Datatype.Content.DateModified
	}

	doc["_meta"] = meta

	// Transform fields to flat properties
	for _, field := range node.Fields {
		key := fieldLabelToKey(field.Info.Label)
		value := c.parseFieldValue(field.Content.FieldValue, field.Info.Type)
		doc[key] = value
	}

	// Transform child nodes to arrays
	if len(node.Nodes) > 0 {
		children := make([]map[string]any, 0, len(node.Nodes))
		for _, child := range node.Nodes {
			children = append(children, c.transformNode(child))
		}

		if allChildrenSameType(node.Nodes) {
			doc[pluralize(node.Nodes[0].Datatype.Info.Label)] = children
		} else {
			doc["children"] = children
		}
	}

	return doc
}

func (c *CleanTransformer) parseFieldValue(value string, fieldType string) any {
	switch fieldType {
	case "boolean":
		return value == "true"
	case "number", "decimal", "float":
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
		return value
	case "integer", "int":
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
		return value
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

// Parse converts Clean format to ModulaCMS format (INBOUND)
func (c *CleanTransformer) Parse(data []byte) (model.Root, error) {
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return model.Root{}, err
	}

	node := c.parseDocument(doc)
	return model.Root{Node: node}, nil
}

// ParseToNode converts Clean document to ModulaCMS Node (INBOUND)
func (c *CleanTransformer) ParseToNode(data []byte) (*model.Node, error) {
	root, err := c.Parse(data)
	if err != nil {
		return nil, err
	}
	return root.Node, nil
}

func (c *CleanTransformer) parseDocument(doc map[string]any) *model.Node {
	node := &model.Node{
		Datatype: model.Datatype{
			Info: db.DatatypeJSON{},
			Content: db.ContentDataJSON{},
		},
		Fields: []model.Field{},
		Nodes:  []*model.Node{},
	}

	// Extract ID
	if id, ok := doc["id"]; ok {
		if idStr, ok := id.(string); ok {
			node.Datatype.Content.ContentDataID = idStr
		} else if idFloat, ok := id.(float64); ok {
			node.Datatype.Content.ContentDataID = strconv.FormatInt(int64(idFloat), 10)
		} else if idInt, ok := id.(int); ok {
			node.Datatype.Content.ContentDataID = strconv.FormatInt(int64(idInt), 10)
		}
	}

	// Extract type
	if typeVal, ok := doc["type"]; ok {
		if typeStr, ok := typeVal.(string); ok {
			node.Datatype.Info.Label = typeStr
			node.Datatype.Info.Type = toLowerCase(typeStr)
		}
	}

	// Extract metadata
	if meta, ok := doc["_meta"]; ok {
		if metaMap, ok := meta.(map[string]any); ok {
			if authorID, ok := metaMap["authorId"]; ok {
				if authorStr, ok := authorID.(string); ok {
					node.Datatype.Content.AuthorID = authorStr
				} else if authorFloat, ok := authorID.(float64); ok {
					node.Datatype.Content.AuthorID = strconv.FormatInt(int64(authorFloat), 10)
				}
			}
			if routeID, ok := metaMap["routeId"]; ok {
				if routeStr, ok := routeID.(string); ok {
					node.Datatype.Content.RouteID = routeStr
				} else if routeFloat, ok := routeID.(float64); ok {
					node.Datatype.Content.RouteID = strconv.FormatInt(int64(routeFloat), 10)
				}
			}
			if dateCreated, ok := metaMap["dateCreated"]; ok {
				if dateStr, ok := dateCreated.(string); ok {
					node.Datatype.Content.DateCreated = dateStr
				}
			}
			if dateModified, ok := metaMap["dateModified"]; ok {
				if dateStr, ok := dateModified.(string); ok {
					node.Datatype.Content.DateModified = dateStr
				}
			}
		}
	}

	// Extract fields
	fieldID := int64(1)
	for key, value := range doc {
		// Skip system fields
		if key == "id" || key == "type" || key == "_meta" {
			continue
		}

		// Check if it's a child nodes array
		if arr, isArray := value.([]any); isArray && len(arr) > 0 {
			// Check if first item looks like a node
			if firstItem, ok := arr[0].(map[string]any); ok {
				if _, hasID := firstItem["id"]; hasID {
					// Parse as child nodes
					for _, item := range arr {
						if itemMap, ok := item.(map[string]any); ok {
							node.Nodes = append(node.Nodes, c.parseDocument(itemMap))
						}
					}
					continue
				}
			}
		}

		// It's a regular field
		fieldIDStr := strconv.FormatInt(fieldID, 10)
		field := model.Field{
			Info: db.FieldsJSON{
				FieldID: fieldIDStr,
				Label:   c.keyToLabel(key),
				Type:    c.detectFieldType(value),
			},
			Content: db.ContentFieldsJSON{
				ContentFieldID: fieldID,
				FieldID:        fieldID,
				FieldValue:     c.anyToString(value),
			},
		}

		node.Fields = append(node.Fields, field)
		fieldID++
	}

	return node
}

func (c *CleanTransformer) keyToLabel(key string) string {
	// Convert camelCase to "Title Case"
	if key == "" {
		return ""
	}

	result := ""
	for i, char := range key {
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

func (c *CleanTransformer) detectFieldType(value any) string {
	switch value.(type) {
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
		return "text"
	}
}

func (c *CleanTransformer) anyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case map[string]any, []any:
		bytes, _ := json.Marshal(v)
		return string(bytes)
	default:
		return ""
	}
}
