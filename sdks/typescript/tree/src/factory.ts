import type { Tree, TreeNode } from "./types.js";

export function createTree<N extends TreeNode = TreeNode>(): Tree<N> {
  return {
    nodes: new Map(),
    rootId: null,
  };
}

export function createNode(
  id: string,
  extra?: Partial<TreeNode>,
): TreeNode {
  return {
    id,
    parentId: extra?.parentId ?? null,
    firstChildId: extra?.firstChildId ?? null,
    nextSiblingId: extra?.nextSiblingId ?? null,
    prevSiblingId: extra?.prevSiblingId ?? null,
  };
}

export function addNode<N extends TreeNode>(tree: Tree<N>, node: N): void {
  tree.nodes.set(node.id, node);
}

export function fromFlat<N extends TreeNode>(
  nodes: N[],
  rootId?: string | null,
): Tree<N> {
  const tree: Tree<N> = {
    nodes: new Map(),
    rootId: rootId ?? null,
  };

  for (const node of nodes) {
    tree.nodes.set(node.id, node);
  }

  // Auto-detect rootId: first node with null parentId and null prevSiblingId
  if (rootId === undefined) {
    for (const node of nodes) {
      if (node.parentId === null && node.prevSiblingId === null) {
        tree.rootId = node.id;
        break;
      }
    }
  }

  return tree;
}
