package query

import (
	"encoding/json"
)

// QueryTransformer transforms a QueryResult into a wire-ready format.
type QueryTransformer interface {
	TransformResult(result QueryResult) (any, error)
	TransformResultToJSON(result QueryResult) ([]byte, error)
}

// RawQueryTransformer outputs query results with raw string field values.
type RawQueryTransformer struct{}

type rawResultEnvelope struct {
	Data     []rawItem   `json:"data"`
	Total    int64       `json:"total"`
	Limit    int64       `json:"limit"`
	Offset   int64       `json:"offset"`
	Datatype rawDatatype `json:"datatype"`
}

type rawItem struct {
	ContentDataID string            `json:"content_data_id"`
	DatatypeID    string            `json:"datatype_id"`
	AuthorID      string            `json:"author_id"`
	Status        string            `json:"status"`
	DateCreated   string            `json:"date_created"`
	DateModified  string            `json:"date_modified"`
	PublishedAt   string            `json:"published_at"`
	Fields        map[string]string `json:"fields"`
}

type rawDatatype struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

// TransformResult converts a QueryResult into the raw envelope structure.
func (t *RawQueryTransformer) TransformResult(result QueryResult) (any, error) {
	return buildRawEnvelope(result), nil
}

// TransformResultToJSON marshals a QueryResult into JSON bytes.
func (t *RawQueryTransformer) TransformResultToJSON(result QueryResult) ([]byte, error) {
	env := buildRawEnvelope(result)
	return json.Marshal(env)
}

func buildRawEnvelope(result QueryResult) rawResultEnvelope {
	items := make([]rawItem, 0, len(result.Items))
	for _, qi := range result.Items {
		items = append(items, toRawItem(qi))
	}
	return rawResultEnvelope{
		Data:   items,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
		Datatype: rawDatatype{
			Name:  result.Datatype.Name,
			Label: result.Datatype.Label,
		},
	}
}

func toRawItem(qi QueryItem) rawItem {
	cd := qi.ContentData
	return rawItem{
		ContentDataID: cd.ContentDataID.String(),
		DatatypeID:    cd.DatatypeID.String(),
		AuthorID:      cd.AuthorID.String(),
		Status:        string(cd.Status),
		DateCreated:   cd.DateCreated.String(),
		DateModified:  cd.DateModified.String(),
		PublishedAt:   cd.PublishedAt.String(),
		Fields:        copyFields(qi.Fields),
	}
}

func copyFields(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// DefaultTransformer returns the default transformer (raw).
func DefaultTransformer() QueryTransformer {
	return &RawQueryTransformer{}
}
