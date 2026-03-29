package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hegner123/modulacms/internal/config"
	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/audited"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/publishing"
	"github.com/hegner123/modulacms/internal/tree/ops"
	"github.com/hegner123/modulacms/internal/utility"
	"github.com/hegner123/modulacms/internal/webhooks"
)

// ContentService manages public content CRUD, content fields, tree operations
// (reorder, move, save), publishing, versioning, batch updates, and heal.
// Implements ops.Backend[types.ContentID] for sibling-pointer tree algorithms.
type ContentService struct {
	driver     db.DbDriver
	mgr        *config.Manager
	dispatcher publishing.WebhookDispatcher
}

// NewContentService creates a ContentService with the given dependencies.
func NewContentService(
	driver db.DbDriver,
	mgr *config.Manager,
	dispatcher publishing.WebhookDispatcher,
) *ContentService {
	return &ContentService{
		driver:     driver,
		mgr:        mgr,
		dispatcher: dispatcher,
	}
}

// GetNode implements ops.Backend[types.ContentID]. Fetches a content row
// and maps its sibling pointers into an ops.Node.
func (s *ContentService) GetNode(ctx context.Context, id types.ContentID) (*ops.Node[types.ContentID], error) {
	cd, err := s.driver.GetContentData(id)
	if err != nil {
		return nil, fmt.Errorf("get content node %s: %w", id, err)
	}
	return &ops.Node[types.ContentID]{
		ID:            cd.ContentDataID,
		ParentID:      contentIDToNullable(cd.ParentID),
		FirstChildID:  contentIDToNullable(cd.FirstChildID),
		NextSiblingID: contentIDToNullable(cd.NextSiblingID),
		PrevSiblingID: contentIDToNullable(cd.PrevSiblingID),
	}, nil
}

// UpdatePointers implements ops.Backend[types.ContentID]. Performs a
// read-modify-write to update only the pointer fields on a content row.
func (s *ContentService) UpdatePointers(ctx context.Context, ac audited.AuditContext, id types.ContentID, ptrs ops.Pointers[types.ContentID]) error {
	cd, err := s.driver.GetContentData(id)
	if err != nil {
		return fmt.Errorf("update pointers: get %s: %w", id, err)
	}
	_, err = s.driver.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
		ContentDataID: cd.ContentDataID,
		ParentID:      nullableToContentID(ptrs.ParentID),
		FirstChildID:  nullableToContentID(ptrs.FirstChildID),
		NextSiblingID: nullableToContentID(ptrs.NextSiblingID),
		PrevSiblingID: nullableToContentID(ptrs.PrevSiblingID),
		RouteID:       cd.RouteID,
		DatatypeID:    cd.DatatypeID,
		AuthorID:      cd.AuthorID,
		Status:        cd.Status,
		DateCreated:   cd.DateCreated,
		DateModified:  types.TimestampNow(),
	})
	if err != nil {
		return fmt.Errorf("update pointers: update %s: %w", id, err)
	}
	return nil
}

// WithTx executes fn within a database transaction. The provided Backend reads
// and writes through the transaction, so all pointer mutations are atomic —
// if any step fails, everything rolls back.
func (s *ContentService) WithTx(ctx context.Context, fn func(ops.Backend[types.ContentID]) error) error {
	conn, _, err := s.driver.GetConnection()
	if err != nil {
		return fmt.Errorf("with tx: get connection: %w", err)
	}
	return types.WithTransaction(ctx, conn, func(tx *sql.Tx) error {
		return fn(&txContentBackend{driver: s.driver, tx: tx})
	})
}

// txContentBackend implements ops.Backend[types.ContentID] with all reads and
// writes scoped to a database transaction. This ensures multi-step pointer
// operations (Unlink, InsertAt, Move, Reorder) are atomic.
type txContentBackend struct {
	driver db.DbDriver
	tx     *sql.Tx
}

// GetNode reads a content row within the transaction to see uncommitted writes.
func (tb *txContentBackend) GetNode(ctx context.Context, id types.ContentID) (*ops.Node[types.ContentID], error) {
	cd, err := db.GetContentDataInTx(tb.driver, ctx, tb.tx, id)
	if err != nil {
		return nil, fmt.Errorf("tx get content node %s: %w", id, err)
	}
	return &ops.Node[types.ContentID]{
		ID:            cd.ContentDataID,
		ParentID:      contentIDToNullable(cd.ParentID),
		FirstChildID:  contentIDToNullable(cd.FirstChildID),
		NextSiblingID: contentIDToNullable(cd.NextSiblingID),
		PrevSiblingID: contentIDToNullable(cd.PrevSiblingID),
	}, nil
}

// UpdatePointers updates pointer fields within the transaction with full audit trail.
func (tb *txContentBackend) UpdatePointers(ctx context.Context, ac audited.AuditContext, id types.ContentID, ptrs ops.Pointers[types.ContentID]) error {
	cd, err := db.GetContentDataInTx(tb.driver, ctx, tb.tx, id)
	if err != nil {
		return fmt.Errorf("tx update pointers: get %s: %w", id, err)
	}
	return db.UpdateContentDataInTx(tb.driver, ctx, tb.tx, ac, db.UpdateContentDataParams{
		ContentDataID: cd.ContentDataID,
		ParentID:      nullableToContentID(ptrs.ParentID),
		FirstChildID:  nullableToContentID(ptrs.FirstChildID),
		NextSiblingID: nullableToContentID(ptrs.NextSiblingID),
		PrevSiblingID: nullableToContentID(ptrs.PrevSiblingID),
		RouteID:       cd.RouteID,
		DatatypeID:    cd.DatatypeID,
		AuthorID:      cd.AuthorID,
		Status:        cd.Status,
		DateCreated:   cd.DateCreated,
		DateModified:  types.TimestampNow(),
	})
}

// --- Content Data CRUD ---

// Get retrieves a single content data row by ID.
func (s *ContentService) Get(ctx context.Context, id types.ContentID) (*db.ContentData, error) {
	cd, err := s.driver.GetContentData(id)
	if err != nil {
		return nil, &NotFoundError{Resource: "content_data", ID: string(id)}
	}
	return cd, nil
}

// GetFull retrieves a composed content data view with author, datatype, and fields.
func (s *ContentService) GetFull(ctx context.Context, id types.ContentID) (*db.ContentDataView, error) {
	view, err := db.AssembleContentDataView(s.driver, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "content_data", ID: string(id)}
	}
	return view, nil
}

// List returns all content data rows.
func (s *ContentService) List(ctx context.Context) (*[]db.ContentData, error) {
	return s.driver.ListContentData()
}

// ListPaginated returns a paginated list of top-level content data.
func (s *ContentService) ListPaginated(ctx context.Context, params db.PaginationParams) (*db.PaginatedResponse[db.ContentDataTopLevel], error) {
	items, err := s.driver.ListContentDataTopLevelPaginated(params)
	if err != nil {
		return nil, fmt.Errorf("list content paginated: %w", err)
	}
	total, err := s.driver.CountContentData()
	if err != nil {
		return nil, fmt.Errorf("count content data: %w", err)
	}
	return &db.PaginatedResponse[db.ContentDataTopLevel]{
		Data:   *items,
		Total:  *total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

// Create creates a new content data row. If ParentID is set, the node is
// spliced into the parent's sibling chain via AppendChild. Auto-creates
// empty content fields for every field belonging to the assigned datatype.
func (s *ContentService) Create(ctx context.Context, ac audited.AuditContext, params db.CreateContentDataParams) (*db.ContentData, error) {
	// Fill in defaults for fields that callers may omit.
	now := types.TimestampNow()
	if params.DateCreated.IsZero() {
		params.DateCreated = now
	}
	if params.DateModified.IsZero() {
		params.DateModified = now
	}
	if params.AuthorID.IsZero() && !ac.UserID.IsZero() {
		params.AuthorID = ac.UserID
	}
	if params.Status == "" {
		params.Status = types.ContentStatusDraft
	}

	// Resolve root_id before creation when the caller hasn't set it.
	// Child nodes inherit root_id from their parent.
	if !params.RootID.Valid && params.ParentID.Valid {
		parent, lookupErr := s.driver.GetContentData(params.ParentID.ID)
		if lookupErr != nil {
			utility.DefaultLogger.Error("create: failed to look up parent for root_id", lookupErr)
		} else if parent != nil {
			params.RootID = parent.RootID
		}
	}

	cd, err := s.driver.CreateContentData(ctx, ac, params)
	if err != nil {
		return nil, fmt.Errorf("create content data: %w", err)
	}

	// Root nodes (no parent) set root_id to self. The ID is not available
	// until after the INSERT, so a follow-up update is required.
	if !cd.ParentID.Valid && !cd.RootID.Valid {
		cd.RootID = types.NullableContentID{ID: cd.ContentDataID, Valid: true}
		_, rootErr := s.driver.UpdateContentData(ctx, ac, db.UpdateContentDataParams{
			ContentDataID: cd.ContentDataID,
			RootID:        cd.RootID,
			ParentID:      cd.ParentID,
			FirstChildID:  cd.FirstChildID,
			NextSiblingID: cd.NextSiblingID,
			PrevSiblingID: cd.PrevSiblingID,
			RouteID:       cd.RouteID,
			DatatypeID:    cd.DatatypeID,
			AuthorID:      cd.AuthorID,
			Status:        cd.Status,
			DateCreated:   cd.DateCreated,
			DateModified:  types.TimestampNow(),
		})
		if rootErr != nil {
			utility.DefaultLogger.Error("create: failed to set root_id on new root content", rootErr)
		}
	}

	// Splice into parent's sibling chain
	if cd.ParentID.Valid {
		if err := ops.AppendChild(ctx, ac, s, cd.ParentID.ID, cd.ContentDataID); err != nil {
			utility.DefaultLogger.Error("create: sibling chain error", err)
			// Non-fatal: node is created but not linked
		}
		// Re-fetch to get updated pointers
		cd, err = s.driver.GetContentData(cd.ContentDataID)
		if err != nil {
			return nil, fmt.Errorf("create: re-fetch after splice: %w", err)
		}
	}

	// Auto-create empty content fields for every field in the datatype
	if cd.DatatypeID.Valid {
		s.autoCreateFields(ctx, ac, cd)
	}

	return cd, nil
}

// autoCreateFields creates empty content fields for each field in the datatype.
func (s *ContentService) autoCreateFields(ctx context.Context, ac audited.AuditContext, cd *db.ContentData) {
	fieldList, err := s.driver.ListFieldsByDatatypeID(types.NullableDatatypeID{ID: cd.DatatypeID.ID, Valid: true})
	if err != nil {
		utility.DefaultLogger.Error("create: failed to list fields for auto-creation", err)
		return
	}
	if fieldList == nil {
		return
	}
	now := types.TimestampNow()
	for _, field := range *fieldList {
		_, cfErr := s.driver.CreateContentField(ctx, ac, db.CreateContentFieldParams{
			RouteID:       cd.RouteID,
			RootID:        cd.RootID,
			ContentDataID: types.NullableContentID{ID: cd.ContentDataID, Valid: true},
			FieldID:       types.NullableFieldID{ID: field.FieldID, Valid: true},
			FieldValue:    "",
			AuthorID:      cd.AuthorID,
			DateCreated:   now,
			DateModified:  now,
		})
		if cfErr != nil {
			utility.DefaultLogger.Error(fmt.Sprintf("create: failed to auto-create content field for field %s", field.FieldID), cfErr)
		}
	}
}

// Update updates a content data row. If revision > 0, uses optimistic locking.
// Returns a ConflictError on revision mismatch.
func (s *ContentService) Update(ctx context.Context, ac audited.AuditContext, params db.UpdateContentDataParams, revision int64) (*db.ContentData, error) {
	if revision > 0 {
		revParams := db.UpdateContentDataWithRevisionParams{
			RouteID:       params.RouteID,
			ParentID:      params.ParentID,
			FirstChildID:  params.FirstChildID,
			NextSiblingID: params.NextSiblingID,
			PrevSiblingID: params.PrevSiblingID,
			DatatypeID:    params.DatatypeID,
			AuthorID:      params.AuthorID,
			Status:        params.Status,
			DateCreated:   params.DateCreated,
			DateModified:  params.DateModified,
			ContentDataID: params.ContentDataID,
			Revision:      revision,
		}
		err := s.driver.UpdateContentDataWithRevision(ctx, revParams)
		if errors.Is(err, db.ErrRevisionConflict) {
			return nil, &ConflictError{
				Resource: "content_data",
				ID:       string(params.ContentDataID),
				Detail:   "revision mismatch",
			}
		}
		if err != nil {
			return nil, fmt.Errorf("update content data with revision: %w", err)
		}
	} else {
		_, err := s.driver.UpdateContentData(ctx, ac, params)
		if err != nil {
			return nil, fmt.Errorf("update content data: %w", err)
		}
	}

	updated, err := s.driver.GetContentData(params.ContentDataID)
	if err != nil {
		return nil, fmt.Errorf("update: re-fetch: %w", err)
	}

	if s.dispatcher != nil {
		s.dispatcher.Dispatch(ctx, webhooks.EventContentUpdated, map[string]any{
			"content_data_id": params.ContentDataID.String(),
			"status":          string(params.Status),
		})
	}

	return updated, nil
}

// Delete removes a content data row after unlinking it from the sibling chain.
// If recursive is true, collects and deletes all descendants first (leaves first).
func (s *ContentService) Delete(ctx context.Context, ac audited.AuditContext, id types.ContentID, recursive bool) ([]types.ContentID, error) {
	if recursive {
		return s.deleteRecursive(ctx, ac, id)
	}

	// Single delete with sibling repair via Unlink
	node, err := s.GetNode(ctx, id)
	if err != nil {
		return nil, &NotFoundError{Resource: "content_data", ID: string(id)}
	}
	if err := ops.Unlink(ctx, ac, s, node); err != nil {
		return nil, fmt.Errorf("delete: unlink: %w", err)
	}
	if err := s.driver.DeleteContentData(ctx, ac, id); err != nil {
		return nil, fmt.Errorf("delete content data: %w", err)
	}

	if s.dispatcher != nil {
		s.dispatcher.Dispatch(ctx, webhooks.EventContentDeleted, map[string]any{
			"content_data_id": id.String(),
		})
	}

	return []types.ContentID{id}, nil
}

// deleteRecursive collects all descendants in leaves-first order and deletes them.
func (s *ContentService) deleteRecursive(ctx context.Context, ac audited.AuditContext, id types.ContentID) ([]types.ContentID, error) {
	descendants, err := s.collectDescendants(id, 100, 1000)
	if err != nil {
		return nil, fmt.Errorf("recursive delete: collect: %w", err)
	}

	deleteSet := make(map[types.ContentID]struct{}, len(descendants)+1)
	deleteSet[id] = struct{}{}
	for _, did := range descendants {
		deleteSet[did] = struct{}{}
	}

	deletedIDs := make([]types.ContentID, 0, len(descendants)+1)

	// Delete descendants (leaves first). Skip sibling repair when parent is also being deleted.
	for _, descID := range descendants {
		descNode, gErr := s.driver.GetContentData(descID)
		if gErr != nil {
			return deletedIDs, fmt.Errorf("recursive delete: get %s: %w", descID, gErr)
		}
		_, parentInSet := deleteSet[descNode.ParentID.ID]
		if descNode.ParentID.Valid && parentInSet {
			if err := s.driver.DeleteContentData(ctx, ac, descID); err != nil {
				return deletedIDs, fmt.Errorf("recursive delete: delete %s: %w", descID, err)
			}
		} else {
			node, nErr := s.GetNode(ctx, descID)
			if nErr != nil {
				return deletedIDs, fmt.Errorf("recursive delete: get node %s: %w", descID, nErr)
			}
			if err := ops.Unlink(ctx, ac, s, node); err != nil {
				return deletedIDs, fmt.Errorf("recursive delete: unlink %s: %w", descID, err)
			}
			if err := s.driver.DeleteContentData(ctx, ac, descID); err != nil {
				return deletedIDs, fmt.Errorf("recursive delete: delete %s: %w", descID, err)
			}
		}
		deletedIDs = append(deletedIDs, descID)
	}

	// Delete the root node with sibling repair
	rootNode, err := s.GetNode(ctx, id)
	if err != nil {
		return deletedIDs, fmt.Errorf("recursive delete: get root: %w", err)
	}
	if err := ops.Unlink(ctx, ac, s, rootNode); err != nil {
		return deletedIDs, fmt.Errorf("recursive delete: unlink root: %w", err)
	}
	if err := s.driver.DeleteContentData(ctx, ac, id); err != nil {
		return deletedIDs, fmt.Errorf("recursive delete: delete root: %w", err)
	}
	deletedIDs = append(deletedIDs, id)

	return deletedIDs, nil
}

// collectDescendants returns all descendant content IDs in leaves-first order.
func (s *ContentService) collectDescendants(nodeID types.ContentID, maxDepth, maxCount int) ([]types.ContentID, error) {
	var result []types.ContentID
	if err := s.collectDescendantsRecursive(nodeID, 0, maxDepth, maxCount, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ContentService) collectDescendantsRecursive(nodeID types.ContentID, depth, maxDepth, maxCount int, result *[]types.ContentID) error {
	if depth >= maxDepth {
		return fmt.Errorf("depth limit %d exceeded", maxDepth)
	}
	node, err := s.driver.GetContentData(nodeID)
	if err != nil {
		return fmt.Errorf("failed to fetch node %s: %w", nodeID, err)
	}
	if node.FirstChildID.Valid {
		childID := node.FirstChildID.ID
		for {
			child, cErr := s.driver.GetContentData(childID)
			if cErr != nil {
				return fmt.Errorf("failed to fetch child %s: %w", childID, cErr)
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

// Move moves a content node to a new parent at the given position.
// All pointer mutations execute within a single transaction — if any step
// fails (cycle detection, unlink, insert, or validation), everything rolls back.
func (s *ContentService) Move(ctx context.Context, ac audited.AuditContext, params ops.MoveParams[types.ContentID]) (*ops.MoveResult[types.ContentID], error) {
	var result *ops.MoveResult[types.ContentID]
	err := s.WithTx(ctx, func(tb ops.Backend[types.ContentID]) error {
		var moveErr error
		result, moveErr = ops.Move(ctx, ac, tb, params)
		return moveErr
	})
	if err != nil {
		return nil, fmt.Errorf("move content: %w", err)
	}
	return result, nil
}

// Reorder atomically reorders sibling content nodes under a parent.
// All pointer mutations execute within a single transaction.
func (s *ContentService) Reorder(ctx context.Context, ac audited.AuditContext, parentID ops.NullableID[types.ContentID], orderedIDs []types.ContentID) (*ops.ReorderResult[types.ContentID], error) {
	var result *ops.ReorderResult[types.ContentID]
	err := s.WithTx(ctx, func(tb ops.Backend[types.ContentID]) error {
		var reorderErr error
		result, reorderErr = ops.Reorder(ctx, ac, tb, parentID, orderedIDs)
		return reorderErr
	})
	if err != nil {
		return nil, fmt.Errorf("reorder content: %w", err)
	}
	return result, nil
}

// contentIDToNullable converts types.NullableContentID to ops.NullableID.
func contentIDToNullable(n types.NullableContentID) ops.NullableID[types.ContentID] {
	if !n.Valid {
		return ops.EmptyID[types.ContentID]()
	}
	return ops.NullID(n.ID)
}

// nullableToContentID converts ops.NullableID to types.NullableContentID.
func nullableToContentID(n ops.NullableID[types.ContentID]) types.NullableContentID {
	if !n.Valid {
		return types.NullableContentID{}
	}
	return types.NullableContentID{ID: n.Value, Valid: true}
}

// ListByRoute returns content tree nodes for a specific route.
func (s *ContentService) ListByRoute(ctx context.Context, routeID types.RouteID) ([]db.RouteContentNodeView, error) {
	nullRouteID := types.NullableRouteID{ID: routeID, Valid: true}
	rows, err := s.driver.GetContentTreeByRoute(nullRouteID)
	if err != nil {
		return nil, fmt.Errorf("get content tree by route: %w", err)
	}
	views := make([]db.RouteContentNodeView, 0, len(*rows))
	for _, row := range *rows {
		views = append(views, db.RouteContentNodeView{
			ContentDataID: row.ContentDataID,
			ParentID:      row.ParentID,
			DatatypeLabel: row.DatatypeLabel,
			DatatypeType:  row.DatatypeType,
			Status:        row.Status,
			DateCreated:   row.DateCreated,
			DateModified:  row.DateModified,
		})
	}
	return views, nil
}

// GetTree returns the content tree for a route.
func (s *ContentService) GetTree(ctx context.Context, routeID types.NullableRouteID) (*[]db.GetContentTreeByRouteRow, error) {
	tree, err := s.driver.GetContentTreeByRoute(routeID)
	if err != nil {
		return nil, fmt.Errorf("get content tree: %w", err)
	}
	return tree, nil
}
