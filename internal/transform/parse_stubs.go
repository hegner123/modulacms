package transform

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/model"
)

// Parse stubs for transformers not yet implemented for inbound transformation
// TODO: Implement full Parse() methods for these transformers

// Sanity Parse (stub - to be implemented)
func (s *SanityTransformer) Parse(data []byte) (model.Root, error) {
	return model.Root{}, fmt.Errorf("Sanity inbound parsing not yet implemented - use format='clean' or 'raw' for POST/PUT requests")
}

func (s *SanityTransformer) ParseToNode(data []byte) (*model.Node, error) {
	return nil, fmt.Errorf("Sanity inbound parsing not yet implemented - use format='clean' or 'raw' for POST/PUT requests")
}

// Strapi Parse (stub - to be implemented)
func (s *StrapiTransformer) Parse(data []byte) (model.Root, error) {
	return model.Root{}, fmt.Errorf("Strapi inbound parsing not yet implemented - use format='clean' or 'raw' for POST/PUT requests")
}

func (s *StrapiTransformer) ParseToNode(data []byte) (*model.Node, error) {
	return nil, fmt.Errorf("Strapi inbound parsing not yet implemented - use format='clean' or 'raw' for POST/PUT requests")
}

// WordPress Parse (stub - to be implemented)
func (w *WordPressTransformer) Parse(data []byte) (model.Root, error) {
	return model.Root{}, fmt.Errorf("WordPress inbound parsing not yet implemented - use format='clean' or 'raw' for POST/PUT requests")
}

func (w *WordPressTransformer) ParseToNode(data []byte) (*model.Node, error) {
	return nil, fmt.Errorf("WordPress inbound parsing not yet implemented - use format='clean' or 'raw' for POST/PUT requests")
}
