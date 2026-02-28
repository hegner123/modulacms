export type TreeNode = {
  id: string
  parentId: string | null
  firstChildId: string | null
  nextSiblingId: string | null
  prevSiblingId: string | null
}

export type Tree<N extends TreeNode = TreeNode> = {
  nodes: Map<string, N>
  rootId: string | null
}
