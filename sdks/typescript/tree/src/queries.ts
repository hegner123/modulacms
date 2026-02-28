import type { Tree, TreeNode } from "./types.js";

export function getChildren<N extends TreeNode>(
  tree: Tree<N>,
  parentId: string,
): N[] {
  const parent = tree.nodes.get(parentId);
  if (!parent?.firstChildId) return [];

  const children: N[] = [];
  let currentId: string | null = parent.firstChildId;
  while (currentId) {
    const child = tree.nodes.get(currentId);
    if (!child) break;
    children.push(child);
    currentId = child.nextSiblingId;
  }
  return children;
}

export function getChildCount<N extends TreeNode>(
  tree: Tree<N>,
  parentId: string,
): number {
  const parent = tree.nodes.get(parentId);
  if (!parent?.firstChildId) return 0;

  let count = 0;
  let currentId: string | null = parent.firstChildId;
  while (currentId) {
    count++;
    const child = tree.nodes.get(currentId);
    if (!child) break;
    currentId = child.nextSiblingId;
  }
  return count;
}

export function getSiblings<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
): N[] {
  const node = tree.nodes.get(nodeId);
  if (!node) return [];

  // Walk back to head of sibling chain
  let headId: string = nodeId;
  let head: N = node;
  while (head.prevSiblingId) {
    const prev = tree.nodes.get(head.prevSiblingId);
    if (!prev) break;
    headId = head.prevSiblingId;
    head = prev;
  }

  // Walk forward from head
  const siblings: N[] = [];
  let currentId: string | null = headId;
  while (currentId) {
    const current = tree.nodes.get(currentId);
    if (!current) break;
    siblings.push(current);
    currentId = current.nextSiblingId;
  }
  return siblings;
}

export function findLastSibling<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
): string {
  let currentId = nodeId;
  let current = tree.nodes.get(currentId);
  while (current?.nextSiblingId) {
    currentId = current.nextSiblingId;
    current = tree.nodes.get(currentId);
  }
  return currentId;
}

export function findChildAt<N extends TreeNode>(
  tree: Tree<N>,
  parentId: string,
  index: number,
): string | null {
  const parent = tree.nodes.get(parentId);
  if (!parent?.firstChildId) return null;

  let currentId: string | null = parent.firstChildId;
  let i = 0;
  while (currentId && i < index) {
    const current = tree.nodes.get(currentId);
    if (!current) return null;
    currentId = current.nextSiblingId;
    i++;
  }
  return currentId;
}

export function getDepth<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
): number {
  let depth = 0;
  let currentId = tree.nodes.get(nodeId)?.parentId ?? null;
  while (currentId) {
    depth++;
    currentId = tree.nodes.get(currentId)?.parentId ?? null;
  }
  return depth;
}

export function isDescendant<N extends TreeNode>(
  tree: Tree<N>,
  candidateId: string,
  ancestorId: string,
): boolean {
  let currentId = tree.nodes.get(candidateId)?.parentId ?? null;
  while (currentId) {
    if (currentId === ancestorId) return true;
    currentId = tree.nodes.get(currentId)?.parentId ?? null;
  }
  return false;
}

export function isAncestor<N extends TreeNode>(
  tree: Tree<N>,
  candidateId: string,
  descendantId: string,
): boolean {
  return isDescendant(tree, descendantId, candidateId);
}

export function getAncestors<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
): string[] {
  const ancestors: string[] = [];
  let currentId = tree.nodes.get(nodeId)?.parentId ?? null;
  while (currentId) {
    ancestors.push(currentId);
    currentId = tree.nodes.get(currentId)?.parentId ?? null;
  }
  return ancestors;
}

export function collectDescendants<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
): string[] {
  const result: string[] = [];
  const node = tree.nodes.get(nodeId);
  if (!node?.firstChildId) return result;

  const stack = [node.firstChildId];
  while (stack.length > 0) {
    const id = stack.pop()!;
    const child = tree.nodes.get(id);
    if (!child) continue;
    result.push(id);
    if (child.nextSiblingId) stack.push(child.nextSiblingId);
    if (child.firstChildId) stack.push(child.firstChildId);
  }
  return result;
}

export function traverse<N extends TreeNode>(tree: Tree<N>): string[] {
  const order: string[] = [];
  if (!tree.rootId) return order;

  const stack = [tree.rootId];
  while (stack.length > 0) {
    const id = stack.pop()!;
    const node = tree.nodes.get(id);
    if (!node) continue;
    order.push(id);
    // Push nextSibling first so firstChild is processed first (LIFO)
    if (node.nextSiblingId) stack.push(node.nextSiblingId);
    if (node.firstChildId) stack.push(node.firstChildId);
  }
  return order;
}

export function getRootList<N extends TreeNode>(tree: Tree<N>): N[] {
  if (!tree.rootId) return [];

  const roots: N[] = [];
  let currentId: string | null = tree.rootId;
  while (currentId) {
    const node = tree.nodes.get(currentId);
    if (!node) break;
    roots.push(node);
    currentId = node.nextSiblingId;
  }
  return roots;
}
