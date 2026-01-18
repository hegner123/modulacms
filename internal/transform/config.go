package transform

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/model"
)

// TransformConfig holds configuration for transformation
type TransformConfig struct {
	Format  OutputFormat
	SiteURL string
	SpaceID string
	Driver  db.DbDriver
}

// NewTransformConfig creates a new TransformConfig from config values
func NewTransformConfig(format OutputFormat, siteURL string, spaceID string, driver db.DbDriver) *TransformConfig {
	return &TransformConfig{
		Format:  format,
		SiteURL: siteURL,
		SpaceID: spaceID,
		Driver:  driver,
	}
}

// NewTransformConfigFromString creates a new TransformConfig from a string format
func NewTransformConfigFromString(format string, siteURL string, spaceID string, driver db.DbDriver) *TransformConfig {
	return &TransformConfig{
		Format:  OutputFormat(format),
		SiteURL: siteURL,
		SpaceID: spaceID,
		Driver:  driver,
	}
}

// GetTransformer returns the appropriate transformer based on config
func (tc *TransformConfig) GetTransformer() (Transformer, error) {
	base := BaseTransformer{
		SiteURL: tc.SiteURL,
		SpaceID: tc.SpaceID,
	}

	switch tc.Format {
	case FormatContentful:
		return &ContentfulTransformer{BaseTransformer: base}, nil
	case FormatSanity:
		return &SanityTransformer{BaseTransformer: base}, nil
	case FormatStrapi:
		return &StrapiTransformer{BaseTransformer: base}, nil
	case FormatWordPress:
		return &WordPressTransformer{BaseTransformer: base}, nil
	case FormatClean:
		return &CleanTransformer{BaseTransformer: base}, nil
	case FormatRaw, "":
		// No transformation - return raw ModulaCMS format
		return &RawTransformer{}, nil
	default:
		return nil, fmt.Errorf("unknown output format: %s", tc.Format)
	}
}

// TransformAndWrite transforms data and writes JSON response (OUTBOUND)
func (tc *TransformConfig) TransformAndWrite(w http.ResponseWriter, root model.Root) error {
	transformer, err := tc.GetTransformer()
	if err != nil {
		return err
	}

	data, err := transformer.TransformToJSON(root)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write(data)
	return writeErr
}

// ParseRequest parses incoming request body from CMS format to ModulaCMS (INBOUND)
func (tc *TransformConfig) ParseRequest(r *http.Request) (model.Root, error) {
	transformer, err := tc.GetTransformer()
	if err != nil {
		return model.Root{}, err
	}

	// Read request body
	body := make([]byte, r.ContentLength)
	_, err = r.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		return model.Root{}, err
	}
	defer r.Body.Close()

	// Parse from CMS format to ModulaCMS
	return transformer.Parse(body)
}

// RawTransformer returns ModulaCMS data without transformation
type RawTransformer struct{}

func (r *RawTransformer) Transform(root model.Root) (any, error) {
	return root, nil
}

func (r *RawTransformer) TransformToJSON(root model.Root) ([]byte, error) {
	return json.Marshal(root)
}

func (r *RawTransformer) Parse(data []byte) (model.Root, error) {
	var root model.Root
	err := json.Unmarshal(data, &root)
	return root, err
}

func (r *RawTransformer) ParseToNode(data []byte) (*model.Node, error) {
	var root model.Root
	err := json.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}
	return root.Node, nil
}
