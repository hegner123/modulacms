# Search Package Plan

**Status:** Draft
**Created:** 2026-03-10
**Location:** `internal/search/`
**Dependencies:** None (stdlib only)

---

## Design

A dependency-free inverted index with BM25 ranking, prefix matching, field weighting, section-level indexing, and snippet extraction. The index lives in memory, persists to a single file on disk, and updates incrementally on publish/unpublish events.

**Data source:** Published content version snapshots (`content_versions WHERE published = 1` and `admin_content_versions WHERE published = 1`). Each snapshot is a complete denormalized JSON blob containing all field values, datatypes, and tree structure. No additional DB queries needed to build search documents.

**Dual content systems:** ModulaCMS has two independent content systems with identical schema: public tables (site content served to frontends) and admin tables (content powering the admin panel UI). These are not a draft/published pipeline — they are parallel systems. The search package indexes both independently. Both have their own publish flow, versioning, snapshots, routes (with slugs), and tree structure.

**Indexing model:** Each published snapshot produces one or more search documents. Text-heavy fields (richtext, textarea) are split by headings into section-level documents so search results can link to `#heading-id` anchors.

---

## Package Structure

```
internal/search/
  index.go       Core inverted index: postings lists, add/remove documents
  tokenizer.go   Text splitting, normalization, stop word removal
  scoring.go     BM25 ranking + field weight aggregation
  document.go    SearchDocument type, field extraction from snapshots
  builder.go     Build SearchDocuments from publishing.Snapshot
  storage.go     Gob serialize/deserialize index to disk file
  search.go      Query parsing, execution, result construction
  snippet.go     Extract matching text windows with term positions
  config.go      SearchConfig (field weights, stop words, limits)
```

**Not in this package:**
- HTTP handler → `internal/router/search.go`
- Publish hook → `internal/publishing/publishing.go` (3-line addition)
- Config fields → `internal/config/config.go`
- Startup wiring → `cmd/serve.go`

---

## Data Structures

### SearchDocument

The unit of indexing. One published content tree produces one or more documents.

```go
type SearchDocument struct {
    // Identity
    ID            string // "{content_data_id}" or "{content_data_id}#{anchor}"
    ContentDataID string // always the root content_data_id
    RouteSlug     string // URL path: "/blog/my-post"
    RouteTitle    string // route display name

    // Classification
    DatatypeName  string // "blog_post", "page", "doc_page"
    DatatypeLabel string // "Blog Post", "Page"
    Locale        string // "" or "en", "fr", etc.

    // Section (if split by heading)
    Section       string // heading text: "Installation" (empty for root doc)
    SectionAnchor string // heading slug: "installation" (empty for root doc)

    // Searchable content — field name → text value
    // Only text-bearing field types indexed: text, textarea, richtext, slug, email, url, _title
    Fields map[string]string

    // Metadata (not indexed, returned in results)
    PublishedAt string // RFC3339
    AuthorID    string
}
```

### Posting

A single occurrence of a term in a document field.

```go
type Posting struct {
    DocIdx   uint32 // index into Index.docs slice
    FieldIdx uint8  // index into Index.fieldNames slice (compact)
    Position uint16 // word position within the field (for phrase/proximity)
}
```

Using fixed-width integers keeps memory tight. `uint16` position supports up to 65K tokens per field — sufficient for any CMS document.

### Index

The core inverted index. Concurrent-safe via `sync.RWMutex`.

```go
type Index struct {
    mu sync.RWMutex

    // Document storage
    docs []SearchDocument // indexed by DocIdx

    // Lookup maps
    docsByContentID map[string][]int // content_data_id → doc indices (one root + N sections)

    // Inverted index
    postings map[string][]Posting // normalized term → sorted posting list

    // Prefix search support
    sortedTerms []string // sorted copy of postings keys, rebuilt on mutation

    // BM25 precomputed stats
    docCount     int
    fieldLengths []map[uint8]int // per-doc field token counts (index = DocIdx)
    avgFieldLen  map[uint8]float64 // average tokens per field across all docs

    // Configuration
    fieldNames   []string           // ordered field name registry (index = FieldIdx)
    fieldNameIdx map[string]uint8   // field name → FieldIdx
    config       SearchConfig
}
```

### SearchConfig

```go
type SearchConfig struct {
    // BM25 tuning
    K1 float64 // term saturation (default 1.2)
    B  float64 // length normalization (default 0.75)

    // Field weights — higher = more important in ranking
    // Missing fields default to 1.0
    FieldWeights map[string]float64

    // Indexing control
    IndexableFieldTypes map[string]bool // field types to index (text, textarea, richtext, etc.)
    StopWords           map[string]bool // common words to skip ("the", "a", "is", etc.)
    MinTermLength       int             // skip terms shorter than this (default 1)

    // Storage
    IndexPath string // file path for persistence (default: alongside DB, "search.idx")

    // Query limits
    MaxResults     int // hard cap on results per query (default 100)
    SnippetLength  int // max characters per snippet (default 200)
    DefaultLimit   int // default results per query (default 20)
}
```

Default field weights:
```go
var DefaultFieldWeights = map[string]float64{
    "_title":   3.0, // title fields rank highest
    "title":    3.0,
    "name":     2.5,
    "slug":     1.5,
    "text":     1.0,
    "textarea": 1.0,
    "richtext": 1.0,
}
```

### SearchResult

Returned to the caller.

```go
type SearchResult struct {
    ID            string  `json:"id"`              // document ID
    ContentDataID string  `json:"content_data_id"`
    RouteSlug     string  `json:"route_slug"`
    RouteTitle    string  `json:"route_title"`
    DatatypeName  string  `json:"datatype_name"`
    DatatypeLabel string  `json:"datatype_label"`
    Locale        string  `json:"locale,omitempty"`
    Section       string  `json:"section,omitempty"`       // heading text
    SectionAnchor string  `json:"section_anchor,omitempty"` // #anchor
    Score         float64 `json:"score"`
    Snippet       string  `json:"snippet"`          // matching text with context
    PublishedAt   string  `json:"published_at"`
}

type SearchResponse struct {
    Query   string         `json:"query"`
    Results []SearchResult `json:"results"`
    Total   int            `json:"total"`   // total matches (before limit)
    Limit   int            `json:"limit"`
    Offset  int            `json:"offset"`
}
```

---

## Phase 1: Tokenizer

**File:** `internal/search/tokenizer.go`

### 1.1 Tokenize function

```go
// Tokenize splits text into normalized terms.
// Returns slice of terms in order (positions preserved by index).
func Tokenize(text string) []string
```

Rules:
1. Convert to lowercase (ASCII-aware, `unicode.ToLower` for non-ASCII)
2. Replace HTML tags with spaces (strip `<...>` — richtext fields contain HTML)
3. Split on non-letter, non-digit characters (whitespace, punctuation, symbols)
4. Discard empty strings
5. No stemming in v1 (add later if needed)

### 1.2 HTML stripping

```go
// StripHTML removes HTML tags and decodes common entities.
// Does NOT use regex — character-by-character state machine.
func StripHTML(html string) string
```

Two states: `inTag` and `inText`. Walk runes:
- `<` → enter `inTag`
- `>` → exit `inTag`, emit space
- `&amp;` → `&`, `&lt;` → `<`, `&gt;` → `>`, `&quot;` → `"`, `&#39;` → `'`
- Everything else in `inText` → emit as-is

### 1.3 Stop words

English stop words hardcoded (30-50 words):
```go
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
```

Stop words are removed during indexing but NOT during querying (a query for "to be or not to be" should still find documents containing those words — the posting lists just won't have entries for stop words alone).

### 1.4 TokenizeAndFilter

```go
// TokenizeAndFilter tokenizes text and removes stop words.
// Returns (terms, positions) where positions[i] is the original
// word position of terms[i] (accounting for removed stop words).
func TokenizeAndFilter(text string, stopWords map[string]bool) (terms []string, positions []int)
```

---

## Phase 2: Core Index

**File:** `internal/search/index.go`

### 2.1 Constructor

```go
func NewIndex(cfg SearchConfig) *Index
```

Initializes empty index with config. Sets up `fieldNames`, `fieldNameIdx`, `postings` maps.

### 2.2 Add document

```go
func (idx *Index) Add(doc SearchDocument)
```

1. Acquire write lock
2. Append doc to `idx.docs`, get `docIdx`
3. Register in `idx.docsByContentID`
4. For each field in `doc.Fields`:
   a. Resolve `fieldIdx` from `idx.fieldNameIdx` (register new field names as needed)
   b. Tokenize field value: `terms, positions := TokenizeAndFilter(StripHTML(value), cfg.StopWords)`
   c. Count term frequencies: `tf := map[string]int`
   d. For each unique term, append `Posting{DocIdx: docIdx, FieldIdx: fieldIdx, Position: pos}` to `idx.postings[term]`
   e. Record field length: `idx.fieldLengths[docIdx][fieldIdx] = len(terms)`
5. Mark `sortedTerms` as stale (rebuild lazily on next prefix search)
6. Recalculate `avgFieldLen`

### 2.3 Remove document(s) by content_data_id

```go
func (idx *Index) RemoveByContentID(contentDataID string)
```

1. Acquire write lock
2. Look up doc indices from `idx.docsByContentID`
3. Mark those indices as tombstoned (set a `deleted` flag or nil the doc)
4. Remove postings that reference tombstoned doc indices
5. Periodically compact (rebuild index if tombstone ratio > 25%)

**Simpler alternative for v1:** Since the index is fully rebuildable and the corpus is small, just rebuild from scratch on remove. Remove is rare (unpublish) and rebuild is fast for CMS-scale data.

### 2.4 Document count

```go
func (idx *Index) Len() int // active (non-tombstoned) document count
```

---

## Phase 3: Scoring

**File:** `internal/search/scoring.go`

### 3.1 BM25

```go
// BM25 computes the BM25 score for a single term in a single field of a single document.
func BM25(tf float64, df float64, docLen float64, avgDocLen float64, totalDocs int, k1 float64, b float64) float64 {
    idf := math.Log((float64(totalDocs)-df+0.5)/(df+0.5) + 1.0)
    tfNorm := (tf * (k1 + 1.0)) / (tf + k1*(1.0-b+b*(docLen/avgDocLen)))
    return idf * tfNorm
}
```

### 3.2 Multi-field scoring

```go
// ScoreDocument computes the aggregate score for a document across all query terms and fields.
func ScoreDocument(idx *Index, docIdx int, queryTerms []string, termDFs map[string]int) float64
```

For each query term:
1. Look up postings for this term
2. Filter to postings matching `docIdx`
3. Group by field
4. For each field: compute `BM25(tf, df, fieldLen, avgFieldLen, totalDocs, k1, b)`
5. Multiply by field weight: `score *= idx.config.FieldWeights[fieldName]`
6. Sum across all fields and terms

### 3.3 Phrase proximity bonus

When multiple query terms appear near each other in the same field, boost the score:

```go
// ProximityBonus returns a multiplier based on how close query terms are
// to each other within a field. Adjacent terms (phrase match) get the
// highest bonus.
func ProximityBonus(positions [][]int, windowSize int) float64
```

- Exact phrase (adjacent positions): 2.0x multiplier
- Within 3 words: 1.5x
- Within 10 words: 1.2x
- Otherwise: 1.0x (no bonus)

---

## Phase 4: Document Builder

**File:** `internal/search/builder.go`

### 4.1 Build from snapshot

```go
// BuildDocuments extracts SearchDocuments from a published content version snapshot.
// Returns one root document plus zero or more section documents (split by headings).
func BuildDocuments(snapshot *publishing.Snapshot, version db.ContentVersion) []SearchDocument
```

**Algorithm:**

1. Extract route info: `slug`, `title` from `snapshot.Route`
2. Build field lookup: `fieldsByID := map[fieldID]FieldsJSON` from `snapshot.Fields`
3. Build datatype lookup: `datatypesByID := map[datatypeID]DatatypeJSON` from `snapshot.Datatypes`
4. Find root content node (the one with `ParentID == ""`)
5. Get root's datatype name and label
6. Collect all text-bearing field values for the root content node:
   - Iterate `snapshot.ContentFields`
   - Filter to fields whose type is in `config.IndexableFieldTypes`
   - Map: `fieldName → fieldValue`
7. Create root `SearchDocument` with all collected fields
8. For richtext fields: call `SplitByHeadings(fieldValue)` to produce section documents
9. For child content nodes in the tree: recursively collect their field values into the root document's fields (prefixed by datatype: `"{datatypeName}.{fieldName}"`)

### 4.2 Default indexable field types

```go
var DefaultIndexableFieldTypes = map[string]bool{
    "text":     true,
    "textarea": true,
    "richtext": true,
    "slug":     true,
    "email":    true,
    "url":      true,
    "_title":   true,
}
```

Fields of type `number`, `date`, `datetime`, `boolean`, `select`, `media`, `relation`, `json`, `_id` are not indexed for full-text search.

### 4.3 Section splitting for richtext

```go
// SplitByHeadings splits HTML content by <h1>-<h6> tags into sections.
// Returns slice of sections, each with heading text, anchor slug, and body text.
func SplitByHeadings(html string) []Section

type Section struct {
    Heading string // "Installation Guide"
    Anchor  string // "installation-guide"
    Body    string // text after heading until next heading (HTML stripped)
}
```

**Algorithm:**
1. Scan for `<h[1-6]` tags (character-by-character, no regex)
2. Extract heading text (strip inner HTML)
3. Generate anchor slug: lowercase, replace spaces/special chars with hyphens, dedup hyphens
4. Collect body text until next heading or end
5. Strip HTML from body text

Each section becomes a separate `SearchDocument` with:
- `ID`: `"{contentDataID}#{anchor}"`
- `Section`: heading text
- `SectionAnchor`: anchor slug
- `Fields`: `{"_section_heading": heading, "_section_body": body}`

Section heading field gets weight 2.5 (between title and body).

### 4.4 Build from admin snapshot

```go
// BuildAdminDocuments extracts SearchDocuments from an admin content version snapshot.
func BuildAdminDocuments(snapshot *publishing.AdminSnapshot, version db.AdminContentVersion) []SearchDocument
```

Same logic using admin types (`AdminContentDataJSON`, `AdminSnapshotContentFieldJSON`, etc.).

**Prerequisite:** `AdminSnapshotRoute` currently only stores `AdminRouteID`. The `admin_routes` table has `slug` and `title` columns, but `AdminSnapshotRoute` doesn't capture them. Before implementing search, add `Slug` and `Title` fields to `AdminSnapshotRoute` (matching `SnapshotRoute`) and update the admin publish flow in `internal/publishing/publishing.go` to populate them. This is a ~5-line change in the snapshot builder.

---

## Phase 5: Search Execution

**File:** `internal/search/search.go`

### 5.1 Query parsing

```go
// ParseQuery splits a search query into terms and optional phrases.
// Quoted strings are treated as phrases: "machine learning" → phrase match.
// Unquoted words are individual terms: machine learning → term match.
func ParseQuery(query string) ParsedQuery

type ParsedQuery struct {
    Terms   []string   // individual terms (lowercased)
    Phrases [][]string // quoted phrases as term slices
}
```

Parsing rules:
1. Split by whitespace
2. If a token starts with `"`, accumulate tokens until closing `"`
3. Lowercase all terms
4. Remove stop words from individual terms (but NOT from phrases)

### 5.2 Search execution

```go
// Search executes a search query against the index and returns ranked results.
func (idx *Index) Search(query string, opts SearchOptions) SearchResponse

type SearchOptions struct {
    Limit        int    // max results (default: config.DefaultLimit)
    Offset       int    // pagination offset
    DatatypeName string // filter to specific datatype (optional)
    Locale       string // filter to specific locale (optional)
}
```

**Algorithm:**

1. Parse query: `parsed := ParseQuery(query)`
2. Acquire read lock
3. For each term in `parsed.Terms`:
   a. Look up `idx.postings[term]` → posting list
   b. Collect candidate doc indices
4. For phrase terms: intersect posting lists and verify adjacency via positions
5. Union all candidate doc indices
6. Apply filters (datatype, locale) — skip docs that don't match
7. Score each candidate: `ScoreDocument(idx, docIdx, terms, termDFs)`
8. Sort by score descending
9. Apply offset/limit
10. Build `SearchResult` for each result (including snippet via Phase 6)
11. Return `SearchResponse`

### 5.3 Prefix matching

```go
// SearchPrefix finds all terms in the index that start with the given prefix.
// Used for search-as-you-type: user types "conf" → matches "configuration", "config", "configure".
func (idx *Index) SearchPrefix(prefix string) []string
```

Implementation:
1. Ensure `idx.sortedTerms` is up to date (lazy rebuild after mutations)
2. Binary search for insertion point of `prefix`
3. Scan forward collecting terms that start with `prefix`
4. Return up to 20 matching terms

For search-as-you-type, the last term in the query is treated as a prefix:
```
query = "install conf"
→ terms = ["install"]
→ prefix = "conf"
→ expand "conf" to all matching terms, score each variant, merge results
```

### 5.4 Search-as-you-type

```go
// SearchWithPrefix handles queries where the last term may be a partial word.
// The last term is expanded via prefix matching; all other terms are exact.
func (idx *Index) SearchWithPrefix(query string, opts SearchOptions) SearchResponse
```

Algorithm:
1. Parse query
2. Pop last term as prefix
3. Find all terms matching prefix (up to 20)
4. For each expansion: score as if the full term was used
5. Take best score per document across all expansions
6. Merge and rank

---

## Phase 6: Snippets

**File:** `internal/search/snippet.go`

### 6.1 Extract snippet

```go
// ExtractSnippet finds the best matching window of text from a document's fields,
// centered on query term occurrences.
func ExtractSnippet(doc SearchDocument, queryTerms []string, maxLen int) string
```

**Algorithm:**

1. Concatenate all text fields (strip HTML) with field-separator markers
2. Tokenize into terms with character offsets (not just positions)
3. Find windows that contain the most query terms
4. Score each window by: (number of query terms present) × (closeness of terms)
5. Select best window
6. Extract substring at character offsets, expanding to word boundaries
7. Add `...` prefix/suffix if truncated
8. Return plain text snippet

### 6.2 Character-offset tokenizer

```go
// TokenizeWithOffsets returns terms with their byte offsets in the original text.
type TermOffset struct {
    Term  string
    Start int // byte offset in original text
    End   int
}

func TokenizeWithOffsets(text string) []TermOffset
```

---

## Phase 7: Persistence

**File:** `internal/search/storage.go`

### 7.1 Save to disk

```go
// Save serializes the index to a file using encoding/gob.
func (idx *Index) Save(path string) error
```

Writes:
1. A header: magic bytes `MCMS` + version uint32 (for forward compatibility)
2. The serializable subset of Index fields (docs, postings, field metadata, stats)
3. Flush and sync

### 7.2 Load from disk

```go
// Load deserializes an index from a file.
// Returns a new Index (does not modify receiver).
func Load(path string, cfg SearchConfig) (*Index, error)
```

1. Read header, verify magic bytes and version
2. Decode gob data
3. Rebuild computed fields (`sortedTerms`, lookup maps)
4. Return ready-to-use index

### 7.3 Rebuild from database

```go
// Rebuild creates a fresh index from all published content versions.
// This is the authoritative rebuild — the persisted file is a cache.
func Rebuild(driver db.DbDriver, cfg SearchConfig) (*Index, error)
```

1. Create empty index
2. Query `content_versions WHERE published = 1` (all locales)
3. For each version:
   a. Deserialize snapshot JSON
   b. Call `BuildDocuments(snapshot, version)`
   c. Add each document to index
4. Same for admin content versions
5. Save to disk
6. Return index

---

## Phase 8: Integration

### 8.1 Publish hook

**File:** `internal/publishing/publishing.go`

Add a `SearchIndexer` interface parameter or callback to `PublishContent()`:

```go
// SearchIndexer is called after successful publish to update the search index.
type SearchIndexer interface {
    OnPublish(snapshot *Snapshot, version db.ContentVersion)
    OnUnpublish(contentDataID string)
}
```

In `PublishContent()`, after step 6 (CreateContentVersion with published=true):

```go
// Update search index (non-blocking)
if indexer != nil {
    go func() {
        indexer.OnPublish(&snapshot, version)
    }()
}
```

The goroutine is fire-and-forget — search index lag is acceptable (milliseconds). The index is not on the critical publish path.

**Same pattern for unpublish** — call `indexer.OnUnpublish(contentDataID)` which removes documents from the index.

### 8.2 Index service

**File:** `internal/search/service.go`

```go
// Service wraps the Index with lifecycle management.
type Service struct {
    index    *Index
    config   SearchConfig
    driver   db.DbDriver
    savePath string
}

func NewService(driver db.DbDriver, cfg SearchConfig) *Service

// Start loads the index from disk (if exists) or rebuilds from DB.
func (s *Service) Start(ctx context.Context) error

// OnPublish implements SearchIndexer.
func (s *Service) OnPublish(snapshot *publishing.Snapshot, version db.ContentVersion)

// OnUnpublish implements SearchIndexer.
func (s *Service) OnUnpublish(contentDataID string)

// Search executes a query.
func (s *Service) Search(query string, opts SearchOptions) SearchResponse

// SearchPrefix executes a prefix query (search-as-you-type).
func (s *Service) SearchPrefix(query string, opts SearchOptions) SearchResponse

// Rebuild forces a full rebuild from the database.
func (s *Service) Rebuild() error
```

`OnPublish` implementation:
1. Build documents from snapshot
2. Remove any existing documents for this content_data_id
3. Add new documents
4. Async save to disk (debounced — don't save on every publish, batch within 5 seconds)

### 8.3 Config fields

**File:** `internal/config/config.go`

```go
// Search
Search_Enabled bool    `json:"search_enabled"` // default: false
Search_Path    string  `json:"search_path"`    // default: "search.idx" (relative to DB)
```

When `search_enabled` is false, no index is created, no hooks fire, the search endpoint returns 404.

### 8.4 Startup wiring

**File:** `cmd/serve.go`

After service registry creation (~line 272):

```go
// Search index
var searchSvc *search.Service
if cfg.Search_Enabled {
    searchSvc = search.NewService(driver, search.DefaultConfig(cfg))
    if err := searchSvc.Start(rootCtx); err != nil {
        utility.DefaultLogger.Error("search index failed to start", err)
        // Non-fatal: CMS runs without search
    }
}
```

Pass `searchSvc` to the publishing pipeline and router.

### 8.5 HTTP endpoint

**File:** `internal/router/search.go`

```go
// GET /api/v1/search?q=...&type=...&locale=...&limit=...&offset=...
func SearchHandler(w http.ResponseWriter, r *http.Request, searchSvc *search.Service)
```

Query parameters:
- `q` (required) — search query string
- `type` (optional) — filter to datatype name
- `locale` (optional) — filter to locale
- `limit` (optional) — default 20, max 100
- `offset` (optional) — default 0
- `prefix` (optional, bool) — enable prefix matching for last term (default true)

Response: JSON `SearchResponse`

```json
{
  "query": "installation guide",
  "results": [
    {
      "id": "01ARZ3NDEK...",
      "content_data_id": "01ARZ3NDEK...",
      "route_slug": "/docs/getting-started",
      "route_title": "Getting Started",
      "datatype_name": "doc_page",
      "datatype_label": "Documentation Page",
      "section": "Installation",
      "section_anchor": "installation",
      "score": 8.42,
      "snippet": "...follow the installation guide to set up ModulaCMS on your server...",
      "published_at": "2026-03-10T12:00:00Z"
    }
  ],
  "total": 3,
  "limit": 20,
  "offset": 0
}
```

**Route registration** in `mux.go`:

```go
if searchSvc != nil {
    mux.HandleFunc("GET /api/v1/search", func(w http.ResponseWriter, r *http.Request) {
        SearchHandler(w, r, searchSvc)
    })
}
```

No auth required — search only indexes published content.

**Admin content search** uses the same endpoint with a `scope` parameter:
- `GET /api/v1/search?q=...&scope=public` (default) — searches public content
- `GET /api/v1/search?q=...&scope=admin` — searches admin panel content

Both scopes use the same index and ranking logic. The `scope` parameter filters by document source. Alternatively, two separate endpoints could be used, but a single endpoint with scope filtering is simpler for SDK consumers.

### 8.6 Admin rebuild endpoint

```go
// POST /api/v1/admin/search/rebuild
```

Protected by `search:update` permission. Triggers full rebuild from database. Returns `{"status": "ok", "documents": 1234}`.

### 8.7 MCP tool

The MCP server can expose search via the existing tool pattern. No special handling needed — the HTTP endpoint is sufficient. If desired, add a `search` tool that calls the search service directly.

---

## Phase 9: SDK Support

### 9.1 TypeScript SDK

```typescript
// In @modulacms/sdk
async search(query: string, options?: SearchOptions): Promise<SearchResponse>

type SearchOptions = {
  type?: string
  locale?: string
  limit?: number
  offset?: number
  prefix?: boolean
}

type SearchResult = {
  id: string
  content_data_id: string
  route_slug: string
  route_title: string
  datatype_name: string
  datatype_label: string
  section?: string
  section_anchor?: string
  score: number
  snippet: string
  published_at: string
}

type SearchResponse = {
  query: string
  results: SearchResult[]
  total: number
  limit: number
  offset: number
}
```

### 9.2 Go SDK

```go
type SearchOptions struct {
    Type   string
    Locale string
    Limit  int
    Offset int
    Prefix bool
}

type SearchResult struct {
    ID            string  `json:"id"`
    ContentDataID string  `json:"content_data_id"`
    RouteSlug     string  `json:"route_slug"`
    RouteTitle    string  `json:"route_title"`
    DatatypeName  string  `json:"datatype_name"`
    DatatypeLabel string  `json:"datatype_label"`
    Section       string  `json:"section,omitempty"`
    SectionAnchor string  `json:"section_anchor,omitempty"`
    Score         float64 `json:"score"`
    Snippet       string  `json:"snippet"`
    PublishedAt   string  `json:"published_at"`
}

func (c *Client) Search(ctx context.Context, query string, opts *SearchOptions) (*SearchResponse, error)
```

### 9.3 Swift SDK

```swift
public struct SearchOptions: Encodable, Sendable {
    public let type: String?
    public let locale: String?
    public let limit: Int?
    public let offset: Int?
    public let prefix: Bool?
}

public struct SearchResult: Codable, Sendable {
    public let id: String
    public let contentDataID: String
    public let routeSlug: String
    public let routeTitle: String
    public let datatypeName: String
    public let datatypeLabel: String
    public let section: String?
    public let sectionAnchor: String?
    public let score: Double
    public let snippet: String
    public let publishedAt: String
}

public func search(_ query: String, options: SearchOptions? = nil) async throws -> SearchResponse
```

---

## Phase 10: Tests

### 10.1 Tokenizer tests

| Input | Expected Terms |
|-------|---------------|
| `"Hello, World!"` | `["hello", "world"]` |
| `"<p>Hello <b>world</b></p>"` | `["hello", "world"]` |
| `"it's a test"` | `["it", "s", "test"]` (stop: "a") |
| `"café résumé"` | `["café", "résumé"]` |
| `""` | `[]` |
| `"C++ is great"` | `["c", "great"]` (stop: "is") |

### 10.2 Index add/search tests

1. Add 3 documents → search for term in doc 2 → returns doc 2 first
2. Add document with title "Installation Guide" → search "install" → matches via prefix
3. Add document with title weight 3.0 and body weight 1.0 → title match ranks higher
4. Search for term not in index → empty results
5. Remove document by content ID → no longer appears in search

### 10.3 BM25 scoring tests

1. Rare term in short document scores higher than common term in long document
2. Same term frequency: shorter document scores higher (length normalization)
3. Term appearing in 1 of 100 docs scores higher than term in 50 of 100 docs (IDF)

### 10.4 Section splitting tests

| Input HTML | Expected Sections |
|-----------|-------------------|
| `"<h2>Setup</h2><p>Install it.</p><h2>Usage</h2><p>Run it.</p>"` | `[{Heading:"Setup", Anchor:"setup", Body:"Install it."}, {Heading:"Usage", Anchor:"usage", Body:"Run it."}]` |
| `"<p>No headings here.</p>"` | `[]` (no sections — content stays in root doc) |
| `"<h3>Nested <em>Heading</em></h3><p>Body</p>"` | `[{Heading:"Nested Heading", Anchor:"nested-heading", Body:"Body"}]` |

### 10.5 Snippet extraction tests

1. Query terms near beginning → snippet starts at beginning (no `...` prefix)
2. Query terms in middle → snippet centered on terms with `...` prefix/suffix
3. Multiple query terms → snippet window covers as many as possible
4. No matching terms → return first N characters of first text field

### 10.6 Persistence round-trip tests

1. Build index → save to temp file → load from temp file → search produces same results
2. Corrupt file → load returns error → rebuild from DB succeeds
3. Empty index → save → load → empty index

### 10.7 Integration tests

1. Publish content → verify document appears in search
2. Unpublish → verify document removed
3. Republish with changed content → verify updated content searchable
4. Search with datatype filter → only matching type returned
5. Search with locale filter → only matching locale returned

### 10.8 Adversarial tests

1. Query with only stop words → returns empty (gracefully)
2. Query with 1000-character string → truncated, no crash
3. Document with 100K character richtext field → indexes without timeout
4. Concurrent reads and writes → no data races (run with `-race`)
5. Empty index → search returns empty results, not error
6. Search for `<script>alert('xss')</script>` → treated as text tokens, no injection

---

## Implementation Order

```
Phase 1 (Tokenizer)       ← standalone, no dependencies
Phase 2 (Core Index)      ← depends on Phase 1
Phase 3 (Scoring)         ← depends on Phase 2
Phase 4 (Document Builder)← depends on Phase 1, needs publishing types
Phase 5 (Search Execution)← depends on Phase 2, 3
Phase 6 (Snippets)        ← depends on Phase 1
Phase 7 (Persistence)     ← depends on Phase 2
    ↓
Phase 8 (Integration)     ← depends on all above
Phase 9 (SDKs)            ← depends on Phase 8 (API shape)
Phase 10 (Tests)          ← throughout, but integration tests last
```

Phases 1, 3, 4, 6 can be developed in parallel. Phase 2 is the critical path.

---

## Estimated Size

| File | Lines |
|------|-------|
| tokenizer.go | ~80 |
| index.go | ~150 |
| scoring.go | ~60 |
| document.go | ~30 |
| builder.go | ~120 |
| storage.go | ~60 |
| search.go | ~130 |
| snippet.go | ~80 |
| config.go | ~50 |
| service.go | ~80 |
| **Total package** | **~840** |
| router/search.go | ~60 |
| Integration (publish hook, config, startup) | ~40 |
| Tests | ~500 |
| SDK additions (3 languages) | ~120 |
| **Grand total** | **~1560** |

Zero dependencies. ~840 lines of search logic. The entire package is smaller than many single dependency files.
