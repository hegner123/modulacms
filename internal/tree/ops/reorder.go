package ops

import (
	"context"
	"fmt"

	"github.com/hegner123/modulacms/internal/db/audited"
)

// Reorder rewrites sibling pointers for a set of nodes under a common parent.
// orderedIDs specifies the desired order from first to last. All nodes must
// share the same parentID. If parentID is valid, the parent's FirstChildID
// is updated to orderedIDs[0].
//
// Safety: when parentID is valid, the function walks the existing sibling
// chain and verifies that every current child is present in orderedIDs.
// Missing children cause an error rather than silent orphaning.
//
// Returns the number of nodes updated.
func Reorder[ID ~string](ctx context.Context, ac audited.AuditContext, b Backend[ID], parentID NullableID[ID], orderedIDs []ID) (int, error) {
	if len(orderedIDs) == 0 {
		return 0, fmt.Errorf("reorder: ordered_ids must not be empty")
	}

	// Build lookup set for the provided IDs.
	provided := make(map[string]struct{}, len(orderedIDs))
	for _, id := range orderedIDs {
		provided[string(id)] = struct{}{}
	}

	// Verify all nodes exist and belong to the stated parent
	nodes := make([]*Node[ID], 0, len(orderedIDs))
	for _, id := range orderedIDs {
		node, err := b.GetNode(ctx, id)
		if err != nil {
			return 0, fmt.Errorf("reorder: get node %s: %w", string(id), err)
		}
		if node.ParentID != parentID {
			return 0, fmt.Errorf("reorder: node %s does not belong to parent", string(id))
		}
		nodes = append(nodes, node)
	}

	// Walk existing sibling chain to detect missing children.
	// This prevents silent orphaning when the caller omits siblings.
	if parentID.Valid {
		parent, err := b.GetNode(ctx, parentID.Value)
		if err != nil {
			return 0, fmt.Errorf("reorder: get parent: %w", err)
		}
		if parent.FirstChildID.Valid {
			var missing []string
			visited := make(map[string]bool)
			cur := parent.FirstChildID.Value
			for {
				key := string(cur)
				if visited[key] {
					break // cycle guard
				}
				visited[key] = true
				if _, ok := provided[key]; !ok {
					missing = append(missing, key)
				}
				sibling, err := b.GetNode(ctx, cur)
				if err != nil {
					break // node deleted or inaccessible, skip
				}
				if !sibling.NextSiblingID.Valid {
					break
				}
				cur = sibling.NextSiblingID.Value
			}
			if len(missing) > 0 {
				return 0, fmt.Errorf("reorder: ordered_ids is missing existing children: %v — include all siblings to prevent orphans", missing)
			}
		}

		// Update parent's first_child_id
		err = b.UpdatePointers(ctx, ac, parent.ID, Pointers[ID]{
			ParentID:      parent.ParentID,
			FirstChildID:  NullID(orderedIDs[0]),
			NextSiblingID: parent.NextSiblingID,
			PrevSiblingID: parent.PrevSiblingID,
		})
		if err != nil {
			return 0, fmt.Errorf("reorder: update parent first_child: %w", err)
		}
	}

	// Rewrite sibling pointers for each node
	lastIdx := len(orderedIDs) - 1
	for i, node := range nodes {
		var prevSibling NullableID[ID]
		var nextSibling NullableID[ID]

		if i > 0 {
			prevSibling = NullID(orderedIDs[i-1])
		}
		if i < lastIdx {
			nextSibling = NullID(orderedIDs[i+1])
		}

		err := b.UpdatePointers(ctx, ac, node.ID, Pointers[ID]{
			ParentID:      node.ParentID,
			FirstChildID:  node.FirstChildID,
			NextSiblingID: nextSibling,
			PrevSiblingID: prevSibling,
		})
		if err != nil {
			return i, fmt.Errorf("reorder: update node %s: %w", string(node.ID), err)
		}
	}

	return len(orderedIDs), nil
}
