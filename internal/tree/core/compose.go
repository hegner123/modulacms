package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
	"golang.org/x/sync/errgroup"
)

// TreeFetcher abstracts the full tree-building pipeline for composition.
// Each call fetches all data for a content tree and builds it using the
// standard core.BuildTree path.
type TreeFetcher interface {
	FetchAndBuildTree(ctx context.Context, id types.ContentID) (*Root, error)
}

// ComposeOptions controls composition behavior.
type ComposeOptions struct {
	MaxDepth       int // Maximum composition nesting depth (default 10)
	MaxConcurrency int // Max goroutines for parallel reference resolution (default 10)
}

// composeState tracks shared state across concurrent recursive calls.
type composeState struct {
	mu      sync.Mutex
	visited map[types.ContentID]bool
}

func (s *composeState) markVisited(id types.ContentID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.visited[id] {
		return false // already visited
	}
	s.visited[id] = true
	return true
}

// ComposeTrees resolves all _reference nodes in the tree by fetching
// and building the referenced content trees via the standard BuildTree
// pipeline. Operates recursively up to MaxDepth levels. Uses errgroup
// for concurrent resolution of sibling references. Broken references
// produce _system_log nodes instead of errors.
func ComposeTrees(ctx context.Context, root *Root, fetcher TreeFetcher, opts ComposeOptions) error {
	if opts.MaxDepth <= 0 {
		opts.MaxDepth = 10
	}
	if opts.MaxConcurrency <= 0 {
		opts.MaxConcurrency = 10
	}
	state := &composeState{
		visited: make(map[types.ContentID]bool),
	}
	// Mark the root content ID as visited to prevent self-reference
	if root.Node != nil && root.Node.ContentData != nil {
		state.markVisited(root.Node.ContentData.ContentDataID)
	}
	return composeTrees(ctx, root, fetcher, opts, state, 0)
}

func composeTrees(ctx context.Context, root *Root, fetcher TreeFetcher, opts ComposeOptions, state *composeState, depth int) error {
	if depth >= opts.MaxDepth {
		return nil
	}

	// Walk tree looking for _reference nodes
	refs := findReferenceNodes(root.Node)
	if len(refs) == 0 {
		return nil
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(opts.MaxConcurrency)

	for _, refNode := range refs {
		refNode := refNode // capture
		g.Go(func() error {
			return resolveReferenceNode(gctx, refNode, root, fetcher, opts, state, depth)
		})
	}

	return g.Wait()
}

// findReferenceNodes walks the tree and collects all nodes whose type is _reference.
func findReferenceNodes(node *Node) []*Node {
	if node == nil {
		return nil
	}
	var refs []*Node
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		if types.DatatypeType(n.Datatype.Type) == types.DatatypeTypeReference {
			refs = append(refs, n)
		}
		walk(n.FirstChild)
		walk(n.NextSibling)
	}
	walk(node)
	return refs
}

// resolveReferenceNode resolves all content_tree_ref field values on a _reference node.
// Subtrees are attached as children in field order.
func resolveReferenceNode(ctx context.Context, refNode *Node, root *Root, fetcher TreeFetcher, opts ComposeOptions, state *composeState, depth int) error {
	parentLabel := refNode.Datatype.Label

	// Collect content_tree_ref field values in order
	var refIDs []types.ContentID
	for i, f := range refNode.Fields {
		if f.Type != types.FieldTypeContentTreeRef {
			continue
		}
		val := refNode.ContentFields[i].FieldValue
		if val == "" {
			continue
		}
		refIDs = append(refIDs, types.ContentID(val))
	}

	// Resolve each reference sequentially (preserves ordering)
	for _, refID := range refIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		child := resolveOneReference(ctx, refID, parentLabel, root, fetcher, opts, state, depth)
		attachChildToNode(refNode, child)
	}

	return nil
}

// resolveOneReference resolves a single content_tree_ref value, returning either the
// resolved subtree root or a _system_log node on failure.
func resolveOneReference(ctx context.Context, refID types.ContentID, parentLabel string, root *Root, fetcher TreeFetcher, opts ComposeOptions, state *composeState, depth int) *Node {
	// Cycle detection
	if !state.markVisited(refID) {
		return newSystemLogNode(parentLabel, refID, "circular_reference",
			fmt.Sprintf("Reference '%s' points to content_data_id %s which creates a circular reference (this content is already in the current composition chain). Each reference must point to content outside the current composition chain.", parentLabel, refID))
	}

	// Depth check
	if depth+1 >= opts.MaxDepth {
		return newSystemLogNode(parentLabel, refID, "max_depth",
			fmt.Sprintf("Reference '%s' points to content_data_id %s but the maximum composition depth (%d) has been reached. This limit is configurable via composition_max_depth in config.json.", parentLabel, refID, opts.MaxDepth))
	}

	// Fetch and build the referenced tree
	subtree, err := fetcher.FetchAndBuildTree(ctx, refID)
	if err != nil {
		return newSystemLogNode(parentLabel, refID, "not_found",
			fmt.Sprintf("Reference '%s' points to content_data_id %s which does not exist or could not be fetched: %s. Check that the content has not been deleted, or update the reference field to point to a valid content_data_id.", parentLabel, refID, err))
	}
	if subtree == nil || subtree.Node == nil {
		return newSystemLogNode(parentLabel, refID, "build_failed",
			fmt.Sprintf("Reference '%s' resolved content_data_id %s but the subtree failed to build. The referenced content may have data integrity issues.", parentLabel, refID))
	}

	// Recursively compose the subtree
	if composeErr := composeTrees(ctx, subtree, fetcher, opts, state, depth+1); composeErr != nil {
		// Subtree composition failed, but the subtree itself is still usable
		// Log the error but don't replace the subtree with a system log
	}

	return subtree.Node
}

// attachChildToNode appends a child node as the last child of parent.
func attachChildToNode(parent, child *Node) {
	child.Parent = parent
	child.PrevSibling = nil
	child.NextSibling = nil

	if parent.FirstChild == nil {
		parent.FirstChild = child
		return
	}
	current := parent.FirstChild
	for current.NextSibling != nil {
		current = current.NextSibling
	}
	current.NextSibling = child
	child.PrevSibling = current
}

// newSystemLogNode creates a synthetic _system_log node with structured fields
// describing a composition failure.
func newSystemLogNode(parentLabel string, refID types.ContentID, errType, message string) *Node {
	nodeID := types.NewContentID()
	dtID := types.NewDatatypeID()

	errorFieldID := types.NewFieldID()
	messageFieldID := types.NewFieldID()
	refFieldID := types.NewFieldID()
	parentFieldID := types.NewFieldID()

	now := types.TimestampNow()

	return &Node{
		ContentData: &db.ContentData{
			ContentDataID: nodeID,
			DatatypeID:    types.NullableDatatypeID{ID: dtID, Valid: true},
			Status:        types.ContentStatusPublished,
			DateCreated:   now,
			DateModified:  now,
		},
		Datatype: db.Datatypes{
			DatatypeID:   dtID,
			Label:        "_system_log",
			Type:         string(types.DatatypeTypeSystemLog),
			DateCreated:  now,
			DateModified: now,
		},
		ContentFields: []db.ContentFields{
			{
				ContentFieldID: types.NewContentFieldID(),
				ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
				FieldID:        types.NullableFieldID{ID: errorFieldID, Valid: true},
				FieldValue:     errType,
				DateCreated:    now,
				DateModified:   now,
			},
			{
				ContentFieldID: types.NewContentFieldID(),
				ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
				FieldID:        types.NullableFieldID{ID: messageFieldID, Valid: true},
				FieldValue:     message,
				DateCreated:    now,
				DateModified:   now,
			},
			{
				ContentFieldID: types.NewContentFieldID(),
				ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
				FieldID:        types.NullableFieldID{ID: refFieldID, Valid: true},
				FieldValue:     refID.String(),
				DateCreated:    now,
				DateModified:   now,
			},
			{
				ContentFieldID: types.NewContentFieldID(),
				ContentDataID:  types.NullableContentID{ID: nodeID, Valid: true},
				FieldID:        types.NullableFieldID{ID: parentFieldID, Valid: true},
				FieldValue:     parentLabel,
				DateCreated:    now,
				DateModified:   now,
			},
		},
		Fields: []db.Fields{
			{
				FieldID:      errorFieldID,
				Label:        "error",
				Type:         types.FieldTypeText,
				DateCreated:  now,
				DateModified: now,
			},
			{
				FieldID:      messageFieldID,
				Label:        "message",
				Type:         types.FieldTypeText,
				DateCreated:  now,
				DateModified: now,
			},
			{
				FieldID:      refFieldID,
				Label:        "reference_id",
				Type:         types.FieldTypeText,
				DateCreated:  now,
				DateModified: now,
			},
			{
				FieldID:      parentFieldID,
				Label:        "parent_label",
				Type:         types.FieldTypeText,
				DateCreated:  now,
				DateModified: now,
			},
		},
	}
}
