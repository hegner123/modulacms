package tui

import (
	"sort"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// AdminMediaTreeNode represents a node in the admin media folder tree.
type AdminMediaTreeNode struct {
	Kind        MediaNodeKind
	Label       string
	Path        string
	FolderID    types.AdminMediaFolderID
	Depth       int
	Expand      bool
	Media       *db.AdminMedia
	FirstChild  *AdminMediaTreeNode
	NextSibling *AdminMediaTreeNode
	PrevSibling *AdminMediaTreeNode
}

// hasChildren returns true if the node has any children.
func (n *AdminMediaTreeNode) hasChildren() bool {
	return n.FirstChild != nil
}

// appendAdminMediaChildNode appends child to parent's sibling-linked child list.
func appendAdminMediaChildNode(parent, child *AdminMediaTreeNode) {
	if parent.FirstChild == nil {
		parent.FirstChild = child
		return
	}
	last := parent.FirstChild
	for last.NextSibling != nil {
		last = last.NextSibling
	}
	last.NextSibling = child
	child.PrevSibling = last
}

// BuildAdminMediaTree builds an admin media tree from database-backed folders and media items.
// Mirrors BuildMediaTree but uses admin types.
func BuildAdminMediaTree(folders []db.AdminMediaFolder, items []db.AdminMedia) []*AdminMediaTreeNode {
	if len(items) == 0 && len(folders) == 0 {
		return nil
	}

	// Index folder nodes by FolderID for parent lookup.
	folderNodes := make(map[types.AdminMediaFolderID]*AdminMediaTreeNode, len(folders))
	var rootFolders []*AdminMediaTreeNode

	// Sort folders alphabetically by name for deterministic ordering.
	sortedFolders := make([]db.AdminMediaFolder, len(folders))
	copy(sortedFolders, folders)
	sort.Slice(sortedFolders, func(i, j int) bool {
		return strings.ToLower(sortedFolders[i].Name) < strings.ToLower(sortedFolders[j].Name)
	})

	// First pass: create all folder nodes.
	for i := range sortedFolders {
		f := sortedFolders[i]
		node := &AdminMediaTreeNode{
			Kind:     MediaNodeFolder,
			Label:    f.Name,
			Path:     f.Name,
			FolderID: f.AdminFolderID,
			Expand:   true,
		}
		folderNodes[f.AdminFolderID] = node
	}

	// Second pass: link folders to parents.
	for i := range sortedFolders {
		f := sortedFolders[i]
		node := folderNodes[f.AdminFolderID]
		if f.ParentID.Valid && !f.ParentID.ID.IsZero() {
			if parent, ok := folderNodes[f.ParentID.ID]; ok {
				node.Depth = parent.Depth + 1
				appendAdminMediaChildNode(parent, node)
				continue
			}
		}
		// Root-level folder (no parent or parent not found).
		rootFolders = append(rootFolders, node)
	}

	// Sort media items alphabetically by name for stable ordering.
	sortedItems := make([]db.AdminMedia, len(items))
	copy(sortedItems, items)
	sort.Slice(sortedItems, func(i, j int) bool {
		nameI := adminMediaItemSortName(sortedItems[i])
		nameJ := adminMediaItemSortName(sortedItems[j])
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})

	// Place media items under their folder or at root.
	var rootFiles []*AdminMediaTreeNode
	for i := range sortedItems {
		m := sortedItems[i]
		name := adminMediaItemDisplayName(m)
		fileNode := &AdminMediaTreeNode{
			Kind:  MediaNodeFile,
			Label: name,
			Path:  string(m.URL),
			Media: &m,
		}

		if m.FolderID.Valid && !m.FolderID.ID.IsZero() {
			if parent, ok := folderNodes[m.FolderID.ID]; ok {
				fileNode.Depth = parent.Depth + 1
				appendAdminMediaChildNode(parent, fileNode)
				continue
			}
		}
		// Unfiled media goes to root.
		rootFiles = append(rootFiles, fileNode)
	}

	// Combine: folders first, then unfiled media at root.
	roots := make([]*AdminMediaTreeNode, 0, len(rootFolders)+len(rootFiles))
	roots = append(roots, rootFolders...)
	roots = append(roots, rootFiles...)

	return roots
}

// FlattenAdminMediaTree walks depth-first respecting Expand state, returning
// navigable nodes (both folders for expand/collapse and files for selection).
func FlattenAdminMediaTree(roots []*AdminMediaTreeNode) []*AdminMediaTreeNode {
	var result []*AdminMediaTreeNode
	for _, node := range roots {
		flattenAdminMediaNode(node, &result)
	}
	return result
}

func flattenAdminMediaNode(node *AdminMediaTreeNode, result *[]*AdminMediaTreeNode) {
	*result = append(*result, node)

	if !node.Expand {
		return
	}

	child := node.FirstChild
	for child != nil {
		flattenAdminMediaNode(child, result)
		child = child.NextSibling
	}
}

// FilterAdminMediaTree filters admin media items by query and returns a new tree
// that preserves ancestor folders of matching items.
func FilterAdminMediaTree(folders []db.AdminMediaFolder, items []db.AdminMedia, query string) ([]db.AdminMediaFolder, []db.AdminMedia) {
	if query == "" {
		return folders, items
	}

	filteredItems := FilterAdminMediaList(items, query)
	if len(filteredItems) == 0 {
		return nil, nil
	}

	// Collect all folder IDs that contain matching media (directly or as ancestors).
	neededFolders := make(map[types.AdminMediaFolderID]bool)
	folderIndex := make(map[types.AdminMediaFolderID]db.AdminMediaFolder, len(folders))
	for _, f := range folders {
		folderIndex[f.AdminFolderID] = f
	}

	for _, item := range filteredItems {
		if item.FolderID.Valid && !item.FolderID.ID.IsZero() {
			// Walk up the folder chain to mark all ancestors as needed.
			fid := item.FolderID.ID
			for !fid.IsZero() {
				if neededFolders[fid] {
					break // already traversed this branch
				}
				neededFolders[fid] = true
				if f, ok := folderIndex[fid]; ok && f.ParentID.Valid && !f.ParentID.ID.IsZero() {
					fid = f.ParentID.ID
				} else {
					break
				}
			}
		}
	}

	var filteredFolders []db.AdminMediaFolder
	for _, f := range folders {
		if neededFolders[f.AdminFolderID] {
			filteredFolders = append(filteredFolders, f)
		}
	}

	return filteredFolders, filteredItems
}

// FilterAdminMediaList returns the subset of items matching query (case-insensitive)
// against name, display name, mimetype, and URL path.
func FilterAdminMediaList(items []db.AdminMedia, query string) []db.AdminMedia {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var result []db.AdminMedia
	for _, item := range items {
		if matchesAdminMediaQuery(item, q) {
			result = append(result, item)
		}
	}
	return result
}

func matchesAdminMediaQuery(item db.AdminMedia, query string) bool {
	if item.Name.Valid && strings.Contains(strings.ToLower(item.Name.String), query) {
		return true
	}
	if item.DisplayName.Valid && strings.Contains(strings.ToLower(item.DisplayName.String), query) {
		return true
	}
	if item.Mimetype.Valid && strings.Contains(strings.ToLower(item.Mimetype.String), query) {
		return true
	}
	if strings.Contains(strings.ToLower(string(item.URL)), query) {
		return true
	}
	return false
}

// adminMediaItemSortName returns the best name for sorting an admin media item.
func adminMediaItemSortName(m db.AdminMedia) string {
	if m.DisplayName.Valid && m.DisplayName.String != "" {
		return m.DisplayName.String
	}
	if m.Name.Valid && m.Name.String != "" {
		return m.Name.String
	}
	return string(m.URL)
}

// adminMediaItemDisplayName returns the best display name for an admin media item.
func adminMediaItemDisplayName(m db.AdminMedia) string {
	if m.DisplayName.Valid && m.DisplayName.String != "" {
		return m.DisplayName.String
	}
	if m.Name.Valid && m.Name.String != "" {
		return m.Name.String
	}
	return mediaFileName(string(m.URL))
}
