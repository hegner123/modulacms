package ops

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
)

// Unlink detaches a node from its sibling chain in 3 steps:
//  1. Repair prev sibling's NextSiblingID (skip over node)
//  2. Repair next sibling's PrevSiblingID (skip over node)
//  3. If node was parent's first child, update parent's FirstChildID
//
// After Unlink, the node's own ParentID/PrevSiblingID/NextSiblingID are
// cleared but FirstChildID is preserved (children stay attached).
func Unlink[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], node *Node[ID]) error {
	// Step 1: repair prev sibling
	if node.PrevSiblingID.Valid {
		prev, err := b.GetNode(ctx, node.PrevSiblingID.Value)
		if err != nil {
			return fmt.Errorf("unlink: get prev sibling: %w", err)
		}
		err = b.UpdatePointers(ctx, ac, prev.ID, Pointers[ID]{
			ParentID:      prev.ParentID,
			FirstChildID:  prev.FirstChildID,
			NextSiblingID: node.NextSiblingID, // skip over node
			PrevSiblingID: prev.PrevSiblingID,
		})
		if err != nil {
			return fmt.Errorf("unlink: update prev sibling: %w", err)
		}
	}

	// Step 2: repair next sibling
	if node.NextSiblingID.Valid {
		next, err := b.GetNode(ctx, node.NextSiblingID.Value)
		if err != nil {
			return fmt.Errorf("unlink: get next sibling: %w", err)
		}
		err = b.UpdatePointers(ctx, ac, next.ID, Pointers[ID]{
			ParentID:      next.ParentID,
			FirstChildID:  next.FirstChildID,
			NextSiblingID: next.NextSiblingID,
			PrevSiblingID: node.PrevSiblingID, // skip over node
		})
		if err != nil {
			return fmt.Errorf("unlink: update next sibling: %w", err)
		}
	}

	// Step 3: repair parent's first_child if needed
	if node.ParentID.Valid {
		parent, err := b.GetNode(ctx, node.ParentID.Value)
		if err != nil {
			return fmt.Errorf("unlink: get parent: %w", err)
		}
		if parent.FirstChildID.Valid && parent.FirstChildID.Value == node.ID {
			err = b.UpdatePointers(ctx, ac, parent.ID, Pointers[ID]{
				ParentID:      parent.ParentID,
				FirstChildID:  node.NextSiblingID, // promote next sibling
				NextSiblingID: parent.NextSiblingID,
				PrevSiblingID: parent.PrevSiblingID,
			})
			if err != nil {
				return fmt.Errorf("unlink: update parent first_child: %w", err)
			}
		}
	}

	// Clear the node's own chain pointers (keep FirstChildID)
	err := b.UpdatePointers(ctx, ac, node.ID, Pointers[ID]{
		ParentID:      EmptyID[ID](),
		FirstChildID:  node.FirstChildID,
		NextSiblingID: EmptyID[ID](),
		PrevSiblingID: EmptyID[ID](),
	})
	if err != nil {
		return fmt.Errorf("unlink: clear node pointers: %w", err)
	}

	return nil
}
