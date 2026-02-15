package model

import (
	"testing"

	"github.com/hegner123/modulacms/internal/db"
)

// testLogger implements Logger for tests by discarding all output.
type testLogger struct{}

func (testLogger) Warn(string, error, ...any) {}

// makeDatatype is a helper that creates a Datatype with the given content data fields.
func makeDatatype(contentDataID, parentID, firstChildID, nextSiblingID, prevSiblingID, typeName string) Datatype {
	return Datatype{
		Info: db.DatatypeJSON{Type: typeName},
		Content: db.ContentDataJSON{
			ContentDataID: contentDataID,
			ParentID:      parentID,
			FirstChildID:  firstChildID,
			NextSiblingID: nextSiblingID,
			PrevSiblingID: prevSiblingID,
		},
	}
}

// childIDs extracts ContentDataIDs from a node's children for easy comparison.
func childIDs(n *Node) []string {
	ids := make([]string, len(n.Nodes))
	for i, child := range n.Nodes {
		ids[i] = child.Datatype.Content.ContentDataID
	}
	return ids
}

func TestBuildNodes_SiblingOrdering(t *testing.T) {
	tests := []struct {
		name      string
		datatypes []Datatype
		wantOrder []string // expected child IDs of the root node, in order
	}{
		{
			name: "ordered children via complete chain",
			datatypes: []Datatype{
				// Root: first child is C
				makeDatatype("root", "", "C", "", "", "ROOT"),
				// Children in arbitrary slice order: A, B, C
				// Chain: C -> A -> B
				makeDatatype("A", "root", "", "B", "C", "page"),
				makeDatatype("B", "root", "", "", "A", "page"),
				makeDatatype("C", "root", "", "A", "", "page"),
			},
			wantOrder: []string{"C", "A", "B"},
		},
		{
			name: "broken chain — partial order plus remainder",
			datatypes: []Datatype{
				// Root: first child is C
				makeDatatype("root", "", "C", "", "", "ROOT"),
				// Chain: C -> A, then A's next points to "missing"
				makeDatatype("A", "root", "", "missing", "C", "page"),
				makeDatatype("B", "root", "", "", "", "page"),
				makeDatatype("C", "root", "", "A", "", "page"),
			},
			// C and A are in the chain; B is not in the chain so appended at end
			wantOrder: []string{"C", "A", "B"},
		},
		{
			name: "cycle detection — original order preserved",
			datatypes: []Datatype{
				// Root: first child is A
				makeDatatype("root", "", "A", "", "", "ROOT"),
				// Cycle: A -> B -> A
				makeDatatype("A", "root", "", "B", "", "page"),
				makeDatatype("B", "root", "", "A", "A", "page"),
				makeDatatype("C", "root", "", "", "", "page"),
			},
			// Cycle detected, chain returns nil, so order is whatever BuildNodes produces.
			// We just check all three children are present (no crash, no data loss).
			wantOrder: nil, // special: just check length == 3
		},
		{
			name: "single child — no reordering needed",
			datatypes: []Datatype{
				makeDatatype("root", "", "only", "", "", "ROOT"),
				makeDatatype("only", "root", "", "", "", "page"),
			},
			wantOrder: []string{"only"},
		},
		{
			name: "empty FirstChildID — no crash, children preserved",
			datatypes: []Datatype{
				// Root has no FirstChildID set
				makeDatatype("root", "", "", "", "", "ROOT"),
				makeDatatype("X", "root", "", "", "", "page"),
				makeDatatype("Y", "root", "", "", "", "page"),
			},
			// No reordering attempted; both children present
			wantOrder: nil, // special: just check length == 2
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, err := BuildNodes(testLogger{}, tc.datatypes, nil)
			if err != nil {
				t.Fatalf("BuildNodes returned unexpected error: %v", err)
			}
			if root == nil {
				t.Fatal("BuildNodes returned nil root")
			}

			got := childIDs(root)

			if tc.wantOrder != nil {
				// Exact order check
				if len(got) != len(tc.wantOrder) {
					t.Fatalf("got %d children, want %d: %v", len(got), len(tc.wantOrder), got)
				}
				for i := range tc.wantOrder {
					if got[i] != tc.wantOrder[i] {
						t.Errorf("child[%d] = %q, want %q (full order: %v)", i, got[i], tc.wantOrder[i], got)
					}
				}
			} else {
				// Just verify all children are present (no exact order guarantee)
				expectedCount := 0
				for _, dt := range tc.datatypes {
					if dt.Info.Type != "ROOT" {
						expectedCount++
					}
				}
				if len(got) != expectedCount {
					t.Fatalf("got %d children, want %d: %v", len(got), expectedCount, got)
				}
			}
		})
	}
}

func TestBuildSiblingChain(t *testing.T) {
	tests := []struct {
		name         string
		firstChildID string
		nodes        map[string]*Node
		wantIDs      []string // nil means expect nil return
	}{
		{
			name:         "firstChildID not in index",
			firstChildID: "missing",
			nodes:        map[string]*Node{},
			wantIDs:      nil,
		},
		{
			name:         "single node chain",
			firstChildID: "A",
			nodes: map[string]*Node{
				"A": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "A", NextSiblingID: ""}}},
			},
			wantIDs: []string{"A"},
		},
		{
			name:         "three node chain",
			firstChildID: "A",
			nodes: map[string]*Node{
				"A": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "A", NextSiblingID: "B"}}},
				"B": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "B", NextSiblingID: "C"}}},
				"C": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "C", NextSiblingID: ""}}},
			},
			wantIDs: []string{"A", "B", "C"},
		},
		{
			name:         "cycle returns nil",
			firstChildID: "A",
			nodes: map[string]*Node{
				"A": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "A", NextSiblingID: "B"}}},
				"B": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "B", NextSiblingID: "A"}}},
			},
			wantIDs: nil,
		},
		{
			name:         "broken chain returns partial",
			firstChildID: "A",
			nodes: map[string]*Node{
				"A": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "A", NextSiblingID: "B"}}},
				"B": {Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: "B", NextSiblingID: "missing"}}},
			},
			wantIDs: []string{"A", "B"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildSiblingChain(tc.firstChildID, tc.nodes)

			if tc.wantIDs == nil {
				if got != nil {
					t.Fatalf("expected nil, got %d nodes", len(got))
				}
				return
			}

			if len(got) != len(tc.wantIDs) {
				t.Fatalf("got %d nodes, want %d", len(got), len(tc.wantIDs))
			}
			for i, want := range tc.wantIDs {
				if got[i].Datatype.Content.ContentDataID != want {
					t.Errorf("chain[%d] = %q, want %q", i, got[i].Datatype.Content.ContentDataID, want)
				}
			}
		})
	}
}

func TestMergeOrdered(t *testing.T) {
	mkNode := func(id string) *Node {
		return &Node{Datatype: Datatype{Content: db.ContentDataJSON{ContentDataID: id}}}
	}

	a, b, c, d := mkNode("A"), mkNode("B"), mkNode("C"), mkNode("D")

	tests := []struct {
		name     string
		chain    []*Node
		existing []*Node
		wantIDs  []string
	}{
		{
			name:     "all in chain",
			chain:    []*Node{c, a, b},
			existing: []*Node{a, b, c},
			wantIDs:  []string{"C", "A", "B"},
		},
		{
			name:     "extra node not in chain appended",
			chain:    []*Node{c, a},
			existing: []*Node{a, b, c, d},
			wantIDs:  []string{"C", "A", "B", "D"},
		},
		{
			name:     "single node chain with extras",
			chain:    []*Node{b},
			existing: []*Node{a, b, c},
			wantIDs:  []string{"B", "A", "C"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeOrdered(tc.chain, tc.existing)
			if len(got) != len(tc.wantIDs) {
				t.Fatalf("got %d nodes, want %d", len(got), len(tc.wantIDs))
			}
			for i, want := range tc.wantIDs {
				if got[i].Datatype.Content.ContentDataID != want {
					t.Errorf("result[%d] = %q, want %q", i, got[i].Datatype.Content.ContentDataID, want)
				}
			}
		})
	}
}
