package tui

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
)

// =============================================================================
// TREE OPERATIONS ABSTRACTION
// =============================================================================
//
// treeOps provides parameterized tree pointer manipulation that works for both
// regular and admin content. Instead of duplicating detach/attach/swap for each
// content type, we normalize tree pointers to strings and parameterize the
// get/update operations via function fields.

// treeNullID is a nullable node ID normalized to string.
type treeNullID struct {
	ID    string
	Valid bool
}

func (n treeNullID) isEmpty() bool {
	return !n.Valid || n.ID == ""
}

// treeNode holds normalized tree pointers and the original fetched struct.
// The source field holds the original *ContentData or *AdminContentData,
// allowing the updateNode function to rebuild full update params while
// preserving all non-pointer fields.
type treeNode struct {
	id            string
	parentID      treeNullID
	firstChildID  treeNullID
	nextSiblingID treeNullID
	prevSiblingID treeNullID
	source        any
}

// treeOps provides parameterized get/update operations for tree nodes.
type treeOps struct {
	getNode    func(id string) (*treeNode, error)
	updateNode func(ctx context.Context, ac audited.AuditContext, node *treeNode) error
}

// newContentTreeOps creates tree operations for regular content.
func newContentTreeOps(d db.DbDriver) treeOps {
	return treeOps{
		getNode: func(id string) (*treeNode, error) {
			c, err := d.GetContentData(types.ContentID(id))
			if err != nil {
				return nil, err
			}
			if c == nil {
				return nil, nil
			}
			return contentToTreeNode(c), nil
		},
		updateNode: func(ctx context.Context, ac audited.AuditContext, node *treeNode) error {
			c := node.source.(*db.ContentData)
			_, err := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: c.ContentDataID,
				RouteID:       c.RouteID,
				ParentID:      types.NullableContentID{ID: types.ContentID(node.parentID.ID), Valid: node.parentID.Valid},
				FirstChildID:  types.NullableContentID{ID: types.ContentID(node.firstChildID.ID), Valid: node.firstChildID.Valid},
				NextSiblingID: types.NullableContentID{ID: types.ContentID(node.nextSiblingID.ID), Valid: node.nextSiblingID.Valid},
				PrevSiblingID: types.NullableContentID{ID: types.ContentID(node.prevSiblingID.ID), Valid: node.prevSiblingID.Valid},
				DatatypeID:    c.DatatypeID,
				AuthorID:      c.AuthorID,
				Status:        c.Status,
				DateCreated:   c.DateCreated,
				DateModified:  types.TimestampNow(),
			})
			return err
		},
	}
}

// newAdminTreeOps creates tree operations for admin content.
func newAdminTreeOps(d db.DbDriver) treeOps {
	return treeOps{
		getNode: func(id string) (*treeNode, error) {
			c, err := d.GetAdminContentData(types.AdminContentID(id))
			if err != nil {
				return nil, err
			}
			if c == nil {
				return nil, nil
			}
			return adminContentToTreeNode(c), nil
		},
		updateNode: func(ctx context.Context, ac audited.AuditContext, node *treeNode) error {
			c := node.source.(*db.AdminContentData)
			_, err := d.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
				AdminContentDataID: c.AdminContentDataID,
				AdminRouteID:       c.AdminRouteID,
				AdminDatatypeID:    c.AdminDatatypeID,
				ParentID:           types.NullableAdminContentID{ID: types.AdminContentID(node.parentID.ID), Valid: node.parentID.Valid},
				FirstChildID:       types.NullableAdminContentID{ID: types.AdminContentID(node.firstChildID.ID), Valid: node.firstChildID.Valid},
				NextSiblingID:      types.NullableAdminContentID{ID: types.AdminContentID(node.nextSiblingID.ID), Valid: node.nextSiblingID.Valid},
				PrevSiblingID:      types.NullableAdminContentID{ID: types.AdminContentID(node.prevSiblingID.ID), Valid: node.prevSiblingID.Valid},
				AuthorID:           c.AuthorID,
				Status:             c.Status,
				DateCreated:        c.DateCreated,
				DateModified:       types.TimestampNow(),
			})
			return err
		},
	}
}

// contentToTreeNode converts a ContentData to a normalized treeNode.
func contentToTreeNode(c *db.ContentData) *treeNode {
	return &treeNode{
		id:            string(c.ContentDataID),
		parentID:      treeNullID{ID: string(c.ParentID.ID), Valid: c.ParentID.Valid},
		firstChildID:  treeNullID{ID: string(c.FirstChildID.ID), Valid: c.FirstChildID.Valid},
		nextSiblingID: treeNullID{ID: string(c.NextSiblingID.ID), Valid: c.NextSiblingID.Valid},
		prevSiblingID: treeNullID{ID: string(c.PrevSiblingID.ID), Valid: c.PrevSiblingID.Valid},
		source:        c,
	}
}

// adminContentToTreeNode converts an AdminContentData to a normalized treeNode.
func adminContentToTreeNode(c *db.AdminContentData) *treeNode {
	return &treeNode{
		id:            string(c.AdminContentDataID),
		parentID:      treeNullID{ID: string(c.ParentID.ID), Valid: c.ParentID.Valid},
		firstChildID:  treeNullID{ID: string(c.FirstChildID.ID), Valid: c.FirstChildID.Valid},
		nextSiblingID: treeNullID{ID: string(c.NextSiblingID.ID), Valid: c.NextSiblingID.Valid},
		prevSiblingID: treeNullID{ID: string(c.PrevSiblingID.ID), Valid: c.PrevSiblingID.Valid},
		source:        c,
	}
}

// =============================================================================
// UNIFIED TREE POINTER OPERATIONS
// =============================================================================
//
// These three functions replace the 6 duplicated regular/admin variants
// previously in tree_ops.go. They operate on the normalized treeNode type
// and use treeOps function fields for database access.
//
// Concurrency warning: Multi-step tree mutations perform read-then-write
// sequences without transaction isolation. Two concurrent SSH sessions editing
// the same content tree can interleave operations and corrupt sibling pointers.

// detachFromSiblings removes a node from its sibling chain.
// Updates prev.next, next.prev, and parent.firstChild as needed. Each affected
// node is fetched before update to preserve all non-pointer fields (Golden Rule).
func detachFromSiblings(ctx context.Context, ac audited.AuditContext, ops treeOps, node *treeNode) []error {
	var errs []error

	// If this node has a previous sibling, point its next to our next
	if !node.prevSiblingID.isEmpty() {
		prev, err := ops.getNode(node.prevSiblingID.ID)
		if err == nil && prev != nil {
			prev.nextSiblingID = node.nextSiblingID
			if updateErr := ops.updateNode(ctx, ac, prev); updateErr != nil {
				errs = append(errs, fmt.Errorf("update prev sibling: %w", updateErr))
			}
		}
	}

	// If this node has a next sibling, point its prev to our prev
	if !node.nextSiblingID.isEmpty() {
		next, err := ops.getNode(node.nextSiblingID.ID)
		if err == nil && next != nil {
			next.prevSiblingID = node.prevSiblingID
			if updateErr := ops.updateNode(ctx, ac, next); updateErr != nil {
				errs = append(errs, fmt.Errorf("update next sibling: %w", updateErr))
			}
		}
	}

	// If this is the first child of parent, update parent's first_child to our next sibling
	if !node.parentID.isEmpty() {
		parent, err := ops.getNode(node.parentID.ID)
		if err == nil && parent != nil {
			if !parent.firstChildID.isEmpty() && parent.firstChildID.ID == node.id {
				parent.firstChildID = node.nextSiblingID
				if updateErr := ops.updateNode(ctx, ac, parent); updateErr != nil {
					errs = append(errs, fmt.Errorf("update parent first_child: %w", updateErr))
				}
			}
		}
	}

	return errs
}

// attachAsLastChild attaches sourceID as the last child of targetID.
// If target has no children, sourceID becomes the first child and an empty
// treeNullID is returned. If target has children, the sibling chain is walked
// to find the last child, whose NextSiblingID is set to sourceID, and that
// last child's ID is returned (for setting source.PrevSiblingID).
func attachAsLastChild(ctx context.Context, ac audited.AuditContext, ops treeOps, sourceID string, targetID string) (treeNullID, error) {
	target, err := ops.getNode(targetID)
	if err != nil || target == nil {
		return treeNullID{}, fmt.Errorf("target content not found: %w", err)
	}

	if target.firstChildID.isEmpty() {
		// Target has no children -- source becomes first child
		target.firstChildID = treeNullID{ID: sourceID, Valid: true}
		if updateErr := ops.updateNode(ctx, ac, target); updateErr != nil {
			return treeNullID{}, fmt.Errorf("failed to set target first_child: %w", updateErr)
		}
		return treeNullID{}, nil
	}

	// Target has children -- walk to last sibling
	currentID := target.firstChildID.ID
	for {
		current, walkErr := ops.getNode(currentID)
		if walkErr != nil || current == nil {
			return treeNullID{}, fmt.Errorf("failed to walk sibling chain at %s: %w", currentID, walkErr)
		}
		if current.nextSiblingID.isEmpty() {
			// Found last sibling -- update its NextSiblingID to source
			current.nextSiblingID = treeNullID{ID: sourceID, Valid: true}
			if updateErr := ops.updateNode(ctx, ac, current); updateErr != nil {
				return treeNullID{}, fmt.Errorf("failed to update last sibling next pointer: %w", updateErr)
			}
			return treeNullID{ID: current.id, Valid: true}, nil
		}
		currentID = current.nextSiblingID.ID
	}
}

// spliceAfter inserts newID as the next sibling of node.
// Updates node.next -> newID and, if node had a next sibling, updates
// that old next sibling's prev -> newID.
func spliceAfter(ctx context.Context, ac audited.AuditContext, ops treeOps, node *treeNode, newID string) error {
	origNext := node.nextSiblingID

	// Update node's next -> new
	node.nextSiblingID = treeNullID{ID: newID, Valid: true}
	if err := ops.updateNode(ctx, ac, node); err != nil {
		return fmt.Errorf("update source next pointer: %w", err)
	}

	// If had a next sibling, update its prev -> new
	if !origNext.isEmpty() {
		d, err := ops.getNode(origNext.ID)
		if err == nil && d != nil {
			d.prevSiblingID = treeNullID{ID: newID, Valid: true}
			if updateErr := ops.updateNode(ctx, ac, d); updateErr != nil {
				return fmt.Errorf("update next sibling prev pointer: %w", updateErr)
			}
		}
	}

	return nil
}

// swapSiblings swaps two adjacent content siblings in the sibling chain.
// For direction "up": Before: [C?] <-> B <-> A <-> [D?], After: [C?] <-> A <-> B <-> [D?]
//   - a is the node moving up, b is a's previous sibling
//
// For direction "down": Before: [C?] <-> A <-> B <-> [D?], After: [C?] <-> B <-> A <-> [D?]
//   - a is the node moving down, b is a's next sibling
//
// Each affected node is fetched before update to preserve all non-pointer fields.
func swapSiblings(ctx context.Context, ac audited.AuditContext, ops treeOps, a *treeNode, b *treeNode, direction string) []error {
	var errs []error

	// Capture original pointer values before any modifications
	origAPrev := a.prevSiblingID
	origANext := a.nextSiblingID
	origBPrev := b.prevSiblingID
	origBNext := b.nextSiblingID

	if direction == "up" {
		// Before: [C?] <-> B <-> A <-> [D?]
		// After:  [C?] <-> A <-> B <-> [D?]

		// If B has a prev (C), update C.NextSiblingID -> A
		if !origBPrev.isEmpty() {
			c, cErr := ops.getNode(origBPrev.ID)
			if cErr == nil && c != nil {
				c.nextSiblingID = treeNullID{ID: a.id, Valid: true}
				if updateErr := ops.updateNode(ctx, ac, c); updateErr != nil {
					errs = append(errs, fmt.Errorf("update C next sibling: %w", updateErr))
				}
			}
		}

		// If A has a next (D), update D.PrevSiblingID -> B
		if !origANext.isEmpty() {
			d, dErr := ops.getNode(origANext.ID)
			if dErr == nil && d != nil {
				d.prevSiblingID = treeNullID{ID: b.id, Valid: true}
				if updateErr := ops.updateNode(ctx, ac, d); updateErr != nil {
					errs = append(errs, fmt.Errorf("update D prev sibling: %w", updateErr))
				}
			}
		}

		// If parent.FirstChildID == B, update to A
		if !a.parentID.isEmpty() {
			parent, pErr := ops.getNode(a.parentID.ID)
			if pErr == nil && parent != nil && !parent.firstChildID.isEmpty() && parent.firstChildID.ID == b.id {
				parent.firstChildID = treeNullID{ID: a.id, Valid: true}
				if updateErr := ops.updateNode(ctx, ac, parent); updateErr != nil {
					errs = append(errs, fmt.Errorf("update parent first_child: %w", updateErr))
				}
			}
		}

		if len(errs) > 0 {
			return errs
		}

		// Update A: prev = B's old prev, next = B
		a.prevSiblingID = origBPrev
		a.nextSiblingID = treeNullID{ID: b.id, Valid: true}
		if aErr := ops.updateNode(ctx, ac, a); aErr != nil {
			errs = append(errs, fmt.Errorf("update A: %w", aErr))
		}

		// Update B: prev = A, next = A's old next
		b.prevSiblingID = treeNullID{ID: a.id, Valid: true}
		b.nextSiblingID = origANext
		if bErr := ops.updateNode(ctx, ac, b); bErr != nil {
			errs = append(errs, fmt.Errorf("update B: %w", bErr))
		}

	} else {
		// direction == "down"
		// Before: [C?] <-> A <-> B <-> [D?]
		// After:  [C?] <-> B <-> A <-> [D?]

		// If A has a prev (C), update C.NextSiblingID -> B
		if !origAPrev.isEmpty() {
			c, cErr := ops.getNode(origAPrev.ID)
			if cErr == nil && c != nil {
				c.nextSiblingID = treeNullID{ID: b.id, Valid: true}
				if updateErr := ops.updateNode(ctx, ac, c); updateErr != nil {
					errs = append(errs, fmt.Errorf("update C next sibling: %w", updateErr))
				}
			}
		}

		// If B has a next (D), update D.PrevSiblingID -> A
		if !origBNext.isEmpty() {
			d, dErr := ops.getNode(origBNext.ID)
			if dErr == nil && d != nil {
				d.prevSiblingID = treeNullID{ID: a.id, Valid: true}
				if updateErr := ops.updateNode(ctx, ac, d); updateErr != nil {
					errs = append(errs, fmt.Errorf("update D prev sibling: %w", updateErr))
				}
			}
		}

		// If parent.FirstChildID == A, update to B
		if !a.parentID.isEmpty() {
			parent, pErr := ops.getNode(a.parentID.ID)
			if pErr == nil && parent != nil && !parent.firstChildID.isEmpty() && parent.firstChildID.ID == a.id {
				parent.firstChildID = treeNullID{ID: b.id, Valid: true}
				if updateErr := ops.updateNode(ctx, ac, parent); updateErr != nil {
					errs = append(errs, fmt.Errorf("update parent first_child: %w", updateErr))
				}
			}
		}

		if len(errs) > 0 {
			return errs
		}

		// Update B: prev = A's old prev, next = A
		b.prevSiblingID = origAPrev
		b.nextSiblingID = treeNullID{ID: a.id, Valid: true}
		if bErr := ops.updateNode(ctx, ac, b); bErr != nil {
			errs = append(errs, fmt.Errorf("update B: %w", bErr))
		}

		// Update A: prev = B, next = B's old next
		a.prevSiblingID = treeNullID{ID: b.id, Valid: true}
		a.nextSiblingID = origBNext
		if aErr := ops.updateNode(ctx, ac, a); aErr != nil {
			errs = append(errs, fmt.Errorf("update A: %w", aErr))
		}
	}

	return errs
}
