package search

type SearchDocument struct {
	ID            string
	ContentDataID string
	RouteSlug     string
	RouteTitle    string
	DatatypeName  string
	DatatypeLabel string
	Locale        string
	Section       string
	SectionAnchor string
	Fields        map[string]string
	PublishedAt   string
	AuthorID      string
}

type SearchResult struct {
	ID            string  `json:"id"`
	ContentDataID string  `json:"content_data_id"`
	RouteSlug     string  `json:"route_slug"`
	RouteTitle    string  `json:"route_title"`
	DatatypeName  string  `json:"datatype_name"`
	DatatypeLabel string  `json:"datatype_label"`
	Locale        string  `json:"locale,omitempty"`
	Section       string  `json:"section,omitempty"`
	SectionAnchor string  `json:"section_anchor,omitempty"`
	Score         float64 `json:"score"`
	Snippet       string  `json:"snippet"`
	PublishedAt   string  `json:"published_at"`
}

type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

type SearchOptions struct {
	Limit        int
	Offset       int
	DatatypeName string
	Locale       string
}

type IndexStats struct {
	Documents   int   `json:"documents"`
	Terms       int   `json:"terms"`
	Postings    int   `json:"postings"`
	Fields      int   `json:"fields"`
	MemEstimate int64 `json:"mem_bytes"`
}

type Section struct {
	Heading string
	Anchor  string
	Body    string
}
