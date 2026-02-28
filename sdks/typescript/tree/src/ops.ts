import type { Tree, TreeNode } from "./types.js";
import { findLastSibling } from "./queries.js";

export function unlink<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
): void {
  const node = tree.nodes.get(nodeId);
  if (!node) return;

  if (node.prevSiblingId) {
    const prev = tree.nodes.get(node.prevSiblingId);
    if (prev) prev.nextSiblingId = node.nextSiblingId;
  }
  if (node.nextSiblingId) {
    const next = tree.nodes.get(node.nextSiblingId);
    if (next) next.prevSiblingId = node.prevSiblingId;
  }

  if (node.parentId) {
    const parent = tree.nodes.get(node.parentId);
    if (parent && parent.firstChildId === nodeId) {
      parent.firstChildId = node.nextSiblingId;
    }
  }

  if (tree.rootId === nodeId) {
    tree.rootId = node.nextSiblingId;
  }

  node.parentId = null;
  node.prevSiblingId = null;
  node.nextSiblingId = null;
}

export function insertBefore<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  targetId: string,
): void {
  const node = tree.nodes.get(nodeId);
  const target = tree.nodes.get(targetId);
  if (!node || !target) return;

  node.parentId = target.parentId;
  node.nextSiblingId = target.id;
  node.prevSiblingId = target.prevSiblingId;

  if (target.prevSiblingId) {
    const prev = tree.nodes.get(target.prevSiblingId);
    if (prev) prev.nextSiblingId = node.id;
  }
  target.prevSiblingId = node.id;

  if (target.parentId) {
    const parent = tree.nodes.get(target.parentId);
    if (parent && parent.firstChildId === targetId) {
      parent.firstChildId = node.id;
    }
  }

  if (tree.rootId === targetId) {
    tree.rootId = node.id;
  }
}

export function insertAfter<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  targetId: string,
): void {
  const node = tree.nodes.get(nodeId);
  const target = tree.nodes.get(targetId);
  if (!node || !target) return;

  node.parentId = target.parentId;
  node.prevSiblingId = target.id;
  node.nextSiblingId = target.nextSiblingId;

  if (target.nextSiblingId) {
    const next = tree.nodes.get(target.nextSiblingId);
    if (next) next.prevSiblingId = node.id;
  }
  target.nextSiblingId = node.id;
}

export function prependChild<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  parentId: string,
): void {
  const node = tree.nodes.get(nodeId);
  const parent = tree.nodes.get(parentId);
  if (!node || !parent) return;

  node.parentId = parent.id;
  node.prevSiblingId = null;
  node.nextSiblingId = parent.firstChildId;

  if (parent.firstChildId) {
    const firstChild = tree.nodes.get(parent.firstChildId);
    if (firstChild) firstChild.prevSiblingId = node.id;
  }
  parent.firstChildId = node.id;
}

export function appendChild<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  parentId: string,
): void {
  const parent = tree.nodes.get(parentId);
  if (!parent) return;

  if (!parent.firstChildId) {
    prependChild(tree, nodeId, parentId);
    return;
  }
  const lastId = findLastSibling(tree, parent.firstChildId);
  insertAfter(tree, nodeId, lastId);
}

export function insertChildAt<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  parentId: string,
  index: number,
): void {
  const parent = tree.nodes.get(parentId);
  if (!parent) return;

  if (index <= 0 || !parent.firstChildId) {
    prependChild(tree, nodeId, parentId);
    return;
  }

  let currentId: string | null = parent.firstChildId;
  let i = 0;
  while (i < index && currentId) {
    const current = tree.nodes.get(currentId);
    if (!current?.nextSiblingId) {
      insertAfter(tree, nodeId, currentId);
      return;
    }
    currentId = current.nextSiblingId;
    i++;
  }

  if (currentId) {
    insertBefore(tree, nodeId, currentId);
  }
}

export function appendSibling<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  targetId: string,
): void {
  insertAfter(tree, nodeId, targetId);
}

export function prependSibling<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  targetId: string,
): void {
  insertBefore(tree, nodeId, targetId);
}

export function remove<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
): string[] {
  const node = tree.nodes.get(nodeId);
  if (!node) return [];

  // Collect descendants while pointers are intact
  const removed = collectDescendantIds(tree, nodeId);
  removed.push(nodeId);

  unlink(tree, nodeId);

  for (const id of removed) {
    tree.nodes.delete(id);
  }

  return removed;
}

function collectDescendantIds<N extends TreeNode>(
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

export function move<N extends TreeNode>(
  tree: Tree<N>,
  nodeId: string,
  targetId: string,
  position: "before" | "after" | "inside",
): void {
  if (nodeId === targetId) return;
  const node = tree.nodes.get(nodeId);
  const target = tree.nodes.get(targetId);
  if (!node || !target) return;

  unlink(tree, nodeId);

  if (position === "before") {
    insertBefore(tree, nodeId, targetId);
  } else if (position === "after") {
    insertAfter(tree, nodeId, targetId);
  } else {
    prependChild(tree, nodeId, targetId);
  }
}
