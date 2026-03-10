package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/tree/ops"
	"github.com/hegner123/modulacms/internal/utility"
)

// AdminContentService manages public content CRUD, content fields, tree operations
// (reorder, move, save), publishing, versioning, batch updates, and heal.
// Implements ops.Backend[types.AdminContentID] for sibling-pointer tree algorithms.
type AdminContentService struct {
	driver     db.DbDriver
	mgr        *config.Manager
	dispatcher publishing.WebhookDispatcher
}

// NewAdminContentService creates a AdminContentService with the given dependencies.
func NewAdminContentService(
	driver db.DbDriver,
	mgr *config.Manager,
	dispatcher publishing.WebhookDispatcher,
) *AdminContentService {
	return &AdminContentService{
		driver:     driver,
		mgr:        mgr,
		dispatcher: dispatcher,
	}
}

// GetNode implements ops.Backend[types.AdminContentID]. Fetches an admin content
// row and maps its sibling pointers into an ops.Node.
func (s *AdminContentService) GetNode(ctx context.Context, id types.AdminContentID) (*ops.Node[types.AdminContentID], error) {
	cd, err := s.driver.GetAdminContentData(id)
	if err != nil {
		return nil, fmt.Errorf("get admin content node %s: %w", id, err)
	}
	return &ops.Node[types.AdminContentID]{
		ID:            cd.AdminContentDataID,
		ParentID:      adminContentIDToNullable(cd.ParentID),
		FirstChildID:  adminContentIDToNullable(cd.FirstChildID),
		NextSiblingID: adminContentIDToNullable(cd.NextSiblingID),
		PrevSiblingID: adminContentIDToNullable(cd.PrevSiblingID),
	}, nil
}

// UpdatePointers implements ops.Backend[types.AdminContentID]. Performs a
// read-modify-write to update only the pointer fields on an admin content row.
func (s *AdminContentService) UpdatePointers(ctx context.Context, ac audited.AuditContext, id types.AdminContentID, ptrs ops.Pointers[types.AdminContentID]) error {
	cd, err := s.driver.GetAdminContentData(id)
	if err != nil {
		return fmt.Errorf("update admin pointers: get %s: %w", id, err)
	}
	_, err = s.driver.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
		AdminContentDataID: cd.AdminContentDataID,
		ParentID:           nullableToAdminContentID(ptrs.ParentID),
		FirstChildID:       nullableToAdminContentID(ptrs.FirstChildID),
		NextSiblingID:      nullableToAdminContentID(ptrs.NextSiblingID),
		PrevSiblingID:      nullableToAdminContentID(ptrs.PrevSiblingID),
		AdminRouteID:       cd.AdminRouteID,
		AdminDatatypeID:    cd.AdminDatatypeID,
		AuthorID:           cd.AuthorID,
		Status:             cd.Status,
		DateCreated:        cd.DateCreated,
		DateModified:       types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("update admin pointers: update %s: %w", id, err)
	}
	return nil
}

// WithTx is a no-op passthrough until real transactions are added to DbDriver.
func (s *AdminContentService) WithTx(ctx context.Context, fn func(ops.Backend[types.AdminContentID]) error) error {
	return fn(s)
}

// --- Admin Content Data CRUD ---

// Get retrieves a single admin content data row by ID.
func (s *AdminContentService) Get(ctx context.Context, id types.AdminContentID) (*db.AdminContentData, error) {
	cd, err := s.driver.GetAdminContentData(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "admin_content_data", ID: string(id)}
	}
	return cd, nil
}

// GetFull retrieves a composed admin content data view with author, datatype, and fields.
func (s *AdminContentService) GetFull(ctx context.Context, id types.AdminContentID) (*db.AdminContentDataView, error) {
	view, err := db.AssembleAdminContentDataView(s.driver, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "admin_content_data", ID: string(id)}
	}
	return view, nil
}

// List returns all admin content data rows.
func (s *AdminContentService) List(ctx context.Context) (*[]db.AdminContentData, error) {
	return s.driver.ListAdminContentData()
}

// ListPaginated returns a paginated list of top-level admin content data.
func (s *AdminContentService) ListPaginated(ctx context.Context, params db.PaginationParams) (*db.PaginatedResponse[db.AdminContentDataTopLevel], error) {
	items, err := s.driver.ListAdminContentDataTopLevelPaginated(params)
	if err != nil {
		return nil, fmt.Errorf("list admin content paginated: %w", err)
	}
	total, err := s.driver.CountAdminContentDataTopLevel()
	if err != nil {
		return nil, fmt.Errorf("count admin content data: %w", err)
	}
	return &db.PaginatedResponse[db.AdminContentDataTopLevel]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// Create creates a new admin content data row. If ParentID is set, the node is
// spliced into the parent's sibling chain via AppendChild. Auto-creates empty
// admin content fields for every field belonging to the assigned admin datatype.
func (s *AdminContentService) Create(ctx context.Context, ac audited.AuditContext, params db.CreateAdminContentDataParams) (*db.AdminContentData, error) {
	// Resolve root_id before creation when the caller hasn't set it.
	// Child nodes inherit root_id from their parent.
	if !params.RootID.Valid && params.ParentID.Valid {
		parent, lookupErr := s.driver.GetAdminContentData(params.ParentID.ID)
		if lookupErr != nil {
			utility.DefaultLogger.Error("admin create: failed to look up parent for root_id", lookupErr)
		} else if parent != nil {
			params.RootID = parent.RootID
		}
	}

	cd, err := s.driver.CreateAdminContentData(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create admin content data: %w", err)
	}

	// Root nodes (no parent) set root_id to self. The ID is not available
	// until after the INSERT, so a follow-up update is required.
	if !cd.ParentID.Valid && !cd.RootID.Valid {
		cd.RootID = types.NullableAdminContentID{ID: cd.AdminContentDataID, Valid: true}
		_, rootErr := s.driver.UpdateAdminContentData(ctx, ac, db.UpdateAdminContentDataParams{
			AdminContentDataID: cd.AdminContentDataID,
			RootID:             cd.RootID,
			ParentID:           cd.ParentID,
			FirstChildID:       cd.FirstChildID,
			NextSiblingID:      cd.NextSiblingID,
			PrevSiblingID:      cd.PrevSiblingID,
			AdminRouteID:       cd.AdminRouteID,
			AdminDatatypeID:    cd.AdminDatatypeID,
			AuthorID:           cd.AuthorID,
			Status:             cd.Status,
			DateCreated:        cd.DateCreated,
			DateModified:       types.TimestampNow(),
		})
		if rootErr != nil {
			utility.DefaultLogger.Error("admin create: failed to set root_id on new root content", rootErr)
		}
	}

	// Splice into parent's sibling chain
	if cd.ParentID.Valid {
		if err := ops.AppendChild(ctx, ac, s, cd.ParentID.ID, cd.AdminContentDataID); err != nil {
			utility.DefaultLogger.Error("admin create: sibling chain error", err)
			// Non-fatal: node is created but not linked
		}
		// Re-fetch to get updated pointers
		cd, err = s.driver.GetAdminContentData(cd.AdminContentDataID)
		if err != nil {
			return nil, fmt.Errorf("admin create: re-fetch after splice: %w", err)
		}
	}

	// Auto-create empty admin content fields for every field in the admin datatype
	if cd.AdminDatatypeID.Valid {
		s.autoCreateAdminFields(ctx, ac, cd)
	}

	return cd, nil
}

// autoCreateAdminFields creates empty admin content fields for each field in the admin datatype.
func (s *AdminContentService) autoCreateAdminFields(ctx context.Context, ac audited.AuditContext, cd *db.AdminContentData) {
	fieldList, err := s.driver.ListAdminFieldsByDatatypeID(types.NullableAdminDatatypeID{ID: cd.AdminDatatypeID.ID, Valid: true})
	if err != nil {
		utility.DefaultLogger.Error("admin create: failed to list admin fields for auto-creation", err)
		return
	}
	if fieldList == nil {
		return
	}
	now := types.TimestampNow()
	for _, field := range *fieldList {
		_, cfErr := s.driver.CreateAdminContentField(ctx, ac, db.CreateAdminContentFieldParams{
			AdminRouteID:       cd.AdminRouteID,
			RootID:             cd.RootID,
			AdminContentDataID: types.NullableAdminContentID{ID: cd.AdminContentDataID, Valid: true},
			AdminFieldID:       types.NullableAdminFieldID{ID: field.AdminFieldID, Valid: true},
			AdminFieldValue:    "",
			AuthorID:           cd.AuthorID,
			DateCreated:        now,
			DateModified:       now,
		})
		if cfErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("admin create: failed to auto-create admin content field for field %s", field.AdminFieldID), cfErr)
		}
	}
}

// Update updates an admin content data row. If revision > 0, uses optimistic locking.
// Returns a ConflictError on revision mismatch.
func (s *AdminContentService) Update(ctx context.Context, ac audited.AuditContext, params db.UpdateAdminContentDataParams, revision int64) (*db.AdminContentData, error) {
	if revision > 0 {
		revParams := db.UpdateAdminContentDataWithRevisionParams{
			AdminRouteID:       params.AdminRouteID,
			ParentID:           params.ParentID,
			FirstChildID:       params.FirstChildID,
			NextSiblingID:      params.NextSiblingID,
			PrevSiblingID:      params.PrevSiblingID,
			AdminDatatypeID:    params.AdminDatatypeID,
			AuthorID:           params.AuthorID,
			Status:             params.Status,
			DateCreated:        params.DateCreated,
			DateModified:       params.DateModified,
			AdminContentDataID: params.AdminContentDataID,
			Revision:           revision,
		}
		err := s.driver.UpdateAdminContentDataWithRevision(ctx, revParams)
		if errors.Is(err, db.ErrRevisionConflict) {
			return nil, &ConflictError{
				Resource: "admin_content_data",
				ID:       string(params.AdminContentDataID),
				Detail:   "revision mismatch",
			}
		}
		if err != nil {
			return nil, fmt.Errorf("update admin content data with revision: %w", err)
		}
	} else {
		_, err := s.driver.UpdateAdminContentData(ctx, ac, params)
		if err != nil {
			return nil, fmt.Errorf("update admin content data: %w", err)
		}
	}

	updated, err := s.driver.GetAdminContentData(params.AdminContentDataID)
	if err != nil {
		return nil, fmt.Errorf("admin update: re-fetch: %w", err)
	}
	return updated, nil
}

// Delete removes an admin content data row after unlinking it from the sibling chain.
// If recursive is true, collects and deletes all descendants first (leaves first).
func (s *AdminContentService) Delete(ctx context.Context, ac audited.AuditContext, id types.AdminContentID, recursive bool) ([]types.AdminContentID, error) {
	if recursive {
		return s.deleteRecursive(ctx, ac, id)
	}

	// Single delete with sibling repair via Unlink
	node, err := s.GetNode(ctx, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "admin_content_data", ID: string(id)}
	}
	if err := ops.Unlink(ctx, ac, s, node); err != nil {
		return nil, fmt.Errorf("admin delete: unlink: %w", err)
	}
	if err := s.driver.DeleteAdminContentData(ctx, ac, id); err != nil {
		return nil, fmt.Errorf("delete admin content data: %w", err)
	}
	return []types.AdminContentID{id}, nil
}

// deleteRecursive collects all descendants in leaves-first order and deletes them.
func (s *AdminContentService) deleteRecursive(ctx context.Context, ac audited.AuditContext, id types.AdminContentID) ([]types.AdminContentID, error) {
	descendants, err := s.collectDescendants(id, 100, 1000)
	if err != nil {
		return nil, fmt.Errorf("admin recursive delete: collect: %w", err)
	}

	deleteSet := make(map[types.AdminContentID]struct{}, len(descendants)+1)
	deleteSet[id] = struct{}{}
	for _, did := range descendants {
		deleteSet[did] = struct{}{}
	}

	deletedIDs := make([]types.AdminContentID, 0, len(descendants)+1)

	// Delete descendants (leaves first). Skip sibling repair when parent is also being deleted.
	for _, descID := range descendants {
		descNode, gErr := s.driver.GetAdminContentData(descID)
		if gErr != nil {
			return deletedIDs, fmt.Errorf("admin recursive delete: get %s: %w", descID, gErr)
		}
		_, parentInSet := deleteSet[descNode.ParentID.ID]
		if descNode.ParentID.Valid && parentInSet {
			if err := s.driver.DeleteAdminContentData(ctx, ac, descID); err != nil {
				return deletedIDs, fmt.Errorf("admin recursive delete: delete %s: %w", descID, err)
			}
		} else {
			node, nErr := s.GetNode(ctx, descID)
			if nErr != nil {
				return deletedIDs, fmt.Errorf("admin recursive delete: get node %s: %w", descID, nErr)
			}
			if err := ops.Unlink(ctx, ac, s, node); err != nil {
				return deletedIDs, fmt.Errorf("admin recursive delete: unlink %s: %w", descID, err)
			}
			if err := s.driver.DeleteAdminContentData(ctx, ac, descID); err != nil {
				return deletedIDs, fmt.Errorf("admin recursive delete: delete %s: %w", descID, err)
			}
		}
		deletedIDs = append(deletedIDs, descID)
	}

	// Delete the root node with sibling repair
	rootNode, err := s.GetNode(ctx, id)
	if err != nil {
		return deletedIDs, fmt.Errorf("admin recursive delete: get root: %w", err)
	}
	if err := ops.Unlink(ctx, ac, s, rootNode); err != nil {
		return deletedIDs, fmt.Errorf("admin recursive delete: unlink root: %w", err)
	}
	if err := s.driver.DeleteAdminContentData(ctx, ac, id); err != nil {
		return deletedIDs, fmt.Errorf("admin recursive delete: delete root: %w", err)
	}
	deletedIDs = append(deletedIDs, id)

	return deletedIDs, nil
}

// collectDescendants returns all descendant admin content IDs in leaves-first order.
func (s *AdminContentService) collectDescendants(nodeID types.AdminContentID, maxDepth, maxCount int) ([]types.AdminContentID, error) {
	var result []types.AdminContentID
	if err := s.collectDescendantsRecursive(nodeID, 0, maxDepth, maxCount, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *AdminContentService) collectDescendantsRecursive(nodeID types.AdminContentID, depth, maxDepth, maxCount int, result *[]types.AdminContentID) error {
	if depth >= maxDepth {
		return fmt.Errorf("depth limit %d exceeded", maxDepth)
	}
	node, err := s.driver.GetAdminContentData(nodeID)
	if err != nil {
		return fmt.Errorf("failed to fetch admin node %s: %w", nodeID, err)
	}
	if node.FirstChildID.Valid {
		childID := node.FirstChildID.ID
		for {
			child, cErr := s.driver.GetAdminContentData(childID)
			if cErr != nil {
				return fmt.Errorf("failed to fetch admin child %s: %w", childID, cErr)
			}
			if err := s.collectDescendantsRecursive(childID, depth+1, maxDepth, maxCount, result); err != nil {
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
	if depth > 0 {
		*result = append(*result, nodeID)
	}
	return nil
}

// --- Tree Operations ---

// Move moves an admin content node to a new parent at the given position.
// Delegates to ops.Move which handles cycle detection, unlinking, and reinsertion.
func (s *AdminContentService) Move(ctx context.Context, ac audited.AuditContext, params ops.MoveParams[types.AdminContentID]) (*ops.MoveResult[types.AdminContentID], error) {
	result, err := ops.Move(ctx, ac, s, params)
	if err != nil {
		return nil, fmt.Errorf("admin move content: %w", err)
	}
	return result, nil
}

// Reorder atomically reorders sibling admin content nodes under a parent.
// Delegates to ops.Reorder which rewrites all sibling pointers and parent first_child.
func (s *AdminContentService) Reorder(ctx context.Context, ac audited.AuditContext, parentID ops.NullableID[types.AdminContentID], orderedIDs []types.AdminContentID) (int, error) {
	updated, err := ops.Reorder(ctx, ac, s, parentID, orderedIDs)
	if err != nil {
		return 0, fmt.Errorf("admin reorder content: %w", err)
	}
	return updated, nil
}

// adminContentIDToNullable converts types.NullableAdminContentID to ops.NullableID.
func adminContentIDToNullable(n types.NullableAdminContentID) ops.NullableID[types.AdminContentID] {
	if !n.Valid {
		return ops.EmptyID[types.AdminContentID]()
	}
	return ops.NullID(n.ID)
}

// nullableToAdminContentID converts ops.NullableID to types.NullableAdminContentID.
func nullableToAdminContentID(n ops.NullableID[types.AdminContentID]) types.NullableAdminContentID {
	if !n.Valid {
		return types.NullableAdminContentID{}
	}
	return types.NullableAdminContentID{ID: n.Value, Valid: true}
}
