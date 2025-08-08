package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hegner123/modulacms/internal/db"
)

type EfficientTreeData struct {
	ContentNodes  []db.GetContentTreeByRouteRow
	FieldDefs     []db.GetFieldDefinitionsByRouteRow
	ContentFields []db.GetContentFieldsByRouteRow
}

type HybridTreeBuilder struct {
	nodeMap          map[int64]*TreeNode
	allNodes         []*TreeNode
	childrenByParent map[int64][]*TreeNode
	fieldDefsByType  map[int64][]db.GetFieldDefinitionsByRouteRow
	contentFieldsMap map[int64][]db.GetContentFieldsByRouteRow
}

func LoadedShallowTreeCmd(t *TreeRoot) tea.Cmd {
	return func() tea.Msg {
		return LoadedShallowTreeMsg{
			TreeRoot: t,
		}
	}
}

type LoadedShallowTreeMsg struct {
	TreeRoot *TreeRoot
}

func (m Model) LoadShallowTree() tea.Cmd {
	root := TreeRoot{}
	return LoadedShallowTreeCmd(&root)
}

func (m Model) BuildTree(rows []db.GetRouteTreeByRouteIDRow) tea.Cmd {
	tree := NewTreeRoot()
	return tea.Batch(LoadedShallowTreeCmd(tree))
}

func (m Model) LoadEfficientTree(database *db.Database, routeID int64) tea.Cmd {
	return func() tea.Msg {
		tree, err := BuildEfficientTree(database, routeID)
		if err != nil {
			return DbErrMsg{Error: fmt.Errorf("failed to build tree: %w", err)}
		}
		return LoadedShallowTreeMsg{TreeRoot: tree}
	}
}

func BuildEfficientTree(database *db.Database, routeID int64) (*TreeRoot, error) {

	contentNodes, err := database.GetContentTreeByRoute(routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content tree: %w", err)
	}

	fieldDefs, err := database.GetFieldDefinitionsByRoute(routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get field definitions: %w", err)
	}

	contentFields, err := database.GetContentFieldsByRoute(routeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get content fields: %w", err)
	}

	if len(*contentNodes) == 0 {
		return NewTreeRoot(), nil
	}

	builder := NewHybridTreeBuilder()
	builder.prepareLookupMaps(*fieldDefs, *contentFields)

	var rootNode *TreeNode
	for _, contentNode := range *contentNodes {
		treeNode := builder.createTreeNode(contentNode)

		builder.nodeMap[contentNode.ContentDataID] = treeNode
		builder.allNodes = append(builder.allNodes, treeNode)

		if !contentNode.ParentID.Valid {
			rootNode = treeNode
		} else {
			parentID := contentNode.ParentID.Int64
			builder.childrenByParent[parentID] = append(builder.childrenByParent[parentID], treeNode)
		}
	}

	if rootNode == nil {
		return nil, fmt.Errorf("no root node found")
	}

	return builder.buildTree(rootNode)
}

func NewHybridTreeBuilder() *HybridTreeBuilder {
	return &HybridTreeBuilder{
		nodeMap:          make(map[int64]*TreeNode),
		allNodes:         make([]*TreeNode, 0),
		childrenByParent: make(map[int64][]*TreeNode),
		fieldDefsByType:  make(map[int64][]db.GetFieldDefinitionsByRouteRow),
		contentFieldsMap: make(map[int64][]db.GetContentFieldsByRouteRow),
	}
}

func (b *HybridTreeBuilder) prepareLookupMaps(
	fieldDefs []db.GetFieldDefinitionsByRouteRow,
	contentFields []db.GetContentFieldsByRouteRow,
) {
	for _, fd := range fieldDefs {
		b.fieldDefsByType[fd.DatatypeID] = append(b.fieldDefsByType[fd.DatatypeID], fd)
	}

	for _, cf := range contentFields {
		b.contentFieldsMap[cf.ContentDataID] = append(b.contentFieldsMap[cf.ContentDataID], cf)
	}
}

func (b *HybridTreeBuilder) createTreeNode(contentNode db.GetContentTreeByRouteRow) *TreeNode {
	treeNode := NewTreeNodeFromContentTree(contentNode)

	if defs, exists := b.fieldDefsByType[contentNode.DatatypeID]; exists {
		fieldTypes := make([]db.Fields, 0, len(defs))
		for _, def := range defs {
			field := db.Fields{
				FieldID: def.FieldID,
				Label:   def.Label,
				Type:    def.Type,
			}
			fieldTypes = append(fieldTypes, field)
		}
		treeNode.NodeFieldTypes = fieldTypes
	}

	if fields, exists := b.contentFieldsMap[contentNode.ContentDataID]; exists {
		nodeFields := make([]db.ContentFields, 0, len(fields))
		for _, field := range fields {
			cf := db.ContentFields{
				ContentFieldID: 0,
				ContentDataID:  field.ContentDataID,
				FieldID:        field.FieldID,
				FieldValue:     field.FieldValue,
			}
			nodeFields = append(nodeFields, cf)
		}
		treeNode.NodeFields = nodeFields
	}

	return treeNode
}

func (b *HybridTreeBuilder) buildTree(rootNode *TreeNode) (*TreeRoot, error) {
	tree := NewTreeRoot()
	tree.Root = rootNode
	tree.NodeIndex[rootNode.Node.ContentDataID] = rootNode

	for _, node := range b.allNodes {
		if children, hasChildren := b.childrenByParent[node.Node.ContentDataID]; hasChildren {
			node.Nodes = &children
		}

		if node != rootNode {
			tree.NodeIndex[node.Node.ContentDataID] = node
		}
	}

	return tree, nil
}

func (tree *TreeRoot) GetNodesByType(nodeType string) []*TreeNode {
	var matchingNodes []*TreeNode
	tree.traverseAndCollect(tree.Root, nodeType, &matchingNodes)
	return matchingNodes
}

func (tree *TreeRoot) GetOrderedChildren(parentID int64) []*TreeNode {
	if parentNode, exists := tree.NodeIndex[parentID]; exists && parentNode.Nodes != nil {
		return *parentNode.Nodes
	}
	return nil
}

func (tree *TreeRoot) GetChildrenAtDepth(parentID int64, depth int) []*TreeNode {
	if depth == 0 {
		return tree.GetOrderedChildren(parentID)
	}

	var result []*TreeNode
	children := tree.GetOrderedChildren(parentID)
	for _, child := range children {
		grandchildren := tree.GetChildrenAtDepth(child.Node.ContentDataID, depth-1)
		result = append(result, grandchildren...)
	}
	return result
}

func (tree *TreeRoot) traverseAndCollect(node *TreeNode, nodeType string, collection *[]*TreeNode) {
	if node == nil {
		return
	}

	if node.NodeDatatype.Type == nodeType {
		*collection = append(*collection, node)
	}

	if node.Nodes != nil {
		for _, child := range *node.Nodes {
			tree.traverseAndCollect(child, nodeType, collection)
		}
	}
}

func (tree *TreeRoot) PrintOrderedStructure() string {
	if tree.Root == nil {
		return "Empty tree"
	}
	return tree.printNode(tree.Root, 0)
}

func (tree *TreeRoot) printNode(node *TreeNode, indent int) string {
	result := ""
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	result += fmt.Sprintf("%s├─ %s (ID: %d, Type: %s)\n",
		prefix,
		node.NodeDatatype.Label,
		node.Node.ContentDataID,
		node.NodeDatatype.Type)

	if node.Nodes != nil {
		for _, child := range *node.Nodes {
			result += tree.printNode(child, indent+1)
		}
	}

	return result
}

// SetExpand sets the expand value for a specific node using fast lookup
func (tree *TreeRoot) SetExpand(nodeID int64, expand bool) bool {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		node.Expand = expand
		return true
	}
	return false
}

// SetWrapped sets the wrapped value for a specific node using fast lookup
func (tree *TreeRoot) SetWrapped(nodeID int64, wrapped int) bool {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		node.Wrapped = wrapped
		return true
	}
	return false
}

// SetIndent sets the indent value for a specific node using fast lookup
func (tree *TreeRoot) SetIndent(nodeID int64, indent int) bool {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		node.Indent = indent
		return true
	}
	return false
}

// ExpandNode expands a specific node and optionally its children
func (tree *TreeRoot) ExpandNode(nodeID int64, recursive bool) bool {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		node.Expand = true

		if recursive && node.Nodes != nil {
			for _, child := range *node.Nodes {
				tree.ExpandNode(child.Node.ContentDataID, recursive)
			}
		}
		return true
	}
	return false
}

// CollapseNode collapses a specific node and optionally its children
func (tree *TreeRoot) CollapseNode(nodeID int64, recursive bool) bool {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		node.Expand = false

		if recursive && node.Nodes != nil {
			for _, child := range *node.Nodes {
				tree.CollapseNode(child.Node.ContentDataID, recursive)
			}
		}
		return true
	}
	return false
}

// ToggleExpand toggles the expand state of a specific node
func (tree *TreeRoot) ToggleExpand(nodeID int64) bool {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		node.Expand = !node.Expand
		return true
	}
	return false
}

// SetExpandForType sets expand state for all nodes of a specific type
func (tree *TreeRoot) SetExpandForType(nodeType string, expand bool) {
	tree.traverseAndSetExpand(tree.Root, nodeType, expand)
}

func (tree *TreeRoot) traverseAndSetExpand(node *TreeNode, nodeType string, expand bool) {
	if node == nil {
		return
	}

	if node.NodeDatatype.Type == nodeType {
		node.Expand = expand
	}

	if node.Nodes != nil {
		for _, child := range *node.Nodes {
			tree.traverseAndSetExpand(child, nodeType, expand)
		}
	}
}

// CalculateDepths calculates and sets depth values for all nodes in the tree
func (tree *TreeRoot) CalculateDepths() {
	if tree.Root != nil {
		tree.setNodeDepth(tree.Root, 0)
	}
}

func (tree *TreeRoot) setNodeDepth(node *TreeNode, depth int) {
	node.Indent = depth

	if node.Nodes != nil {
		for _, child := range *node.Nodes {
			tree.setNodeDepth(child, depth+1)
		}
	}
}

// SetDepthsFromParent sets depth values starting from a specific parent node
func (tree *TreeRoot) SetDepthsFromParent(parentID int64, startDepth int) bool {
	if parentNode, exists := tree.NodeIndex[parentID]; exists {
		tree.setNodeDepth(parentNode, startDepth)
		return true
	}
	return false
}

// GetNodeState returns the current state of a node (expand, wrapped, indent)
type NodeState struct {
	NodeID  int64
	Expand  bool
	Wrapped int
	Indent  int
}

func (tree *TreeRoot) GetNodeState(nodeID int64) (*NodeState, bool) {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		return &NodeState{
			NodeID:  nodeID,
			Expand:  node.Expand,
			Wrapped: node.Wrapped,
			Indent:  node.Indent,
		}, true
	}
	return nil, false
}

// SetNodeState applies a complete state to a node
func (tree *TreeRoot) SetNodeState(nodeID int64, state NodeState) bool {
	if node, exists := tree.NodeIndex[nodeID]; exists {
		node.Expand = state.Expand
		node.Wrapped = state.Wrapped
		node.Indent = state.Indent
		return true
	}
	return false
}

// ApplyStates applies multiple node states efficiently
func (tree *TreeRoot) ApplyStates(states []NodeState) {
	for _, state := range states {
		tree.SetNodeState(state.NodeID, state)
	}
}

// GetAllNodeStates returns the current state of all nodes
func (tree *TreeRoot) GetAllNodeStates() []NodeState {
	var states []NodeState

	for nodeID, node := range tree.NodeIndex {
		states = append(states, NodeState{
			NodeID:  nodeID,
			Expand:  node.Expand,
			Wrapped: node.Wrapped,
			Indent:  node.Indent,
		})
	}

	return states
}

// InitializeViewStates sets up initial view states based on content type rules
func (tree *TreeRoot) InitializeViewStates() {
	if tree.Root == nil {
		return
	}

	// Calculate depths first
	tree.CalculateDepths()

	// Set expand states based on content type and depth
	tree.initializeExpandStates(tree.Root)
}

func (tree *TreeRoot) initializeExpandStates(node *TreeNode) {
	if node == nil {
		return
	}

	// Default expand rules based on your testing logic
	switch node.NodeDatatype.Type {
	case "Page":
		node.Expand = true // Always expand pages
	case "Navigation", "Hero", "Footer":
		node.Expand = node.Indent <= 1 // Expand structure elements at shallow depths
	case "Container":
		node.Expand = node.Indent <= 1 // Expand containers for layout
	case "Row":
		node.Expand = node.Indent <= 2 // Expand rows when container accessed
	default:
		node.Expand = false // Collapse content elements by default (lazy load)
	}

	// Recursively initialize children
	if node.Nodes != nil {
		for _, child := range *node.Nodes {
			tree.initializeExpandStates(child)
		}
	}
}
