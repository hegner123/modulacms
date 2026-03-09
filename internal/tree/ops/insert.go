package ops

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
)

// AppendChild inserts childID as the last child of parentID.
// If parentID has no children, childID becomes the first child.
// Otherwise, walks the sibling chain to the end and appends.
func AppendChild[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], parentID, childID ID) error {
	parent, err := b.GetNode(ctx, parentID)
	if err != nil {
		return fmt.Errorf("append child: get parent: %w", err)
	}

	if !parent.FirstChildID.Valid {
		// No children — set child as first child of parent
		err = b.UpdatePointers(ctx, ac, parent.ID, Pointers[ID]{
			ParentID:      parent.ParentID,
			FirstChildID:  NullID(childID),
			NextSiblingID: parent.NextSiblingID,
			PrevSiblingID: parent.PrevSiblingID,
		})
		if err != nil {
			return fmt.Errorf("append child: update parent first_child: %w", err)
		}
		// Set child's parent pointer
		child, cErr := b.GetNode(ctx, childID)
		if cErr != nil {
			return fmt.Errorf("append child: get child: %w", cErr)
		}
		return b.UpdatePointers(ctx, ac, childID, Pointers[ID]{
			ParentID:      NullID(parentID),
			FirstChildID:  child.FirstChildID,
			NextSiblingID: EmptyID[ID](),
			PrevSiblingID: EmptyID[ID](),
		})
	}

	// Walk to end of sibling chain
	current, err := b.GetNode(ctx, parent.FirstChildID.Value)
	if err != nil {
		return fmt.Errorf("append child: get first child: %w", err)
	}
	for current.NextSiblingID.Valid {
		current, err = b.GetNode(ctx, current.NextSiblingID.Value)
		if err != nil {
			return fmt.Errorf("append child: walk siblings: %w", err)
		}
	}

	// Link current (last sibling) -> child
	err = b.UpdatePointers(ctx, ac, current.ID, Pointers[ID]{
		ParentID:      current.ParentID,
		FirstChildID:  current.FirstChildID,
		NextSiblingID: NullID(childID),
		PrevSiblingID: current.PrevSiblingID,
	})
	if err != nil {
		return fmt.Errorf("append child: update last sibling next: %w", err)
	}

	// Set child's pointers
	child, err := b.GetNode(ctx, childID)
	if err != nil {
		return fmt.Errorf("append child: get child: %w", err)
	}
	return b.UpdatePointers(ctx, ac, childID, Pointers[ID]{
		ParentID:      NullID(parentID),
		FirstChildID:  child.FirstChildID,
		NextSiblingID: EmptyID[ID](),
		PrevSiblingID: NullID(current.ID),
	})
}

// InsertAt inserts childID at the given position under parentID.
// Position 0 inserts as the first child. Positions beyond the end
// of the sibling chain append to the end.
func InsertAt[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], parentID, childID ID, position int) error {
	parent, err := b.GetNode(ctx, parentID)
	if err != nil {
		return fmt.Errorf("insert at: get parent: %w", err)
	}

	// Empty parent or position 0 — insert as first child
	if !parent.FirstChildID.Valid || position == 0 {
		oldFirstChildID := parent.FirstChildID

		// Update parent's first_child to the new child
		err = b.UpdatePointers(ctx, ac, parent.ID, Pointers[ID]{
			ParentID:      parent.ParentID,
			FirstChildID:  NullID(childID),
			NextSiblingID: parent.NextSiblingID,
			PrevSiblingID: parent.PrevSiblingID,
		})
		if err != nil {
			return fmt.Errorf("insert at: update parent first_child: %w", err)
		}

		// Update old first child's prev pointer
		if oldFirstChildID.Valid {
			oldFirst, ofErr := b.GetNode(ctx, oldFirstChildID.Value)
			if ofErr != nil {
				return fmt.Errorf("insert at: get old first child: %w", ofErr)
			}
			ofErr = b.UpdatePointers(ctx, ac, oldFirst.ID, Pointers[ID]{
				ParentID:      oldFirst.ParentID,
				FirstChildID:  oldFirst.FirstChildID,
				NextSiblingID: oldFirst.NextSiblingID,
				PrevSiblingID: NullID(childID),
			})
			if ofErr != nil {
				return fmt.Errorf("insert at: update old first child prev: %w", ofErr)
			}
		}

		// Set child's pointers
		child, cErr := b.GetNode(ctx, childID)
		if cErr != nil {
			return fmt.Errorf("insert at: get child: %w", cErr)
		}
		return b.UpdatePointers(ctx, ac, childID, Pointers[ID]{
			ParentID:      NullID(parentID),
			FirstChildID:  child.FirstChildID,
			NextSiblingID: oldFirstChildID,
			PrevSiblingID: EmptyID[ID](),
		})
	}

	// Walk to position-1 to find the node to insert after
	current, err := b.GetNode(ctx, parent.FirstChildID.Value)
	if err != nil {
		return fmt.Errorf("insert at: get first child: %w", err)
	}
	for i := 0; i < position-1 && current.NextSiblingID.Valid; i++ {
		current, err = b.GetNode(ctx, current.NextSiblingID.Value)
		if err != nil {
			return fmt.Errorf("insert at: walk siblings: %w", err)
		}
	}

	// Insert after current
	oldNext := current.NextSiblingID

	// Update current's next to point to child
	err = b.UpdatePointers(ctx, ac, current.ID, Pointers[ID]{
		ParentID:      current.ParentID,
		FirstChildID:  current.FirstChildID,
		NextSiblingID: NullID(childID),
		PrevSiblingID: current.PrevSiblingID,
	})
	if err != nil {
		return fmt.Errorf("insert at: update insert-after node next: %w", err)
	}

	// Update the node after current's prev to point to child
	if oldNext.Valid {
		afterNode, anErr := b.GetNode(ctx, oldNext.Value)
		if anErr != nil {
			return fmt.Errorf("insert at: get after-insertion node: %w", anErr)
		}
		anErr = b.UpdatePointers(ctx, ac, afterNode.ID, Pointers[ID]{
			ParentID:      afterNode.ParentID,
			FirstChildID:  afterNode.FirstChildID,
			NextSiblingID: afterNode.NextSiblingID,
			PrevSiblingID: NullID(childID),
		})
		if anErr != nil {
			return fmt.Errorf("insert at: update after-insertion node prev: %w", anErr)
		}
	}

	// Set child's pointers
	child, err := b.GetNode(ctx, childID)
	if err != nil {
		return fmt.Errorf("insert at: get child: %w", err)
	}
	return b.UpdatePointers(ctx, ac, childID, Pointers[ID]{
		ParentID:      NullID(parentID),
		FirstChildID:  child.FirstChildID,
		NextSiblingID: oldNext,
		PrevSiblingID: NullID(current.ID),
	})
}
