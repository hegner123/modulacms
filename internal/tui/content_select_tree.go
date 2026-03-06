package tui

import (
	"sort"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
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
func BuildContentSelectTree(items []db.ContentDataTopLevel) []*ContentSelectNode {
	var routed, standalone []db.ContentDataTopLevel
	for i := range items {
		if items[i].RouteID.Valid {
			routed = append(routed, items[i])
		} else {
			standalone = append(standalone, items[i])
		}
	}

	var roots []*ContentSelectNode

	if len(routed) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Pages"})
		roots = append(roots, buildRoutedNodes(routed)...)
	}

	if len(standalone) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Standalone"})
		roots = append(roots, buildStandaloneNodes(standalone)...)
	}

	return roots
}

// BuildAdminContentSelectTree builds the slug tree for admin content items.
func BuildAdminContentSelectTree(items []db.AdminContentDataTopLevel) []*ContentSelectNode {
	var routed, standalone []db.AdminContentDataTopLevel
	for i := range items {
		if items[i].AdminRouteID.Valid {
			routed = append(routed, items[i])
		} else {
			standalone = append(standalone, items[i])
		}
	}

	var roots []*ContentSelectNode

	if len(routed) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Pages"})
		roots = append(roots, buildAdminRoutedNodes(routed)...)
	}

	if len(standalone) > 0 {
		roots = append(roots, &ContentSelectNode{Kind: NodeSection, Label: "Standalone"})
		roots = append(roots, buildAdminStandaloneNodes(standalone)...)
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
func buildStandaloneNodes(items []db.ContentDataTopLevel) []*ContentSelectNode {
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
			title := string(item.RouteTitle)
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
func buildAdminStandaloneNodes(items []db.AdminContentDataTopLevel) []*ContentSelectNode {
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
			title := string(item.RouteTitle)
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
