package router

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// appendToSiblingChain links a newly created content node into its parent's
// sibling chain. If prev/next sibling IDs are provided in createParams, the
// node is spliced at that position; otherwise it is appended to the end.
func appendToSiblingChain(
	ctx context.Context,
	ac audited.AuditContext,
	d db.DbDriver,
	node *db.ContentData,
	reqPrev, reqNext types.NullableContentID,
) (*db.ContentData, error) {
	if !node.ParentID.Valid {
		return node, nil
	}

	now := types.TimestampNow()
	parent, err := d.GetContentData(node.ParentID.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch parent: %w", err)
	}

	hasPosition := reqPrev.Valid || reqNext.Valid

	if !parent.FirstChildID.Valid {
		// Parent has no children — set first_child_id to new node.
		_, err = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: parent.ContentDataID,
			ParentID:      parent.ParentID,
			FirstChildID:  types.NullableContentID{ID: node.ContentDataID, Valid: true},
			NextSiblingID: parent.NextSiblingID,
			PrevSiblingID: parent.PrevSiblingID,
			RouteID:       parent.RouteID,
			DatatypeID:    parent.DatatypeID,
			AuthorID:      parent.AuthorID,
			Status:        parent.Status,
			DateCreated:   parent.DateCreated,
			DateModified:  now,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set parent first_child_id: %w", err)
		}
		return node, nil
	}

	if hasPosition {
		// Splice into specific position in the sibling chain.
		if reqPrev.Valid {
			prev, gErr := d.GetContentData(reqPrev.ID)
			if gErr != nil {
				utility.DefaultLogger.Error("appendToSiblingChain: failed to fetch prev sibling", gErr)
			} else {
				_, uErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
					ContentDataID: prev.ContentDataID,
					ParentID:      prev.ParentID,
					FirstChildID:  prev.FirstChildID,
					NextSiblingID: types.NullableContentID{ID: node.ContentDataID, Valid: true},
					PrevSiblingID: prev.PrevSiblingID,
					RouteID:       prev.RouteID,
					DatatypeID:    prev.DatatypeID,
					AuthorID:      prev.AuthorID,
					Status:        prev.Status,
					DateCreated:   prev.DateCreated,
					DateModified:  now,
				})
				if uErr != nil {
					utility.DefaultLogger.Error("appendToSiblingChain: failed to update prev sibling", uErr)
				}
			}
		} else {
			// Inserting at the start — update parent's first_child_id.
			_, uErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: parent.ContentDataID,
				ParentID:      parent.ParentID,
				FirstChildID:  types.NullableContentID{ID: node.ContentDataID, Valid: true},
				NextSiblingID: parent.NextSiblingID,
				PrevSiblingID: parent.PrevSiblingID,
				RouteID:       parent.RouteID,
				DatatypeID:    parent.DatatypeID,
				AuthorID:      parent.AuthorID,
				Status:        parent.Status,
				DateCreated:   parent.DateCreated,
				DateModified:  now,
			})
			if uErr != nil {
				utility.DefaultLogger.Error("appendToSiblingChain: failed to update parent first_child_id for insert-at-start", uErr)
			}
		}

		if reqNext.Valid {
			next, gErr := d.GetContentData(reqNext.ID)
			if gErr != nil {
				utility.DefaultLogger.Error("appendToSiblingChain: failed to fetch next sibling", gErr)
			} else {
				_, uErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
					ContentDataID: next.ContentDataID,
					ParentID:      next.ParentID,
					FirstChildID:  next.FirstChildID,
					NextSiblingID: next.NextSiblingID,
					PrevSiblingID: types.NullableContentID{ID: node.ContentDataID, Valid: true},
					RouteID:       next.RouteID,
					DatatypeID:    next.DatatypeID,
					AuthorID:      next.AuthorID,
					Status:        next.Status,
					DateCreated:   next.DateCreated,
					DateModified:  now,
				})
				if uErr != nil {
					utility.DefaultLogger.Error("appendToSiblingChain: failed to update next sibling", uErr)
				}
			}
		}

		// Update the new node's sibling pointers.
		_, wErr := d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: node.ContentDataID,
			ParentID:      node.ParentID,
			FirstChildID:  node.FirstChildID,
			NextSiblingID: reqNext,
			PrevSiblingID: reqPrev,
			RouteID:       node.RouteID,
			DatatypeID:    node.DatatypeID,
			AuthorID:      node.AuthorID,
			Status:        node.Status,
			DateCreated:   node.DateCreated,
			DateModified:  now,
		})
		if wErr != nil {
			utility.DefaultLogger.Error("appendToSiblingChain: failed to set new node sibling pointers", wErr)
		}

		// Re-fetch to return updated state.
		updated, rErr := d.GetContentData(node.ContentDataID)
		if rErr != nil {
			return node, nil
		}
		return updated, nil
	}

	// No position specified — append to end of sibling chain.
	last, err := d.GetContentData(parent.FirstChildID.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch first child: %w", err)
	}
	for last.NextSiblingID.Valid {
		last, err = d.GetContentData(last.NextSiblingID.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to walk sibling chain: %w", err)
		}
	}

	// Link last sibling → new node
	_, err = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
		ContentDataID: last.ContentDataID,
		ParentID:      last.ParentID,
		FirstChildID:  last.FirstChildID,
		NextSiblingID: types.NullableContentID{ID: node.ContentDataID, Valid: true},
		PrevSiblingID: last.PrevSiblingID,
		RouteID:       last.RouteID,
		DatatypeID:    last.DatatypeID,
		AuthorID:      last.AuthorID,
		Status:        last.Status,
		DateCreated:   last.DateCreated,
		DateModified:  now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to link last sibling: %w", err)
	}

	// Set new node's prev_sibling_id to last sibling
	_, err = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
		ContentDataID: node.ContentDataID,
		ParentID:      node.ParentID,
		FirstChildID:  node.FirstChildID,
		NextSiblingID: node.NextSiblingID,
		PrevSiblingID: types.NullableContentID{ID: last.ContentDataID, Valid: true},
		RouteID:       node.RouteID,
		DatatypeID:    node.DatatypeID,
		AuthorID:      node.AuthorID,
		Status:        node.Status,
		DateCreated:   node.DateCreated,
		DateModified:  now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set new node prev_sibling_id: %w", err)
	}

	updated, err := d.GetContentData(node.ContentDataID)
	if err != nil {
		return node, nil
	}
	return updated, nil
}

// deleteContentWithSiblingRepair removes a content node and repairs the
// sibling linked list (prev/next pointers and parent's first_child_id).
func deleteContentWithSiblingRepair(
	ctx context.Context,
	ac audited.AuditContext,
	d db.DbDriver,
	cdID types.ContentID,
) error {
	node, err := d.GetContentData(cdID)
	if err != nil {
		return fmt.Errorf("failed to fetch node %s: %w", cdID, err)
	}

	now := types.TimestampNow()

	// Repair prev sibling's next pointer
	if node.PrevSiblingID.Valid {
		prev, pErr := d.GetContentData(node.PrevSiblingID.ID)
		if pErr != nil {
			return fmt.Errorf("failed to fetch prev sibling: %w", pErr)
		}
		_, pErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: prev.ContentDataID,
			ParentID:      prev.ParentID,
			FirstChildID:  prev.FirstChildID,
			NextSiblingID: node.NextSiblingID,
			PrevSiblingID: prev.PrevSiblingID,
			RouteID:       prev.RouteID,
			DatatypeID:    prev.DatatypeID,
			AuthorID:      prev.AuthorID,
			Status:        prev.Status,
			DateCreated:   prev.DateCreated,
			DateModified:  now,
		})
		if pErr != nil {
			return fmt.Errorf("failed to update prev sibling: %w", pErr)
		}
	}

	// Repair next sibling's prev pointer
	if node.NextSiblingID.Valid {
		next, nErr := d.GetContentData(node.NextSiblingID.ID)
		if nErr != nil {
			return fmt.Errorf("failed to fetch next sibling: %w", nErr)
		}
		_, nErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: next.ContentDataID,
			ParentID:      next.ParentID,
			FirstChildID:  next.FirstChildID,
			NextSiblingID: next.NextSiblingID,
			PrevSiblingID: node.PrevSiblingID,
			RouteID:       next.RouteID,
			DatatypeID:    next.DatatypeID,
			AuthorID:      next.AuthorID,
			Status:        next.Status,
			DateCreated:   next.DateCreated,
			DateModified:  now,
		})
		if nErr != nil {
			return fmt.Errorf("failed to update next sibling: %w", nErr)
		}
	}

	// If this node is the parent's first child, update parent
	if node.ParentID.Valid {
		parent, pErr := d.GetContentData(node.ParentID.ID)
		if pErr != nil {
			return fmt.Errorf("failed to fetch parent: %w", pErr)
		}
		if parent.FirstChildID.Valid && parent.FirstChildID.ID == cdID {
			_, pErr = d.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
				ContentDataID: parent.ContentDataID,
				ParentID:      parent.ParentID,
				FirstChildID:  node.NextSiblingID,
				NextSiblingID: parent.NextSiblingID,
				PrevSiblingID: parent.PrevSiblingID,
				RouteID:       parent.RouteID,
				DatatypeID:    parent.DatatypeID,
				AuthorID:      parent.AuthorID,
				Status:        parent.Status,
				DateCreated:   parent.DateCreated,
				DateModified:  now,
			})
			if pErr != nil {
				return fmt.Errorf("failed to update parent first_child_id: %w", pErr)
			}
		}
	}

	return d.DeleteContentData(ctx, ac, cdID)
}

// collectDescendants returns all descendant content IDs in delete-safe order
// (leaves first). maxDepth caps recursion (use 100 for default). maxCount
// caps total nodes (use 1000 for default). Returns an error if limits are
// exceeded.
func collectDescendants(d db.DbDriver, nodeID types.ContentID, maxDepth, maxCount int) ([]types.ContentID, error) {
	var result []types.ContentID
	err := collectDescendantsRecursive(d, nodeID, 0, maxDepth, maxCount, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func collectDescendantsRecursive(d db.DbDriver, nodeID types.ContentID, depth, maxDepth, maxCount int, result *[]types.ContentID) error {
	if depth >= maxDepth {
		return fmt.Errorf("depth limit %d exceeded", maxDepth)
	}

	node, err := d.GetContentData(nodeID)
	if err != nil {
		return fmt.Errorf("failed to fetch node %s: %w", nodeID, err)
	}

	// Walk children via first_child_id → next_sibling_id chain
	if node.FirstChildID.Valid {
		childID := node.FirstChildID.ID
		for {
			child, cErr := d.GetContentData(childID)
			if cErr != nil {
				return fmt.Errorf("failed to fetch child %s: %w", childID, cErr)
			}

			// Recurse into child's children first (depth-first, leaves first)
			if err := collectDescendantsRecursive(d, childID, depth+1, maxDepth, maxCount, result); err != nil {
				return err
			}

			if len(*result) >= maxCount {
				return fmt.Errorf("node count limit %d exceeded", maxCount)
			}

			if !child.NextSiblingID.Valid {
				break
			}
			childID = child.NextSiblingID.ID
		}
	}

	// Add this node after its children (leaves-first order), but skip the
	// root node itself — the caller handles that.
	if depth > 0 {
		*result = append(*result, nodeID)
	}

	return nil
}
