package tui

import (
	"sort"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

// MediaNodeKind discriminates folder and file nodes in the media tree.
type MediaNodeKind int

const (
	MediaNodeFile   MediaNodeKind = iota // actual media item (leaf)
	MediaNodeFolder                      // database-backed folder
)

// MediaTreeNode represents a node in the media folder tree.
type MediaTreeNode struct {
	Kind        MediaNodeKind
	Label       string               // folder name or filename
	Path        string               // full path for sorting and grouping (legacy compat)
	FolderID    types.MediaFolderID  // non-zero for MediaNodeFolder
	Depth       int
	Expand      bool
	Media       *db.Media // non-nil only for MediaNodeFile
	FirstChild  *MediaTreeNode
	NextSibling *MediaTreeNode
	PrevSibling *MediaTreeNode
}

// hasChildren returns true if the node has any children.
func (n *MediaTreeNode) hasChildren() bool {
	return n.FirstChild != nil
}

// appendMediaChildNode appends child to parent's sibling-linked child list.
func appendMediaChildNode(parent, child *MediaTreeNode) {
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

// BuildMediaTree builds a media tree from database-backed folders and media items.
// Folders are organized by ParentID; media items are placed under their FolderID.
// Unfiled media (FolderID not valid) appears at root level.
// Folders are sorted alphabetically by Name; media items are sorted by Name within each folder.
func BuildMediaTree(folders []db.MediaFolder, items []db.Media) []*MediaTreeNode {
	if len(items) == 0 && len(folders) == 0 {
		return nil
	}

	// Index folder nodes by FolderID for parent lookup.
	folderNodes := make(map[types.MediaFolderID]*MediaTreeNode, len(folders))
	var rootFolders []*MediaTreeNode

	// Sort folders alphabetically by name for deterministic ordering.
	sortedFolders := make([]db.MediaFolder, len(folders))
	copy(sortedFolders, folders)
	sort.Slice(sortedFolders, func(i, j int) bool {
		return strings.ToLower(sortedFolders[i].Name) < strings.ToLower(sortedFolders[j].Name)
	})

	// First pass: create all folder nodes.
	for i := range sortedFolders {
		f := sortedFolders[i]
		node := &MediaTreeNode{
			Kind:     MediaNodeFolder,
			Label:    f.Name,
			Path:     f.Name,
			FolderID: f.FolderID,
			Expand:   true,
		}
		folderNodes[f.FolderID] = node
	}

	// Second pass: link folders to parents.
	for i := range sortedFolders {
		f := sortedFolders[i]
		node := folderNodes[f.FolderID]
		if f.ParentID.Valid && !f.ParentID.ID.IsZero() {
			if parent, ok := folderNodes[f.ParentID.ID]; ok {
				node.Depth = parent.Depth + 1
				appendMediaChildNode(parent, node)
				continue
			}
		}
		// Root-level folder (no parent or parent not found).
		rootFolders = append(rootFolders, node)
	}

	// Sort media items alphabetically by name for stable ordering.
	sortedItems := make([]db.Media, len(items))
	copy(sortedItems, items)
	sort.Slice(sortedItems, func(i, j int) bool {
		nameI := mediaItemSortName(sortedItems[i])
		nameJ := mediaItemSortName(sortedItems[j])
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})

	// Place media items under their folder or at root.
	var rootFiles []*MediaTreeNode
	for i := range sortedItems {
		m := sortedItems[i]
		name := mediaItemDisplayName(m)
		fileNode := &MediaTreeNode{
			Kind:  MediaNodeFile,
			Label: name,
			Path:  string(m.URL),
			Media: &m,
		}

		if m.FolderID.Valid && !m.FolderID.ID.IsZero() {
			if parent, ok := folderNodes[m.FolderID.ID]; ok {
				fileNode.Depth = parent.Depth + 1
				appendMediaChildNode(parent, fileNode)
				continue
			}
		}
		// Unfiled media goes to root.
		rootFiles = append(rootFiles, fileNode)
	}

	// Combine: folders first, then unfiled media at root.
	roots := make([]*MediaTreeNode, 0, len(rootFolders)+len(rootFiles))
	roots = append(roots, rootFolders...)
	roots = append(roots, rootFiles...)

	return roots
}

// BuildMediaTreeLegacy converts a flat media list into a URL-path-grouped folder tree.
// Items are grouped by shared path prefixes extracted from their URL field.
// Retained for import functionality that uses URL-based grouping.
func BuildMediaTreeLegacy(items []db.Media) []*MediaTreeNode {
	if len(items) == 0 {
		return nil
	}

	type pathItem struct {
		segments []string
		media    db.Media
	}

	parsed := make([]pathItem, 0, len(items))
	for _, item := range items {
		segs := splitMediaPath(string(item.URL))
		parsed = append(parsed, pathItem{segments: segs, media: item})
	}

	sort.Slice(parsed, func(i, j int) bool {
		return compareSegments(parsed[i].segments, parsed[j].segments) < 0
	})

	// Build tree using a trie-like approach: track folder nodes by path prefix.
	var roots []*MediaTreeNode
	folders := make(map[string]*MediaTreeNode)

	for i := range parsed {
		segs := parsed[i].segments
		if len(segs) == 0 {
			// Root-level file (no path segments)
			m := parsed[i].media
			roots = append(roots, &MediaTreeNode{
				Kind:  MediaNodeFile,
				Label: mediaFileName(string(m.URL)),
				Path:  string(m.URL),
				Media: &m,
			})
			continue
		}

		// Ensure all ancestor folders exist
		for depth := range len(segs) - 1 {
			prefix := strings.Join(segs[:depth+1], "/")
			if _, exists := folders[prefix]; exists {
				continue
			}
			folder := &MediaTreeNode{
				Kind:   MediaNodeFolder,
				Label:  segs[depth],
				Path:   prefix,
				Depth:  depth,
				Expand: true,
			}
			folders[prefix] = folder

			if depth == 0 {
				roots = append(roots, folder)
			} else {
				parentPrefix := strings.Join(segs[:depth], "/")
				if parent, ok := folders[parentPrefix]; ok {
					folder.Depth = parent.Depth + 1
					appendMediaChildNode(parent, folder)
				}
			}
		}

		// Add the file node
		m := parsed[i].media
		fileName := segs[len(segs)-1]
		fileNode := &MediaTreeNode{
			Kind:  MediaNodeFile,
			Label: fileName,
			Path:  strings.Join(segs, "/"),
			Media: &m,
		}

		parentPrefix := strings.Join(segs[:len(segs)-1], "/")
		if parent, ok := folders[parentPrefix]; ok {
			fileNode.Depth = parent.Depth + 1
			appendMediaChildNode(parent, fileNode)
		} else {
			roots = append(roots, fileNode)
		}
	}

	return roots
}

// FlattenMediaTree walks depth-first respecting Expand state, returning
// navigable nodes (both folders for expand/collapse and files for selection).
func FlattenMediaTree(roots []*MediaTreeNode) []*MediaTreeNode {
	var result []*MediaTreeNode
	for _, node := range roots {
		flattenMediaNode(node, &result)
	}
	return result
}

func flattenMediaNode(node *MediaTreeNode, result *[]*MediaTreeNode) {
	*result = append(*result, node)

	if !node.Expand {
		return
	}

	child := node.FirstChild
	for child != nil {
		flattenMediaNode(child, result)
		child = child.NextSibling
	}
}

// FilterMediaTree filters media items by query and returns a new tree that preserves
// ancestor folders of matching items. Folders that contain no matching descendants are removed.
func FilterMediaTree(folders []db.MediaFolder, items []db.Media, query string) ([]db.MediaFolder, []db.Media) {
	if query == "" {
		return folders, items
	}

	filteredItems := FilterMediaList(items, query)
	if len(filteredItems) == 0 {
		return nil, nil
	}

	// Collect all folder IDs that contain matching media (directly or as ancestors).
	neededFolders := make(map[types.MediaFolderID]bool)
	folderIndex := make(map[types.MediaFolderID]db.MediaFolder, len(folders))
	for _, f := range folders {
		folderIndex[f.FolderID] = f
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

	var filteredFolders []db.MediaFolder
	for _, f := range folders {
		if neededFolders[f.FolderID] {
			filteredFolders = append(filteredFolders, f)
		}
	}

	return filteredFolders, filteredItems
}

// FilterMediaList returns the subset of items matching query (case-insensitive)
// against name, display name, mimetype, and URL path.
func FilterMediaList(items []db.Media, query string) []db.Media {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var result []db.Media
	for _, item := range items {
		if matchesMediaQuery(item, q) {
			result = append(result, item)
		}
	}
	return result
}

func matchesMediaQuery(item db.Media, query string) bool {
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

// mediaItemSortName returns the best name for sorting a media item.
func mediaItemSortName(m db.Media) string {
	if m.DisplayName.Valid && m.DisplayName.String != "" {
		return m.DisplayName.String
	}
	if m.Name.Valid && m.Name.String != "" {
		return m.Name.String
	}
	return string(m.URL)
}

// mediaItemDisplayName returns the best display name for a media item.
func mediaItemDisplayName(m db.Media) string {
	if m.DisplayName.Valid && m.DisplayName.String != "" {
		return m.DisplayName.String
	}
	if m.Name.Valid && m.Name.String != "" {
		return m.Name.String
	}
	return mediaFileName(string(m.URL))
}

// splitMediaPath extracts path segments from a URL, stripping scheme and host.
func splitMediaPath(rawURL string) []string {
	path := rawURL
	// Strip scheme
	if idx := strings.Index(path, "://"); idx >= 0 {
		path = path[idx+3:]
	}
	// Strip host (everything before the first / after scheme removal)
	if idx := strings.Index(path, "/"); idx >= 0 {
		path = path[idx:]
	}
	// Strip leading /
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	var segments []string
	for _, p := range parts {
		if p != "" {
			segments = append(segments, p)
		}
	}
	return segments
}

// mediaFileName extracts the filename (last path segment) from a URL.
func mediaFileName(rawURL string) string {
	segs := splitMediaPath(rawURL)
	if len(segs) == 0 {
		return rawURL
	}
	return segs[len(segs)-1]
}

// compareSegments compares two segment slices lexicographically.
// Shorter slices sort before longer ones with the same prefix.
func compareSegments(a, b []string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := range minLen {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}
