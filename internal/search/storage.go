package search

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// indexData is the serializable snapshot of an Index.
type indexData struct {
	Docs            []SearchDocument
	DocsByContentID map[string][]int
	Postings        map[string][]Posting
	FieldNames      []string
	FieldNameIdx    map[string]uint16
	FieldLengths    []map[uint16]int
	AvgFieldLen     map[uint16]float64
	DocCount        int
}

var magicBytes = [4]byte{'M', 'C', 'M', 'S'}

const indexVersion uint32 = 1

// Save writes the index to disk at path atomically. It writes to a temporary
// file in the same directory, then renames to the target path.
func (idx *Index) Save(path string) error {
	idx.mu.RLock()
	data := indexData{
		Docs:            idx.docs,
		DocsByContentID: idx.docsByContentID,
		Postings:        idx.postings,
		FieldNames:      idx.fieldNames,
		FieldNameIdx:    idx.fieldNameIdx,
		FieldLengths:    idx.fieldLengths,
		AvgFieldLen:     idx.avgFieldLen,
		DocCount:        idx.docCount,
	}
	idx.mu.RUnlock()

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "mcms-idx-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file for index save: %w", err)
	}
	tmpName := tmp.Name()

	// Clean up temp file on any failure path
	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	// Write magic bytes
	if _, err := tmp.Write(magicBytes[:]); err != nil {
		return fmt.Errorf("write magic bytes: %w", err)
	}

	// Write version
	if err := binary.Write(tmp, binary.LittleEndian, indexVersion); err != nil {
		return fmt.Errorf("write index version: %w", err)
	}

	// Encode index data via gob through a buffered writer
	bw := bufio.NewWriter(tmp)
	enc := gob.NewEncoder(bw)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode index data: %w", err)
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush index data: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync index file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp index file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp index file: %w", err)
	}

	success = true
	return nil
}

// Load reads an index from disk at path and returns a fully populated Index.
func Load(path string, cfg SearchConfig) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open index file: %w", err)
	}
	defer f.Close()

	// Verify magic bytes
	var magic [4]byte
	if _, err := f.Read(magic[:]); err != nil {
		return nil, fmt.Errorf("read magic bytes: %w", err)
	}
	if magic != magicBytes {
		return nil, fmt.Errorf("invalid index file: wrong magic bytes")
	}

	// Verify version
	var version uint32
	if err := binary.Read(f, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("read index version: %w", err)
	}
	if version != indexVersion {
		return nil, fmt.Errorf("unsupported index version: got %d, want %d", version, indexVersion)
	}

	// Decode index data
	br := bufio.NewReader(f)
	dec := gob.NewDecoder(br)
	var data indexData
	if err := dec.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode index data: %w", err)
	}

	idx := NewIndex(cfg)
	idx.docs = data.Docs
	idx.docsByContentID = data.DocsByContentID
	idx.postings = data.Postings
	idx.fieldNames = data.FieldNames
	idx.fieldNameIdx = data.FieldNameIdx
	idx.fieldLengths = data.FieldLengths
	idx.avgFieldLen = data.AvgFieldLen
	idx.docCount = data.DocCount
	idx.sortedDirty = true

	return idx, nil
}
