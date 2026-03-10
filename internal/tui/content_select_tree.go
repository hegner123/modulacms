package tui

import (
	"context"
	"sort"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"github.com/hegner123/modulacms/internal/utility"
)

// ContentSelectNodeKind discriminates the three node types in the select tree.
type ContentSelectNodeKind int

const (
	NodeContent ContentSelectNodeKind = iota // selectable content item
	NodeSection                              // non-selectable section header ("Pages", "Standalone")
	NodeGroup                                // collapsible group (slug prefix or datatype label)
)

// ContentSelectNode represents a node in the slug-grouped select tree.
// Children are linked via sibling pointers (FirstChild/NextSibling/PrevSibling),
// matching the tree.Node pattern used elsewhere in the codebase.
type ContentSelectNode struct {
	Kind         ContentSelectNodeKind
	Label        string                       // display label: "/" or "about" or "Main Menu"
	Slug         string                       // full slug for sorting (empty for standalone)
	Depth        int                          // visual indent level
	Expand       bool                         // collapsible group state (only meaningful for NodeGroup)
	Content      *db.ContentDataTopLevel      // non-nil for regular mode NodeContent
	AdminContent *db.AdminContentDataTopLevel // non-nil for admin mode NodeContent
	FirstChild   *ContentSelectNode           // first child node
	NextSibling  *ContentSelectNode           // next sibling
	PrevSibling  *ContentSelectNode           // previous sibling
}

// appendChildNode appends child to parent's sibling-linked child list.
func appendChildNode(parent, child *ContentSelectNode) {
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

// hasChildren returns true if the node has any children.
func (n *ContentSelectNode) hasChildren() bool {
	return n.FirstChild != nil
}

// BuildContentSelectTree builds a slug-grouped hierarchical tree from flat
// top-level content items. Routed items are grouped by slug prefix under a
// "Pages" section. Standalone items (no route) are grouped by datatype label
// under a "Standalone" section.
func BuildContentSelectTree(items []db.ContentDataTopLevel, titleMap map[string]string) []*ContentSelectNode {
	var routed, globals, standalone []db.ContentDataTopLevel
	for i := range items {
		if items[i].RouteID.Valid {
			routed = append(routed, items[i])
		} else if types.DatatypeType(items[i].DatatypeType).IsGlobalType() {
			globals = append(globals, items[i])
		} else {
			standalone = append(standalone, items[i])
		}
	}

	var roots []*ContentSelectNode

	if len(routed) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Pages"})
		roots = append(roots, buildRoutedNodes(routed)...)
	}

	if len(globals) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Globals"})
		roots = append(roots, buildStandaloneNodes(globals, titleMap)...)
	}

	if len(standalone) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Standalone"})
		roots = append(roots, buildStandaloneNodes(standalone, titleMap)...)
	}

	return roots
}

// BuildAdminContentSelectTree builds the slug tree for admin content items.
func BuildAdminContentSelectTree(items []db.AdminContentDataTopLevel, titleMap map[string]string) []*ContentSelectNode {
	var routed, globals, standalone []db.AdminContentDataTopLevel
	for i := range items {
		if items[i].AdminRouteID.Valid {
			routed = append(routed, items[i])
		} else if types.DatatypeType(items[i].DatatypeType).IsGlobalType() {
			globals = append(globals, items[i])
		} else {
			standalone = append(standalone, items[i])
		}
	}

	var roots []*ContentSelectNode

	if len(routed) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Pages"})
		roots = append(roots, buildAdminRoutedNodes(routed)...)
	}

	if len(globals) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Globals"})
		roots = append(roots, buildAdminStandaloneNodes(globals, titleMap)...)
	}

	if len(standalone) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Standalone"})
		roots = append(roots, buildAdminStandaloneNodes(standalone, titleMap)...)
	}

	return roots
}

// FlattenSelectTree produces a flat slice for cursor navigation by walking
// depth-first, respecting Expand state. NodeSection items are excluded —
// cursor always maps to a selectable/expandable item.
func FlattenSelectTree(roots []*ContentSelectNode) []*ContentSelectNode {
	var result []*ContentSelectNode
	for _, node := range roots {
		flattenNode(node, &result)
	}
	return result
}

func flattenNode(node *ContentSelectNode, result *[]*ContentSelectNode) {
	if node.Kind == NodeSection {
		return
	}

	*result = append(*result, node)

	if node.Expand {
		child := node.FirstChild
		for child != nil {
			flattenNode(child, result)
			child = child.NextSibling
		}
	}
}

// splitSlug splits a slug into segments, filtering empty strings from leading "/".
func splitSlug(slug string) []string {
	parts := strings.Split(string(slug), "/")
	var segments []string
	for _, p := range parts {
		if p != "" {
			segments = append(segments, p)
		}
	}
	return segments
}

// buildRoutedNodes sorts routed content by slug segments and groups items
// sharing a first segment under collapsible group nodes.
func buildRoutedNodes(items []db.ContentDataTopLevel) []*ContentSelectNode {
	sort.Slice(items, func(i, j int) bool {
		return compareSlugSegments(string(items[i].RouteSlug), string(items[j].RouteSlug)) < 0
	})

	// Group by first segment
	type group struct {
		prefix string
		items  []db.ContentDataTopLevel
	}
	var groups []group
	for _, item := range items {
		segs := splitSlug(string(item.RouteSlug))
		prefix := ""
		if len(segs) > 0 {
			prefix = segs[0]
		}

		if len(groups) == 0 || groups[len(groups)-1].prefix != prefix {
			groups = append(groups, group{prefix: prefix, items: []db.ContentDataTopLevel{item}})
		} else {
			groups[len(groups)-1].items = append(groups[len(groups)-1].items, item)
		}
	}

	var nodes []*ContentSelectNode
	for _, g := range groups {
		if len(g.items) == 1 {
			item := g.items[0]
			nodes = append(nodes, &ContentSelectNode{
				Kind:    NodeContent,
				Label:   slugLabel(string(item.RouteSlug)),
				Slug:    string(item.RouteSlug),
				Depth:   0,
				Content: &item,
			})
			continue
		}

		// Multiple items share this prefix — first item might be the prefix itself
		groupNode := &ContentSelectNode{
			Kind:   NodeGroup,
			Label:  "/" + g.prefix,
			Slug:   "/" + g.prefix,
			Depth:  0,
			Expand: true,
		}

		for _, item := range g.items {
			segs := splitSlug(string(item.RouteSlug))
			if len(segs) <= 1 {
				// This is the prefix item itself (e.g., "/about") — keep Kind as
				// NodeGroup so the renderer recurses children. Content is set so
				// it remains selectable and shows status/datatype.
				groupNode.Content = func() *db.ContentDataTopLevel { c := item; return &c }()
				groupNode.Label = slugLabel(string(item.RouteSlug))
				groupNode.Slug = string(item.RouteSlug)
				continue
			}
			// Child of the group
			child := &ContentSelectNode{
				Kind:    NodeContent,
				Label:   slugLabel(string(item.RouteSlug)),
				Slug:    string(item.RouteSlug),
				Depth:   1,
				Content: func() *db.ContentDataTopLevel { c := item; return &c }(),
			}
			appendChildNode(groupNode, child)
		}

		nodes = append(nodes, groupNode)
	}

	return nodes
}

// buildAdminRoutedNodes is the admin variant of buildRoutedNodes.
func buildAdminRoutedNodes(items []db.AdminContentDataTopLevel) []*ContentSelectNode {
	sort.Slice(items, func(i, j int) bool {
		return compareSlugSegments(string(items[i].RouteSlug), string(items[j].RouteSlug)) < 0
	})

	type group struct {
		prefix string
		items  []db.AdminContentDataTopLevel
	}
	var groups []group
	for _, item := range items {
		segs := splitSlug(string(item.RouteSlug))
		prefix := ""
		if len(segs) > 0 {
			prefix = segs[0]
		}
		if len(groups) == 0 || groups[len(groups)-1].prefix != prefix {
			groups = append(groups, group{prefix: prefix, items: []db.AdminContentDataTopLevel{item}})
		} else {
			groups[len(groups)-1].items = append(groups[len(groups)-1].items, item)
		}
	}

	var nodes []*ContentSelectNode
	for _, g := range groups {
		if len(g.items) == 1 {
			item := g.items[0]
			nodes = append(nodes, &ContentSelectNode{
				Kind:         NodeContent,
				Label:        slugLabel(string(item.RouteSlug)),
				Slug:         string(item.RouteSlug),
				Depth:        0,
				AdminContent: &item,
			})
			continue
		}

		groupNode := &ContentSelectNode{
			Kind:   NodeGroup,
			Label:  "/" + g.prefix,
			Slug:   "/" + g.prefix,
			Depth:  0,
			Expand: true,
		}

		for _, item := range g.items {
			segs := splitSlug(string(item.RouteSlug))
			if len(segs) <= 1 {
				groupNode.AdminContent = func() *db.AdminContentDataTopLevel { c := item; return &c }()
				groupNode.Label = slugLabel(string(item.RouteSlug))
				groupNode.Slug = string(item.RouteSlug)
				continue
			}
			child := &ContentSelectNode{
				Kind:         NodeContent,
				Label:        slugLabel(string(item.RouteSlug)),
				Slug:         string(item.RouteSlug),
				Depth:        1,
				AdminContent: func() *db.AdminContentDataTopLevel { c := item; return &c }(),
			}
			appendChildNode(groupNode, child)
		}

		nodes = append(nodes, groupNode)
	}

	return nodes
}

// buildStandaloneNodes groups standalone content by datatype label.
func buildStandaloneNodes(items []db.ContentDataTopLevel, titleMap map[string]string) []*ContentSelectNode {
	sort.Slice(items, func(i, j int) bool {
		if items[i].DatatypeLabel != items[j].DatatypeLabel {
			return items[i].DatatypeLabel < items[j].DatatypeLabel
		}
		return items[i].ContentDataID < items[j].ContentDataID
	})

	type group struct {
		label string
		items []db.ContentDataTopLevel
	}
	var groups []group
	for _, item := range items {
		lbl := item.DatatypeLabel
		if lbl == "" {
			lbl = "Other"
		}
		if len(groups) == 0 || groups[len(groups)-1].label != lbl {
			groups = append(groups, group{label: lbl, items: []db.ContentDataTopLevel{item}})
		} else {
			groups[len(groups)-1].items = append(groups[len(groups)-1].items, item)
		}
	}

	var nodes []*ContentSelectNode
	for _, g := range groups {
		groupNode := &ContentSelectNode{
			Kind:   NodeGroup,
			Label:  g.label,
			Depth:  0,
			Expand: true,
		}
		for _, item := range g.items {
			title := titleMap[string(item.ContentDataID)]
			if title == "" {
				title = string(item.RouteTitle)
			}
			if title == "" {
				title = item.DatatypeLabel
			}
			if title == "" {
				title = string(item.ContentDataID)
			}
			child := &ContentSelectNode{
				Kind:    NodeContent,
				Label:   title,
				Depth:   1,
				Content: func() *db.ContentDataTopLevel { c := item; return &c }(),
			}
			appendChildNode(groupNode, child)
		}
		nodes = append(nodes, groupNode)
	}

	return nodes
}

// buildAdminStandaloneNodes is the admin variant of buildStandaloneNodes.
func buildAdminStandaloneNodes(items []db.AdminContentDataTopLevel, titleMap map[string]string) []*ContentSelectNode {
	sort.Slice(items, func(i, j int) bool {
		if items[i].DatatypeLabel != items[j].DatatypeLabel {
			return items[i].DatatypeLabel < items[j].DatatypeLabel
		}
		return items[i].AdminContentDataID < items[j].AdminContentDataID
	})

	type group struct {
		label string
		items []db.AdminContentDataTopLevel
	}
	var groups []group
	for _, item := range items {
		lbl := item.DatatypeLabel
		if lbl == "" {
			lbl = "Other"
		}
		if len(groups) == 0 || groups[len(groups)-1].label != lbl {
			groups = append(groups, group{label: lbl, items: []db.AdminContentDataTopLevel{item}})
		} else {
			groups[len(groups)-1].items = append(groups[len(groups)-1].items, item)
		}
	}

	var nodes []*ContentSelectNode
	for _, g := range groups {
		groupNode := &ContentSelectNode{
			Kind:   NodeGroup,
			Label:  g.label,
			Depth:  0,
			Expand: true,
		}
		for _, item := range g.items {
			title := titleMap[string(item.AdminContentDataID)]
			if title == "" {
				title = string(item.RouteTitle)
			}
			if title == "" {
				title = item.DatatypeLabel
			}
			if title == "" {
				title = string(item.AdminContentDataID)
			}
			child := &ContentSelectNode{
				Kind:         NodeContent,
				Label:        title,
				Depth:        1,
				AdminContent: func() *db.AdminContentDataTopLevel { c := item; return &c }(),
			}
			appendChildNode(groupNode, child)
		}
		nodes = append(nodes, groupNode)
	}

	return nodes
}

// compareSlugSegments compares two slugs by their segments lexicographically.
// "/" (homepage, 0 segments) sorts first.
func compareSlugSegments(a, b string) int {
	sa := splitSlug(a)
	sb := splitSlug(b)

	// Homepage (0 segments) sorts first
	if len(sa) == 0 && len(sb) > 0 {
		return -1
	}
	if len(sa) > 0 && len(sb) == 0 {
		return 1
	}

	minLen := len(sa)
	if len(sb) < minLen {
		minLen = len(sb)
	}

	for i := range minLen {
		if sa[i] < sb[i] {
			return -1
		}
		if sa[i] > sb[i] {
			return 1
		}
	}

	if len(sa) < len(sb) {
		return -1
	}
	if len(sa) > len(sb) {
		return 1
	}
	return 0
}

// slugLabel returns a display label for a slug.
// "/" stays as "/", "/about" becomes "/about", "/about/team" becomes "/about/team".
func slugLabel(slug string) string {
	if slug == "" || slug == "/" {
		return "/"
	}
	return slug
}

// resolveTitleFields builds a map of ContentDataID → title by finding fields
// with type _title and looking up their content field values. Only queries for
// standalone/global items (no route).
func resolveTitleFields(d db.DbDriver, items []db.ContentDataTopLevel) map[string]string {
	// Collect content IDs of standalone/global items that need title resolution
	var contentIDs []types.ContentID
	for _, item := range items {
		if item.RouteID.Valid {
			continue
		}
		contentIDs = append(contentIDs, item.ContentDataID)
	}
	if len(contentIDs) == 0 {
		return nil
	}

	// Find all _title field IDs
	allFields, err := d.ListFields()
	if err != nil || allFields == nil {
		utility.DefaultLogger.Error("resolveTitleFields: ListFields failed", err)
		return nil
	}
	titleFieldIDs := make(map[string]struct{})
	for _, f := range *allFields {
		if f.Type == types.FieldTypeTitle {
			titleFieldIDs[string(f.FieldID)] = struct{}{}
		}
	}
	if len(titleFieldIDs) == 0 {
		return nil
	}

	// Batch-fetch content fields for these content items
	contentFields, err := d.ListContentFieldsByContentDataIDs(context.Background(), contentIDs, "")
	if err != nil || contentFields == nil {
		utility.DefaultLogger.Error("resolveTitleFields: ListContentFieldsByContentDataIDs failed", err)
		return nil
	}

	// Build title map: first _title field value wins per content item
	titleMap := make(map[string]string)
	for _, cf := range *contentFields {
		if !cf.FieldID.Valid {
			continue
		}
		if _, ok := titleFieldIDs[string(cf.FieldID.ID)]; !ok {
			continue
		}
		cid := string(cf.ContentDataID.ID)
		if _, exists := titleMap[cid]; !exists && cf.FieldValue != "" {
			titleMap[cid] = cf.FieldValue
		}
	}
	return titleMap
}

// resolveAdminTitleFields is the admin variant of resolveTitleFields.
func resolveAdminTitleFields(d db.DbDriver, items []db.AdminContentDataTopLevel) map[string]string {
	var contentIDs []types.AdminContentID
	for _, item := range items {
		if item.AdminRouteID.Valid {
			continue
		}
		contentIDs = append(contentIDs, item.AdminContentDataID)
	}
	if len(contentIDs) == 0 {
		return nil
	}

	allFields, err := d.ListAdminFields()
	if err != nil || allFields == nil {
		utility.DefaultLogger.Error("resolveAdminTitleFields: ListAdminFields failed", err)
		return nil
	}
	titleFieldIDs := make(map[string]struct{})
	for _, f := range *allFields {
		if f.Type == types.FieldTypeTitle {
			titleFieldIDs[string(f.AdminFieldID)] = struct{}{}
		}
	}
	if len(titleFieldIDs) == 0 {
		return nil
	}

	contentFields, err := d.ListAdminContentFieldsByContentDataIDs(context.Background(), contentIDs, "")
	if err != nil || contentFields == nil {
		utility.DefaultLogger.Error("resolveAdminTitleFields: ListAdminContentFieldsByContentDataIDs failed", err)
		return nil
	}

	titleMap := make(map[string]string)
	for _, cf := range *contentFields {
		if !cf.AdminFieldID.Valid {
			continue
		}
		if _, ok := titleFieldIDs[string(cf.AdminFieldID.ID)]; !ok {
			continue
		}
		cid := string(cf.AdminContentDataID.ID)
		if _, exists := titleMap[cid]; !exists && cf.AdminFieldValue != "" {
			titleMap[cid] = cf.AdminFieldValue
		}
	}
	return titleMap
}
