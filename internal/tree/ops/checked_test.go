package ops_test

import (
	"testing"

	"github.com/hegner123/modulacms/internal/tree/ops"
)

// --- UnlinkChecked tests ---

func TestUnlinkChecked_Healthy(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), null("C"), null("A"))
	m.addNode("C", null("P"), empty(), empty(), null("B"))

	node, _ := m.GetNode(ctx, "B")
	report, err := ops.UnlinkChecked(ctx, ac, m, node)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Before == nil || report.After == nil {
		t.Fatal("expected both before and after snapshots")
	}
	if len(report.Violations) > 0 {
		t.Errorf("unexpected violations: %+v", report.Violations)
	}
	for _, a := range report.Assertions {
		if !a.Passed {
			t.Errorf("assertion %q failed: %s", a.Name, a.Detail)
		}
	}

	// Verify B was unlinked
	b, _ := m.GetNode(ctx, "B")
	if b.ParentID.Valid || b.PrevSiblingID.Valid || b.NextSiblingID.Valid {
		t.Error("expected B's parent/prev/next to be cleared")
	}
	// Verify A.next = C (skip over B)
	a, _ := m.GetNode(ctx, "A")
	if !a.NextSiblingID.Valid || a.NextSiblingID.Value != "C" {
		t.Error("expected A.next = C")
	}
}

func TestUnlinkChecked_MangledData_BlocksOperation(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	// B.prev = C instead of A — broken symmetry (A.next=B but B.prev=C)
	m.addNode("B", null("P"), empty(), empty(), null("C"))
	m.addNode("C", null("P"), empty(), empty(), empty())

	node, _ := m.GetNode(ctx, "B")
	report, err := ops.UnlinkChecked(ctx, ac, m, node)
	if !ops.IsChainError(err) {
		t.Fatalf("expected ChainError, got: %v", err)
	}
	if report == nil {
		t.Fatal("expected report even on error")
	}
	if len(report.Violations) == 0 {
		t.Error("expected violations in report")
	}
	if report.After != nil {
		t.Error("expected no after-snapshot when operation is blocked")
	}
}

// --- InsertAtChecked tests ---

func TestInsertAtChecked_Healthy(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), empty(), null("A"))
	m.addNode("C", empty(), empty(), empty(), empty()) // new child to insert

	report, err := ops.InsertAtChecked(ctx, ac, m, "P", "C", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Before == nil || report.After == nil {
		t.Fatal("expected both snapshots")
	}
	for _, a := range report.Assertions {
		if !a.Passed {
			t.Errorf("assertion %q failed: %s", a.Name, a.Detail)
		}
	}

	// Verify C is between A and B
	c, _ := m.GetNode(ctx, "C")
	if !c.PrevSiblingID.Valid || c.PrevSiblingID.Value != "A" {
		t.Error("expected C.prev = A")
	}
	if !c.NextSiblingID.Valid || c.NextSiblingID.Value != "B" {
		t.Error("expected C.next = B")
	}
}

func TestInsertAtChecked_MangledData_BlocksOperation(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("A"), empty()) // self-reference: A.next = A
	m.addNode("C", empty(), empty(), empty(), empty())

	report, err := ops.InsertAtChecked(ctx, ac, m, "P", "C", 0)
	if !ops.IsChainError(err) {
		t.Fatalf("expected ChainError, got: %v", err)
	}
	if report == nil {
		t.Fatal("expected report even on error")
	}
	if len(report.Violations) == 0 {
		t.Error("expected violations")
	}
}

// --- AppendChildChecked tests ---

func TestAppendChildChecked_Healthy(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), empty(), empty())
	m.addNode("C", empty(), empty(), empty(), empty())

	report, err := ops.AppendChildChecked(ctx, ac, m, "P", "C")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Before == nil || report.After == nil {
		t.Fatal("expected both snapshots")
	}

	// Verify C is last child
	c, _ := m.GetNode(ctx, "C")
	if !c.PrevSiblingID.Valid || c.PrevSiblingID.Value != "A" {
		t.Error("expected C.prev = A")
	}
	if c.NextSiblingID.Valid {
		t.Error("expected C.next = null")
	}
}

func TestAppendChildChecked_MangledData_BlocksOperation(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), empty(), null("WRONG")) // B.prev should be A
	m.addNode("C", empty(), empty(), empty(), empty())

	report, err := ops.AppendChildChecked(ctx, ac, m, "P", "C")
	if !ops.IsChainError(err) {
		t.Fatalf("expected ChainError, got: %v", err)
	}
	if report == nil {
		t.Fatal("expected report even on error")
	}
	if len(report.Violations) == 0 {
		t.Error("expected violations")
	}
}

// --- AppendChild cycle guard test ---

func TestAppendChild_CycleInSiblingChain(t *testing.T) {
	m := newMockBackend()
	m.addNode("P", empty(), null("A"), empty(), empty())
	m.addNode("A", null("P"), empty(), null("B"), empty())
	m.addNode("B", null("P"), empty(), null("A"), null("A")) // cycle: B -> A
	m.addNode("C", empty(), empty(), empty(), empty())

	err := ops.AppendChild(ctx, ac, m, "P", "C")
	if err == nil {
		t.Fatal("expected error for cycle in sibling chain")
	}
}
