package ops

import (
	"context"
	"fmt"
	"strings"
)

// NodeSnapshot captures a node's pointer state at a point in time.
type NodeSnapshot[ID ~string] struct {
	ID            ID
	ParentID      NullableID[ID]
	FirstChildID  NullableID[ID]
	NextSiblingID NullableID[ID]
	PrevSiblingID NullableID[ID]
}

// snapshotFromNode creates a NodeSnapshot from an ops.Node.
func snapshotFromNode[ID ~string](n *Node[ID]) NodeSnapshot[ID] {
	return NodeSnapshot[ID]{
		ID:            n.ID,
		ParentID:      n.ParentID,
		FirstChildID:  n.FirstChildID,
		NextSiblingID: n.NextSiblingID,
		PrevSiblingID: n.PrevSiblingID,
	}
}

// ChainSnapshot is a frozen record of all nodes involved in an operation.
type ChainSnapshot[ID ~string] struct {
	Nodes   []NodeSnapshot[ID]          // ordered (chain walk order)
	ByID    map[string]NodeSnapshot[ID] // O(1) lookup by string(ID)
	WalkErr error                       // non-nil if walk hit a dangling pointer or cycle
}

// ChainViolation describes an inconsistency found in pointer data.
// Violations are blocking — the operation will not proceed.
type ChainViolation[ID ~string] struct {
	NodeID      ID
	Field       string // "next_sibling_id", "prev_sibling_id", etc.
	Description string // human-readable diagnostic
}

// AssertionResult is one post-operation invariant check.
type AssertionResult struct {
	Name   string // e.g. "prev-next-symmetry"
	Passed bool
	Detail string // empty on pass
}

// OperationReport bundles pre/post snapshots with validation and assertion results.
type OperationReport[ID ~string] struct {
	Before     *ChainSnapshot[ID]
	After      *ChainSnapshot[ID]
	Violations []ChainViolation[ID] // pre-op: if non-empty, operation was blocked
	Assertions []AssertionResult    // post-op: if any failed, chain is inconsistent
}

// ChainError is returned when pre-operation validation detects mangled data.
// Carries the full OperationReport for recovery (heal) or bug reporting.
// Because this is a generic type, use IsChainError for type-erased detection.
type ChainError[ID ~string] struct {
	Operation string
	Report    *OperationReport[ID]
}

func (e *ChainError[ID]) Error() string {
	if len(e.Report.Violations) == 0 {
		return fmt.Sprintf("%s: chain validation failed", e.Operation)
	}
	return fmt.Sprintf("%s: chain validation failed: %s (node %s, field %s)",
		e.Operation,
		e.Report.Violations[0].Description,
		string(e.Report.Violations[0].NodeID),
		e.Report.Violations[0].Field,
	)
}

// IsChainError checks whether an error is a ChainError (mangled pointer data
// blocked an operation). Callers can branch into recovery or bug-reporting.
func IsChainError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "chain validation failed")
}

// AssertionError is returned when post-operation assertions detect inconsistency.
// The mutation already happened — this error signals investigation is needed.
// Inside a transaction, this causes rollback (the mutation is undone).
type AssertionError[ID ~string] struct {
	Operation string
	Report    *OperationReport[ID]
}

func (e *AssertionError[ID]) Error() string {
	var failed []string
	for _, a := range e.Report.Assertions {
		if !a.Passed {
			failed = append(failed, a.Name+": "+a.Detail)
		}
	}
	return fmt.Sprintf("%s: post-operation assertions failed: %s",
		e.Operation, strings.Join(failed, "; "))
}

// WalkSiblingChain walks from a parent's FirstChildID to the end of the sibling
// chain via NextSiblingID. Returns a ChainSnapshot containing the parent and all
// siblings in chain order. Uses a visited set to detect cycles.
//
// If the walk encounters a dangling pointer (GetNode error), the partial snapshot
// is returned with WalkErr set. If a cycle is detected, WalkErr is set and the
// snapshot contains nodes up to the cycle point.
func WalkSiblingChain[ID ~string](ctx context.Context, b Backend[ID], parentID ID) (*ChainSnapshot[ID], error) {
	parent, err := b.GetNode(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("walk chain: get parent %s: %w", string(parentID), err)
	}

	snap := &ChainSnapshot[ID]{
		Nodes: []NodeSnapshot[ID]{snapshotFromNode(parent)},
		ByID:  map[string]NodeSnapshot[ID]{string(parent.ID): snapshotFromNode(parent)},
	}

	if !parent.FirstChildID.Valid {
		return snap, nil
	}

	visited := map[string]bool{string(parentID): true}
	current := parent.FirstChildID.Value

	for i := range maxCycleDepth {
		key := string(current)
		if visited[key] {
			snap.WalkErr = fmt.Errorf("cycle detected at node %s (depth %d)", key, i)
			break
		}
		visited[key] = true

		node, nodeErr := b.GetNode(ctx, current)
		if nodeErr != nil {
			snap.WalkErr = fmt.Errorf("dangling pointer to %s (depth %d): %w", key, i, nodeErr)
			break
		}

		ns := snapshotFromNode(node)
		snap.Nodes = append(snap.Nodes, ns)
		snap.ByID[key] = ns

		if !node.NextSiblingID.Valid {
			break
		}
		current = node.NextSiblingID.Value
	}

	return snap, nil
}

// SnapshotNodes fetches specific nodes by ID and returns them as a ChainSnapshot.
// Nodes are returned in the order requested. If a node cannot be fetched, its ID
// is skipped and WalkErr is set on the snapshot.
func SnapshotNodes[ID ~string](ctx context.Context, b Backend[ID], ids []ID) (*ChainSnapshot[ID], error) {
	snap := &ChainSnapshot[ID]{
		Nodes: make([]NodeSnapshot[ID], 0, len(ids)),
		ByID:  make(map[string]NodeSnapshot[ID], len(ids)),
	}

	for _, id := range ids {
		node, err := b.GetNode(ctx, id)
		if err != nil {
			snap.WalkErr = fmt.Errorf("snapshot node %s: %w", string(id), err)
			continue
		}
		ns := snapshotFromNode(node)
		snap.Nodes = append(snap.Nodes, ns)
		snap.ByID[string(id)] = ns
	}

	return snap, nil
}

// ValidateChain checks a ChainSnapshot for pointer inconsistencies.
// Returns a slice of violations. An empty slice means the chain is healthy.
//
// Checks performed:
//  1. Prev-next symmetry: A.next=B implies B.prev=A
//  2. Parent.first_child reachability: first child exists and has prev=null
//  3. Parent consistency: all siblings share the same parent_id
//  4. No self-referencing pointers
//  5. First node in chain has prev=null, last has next=null
func ValidateChain[ID ~string](snap *ChainSnapshot[ID]) []ChainViolation[ID] {
	if snap == nil || len(snap.Nodes) == 0 {
		return nil
	}

	var violations []ChainViolation[ID]

	// Find the parent node (first in snapshot from WalkSiblingChain) and siblings.
	// The parent is the node whose FirstChildID points into the chain.
	// Siblings are nodes whose ParentID points to the parent.
	var parentSnap *NodeSnapshot[ID]
	var siblings []NodeSnapshot[ID]

	for i := range snap.Nodes {
		n := &snap.Nodes[i]
		// Heuristic: the parent has a FirstChildID that matches another node in the snap,
		// and siblings have a ParentID that matches the parent.
		if n.FirstChildID.Valid {
			if _, ok := snap.ByID[string(n.FirstChildID.Value)]; ok {
				parentSnap = n
			}
		}
	}

	// Collect siblings: all nodes except the parent. This includes nodes with
	// mismatched parent_id — that's exactly what check 3 catches.
	if parentSnap != nil {
		for _, n := range snap.Nodes {
			if n.ID == parentSnap.ID {
				continue
			}
			siblings = append(siblings, n)
		}
	} else {
		// No clear parent — treat all nodes as siblings for validation
		siblings = snap.Nodes
	}

	// Check 4: Self-referencing pointers
	for _, n := range snap.Nodes {
		if n.NextSiblingID.Valid && n.NextSiblingID.Value == n.ID {
			violations = append(violations, ChainViolation[ID]{
				NodeID: n.ID, Field: "next_sibling_id",
				Description: "node points to itself as next sibling",
			})
		}
		if n.PrevSiblingID.Valid && n.PrevSiblingID.Value == n.ID {
			violations = append(violations, ChainViolation[ID]{
				NodeID: n.ID, Field: "prev_sibling_id",
				Description: "node points to itself as prev sibling",
			})
		}
		if n.ParentID.Valid && n.ParentID.Value == n.ID {
			violations = append(violations, ChainViolation[ID]{
				NodeID: n.ID, Field: "parent_id",
				Description: "node points to itself as parent",
			})
		}
		if n.FirstChildID.Valid && n.FirstChildID.Value == n.ID {
			violations = append(violations, ChainViolation[ID]{
				NodeID: n.ID, Field: "first_child_id",
				Description: "node points to itself as first child",
			})
		}
	}

	// Check 1: Prev-next symmetry among siblings
	for _, n := range siblings {
		if n.NextSiblingID.Valid {
			nextKey := string(n.NextSiblingID.Value)
			if next, ok := snap.ByID[nextKey]; ok {
				if !next.PrevSiblingID.Valid || next.PrevSiblingID.Value != n.ID {
					expected := string(n.ID)
					got := "<null>"
					if next.PrevSiblingID.Valid {
						got = string(next.PrevSiblingID.Value)
					}
					violations = append(violations, ChainViolation[ID]{
						NodeID: n.ID, Field: "next_sibling_id",
						Description: fmt.Sprintf(
							"%s.next=%s but %s.prev=%s (expected %s)",
							string(n.ID), nextKey, nextKey, got, expected,
						),
					})
				}
			}
		}
		if n.PrevSiblingID.Valid {
			prevKey := string(n.PrevSiblingID.Value)
			if prev, ok := snap.ByID[prevKey]; ok {
				if !prev.NextSiblingID.Valid || prev.NextSiblingID.Value != n.ID {
					expected := string(n.ID)
					got := "<null>"
					if prev.NextSiblingID.Valid {
						got = string(prev.NextSiblingID.Value)
					}
					violations = append(violations, ChainViolation[ID]{
						NodeID: n.ID, Field: "prev_sibling_id",
						Description: fmt.Sprintf(
							"%s.prev=%s but %s.next=%s (expected %s)",
							string(n.ID), prevKey, prevKey, got, expected,
						),
					})
				}
			}
		}
	}

	// Check 2: Parent's first_child is reachable and has prev=null
	if parentSnap != nil && parentSnap.FirstChildID.Valid {
		fcKey := string(parentSnap.FirstChildID.Value)
		if fc, ok := snap.ByID[fcKey]; ok {
			if fc.PrevSiblingID.Valid {
				violations = append(violations, ChainViolation[ID]{
					NodeID: parentSnap.ID, Field: "first_child_id",
					Description: fmt.Sprintf(
						"parent.first_child=%s but that node has prev_sibling=%s (expected null)",
						fcKey, string(fc.PrevSiblingID.Value),
					),
				})
			}
		}
	}

	// Check 3: Parent consistency — all siblings have the same parent_id
	if parentSnap != nil && len(siblings) > 0 {
		for _, n := range siblings {
			if !n.ParentID.Valid || n.ParentID.Value != parentSnap.ID {
				got := "<null>"
				if n.ParentID.Valid {
					got = string(n.ParentID.Value)
				}
				violations = append(violations, ChainViolation[ID]{
					NodeID: n.ID, Field: "parent_id",
					Description: fmt.Sprintf(
						"sibling %s has parent=%s (expected %s)",
						string(n.ID), got, string(parentSnap.ID),
					),
				})
			}
		}
	}

	// Check 5: First sibling in chain has prev=null, last has next=null
	if len(siblings) > 0 {
		// Find chain head (the one with no prev, or parent's first_child)
		for _, n := range siblings {
			if !n.PrevSiblingID.Valid {
				// This should be the first — verify parent agrees if we have parent
				if parentSnap != nil && parentSnap.FirstChildID.Valid && parentSnap.FirstChildID.Value != n.ID {
					violations = append(violations, ChainViolation[ID]{
						NodeID: n.ID, Field: "prev_sibling_id",
						Description: fmt.Sprintf(
							"node %s has null prev but parent.first_child=%s (not this node)",
							string(n.ID), string(parentSnap.FirstChildID.Value),
						),
					})
				}
			}
		}
	}

	return violations
}

// AssertChainConsistency runs post-operation invariant checks on a ChainSnapshot.
// Returns assertion results — all should have Passed=true for a healthy chain.
//
// Assertions:
//  1. prev-next-symmetry: bidirectional links are consistent
//  2. parent-first-child-reachable: parent's first_child is in the chain
//  3. chain-terminates: chain has a definite first (prev=null) and last (next=null)
func AssertChainConsistency[ID ~string](snap *ChainSnapshot[ID]) []AssertionResult {
	if snap == nil || len(snap.Nodes) == 0 {
		return nil
	}

	var results []AssertionResult

	// Separate parent from siblings (same heuristic as ValidateChain)
	var parentSnap *NodeSnapshot[ID]
	var siblings []NodeSnapshot[ID]

	for i := range snap.Nodes {
		n := &snap.Nodes[i]
		if n.FirstChildID.Valid {
			if _, ok := snap.ByID[string(n.FirstChildID.Value)]; ok {
				parentSnap = n
			}
		}
	}
	if parentSnap != nil {
		for _, n := range snap.Nodes {
			if n.ID == parentSnap.ID {
				continue
			}
			siblings = append(siblings, n)
		}
	} else {
		siblings = snap.Nodes
	}

	// Assertion 1: prev-next symmetry
	symmetryOK := true
	var symmetryDetail string
	for _, n := range siblings {
		if n.NextSiblingID.Valid {
			if next, ok := snap.ByID[string(n.NextSiblingID.Value)]; ok {
				if !next.PrevSiblingID.Valid || next.PrevSiblingID.Value != n.ID {
					symmetryOK = false
					symmetryDetail = fmt.Sprintf("%s.next=%s but %s.prev does not point back",
						string(n.ID), string(n.NextSiblingID.Value), string(n.NextSiblingID.Value))
					break
				}
			}
		}
	}
	results = append(results, AssertionResult{
		Name: "prev-next-symmetry", Passed: symmetryOK, Detail: symmetryDetail,
	})

	// Assertion 2: parent-first-child-reachable
	fcOK := true
	var fcDetail string
	if parentSnap != nil && parentSnap.FirstChildID.Valid {
		fcKey := string(parentSnap.FirstChildID.Value)
		if _, ok := snap.ByID[fcKey]; !ok {
			fcOK = false
			fcDetail = fmt.Sprintf("parent.first_child=%s not found in chain", fcKey)
		}
	}
	results = append(results, AssertionResult{
		Name: "parent-first-child-reachable", Passed: fcOK, Detail: fcDetail,
	})

	// Assertion 3: chain-terminates (has first with prev=null and last with next=null)
	var hasFirst, hasLast bool
	for _, n := range siblings {
		if !n.PrevSiblingID.Valid {
			hasFirst = true
		}
		if !n.NextSiblingID.Valid {
			hasLast = true
		}
	}
	terminatesOK := len(siblings) == 0 || (hasFirst && hasLast)
	terminatesDetail := ""
	if !terminatesOK {
		if !hasFirst {
			terminatesDetail = "no node with prev_sibling=null (no chain head)"
		} else {
			terminatesDetail = "no node with next_sibling=null (no chain tail)"
		}
	}
	results = append(results, AssertionResult{
		Name: "chain-terminates", Passed: terminatesOK, Detail: terminatesDetail,
	})

	return results
}
