package transform

import (
	"encoding/json"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/model"
)

// Use OutputFormat from config package
type OutputFormat = config.OutputFormat

const (
	FormatContentful = config.FormatContentful
	FormatSanity     = config.FormatSanity
	FormatStrapi     = config.FormatStrapi
	FormatWordPress  = config.FormatWordPress
	FormatClean      = config.FormatClean
	FormatRaw        = config.FormatRaw
)

// Transformer interface for converting ModulaCMS data to different CMS formats
type Transformer interface {
	// Transform converts a ModulaCMS Root/Node to the target CMS format (OUTBOUND)
	Transform(root model.Root) (any, error)

	// TransformToJSON converts and marshals to JSON bytes (OUTBOUND)
	TransformToJSON(root model.Root) ([]byte, error)

	// Parse converts from CMS format to ModulaCMS format (INBOUND)
	Parse(data []byte) (model.Root, error)

	// ParseToNode converts from CMS format to ModulaCMS Node (INBOUND)
	ParseToNode(data []byte) (*model.Node, error)
}

// BaseTransformer provides common functionality for all transformers
type BaseTransformer struct {
	SiteURL string
	SpaceID string
}

// TransformToJSON is a helper that transforms and marshals to JSON
func (b *BaseTransformer) TransformToJSON(transformer Transformer, root model.Root) ([]byte, error) {
	result, err := transformer.Transform(root)
	if err != nil {
		return nil, err
	}

	return json.Marshal(result)
}

// fieldLabelToKey converts field labels to camelCase keys
// "Title" → "title"
// "Featured Image" → "featuredImage"
// "SEO Meta Description" → "seoMetaDescription"
func fieldLabelToKey(label string) string {
	if label == "" {
		return ""
	}

	words := splitWords(label)
	if len(words) == 0 {
		return ""
	}

	result := ""
	for i, word := range words {
		if i == 0 {
			result += toLowerCase(word)
		} else {
			result += toTitleCase(word)
		}
	}

	return result
}

// splitWords splits a string into words
func splitWords(s string) []string {
	words := []string{}
	currentWord := ""

	for _, char := range s {
		if char == ' ' || char == '_' || char == '-' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(char)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// toLowerCase converts string to lowercase
func toLowerCase(s string) string {
	result := ""
	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			result += string(char + 32)
		} else {
			result += string(char)
		}
	}
	return result
}

// toTitleCase converts string to TitleCase
func toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	result := ""
	for i, char := range s {
		if i == 0 {
			if char >= 'a' && char <= 'z' {
				result += string(char - 32)
			} else {
				result += string(char)
			}
		} else {
			if char >= 'A' && char <= 'Z' {
				result += string(char + 32)
			} else {
				result += string(char)
			}
		}
	}
	return result
}

// pluralize converts singular word to plural (simplified)
func pluralize(word string) string {
	if word == "" {
		return ""
	}

	lower := toLowerCase(word)

	// Already plural
	if len(lower) > 0 && lower[len(lower)-1] == 's' {
		return lower
	}

	// Ends with 'y'
	if len(lower) > 1 && lower[len(lower)-1] == 'y' {
		return lower[:len(lower)-1] + "ies"
	}

	// Default: add 's'
	return lower + "s"
}
