import { describe, it, expect, beforeEach } from "vitest";
import type { Tree, TreeNode } from "../src/types.js";
import {
  unlink,
  insertBefore,
  insertAfter,
  prependChild,
  appendChild,
  insertChildAt,
  appendSibling,
  prependSibling,
  remove,
  move,
} from "../src/ops.js";

/**
 * Test tree structure:
 *
 *   Root level: A → B → C
 *   Children of A: D → E
 *   Children of C: F
 *   Children of D: G
 *
 * Depths: A=0, B=0, C=0, D=1, E=1, F=1, G=2
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

function detachedNode(id: string): TreeNode {
  return {
    id,
    parentId: null,
    firstChildId: null,
    nextSiblingId: null,
    prevSiblingId: null,
  };
}

describe("unlink", () => {
  let tree: Tree;
  beforeEach(() => {
    tree = buildTree();
  });

  it("removes a middle sibling from the chain", () => {
    unlink(tree, "B");
    const b = tree.nodes.get("B")!;
    expect(b.parentId).toBeNull();
    expect(b.prevSiblingId).toBeNull();
    expect(b.nextSiblingId).toBeNull();

    const a = tree.nodes.get("A")!;
    expect(a.nextSiblingId).toBe("C");
    const c = tree.nodes.get("C")!;
    expect(c.prevSiblingId).toBe("A");
  });

  it("updates rootId when unlinking the root head", () => {
    unlink(tree, "A");
    expect(tree.rootId).toBe("B");
    const b = tree.nodes.get("B")!;
    expect(b.prevSiblingId).toBeNull();
  });

  it("updates parent firstChildId when unlinking first child", () => {
    unlink(tree, "D");
    const a = tree.nodes.get("A")!;
    expect(a.firstChildId).toBe("E");
    const e = tree.nodes.get("E")!;
    expect(e.prevSiblingId).toBeNull();
  });

  it("removes a tail sibling", () => {
    unlink(tree, "C");
    const b = tree.nodes.get("B")!;
    expect(b.nextSiblingId).toBeNull();
  });

  it("is a no-op for nonexistent node", () => {
    unlink(tree, "Z");
    expect(tree.nodes.size).toBe(7);
  });
});

describe("insertBefore", () => {
  let tree: Tree;
  beforeEach(() => {
    tree = buildTree();
  });

  it("inserts a new node before a middle sibling", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertBefore(tree, "X", "B");

    const x = tree.nodes.get("X")!;
    expect(x.parentId).toBeNull();
    expect(x.prevSiblingId).toBe("A");
    expect(x.nextSiblingId).toBe("B");
    expect(tree.nodes.get("A")!.nextSiblingId).toBe("X");
    expect(tree.nodes.get("B")!.prevSiblingId).toBe("X");
  });

  it("inserts before the root head and updates rootId", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertBefore(tree, "X", "A");

    expect(tree.rootId).toBe("X");
    const x = tree.nodes.get("X")!;
    expect(x.prevSiblingId).toBeNull();
    expect(x.nextSiblingId).toBe("A");
  });

  it("inserts before first child and updates parent firstChildId", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertBefore(tree, "X", "D");

    expect(tree.nodes.get("A")!.firstChildId).toBe("X");
    const x = tree.nodes.get("X")!;
    expect(x.parentId).toBe("A");
    expect(x.nextSiblingId).toBe("D");
  });
});

describe("insertAfter", () => {
  let tree: Tree;
  beforeEach(() => {
    tree = buildTree();
  });

  it("inserts a new node after a middle sibling", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertAfter(tree, "X", "A");

    const x = tree.nodes.get("X")!;
    expect(x.parentId).toBeNull();
    expect(x.prevSiblingId).toBe("A");
    expect(x.nextSiblingId).toBe("B");
    expect(tree.nodes.get("A")!.nextSiblingId).toBe("X");
    expect(tree.nodes.get("B")!.prevSiblingId).toBe("X");
  });

  it("inserts after the tail sibling", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertAfter(tree, "X", "C");

    const x = tree.nodes.get("X")!;
    expect(x.prevSiblingId).toBe("C");
    expect(x.nextSiblingId).toBeNull();
    expect(tree.nodes.get("C")!.nextSiblingId).toBe("X");
  });

  it("inherits the parentId from target", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertAfter(tree, "X", "D");

    expect(tree.nodes.get("X")!.parentId).toBe("A");
  });
});

describe("prependChild", () => {
  let tree: Tree;
  beforeEach(() => {
    tree = buildTree();
  });

  it("prepends to a parent with existing children", () => {
    tree.nodes.set("X", detachedNode("X"));
    prependChild(tree, "X", "A");

    expect(tree.nodes.get("A")!.firstChildId).toBe("X");
    const x = tree.nodes.get("X")!;
    expect(x.parentId).toBe("A");
    expect(x.prevSiblingId).toBeNull();
    expect(x.nextSiblingId).toBe("D");
    expect(tree.nodes.get("D")!.prevSiblingId).toBe("X");
  });

  it("prepends to a parent with no children", () => {
    tree.nodes.set("X", detachedNode("X"));
    prependChild(tree, "X", "B");

    expect(tree.nodes.get("B")!.firstChildId).toBe("X");
    const x = tree.nodes.get("X")!;
    expect(x.parentId).toBe("B");
    expect(x.prevSiblingId).toBeNull();
    expect(x.nextSiblingId).toBeNull();
  });
});

describe("appendChild", () => {
  let tree: Tree;
  beforeEach(() => {
    tree = buildTree();
  });

  it("appends after the last child", () => {
    tree.nodes.set("X", detachedNode("X"));
    appendChild(tree, "X", "A");

    const x = tree.nodes.get("X")!;
    expect(x.parentId).toBe("A");
    expect(x.prevSiblingId).toBe("E");
    expect(x.nextSiblingId).toBeNull();
    expect(tree.nodes.get("E")!.nextSiblingId).toBe("X");
  });

  it("appends to an empty parent (becomes first child)", () => {
    tree.nodes.set("X", detachedNode("X"));
    appendChild(tree, "X", "B");

    expect(tree.nodes.get("B")!.firstChildId).toBe("X");
    expect(tree.nodes.get("X")!.parentId).toBe("B");
  });
});

describe("insertChildAt", () => {
  let tree: Tree;
  beforeEach(() => {
    tree = buildTree();
  });

  it("index 0 prepends", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertChildAt(tree, "X", "A", 0);

    expect(tree.nodes.get("A")!.firstChildId).toBe("X");
    expect(tree.nodes.get("X")!.nextSiblingId).toBe("D");
  });

  it("index 1 inserts between first and second child", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertChildAt(tree, "X", "A", 1);

    const x = tree.nodes.get("X")!;
    expect(x.prevSiblingId).toBe("D");
    expect(x.nextSiblingId).toBe("E");
  });

  it("index beyond child count appends", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertChildAt(tree, "X", "A", 99);

    const x = tree.nodes.get("X")!;
    expect(x.prevSiblingId).toBe("E");
    expect(x.nextSiblingId).toBeNull();
  });

  it("inserts into empty parent", () => {
    tree.nodes.set("X", detachedNode("X"));
    insertChildAt(tree, "X", "B", 0);

    expect(tree.nodes.get("B")!.firstChildId).toBe("X");
  });
});

describe("appendSibling / prependSibling", () => {
  it("appendSibling delegates to insertAfter", () => {
    const tree = buildTree();
    tree.nodes.set("X", detachedNode("X"));
    appendSibling(tree, "X", "A");

    expect(tree.nodes.get("X")!.prevSiblingId).toBe("A");
    expect(tree.nodes.get("A")!.nextSiblingId).toBe("X");
  });

  it("prependSibling delegates to insertBefore", () => {
    const tree = buildTree();
    tree.nodes.set("X", detachedNode("X"));
    prependSibling(tree, "X", "B");

    expect(tree.nodes.get("X")!.nextSiblingId).toBe("B");
    expect(tree.nodes.get("B")!.prevSiblingId).toBe("X");
  });
});

describe("remove", () => {
  it("removes a leaf node", () => {
    const tree = buildTree();
    const removed = remove(tree, "G");

    expect(removed).toEqual(["G"]);
    expect(tree.nodes.has("G")).toBe(false);
    expect(tree.nodes.get("D")!.firstChildId).toBeNull();
  });

  it("removes a node and all descendants", () => {
    const tree = buildTree();
    const removed = remove(tree, "D");

    expect(removed).toContain("D");
    expect(removed).toContain("G");
    expect(removed).toHaveLength(2);
    expect(tree.nodes.has("D")).toBe(false);
    expect(tree.nodes.has("G")).toBe(false);
    expect(tree.nodes.get("A")!.firstChildId).toBe("E");
  });

  it("removes a root-level subtree", () => {
    const tree = buildTree();
    const removed = remove(tree, "A");

    expect(removed).toContain("A");
    expect(removed).toContain("D");
    expect(removed).toContain("E");
    expect(removed).toContain("G");
    expect(removed).toHaveLength(4);
    expect(tree.rootId).toBe("B");
  });

  it("returns empty array for nonexistent node", () => {
    const tree = buildTree();
    expect(remove(tree, "Z")).toEqual([]);
  });
});

describe("move", () => {
  it("moves a node before a target", () => {
    const tree = buildTree();
    move(tree, "C", "A", "before");

    expect(tree.rootId).toBe("C");
    expect(tree.nodes.get("C")!.nextSiblingId).toBe("A");
    expect(tree.nodes.get("B")!.prevSiblingId).toBe("A");
  });

  it("moves a node after a target", () => {
    const tree = buildTree();
    move(tree, "A", "C", "after");

    expect(tree.rootId).toBe("B");
    expect(tree.nodes.get("C")!.nextSiblingId).toBe("A");
    expect(tree.nodes.get("A")!.prevSiblingId).toBe("C");
  });

  it("moves a node inside a target (as first child)", () => {
    const tree = buildTree();
    move(tree, "E", "C", "inside");

    expect(tree.nodes.get("E")!.parentId).toBe("C");
    expect(tree.nodes.get("C")!.firstChildId).toBe("E");
    expect(tree.nodes.get("E")!.nextSiblingId).toBe("F");
  });

  it("is a no-op when nodeId equals targetId", () => {
    const tree = buildTree();
    move(tree, "A", "A", "before");
    expect(tree.rootId).toBe("A");
  });
});
