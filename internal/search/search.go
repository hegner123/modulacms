package search

import (
	"sort"
	"strings"
	"unicode"
)

// ParsedQuery holds the result of parsing a search query string into
// individual terms and quoted phrase groups.
type ParsedQuery struct {
	Terms   []string
	Phrases [][]string
}

// ParseQuery breaks a query string into individual terms and quoted phrases.
// Stop words are removed from both. Phrases reduced to a single term after
// filtering are promoted to Terms; empty phrases are dropped.
func ParseQuery(query string, stopWords map[string]bool) ParsedQuery {
	var pq ParsedQuery
	var current strings.Builder
	inPhrase := false
	var phraseTerms []string

	flush := func() {
		if current.Len() == 0 {
			return
		}
		word := strings.ToLower(current.String())
		current.Reset()
		if stopWords[word] {
			return
		}
		if inPhrase {
			phraseTerms = append(phraseTerms, word)
		} else {
			pq.Terms = append(pq.Terms, word)
		}
	}

	for _, r := range query {
		if r == '"' {
			if inPhrase {
				flush()
				inPhrase = false
				if len(phraseTerms) == 1 {
					pq.Terms = append(pq.Terms, phraseTerms[0])
				} else if len(phraseTerms) > 1 {
					pq.Phrases = append(pq.Phrases, phraseTerms)
				}
				phraseTerms = nil
			} else {
				flush()
				inPhrase = true
			}
			continue
		}

		if unicode.IsSpace(r) {
			flush()
			continue
		}

		current.WriteRune(r)
	}

	// Handle unclosed quote or trailing token.
	flush()
	if inPhrase && len(phraseTerms) > 0 {
		if len(phraseTerms) == 1 {
			pq.Terms = append(pq.Terms, phraseTerms[0])
		} else {
			pq.Phrases = append(pq.Phrases, phraseTerms)
		}
	}

	return pq
}

// ScoreDocument computes the BM25-based relevance score for a single document
// across all query terms and fields.
func ScoreDocument(idx *Index, docIdx int, queryTerms []string, termDFs map[string]int) float64 {
	var totalScore float64

	// For proximity bonus: collect per-field positions grouped by term index.
	// Key: fieldIdx, Value: slice of position slices (one per query term).
	fieldTermPositions := make(map[uint16][][]int)

	for termI, term := range queryTerms {
		postings, ok := idx.postings[term]
		if !ok {
			continue
		}

		// Group postings for this doc by field.
		type fieldGroup struct {
			count     int
			positions []int
		}
		groups := make(map[uint16]*fieldGroup)

		for _, p := range postings {
			if int(p.DocIdx) != docIdx {
				continue
			}
			g, exists := groups[p.FieldIdx]
			if !exists {
				g = &fieldGroup{}
				groups[p.FieldIdx] = g
			}
			g.count++
			g.positions = append(g.positions, int(p.Position))
		}

		df := float64(termDFs[term])

		for fieldIdx, g := range groups {
			fieldLen := 1.0
			if docIdx < len(idx.fieldLengths) && idx.fieldLengths[docIdx] != nil {
				if l, ok := idx.fieldLengths[docIdx][fieldIdx]; ok && l > 0 {
					fieldLen = float64(l)
				}
			}

			avgFL := 1.0
			if a, ok := idx.avgFieldLen[fieldIdx]; ok && a > 0 {
				avgFL = a
			}

			score := BM25(float64(g.count), df, fieldLen, avgFL, idx.docCount, idx.config.K1, idx.config.B)

			fieldName := idx.getFieldName(fieldIdx)
			weight := 1.0
			if w, ok := idx.config.FieldWeights[fieldName]; ok {
				weight = w
			}

			totalScore += score * weight

			// Collect positions for proximity bonus.
			if _, ok := fieldTermPositions[fieldIdx]; !ok {
				fieldTermPositions[fieldIdx] = make([][]int, len(queryTerms))
			}
			for len(fieldTermPositions[fieldIdx]) <= termI {
				fieldTermPositions[fieldIdx] = append(fieldTermPositions[fieldIdx], nil)
			}
			fieldTermPositions[fieldIdx][termI] = g.positions
		}
	}

	// Apply proximity bonus per field.
	for _, termPositions := range fieldTermPositions {
		bonus := ProximityBonus(termPositions, len(queryTerms))
		if bonus > 1.0 {
			totalScore *= bonus
		}
	}

	return totalScore
}

// computeTermDFs calculates document frequency for each term. A term's DF is
// the number of unique documents it appears in.
func computeTermDFs(idx *Index, terms []string) map[string]int {
	dfs := make(map[string]int, len(terms))
	for _, term := range terms {
		postings, ok := idx.postings[term]
		if !ok {
			continue
		}
		seen := make(map[uint32]bool)
		for _, p := range postings {
			seen[p.DocIdx] = true
		}
		dfs[term] = len(seen)
	}
	return dfs
}

// collectCandidates returns the set of unique document indices that contain
// at least one of the given terms.
func collectCandidates(idx *Index, terms []string) map[int]bool {
	candidates := make(map[int]bool)
	for _, term := range terms {
		for _, p := range idx.postings[term] {
			candidates[int(p.DocIdx)] = true
		}
	}
	return candidates
}

// phraseMatchesInDoc returns true if the phrase terms appear adjacent in the
// same field within the document at docIdx.
func phraseMatchesInDoc(idx *Index, docIdx int, phrase []string) bool {
	if len(phrase) < 2 {
		return true
	}

	// Collect per-field postings for the first term.
	type fieldPosting struct {
		fieldIdx uint16
		position uint16
	}

	firstTermPostings := make(map[uint16][]uint16)
	if postings, ok := idx.postings[phrase[0]]; ok {
		for _, p := range postings {
			if int(p.DocIdx) == docIdx {
				firstTermPostings[p.FieldIdx] = append(firstTermPostings[p.FieldIdx], p.Position)
			}
		}
	}

	if len(firstTermPostings) == 0 {
		return false
	}

	// For each field that has the first term, check if subsequent terms
	// appear at consecutive positions.
	for fieldIdx, startPositions := range firstTermPostings {
		for _, startPos := range startPositions {
			matched := true
			for offset := 1; offset < len(phrase); offset++ {
				targetPos := startPos + uint16(offset)
				found := false
				if postings, ok := idx.postings[phrase[offset]]; ok {
					for _, p := range postings {
						if int(p.DocIdx) == docIdx && p.FieldIdx == fieldIdx && p.Position == targetPos {
							found = true
							break
						}
					}
				}
				if !found {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
	}

	return false
}

// Search performs a full-text search against the index and returns paginated,
// scored results.
func (idx *Index) Search(query string, opts SearchOptions) SearchResponse {
	parsed := ParseQuery(query, idx.config.StopWords)

	// Collect all query terms (individual + flattened phrases).
	allTerms := collectAllTerms(parsed)

	if len(allTerms) == 0 {
		return SearchResponse{Query: query}
	}

	idx.mu.RLock()
	defer idx.mu.RUnlock()

	termDFs := computeTermDFs(idx, allTerms)
	candidates := collectCandidates(idx, allTerms)

	// Phrase filtering: remove candidates that don't contain all phrases.
	for _, phrase := range parsed.Phrases {
		for docIdx := range candidates {
			if !phraseMatchesInDoc(idx, docIdx, phrase) {
				delete(candidates, docIdx)
			}
		}
	}

	// Apply search option filters.
	for docIdx := range candidates {
		doc := idx.docs[docIdx]
		if opts.DatatypeName != "" && doc.DatatypeName != opts.DatatypeName {
			delete(candidates, docIdx)
			continue
		}
		if opts.Locale != "" && doc.Locale != opts.Locale {
			delete(candidates, docIdx)
		}
	}

	// Score candidates.
	type scored struct {
		docIdx int
		score  float64
	}
	results := make([]scored, 0, len(candidates))
	for docIdx := range candidates {
		s := ScoreDocument(idx, docIdx, allTerms, termDFs)
		results = append(results, scored{docIdx: docIdx, score: s})
	}

	// Sort by score descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	total := len(results)

	// Resolve limit and offset.
	limit := opts.Limit
	if limit <= 0 {
		limit = idx.config.DefaultLimit
	}
	if limit > idx.config.MaxResults {
		limit = idx.config.MaxResults
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	// Paginate.
	if offset >= len(results) {
		return SearchResponse{
			Query:  query,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	page := results[offset:end]

	// Build response results.
	searchResults := make([]SearchResult, len(page))
	for i, entry := range page {
		doc := idx.docs[entry.docIdx]
		searchResults[i] = SearchResult{
			ID:            doc.ID,
			ContentDataID: doc.ContentDataID,
			RouteSlug:     doc.RouteSlug,
			RouteTitle:    doc.RouteTitle,
			DatatypeName:  doc.DatatypeName,
			DatatypeLabel: doc.DatatypeLabel,
			Locale:        doc.Locale,
			Section:       doc.Section,
			SectionAnchor: doc.SectionAnchor,
			Score:         entry.score,
			Snippet:       ExtractSnippet(doc, allTerms, idx.config.SnippetLength),
			PublishedAt:   doc.PublishedAt,
		}
	}

	return SearchResponse{
		Query:   query,
		Results: searchResults,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
}

// SearchPrefix returns up to 20 terms from the index that start with the
// given prefix. Uses binary search over sorted terms for efficiency.
// Acquires a write lock because ensureSorted may rebuild the sorted terms cache.
func (idx *Index) SearchPrefix(prefix string) []string {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.searchPrefixUnlocked(prefix)
}

// searchPrefixUnlocked performs prefix search without acquiring a lock.
// The caller must hold a write lock (ensureSorted may mutate sortedTerms).
func (idx *Index) searchPrefixUnlocked(prefix string) []string {
	idx.ensureSorted()

	prefix = strings.ToLower(prefix)
	if prefix == "" {
		return nil
	}

	i := sort.SearchStrings(idx.sortedTerms, prefix)

	var matches []string
	for i < len(idx.sortedTerms) && len(matches) < 20 {
		if !strings.HasPrefix(idx.sortedTerms[i], prefix) {
			break
		}
		matches = append(matches, idx.sortedTerms[i])
		i++
	}

	return matches
}

// SearchWithPrefix performs a search where the last term is treated as a
// prefix, expanding it against the index vocabulary. Results from all
// expansions are merged by best score per document.
func (idx *Index) SearchWithPrefix(query string, opts SearchOptions) SearchResponse {
	parsed := ParseQuery(query, idx.config.StopWords)
	allTerms := collectAllTerms(parsed)

	if len(allTerms) == 0 {
		return SearchResponse{Query: query}
	}

	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Split: all terms except the last are exact, the last is a prefix.
	exactTerms := allTerms[:len(allTerms)-1]
	prefixTerm := allTerms[len(allTerms)-1]

	// Find prefix expansions.
	expansions := idx.searchPrefixUnlocked(prefixTerm)
	if len(expansions) == 0 {
		// No expansions found: fall back to treating prefix as exact term.
		expansions = []string{prefixTerm}
	}

	// For each expansion, build a full term list and score candidates.
	// Track best score per document across all expansions.
	bestScores := make(map[int]float64)

	for _, expanded := range expansions {
		queryTerms := make([]string, len(exactTerms)+1)
		copy(queryTerms, exactTerms)
		queryTerms[len(exactTerms)] = expanded

		termDFs := computeTermDFs(idx, queryTerms)
		candidates := collectCandidates(idx, queryTerms)

		// Phrase filtering.
		for _, phrase := range parsed.Phrases {
			for docIdx := range candidates {
				if !phraseMatchesInDoc(idx, docIdx, phrase) {
					delete(candidates, docIdx)
				}
			}
		}

		// Option filters.
		for docIdx := range candidates {
			doc := idx.docs[docIdx]
			if opts.DatatypeName != "" && doc.DatatypeName != opts.DatatypeName {
				delete(candidates, docIdx)
				continue
			}
			if opts.Locale != "" && doc.Locale != opts.Locale {
				delete(candidates, docIdx)
			}
		}

		// Score and merge.
		for docIdx := range candidates {
			s := ScoreDocument(idx, docIdx, queryTerms, termDFs)
			if s > bestScores[docIdx] {
				bestScores[docIdx] = s
			}
		}
	}

	// Collect and sort by score.
	type scored struct {
		docIdx int
		score  float64
	}
	results := make([]scored, 0, len(bestScores))
	for docIdx, s := range bestScores {
		results = append(results, scored{docIdx: docIdx, score: s})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	total := len(results)

	// Resolve limit and offset.
	limit := opts.Limit
	if limit <= 0 {
		limit = idx.config.DefaultLimit
	}
	if limit > idx.config.MaxResults {
		limit = idx.config.MaxResults
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	// Paginate.
	if offset >= len(results) {
		return SearchResponse{
			Query:  query,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
	}
	end := offset + limit
	if end > len(results) {
		end = len(results)
	}
	page := results[offset:end]

	// Build response.
	searchResults := make([]SearchResult, len(page))
	for i, entry := range page {
		doc := idx.docs[entry.docIdx]
		searchResults[i] = SearchResult{
			ID:            doc.ID,
			ContentDataID: doc.ContentDataID,
			RouteSlug:     doc.RouteSlug,
			RouteTitle:    doc.RouteTitle,
			DatatypeName:  doc.DatatypeName,
			DatatypeLabel: doc.DatatypeLabel,
			Locale:        doc.Locale,
			Section:       doc.Section,
			SectionAnchor: doc.SectionAnchor,
			Score:         entry.score,
			Snippet:       ExtractSnippet(doc, allTerms, idx.config.SnippetLength),
			PublishedAt:   doc.PublishedAt,
		}
	}

	return SearchResponse{
		Query:   query,
		Results: searchResults,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
}

// collectAllTerms gathers all unique terms from a ParsedQuery: individual
// terms plus all terms from phrases, preserving order and deduplicating.
func collectAllTerms(pq ParsedQuery) []string {
	seen := make(map[string]bool)
	var all []string
	for _, t := range pq.Terms {
		if !seen[t] {
			seen[t] = true
			all = append(all, t)
		}
	}
	for _, phrase := range pq.Phrases {
		for _, t := range phrase {
			if !seen[t] {
				seen[t] = true
				all = append(all, t)
			}
		}
	}
	return all
}
