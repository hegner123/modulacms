package ops

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
)

// Save performs a multi-phase bulk tree operation:
//  1. Create new nodes (via CreateFn), collecting temp-to-real ID mappings
//  2. Remap temp IDs in the Order slice
//  3. Run non-pointer updates (via UpdateFn)
//  4. Delete nodes
//  5. Reorder remaining siblings using the remapped Order
//
// Errors from individual creates/updates/deletes are collected; Save does
// not abort on the first failure. The returned SaveResult reports counts
// and the ID map.
func Save[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], params SaveParams[ID]) (*SaveResult[ID], error) {
	result := &SaveResult[ID]{
		IDMap: make(map[ID]ID, len(params.Creates)),
	}

	// Phase 1: Create nodes
	var createErrors []error
	for _, cr := range params.Creates {
		realID, err := cr.CreateFn(ctx, ac)
		if err != nil {
			createErrors = append(createErrors, fmt.Errorf("save: create %s: %w", string(cr.TempID), err))
			continue
		}
		result.IDMap[cr.TempID] = realID
		result.Created++
	}

	// Phase 2: Remap temp IDs in Order slice
	remappedOrder := make([]ID, 0, len(params.Order))
	for _, id := range params.Order {
		if realID, ok := result.IDMap[id]; ok {
			remappedOrder = append(remappedOrder, realID)
		} else {
			remappedOrder = append(remappedOrder, id)
		}
	}

	// Phase 3: Run non-pointer updates
	var updateErrors []error
	for _, upd := range params.Updates {
		if err := upd.UpdateFn(ctx, ac); err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("save: update %s: %w", string(upd.ID), err))
			continue
		}
		result.Updated++
	}

	// Phase 4: Delete nodes
	// Deletes use Unlink to repair the sibling chain before removal.
	// The actual DB delete is the caller's responsibility via DeleteFn or
	// a post-Save cleanup — Save only unlinks.
	var deleteErrors []error
	for _, id := range params.Deletes {
		node, err := b.GetNode(ctx, id)
		if err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("save: get delete target %s: %w", string(id), err))
			continue
		}
		if err := Unlink(ctx, ac, b, node); err != nil {
			deleteErrors = append(deleteErrors, fmt.Errorf("save: unlink delete target %s: %w", string(id), err))
			continue
		}
		result.Deleted++
	}

	// Remove deleted IDs from remapped order
	if len(params.Deletes) > 0 {
		deleteSet := make(map[ID]struct{}, len(params.Deletes))
		for _, id := range params.Deletes {
			deleteSet[id] = struct{}{}
		}
		filtered := make([]ID, 0, len(remappedOrder))
		for _, id := range remappedOrder {
			if _, deleted := deleteSet[id]; !deleted {
				filtered = append(filtered, id)
			}
		}
		remappedOrder = filtered
	}

	// Phase 5: Reorder siblings if an order was specified
	if len(remappedOrder) > 0 {
		_, err := Reorder(ctx, ac, b, params.ParentID, remappedOrder)
		if err != nil {
			return result, fmt.Errorf("save: reorder: %w", err)
		}
	}

	// Collect errors
	allErrors := append(append(createErrors, updateErrors...), deleteErrors...)
	if len(allErrors) > 0 {
		return result, fmt.Errorf("save completed with %d errors; first: %w", len(allErrors), allErrors[0])
	}

	return result, nil
}
