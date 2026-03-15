package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hegner123/modulacms/internal/db/types"
)

var mediaFolderNameMu sync.Mutex

// GetMediaFolderBreadcrumb walks up the parent chain from the given folder to root.
// Returns folders in order from root to the given folder. Max 10 iterations.
func (d Database) GetMediaFolderBreadcrumb(folderID types.MediaFolderID) ([]MediaFolder, error) {
	var chain []MediaFolder
	currentID := folderID

	for i := range 10 {
		_ = i
		folder, err := d.GetMediaFolder(currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get media folder %s in breadcrumb: %w", currentID, err)
		}
		chain = append(chain, *folder)

		if !folder.ParentID.Valid {
			break
		}
		currentID = types.MediaFolderID(folder.ParentID.ID)

		if len(chain) > 10 {
			return nil, fmt.Errorf("media folder breadcrumb exceeded max depth of 10 (circular reference?)")
		}
	}

	// Reverse: chain is leaf-to-root, we want root-to-leaf
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

// ValidateMediaFolderName validates folder name rules and uniqueness within parent.
// Must be called before create/update. Protected by mutex to prevent race conditions.
func (d Database) ValidateMediaFolderName(name string, parentID types.NullableMediaFolderID) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("media folder name cannot be empty")
	}
	if len(trimmed) > 255 {
		return fmt.Errorf("media folder name cannot exceed 255 characters")
	}
	if strings.ContainsAny(trimmed, "/\\\x00") {
		return fmt.Errorf("media folder name cannot contain '/', '\\', or null bytes")
	}
	if trimmed == "." || trimmed == ".." {
		return fmt.Errorf("media folder name cannot be %q", trimmed)
	}

	mediaFolderNameMu.Lock()
	defer mediaFolderNameMu.Unlock()

	if parentID.Valid {
		_, err := d.GetMediaFolderByNameAndParent(types.MediaFolderID(parentID.ID), trimmed)
		if err == nil {
			return fmt.Errorf("a folder named %q already exists in this parent", trimmed)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check folder name uniqueness: %w", err)
		}
	} else {
		_, err := d.GetMediaFolderByNameAtRoot(trimmed)
		if err == nil {
			return fmt.Errorf("a folder named %q already exists at root", trimmed)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check folder name uniqueness: %w", err)
		}
	}

	return nil
}

// ValidateMediaFolderMove checks that moving a folder to a new parent won't create
// a cycle or exceed max depth (10 levels).
func (d Database) ValidateMediaFolderMove(folderID types.MediaFolderID, newParentID types.NullableMediaFolderID) error {
	// Moving to root: just check subtree depth
	if !newParentID.Valid {
		depth, err := d.maxSubtreeDepth(folderID)
		if err != nil {
			return fmt.Errorf("failed to calculate subtree depth: %w", err)
		}
		// At root the folder itself is at depth 1, plus its subtree
		if 1+depth > 10 {
			return fmt.Errorf("move would exceed maximum folder depth of 10")
		}
		return nil
	}

	// Walk from newParentID up to root -- if folderID is encountered, it's a cycle
	currentID := types.MediaFolderID(newParentID.ID)
	ancestorDepth := 0
	for {
		ancestorDepth++
		if currentID == folderID {
			return fmt.Errorf("move would create a circular reference")
		}
		if ancestorDepth > 10 {
			return fmt.Errorf("move would exceed maximum folder depth of 10")
		}

		folder, err := d.GetMediaFolder(currentID)
		if err != nil {
			return fmt.Errorf("failed to get ancestor folder %s: %w", currentID, err)
		}
		if !folder.ParentID.Valid {
			break
		}
		currentID = types.MediaFolderID(folder.ParentID.ID)
	}

	// Calculate total depth: ancestors + 1 (for the folder itself) + subtree
	subtreeDepth, err := d.maxSubtreeDepth(folderID)
	if err != nil {
		return fmt.Errorf("failed to calculate subtree depth: %w", err)
	}
	totalDepth := ancestorDepth + 1 + subtreeDepth
	if totalDepth > 10 {
		return fmt.Errorf("move would exceed maximum folder depth of 10")
	}

	return nil
}

// maxSubtreeDepth recursively finds the deepest leaf under folderID.
// Returns 0 if folder has no children.
func (d Database) maxSubtreeDepth(folderID types.MediaFolderID) (int, error) {
	children, err := d.ListMediaFoldersByParent(folderID)
	if err != nil {
		return 0, fmt.Errorf("failed to list children of folder %s: %w", folderID, err)
	}
	if children == nil || len(*children) == 0 {
		return 0, nil
	}

	maxDepth := 0
	for _, child := range *children {
		childDepth, err := d.maxSubtreeDepth(child.FolderID)
		if err != nil {
			return 0, err
		}
		depth := 1 + childDepth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth, nil
}

// ===== MySQL =====

// GetMediaFolderBreadcrumb walks up the parent chain from the given folder to root (MySQL).
// Returns folders in order from root to the given folder. Max 10 iterations.
func (d MysqlDatabase) GetMediaFolderBreadcrumb(folderID types.MediaFolderID) ([]MediaFolder, error) {
	var chain []MediaFolder
	currentID := folderID

	for i := range 10 {
		_ = i
		folder, err := d.GetMediaFolder(currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get media folder %s in breadcrumb: %w", currentID, err)
		}
		chain = append(chain, *folder)

		if !folder.ParentID.Valid {
			break
		}
		currentID = types.MediaFolderID(folder.ParentID.ID)

		if len(chain) > 10 {
			return nil, fmt.Errorf("media folder breadcrumb exceeded max depth of 10 (circular reference?)")
		}
	}

	// Reverse: chain is leaf-to-root, we want root-to-leaf
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

// ValidateMediaFolderName validates folder name rules and uniqueness within parent (MySQL).
// Must be called before create/update. Protected by mutex to prevent race conditions.
func (d MysqlDatabase) ValidateMediaFolderName(name string, parentID types.NullableMediaFolderID) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("media folder name cannot be empty")
	}
	if len(trimmed) > 255 {
		return fmt.Errorf("media folder name cannot exceed 255 characters")
	}
	if strings.ContainsAny(trimmed, "/\\\x00") {
		return fmt.Errorf("media folder name cannot contain '/', '\\', or null bytes")
	}
	if trimmed == "." || trimmed == ".." {
		return fmt.Errorf("media folder name cannot be %q", trimmed)
	}

	mediaFolderNameMu.Lock()
	defer mediaFolderNameMu.Unlock()

	if parentID.Valid {
		_, err := d.GetMediaFolderByNameAndParent(types.MediaFolderID(parentID.ID), trimmed)
		if err == nil {
			return fmt.Errorf("a folder named %q already exists in this parent", trimmed)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check folder name uniqueness: %w", err)
		}
	} else {
		_, err := d.GetMediaFolderByNameAtRoot(trimmed)
		if err == nil {
			return fmt.Errorf("a folder named %q already exists at root", trimmed)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check folder name uniqueness: %w", err)
		}
	}

	return nil
}

// ValidateMediaFolderMove checks that moving a folder to a new parent won't create
// a cycle or exceed max depth (10 levels) (MySQL).
func (d MysqlDatabase) ValidateMediaFolderMove(folderID types.MediaFolderID, newParentID types.NullableMediaFolderID) error {
	// Moving to root: just check subtree depth
	if !newParentID.Valid {
		depth, err := d.maxSubtreeDepth(folderID)
		if err != nil {
			return fmt.Errorf("failed to calculate subtree depth: %w", err)
		}
		// At root the folder itself is at depth 1, plus its subtree
		if 1+depth > 10 {
			return fmt.Errorf("move would exceed maximum folder depth of 10")
		}
		return nil
	}

	// Walk from newParentID up to root -- if folderID is encountered, it's a cycle
	currentID := types.MediaFolderID(newParentID.ID)
	ancestorDepth := 0
	for {
		ancestorDepth++
		if currentID == folderID {
			return fmt.Errorf("move would create a circular reference")
		}
		if ancestorDepth > 10 {
			return fmt.Errorf("move would exceed maximum folder depth of 10")
		}

		folder, err := d.GetMediaFolder(currentID)
		if err != nil {
			return fmt.Errorf("failed to get ancestor folder %s: %w", currentID, err)
		}
		if !folder.ParentID.Valid {
			break
		}
		currentID = types.MediaFolderID(folder.ParentID.ID)
	}

	// Calculate total depth: ancestors + 1 (for the folder itself) + subtree
	subtreeDepth, err := d.maxSubtreeDepth(folderID)
	if err != nil {
		return fmt.Errorf("failed to calculate subtree depth: %w", err)
	}
	totalDepth := ancestorDepth + 1 + subtreeDepth
	if totalDepth > 10 {
		return fmt.Errorf("move would exceed maximum folder depth of 10")
	}

	return nil
}

// maxSubtreeDepth recursively finds the deepest leaf under folderID (MySQL).
// Returns 0 if folder has no children.
func (d MysqlDatabase) maxSubtreeDepth(folderID types.MediaFolderID) (int, error) {
	children, err := d.ListMediaFoldersByParent(folderID)
	if err != nil {
		return 0, fmt.Errorf("failed to list children of folder %s: %w", folderID, err)
	}
	if children == nil || len(*children) == 0 {
		return 0, nil
	}

	maxDepth := 0
	for _, child := range *children {
		childDepth, err := d.maxSubtreeDepth(child.FolderID)
		if err != nil {
			return 0, err
		}
		depth := 1 + childDepth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth, nil
}

// ===== PostgreSQL =====

// GetMediaFolderBreadcrumb walks up the parent chain from the given folder to root (PostgreSQL).
// Returns folders in order from root to the given folder. Max 10 iterations.
func (d PsqlDatabase) GetMediaFolderBreadcrumb(folderID types.MediaFolderID) ([]MediaFolder, error) {
	var chain []MediaFolder
	currentID := folderID

	for i := range 10 {
		_ = i
		folder, err := d.GetMediaFolder(currentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get media folder %s in breadcrumb: %w", currentID, err)
		}
		chain = append(chain, *folder)

		if !folder.ParentID.Valid {
			break
		}
		currentID = types.MediaFolderID(folder.ParentID.ID)

		if len(chain) > 10 {
			return nil, fmt.Errorf("media folder breadcrumb exceeded max depth of 10 (circular reference?)")
		}
	}

	// Reverse: chain is leaf-to-root, we want root-to-leaf
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

// ValidateMediaFolderName validates folder name rules and uniqueness within parent (PostgreSQL).
// Must be called before create/update. Protected by mutex to prevent race conditions.
func (d PsqlDatabase) ValidateMediaFolderName(name string, parentID types.NullableMediaFolderID) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("media folder name cannot be empty")
	}
	if len(trimmed) > 255 {
		return fmt.Errorf("media folder name cannot exceed 255 characters")
	}
	if strings.ContainsAny(trimmed, "/\\\x00") {
		return fmt.Errorf("media folder name cannot contain '/', '\\', or null bytes")
	}
	if trimmed == "." || trimmed == ".." {
		return fmt.Errorf("media folder name cannot be %q", trimmed)
	}

	mediaFolderNameMu.Lock()
	defer mediaFolderNameMu.Unlock()

	if parentID.Valid {
		_, err := d.GetMediaFolderByNameAndParent(types.MediaFolderID(parentID.ID), trimmed)
		if err == nil {
			return fmt.Errorf("a folder named %q already exists in this parent", trimmed)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check folder name uniqueness: %w", err)
		}
	} else {
		_, err := d.GetMediaFolderByNameAtRoot(trimmed)
		if err == nil {
			return fmt.Errorf("a folder named %q already exists at root", trimmed)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to check folder name uniqueness: %w", err)
		}
	}

	return nil
}

// ValidateMediaFolderMove checks that moving a folder to a new parent won't create
// a cycle or exceed max depth (10 levels) (PostgreSQL).
func (d PsqlDatabase) ValidateMediaFolderMove(folderID types.MediaFolderID, newParentID types.NullableMediaFolderID) error {
	// Moving to root: just check subtree depth
	if !newParentID.Valid {
		depth, err := d.maxSubtreeDepth(folderID)
		if err != nil {
			return fmt.Errorf("failed to calculate subtree depth: %w", err)
		}
		// At root the folder itself is at depth 1, plus its subtree
		if 1+depth > 10 {
			return fmt.Errorf("move would exceed maximum folder depth of 10")
		}
		return nil
	}

	// Walk from newParentID up to root -- if folderID is encountered, it's a cycle
	currentID := types.MediaFolderID(newParentID.ID)
	ancestorDepth := 0
	for {
		ancestorDepth++
		if currentID == folderID {
			return fmt.Errorf("move would create a circular reference")
		}
		if ancestorDepth > 10 {
			return fmt.Errorf("move would exceed maximum folder depth of 10")
		}

		folder, err := d.GetMediaFolder(currentID)
		if err != nil {
			return fmt.Errorf("failed to get ancestor folder %s: %w", currentID, err)
		}
		if !folder.ParentID.Valid {
			break
		}
		currentID = types.MediaFolderID(folder.ParentID.ID)
	}

	// Calculate total depth: ancestors + 1 (for the folder itself) + subtree
	subtreeDepth, err := d.maxSubtreeDepth(folderID)
	if err != nil {
		return fmt.Errorf("failed to calculate subtree depth: %w", err)
	}
	totalDepth := ancestorDepth + 1 + subtreeDepth
	if totalDepth > 10 {
		return fmt.Errorf("move would exceed maximum folder depth of 10")
	}

	return nil
}

// maxSubtreeDepth recursively finds the deepest leaf under folderID (PostgreSQL).
// Returns 0 if folder has no children.
func (d PsqlDatabase) maxSubtreeDepth(folderID types.MediaFolderID) (int, error) {
	children, err := d.ListMediaFoldersByParent(folderID)
	if err != nil {
		return 0, fmt.Errorf("failed to list children of folder %s: %w", folderID, err)
	}
	if children == nil || len(*children) == 0 {
		return 0, nil
	}

	maxDepth := 0
	for _, child := range *children {
		childDepth, err := d.maxSubtreeDepth(child.FolderID)
		if err != nil {
			return 0, err
		}
		depth := 1 + childDepth
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth, nil
}
