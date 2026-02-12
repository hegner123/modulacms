package model

import (
	"fmt"
	"strings"

	"github.com/hegner123/modulacms/internal/db"
)

// BuildTree constructs a content tree from public (non-admin) database entities.
// It accepts four parallel slices that are assumed to be correlated by index:
//   - cd and dt are paired: cd[i] is the content instance, dt[i] is its type definition.
//   - cf and df are paired: cf[i] is the field value, df[i] is its field definition.
//
// Each pair is mapped into the model's Datatype and Field types using the db
// package's MapXxxJSON converters, then passed to BuildNodes to assemble the
// hierarchy.
//
// Called by the router layer (slugs.go) to build trees for public API responses.
func BuildTree(log Logger, cd []db.ContentData, dt []db.Datatypes, cf []db.ContentFields, df []db.Fields) (Root, error) {
	if len(cd) != len(dt) {
		return Root{}, fmt.Errorf("BuildTree: content data length (%d) != datatypes length (%d)", len(cd), len(dt))
	}
	if len(cf) != len(df) {
		return Root{}, fmt.Errorf("BuildTree: content fields length (%d) != fields length (%d)", len(cf), len(df))
	}

	// Map each content-data/datatype pair into a model Datatype.
	d := make([]Datatype, len(cd))
	f := make([]Field, len(cf))
	for i, v := range cd {
		d[i].Info = db.MapDatatypeJSON(dt[i])
		d[i].Content = db.MapContentDataJSON(v)
	}

	// Map each content-field/field-definition pair into a model Field.
	for i, v := range cf {
		f[i].Info = db.MapFieldJSON(df[i])
		f[i].Content = db.MapContentFieldJSON(v)
	}

	// Assemble the flat slices into a parent-child tree and wrap in a Root.
	nodes, err := BuildNodes(log, d, f)
	root := NewRoot()
	root.Node = nodes

	return root, err
}

// BuildAdminTree constructs a content tree from admin database entities.
// Same structure as BuildTree but uses admin-prefixed DB types.
//
// Called by the router layer (adminTree.go) to build trees for admin API responses.
//
// Note: Mutates the mapped FieldsJSON.ParentID after creation because
// AdminFields.ParentID refers to the datatype owner (type-level), but
// BuildNodes needs ParentID to point to the content data instance. The fix
// reassigns ParentID from AdminContentFields.AdminContentDataID.
func BuildAdminTree(log Logger, cd []db.AdminContentData, dt []db.AdminDatatypes, cf []db.AdminContentFields, df []db.AdminFields) (Root, error) {
	if len(cd) != len(dt) {
		return Root{}, fmt.Errorf("BuildAdminTree: content data length (%d) != datatypes length (%d)", len(cd), len(dt))
	}
	if len(cf) != len(df) {
		return Root{}, fmt.Errorf("BuildAdminTree: content fields length (%d) != fields length (%d)", len(cf), len(df))
	}

	// Map each admin content-data/datatype pair into a model Datatype.
	d := make([]Datatype, len(cd))
	f := make([]Field, len(cf))
	for i, v := range cd {
		d[i].Info = db.MapAdminDatatypeJSON(dt[i])
		d[i].Content = db.MapAdminContentDataJSON(v)
	}

	// Map each admin content-field/field-definition pair into a model Field.
	// The ParentID override is necessary because AdminFields.ParentID is the
	// datatype owner (field definition scope), not the content node the field
	// value belongs to. We replace it with AdminContentDataID so BuildNodes
	// can match fields to the correct nodes by content instance ID.
	for i, v := range cf {
		info := db.MapAdminFieldJSON(df[i])
		info.ParentID = v.AdminContentDataID.String()
		f[i].Info = info
		f[i].Content = db.MapAdminContentFieldJSON(v)
	}

	nodes, err := BuildNodes(log, d, f)
	root := NewRoot()
	root.Node = nodes

	return root, err
}

// BuildNodes is the core tree-assembly algorithm. It takes flat slices of
// Datatypes (one per content node) and Fields (one per field value) and
// produces a tree by:
//
//  1. Creating one Node per Datatype with empty Fields and Nodes slices.
//  2. Building a map[string]*Node index for O(1) lookups by ContentDataID.
//  3. Identifying the root node (Type == "ROOT").
//  4. Linking each non-root node to its parent via ContentData.ParentID.
//  5. Attaching fields to their owning node via Field.Info.ParentID.
//
// Returns the root Node pointer (nil if no node has Type "ROOT") and an error
// if orphan nodes or fields were encountered.
func BuildNodes(log Logger, datatypes []Datatype, fields []Field) (*Node, error) {
	// Phase 1: Create a flat slice of Node pointers, one per Datatype.
	// Fields and Nodes are initialized to empty slices (not nil) so they
	// marshal to JSON as [] rather than null.
	nodes := make([]*Node, len(datatypes))
	for i, dt := range datatypes {
		nodes[i] = &Node{
			Datatype: dt,
			Fields:   []Field{},
			Nodes:    []*Node{},
		}
	}

	// Build a map index for O(1) lookups by ContentDataID.
	// The nodes slice is still used for ordered iteration in phases 2 and 3.
	nodeIndex := make(map[string]*Node, len(nodes))
	for _, n := range nodes {
		nodeIndex[n.Datatype.Content.ContentDataID] = n
	}

	var root *Node
	var orphans []string

	// Phase 2: Build the tree hierarchy by linking children to parents.
	// Iterates all nodes; the one with Type "ROOT" becomes the tree root,
	// all others are appended to their parent's Nodes slice.
	for _, node := range nodes {
		if node.Datatype.Info.Type == "ROOT" {
			root = node
			continue
		}

		// Guard against self-parenting which would create an infinite loop
		// during traversal or serialization.
		if node.Datatype.Content.ParentID == node.Datatype.Content.ContentDataID {
			continue
		}

		parent := nodeIndex[node.Datatype.Content.ParentID]
		if parent == nil {
			orphans = append(orphans, fmt.Sprintf("orphan node %s (parent %s not found)", node.Datatype.Content.ContentDataID, node.Datatype.Content.ParentID))
			continue
		}
		parent.Nodes = append(parent.Nodes, node)
	}

	// Phase 2.5: Reorder children to match stored sibling pointer ordering.
	reorderChildren(nodeIndex)

	// Phase 3: Associate each field with its owning node.
	// Field.Info.ParentID must match a node's Datatype.Content.ContentDataID.
	// For admin fields, ParentID was rewritten in BuildAdminTree to point to
	// the content instance rather than the datatype definition.
	for _, field := range fields {
		node := nodeIndex[field.Info.ParentID]
		if node == nil {
			orphans = append(orphans, fmt.Sprintf("orphan field (parent %s not found)", field.Info.ParentID))
			continue
		}
		node.Fields = append(node.Fields, field)
	}

	var err error
	if len(orphans) > 0 {
		err = fmt.Errorf("BuildNodes: %d orphan(s): %s", len(orphans), strings.Join(orphans, "; "))
		log.Warn("BuildNodes: orphans detected", err)
	}

	return root, err
}

// reorderChildren sorts each node's Nodes slice to match the stored
// sibling-pointer ordering (FirstChildID -> NextSiblingID chain).
// Nodes not found in the chain are appended at the end, preserving
// partial ordering when chains are incomplete.
func reorderChildren(nodeIndex map[string]*Node) {
	for _, node := range nodeIndex {
		if len(node.Nodes) <= 1 {
			continue
		}
		firstChildID := node.Datatype.Content.FirstChildID
		if firstChildID == "" {
			continue
		}
		chain := buildSiblingChain(firstChildID, nodeIndex)
		if chain == nil {
			continue
		}
		node.Nodes = mergeOrdered(chain, node.Nodes)
	}
}

// buildSiblingChain follows the NextSiblingID chain starting from firstChildID
// and returns nodes in correct display order. Returns nil on cycle detection
// or if firstChildID is not found in the index.
func buildSiblingChain(firstChildID string, nodeIndex map[string]*Node) []*Node {
	first := nodeIndex[firstChildID]
	if first == nil {
		return nil
	}

	var chain []*Node
	visited := make(map[string]bool)
	current := first

	for current != nil {
		id := current.Datatype.Content.ContentDataID
		if visited[id] {
			return nil // cycle detected
		}
		visited[id] = true
		chain = append(chain, current)

		nextID := current.Datatype.Content.NextSiblingID
		if nextID == "" {
			break
		}
		current = nodeIndex[nextID]
	}

	return chain
}

// mergeOrdered returns a slice with chain nodes first, followed by any
// nodes in existing that aren't in the chain. Preserves all children
// even when the pointer chain is incomplete.
func mergeOrdered(chain []*Node, existing []*Node) []*Node {
	inChain := make(map[string]bool, len(chain))
	for _, n := range chain {
		inChain[n.Datatype.Content.ContentDataID] = true
	}

	result := make([]*Node, len(chain), len(existing))
	copy(result, chain)

	for _, n := range existing {
		if !inChain[n.Datatype.Content.ContentDataID] {
			result = append(result, n)
		}
	}

	return result
}

