// Package core provides the shared tree-building algorithms and data structures
// used by the TUI (internal/tree), API (internal/model), and admin panel layers.
// It operates on raw DB types (db.ContentData, db.Datatypes, etc.) and uses
// sibling pointers (FirstChild/NextSibling/PrevSibling) for O(1) manipulation.
package core

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// Node is the shared tree node used by all consumers.
// It holds raw DB types and uses sibling pointers for O(1) manipulation.
type Node struct {
	ContentData   *db.ContentData    // Pointer: nil for synthetic nodes
	Datatype      db.Datatypes       // Value: always populated
	ContentFields []db.ContentFields // Field values for this node
	Fields        []db.Fields        // Field definitions for this node's content fields
	Parent        *Node
	FirstChild    *Node
	NextSibling   *Node
	PrevSibling   *Node
}

// Root is the top-level tree container with an index for O(1) lookups.
type Root struct {
	Node      *Node
	NodeIndex map[types.ContentID]*Node
}

// LoadStats tracks tree construction diagnostics.
type LoadStats struct {
	NodesCount      int
	OrphansResolved int
	RetryAttempts   int
	CircularRefs    []types.ContentID
	FinalOrphans    []types.ContentID
}

// String returns a human-readable summary of LoadStats.
func (stats LoadStats) String() string {
	return fmt.Sprintf("Nodes Count: %d\nOrphans Resolved %d\nRetry Attempts %d\nCircular Refs %v\nFinal Orphans %v\n",
		stats.NodesCount,
		stats.OrphansResolved,
		stats.RetryAttempts,
		stats.CircularRefs,
		stats.FinalOrphans,
	)
}

// newLoadStats returns a LoadStats with initialized slices.
func newLoadStats() *LoadStats {
	return &LoadStats{
		CircularRefs: make([]types.ContentID, 0),
		FinalOrphans: make([]types.ContentID, 0),
	}
}
