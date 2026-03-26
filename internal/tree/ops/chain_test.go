package ops_test

import (
	"context"
	"testing"

	"github.com/hegner123/modulacms/internal/tree/ops"
)

// --- WalkSiblingChain tests ---

func TestWalkSiblingChain_Empty(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), empty(), empty(), empty())

	snap, err := ops.WalkSiblingChain(ctx, m, "P")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.WalkErr != nil {
		t.Fatalf("unexpected walk error: %v", snap.WalkErr)
	}
	if len(snap.Nodes) != 1 {
		t.Errorf("expected 1 node (parent only), got %d", len(snap.Nodes))
	}
}

func TestWalkSiblingChain_LinearChain(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), null("C"), null("A"))
	m.addNode("C", null("P"), empty(), empty(), null("B"))

	snap, err := ops.WalkSiblingChain(ctx, m, "P")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.WalkErr != nil {
		t.Fatalf("unexpected walk error: %v", snap.WalkErr)
	}
	// P + A + B + C = 4
	if len(snap.Nodes) != 4 {
		t.Errorf("expected 4 nodes, got %d", len(snap.Nodes))
	}
	// Verify order: P, A, B, C
	expected := []string{"P", "A", "B", "C"}
	for i, e := range expected {
		if string(snap.Nodes[i].ID) != e {
			t.Errorf("node[%d] = %s, want %s", i, snap.Nodes[i].ID, e)
		}
	}
}

func TestWalkSiblingChain_DetectsCycle(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), null("A"), null("A")) // cycle: B -> A

	snap, err := ops.WalkSiblingChain(ctx, m, "P")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.WalkErr == nil {
		t.Fatal("expected WalkErr for cycle, got nil")
	}
	// Should have partial snapshot: P, A, B (then cycle detected on revisit to A)
	if len(snap.Nodes) < 2 {
		t.Errorf("expected at least 2 nodes in partial snapshot, got %d", len(snap.Nodes))
	}
}

func TestWalkSiblingChain_DanglingPointer(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("MISSING"), empty())

	snap, err := ops.WalkSiblingChain(ctx, m, "P")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.WalkErr == nil {
		t.Fatal("expected WalkErr for dangling pointer, got nil")
	}
	if len(snap.Nodes) != 2 { // P + A
		t.Errorf("expected 2 nodes, got %d", len(snap.Nodes))
	}
}

func TestWalkSiblingChain_ParentNotFound(t *testing.T) {
	m := newMockBackend()

	_, err := ops.WalkSiblingChain(ctx, m, "MISSING")
	if err == nil {
		t.Fatal("expected error for missing parent, got nil")
	}
}

// --- SnapshotNodes tests ---

func TestSnapshotNodes(t *testing.T) {
	m := newMockBackend()
	m.addNode("A", empty(), empty(), null("B"), empty())
	m.addNode("B", empty(), empty(), empty(), null("A"))

	snap, err := ops.SnapshotNodes(ctx, m, []string{"A", "B"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(snap.Nodes))
	}
	if _, ok := snap.ByID["A"]; !ok {
		t.Error("expected node A in ByID")
	}
	if _, ok := snap.ByID["B"]; !ok {
		t.Error("expected node B in ByID")
	}
}

func TestSnapshotNodes_SkipsMissing(t *testing.T) {
	m := newMockBackend()
	m.addNode("A", empty(), empty(), empty(), empty())

	snap, err := ops.SnapshotNodes(ctx, m, []string{"A", "MISSING"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(snap.Nodes))
	}
	if snap.WalkErr == nil {
		t.Error("expected WalkErr for missing node")
	}
}

// --- ValidateChain tests ---

func TestValidateChain_Healthy(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), empty(), null("A"))

	snap, _ := ops.WalkSiblingChain(ctx, m, "P")
	violations := ops.ValidateChain(snap)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d: %+v", len(violations), violations)
	}
}

func TestValidateChain_BrokenSymmetry(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), empty(), null("WRONG")) // B.prev should be A

	snap, _ := ops.WalkSiblingChain(ctx, m, "P")
	violations := ops.ValidateChain(snap)
	if len(violations) == 0 {
		t.Fatal("expected violations for broken symmetry, got none")
	}
	found := false
	for _, v := range violations {
		if v.Field == "next_sibling_id" || v.Field == "prev_sibling_id" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected symmetry violation, got: %+v", violations)
	}
}

func TestValidateChain_SelfReference(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("A"), empty()) // A.next = A (self-ref)

	snap, _ := ops.SnapshotNodes(ctx, m, []string{"P", "A"})
	violations := ops.ValidateChain(snap)
	found := false
	for _, v := range violations {
		if string(v.NodeID) == "A" && v.Field == "next_sibling_id" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected self-reference violation, got: %+v", violations)
	}
}

func TestValidateChain_MixedParents(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("OTHER"), empty(), empty(), null("A")) // B has wrong parent

	snap, _ := ops.WalkSiblingChain(ctx, m, "P")
	violations := ops.ValidateChain(snap)
	found := false
	for _, v := range violations {
		if string(v.NodeID) == "B" && v.Field == "parent_id" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected parent consistency violation, got: %+v", violations)
	}
}

func TestValidateChain_FirstChildHasPrev(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), empty(), null("GHOST")) // first child has prev

	snap, _ := ops.SnapshotNodes(ctx, m, []string{"P", "A"})
	violations := ops.ValidateChain(snap)
	found := false
	for _, v := range violations {
		if v.Field == "first_child_id" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected first_child_id violation (first child has prev), got: %+v", violations)
	}
}

func TestValidateChain_Nil(t *testing.T) {
	violations := ops.ValidateChain[string](nil)
	if violations != nil {
		t.Errorf("expected nil violations for nil snapshot, got: %+v", violations)
	}
}

// --- AssertChainConsistency tests ---

func TestAssertChainConsistency_Healthy(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), empty(), null("A"))

	snap, _ := ops.WalkSiblingChain(ctx, m, "P")
	results := ops.AssertChainConsistency(snap)
	for _, r := range results {
		if !r.Passed {
			t.Errorf("assertion %q failed: %s", r.Name, r.Detail)
		}
	}
}

func TestAssertChainConsistency_BrokenSymmetry(t *testing.T) {
	snap := &ops.ChainSnapshot[string]{
		Nodes: []ops.NodeSnapshot[string]{
			{ID: "P", FirstChildID: null("A")},
			{ID: "A", ParentID: null("P"), NextSiblingID: null("B")},
			{ID: "B", ParentID: null("P"), PrevSiblingID: null("WRONG")}, // should be A
		},
		ByID: map[string]ops.NodeSnapshot[string]{
			"P": {ID: "P", FirstChildID: null("A")},
			"A": {ID: "A", ParentID: null("P"), NextSiblingID: null("B")},
			"B": {ID: "B", ParentID: null("P"), PrevSiblingID: null("WRONG")},
		},
	}

	results := ops.AssertChainConsistency(snap)
	symmetryResult := findAssertion(results, "prev-next-symmetry")
	if symmetryResult == nil {
		t.Fatal("expected prev-next-symmetry assertion")
	}
	if symmetryResult.Passed {
		t.Error("expected prev-next-symmetry to fail")
	}
}

func TestAssertChainConsistency_NoTermination(t *testing.T) {
	// All siblings have both prev and next (no head/tail)
	snap := &ops.ChainSnapshot[string]{
		Nodes: []ops.NodeSnapshot[string]{
			{ID: "P", FirstChildID: null("A")},
			{ID: "A", ParentID: null("P"), NextSiblingID: null("B"), PrevSiblingID: null("B")},
			{ID: "B", ParentID: null("P"), NextSiblingID: null("A"), PrevSiblingID: null("A")},
		},
		ByID: map[string]ops.NodeSnapshot[string]{
			"P": {ID: "P", FirstChildID: null("A")},
			"A": {ID: "A", ParentID: null("P"), NextSiblingID: null("B"), PrevSiblingID: null("B")},
			"B": {ID: "B", ParentID: null("P"), NextSiblingID: null("A"), PrevSiblingID: null("A")},
		},
	}

	results := ops.AssertChainConsistency(snap)
	terminates := findAssertion(results, "chain-terminates")
	if terminates == nil {
		t.Fatal("expected chain-terminates assertion")
	}
	if terminates.Passed {
		t.Error("expected chain-terminates to fail")
	}
}

func TestAssertChainConsistency_Nil(t *testing.T) {
	results := ops.AssertChainConsistency[string](nil)
	if results != nil {
		t.Errorf("expected nil results for nil snapshot, got: %+v", results)
	}
}

// --- IsChainError tests ---

func TestIsChainError(t *testing.T) {
	report := &ops.OperationReport[string]{
		Violations: []ops.ChainViolation[string]{
			{NodeID: "A", Field: "next_sibling_id", Description: "broken"},
		},
	}
	chainErr := &ops.ChainError[string]{Operation: "move", Report: report}

	if !ops.IsChainError(chainErr) {
		t.Error("expected IsChainError to return true for ChainError")
	}
	if ops.IsChainError(nil) {
		t.Error("expected IsChainError to return false for nil")
	}
	if ops.IsChainError(context.DeadlineExceeded) {
		t.Error("expected IsChainError to return false for non-chain error")
	}
}

// --- helpers ---

func findAssertion(results []ops.AssertionResult, name string) *ops.AssertionResult {
	for i := range results {
		if results[i].Name == name {
			return &results[i]
		}
	}
	return nil
}
