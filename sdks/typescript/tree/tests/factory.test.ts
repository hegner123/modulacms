import { describe, it, expect } from "vitest";
import type { TreeNode } from "../src/types.js";
import { createTree, createNode, addNode, fromFlat } from "../src/factory.js";

describe("createTree", () => {
  it("creates an empty tree", () => {
    const tree = createTree();
    expect(tree.nodes.size).toBe(0);
    expect(tree.rootId).toBeNull();
  });
});

describe("createNode", () => {
  it("creates a node with null pointers", () => {
    const node = createNode("x");
    expect(node).toEqual({
      id: "x",
      parentId: null,
      firstChildId: null,
      nextSiblingId: null,
      prevSiblingId: null,
    });
  });

  it("creates a node with partial overrides", () => {
    const node = createNode("x", { parentId: "p", firstChildId: "c" });
    expect(node.parentId).toBe("p");
    expect(node.firstChildId).toBe("c");
    expect(node.nextSiblingId).toBeNull();
    expect(node.prevSiblingId).toBeNull();
  });
});

describe("addNode", () => {
  it("adds a node to the tree map", () => {
    const tree = createTree();
    const node = createNode("a");
    addNode(tree, node);
    expect(tree.nodes.get("a")).toBe(node);
    expect(tree.nodes.size).toBe(1);
  });

  it("does not set rootId", () => {
    const tree = createTree();
    addNode(tree, createNode("a"));
    expect(tree.rootId).toBeNull();
  });
});

describe("fromFlat", () => {
  it("builds a tree from a flat array with explicit rootId", () => {
    const nodes: TreeNode[] = [
      {
        id: "A",
        parentId: null,
        firstChildId: "B",
        nextSiblingId: null,
        prevSiblingId: null,
      },
      {
        id: "B",
        parentId: "A",
        firstChildId: null,
        nextSiblingId: null,
        prevSiblingId: null,
      },
    ];

    const tree = fromFlat(nodes, "A");
    expect(tree.rootId).toBe("A");
    expect(tree.nodes.size).toBe(2);
    expect(tree.nodes.get("A")!.firstChildId).toBe("B");
  });

  it("auto-detects rootId when not provided", () => {
    const nodes: TreeNode[] = [
      {
        id: "B",
        parentId: "A",
        firstChildId: null,
        nextSiblingId: null,
        prevSiblingId: null,
      },
      {
        id: "A",
        parentId: null,
        firstChildId: "B",
        nextSiblingId: "C",
        prevSiblingId: null,
      },
      {
        id: "C",
        parentId: null,
        firstChildId: null,
        nextSiblingId: null,
        prevSiblingId: "A",
      },
    ];

    const tree = fromFlat(nodes);
    // A has null parentId and null prevSiblingId — it's the root head
    expect(tree.rootId).toBe("A");
  });

  it("respects explicit null rootId (empty root)", () => {
    const nodes: TreeNode[] = [
      {
        id: "A",
        parentId: null,
        firstChildId: null,
        nextSiblingId: null,
        prevSiblingId: null,
      },
    ];

    const tree = fromFlat(nodes, null);
    expect(tree.rootId).toBeNull();
    expect(tree.nodes.size).toBe(1);
  });

  it("builds empty tree from empty array", () => {
    const tree = fromFlat([]);
    expect(tree.rootId).toBeNull();
    expect(tree.nodes.size).toBe(0);
  });

  it("preserves extra fields on extended nodes", () => {
    type LabeledNode = TreeNode & { label: string };
    const nodes: LabeledNode[] = [
      {
        id: "A",
        parentId: null,
        firstChildId: null,
        nextSiblingId: null,
        prevSiblingId: null,
        label: "Home",
      },
    ];

    const tree = fromFlat(nodes);
    const node = tree.nodes.get("A")!;
    expect(node.label).toBe("Home");
  });
});
