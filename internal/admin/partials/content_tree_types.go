package partials

import "github.com/hegner123/modulacms/internal/db/types"

// RouteTreeNode represents a node in the route-segment tree.
// Interior nodes (like "docs" or "building-content") may not have a route themselves;
// they are just grouping labels derived from slug segments.
// Leaf nodes (or interior nodes that happen to match a route) have IsRoute=true
// and carry the RouteID and ContentDataID for navigation.
type RouteTreeNode struct {
	// Segment is the slug segment name for this node (e.g. "docs", "building-content", "media").
	Segment string
	// Label is the human-readable display name. Uses route Title when available,
	// falls back to the segment name with dashes replaced by spaces and title-cased.
	Label string
	// IsRoute is true when this node corresponds to an actual route in the database.
	IsRoute bool
	// RouteID is set when IsRoute is true.
	RouteID types.RouteID
	// ContentDataID is the root content_data ID for this route (for page view navigation).
	ContentDataID types.ContentID
	// IsGlobal is true for unrouted global content nodes.
	IsGlobal bool
	// Children are the nested sub-segment nodes.
	Children []RouteTreeNode
}

// ContentBlockSummary represents a content block displayed inline on the page view.
type ContentBlockSummary struct {
	ContentDataID types.ContentID
	DatatypeLabel string
	DatatypeType  string
	// ResolvedLabel is a human-readable name for reference types (e.g. "main Navigation"
	// instead of a raw content ID). Empty for non-reference blocks.
	ResolvedLabel string
	Fields        []BlockFieldSummary
	ChildCount    int
	Depth         int
	// IsDirty is true when the block has been modified since last publish.
	IsDirty bool
}

// BlockFieldSummary holds a field value for inline display.
type BlockFieldSummary struct {
	Label string
	Type  string
	Value string
	// ResolvedValue is a human-readable resolved value for _id fields.
	// When set, displayed instead of Value.
	ResolvedValue string
}

// ContentBreadcrumb holds the navigation segments for the breadcrumb bar.
type ContentBreadcrumb struct {
	Segments []BreadcrumbSegment
}

// BreadcrumbSegment is a single clickable segment in the breadcrumb.
type BreadcrumbSegment struct {
	Label string
	Href  string
}
