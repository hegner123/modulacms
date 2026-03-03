import { describe, it, expect } from "vitest";
import type { Tree } from "../src/types.js";
import {
  getChildren,
  getChildCount,
  getSiblings,
  findLastSibling,
  findChildAt,
  getDepth,
  isDescendant,
  isAncestor,
  getAncestors,
  collectDescendants,
  traverse,
  getRootList,
} from "../src/queries.js";

/**
 * Test tree structure:
 *
 *   Root level: A → B → C
 *   Children of A: D → E
 *   Children of C: F
 *   Children of D: G
 */
function buildTree(): Tree {
  const tree: Tree = { nodes: new Map(), rootId: "A" };

  tree.nodes.set("A", {
    id: "A",
    parentId: null,
    firstChildId: "D",
    nextSiblingId: "B",
    prevSiblingId: null,
  });
  tree.nodes.set("B", {
    id: "B",
    parentId: null,
    firstChildId: null,
    nextSiblingId: "C",
    prevSiblingId: "A",
  });
  tree.nodes.set("C", {
    id: "C",
    parentId: null,
    firstChildId: "F",
    nextSiblingId: null,
    prevSiblingId: "B",
  });
  tree.nodes.set("D", {
    id: "D",
    parentId: "A",
    firstChildId: "G",
    nextSiblingId: "E",
    prevSiblingId: null,
  });
  tree.nodes.set("E", {
    id: "E",
    parentId: "A",
    firstChildId: null,
    nextSiblingId: null,
    prevSiblingId: "D",
  });
  tree.nodes.set("F", {
    id: "F",
    parentId: "C",
    firstChildId: null,
    nextSiblingId: null,
    prevSiblingId: null,
  });
  tree.nodes.set("G", {
    id: "G",
    parentId: "D",
    firstChildId: null,
    nextSiblingId: null,
    prevSiblingId: null,
  });

  return tree;
}

describe("getChildren", () => {
  const tree = buildTree();

  it("returns ordered children of a parent", () => {
    const children = getChildren(tree, "A");
    expect(children.map((c) => c.id)).toEqual(["D", "E"]);
  });

  it("returns empty array for a leaf node", () => {
    expect(getChildren(tree, "G")).toEqual([]);
  });

  it("returns empty array for nonexistent node", () => {
    expect(getChildren(tree, "Z")).toEqual([]);
  });
});

describe("getChildCount", () => {
  const tree = buildTree();

  it("counts children without allocating", () => {
    expect(getChildCount(tree, "A")).toBe(2);
  });

  it("returns 0 for leaf nodes", () => {
    expect(getChildCount(tree, "E")).toBe(0);
  });

  it("returns 1 for single-child parent", () => {
    expect(getChildCount(tree, "C")).toBe(1);
    expect(getChildCount(tree, "D")).toBe(1);
  });
});

describe("getSiblings", () => {
  const tree = buildTree();

  it("returns all root-level siblings from any root node", () => {
    const sibs = getSiblings(tree, "B");
    expect(sibs.map((s) => s.id)).toEqual(["A", "B", "C"]);
  });

  it("returns all siblings from the head node", () => {
    expect(getSiblings(tree, "A").map((s) => s.id)).toEqual(["A", "B", "C"]);
  });

  it("returns siblings of child nodes", () => {
    expect(getSiblings(tree, "E").map((s) => s.id)).toEqual(["D", "E"]);
  });

  it("returns single-element array for only child", () => {
    expect(getSiblings(tree, "G").map((s) => s.id)).toEqual(["G"]);
  });
});

describe("findLastSibling", () => {
  const tree = buildTree();

  it("walks to end of root chain", () => {
    expect(findLastSibling(tree, "A")).toBe("C");
  });

  it("returns self when already last", () => {
    expect(findLastSibling(tree, "C")).toBe("C");
  });

  it("walks to end of child chain", () => {
    expect(findLastSibling(tree, "D")).toBe("E");
  });
});

describe("findChildAt", () => {
  const tree = buildTree();

  it("returns first child at index 0", () => {
    expect(findChildAt(tree, "A", 0)).toBe("D");
  });

  it("returns second child at index 1", () => {
    expect(findChildAt(tree, "A", 1)).toBe("E");
  });

  it("returns null for index beyond child count", () => {
    expect(findChildAt(tree, "A", 5)).toBeNull();
  });

  it("returns null for leaf node", () => {
    expect(findChildAt(tree, "G", 0)).toBeNull();
  });
});

describe("getDepth", () => {
  const tree = buildTree();

  it("returns 0 for root-level nodes", () => {
    expect(getDepth(tree, "A")).toBe(0);
    expect(getDepth(tree, "B")).toBe(0);
    expect(getDepth(tree, "C")).toBe(0);
  });

  it("returns 1 for first-level children", () => {
    expect(getDepth(tree, "D")).toBe(1);
    expect(getDepth(tree, "E")).toBe(1);
    expect(getDepth(tree, "F")).toBe(1);
  });

  it("returns 2 for second-level children", () => {
    expect(getDepth(tree, "G")).toBe(2);
  });
});

describe("isDescendant", () => {
  const tree = buildTree();

  it("returns true for direct child", () => {
    expect(isDescendant(tree, "D", "A")).toBe(true);
  });

  it("returns true for deep descendant", () => {
    expect(isDescendant(tree, "G", "A")).toBe(true);
  });

  it("returns false for unrelated nodes", () => {
    expect(isDescendant(tree, "F", "A")).toBe(false);
  });

  it("returns false for self", () => {
    expect(isDescendant(tree, "A", "A")).toBe(false);
  });

  it("returns false for ancestor-to-descendant direction", () => {
    expect(isDescendant(tree, "A", "G")).toBe(false);
  });
});

describe("isAncestor", () => {
  const tree = buildTree();

  it("returns true when candidate is ancestor", () => {
    expect(isAncestor(tree, "A", "G")).toBe(true);
  });

  it("returns false when candidate is not ancestor", () => {
    expect(isAncestor(tree, "B", "G")).toBe(false);
  });
});

describe("getAncestors", () => {
  const tree = buildTree();

  it("returns parent-to-root path for deep node", () => {
    expect(getAncestors(tree, "G")).toEqual(["D", "A"]);
  });

  it("returns empty array for root-level node", () => {
    expect(getAncestors(tree, "A")).toEqual([]);
  });

  it("returns single parent for depth-1 node", () => {
    expect(getAncestors(tree, "D")).toEqual(["A"]);
  });
});

describe("collectDescendants", () => {
  const tree = buildTree();

  it("collects all descendants of A", () => {
    const desc = collectDescendants(tree, "A");
    expect(desc.sort()).toEqual(["D", "E", "G"]);
  });

  it("collects single child", () => {
    expect(collectDescendants(tree, "D")).toEqual(["G"]);
  });

  it("returns empty for leaf", () => {
    expect(collectDescendants(tree, "G")).toEqual([]);
  });
});

describe("traverse", () => {
  it("returns DFS order from root", () => {
    const tree = buildTree();
    const order = traverse(tree);
    // DFS: A, D, G, E, B, C, F
    expect(order).toEqual(["A", "D", "G", "E", "B", "C", "F"]);
  });

  it("returns empty array for empty tree", () => {
    const tree: Tree = { nodes: new Map(), rootId: null };
    expect(traverse(tree)).toEqual([]);
  });
});

describe("getRootList", () => {
  it("returns all root-level nodes in order", () => {
    const tree = buildTree();
    const roots = getRootList(tree);
    expect(roots.map((r) => r.id)).toEqual(["A", "B", "C"]);
  });

  it("returns empty array for empty tree", () => {
    const tree: Tree = { nodes: new Map(), rootId: null };
    expect(getRootList(tree)).toEqual([]);
  });
});
