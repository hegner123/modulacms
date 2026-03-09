package ops

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
)

// maxCycleDepth prevents infinite loops in cycle detection.
const maxCycleDepth = 1000

// DetectCycle walks the parent chain from proposedParentID upward. Returns an
// error if nodeID is encountered (the move would create a cycle).
func DetectCycle[ID ~string](ctx context.Context, b Backend[ID], nodeID ID, proposedParentID NullableID[ID]) error {
	if !proposedParentID.Valid {
		return nil // moving to root — no cycle possible
	}

	current := proposedParentID.Value
	for i := range maxCycleDepth {
		if current == nodeID {
			return fmt.Errorf("cycle detected: node %s would become its own ancestor", string(nodeID))
		}
		node, err := b.GetNode(ctx, current)
		if err != nil {
			return fmt.Errorf("cycle detection at depth %d: %w", i, err)
		}
		if !node.ParentID.Valid {
			return nil // reached root — no cycle
		}
		current = node.ParentID.Value
	}
	return fmt.Errorf("cycle detection exceeded max depth %d from node %s", maxCycleDepth, string(nodeID))
}

// Move performs a full tree move: cycle detection, unlink from old position,
// and insert at new position under the new parent.
//
// If NewParentID is not valid, the node is moved to root level (unlinked
// with parent/sibling pointers cleared).
func Move[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], params MoveParams[ID]) (*MoveResult[ID], error) {
	// Fetch the node being moved
	node, err := b.GetNode(ctx, params.NodeID)
	if err != nil {
		return nil, fmt.Errorf("move: get node: %w", err)
	}

	oldParentID := node.ParentID

	// Compute old position by walking from parent's first child
	oldPosition := computePosition(ctx, b, node)

	// Cycle detection
	if err := DetectCycle(ctx, b, params.NodeID, params.NewParentID); err != nil {
		return nil, err
	}

	// Unlink from current position
	if err := Unlink(ctx, ac, b, node); err != nil {
		return nil, fmt.Errorf("move: %w", err)
	}

	// Insert at new position
	if params.NewParentID.Valid {
		if err := InsertAt(ctx, ac, b, params.NewParentID.Value, params.NodeID, params.Position); err != nil {
			return nil, fmt.Errorf("move: %w", err)
		}
	}
	// If NewParentID is not valid, Unlink already cleared the pointers.
	// The node is now a root-level orphan.

	return &MoveResult[ID]{
		OldParentID: oldParentID,
		NewParentID: params.NewParentID,
		OldPosition: oldPosition,
		NewPosition: params.Position,
	}, nil
}

// computePosition walks from parent's first child to determine the 0-based
// position of node in its sibling chain. Returns -1 if not found or no parent.
func computePosition[ID ~string](ctx context.Context, b Backend[ID], node *Node[ID]) int {
	if !node.ParentID.Valid {
		return -1
	}
	parent, err := b.GetNode(ctx, node.ParentID.Value)
	if err != nil || !parent.FirstChildID.Valid {
		return -1
	}

	current := parent.FirstChildID.Value
	for pos := 0; pos < maxCycleDepth; pos++ {
		if current == node.ID {
			return pos
		}
		cur, err := b.GetNode(ctx, current)
		if err != nil || !cur.NextSiblingID.Valid {
			return -1
		}
		current = cur.NextSiblingID.Value
	}
	return -1
}
