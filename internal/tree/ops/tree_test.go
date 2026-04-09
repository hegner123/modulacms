package ops_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/tree/ops"
)

// testID is a plain string type satisfying the ~string constraint.
type testID = string

// mockBackend is an in-memory implementation of ops.Backend[string] for testing.
type mockBackend struct {
	nodes   map[testID]*ops.Node[testID]
	updates int // tracks total UpdatePointers calls
}

func newMockBackend() *mockBackend {
	return &mockBackend{nodes: make(map[testID]*ops.Node[testID])}
}

func (m *mockBackend) GetNode(_ context.Context, id testID) (*ops.Node[testID], error) {
	n, ok := m.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
	}
	// Return a copy to avoid aliasing
	cp := *n
	return &cp, nil
}

func (m *mockBackend) UpdatePointers(_ context.Context, _ audited.AuditContext, id testID, ptrs ops.Pointers[testID]) error {
	n, ok := m.nodes[id]
	if !ok {
		return fmt.Errorf("node %s not found", id)
	}
	n.ParentID = ptrs.ParentID
	n.FirstChildID = ptrs.FirstChildID
	n.NextSiblingID = ptrs.NextSiblingID
	n.PrevSiblingID = ptrs.PrevSiblingID
	m.updates++
	return nil
}

// addNode is a test helper to insert a node into the mock backend.
func (m *mockBackend) addNode(id testID, parentID, firstChild, nextSib, prevSib ops.NullableID[testID]) {
	m.nodes[id] = &ops.Node[testID]{
		ID:            id,
		ParentID:      parentID,
		FirstChildID:  firstChild,
		NextSiblingID: nextSib,
		PrevSiblingID: prevSib,
	}
}

var (
	ac  = audited.AuditContext{NodeID: types.NodeID("test-node"), UserID: types.UserID("test-user")}
	ctx = context.Background()
)

func null(id testID) ops.NullableID[testID] { return ops.NullID[testID](id) }
func empty() ops.NullableID[testID]         { return ops.EmptyID[testID]() }

// --- DetectCycle tests ---

func TestDetectCycle(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*mockBackend)
		nodeID         testID
		proposedParent ops.NullableID[testID]
		wantErr        bool
	}{
		{
			name:           "no cycle - moving to root",
			setup:          func(m *mockBackend) {},
			nodeID:         "A",
			proposedParent: empty(),
			wantErr:        false,
		},
		{
			name: "no cycle - linear chain",
			setup: func(m *mockBackend) {
				m.addNode("A", empty(), null("B"), empty(), empty())
				m.addNode("B", null("A"), empty(), empty(), empty())
				m.addNode("C", empty(), empty(), empty(), empty())
			},
			nodeID:         "C",
			proposedParent: null("B"),
			wantErr:        false,
		},
		{
			name: "cycle - node would become own ancestor",
			setup: func(m *mockBackend) {
				m.addNode("A", empty(), null("B"), empty(), empty())
				m.addNode("B", null("A"), null("C"), empty(), empty())
				m.addNode("C", null("B"), empty(), empty(), empty())
			},
			nodeID:         "A",
			proposedParent: null("C"),
			wantErr:        true,
		},
		{
			name: "cycle - direct self-move",
			setup: func(m *mockBackend) {
				m.addNode("A", empty(), empty(), empty(), empty())
			},
			nodeID:         "A",
			proposedParent: null("A"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := newMockBackend()
			tt.setup(mb)
			err := ops.DetectCycle(ctx, mb, tt.nodeID, tt.proposedParent)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectCycle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// --- Unlink tests ---

func TestUnlink(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*mockBackend)
		nodeID testID
		verify func(*testing.T, *mockBackend)
	}{
		{
			name: "unlink first child with next sibling",
			setup: func(m *mockBackend) {
				// P -> A -> B
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), null("B"), empty())
				m.addNode("B", null("P"), empty(), empty(), null("A"))
			},
			nodeID: "A",
			verify: func(t *testing.T, m *mockBackend) {
				// P's first child should now be B
				p := m.nodes["P"]
				if !p.FirstChildID.Valid || p.FirstChildID.Value != "B" {
					t.Errorf("parent first_child = %v, want B", p.FirstChildID)
				}
				// B's prev should be empty
				b := m.nodes["B"]
				if b.PrevSiblingID.Valid {
					t.Errorf("B prev_sibling = %v, want empty", b.PrevSiblingID)
				}
				// A should be fully detached
				a := m.nodes["A"]
				if a.ParentID.Valid || a.NextSiblingID.Valid || a.PrevSiblingID.Valid {
					t.Errorf("A should be detached, got parent=%v next=%v prev=%v",
						a.ParentID, a.NextSiblingID, a.PrevSiblingID)
				}
			},
		},
		{
			name: "unlink middle child",
			setup: func(m *mockBackend) {
				// P -> A -> B -> C
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), null("B"), empty())
				m.addNode("B", null("P"), empty(), null("C"), null("A"))
				m.addNode("C", null("P"), empty(), empty(), null("B"))
			},
			nodeID: "B",
			verify: func(t *testing.T, m *mockBackend) {
				// A -> C now
				a := m.nodes["A"]
				if !a.NextSiblingID.Valid || a.NextSiblingID.Value != "C" {
					t.Errorf("A next_sibling = %v, want C", a.NextSiblingID)
				}
				c := m.nodes["C"]
				if !c.PrevSiblingID.Valid || c.PrevSiblingID.Value != "A" {
					t.Errorf("C prev_sibling = %v, want A", c.PrevSiblingID)
				}
				// B detached
				b := m.nodes["B"]
				if b.ParentID.Valid || b.NextSiblingID.Valid || b.PrevSiblingID.Valid {
					t.Errorf("B should be detached")
				}
			},
		},
		{
			name: "unlink last child",
			setup: func(m *mockBackend) {
				// P -> A -> B
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), null("B"), empty())
				m.addNode("B", null("P"), empty(), empty(), null("A"))
			},
			nodeID: "B",
			verify: func(t *testing.T, m *mockBackend) {
				a := m.nodes["A"]
				if a.NextSiblingID.Valid {
					t.Errorf("A next_sibling = %v, want empty", a.NextSiblingID)
				}
				b := m.nodes["B"]
				if b.ParentID.Valid || b.PrevSiblingID.Valid {
					t.Errorf("B should be detached")
				}
			},
		},
		{
			name: "unlink only child",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), empty(), empty())
			},
			nodeID: "A",
			verify: func(t *testing.T, m *mockBackend) {
				p := m.nodes["P"]
				if p.FirstChildID.Valid {
					t.Errorf("parent first_child = %v, want empty", p.FirstChildID)
				}
				a := m.nodes["A"]
				if a.ParentID.Valid {
					t.Errorf("A parent = %v, want empty", a.ParentID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := newMockBackend()
			tt.setup(mb)
			node, err := mb.GetNode(ctx, tt.nodeID)
			if err != nil {
				t.Fatalf("setup: get node: %v", err)
			}
			if err := ops.Unlink(ctx, ac, mb, node); err != nil {
				t.Fatalf("Unlink() error: %v", err)
			}
			tt.verify(t, mb)
		})
	}
}

// --- InsertAt tests ---

func TestInsertAt(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*mockBackend)
		parentID testID
		childID  testID
		position int
		verify   func(*testing.T, *mockBackend)
	}{
		{
			name: "insert into empty parent",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), empty(), empty(), empty())
				m.addNode("X", empty(), empty(), empty(), empty())
			},
			parentID: "P",
			childID:  "X",
			position: 0,
			verify: func(t *testing.T, m *mockBackend) {
				p := m.nodes["P"]
				if !p.FirstChildID.Valid || p.FirstChildID.Value != "X" {
					t.Errorf("P first_child = %v, want X", p.FirstChildID)
				}
				x := m.nodes["X"]
				if !x.ParentID.Valid || x.ParentID.Value != "P" {
					t.Errorf("X parent = %v, want P", x.ParentID)
				}
			},
		},
		{
			name: "insert at position 0 (before first child)",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), empty(), empty())
				m.addNode("X", empty(), empty(), empty(), empty())
			},
			parentID: "P",
			childID:  "X",
			position: 0,
			verify: func(t *testing.T, m *mockBackend) {
				p := m.nodes["P"]
				if !p.FirstChildID.Valid || p.FirstChildID.Value != "X" {
					t.Errorf("P first_child = %v, want X", p.FirstChildID)
				}
				x := m.nodes["X"]
				if !x.NextSiblingID.Valid || x.NextSiblingID.Value != "A" {
					t.Errorf("X next = %v, want A", x.NextSiblingID)
				}
				a := m.nodes["A"]
				if !a.PrevSiblingID.Valid || a.PrevSiblingID.Value != "X" {
					t.Errorf("A prev = %v, want X", a.PrevSiblingID)
				}
			},
		},
		{
			name: "insert at position 1 (between A and B)",
			setup: func(m *mockBackend) {
				// P -> A -> B
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), null("B"), empty())
				m.addNode("B", null("P"), empty(), empty(), null("A"))
				m.addNode("X", empty(), empty(), empty(), empty())
			},
			parentID: "P",
			childID:  "X",
			position: 1,
			verify: func(t *testing.T, m *mockBackend) {
				// Should be P -> A -> X -> B
				a := m.nodes["A"]
				if !a.NextSiblingID.Valid || a.NextSiblingID.Value != "X" {
					t.Errorf("A next = %v, want X", a.NextSiblingID)
				}
				x := m.nodes["X"]
				if !x.PrevSiblingID.Valid || x.PrevSiblingID.Value != "A" {
					t.Errorf("X prev = %v, want A", x.PrevSiblingID)
				}
				if !x.NextSiblingID.Valid || x.NextSiblingID.Value != "B" {
					t.Errorf("X next = %v, want B", x.NextSiblingID)
				}
				b := m.nodes["B"]
				if !b.PrevSiblingID.Valid || b.PrevSiblingID.Value != "X" {
					t.Errorf("B prev = %v, want X", b.PrevSiblingID)
				}
			},
		},
		{
			name: "insert at position beyond end (appends)",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), empty(), empty())
				m.addNode("X", empty(), empty(), empty(), empty())
			},
			parentID: "P",
			childID:  "X",
			position: 99,
			verify: func(t *testing.T, m *mockBackend) {
				// Should be P -> A -> X
				a := m.nodes["A"]
				if !a.NextSiblingID.Valid || a.NextSiblingID.Value != "X" {
					t.Errorf("A next = %v, want X", a.NextSiblingID)
				}
				x := m.nodes["X"]
				if !x.PrevSiblingID.Valid || x.PrevSiblingID.Value != "A" {
					t.Errorf("X prev = %v, want A", x.PrevSiblingID)
				}
				if x.NextSiblingID.Valid {
					t.Errorf("X next = %v, want empty", x.NextSiblingID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := newMockBackend()
			tt.setup(mb)
			if err := ops.InsertAt(ctx, ac, mb, tt.parentID, tt.childID, tt.position); err != nil {
				t.Fatalf("InsertAt() error: %v", err)
			}
			tt.verify(t, mb)
		})
	}
}

// --- AppendChild tests ---

func TestAppendChild(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*mockBackend)
		parentID testID
		childID  testID
		verify   func(*testing.T, *mockBackend)
	}{
		{
			name: "append to empty parent",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), empty(), empty(), empty())
				m.addNode("X", empty(), empty(), empty(), empty())
			},
			parentID: "P",
			childID:  "X",
			verify: func(t *testing.T, m *mockBackend) {
				p := m.nodes["P"]
				if !p.FirstChildID.Valid || p.FirstChildID.Value != "X" {
					t.Errorf("P first_child = %v, want X", p.FirstChildID)
				}
				x := m.nodes["X"]
				if !x.ParentID.Valid || x.ParentID.Value != "P" {
					t.Errorf("X parent = %v, want P", x.ParentID)
				}
			},
		},
		{
			name: "append to parent with one child",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), empty(), empty())
				m.addNode("X", empty(), empty(), empty(), empty())
			},
			parentID: "P",
			childID:  "X",
			verify: func(t *testing.T, m *mockBackend) {
				// P -> A -> X
				a := m.nodes["A"]
				if !a.NextSiblingID.Valid || a.NextSiblingID.Value != "X" {
					t.Errorf("A next = %v, want X", a.NextSiblingID)
				}
				x := m.nodes["X"]
				if !x.PrevSiblingID.Valid || x.PrevSiblingID.Value != "A" {
					t.Errorf("X prev = %v, want A", x.PrevSiblingID)
				}
				if !x.ParentID.Valid || x.ParentID.Value != "P" {
					t.Errorf("X parent = %v, want P", x.ParentID)
				}
			},
		},
		{
			name: "append to parent with two children",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), null("B"), empty())
				m.addNode("B", null("P"), empty(), empty(), null("A"))
				m.addNode("X", empty(), empty(), empty(), empty())
			},
			parentID: "P",
			childID:  "X",
			verify: func(t *testing.T, m *mockBackend) {
				// P -> A -> B -> X
				b := m.nodes["B"]
				if !b.NextSiblingID.Valid || b.NextSiblingID.Value != "X" {
					t.Errorf("B next = %v, want X", b.NextSiblingID)
				}
				x := m.nodes["X"]
				if !x.PrevSiblingID.Valid || x.PrevSiblingID.Value != "B" {
					t.Errorf("X prev = %v, want B", x.PrevSiblingID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := newMockBackend()
			tt.setup(mb)
			if err := ops.AppendChild(ctx, ac, mb, tt.parentID, tt.childID); err != nil {
				t.Fatalf("AppendChild() error: %v", err)
			}
			tt.verify(t, mb)
		})
	}
}

// --- Move tests ---

func TestMove(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*mockBackend)
		params  ops.MoveParams[testID]
		verify  func(*testing.T, *mockBackend, *ops.MoveResult[testID])
		wantErr bool
	}{
		{
			name: "move child to different parent",
			setup: func(m *mockBackend) {
				// P1 -> A -> B, P2 (empty)
				m.addNode("P1", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P1"), empty(), null("B"), empty())
				m.addNode("B", null("P1"), empty(), empty(), null("A"))
				m.addNode("P2", empty(), empty(), empty(), empty())
			},
			params: ops.MoveParams[testID]{
				NodeID:      "A",
				NewParentID: null("P2"),
				Position:    0,
			},
			verify: func(t *testing.T, m *mockBackend, r *ops.MoveResult[testID]) {
				// P1 -> B (first child), P2 -> A
				p1 := m.nodes["P1"]
				if !p1.FirstChildID.Valid || p1.FirstChildID.Value != "B" {
					t.Errorf("P1 first_child = %v, want B", p1.FirstChildID)
				}
				p2 := m.nodes["P2"]
				if !p2.FirstChildID.Valid || p2.FirstChildID.Value != "A" {
					t.Errorf("P2 first_child = %v, want A", p2.FirstChildID)
				}
				b := m.nodes["B"]
				if b.PrevSiblingID.Valid {
					t.Errorf("B prev = %v, want empty", b.PrevSiblingID)
				}
				if r.OldParentID.Value != "P1" || r.NewParentID.Value != "P2" {
					t.Errorf("result parents: old=%v new=%v", r.OldParentID, r.NewParentID)
				}
			},
		},
		{
			name: "move to root (clear parent)",
			setup: func(m *mockBackend) {
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), empty(), empty())
			},
			params: ops.MoveParams[testID]{
				NodeID:      "A",
				NewParentID: empty(),
				Position:    0,
			},
			verify: func(t *testing.T, m *mockBackend, _ *ops.MoveResult[testID]) {
				p := m.nodes["P"]
				if p.FirstChildID.Valid {
					t.Errorf("P first_child = %v, want empty", p.FirstChildID)
				}
				a := m.nodes["A"]
				if a.ParentID.Valid {
					t.Errorf("A parent = %v, want empty", a.ParentID)
				}
			},
		},
		{
			name: "cycle detection blocks move",
			setup: func(m *mockBackend) {
				// P -> A -> B (try to move P under B)
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), null("B"), empty(), empty())
				m.addNode("B", null("A"), empty(), empty(), empty())
			},
			params: ops.MoveParams[testID]{
				NodeID:      "P",
				NewParentID: null("B"),
				Position:    0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := newMockBackend()
			tt.setup(mb)
			result, err := ops.Move(ctx, ac, mb, tt.params)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Move() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.verify != nil && err == nil {
				tt.verify(t, mb, result)
			}
		})
	}
}

// --- Reorder tests ---

func TestReorder(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*mockBackend)
		parentID   ops.NullableID[testID]
		orderedIDs []testID
		verify     func(*testing.T, *mockBackend)
		wantErr    bool
	}{
		{
			name: "reverse two siblings",
			setup: func(m *mockBackend) {
				// P -> A -> B
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), null("B"), empty())
				m.addNode("B", null("P"), empty(), empty(), null("A"))
			},
			parentID:   null("P"),
			orderedIDs: []testID{"B", "A"},
			verify: func(t *testing.T, m *mockBackend) {
				// P -> B -> A
				p := m.nodes["P"]
				if !p.FirstChildID.Valid || p.FirstChildID.Value != "B" {
					t.Errorf("P first_child = %v, want B", p.FirstChildID)
				}
				b := m.nodes["B"]
				if b.PrevSiblingID.Valid {
					t.Errorf("B prev = %v, want empty", b.PrevSiblingID)
				}
				if !b.NextSiblingID.Valid || b.NextSiblingID.Value != "A" {
					t.Errorf("B next = %v, want A", b.NextSiblingID)
				}
				a := m.nodes["A"]
				if !a.PrevSiblingID.Valid || a.PrevSiblingID.Value != "B" {
					t.Errorf("A prev = %v, want B", a.PrevSiblingID)
				}
				if a.NextSiblingID.Valid {
					t.Errorf("A next = %v, want empty", a.NextSiblingID)
				}
			},
		},
		{
			name: "reorder three siblings",
			setup: func(m *mockBackend) {
				// P -> A -> B -> C
				m.addNode("P", empty(), null("A"), empty(), empty())
				m.addNode("A", null("P"), empty(), null("B"), empty())
				m.addNode("B", null("P"), empty(), null("C"), null("A"))
				m.addNode("C", null("P"), empty(), empty(), null("B"))
			},
			parentID:   null("P"),
			orderedIDs: []testID{"C", "A", "B"},
			verify: func(t *testing.T, m *mockBackend) {
				// P -> C -> A -> B
				p := m.nodes["P"]
				if !p.FirstChildID.Valid || p.FirstChildID.Value != "C" {
					t.Errorf("P first_child = %v, want C", p.FirstChildID)
				}
				c := m.nodes["C"]
				if c.PrevSiblingID.Valid {
					t.Errorf("C prev = %v, want empty", c.PrevSiblingID)
				}
				if !c.NextSiblingID.Valid || c.NextSiblingID.Value != "A" {
					t.Errorf("C next = %v, want A", c.NextSiblingID)
				}
				a := m.nodes["A"]
				if !a.PrevSiblingID.Valid || a.PrevSiblingID.Value != "C" {
					t.Errorf("A prev = %v, want C", a.PrevSiblingID)
				}
				if !a.NextSiblingID.Valid || a.NextSiblingID.Value != "B" {
					t.Errorf("A next = %v, want B", a.NextSiblingID)
				}
				b := m.nodes["B"]
				if !b.PrevSiblingID.Valid || b.PrevSiblingID.Value != "A" {
					t.Errorf("B prev = %v, want A", b.PrevSiblingID)
				}
				if b.NextSiblingID.Valid {
					t.Errorf("B next = %v, want empty", b.NextSiblingID)
				}
			},
		},
		{
			name: "wrong parent rejected",
			setup: func(m *mockBackend) {
				m.addNode("P1", empty(), null("A"), empty(), empty())
				m.addNode("P2", empty(), null("B"), empty(), empty())
				m.addNode("A", null("P1"), empty(), empty(), empty())
				m.addNode("B", null("P2"), empty(), empty(), empty())
			},
			parentID:   null("P1"),
			orderedIDs: []testID{"A", "B"},
			wantErr:    true,
		},
		{
			name:       "empty ordered_ids rejected",
			setup:      func(m *mockBackend) {},
			parentID:   empty(),
			orderedIDs: []testID{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mb := newMockBackend()
			tt.setup(mb)
			_, err := ops.Reorder(ctx, ac, mb, tt.parentID, tt.orderedIDs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Reorder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.verify != nil && err == nil {
				tt.verify(t, mb)
			}
		})
	}
}

// --- Save tests ---

func TestSave(t *testing.T) {
	t.Run("create and reorder", func(t *testing.T) {
		mb := newMockBackend()
		mb.addNode("P", empty(), null("A"), empty(), empty())
		mb.addNode("A", null("P"), empty(), empty(), empty())

		params := ops.SaveParams[testID]{
			ParentID: null("P"),
			Creates: []ops.SaveCreate[testID]{
				{
					TempID: "temp-1",
					CreateFn: func(_ context.Context, _ audited.AuditContext) (testID, error) {
						mb.addNode("B", null("P"), empty(), empty(), empty())
						return "B", nil
					},
				},
			},
			Order: []testID{"temp-1", "A"}, // B first, then A
		}

		result, err := ops.Save(ctx, ac, mb, params)
		if err != nil {
			t.Fatalf("Save() error: %v", err)
		}

		if result.Created != 1 {
			t.Errorf("created = %d, want 1", result.Created)
		}
		if result.IDMap["temp-1"] != "B" {
			t.Errorf("IDMap[temp-1] = %v, want B", result.IDMap["temp-1"])
		}

		// Verify order: P -> B -> A
		p := mb.nodes["P"]
		if !p.FirstChildID.Valid || p.FirstChildID.Value != "B" {
			t.Errorf("P first_child = %v, want B", p.FirstChildID)
		}
	})

	t.Run("delete with unlink", func(t *testing.T) {
		mb := newMockBackend()
		// P -> A -> B -> C
		mb.addNode("P", empty(), null("A"), empty(), empty())
		mb.addNode("A", null("P"), empty(), null("B"), empty())
		mb.addNode("B", null("P"), empty(), null("C"), null("A"))
		mb.addNode("C", null("P"), empty(), empty(), null("B"))

		params := ops.SaveParams[testID]{
			ParentID: null("P"),
			Deletes:  []testID{"B"},
			Order:    []testID{"A", "B", "C"}, // B will be filtered out
		}

		result, err := ops.Save(ctx, ac, mb, params)
		if err != nil {
			t.Fatalf("Save() error: %v", err)
		}

		if result.Deleted != 1 {
			t.Errorf("Deleted = %d, want 1", result.Deleted)
		}

		// After save: P -> A -> C (B unlinked, removed from order)
		p := mb.nodes["P"]
		if !p.FirstChildID.Valid || p.FirstChildID.Value != "A" {
			t.Errorf("P first_child = %v, want A", p.FirstChildID)
		}
		a := mb.nodes["A"]
		if !a.NextSiblingID.Valid || a.NextSiblingID.Value != "C" {
			t.Errorf("A next = %v, want C", a.NextSiblingID)
		}
	})
}

// ---------------------------------------------------------------------------
// Edge case: Insert errors
// ---------------------------------------------------------------------------

func TestInsertAt_ParentNotFound(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("X", empty(), empty(), empty(), empty())

	err := ops.InsertAt(ctx, ac, mb, "MISSING", "X", 0)
	if err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestInsertAt_ChildNotFound(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), empty(), empty(), empty())

	err := ops.InsertAt(ctx, ac, mb, "P", "MISSING", 0)
	if err == nil {
		t.Fatal("expected error for missing child")
	}
}

func TestAppendChild_ParentNotFound(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("X", empty(), empty(), empty(), empty())

	err := ops.AppendChild(ctx, ac, mb, "MISSING", "X")
	if err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestAppendChild_ChildNotFound(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), empty(), empty(), empty())

	err := ops.AppendChild(ctx, ac, mb, "P", "MISSING")
	if err == nil {
		t.Fatal("expected error for missing child")
	}
}

func TestInsertAt_ExactEndPosition(t *testing.T) {
	// P -> A -> B, insert X at position 2 (exactly at end, not beyond)
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), null("B"), empty())
	mb.addNode("B", null("P"), empty(), empty(), null("A"))
	mb.addNode("X", empty(), empty(), empty(), empty())

	if err := ops.InsertAt(ctx, ac, mb, "P", "X", 2); err != nil {
		t.Fatalf("InsertAt() error: %v", err)
	}

	// Should be P -> A -> B -> X
	b := mb.nodes["B"]
	if !b.NextSiblingID.Valid || b.NextSiblingID.Value != "X" {
		t.Errorf("B next = %v, want X", b.NextSiblingID)
	}
	x := mb.nodes["X"]
	if !x.PrevSiblingID.Valid || x.PrevSiblingID.Value != "B" {
		t.Errorf("X prev = %v, want B", x.PrevSiblingID)
	}
	if x.NextSiblingID.Valid {
		t.Errorf("X next = %v, want empty", x.NextSiblingID)
	}
}

// ---------------------------------------------------------------------------
// Edge case: Move errors and no-ops
// ---------------------------------------------------------------------------

func TestMove_NodeNotFound(t *testing.T) {
	mb := newMockBackend()

	_, err := ops.Move(ctx, ac, mb, ops.MoveParams[testID]{
		NodeID:      "MISSING",
		NewParentID: empty(),
	})
	if err == nil {
		t.Fatal("expected error for missing node")
	}
}

func TestMove_SameParentAndPosition(t *testing.T) {
	// P -> A -> B, move A to P at position 0 (same place)
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), null("B"), empty())
	mb.addNode("B", null("P"), empty(), empty(), null("A"))

	result, err := ops.Move(ctx, ac, mb, ops.MoveParams[testID]{
		NodeID:      "A",
		NewParentID: null("P"),
		Position:    0,
	})
	if err != nil {
		t.Fatalf("Move() error: %v", err)
	}

	// Should still be P -> A -> B (or A re-inserted at position 0)
	p := mb.nodes["P"]
	if !p.FirstChildID.Valid || p.FirstChildID.Value != "A" {
		t.Errorf("P first_child = %v, want A", p.FirstChildID)
	}
	if result.OldPosition != 0 {
		t.Errorf("OldPosition = %d, want 0", result.OldPosition)
	}
	if result.NewPosition != 0 {
		t.Errorf("NewPosition = %d, want 0", result.NewPosition)
	}
}

func TestMove_DescendantToAncestor(t *testing.T) {
	// P -> A -> B -> C, try to move A under C (cycle)
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), null("B"), empty(), empty())
	mb.addNode("B", null("A"), null("C"), empty(), empty())
	mb.addNode("C", null("B"), empty(), empty(), empty())

	_, err := ops.Move(ctx, ac, mb, ops.MoveParams[testID]{
		NodeID:      "A",
		NewParentID: null("C"),
		Position:    0,
	})
	if err == nil {
		t.Fatal("expected cycle detection error")
	}
}

// ---------------------------------------------------------------------------
// Edge case: Reorder edge cases
// ---------------------------------------------------------------------------

func TestReorder_SingleItem(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), empty(), empty())

	result, err := ops.Reorder(ctx, ac, mb, null("P"), []testID{"A"})
	if err != nil {
		t.Fatalf("Reorder() error: %v", err)
	}
	if result.Updated != 1 {
		t.Errorf("updated = %d, want 1", result.Updated)
	}

	// A should have no siblings
	a := mb.nodes["A"]
	if a.PrevSiblingID.Valid {
		t.Errorf("A prev = %v, want empty", a.PrevSiblingID)
	}
	if a.NextSiblingID.Valid {
		t.Errorf("A next = %v, want empty", a.NextSiblingID)
	}
	// Parent first child should be A
	p := mb.nodes["P"]
	if !p.FirstChildID.Valid || p.FirstChildID.Value != "A" {
		t.Errorf("P first_child = %v, want A", p.FirstChildID)
	}
}

func TestReorder_NodeNotFound(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), empty(), empty())

	_, err := ops.Reorder(ctx, ac, mb, null("P"), []testID{"A", "MISSING"})
	if err == nil {
		t.Fatal("expected error for missing node")
	}
}

func TestReorder_DuplicateIDs(t *testing.T) {
	// P -> A -> B, reorder with ["A", "A"] — should fail because second
	// GetNode("A") sees ParentID already updated and may mismatch, or
	// the sibling chain ends up self-referential.
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), null("B"), empty())
	mb.addNode("B", null("P"), empty(), empty(), null("A"))

	result, err := ops.Reorder(ctx, ac, mb, null("P"), []testID{"A", "A"})
	// The code doesn't validate duplicates — it will link A -> A.
	// This documents the current behavior so we can decide if it needs a fix.
	if err != nil {
		t.Logf("Reorder with duplicates returned error (good): %v", err)
		return
	}
	t.Logf("Reorder with duplicates succeeded (n=%d) — A.NextSiblingID=%v A.PrevSiblingID=%v",
		result.Updated, mb.nodes["A"].NextSiblingID, mb.nodes["A"].PrevSiblingID)

	// If it succeeded, A's next would point to itself — document this as a known gap
	a := mb.nodes["A"]
	if a.NextSiblingID.Valid && a.NextSiblingID.Value == "A" {
		t.Error("KNOWN GAP: duplicate IDs in orderedIDs creates self-referential sibling chain")
	}
}

// ---------------------------------------------------------------------------
// Edge case: Unlink edge cases
// ---------------------------------------------------------------------------

func TestUnlink_RootNode(t *testing.T) {
	// Node with no parent — Unlink should clear pointers without error
	mb := newMockBackend()
	mb.addNode("R", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("R"), empty(), empty(), empty())

	node, err := mb.GetNode(ctx, "R")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := ops.Unlink(ctx, ac, mb, node); err != nil {
		t.Fatalf("Unlink() error: %v", err)
	}

	r := mb.nodes["R"]
	if r.ParentID.Valid {
		t.Errorf("R parent = %v, want empty", r.ParentID)
	}
	if r.NextSiblingID.Valid {
		t.Errorf("R next = %v, want empty", r.NextSiblingID)
	}
	if r.PrevSiblingID.Valid {
		t.Errorf("R prev = %v, want empty", r.PrevSiblingID)
	}
	// FirstChildID should be preserved
	if !r.FirstChildID.Valid || r.FirstChildID.Value != "A" {
		t.Errorf("R first_child = %v, want A (preserved)", r.FirstChildID)
	}
}

func TestUnlink_NodeWithChildren(t *testing.T) {
	// P -> A (A has children B, C)
	// After unlinking A: A's FirstChildID preserved, B and C still under A
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), null("B"), empty(), empty())
	mb.addNode("B", null("A"), empty(), null("C"), empty())
	mb.addNode("C", null("A"), empty(), empty(), null("B"))

	node, err := mb.GetNode(ctx, "A")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := ops.Unlink(ctx, ac, mb, node); err != nil {
		t.Fatalf("Unlink() error: %v", err)
	}

	a := mb.nodes["A"]
	if a.ParentID.Valid {
		t.Errorf("A parent = %v, want empty", a.ParentID)
	}
	// FirstChildID preserved — children stay attached
	if !a.FirstChildID.Valid || a.FirstChildID.Value != "B" {
		t.Errorf("A first_child = %v, want B (preserved)", a.FirstChildID)
	}
	// Parent no longer has A
	p := mb.nodes["P"]
	if p.FirstChildID.Valid {
		t.Errorf("P first_child = %v, want empty (A was only child)", p.FirstChildID)
	}
}

func TestUnlink_AlreadyUnlinked(t *testing.T) {
	// Node with no parent, no siblings — Unlink is idempotent
	mb := newMockBackend()
	mb.addNode("X", empty(), empty(), empty(), empty())

	node, err := mb.GetNode(ctx, "X")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := ops.Unlink(ctx, ac, mb, node); err != nil {
		t.Fatalf("Unlink() error: %v", err)
	}

	x := mb.nodes["X"]
	if x.ParentID.Valid || x.NextSiblingID.Valid || x.PrevSiblingID.Valid {
		t.Errorf("X should remain detached: parent=%v next=%v prev=%v",
			x.ParentID, x.NextSiblingID, x.PrevSiblingID)
	}
}

// ---------------------------------------------------------------------------
// Edge case: Save errors and edge cases
// ---------------------------------------------------------------------------

func TestSave_CreateFailure(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), empty(), empty())

	params := ops.SaveParams[testID]{
		ParentID: null("P"),
		Creates: []ops.SaveCreate[testID]{
			{
				TempID: "temp-fail",
				CreateFn: func(_ context.Context, _ audited.AuditContext) (testID, error) {
					return "", fmt.Errorf("db connection lost")
				},
			},
			{
				TempID: "temp-ok",
				CreateFn: func(_ context.Context, _ audited.AuditContext) (testID, error) {
					mb.addNode("B", null("P"), empty(), empty(), empty())
					return "B", nil
				},
			},
		},
		Order: []testID{"A", "temp-ok"},
	}

	result, err := ops.Save(ctx, ac, mb, params)
	if err == nil {
		t.Fatal("expected error from failed create")
	}
	// The successful create should still have been processed
	if result.Created != 1 {
		t.Errorf("created = %d, want 1 (one succeeded)", result.Created)
	}
	if result.IDMap["temp-ok"] != "B" {
		t.Errorf("IDMap[temp-ok] = %v, want B", result.IDMap["temp-ok"])
	}
}

func TestSave_UpdateFailure(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), null("B"), empty())
	mb.addNode("B", null("P"), empty(), empty(), null("A"))

	updateACalled := false
	params := ops.SaveParams[testID]{
		ParentID: null("P"),
		Updates: []ops.SaveUpdate[testID]{
			{
				ID: "A",
				UpdateFn: func(_ context.Context, _ audited.AuditContext) error {
					updateACalled = true
					return fmt.Errorf("validation failed")
				},
			},
			{
				ID: "B",
				UpdateFn: func(_ context.Context, _ audited.AuditContext) error {
					return nil
				},
			},
		},
		Order: []testID{"A", "B"},
	}

	result, err := ops.Save(ctx, ac, mb, params)
	if err == nil {
		t.Fatal("expected error from failed update")
	}
	if !updateACalled {
		t.Error("UpdateFn for A was not called")
	}
	// B's update should still succeed
	if result.Updated != 1 {
		t.Errorf("Updated = %d, want 1 (one succeeded)", result.Updated)
	}
}

func TestSave_DeleteNonexistent(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), empty(), empty())

	params := ops.SaveParams[testID]{
		ParentID: null("P"),
		Deletes:  []testID{"MISSING"},
		Order:    []testID{"A"},
	}

	result, err := ops.Save(ctx, ac, mb, params)
	if err == nil {
		t.Fatal("expected error from deleting non-existent node")
	}
	if result.Deleted != 0 {
		t.Errorf("Deleted = %d, want 0", result.Deleted)
	}
}

func TestSave_AllDeletedLeavesEmptyOrder(t *testing.T) {
	// Delete all children — order becomes empty after filtering, skip reorder
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), empty(), empty())

	params := ops.SaveParams[testID]{
		ParentID: null("P"),
		Deletes:  []testID{"A"},
		Order:    []testID{"A"}, // filtered to empty after delete
	}

	result, err := ops.Save(ctx, ac, mb, params)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if result.Deleted != 1 {
		t.Errorf("Deleted = %d, want 1", result.Deleted)
	}

	// A should be unlinked, P should have no children
	p := mb.nodes["P"]
	if p.FirstChildID.Valid {
		t.Errorf("P first_child = %v, want empty", p.FirstChildID)
	}
}

func TestSave_EmptyParams(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("P", empty(), null("A"), empty(), empty())
	mb.addNode("A", null("P"), empty(), empty(), empty())

	// No creates, updates, deletes, or order — should be a no-op
	params := ops.SaveParams[testID]{
		ParentID: null("P"),
	}

	result, err := ops.Save(ctx, ac, mb, params)
	if err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if result.Created != 0 || result.Updated != 0 || result.Deleted != 0 {
		t.Errorf("expected all zeros, got created=%d updated=%d deleted=%d",
			result.Created, result.Updated, result.Deleted)
	}
}

// ---------------------------------------------------------------------------
// Edge case: DetectCycle with missing node
// ---------------------------------------------------------------------------

func TestDetectCycle_ProposedParentNotFound(t *testing.T) {
	mb := newMockBackend()
	mb.addNode("A", empty(), empty(), empty(), empty())

	err := ops.DetectCycle(ctx, mb, "A", null("MISSING"))
	if err == nil {
		t.Fatal("expected error when proposed parent doesn't exist")
	}
}
