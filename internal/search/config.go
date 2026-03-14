package search

type SearchConfig struct {
	K1                  float64
	B                   float64
	FieldWeights        map[string]float64
	IndexableFieldTypes map[string]bool
	StopWords           map[string]bool
	MinTermLength       int
	IndexPath           string
	MaxResults          int
	SnippetLength       int
	DefaultLimit        int
}

var DefaultFieldWeights = map[string]float64{
	"_title":           3.0,
	"title":            3.0,
	"name":             2.5,
	"_section_heading": 2.5,
	"slug":             1.5,
	"text":             1.0,
	"textarea":         1.0,
	"richtext":         1.0,
	"_section_body":    1.0,
}

var DefaultIndexableFieldTypes = map[string]bool{
	"text":     true,
	"textarea": true,
	"richtext": true,
	"slug":     true,
	"email":    true,
	"url":      true,
}

var defaultStopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true,
	"at": true, "be": true, "by": true, "for": true, "from": true,
	"has": true, "he": true, "in": true, "is": true, "it": true,
	"its": true, "of": true, "on": true, "or": true, "that": true,
	"the": true, "to": true, "was": true, "were": true, "will": true,
	"with": true, "this": true, "but": true, "not": true, "you": true,
	"all": true, "can": true, "had": true, "her": true, "his": true,
	"one": true, "our": true, "out": true, "do": true,
}

func DefaultConfig() SearchConfig {
	fieldWeights := make(map[string]float64, len(DefaultFieldWeights))
	for k, v := range DefaultFieldWeights {
		fieldWeights[k] = v
	}

	indexableFieldTypes := make(map[string]bool, len(DefaultIndexableFieldTypes))
	for k, v := range DefaultIndexableFieldTypes {
		indexableFieldTypes[k] = v
	}

	stopWords := make(map[string]bool, len(defaultStopWords))
	for k, v := range defaultStopWords {
		stopWords[k] = v
	}

	return SearchConfig{
		K1:                  1.2,
		B:                   0.75,
		FieldWeights:        fieldWeights,
		IndexableFieldTypes: indexableFieldTypes,
		StopWords:           stopWords,
		MinTermLength:       1,
		IndexPath:           "search.idx",
		MaxResults:          100,
		SnippetLength:       200,
		DefaultLimit:        20,
	}
}
