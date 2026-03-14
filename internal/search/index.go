package search

import (
	"sort"
	"sync"
)

// Posting records the location of a single term occurrence in the index.
type Posting struct {
	DocIdx   uint32
	FieldIdx uint16
	Position uint16
}

// Index is an in-memory inverted index for full-text search over SearchDocuments.
type Index struct {
	mu sync.RWMutex

	docs            []SearchDocument
	docsByContentID map[string][]int
	postings        map[string][]Posting
	sortedTerms     []string
	sortedDirty     bool
	docCount        int
	fieldLengths    []map[uint16]int
	avgFieldLen     map[uint16]float64
	fieldNames      []string
	fieldNameIdx    map[string]uint16
	config          SearchConfig
}

// NewIndex creates an empty Index with the given configuration.
func NewIndex(cfg SearchConfig) *Index {
	return &Index{
		docs:            nil,
		docsByContentID: make(map[string][]int),
		postings:        make(map[string][]Posting),
		sortedDirty:     true,
		fieldLengths:    nil,
		avgFieldLen:     make(map[uint16]float64),
		fieldNames:      nil,
		fieldNameIdx:    make(map[string]uint16),
		config:          cfg,
	}
}

// Add inserts a document into the index.
func (idx *Index) Add(doc SearchDocument) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.addUnlocked(doc)
}

// addUnlocked performs the actual document insertion without acquiring the lock.
// It must only be called while idx.mu is held for writing.
func (idx *Index) addUnlocked(doc SearchDocument) {
	idx.docs = append(idx.docs, doc)
	docIdx := uint32(len(idx.docs) - 1)

	idx.docsByContentID[doc.ContentDataID] = append(
		idx.docsByContentID[doc.ContentDataID],
		int(docIdx),
	)

	// Grow fieldLengths slice to accommodate docIdx
	for len(idx.fieldLengths) <= int(docIdx) {
		idx.fieldLengths = append(idx.fieldLengths, nil)
	}
	idx.fieldLengths[docIdx] = make(map[uint16]int)

	for fieldName, value := range doc.Fields {
		fieldIdx := idx.getOrRegisterField(fieldName)

		terms, positions := TokenizeAndFilter(StripHTML(value), idx.config.StopWords)
		idx.fieldLengths[docIdx][fieldIdx] = len(terms)

		for i, term := range terms {
			pos := uint16(0)
			if positions[i] <= int(^uint16(0)) {
				pos = uint16(positions[i])
			}
			idx.postings[term] = append(idx.postings[term], Posting{
				DocIdx:   docIdx,
				FieldIdx: fieldIdx,
				Position: pos,
			})
		}
	}

	idx.sortedDirty = true
	idx.docCount++
	idx.recalcAvgFieldLen()
}

// getOrRegisterField returns the uint16 index for the given field name,
// registering it if not yet known.
func (idx *Index) getOrRegisterField(name string) uint16 {
	if fi, ok := idx.fieldNameIdx[name]; ok {
		return fi
	}
	fi := uint16(len(idx.fieldNames))
	idx.fieldNames = append(idx.fieldNames, name)
	idx.fieldNameIdx[name] = fi
	return fi
}

// recalcAvgFieldLen recomputes avgFieldLen from the current fieldLengths.
func (idx *Index) recalcAvgFieldLen() {
	sums := make(map[uint16]float64)
	counts := make(map[uint16]float64)

	for _, fl := range idx.fieldLengths {
		if fl == nil {
			continue
		}
		for fi, length := range fl {
			sums[fi] += float64(length)
			counts[fi]++
		}
	}

	avg := make(map[uint16]float64, len(sums))
	for fi, s := range sums {
		if counts[fi] > 0 {
			avg[fi] = s / counts[fi]
		}
	}
	idx.avgFieldLen = avg
}

// RemoveByContentID removes all documents that share the given ContentDataID
// and rebuilds the index from the remaining documents.
func (idx *Index) RemoveByContentID(contentDataID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	indices, ok := idx.docsByContentID[contentDataID]
	if !ok || len(indices) == 0 {
		return
	}

	// Build a set of doc indices to remove
	removeSet := make(map[int]bool, len(indices))
	for _, i := range indices {
		removeSet[i] = true
	}

	// Collect retained documents
	var retained []SearchDocument
	for i, doc := range idx.docs {
		if !removeSet[i] {
			retained = append(retained, doc)
		}
	}

	// Reset internal state
	idx.docs = nil
	idx.docsByContentID = make(map[string][]int)
	idx.postings = make(map[string][]Posting)
	idx.fieldLengths = nil
	idx.avgFieldLen = make(map[uint16]float64)
	idx.fieldNames = nil
	idx.fieldNameIdx = make(map[string]uint16)
	idx.docCount = 0
	idx.sortedDirty = true

	// Re-add retained documents
	for _, doc := range retained {
		idx.addUnlocked(doc)
	}
}

// Len returns the number of documents in the index.
func (idx *Index) Len() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.docCount
}

// Stats returns summary statistics about the index.
func (idx *Index) Stats() IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	totalPostings := 0
	for _, pl := range idx.postings {
		totalPostings += len(pl)
	}

	return IndexStats{
		Documents:   idx.docCount,
		Terms:       len(idx.postings),
		Postings:    totalPostings,
		Fields:      len(idx.fieldNames),
		MemEstimate: int64(idx.docCount*200 + totalPostings*8 + len(idx.postings)*20),
	}
}

// ensureSorted rebuilds sortedTerms from the postings map if needed.
// Must be called while holding at least a read lock.
func (idx *Index) ensureSorted() {
	if !idx.sortedDirty {
		return
	}
	terms := make([]string, 0, len(idx.postings))
	for t := range idx.postings {
		terms = append(terms, t)
	}
	sort.Strings(terms)
	idx.sortedTerms = terms
	idx.sortedDirty = false
}

// getFieldName returns the field name for the given index, or an empty
// string if the index is out of range.
func (idx *Index) getFieldName(fieldIdx uint16) string {
	if int(fieldIdx) < len(idx.fieldNames) {
		return idx.fieldNames[fieldIdx]
	}
	return ""
}
