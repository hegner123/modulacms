package core

import (
	"fmt"

	"github.com/hegner123/modulacms/internal/db"
	"github.com/hegner123/modulacms/internal/db/types"
)

const maxRetry = 100

// BuildFromRows constructs a tree from TUI/admin query rows.
// Uses a 4-phase algorithm: create nodes, assign hierarchy, resolve orphans,
// reorder by stored sibling pointers.
func BuildFromRows(rows []db.GetContentTreeByRouteRow) (*Root, *LoadStats, error) {
	stats := newLoadStats()
	root := &Root{
		NodeIndex: make(map[types.ContentID]*Node, len(rows)),
	}
	var rootNodes []*Node

	// Phase 1: Create all nodes and populate index
	for i := range rows {
		row := &rows[i]
		cd := db.ContentData{
			ContentDataID: row.ContentDataID,
			ParentID:      row.ParentID,
			FirstChildID:  row.FirstChildID,
			NextSiblingID: row.NextSiblingID,
			PrevSiblingID: row.PrevSiblingID,
			RouteID:       row.RouteID,
			DatatypeID:    row.DatatypeID,
			AuthorID:      row.AuthorID,
			Status:        row.Status,
			DateCreated:   row.DateCreated,
			DateModified:  row.DateModified,
		}
		dt := db.Datatypes{
			DatatypeID: row.DatatypeID.ID,
			Label:      row.DatatypeLabel,
			Type:       row.DatatypeType,
		}
		node := &Node{
			ContentData: &cd,
			Datatype:    dt,
		}
		root.NodeIndex[cd.ContentDataID] = node
		stats.NodesCount++

		if !cd.ParentID.Valid {
			rootNodes = append(rootNodes, node)
			if root.Node == nil {
				root.Node = node
			}
		}
	}

	// Phase 2: Assign hierarchy for nodes with existing parents
	orphans := make(map[types.ContentID]*Node)
	for id, node := range root.NodeIndex {
		if !node.ContentData.ParentID.Valid {
			continue
		}
		parentID := node.ContentData.ParentID.ID
		parent := root.NodeIndex[parentID]
		if parent != nil {
			attachNodeToParent(node, parent)
		} else {
			orphans[id] = node
		}
	}

	// Phase 3: Iteratively resolve orphaned nodes
	for len(orphans) > 0 && stats.RetryAttempts < maxRetry {
		stats.RetryAttempts++
		orphansResolved := 0
		for id, orphan := range orphans {
			parentID := orphan.ContentData.ParentID.ID
			parent := root.NodeIndex[parentID]
			if parent != nil && parent.Parent != nil {
				attachNodeToParent(orphan, parent)
				delete(orphans, id)
				orphansResolved++
				stats.OrphansResolved++
			}
		}
		if orphansResolved == 0 {
			if detectCircularReferences(orphans, root.NodeIndex, stats) {
				break
			}
		}
	}

	// Phase 4: Reorder siblings to match stored pointer order
	reorderByPointers(root.NodeIndex)
	reorderRootSiblings(root, rootNodes)

	// Validate final state
	for id := range orphans {
		stats.FinalOrphans = append(stats.FinalOrphans, id)
	}
	var err error
	if root.Node == nil {
		err = fmt.Errorf("no root node found")
	} else if len(stats.CircularRefs) > 0 {
		err = fmt.Errorf("circular references detected in nodes: %v", stats.CircularRefs)
	} else if len(stats.FinalOrphans) > 0 {
		err = fmt.Errorf("unresolved orphan nodes: %v (parents may not exist)", stats.FinalOrphans)
	}

	return root, stats, err
}

// BuildTree constructs a tree from the full set of content data, datatypes,
// content fields, and field definitions. Used by the API layer where data
// is fetched as parallel slices.
// Contract: cd[i] pairs with dt[i], cf[i] pairs with df[i].
func BuildTree(cd []db.ContentData, dt []db.Datatypes, cf []db.ContentFields, df []db.Fields) (*Root, *LoadStats, error) {
	if len(cd) != len(dt) {
		return nil, nil, fmt.Errorf("BuildTree: content data length (%d) != datatypes length (%d)", len(cd), len(dt))
	}
	if len(cf) != len(df) {
		return nil, nil, fmt.Errorf("BuildTree: content fields length (%d) != fields length (%d)", len(cf), len(df))
	}

	stats := newLoadStats()
	root := &Root{
		NodeIndex: make(map[types.ContentID]*Node, len(cd)),
	}

	// Phase 1: Create nodes from cd/dt pairs
	nodes := make([]*Node, len(cd))
	for i := range cd {
		cdCopy := cd[i]
		node := &Node{
			ContentData: &cdCopy,
			Datatype:    dt[i],
		}
		nodes[i] = node
		root.NodeIndex[cdCopy.ContentDataID] = node
		stats.NodesCount++
	}

	// Phase 2: Identify root and link hierarchy
	var orphanMessages []string
	for _, node := range nodes {
		if types.DatatypeType(node.Datatype.Type).IsRootType() {
			if root.Node == nil {
				root.Node = node
			}
			continue
		}

		// Guard against self-parenting
		if node.ContentData.ParentID.Valid && node.ContentData.ParentID.ID == node.ContentData.ContentDataID {
			continue
		}

		if !node.ContentData.ParentID.Valid {
			// Node with no parent and not a root type -- treat as potential root fallback
			if root.Node == nil {
				root.Node = node
			}
			continue
		}

		parentID := node.ContentData.ParentID.ID
		parent := root.NodeIndex[parentID]
		if parent == nil {
			orphanMessages = append(orphanMessages, fmt.Sprintf("orphan node %s (parent %s not found)", node.ContentData.ContentDataID, parentID))
			continue
		}
		attachNodeToParent(node, parent)
	}

	// Phase 3: Reorder children by sibling pointers
	reorderByPointers(root.NodeIndex)

	// Phase 4: Attach fields to nodes
	for i := range cf {
		if !cf[i].ContentDataID.Valid {
			orphanMessages = append(orphanMessages, fmt.Sprintf("orphan field (no content_data_id)"))
			continue
		}
		node := root.NodeIndex[cf[i].ContentDataID.ID]
		if node == nil {
			orphanMessages = append(orphanMessages, fmt.Sprintf("orphan field (parent %s not found)", cf[i].ContentDataID.ID))
			continue
		}
		node.ContentFields = append(node.ContentFields, cf[i])
		node.Fields = append(node.Fields, df[i])
	}

	var err error
	if root.Node == nil {
		err = fmt.Errorf("no root node found")
	} else if len(orphanMessages) > 0 {
		err = fmt.Errorf("BuildTree: %d orphan(s)", len(orphanMessages))
	}

	return root, stats, err
}

// BuildAdminTree constructs a tree from admin-prefixed DB types.
// Same algorithm as BuildTree, different input types.
// Contract: cd[i] pairs with dt[i], cf[i] pairs with df[i].
func BuildAdminTree(cd []db.AdminContentData, dt []db.AdminDatatypes, cf []db.AdminContentFields, df []db.AdminFields) (*Root, *LoadStats, error) {
	if len(cd) != len(dt) {
		return nil, nil, fmt.Errorf("BuildAdminTree: content data length (%d) != datatypes length (%d)", len(cd), len(dt))
	}
	if len(cf) != len(df) {
		return nil, nil, fmt.Errorf("BuildAdminTree: content fields length (%d) != fields length (%d)", len(cf), len(df))
	}

	// Convert admin types to public types for unified processing.
	// Admin types use different ID families (AdminContentID vs ContentID)
	// so we convert through string representation into ContentID for the
	// shared NodeIndex map.
	stats := newLoadStats()
	root := &Root{
		NodeIndex: make(map[types.ContentID]*Node, len(cd)),
	}

	nodes := make([]*Node, len(cd))
	for i := range cd {
		// Convert AdminContentData to ContentData shape via string IDs
		contentID := types.ContentID(cd[i].AdminContentDataID.String())
		parentID := types.NullableContentID{Valid: cd[i].ParentID.Valid}
		if cd[i].ParentID.Valid {
			parentID.ID = types.ContentID(cd[i].ParentID.ID.String())
		}
		firstChildID := types.NullableContentID{Valid: cd[i].FirstChildID.Valid}
		if cd[i].FirstChildID.Valid {
			firstChildID.ID = types.ContentID(cd[i].FirstChildID.ID.String())
		}
		nextSiblingID := types.NullableContentID{Valid: cd[i].NextSiblingID.Valid}
		if cd[i].NextSiblingID.Valid {
			nextSiblingID.ID = types.ContentID(cd[i].NextSiblingID.ID.String())
		}
		prevSiblingID := types.NullableContentID{Valid: cd[i].PrevSiblingID.Valid}
		if cd[i].PrevSiblingID.Valid {
			prevSiblingID.ID = types.ContentID(cd[i].PrevSiblingID.ID.String())
		}
		routeID := types.NullableRouteID{Valid: cd[i].AdminRouteID.Valid}
		if cd[i].AdminRouteID.Valid {
			routeID.ID = types.RouteID(cd[i].AdminRouteID.ID.String())
		}
		datatypeID := types.NullableDatatypeID{Valid: cd[i].AdminDatatypeID.Valid}
		if cd[i].AdminDatatypeID.Valid {
			datatypeID.ID = types.DatatypeID(cd[i].AdminDatatypeID.ID.String())
		}

		cdConverted := db.ContentData{
			ContentDataID: contentID,
			ParentID:      parentID,
			FirstChildID:  firstChildID,
			NextSiblingID: nextSiblingID,
			PrevSiblingID: prevSiblingID,
			RouteID:       routeID,
			DatatypeID:    datatypeID,
			AuthorID:      cd[i].AuthorID,
			Status:        cd[i].Status,
			DateCreated:   cd[i].DateCreated,
			DateModified:  cd[i].DateModified,
		}

		dtConverted := db.Datatypes{
			DatatypeID:   types.DatatypeID(dt[i].AdminDatatypeID.String()),
			ParentID:     types.NullableDatatypeID{Valid: dt[i].ParentID.Valid},
			Label:        dt[i].Label,
			Type:         dt[i].Type,
			AuthorID:     dt[i].AuthorID,
			DateCreated:  dt[i].DateCreated,
			DateModified: dt[i].DateModified,
		}
		if dt[i].ParentID.Valid {
			dtConverted.ParentID.ID = types.DatatypeID(dt[i].ParentID.ID.String())
		}

		node := &Node{
			ContentData: &cdConverted,
			Datatype:    dtConverted,
		}
		nodes[i] = node
		root.NodeIndex[contentID] = node
		stats.NodesCount++
	}

	// Phase 2: Identify root and link hierarchy
	var orphanMessages []string
	for _, node := range nodes {
		if types.DatatypeType(node.Datatype.Type).IsRootType() {
			if root.Node == nil {
				root.Node = node
			}
			continue
		}

		if node.ContentData.ParentID.Valid && node.ContentData.ParentID.ID == node.ContentData.ContentDataID {
			continue
		}

		if !node.ContentData.ParentID.Valid {
			if root.Node == nil {
				root.Node = node
			}
			continue
		}

		parentID := node.ContentData.ParentID.ID
		parent := root.NodeIndex[parentID]
		if parent == nil {
			orphanMessages = append(orphanMessages, fmt.Sprintf("orphan node %s (parent %s not found)", node.ContentData.ContentDataID, parentID))
			continue
		}
		attachNodeToParent(node, parent)
	}

	// Phase 3: Reorder children by sibling pointers
	reorderByPointers(root.NodeIndex)

	// Phase 4: Attach fields to nodes (convert admin field types)
	for i := range cf {
		if !cf[i].AdminContentDataID.Valid {
			orphanMessages = append(orphanMessages, "orphan admin field (no admin_content_data_id)")
			continue
		}
		contentID := types.ContentID(cf[i].AdminContentDataID.ID.String())
		node := root.NodeIndex[contentID]
		if node == nil {
			orphanMessages = append(orphanMessages, fmt.Sprintf("orphan admin field (parent %s not found)", contentID))
			continue
		}

		// Convert admin content field to public content field shape
		cfConverted := db.ContentFields{
			ContentFieldID: types.ContentFieldID(cf[i].AdminContentFieldID.String()),
			ContentDataID:  types.NullableContentID{ID: contentID, Valid: true},
			FieldValue:     cf[i].AdminFieldValue,
			AuthorID:       cf[i].AuthorID,
			DateCreated:    cf[i].DateCreated,
			DateModified:   cf[i].DateModified,
		}
		if cf[i].AdminFieldID.Valid {
			cfConverted.FieldID = types.NullableFieldID{ID: types.FieldID(cf[i].AdminFieldID.ID.String()), Valid: true}
		}
		if cf[i].AdminRouteID.Valid {
			cfConverted.RouteID = types.NullableRouteID{ID: types.RouteID(cf[i].AdminRouteID.ID.String()), Valid: true}
		}

		// Convert admin field to public field shape
		dfConverted := db.Fields{
			FieldID:      types.FieldID(df[i].AdminFieldID.String()),
			Label:        df[i].Label,
			Data:         df[i].Data,
			ValidationID: types.NullableValidationID{ID: types.ValidationID(df[i].ValidationID.ID.String()), Valid: df[i].ValidationID.Valid},
			UIConfig:     df[i].UIConfig,
			Type:         df[i].Type,
			AuthorID:     df[i].AuthorID,
			DateCreated:  df[i].DateCreated,
			DateModified: df[i].DateModified,
		}
		if df[i].ParentID.Valid {
			dfConverted.ParentID = types.NullableDatatypeID{ID: types.DatatypeID(df[i].ParentID.ID.String()), Valid: true}
		}

		node.ContentFields = append(node.ContentFields, cfConverted)
		node.Fields = append(node.Fields, dfConverted)
	}

	var err error
	if root.Node == nil {
		err = fmt.Errorf("no root node found")
	} else if len(orphanMessages) > 0 {
		err = fmt.Errorf("BuildAdminTree: %d orphan(s)", len(orphanMessages))
	}

	return root, stats, err
}

// attachNodeToParent links a child node to its parent with proper sibling linking.
func attachNodeToParent(node, parent *Node) {
	node.Parent = parent
	if parent.FirstChild == nil {
		parent.FirstChild = node
	} else {
		current := parent.FirstChild
		for current.NextSibling != nil {
			current = current.NextSibling
		}
		current.NextSibling = node
		node.PrevSibling = current
	}
}

// detectCircularReferences checks orphans for circular parent chains.
func detectCircularReferences(orphans map[types.ContentID]*Node, index map[types.ContentID]*Node, stats *LoadStats) bool {
	var circularRefs []types.ContentID
	for id, orphan := range orphans {
		if hasCircularReference(orphan, index, make(map[types.ContentID]bool)) {
			circularRefs = append(circularRefs, id)
		}
	}
	stats.CircularRefs = circularRefs
	return len(circularRefs) > 0
}

// hasCircularReference walks the parent chain checking for cycles.
func hasCircularReference(node *Node, index map[types.ContentID]*Node, visited map[types.ContentID]bool) bool {
	nodeID := node.ContentData.ContentDataID
	if visited[nodeID] {
		return true
	}
	if !node.ContentData.ParentID.Valid {
		return false
	}
	visited[nodeID] = true
	parentID := node.ContentData.ParentID.ID
	parent := index[parentID]
	if parent == nil {
		return false
	}
	return hasCircularReference(parent, index, visited)
}

// reorderByPointers rebuilds sibling chains to match stored FirstChildID/NextSiblingID pointers.
// When FirstChildID is missing, falls back to finding the first child by looking
// for the child whose PrevSiblingID is empty (head of the chain).
func reorderByPointers(index map[types.ContentID]*Node) {
	for _, node := range index {
		if node.FirstChild == nil {
			continue
		}
		if node.ContentData == nil {
			continue
		}

		var firstChildID types.ContentID
		if node.ContentData.FirstChildID.Valid && node.ContentData.FirstChildID.ID != "" {
			firstChildID = node.ContentData.FirstChildID.ID
		} else {
			// Fallback: find the child with no prev sibling (head of chain).
			firstChildID = findChainHead(node)
			if firstChildID == "" {
				continue
			}
		}

		ordered := buildSiblingChain(firstChildID, index)
		if ordered == nil {
			continue
		}
		applySiblingOrder(node, ordered)
	}
}

// findChainHead walks a parent's children to find the one whose
// PrevSiblingID is empty, indicating it is the first in the chain.
func findChainHead(parent *Node) types.ContentID {
	current := parent.FirstChild
	for current != nil {
		if !current.ContentData.PrevSiblingID.Valid || current.ContentData.PrevSiblingID.ID == "" {
			return current.ContentData.ContentDataID
		}
		current = current.NextSibling
	}
	return ""
}

// buildSiblingChain follows the NextSiblingID chain starting from firstChildID.
// Returns nil if the chain is broken by a cycle.
func buildSiblingChain(firstChildID types.ContentID, index map[types.ContentID]*Node) []*Node {
	first := index[firstChildID]
	if first == nil {
		return nil
	}

	var chain []*Node
	visited := make(map[types.ContentID]bool)
	current := first

	for current != nil {
		if visited[current.ContentData.ContentDataID] {
			return nil // cycle detected
		}
		visited[current.ContentData.ContentDataID] = true
		chain = append(chain, current)

		if !current.ContentData.NextSiblingID.Valid || current.ContentData.NextSiblingID.ID == "" {
			break
		}
		nextID := current.ContentData.NextSiblingID.ID
		current = index[nextID]
		if current == nil {
			break // broken chain, keep what we have
		}
	}

	return chain
}

// applySiblingOrder rewrites FirstChild/NextSibling/PrevSibling pointers
// for a parent's children based on the ordered chain.
func applySiblingOrder(parent *Node, ordered []*Node) {
	// Collect all current children
	existing := make(map[types.ContentID]*Node)
	current := parent.FirstChild
	for current != nil {
		existing[current.ContentData.ContentDataID] = current
		current = current.NextSibling
	}

	// Remove chained nodes from existing -- they go first
	for _, n := range ordered {
		delete(existing, n.ContentData.ContentDataID)
	}

	// Append remaining children at the end
	for _, n := range existing {
		ordered = append(ordered, n)
	}

	// Rewrite pointers
	if len(ordered) == 0 {
		return
	}
	parent.FirstChild = ordered[0]
	for i, n := range ordered {
		n.Parent = parent
		if i == 0 {
			n.PrevSibling = nil
		} else {
			n.PrevSibling = ordered[i-1]
		}
		if i == len(ordered)-1 {
			n.NextSibling = nil
		} else {
			n.NextSibling = ordered[i+1]
		}
	}
}

// reorderRootSiblings reorders root-level nodes using their stored sibling pointers.
func reorderRootSiblings(root *Root, rootNodes []*Node) {
	if len(rootNodes) <= 1 {
		return
	}

	// Find the true first root (PrevSiblingID is null/empty)
	var firstRoot *Node
	for _, n := range rootNodes {
		if !n.ContentData.PrevSiblingID.Valid || n.ContentData.PrevSiblingID.ID == "" {
			firstRoot = n
			break
		}
	}

	if firstRoot == nil {
		return
	}

	chain := buildSiblingChain(firstRoot.ContentData.ContentDataID, root.NodeIndex)
	if len(chain) == 0 {
		return
	}

	// Collect root nodes not in chain
	inChain := make(map[types.ContentID]bool)
	for _, n := range chain {
		inChain[n.ContentData.ContentDataID] = true
	}
	for _, n := range rootNodes {
		if !inChain[n.ContentData.ContentDataID] {
			chain = append(chain, n)
		}
	}

	// Rewrite root sibling pointers
	root.Node = chain[0]
	for i, n := range chain {
		n.Parent = nil
		if i == 0 {
			n.PrevSibling = nil
		} else {
			n.PrevSibling = chain[i-1]
		}
		if i == len(chain)-1 {
			n.NextSibling = nil
		} else {
			n.NextSibling = chain[i+1]
		}
	}
}
