// AUTO-GENERATED — DO NOT EDIT. Source: block-editor-src/. Regenerate: just admin bundle

// internal/admin/static/js/block-editor-src/config.js
var BLOCK_TYPE_CONFIG = {
  text: { label: "Text", canHaveChildren: false },
  heading: { label: "Heading", canHaveChildren: false },
  image: { label: "Image", canHaveChildren: false },
  container: { label: "Container", canHaveChildren: true }
};
function getTypeConfig(type) {
  if (BLOCK_TYPE_CONFIG[type]) return BLOCK_TYPE_CONFIG[type];
  return { label: type, canHaveChildren: true };
}
var MAX_DEPTH = 8;

// internal/admin/static/js/block-editor-src/id.js
function generateId() {
  return crypto.randomUUID();
}

// internal/admin/static/js/block-editor-src/state.js
function createState() {
  return {
    blocks: {},
    rootId: null,
    selectedBlockId: null,
    dirty: false
  };
}

// internal/admin/static/js/block-editor-src/tree-queries.js
function getChildren(state, parentId) {
  const parent = state.blocks[parentId];
  if (!parent || !parent.firstChildId) return [];
  const children = [];
  let currentId = parent.firstChildId;
  while (currentId) {
    const child = state.blocks[currentId];
    children.push(child);
    currentId = child.nextSiblingId;
  }
  return children;
}
function getRootList(state) {
  if (!state.rootId) return [];
  const roots = [];
  let currentId = state.rootId;
  while (currentId) {
    const block = state.blocks[currentId];
    roots.push(block);
    currentId = block.nextSiblingId;
  }
  return roots;
}
function isDescendant(state, candidateId, ancestorId) {
  let currentId = state.blocks[candidateId]?.parentId;
  while (currentId) {
    if (currentId === ancestorId) return true;
    currentId = state.blocks[currentId]?.parentId;
  }
  return false;
}
function getDepth(state, blockId) {
  let depth = 0;
  let currentId = state.blocks[blockId]?.parentId;
  while (currentId) {
    depth++;
    currentId = state.blocks[currentId]?.parentId;
  }
  return depth;
}
function findLastSibling(state, blockId) {
  let currentId = blockId;
  while (state.blocks[currentId]?.nextSiblingId) {
    currentId = state.blocks[currentId].nextSiblingId;
  }
  return currentId;
}
function getBlockTraversalOrder(state) {
  const order = [];
  if (!state.rootId) return order;
  const stack = [state.rootId];
  while (stack.length > 0) {
    const id = stack.pop();
    const block = state.blocks[id];
    if (!block) continue;
    order.push(id);
    if (block.nextSiblingId) stack.push(block.nextSiblingId);
    if (block.firstChildId) stack.push(block.firstChildId);
  }
  return order;
}
function collectDescendants(state, blockId) {
  const result = [];
  const block = state.blocks[blockId];
  if (!block || !block.firstChildId) return result;
  const stack = [block.firstChildId];
  while (stack.length > 0) {
    const id = stack.pop();
    const child = state.blocks[id];
    if (!child) continue;
    result.push(id);
    if (child.nextSiblingId) stack.push(child.nextSiblingId);
    if (child.firstChildId) stack.push(child.firstChildId);
  }
  return result;
}
function getDescendantCount(state, blockId) {
  return collectDescendants(state, blockId).length;
}

// internal/admin/static/js/block-editor-src/tree-ops.js
function unlink(state, blockId) {
  const block = state.blocks[blockId];
  if (!block) return;
  if (block.prevSiblingId) {
    state.blocks[block.prevSiblingId].nextSiblingId = block.nextSiblingId;
  }
  if (block.nextSiblingId) {
    state.blocks[block.nextSiblingId].prevSiblingId = block.prevSiblingId;
  }
  if (block.parentId) {
    const parent = state.blocks[block.parentId];
    if (parent && parent.firstChildId === blockId) {
      parent.firstChildId = block.nextSiblingId;
    }
  }
  if (state.rootId === blockId) {
    state.rootId = block.nextSiblingId;
  }
  block.parentId = null;
  block.prevSiblingId = null;
  block.nextSiblingId = null;
}
function insertBefore(state, blockId, targetId) {
  const block = state.blocks[blockId];
  const target = state.blocks[targetId];
  block.parentId = target.parentId;
  block.nextSiblingId = target.id;
  block.prevSiblingId = target.prevSiblingId;
  if (target.prevSiblingId) {
    state.blocks[target.prevSiblingId].nextSiblingId = block.id;
  }
  target.prevSiblingId = block.id;
  if (target.parentId) {
    const parent = state.blocks[target.parentId];
    if (parent && parent.firstChildId === targetId) {
      parent.firstChildId = block.id;
    }
  }
  if (state.rootId === targetId) {
    state.rootId = block.id;
  }
}
function insertAfter(state, blockId, targetId) {
  const block = state.blocks[blockId];
  const target = state.blocks[targetId];
  block.parentId = target.parentId;
  block.prevSiblingId = target.id;
  block.nextSiblingId = target.nextSiblingId;
  if (target.nextSiblingId) {
    state.blocks[target.nextSiblingId].prevSiblingId = block.id;
  }
  target.nextSiblingId = block.id;
}
function insertAsFirstChild(state, blockId, parentId) {
  const block = state.blocks[blockId];
  const parent = state.blocks[parentId];
  block.parentId = parent.id;
  block.prevSiblingId = null;
  block.nextSiblingId = parent.firstChildId;
  if (parent.firstChildId) {
    state.blocks[parent.firstChildId].prevSiblingId = block.id;
  }
  parent.firstChildId = block.id;
}
function insertAsLastChild(state, blockId, parentId) {
  const parent = state.blocks[parentId];
  if (!parent.firstChildId) {
    insertAsFirstChild(state, blockId, parentId);
    return;
  }
  const lastId = findLastSibling(state, parent.firstChildId);
  insertAfter(state, blockId, lastId);
}
function addBlock(state, type, afterId) {
  const id = generateId();
  const config = getTypeConfig(type);
  const block = {
    id,
    type,
    parentId: null,
    firstChildId: null,
    prevSiblingId: null,
    nextSiblingId: null,
    label: config.label + " Block"
  };
  state.blocks[id] = block;
  if (!state.rootId) {
    state.rootId = id;
  } else if (afterId) {
    insertAfter(state, id, afterId);
  } else {
    const lastId = findLastSibling(state, state.rootId);
    insertAfter(state, id, lastId);
  }
  state.dirty = true;
  return id;
}
function addBlockFromDatatype(state, datatype, position, targetId) {
  var id = generateId();
  var block = {
    id,
    type: datatype.type,
    parentId: null,
    firstChildId: null,
    prevSiblingId: null,
    nextSiblingId: null,
    datatypeId: datatype.id,
    label: datatype.label,
    authorId: "",
    routeId: "",
    status: "draft",
    dateCreated: (/* @__PURE__ */ new Date()).toISOString(),
    dateModified: (/* @__PURE__ */ new Date()).toISOString(),
    fields: []
  };
  state.blocks[id] = block;
  if (!state.rootId) {
    state.rootId = id;
  } else if (position === "before" && targetId) {
    insertBefore(state, id, targetId);
  } else if (position === "after" && targetId) {
    insertAfter(state, id, targetId);
  } else if (position === "inside" && targetId) {
    insertAsFirstChild(state, id, targetId);
  } else {
    var lastId = findLastSibling(state, state.rootId);
    insertAfter(state, id, lastId);
  }
  state.dirty = true;
  return id;
}
function removeBlock(state, blockId) {
  const block = state.blocks[blockId];
  if (!block) return [];
  const removed = collectDescendants(state, blockId);
  removed.push(blockId);
  unlink(state, blockId);
  for (const id of removed) {
    delete state.blocks[id];
  }
  state.dirty = true;
  return removed;
}
function moveBlock(state, blockId, targetId, position) {
  if (blockId === targetId) return;
  const block = state.blocks[blockId];
  const target = state.blocks[targetId];
  if (!block || !target) return;
  unlink(state, blockId);
  if (position === "before") {
    insertBefore(state, blockId, targetId);
  } else if (position === "after") {
    insertAfter(state, blockId, targetId);
  } else if (position === "inside") {
    insertAsFirstChild(state, blockId, targetId);
  }
  state.dirty = true;
}
function indentBlock(state, blockId) {
  const block = state.blocks[blockId];
  if (!block) return false;
  if (!block.prevSiblingId) return false;
  const prevSibling = state.blocks[block.prevSiblingId];
  if (!prevSibling) return false;
  if (getDepth(state, blockId) + 1 >= MAX_DEPTH) return false;
  const config = getTypeConfig(prevSibling.type);
  if (!config.canHaveChildren) return false;
  unlink(state, blockId);
  insertAsLastChild(state, blockId, prevSibling.id);
  state.dirty = true;
  return true;
}
function outdentBlock(state, blockId) {
  const block = state.blocks[blockId];
  if (!block) return false;
  if (!block.parentId) return false;
  const parent = state.blocks[block.parentId];
  if (!parent) return false;
  const parentId = parent.id;
  const youngerSiblings = [];
  let walkId = block.nextSiblingId;
  while (walkId) {
    youngerSiblings.push(walkId);
    walkId = state.blocks[walkId].nextSiblingId;
  }
  if (block.prevSiblingId) {
    state.blocks[block.prevSiblingId].nextSiblingId = null;
  } else {
    parent.firstChildId = null;
  }
  if (youngerSiblings.length > 0) {
    block.nextSiblingId = null;
    state.blocks[youngerSiblings[0]].prevSiblingId = null;
  }
  unlink(state, blockId);
  if (youngerSiblings.length > 0) {
    for (const sibId of youngerSiblings) {
      state.blocks[sibId].parentId = blockId;
    }
    if (block.firstChildId) {
      const lastChildId = findLastSibling(state, block.firstChildId);
      state.blocks[lastChildId].nextSiblingId = youngerSiblings[0];
      state.blocks[youngerSiblings[0]].prevSiblingId = lastChildId;
    } else {
      block.firstChildId = youngerSiblings[0];
    }
  }
  insertAfter(state, blockId, parentId);
  state.dirty = true;
  return true;
}
function duplicateBlock(state, blockId) {
  const block = state.blocks[blockId];
  if (!block) return null;
  const idMap = /* @__PURE__ */ new Map();
  function cloneSubtree(originalId) {
    const original = state.blocks[originalId];
    if (!original) return null;
    const newId = generateId();
    idMap.set(originalId, newId);
    const clone = {
      id: newId,
      type: original.type,
      parentId: null,
      firstChildId: null,
      prevSiblingId: null,
      nextSiblingId: null,
      label: original.label
    };
    state.blocks[newId] = clone;
    let childId = original.firstChildId;
    let prevCloneChildId = null;
    while (childId) {
      const cloneChildId = cloneSubtree(childId);
      state.blocks[cloneChildId].parentId = newId;
      if (!prevCloneChildId) {
        clone.firstChildId = cloneChildId;
      } else {
        state.blocks[prevCloneChildId].nextSiblingId = cloneChildId;
        state.blocks[cloneChildId].prevSiblingId = prevCloneChildId;
      }
      prevCloneChildId = cloneChildId;
      childId = state.blocks[childId].nextSiblingId;
    }
    return newId;
  }
  const cloneId = cloneSubtree(blockId);
  insertAfter(state, cloneId, blockId);
  state.dirty = true;
  return cloneId;
}
function moveBlockUp(state, blockId) {
  const block = state.blocks[blockId];
  if (!block) return false;
  const prevSiblingId = block.prevSiblingId;
  if (!prevSiblingId) return false;
  unlink(state, blockId);
  insertBefore(state, blockId, prevSiblingId);
  state.dirty = true;
  return true;
}
function moveBlockDown(state, blockId) {
  const block = state.blocks[blockId];
  if (!block) return false;
  const nextSiblingId = block.nextSiblingId;
  if (!nextSiblingId) return false;
  unlink(state, blockId);
  insertAfter(state, blockId, nextSiblingId);
  state.dirty = true;
  return true;
}

// internal/admin/static/js/block-editor-src/validate.js
function validateState(state) {
  const errors = [];
  const reachable = /* @__PURE__ */ new Set();
  for (const [id, block] of Object.entries(state.blocks)) {
    if (block.nextSiblingId) {
      const next = state.blocks[block.nextSiblingId];
      if (!next) {
        errors.push(`Block ${id}: nextSiblingId "${block.nextSiblingId}" not found in blocks`);
      } else if (next.prevSiblingId !== id) {
        errors.push(`Block ${id}: nextSiblingId "${block.nextSiblingId}" has prevSiblingId "${next.prevSiblingId}", expected "${id}"`);
      }
    }
    if (block.prevSiblingId) {
      const prev = state.blocks[block.prevSiblingId];
      if (!prev) {
        errors.push(`Block ${id}: prevSiblingId "${block.prevSiblingId}" not found in blocks`);
      } else if (prev.nextSiblingId !== id) {
        errors.push(`Block ${id}: prevSiblingId "${block.prevSiblingId}" has nextSiblingId "${prev.nextSiblingId}", expected "${id}"`);
      }
    }
  }
  for (const [id, block] of Object.entries(state.blocks)) {
    if (block.firstChildId) {
      const firstChild = state.blocks[block.firstChildId];
      if (!firstChild) {
        errors.push(`Block ${id}: firstChildId "${block.firstChildId}" not found in blocks`);
      } else if (firstChild.parentId !== id) {
        errors.push(`Block ${id}: firstChildId "${block.firstChildId}" has parentId "${firstChild.parentId}", expected "${id}"`);
      }
    }
  }
  if (state.rootId) {
    const visited = /* @__PURE__ */ new Set();
    let currentId = state.rootId;
    while (currentId) {
      if (visited.has(currentId)) {
        errors.push(`Cycle detected in root chain at block "${currentId}"`);
        break;
      }
      visited.add(currentId);
      const block = state.blocks[currentId];
      if (!block) {
        errors.push(`Root chain references non-existent block "${currentId}"`);
        break;
      }
      if (block.parentId) {
        errors.push(`Root chain block "${currentId}" has parentId "${block.parentId}", expected null or empty`);
      }
      reachable.add(currentId);
      markChildrenReachable(state, currentId, reachable, errors);
      currentId = block.nextSiblingId;
    }
  }
  for (const id of Object.keys(state.blocks)) {
    if (!reachable.has(id)) {
      errors.push(`Block "${id}" is not reachable from rootId or any firstChildId chain`);
    }
  }
  for (const [id, block] of Object.entries(state.blocks)) {
    if (block.parentId) {
      const parent = state.blocks[block.parentId];
      if (!parent) {
        errors.push(`Block "${id}" has parentId "${block.parentId}" which does not exist`);
      }
    }
  }
  for (const [id, block] of Object.entries(state.blocks)) {
    if (block.firstChildId) {
      const visited = /* @__PURE__ */ new Set();
      let childId = block.firstChildId;
      while (childId) {
        if (visited.has(childId)) {
          errors.push(`Cycle in child chain of block "${id}" at "${childId}"`);
          break;
        }
        visited.add(childId);
        const child = state.blocks[childId];
        if (!child) break;
        if (child.parentId !== id) {
          errors.push(`Child "${childId}" of block "${id}" has parentId "${child.parentId}", expected "${id}"`);
        }
        childId = child.nextSiblingId;
      }
    }
  }
  return errors;
}
function markChildrenReachable(state, parentId, reachable, errors) {
  const parent = state.blocks[parentId];
  if (!parent || !parent.firstChildId) return;
  const visited = /* @__PURE__ */ new Set();
  let childId = parent.firstChildId;
  while (childId) {
    if (visited.has(childId)) break;
    visited.add(childId);
    reachable.add(childId);
    const child = state.blocks[childId];
    if (!child) break;
    markChildrenReachable(state, childId, reachable, errors);
    childId = child.nextSiblingId;
  }
}

// internal/admin/static/js/block-editor-src/dom-patches.js
var domPatchMethods = {
  /**
   * Indent a block: make it the last child of its previous sibling.
   * Handles both state mutation and DOM patching.
   */
  _doIndentBlock(blockId) {
    if (!this._state) return;
    const block = this._state.blocks[blockId];
    if (!block) return;
    const oldParentId = block.parentId;
    const newParentId = block.prevSiblingId;
    this._history.pushUndo(this._state);
    const result = indentBlock(this._state, blockId);
    if (!result) {
      this._history.discardLastUndo();
      return;
    }
    this._devValidate();
    const blockWrapper = this._wrapperRegistry.get(blockId);
    const newParentWrapper = this._wrapperRegistry.get(newParentId);
    if (blockWrapper && newParentWrapper) {
      let childContainer = newParentWrapper.querySelector(":scope > .children-container");
      if (!childContainer) {
        childContainer = document.createElement("div");
        childContainer.className = "children-container";
        childContainer.dataset.parentId = newParentId;
        newParentWrapper.appendChild(childContainer);
      }
      childContainer.appendChild(blockWrapper);
      const newDepth = getDepth(this._state, blockId);
      blockWrapper.style.marginInlineStart = newDepth * 24 + "px";
      this._updateDescendantDepths(blockId);
    }
    this._cleanupEmptyChildrenContainer(oldParentId);
    this._updateChildCountBadge(oldParentId);
    this._updateChildCountBadge(newParentId);
    this._updateSaveButton();
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "indent", blockId }
    }));
  },
  /**
   * Outdent a block: move it out of its parent to become the next sibling
   * of that parent. Younger siblings are reparented under the block.
   * Handles both state mutation and DOM patching.
   */
  _doOutdentBlock(blockId) {
    if (!this._state) return;
    const block = this._state.blocks[blockId];
    if (!block) return;
    const oldParentId = block.parentId;
    if (!oldParentId) return;
    const youngerSiblingIds = [];
    let walkId = block.nextSiblingId;
    while (walkId) {
      youngerSiblingIds.push(walkId);
      walkId = this._state.blocks[walkId].nextSiblingId;
    }
    this._history.pushUndo(this._state);
    const result = outdentBlock(this._state, blockId);
    if (!result) {
      this._history.discardLastUndo();
      return;
    }
    this._devValidate();
    const blockWrapper = this._wrapperRegistry.get(blockId);
    const oldParentWrapper = this._wrapperRegistry.get(oldParentId);
    if (blockWrapper && oldParentWrapper) {
      const nextSibling = oldParentWrapper.nextElementSibling;
      if (nextSibling) {
        oldParentWrapper.parentNode.insertBefore(blockWrapper, nextSibling);
      } else {
        oldParentWrapper.parentNode.appendChild(blockWrapper);
      }
      const newDepth = getDepth(this._state, blockId);
      blockWrapper.style.marginInlineStart = newDepth * 24 + "px";
      if (youngerSiblingIds.length > 0) {
        let childContainer = blockWrapper.querySelector(":scope > .children-container");
        if (!childContainer) {
          childContainer = document.createElement("div");
          childContainer.className = "children-container";
          childContainer.dataset.parentId = blockId;
          blockWrapper.appendChild(childContainer);
        }
        for (const sibId of youngerSiblingIds) {
          const sibWrapper = this._wrapperRegistry.get(sibId);
          if (sibWrapper) {
            childContainer.appendChild(sibWrapper);
          }
        }
      }
      this._updateDescendantDepths(blockId);
    }
    this._cleanupEmptyChildrenContainer(oldParentId);
    this._updateChildCountBadge(oldParentId);
    this._updateChildCountBadge(blockId);
    this._updateSaveButton();
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "outdent", blockId }
    }));
  },
  /**
   * Duplicate a block and its entire subtree. Renders the cloned subtree
   * and inserts it into the DOM after the original's wrapper.
   */
  _doDuplicateBlock(blockId) {
    if (!this._state) return;
    if (!blockId) return;
    const block = this._state.blocks[blockId];
    if (!block) return;
    this._history.pushUndo(this._state);
    const cloneId = duplicateBlock(this._state, blockId);
    if (!cloneId) return;
    this._devValidate();
    const originalWrapper = this._wrapperRegistry.get(blockId);
    if (originalWrapper) {
      const cloneBlock = this._state.blocks[cloneId];
      const depth = getDepth(this._state, cloneId);
      const cloneWrapper = this._renderBlockWrapper(cloneBlock, depth);
      const nextSibling = originalWrapper.nextElementSibling;
      if (nextSibling) {
        originalWrapper.parentNode.insertBefore(cloneWrapper, nextSibling);
      } else {
        originalWrapper.parentNode.appendChild(cloneWrapper);
      }
    }
    const parentId = this._state.blocks[cloneId]?.parentId;
    this._updateChildCountBadge(parentId);
    this._updateSaveButton();
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "duplicate", blockId }
    }));
  },
  /**
   * Move a block before its previous sibling.
   * Handles state mutation and surgical DOM patching.
   */
  _doMoveBlockUp(blockId) {
    if (!this._state) return;
    if (!blockId) return;
    const block = this._state.blocks[blockId];
    if (!block) return;
    const prevSiblingId = block.prevSiblingId;
    if (!prevSiblingId) return;
    this._history.pushUndo(this._state);
    const result = moveBlockUp(this._state, blockId);
    if (!result) {
      this._history.discardLastUndo();
      return;
    }
    this._devValidate();
    const blockWrapper = this._wrapperRegistry.get(blockId);
    const prevWrapper = this._wrapperRegistry.get(prevSiblingId);
    if (blockWrapper && prevWrapper) {
      prevWrapper.parentNode.insertBefore(blockWrapper, prevWrapper);
    }
    this._updateSaveButton();
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "moveUp", blockId }
    }));
  },
  /**
   * Move a block after its next sibling.
   * Handles state mutation and surgical DOM patching.
   */
  _doMoveBlockDown(blockId) {
    if (!this._state) return;
    if (!blockId) return;
    const block = this._state.blocks[blockId];
    if (!block) return;
    const nextSiblingId = block.nextSiblingId;
    if (!nextSiblingId) return;
    this._history.pushUndo(this._state);
    const result = moveBlockDown(this._state, blockId);
    if (!result) {
      this._history.discardLastUndo();
      return;
    }
    this._devValidate();
    const blockWrapper = this._wrapperRegistry.get(blockId);
    const nextWrapper = this._wrapperRegistry.get(nextSiblingId);
    if (blockWrapper && nextWrapper) {
      const afterNext = nextWrapper.nextElementSibling;
      if (afterNext) {
        nextWrapper.parentNode.insertBefore(blockWrapper, afterNext);
      } else {
        nextWrapper.parentNode.appendChild(blockWrapper);
      }
    }
    this._updateSaveButton();
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "moveDown", blockId }
    }));
  },
  /**
   * Add a new block after the specified block (or at end of root list).
   * Used by Enter key and toolbar. Default type is 'text'.
   */
  _doAddBlockAfter(afterBlockId, type) {
    if (!this._state) return;
    if (!afterBlockId) return;
    const afterBlock = this._state.blocks[afterBlockId];
    if (!afterBlock) return;
    this._history.pushUndo(this._state);
    const id = addBlock(this._state, type || "text", afterBlockId);
    this._devValidate();
    const block = this._state.blocks[id];
    const depth = getDepth(this._state, id);
    const wrapper = this._renderBlockWrapper(block, depth);
    const afterWrapper = this._wrapperRegistry.get(afterBlockId);
    if (afterWrapper) {
      const nextSibling = afterWrapper.nextElementSibling;
      if (nextSibling) {
        afterWrapper.parentNode.insertBefore(wrapper, nextSibling);
      } else {
        afterWrapper.parentNode.appendChild(wrapper);
      }
    }
    this._selectBlock(id);
    const parentId = this._state.blocks[id]?.parentId;
    this._updateChildCountBadge(parentId);
    this._updateSaveButton();
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "add", blockId: id }
    }));
  },
  /**
   * Recursively update marginInlineStart on descendant wrappers to match their new depth.
   */
  _updateDescendantDepths(parentBlockId) {
    const children = getChildren(this._state, parentBlockId);
    for (const child of children) {
      const childWrapper = this._wrapperRegistry.get(child.id);
      if (childWrapper) {
        const depth = getDepth(this._state, child.id);
        childWrapper.style.marginInlineStart = depth * 24 + "px";
      }
      this._updateDescendantDepths(child.id);
    }
  },
  /**
   * Remove an empty children container div from a parent's wrapper.
   * This prevents visual artifacts (empty indented space) and incorrect
   * elementFromPoint hit testing (empty container intercepting pointer events).
   */
  _cleanupEmptyChildrenContainer(parentId) {
    if (!parentId) return;
    const parentBlock = this._state?.blocks[parentId];
    if (!parentBlock) return;
    if (parentBlock.firstChildId !== null) return;
    const parentWrapper = this._wrapperRegistry.get(parentId);
    if (!parentWrapper) return;
    const childContainer = parentWrapper.querySelector(":scope > .children-container");
    if (childContainer) {
      childContainer.remove();
    }
  },
  /**
   * Update or remove the child count badge on a block's header element.
   */
  _updateChildCountBadge(blockId) {
    if (!blockId) return;
    const block = this._state?.blocks[blockId];
    if (!block) return;
    const headerEl = this._elementRegistry.get(blockId);
    if (!headerEl) return;
    const existingBadge = headerEl.querySelector(".child-count-badge");
    const childCount = getDescendantCount(this._state, blockId);
    if (childCount > 0) {
      if (existingBadge) {
        existingBadge.textContent = String(childCount);
        existingBadge.title = childCount + " descendant" + (childCount === 1 ? "" : "s");
      } else {
        const countBadge = document.createElement("span");
        countBadge.className = "child-count-badge";
        countBadge.textContent = String(childCount);
        countBadge.title = childCount + " descendant" + (childCount === 1 ? "" : "s");
        const deleteBtn = headerEl.querySelector(".block-delete-btn");
        if (deleteBtn) {
          headerEl.insertBefore(countBadge, deleteBtn);
        } else {
          headerEl.appendChild(countBadge);
        }
      }
    } else {
      if (existingBadge) {
        existingBadge.remove();
      }
    }
  }
};

// internal/admin/static/js/block-editor-src/drag.js
var dragMethods = {
  _onPointerDown(e) {
    if (e.button !== 0) return;
    if (e.target.closest("[data-action]")) return;
    const blockItem = e.target.closest(".block-item");
    if (!blockItem) return;
    const blockId = blockItem.dataset.blockId;
    if (!blockId) return;
    if (e.target.closest(".block-content-preview")) {
      this._selectBlock(blockId);
      return;
    }
    const block = this._state?.blocks[blockId];
    if (!block) return;
    this._pointerSelectActive = true;
    const startX = e.clientX;
    const startY = e.clientY;
    const onPreMove = (moveEvent) => {
      try {
        this._onPreThresholdMove(moveEvent, blockItem, blockId, startX, startY, onPreMove, onPreUp);
      } catch (err) {
        console.error("[block-editor] Error in pre-threshold move:", err);
        blockItem.removeEventListener("pointermove", onPreMove);
        blockItem.removeEventListener("pointerup", onPreUp);
      }
    };
    const onPreUp = () => {
      blockItem.removeEventListener("pointermove", onPreMove);
      blockItem.removeEventListener("pointerup", onPreUp);
      this._pointerSelectActive = false;
      this._selectBlock(blockId);
    };
    blockItem.addEventListener("pointermove", onPreMove);
    blockItem.addEventListener("pointerup", onPreUp);
  },
  _onPreThresholdMove(e, blockItem, blockId, startX, startY, preMoveFn, preUpFn) {
    const dx = e.clientX - startX;
    const dy = e.clientY - startY;
    const distance = Math.sqrt(dx * dx + dy * dy);
    if (distance < 8) return;
    blockItem.removeEventListener("pointermove", preMoveFn);
    blockItem.removeEventListener("pointerup", preUpFn);
    this._startDrag(e, blockItem, blockId, startX, startY);
  },
  _startDrag(e, blockItem, blockId, startX, startY) {
    try {
      blockItem.setPointerCapture(e.pointerId);
      const blockRect = blockItem.getBoundingClientRect();
      const grabOffsetX = startX - blockRect.left;
      const grabOffsetY = startY - blockRect.top;
      const overlay = this._createDragOverlay(blockItem, e.clientX, e.clientY, grabOffsetX, grabOffsetY);
      const onDragMove = (moveEvent) => {
        try {
          this._onDragMove(moveEvent);
        } catch (err) {
          console.error("[block-editor] Error in drag move:", err);
          this._cleanupDrag();
        }
      };
      const onDragEnd = (upEvent) => {
        try {
          this._onDragEnd(upEvent);
        } catch (err) {
          console.error("[block-editor] Error in drag end:", err);
          this._cleanupDrag();
        }
      };
      this._drag = {
        blockId,
        blockItem,
        overlay,
        grabOffsetX,
        grabOffsetY,
        pointerId: e.pointerId,
        onDragMove,
        onDragEnd,
        dropTarget: null
        // { blockId, position: 'before'|'after'|'inside' }
      };
      blockItem.classList.add("dragging");
      blockItem.addEventListener("pointermove", onDragMove);
      blockItem.addEventListener("pointerup", onDragEnd);
      blockItem.addEventListener("pointercancel", onDragEnd);
    } catch (err) {
      console.error("[block-editor] Error starting drag:", err);
      this._cleanupDrag();
    }
  },
  _createDragOverlay(blockItem, clientX, clientY, grabOffsetX, grabOffsetY) {
    const overlay = blockItem.cloneNode(true);
    overlay.className = "block-item drag-overlay";
    overlay.style.width = blockItem.getBoundingClientRect().width + "px";
    overlay.style.transform = "translate(" + (clientX - grabOffsetX) + "px, " + (clientY - grabOffsetY) + "px)";
    this.appendChild(overlay);
    return overlay;
  },
  _onDragMove(e) {
    if (!this._drag) return;
    const { overlay, grabOffsetX, grabOffsetY } = this._drag;
    overlay.style.transform = "translate(" + (e.clientX - grabOffsetX) + "px, " + (e.clientY - grabOffsetY) + "px)";
    this._lastPointerY = e.clientY;
    this._startAutoScroll();
    const dropTarget = this._computeDropZone(e.clientX, e.clientY);
    this._drag.dropTarget = dropTarget;
    if (dropTarget) {
      this._updateDropIndicator(dropTarget);
    } else {
      this._removeDropIndicator();
      this._removeDropInsideHighlight();
    }
  },
  /**
   * Three-zone drop detection against block header rects (not full subtree).
   * Works on ALL blocks (root and nested).
   * beforeZone = Math.max(rect.height * 0.25, 10)
   * afterZone = Math.max(rect.height * 0.25, 10)
   * relativeY in beforeZone -> "before", in afterZone -> "after", else -> "inside"
   *
   * Coercions:
   * - "inside" coerced to "after" if canHaveChildren is false
   * - "inside" coerced to "after" if target depth + 1 > MAX_DEPTH
   * - Skip entirely if target is descendant of dragged block
   */
  _computeDropZone(clientX, clientY) {
    if (!this._drag) return null;
    const allBlockItems = this.querySelectorAll(".block-item");
    let closestTarget = null;
    let closestDistance = Infinity;
    for (const item of allBlockItems) {
      const itemBlockId = item.dataset.blockId;
      if (!itemBlockId) continue;
      if (itemBlockId === this._drag.blockId) continue;
      if (item.classList.contains("drag-overlay")) continue;
      if (isDescendant(this._state, itemBlockId, this._drag.blockId)) continue;
      const rect = item.getBoundingClientRect();
      if (clientY >= rect.top && clientY <= rect.bottom) {
        const beforeZone = Math.max(rect.height * 0.25, 10);
        const afterZone = Math.max(rect.height * 0.25, 10);
        const relativeY = clientY - rect.top;
        let position;
        if (relativeY < beforeZone) {
          position = "before";
        } else if (relativeY > rect.height - afterZone) {
          position = "after";
        } else {
          position = "inside";
        }
        if (position === "inside") {
          const targetBlock = this._state.blocks[itemBlockId];
          const config = getTypeConfig(targetBlock?.type);
          if (!config.canHaveChildren) {
            position = "after";
          }
        }
        if (position === "inside") {
          const targetDepth = getDepth(this._state, itemBlockId);
          if (targetDepth + 1 > MAX_DEPTH) {
            position = "after";
          }
        }
        return { blockId: itemBlockId, position };
      }
      const distToTop = Math.abs(clientY - rect.top);
      const distToBottom = Math.abs(clientY - rect.bottom);
      const minDist = Math.min(distToTop, distToBottom);
      if (minDist < closestDistance) {
        closestDistance = minDist;
        closestTarget = { item, rect, distToTop, distToBottom };
      }
    }
    if (closestTarget) {
      const { item, rect } = closestTarget;
      const itemBlockId = item.dataset.blockId;
      if (clientY < rect.top) {
        return { blockId: itemBlockId, position: "before" };
      }
      return { blockId: itemBlockId, position: "after" };
    }
    return null;
  },
  _updateDropIndicator(dropTarget) {
    const targetEl = this._elementRegistry.get(dropTarget.blockId);
    if (!targetEl) return;
    if (dropTarget.position === "inside") {
      this._removeDropIndicator();
      this._applyDropInsideHighlight(dropTarget.blockId);
      return;
    }
    this._removeDropInsideHighlight();
    const blockList = this.querySelector(".block-list");
    if (!blockList) return;
    if (!this._dropIndicator) {
      this._dropIndicator = document.createElement("div");
      this._dropIndicator.className = "drop-indicator";
      blockList.appendChild(this._dropIndicator);
    }
    const targetRect = targetEl.getBoundingClientRect();
    const listRect = blockList.getBoundingClientRect();
    if (dropTarget.position === "before") {
      this._dropIndicator.style.top = targetRect.top - listRect.top - 1 + "px";
    } else {
      this._dropIndicator.style.top = targetRect.bottom - listRect.top - 1 + "px";
    }
  },
  _applyDropInsideHighlight(blockId) {
    this._removeDropInsideHighlight();
    const targetEl = this._elementRegistry.get(blockId);
    if (targetEl) {
      targetEl.classList.add("drop-inside");
    }
  },
  _removeDropInsideHighlight() {
    const highlighted = this.querySelector(".drop-inside");
    if (highlighted) {
      highlighted.classList.remove("drop-inside");
    }
  },
  _removeDropIndicator() {
    if (this._dropIndicator) {
      this._dropIndicator.remove();
      this._dropIndicator = null;
    }
  },
  _onDragEnd(e) {
    if (!this._drag) return;
    const { dropTarget } = this._drag;
    if (dropTarget) {
      this._executeDrop(dropTarget);
    }
    this._cleanupDrag();
  },
  /**
   * Execute a drop: mutate state, patch DOM, handle nesting.
   * For "inside" drops: move block's wrapper into target's children container.
   * For "before"/"after": move block's wrapper next to target's wrapper.
   */
  _executeDrop(dropTarget) {
    if (!this._drag || !this._state) return;
    const { blockId } = this._drag;
    const { blockId: targetId, position } = dropTarget;
    const oldParentId = this._state.blocks[blockId]?.parentId;
    this._history.pushUndo(this._state);
    moveBlock(this._state, blockId, targetId, position);
    this._devValidate();
    const blockWrapper = this._wrapperRegistry.get(blockId);
    const targetWrapper = this._wrapperRegistry.get(targetId);
    if (blockWrapper && targetWrapper) {
      if (position === "inside") {
        let childContainer = targetWrapper.querySelector(":scope > .children-container");
        if (!childContainer) {
          childContainer = document.createElement("div");
          childContainer.className = "children-container";
          childContainer.dataset.parentId = targetId;
          targetWrapper.appendChild(childContainer);
        }
        const newDepth = getDepth(this._state, blockId);
        blockWrapper.style.marginInlineStart = newDepth * 24 + "px";
        this._updateDescendantDepths(blockId);
        if (childContainer.firstElementChild) {
          childContainer.insertBefore(blockWrapper, childContainer.firstElementChild);
        } else {
          childContainer.appendChild(blockWrapper);
        }
      } else if (position === "before") {
        targetWrapper.parentNode.insertBefore(blockWrapper, targetWrapper);
        const newDepth = getDepth(this._state, blockId);
        blockWrapper.style.marginInlineStart = newDepth * 24 + "px";
        this._updateDescendantDepths(blockId);
      } else {
        const nextSibling = targetWrapper.nextElementSibling;
        if (nextSibling) {
          targetWrapper.parentNode.insertBefore(blockWrapper, nextSibling);
        } else {
          targetWrapper.parentNode.appendChild(blockWrapper);
        }
        const newDepth = getDepth(this._state, blockId);
        blockWrapper.style.marginInlineStart = newDepth * 24 + "px";
        this._updateDescendantDepths(blockId);
      }
    }
    this._cleanupEmptyChildrenContainer(oldParentId);
    this._updateChildCountBadge(oldParentId);
    this._updateChildCountBadge(targetId);
    this._updateChildCountBadge(blockId);
    this._updateSaveButton();
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "move", blockId, targetId, position }
    }));
  },
  // ---- Auto-scroll ----
  _startAutoScroll() {
    if (this._autoScrollRaf !== null) return;
    const editorContainer = this.querySelector("[data-editor-container]");
    if (!editorContainer) return;
    const EDGE_ZONE = 40;
    const MAX_SPEED = 12;
    const scrollStep = () => {
      if (!this._drag) {
        this._stopAutoScroll();
        return;
      }
      const containerRect = editorContainer.getBoundingClientRect();
      const pointerY = this._lastPointerY;
      const distFromTop = pointerY - containerRect.top;
      const distFromBottom = containerRect.bottom - pointerY;
      if (distFromTop < EDGE_ZONE && distFromTop >= 0) {
        const speed = Math.round(MAX_SPEED * (1 - distFromTop / EDGE_ZONE));
        editorContainer.scrollTop = editorContainer.scrollTop - speed;
      } else if (distFromBottom < EDGE_ZONE && distFromBottom >= 0) {
        const speed = Math.round(MAX_SPEED * (1 - distFromBottom / EDGE_ZONE));
        editorContainer.scrollTop = editorContainer.scrollTop + speed;
      } else {
        this._stopAutoScroll();
        return;
      }
      this._autoScrollRaf = requestAnimationFrame(scrollStep);
    };
    this._autoScrollRaf = requestAnimationFrame(scrollStep);
  },
  _stopAutoScroll() {
    if (this._autoScrollRaf !== null) {
      cancelAnimationFrame(this._autoScrollRaf);
      this._autoScrollRaf = null;
    }
  },
  _onEscapeKey(e) {
    if (e.key === "Escape" && this._drag) {
      this._cleanupDrag();
    }
  },
  _cleanupDrag() {
    if (!this._drag) return;
    const { blockItem, overlay, pointerId, onDragMove, onDragEnd } = this._drag;
    this._stopAutoScroll();
    try {
      if (blockItem.hasPointerCapture(pointerId)) {
        blockItem.releasePointerCapture(pointerId);
      }
    } catch (captureErr) {
    }
    blockItem.removeEventListener("pointermove", onDragMove);
    blockItem.removeEventListener("pointerup", onDragEnd);
    blockItem.removeEventListener("pointercancel", onDragEnd);
    if (overlay && overlay.parentNode) {
      overlay.remove();
    }
    blockItem.classList.remove("dragging");
    this._removeDropIndicator();
    this._removeDropInsideHighlight();
    this._pointerSelectActive = false;
    this._drag = null;
  }
};

// internal/admin/static/js/block-editor-src/cache.js
var _dtCache = {
  data: null,
  // Array of {id, parentId, name, label, type} from API
  fetchedAt: 0,
  // timestamp ms
  ttl: 5 * 60 * 1e3,
  // 5 minutes
  pending: null
  // in-flight promise to deduplicate concurrent fetches
};
var SYSTEM_TYPES = { "_root": true, "_nested_root": true, "_system_log": true, "_reference": true };
function fetchDatatypes() {
  var now = Date.now();
  if (_dtCache.data && now - _dtCache.fetchedAt < _dtCache.ttl) {
    return Promise.resolve(_dtCache.data);
  }
  if (_dtCache.pending) return _dtCache.pending;
  _dtCache.pending = fetch("/admin/api/datatypes", { credentials: "same-origin" }).then(function(res) {
    if (!res.ok) throw new Error("Failed to fetch datatypes: " + res.status);
    return res.json();
  }).then(function(datatypes) {
    _dtCache.data = datatypes.map(function(dt) {
      return { id: dt.datatype_id, parentId: dt.parent_id || null, name: dt.name, label: dt.label, type: dt.type };
    });
    _dtCache.fetchedAt = Date.now();
    _dtCache.pending = null;
    return _dtCache.data;
  }).catch(function(err) {
    _dtCache.pending = null;
    throw err;
  });
  return _dtCache.pending;
}
function fetchDatatypesGrouped(rootDatatypeId) {
  return fetchDatatypes().then(function(datatypes) {
    var childrenOf = {};
    var byId = {};
    for (var i = 0; i < datatypes.length; i++) {
      var dt = datatypes[i];
      byId[dt.id] = dt;
      var pid = dt.parentId || "_none";
      if (!childrenOf[pid]) childrenOf[pid] = [];
      childrenOf[pid].push(dt);
    }
    function collectChildren(parentId, baseDepth) {
      var result = [];
      var kids = childrenOf[parentId];
      if (!kids) return result;
      for (var j = 0; j < kids.length; j++) {
        var kid = kids[j];
        if (SYSTEM_TYPES[kid.type]) continue;
        result.push({ id: kid.id, name: kid.name, label: kid.label, type: kid.type, depth: baseDepth });
        var grandchildren = collectChildren(kid.id, baseDepth + 1);
        for (var k = 0; k < grandchildren.length; k++) {
          result.push(grandchildren[k]);
        }
      }
      return result;
    }
    var categories = [];
    if (rootDatatypeId && byId[rootDatatypeId]) {
      var rootDt = byId[rootDatatypeId];
      var rootItems = collectChildren(rootDatatypeId, 0);
      if (rootItems.length > 0) {
        categories.push({ name: rootDt.label, items: rootItems });
      }
    }
    var collectionItems = [];
    for (var ci = 0; ci < datatypes.length; ci++) {
      if (datatypes[ci].type === "_collection") {
        var kids2 = collectChildren(datatypes[ci].id, 0);
        for (var ck = 0; ck < kids2.length; ck++) {
          collectionItems.push(kids2[ck]);
        }
      }
    }
    if (collectionItems.length > 0) {
      categories.push({ name: "Collections", items: collectionItems });
    }
    var globalItems = [];
    for (var gi = 0; gi < datatypes.length; gi++) {
      var gdt = datatypes[gi];
      if (gdt.type === "_global") {
        globalItems.push({ id: gdt.id, name: gdt.name, label: gdt.label, type: gdt.type, depth: 0 });
        var gkids = collectChildren(gdt.id, 1);
        for (var gk = 0; gk < gkids.length; gk++) {
          globalItems.push(gkids[gk]);
        }
      }
    }
    if (globalItems.length > 0) {
      categories.push({ name: "Global", items: globalItems });
    }
    return { categories };
  });
}

// internal/admin/static/js/block-editor-src/picker.js
var pickerMethods = {
  _openPicker: function(insertTargetId, position) {
    this._pickerOpen = true;
    this._pickerInsertTarget = insertTargetId;
    this._pickerInsertPosition = position || "after";
    this._pickerQuery = "";
    this._pickerSelectedIndex = 0;
    var self = this;
    fetchDatatypesGrouped(this._rootDatatypeId).then(function(grouped) {
      self._pickerData = grouped;
      self._renderPicker();
    }).catch(function(err) {
      console.error("[block-editor] Failed to load datatypes for picker:", err);
      self._closePicker();
    });
  },
  _closePicker: function() {
    this._pickerOpen = false;
    if (this._pickerBackdrop) {
      this._pickerBackdrop.remove();
      this._pickerBackdrop = null;
    }
    this._pickerEl = null;
    if (this._pickerEscHandler) {
      document.removeEventListener("keydown", this._pickerEscHandler, true);
      this._pickerEscHandler = null;
    }
    var container = this.querySelector("[data-editor-container]");
    if (container) container.focus();
  },
  _renderPicker: function() {
    if (this._pickerBackdrop) {
      this._pickerBackdrop.remove();
    }
    var backdrop = document.createElement("div");
    backdrop.className = "block-picker-backdrop";
    var picker = document.createElement("div");
    picker.className = "block-picker";
    var results = document.createElement("div");
    results.className = "block-picker-results";
    picker.appendChild(results);
    var inputBar = document.createElement("div");
    inputBar.className = "block-picker-input";
    var prompt = document.createElement("span");
    prompt.className = "block-picker-prompt";
    prompt.textContent = ">";
    inputBar.appendChild(prompt);
    var queryDisplay = document.createElement("span");
    queryDisplay.className = "block-picker-query";
    queryDisplay.textContent = this._pickerQuery;
    inputBar.appendChild(queryDisplay);
    picker.appendChild(inputBar);
    backdrop.appendChild(picker);
    this._pickerEl = picker;
    this._pickerBackdrop = backdrop;
    var self = this;
    backdrop.addEventListener("mousedown", function(e) {
      if (e.target === backdrop) {
        self._closePicker();
      }
    });
    document.body.appendChild(backdrop);
    this._renderPickerResults();
    this._pickerEscHandler = function(e) {
      if (!self._pickerOpen) return;
      self._onPickerKeyDown(e);
    };
    document.addEventListener("keydown", this._pickerEscHandler, true);
  },
  /**
   * Build the flat list of selectable items from picker data,
   * filtered by the current query. Returns an array of
   * { id, label, type, depth, categoryName } objects.
   */
  _getPickerItems: function() {
    if (!this._pickerData) return [];
    var categories = this._pickerData.categories;
    var query = this._pickerQuery.toLowerCase();
    var items = [];
    for (var c = 0; c < categories.length; c++) {
      var cat = categories[c];
      var catItems = [];
      for (var i = 0; i < cat.items.length; i++) {
        var item = cat.items[i];
        if (query && item.label.toLowerCase().indexOf(query) === -1 && (!item.name || item.name.toLowerCase().indexOf(query) === -1)) continue;
        catItems.push(item);
      }
      if (catItems.length === 0) continue;
      items.push({ isHeader: true, name: cat.name });
      for (var j = 0; j < catItems.length; j++) {
        items.push(catItems[j]);
      }
    }
    return items;
  },
  _renderPickerResults: function() {
    var resultsEl = this._pickerEl.querySelector(".block-picker-results");
    if (!resultsEl) return;
    resultsEl.innerHTML = "";
    var items = this._getPickerItems();
    var selectableIndex = 0;
    for (var i = 0; i < items.length; i++) {
      var item = items[i];
      if (item.isHeader) {
        var header = document.createElement("div");
        header.className = "block-picker-header";
        header.textContent = item.name;
        resultsEl.appendChild(header);
        continue;
      }
      var row = document.createElement("div");
      row.className = "block-picker-item";
      row.dataset.selectableIndex = String(selectableIndex);
      row.dataset.datatypeId = item.id;
      row.dataset.datatypeLabel = item.label;
      row.dataset.datatypeType = item.type;
      if (item.depth > 0) {
        row.style.paddingLeft = 12 + item.depth * 16 + "px";
      }
      if (selectableIndex === this._pickerSelectedIndex) {
        row.classList.add("block-picker-item--selected");
      }
      row.textContent = item.label;
      var self = this;
      row.addEventListener("mousedown", /* @__PURE__ */ (function(idx) {
        return function(e) {
          e.preventDefault();
          e.stopPropagation();
          self._pickerSelectedIndex = idx;
          self._pickerInsertBlock();
        };
      })(selectableIndex));
      resultsEl.appendChild(row);
      selectableIndex++;
    }
    var selected = resultsEl.querySelector(".block-picker-item--selected");
    if (selected) {
      selected.scrollIntoView({ block: "nearest" });
    }
    var queryEl = this._pickerEl.querySelector(".block-picker-query");
    if (queryEl) {
      queryEl.textContent = this._pickerQuery;
    }
  },
  /**
   * Get the list of selectable (non-header) items from picker data.
   */
  _getSelectableItems: function() {
    if (!this._pickerData) return [];
    var categories = this._pickerData.categories;
    var query = this._pickerQuery.toLowerCase();
    var items = [];
    for (var c = 0; c < categories.length; c++) {
      var cat = categories[c];
      for (var i = 0; i < cat.items.length; i++) {
        var item = cat.items[i];
        if (query && item.label.toLowerCase().indexOf(query) === -1 && (!item.name || item.name.toLowerCase().indexOf(query) === -1)) continue;
        items.push(item);
      }
    }
    return items;
  },
  _onPickerKeyDown: function(e) {
    e.preventDefault();
    e.stopPropagation();
    if (e.key === "Escape") {
      this._closePicker();
      return;
    }
    var selectableItems = this._getSelectableItems();
    var maxIndex = selectableItems.length - 1;
    if (e.key === "ArrowUp") {
      this._pickerSelectedIndex = Math.max(0, this._pickerSelectedIndex - 1);
      this._renderPickerResults();
      return;
    }
    if (e.key === "ArrowDown") {
      this._pickerSelectedIndex = Math.min(maxIndex, this._pickerSelectedIndex + 1);
      this._renderPickerResults();
      return;
    }
    if (e.key === "Enter") {
      this._pickerInsertBlock();
      return;
    }
    if (e.key === "Backspace") {
      if (this._pickerQuery.length > 0) {
        this._pickerQuery = this._pickerQuery.slice(0, -1);
        this._pickerSelectedIndex = 0;
        this._renderPickerResults();
      }
      return;
    }
    if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
      this._pickerQuery += e.key;
      this._pickerSelectedIndex = 0;
      this._renderPickerResults();
      return;
    }
  },
  _pickerInsertBlock: function() {
    var selectableItems = this._getSelectableItems();
    if (selectableItems.length === 0) return;
    var idx = this._pickerSelectedIndex;
    if (idx < 0 || idx >= selectableItems.length) return;
    var item = selectableItems[idx];
    var datatype = { id: item.id, label: item.label, type: item.type };
    this._history.pushUndo(this._state);
    var id = addBlockFromDatatype(
      this._state,
      datatype,
      this._pickerInsertPosition,
      this._pickerInsertTarget
    );
    this._closePicker();
    this._devValidate();
    this._render();
    this._selectBlock(id);
    this.dispatchEvent(new CustomEvent("block-editor:change", {
      bubbles: true,
      composed: true,
      detail: { action: "add", blockId: id }
    }));
  }
};

// internal/admin/static/js/block-editor-src/history.js
var History = class {
  constructor(maxSize) {
    this._maxSize = maxSize || 50;
    this._undoStack = [];
    this._redoStack = [];
    this._fieldBatchTimer = null;
    this._inFieldBatch = false;
  }
  _cloneState(state) {
    var blocks = {};
    var keys = Object.keys(state.blocks);
    for (var i = 0; i < keys.length; i++) {
      var key = keys[i];
      var block = state.blocks[key];
      blocks[key] = {
        ...block,
        fields: block.fields ? block.fields.map(function(f) {
          return { ...f };
        }) : void 0
      };
    }
    return { blocks, rootId: state.rootId };
  }
  pushUndo(state) {
    clearTimeout(this._fieldBatchTimer);
    this._inFieldBatch = false;
    var entry = {
      snapshot: this._cloneState(state),
      selectedBlockId: state.selectedBlockId
    };
    this._undoStack.push(entry);
    this._redoStack = [];
    if (this._undoStack.length > this._maxSize) {
      this._undoStack.shift();
    }
  }
  discardLastUndo() {
    this._undoStack.pop();
  }
  pushFieldChange(state) {
    if (this._inFieldBatch) return;
    this.pushUndo(state);
    this._inFieldBatch = true;
    var self = this;
    this._fieldBatchTimer = setTimeout(function() {
      self._inFieldBatch = false;
    }, 500);
  }
  get inFieldBatch() {
    return this._inFieldBatch;
  }
  popUndo(currentState) {
    if (this._undoStack.length === 0) return null;
    var redoEntry = {
      snapshot: this._cloneState(currentState),
      selectedBlockId: currentState.selectedBlockId
    };
    this._redoStack.push(redoEntry);
    return this._undoStack.pop();
  }
  popRedo(currentState) {
    if (this._redoStack.length === 0) return null;
    var undoEntry = {
      snapshot: this._cloneState(currentState),
      selectedBlockId: currentState.selectedBlockId
    };
    this._undoStack.push(undoEntry);
    return this._redoStack.pop();
  }
  get canUndo() {
    return this._undoStack.length > 0;
  }
  get canRedo() {
    return this._redoStack.length > 0;
  }
  clear() {
    this._undoStack = [];
    this._redoStack = [];
    clearTimeout(this._fieldBatchTimer);
    this._inFieldBatch = false;
  }
  remapIds(idMap) {
    var clientIds = Object.keys(idMap);
    var stacks = [this._undoStack, this._redoStack];
    for (var s = 0; s < stacks.length; s++) {
      var stack = stacks[s];
      for (var e = 0; e < stack.length; e++) {
        var entry = stack[e];
        var snapshot = entry.snapshot;
        for (var c = 0; c < clientIds.length; c++) {
          var clientId = clientIds[c];
          if (snapshot.blocks[clientId] !== void 0) {
            var block = snapshot.blocks[clientId];
            delete snapshot.blocks[clientId];
            block.id = idMap[clientId];
            snapshot.blocks[idMap[clientId]] = block;
          }
        }
        var blockKeys = Object.keys(snapshot.blocks);
        for (var b = 0; b < blockKeys.length; b++) {
          var blk = snapshot.blocks[blockKeys[b]];
          if (blk.parentId && idMap[blk.parentId] !== void 0) {
            blk.parentId = idMap[blk.parentId];
          }
          if (blk.firstChildId && idMap[blk.firstChildId] !== void 0) {
            blk.firstChildId = idMap[blk.firstChildId];
          }
          if (blk.nextSiblingId && idMap[blk.nextSiblingId] !== void 0) {
            blk.nextSiblingId = idMap[blk.nextSiblingId];
          }
          if (blk.prevSiblingId && idMap[blk.prevSiblingId] !== void 0) {
            blk.prevSiblingId = idMap[blk.prevSiblingId];
          }
        }
        if (snapshot.rootId && idMap[snapshot.rootId] !== void 0) {
          snapshot.rootId = idMap[snapshot.rootId];
        }
        if (entry.selectedBlockId && idMap[entry.selectedBlockId] !== void 0) {
          entry.selectedBlockId = idMap[entry.selectedBlockId];
        }
      }
    }
  }
};

// internal/admin/static/js/block-editor-src/index.js
var isBrowser = typeof window !== "undefined";
if (isBrowser) {
  class BlockEditor extends HTMLElement {
    static get observedAttributes() {
      return ["data-state"];
    }
    constructor() {
      super();
      this._state = null;
      this._elementRegistry = /* @__PURE__ */ new Map();
      this._wrapperRegistry = /* @__PURE__ */ new Map();
      this._collapsedBlocks = /* @__PURE__ */ new Set();
      this._beforeUnloadHandler = this._onBeforeUnload.bind(this);
      this._drag = null;
      this._dropIndicator = null;
      this._escapeHandler = this._onEscapeKey.bind(this);
      this._autoScrollRaf = null;
      this._lastPointerY = 0;
      this._keydownHandler = this._onKeyDown.bind(this);
      this._pointerSelectActive = false;
      this._pickerOpen = false;
      this._pickerEl = null;
      this._pickerBackdrop = null;
      this._pickerInsertTarget = null;
      this._pickerInsertPosition = "after";
      this._pickerQuery = "";
      this._pickerSelectedIndex = 0;
      this._pickerData = null;
      this._rootDatatypeId = null;
      this._history = new History(50);
    }
    get dev() {
      return this.hasAttribute("data-dev");
    }
    get state() {
      return this._state;
    }
    set state(newState) {
      this._state = newState;
      if (this._history) this._history.clear();
      this._state.dirty = false;
      this._elementRegistry.clear();
      this._wrapperRegistry.clear();
      this._collapsedBlocks.clear();
      this._render();
    }
    get dirty() {
      return this._state ? this._state.dirty : false;
    }
    serialize() {
      if (!this._state) return "{}";
      return JSON.stringify({
        blocks: this._state.blocks,
        rootId: this._state.rootId
      });
    }
    getBlock(id) {
      return this._state?.blocks[id] ?? null;
    }
    getFieldData(blockId) {
      return this._state?.blocks[blockId]?.fields || [];
    }
    setFieldValue(blockId, fieldId, value) {
      const block = this._state?.blocks[blockId];
      if (!block?.fields) return;
      const field = block.fields.find((f) => f.fieldId === fieldId);
      if (field) {
        if (!this._history.inFieldBatch) this._history.pushFieldChange(this._state);
        field.value = value;
        this._state.dirty = true;
        this._updateSaveButton();
        this._updateContentPreview(blockId);
      }
    }
    save() {
      if (!this._state) return;
      const serialized = this.serialize();
      this._state.dirty = false;
      this._updateSaveButton();
      this.dispatchEvent(new CustomEvent("block-editor:save", {
        bubbles: true,
        composed: true,
        detail: { state: serialized }
      }));
    }
    // ---- Lifecycle ----
    connectedCallback() {
      window.addEventListener("beforeunload", this._beforeUnloadHandler);
      window.addEventListener("keydown", this._escapeHandler);
      this._initState();
    }
    disconnectedCallback() {
      window.removeEventListener("beforeunload", this._beforeUnloadHandler);
      window.removeEventListener("keydown", this._escapeHandler);
      if (this._history) this._history.clear();
    }
    attributeChangedCallback(name, oldValue, newValue) {
      if (name === "data-state" && this.isConnected) {
        this._initState();
      }
    }
    // ---- State Initialization ----
    _initState() {
      const stateAttr = this.getAttribute("data-state");
      if (stateAttr === null || stateAttr === "") {
        this._state = createState();
        if (this._history) this._history.clear();
        this._render();
        return;
      }
      let parsed;
      try {
        parsed = JSON.parse(stateAttr);
      } catch (parseError) {
        this._renderError("Invalid block data", parseError.message);
        this.dispatchEvent(new CustomEvent("block-editor:error", {
          bubbles: true,
          composed: true,
          detail: { message: "Invalid block data", error: parseError }
        }));
        return;
      }
      const newState = {
        blocks: parsed.blocks || {},
        rootId: parsed.rootId || null,
        selectedBlockId: null,
        dirty: false
      };
      const validationErrors = validateState(newState);
      if (validationErrors.length > 0) {
        const message = "Block data has inconsistent pointers: " + validationErrors[0];
        this._renderError(message, validationErrors.join("\n"));
        this.dispatchEvent(new CustomEvent("block-editor:error", {
          bubbles: true,
          composed: true,
          detail: { message, error: new Error(validationErrors.join("; ")) }
        }));
        return;
      }
      this._state = newState;
      if (this._history) this._history.clear();
      this._rootDatatypeId = this.getAttribute("data-root-datatype-id") || null;
      this._elementRegistry.clear();
      this._wrapperRegistry.clear();
      this._collapsedBlocks.clear();
      this._render();
    }
    // ---- Rendering ----
    _render() {
      this.innerHTML = "";
      this._elementRegistry.clear();
      this._wrapperRegistry.clear();
      const container = document.createElement("div");
      container.className = "editor-container";
      container.setAttribute("data-editor-container", "");
      const header = this._renderHeader();
      container.appendChild(header);
      const blockList = document.createElement("div");
      blockList.className = "block-list";
      this._renderBlocksInto(blockList);
      container.appendChild(blockList);
      container.setAttribute("tabindex", "0");
      container.addEventListener("click", (e) => this._handleClick(e));
      container.addEventListener("pointerdown", (e) => this._onPointerDown(e));
      container.addEventListener("keydown", this._keydownHandler);
      var self = this;
      container.addEventListener("focus", function() {
        if (self._pointerSelectActive) return;
        if (!self._state.selectedBlockId) {
          var order = getBlockTraversalOrder(self._state);
          if (order.length > 0) self._selectBlock(order[0]);
        }
      });
      this.appendChild(container);
    }
    _renderHeader() {
      const header = document.createElement("div");
      header.className = "editor-header";
      const collapseControls = document.createElement("div");
      collapseControls.className = "collapse-controls";
      const expandAllBtn = document.createElement("button");
      expandAllBtn.textContent = "Expand All";
      expandAllBtn.className = "collapse-btn";
      expandAllBtn.dataset.action = "expand-all";
      collapseControls.appendChild(expandAllBtn);
      const collapseAllBtn = document.createElement("button");
      collapseAllBtn.textContent = "Collapse All";
      collapseAllBtn.className = "collapse-btn";
      collapseAllBtn.dataset.action = "collapse-all";
      collapseControls.appendChild(collapseAllBtn);
      header.appendChild(collapseControls);
      const saveBtn = document.createElement("button");
      saveBtn.textContent = "Save";
      saveBtn.className = "save-btn";
      saveBtn.dataset.action = "save";
      header.appendChild(saveBtn);
      return header;
    }
    _renderBlocksInto(container) {
      if (!this._state) return;
      const rootList = getRootList(this._state);
      if (rootList.length === 0) {
        container.appendChild(this._renderEmptyState());
        return;
      }
      container.appendChild(this._renderInsertButton("before", rootList[0].id));
      for (var i = 0; i < rootList.length; i++) {
        const block = rootList[i];
        const wrapper = this._renderBlockWrapper(block, 0);
        container.appendChild(wrapper);
        container.appendChild(this._renderInsertButton("after", block.id));
      }
    }
    /**
     * Render a block-wrapper div containing the block-item header and
     * optionally a children-container. The wrapper is indented by depth.
     */
    _renderBlockWrapper(block, depth) {
      const wrapper = document.createElement("div");
      wrapper.className = "block-wrapper";
      if (this._collapsedBlocks.has(block.id)) {
        wrapper.classList.add("collapsed");
      }
      wrapper.dataset.blockId = block.id;
      wrapper.style.marginInlineStart = depth * 24 + "px";
      const header = this._renderBlockHeader(block);
      wrapper.appendChild(header);
      const typeConfig = getTypeConfig(block.type);
      const children = getChildren(this._state, block.id);
      if (children.length > 0) {
        const childContainer = document.createElement("div");
        childContainer.className = "children-container";
        childContainer.dataset.parentId = block.id;
        childContainer.appendChild(this._renderInsertButton("before", children[0].id));
        for (var ci = 0; ci < children.length; ci++) {
          const child = children[ci];
          childContainer.appendChild(this._renderBlockWrapper(child, depth + 1));
          childContainer.appendChild(this._renderInsertButton("after", child.id));
        }
        wrapper.appendChild(childContainer);
      } else if (typeConfig.canHaveChildren) {
        const insideBtn = this._renderInsertButton("inside", block.id);
        insideBtn.classList.add("insert-btn--inside");
        wrapper.appendChild(insideBtn);
      }
      this._wrapperRegistry.set(block.id, wrapper);
      return wrapper;
    }
    /**
     * Render the block-item header element (badge, label, child count, delete button,
     * type-specific content).
     */
    _renderBlockHeader(block) {
      const el = document.createElement("div");
      el.className = "block-item";
      el.dataset.blockId = block.id;
      if (block.type === "container") {
        el.classList.add("block-item--container");
      }
      const chevron = document.createElement("button");
      chevron.className = "block-chevron";
      chevron.dataset.action = "toggle-collapse";
      chevron.dataset.blockId = block.id;
      chevron.textContent = this._collapsedBlocks.has(block.id) ? "\u25B8" : "\u25BE";
      chevron.title = "Toggle collapse";
      el.appendChild(chevron);
      const badge = document.createElement("span");
      badge.className = "block-type-badge block-type-badge--" + block.type;
      badge.textContent = getTypeConfig(block.type).label;
      el.appendChild(badge);
      const label = document.createElement("span");
      label.className = "block-label";
      label.textContent = block.label;
      el.appendChild(label);
      const childCount = getDescendantCount(this._state, block.id);
      if (childCount > 0) {
        const countBadge = document.createElement("span");
        countBadge.className = "child-count-badge";
        countBadge.textContent = String(childCount);
        countBadge.title = childCount + " descendant" + (childCount === 1 ? "" : "s");
        el.appendChild(countBadge);
      }
      const hoverToolbar = this._renderHoverToolbar(block.id);
      el.appendChild(hoverToolbar);
      const preview = this._renderContentPreview(block);
      if (preview) {
        el.appendChild(preview);
      }
      this._elementRegistry.set(block.id, el);
      return el;
    }
    /**
     * Render the hover toolbar for a block header.
     * Contains: Move Up, Move Down, Indent, Outdent, Duplicate, Delete.
     * All buttons have data-action attributes so _onPointerDown excludes
     * them from drag initiation.
     */
    _renderHoverToolbar(blockId) {
      const toolbar = document.createElement("div");
      toolbar.className = "block-hover-toolbar";
      const actions = [
        { label: "\u2191", action: "toolbar-move-up", title: "Move Up" },
        { label: "\u2193", action: "toolbar-move-down", title: "Move Down" },
        { label: "\u2192", action: "toolbar-indent", title: "Indent (>)" },
        { label: "\u2190", action: "toolbar-outdent", title: "Outdent (<)" },
        { label: "\u2398", action: "toolbar-duplicate", title: "Duplicate (Ctrl+Shift+D)" },
        { label: "\u2715", action: "toolbar-delete", title: "Delete" }
      ];
      for (const def of actions) {
        const btn = document.createElement("button");
        btn.textContent = def.label;
        btn.title = def.title;
        btn.dataset.action = def.action;
        btn.dataset.blockId = blockId;
        toolbar.appendChild(btn);
      }
      return toolbar;
    }
    /**
     * Render content preview showing actual field values.
     * Returns null for container blocks (children ARE the content).
     */
    _renderContentPreview(block) {
      if (block.type === "container") return null;
      var preview = document.createElement("div");
      preview.className = "block-content-preview";
      var fields = block.fields || [];
      var hasContent = false;
      for (var i = 0; i < fields.length; i++) {
        var value = (fields[i].value || "").trim();
        if (!value) continue;
        if (value.length === 26 && /^[0-9A-HJKMNP-TV-Z]{26}$/.test(value)) continue;
        hasContent = true;
        var el = document.createElement("div");
        if (i === 0 && block.type === "heading") {
          el.className = "preview-heading";
        } else if (value.length > 200 || value.indexOf("\n") !== -1) {
          el.className = "preview-text";
        } else {
          el.className = "preview-field";
        }
        if (value.length > 500) {
          value = value.substring(0, 500) + "\u2026";
        }
        el.textContent = value;
        preview.appendChild(el);
      }
      if (!hasContent) {
        var empty = document.createElement("div");
        empty.className = "preview-empty";
        empty.textContent = "No content yet";
        preview.appendChild(empty);
      }
      return preview;
    }
    /**
     * Render children of a parent block into its children container.
     * Used during DOM patching when a new children container is created.
     */
    _renderChildrenInto(childContainer, parentId, parentDepth) {
      const children = getChildren(this._state, parentId);
      for (const child of children) {
        const wrapper = this._renderBlockWrapper(child, parentDepth + 1);
        childContainer.appendChild(wrapper);
      }
    }
    // ---- Insert Buttons & Dialog ----
    _renderEmptyState() {
      var container = document.createElement("div");
      container.className = "insert-empty";
      var btn = document.createElement("button");
      btn.className = "insert-btn insert-btn--empty";
      btn.title = "Add content block";
      btn.innerHTML = "+";
      btn.dataset.action = "insert";
      btn.dataset.position = "root";
      container.appendChild(btn);
      return container;
    }
    _renderInsertButton(position, targetId) {
      var btn = document.createElement("button");
      btn.className = "insert-btn";
      btn.title = "Insert block";
      btn.innerHTML = "+";
      btn.dataset.action = "insert";
      btn.dataset.position = position;
      btn.dataset.targetId = targetId || "";
      return btn;
    }
    _renderError(message, detail) {
      this.innerHTML = "";
      const container = document.createElement("div");
      container.className = "editor-container";
      container.setAttribute("data-editor-container", "");
      const errorDiv = document.createElement("div");
      errorDiv.className = "error-container";
      const msgEl = document.createElement("div");
      msgEl.className = "error-message";
      msgEl.textContent = message;
      errorDiv.appendChild(msgEl);
      if (detail) {
        const detailEl = document.createElement("div");
        detailEl.className = "error-detail";
        detailEl.textContent = detail;
        errorDiv.appendChild(detailEl);
      }
      container.appendChild(errorDiv);
      this.appendChild(container);
    }
    // ---- Event Handling ----
    _handleClick(e) {
      const target = e.target;
      const action = target.dataset?.action;
      if (!action) return;
      if (action === "insert") {
        var position = target.dataset.position;
        var targetId = target.dataset.targetId || null;
        this._openPicker(targetId, position);
        return;
      } else if (action === "add") {
        const blockType = target.dataset.blockType || "text";
        this._doAddBlock(blockType);
      } else if (action === "delete" || action === "toolbar-delete") {
        this._doDeleteBlock(target.dataset.blockId);
      } else if (action === "save") {
        this.save();
      } else if (action === "toolbar-move-up") {
        this._doMoveBlockUp(target.dataset.blockId);
      } else if (action === "toolbar-move-down") {
        this._doMoveBlockDown(target.dataset.blockId);
      } else if (action === "toolbar-indent") {
        this._doIndentBlock(target.dataset.blockId);
      } else if (action === "toolbar-outdent") {
        this._doOutdentBlock(target.dataset.blockId);
      } else if (action === "toolbar-duplicate") {
        this._doDuplicateBlock(target.dataset.blockId);
      } else if (action === "toggle-collapse") {
        this._toggleCollapse(target.dataset.blockId);
      } else if (action === "expand-all") {
        this._expandAll();
      } else if (action === "collapse-all") {
        this._collapseAll();
      }
    }
    _doAddBlock(type) {
      this._history.pushUndo(this._state);
      const id = addBlock(this._state, type);
      this._devValidate();
      const block = this._state.blocks[id];
      const wrapper = this._renderBlockWrapper(block, 0);
      const blockList = this.querySelector(".block-list");
      blockList.appendChild(wrapper);
      this._updateSaveButton();
      this.dispatchEvent(new CustomEvent("block-editor:change", {
        bubbles: true,
        composed: true,
        detail: { action: "add", blockId: id }
      }));
    }
    _doDeleteBlock(blockId) {
      const block = this._state.blocks[blockId];
      if (!block) return;
      const descendantCount = getDescendantCount(this._state, blockId);
      if (descendantCount > 0) {
        var self = this;
        showConfirmDialog({
          title: "Delete Block",
          message: 'Delete "' + block.label + '" and ' + descendantCount + " children?",
          confirmLabel: "Delete",
          destructive: true
        }).then(function(confirmed) {
          if (confirmed) self._executeDeleteBlock(blockId);
        });
        return;
      }
      this._executeDeleteBlock(blockId);
    }
    _executeDeleteBlock(blockId) {
      const block = this._state.blocks[blockId];
      if (!block) return;
      this._history.pushUndo(this._state);
      const parentId = block.parentId;
      const removedIds = removeBlock(this._state, blockId);
      this._devValidate();
      if (this._state.selectedBlockId && removedIds.includes(this._state.selectedBlockId)) {
        this._state.selectedBlockId = null;
        this.dispatchEvent(new CustomEvent("block-editor:select", {
          bubbles: true,
          composed: true,
          detail: { blockId: null }
        }));
      }
      for (const id of removedIds) {
        this._collapsedBlocks.delete(id);
        const wrapper = this._wrapperRegistry.get(id);
        if (wrapper) {
          wrapper.remove();
          this._wrapperRegistry.delete(id);
        }
        const el = this._elementRegistry.get(id);
        if (el) {
          this._elementRegistry.delete(id);
        }
      }
      this._cleanupEmptyChildrenContainer(parentId);
      this._updateChildCountBadge(parentId);
      this._updateSaveButton();
      this.dispatchEvent(new CustomEvent("block-editor:change", {
        bubbles: true,
        composed: true,
        detail: { action: "remove", blockId }
      }));
    }
    // ---- Selection ----
    _selectBlock(blockId) {
      if (!this._state) return;
      if (this._state.selectedBlockId) {
        const prevEl = this._elementRegistry.get(this._state.selectedBlockId);
        if (prevEl) {
          prevEl.classList.remove("selected");
        }
      }
      if (this._state.selectedBlockId === blockId) {
        this._state.selectedBlockId = null;
        this.dispatchEvent(new CustomEvent("block-editor:select", {
          bubbles: true,
          composed: true,
          detail: { blockId: null }
        }));
        return;
      }
      this._state.selectedBlockId = blockId;
      const el = this._elementRegistry.get(blockId);
      if (el) {
        el.classList.add("selected");
        el.scrollIntoView({ block: "nearest", behavior: "smooth" });
      }
      this.dispatchEvent(new CustomEvent("block-editor:select", {
        bubbles: true,
        composed: true,
        detail: { blockId }
      }));
    }
    // ---- Keyboard Shortcuts ----
    _onKeyDown(e) {
      if (this._pickerOpen) {
        this._onPickerKeyDown(e);
        return;
      }
      if (!this._state) return;
      if ((e.ctrlKey || e.metaKey) && !e.shiftKey && e.key === "z") {
        e.preventDefault();
        if (!this._drag) this._undo();
        return;
      }
      if ((e.ctrlKey || e.metaKey) && (e.shiftKey && (e.key === "z" || e.key === "Z") || e.key === "y")) {
        e.preventDefault();
        if (!this._drag) this._redo();
        return;
      }
      var blockId = this._state.selectedBlockId;
      var noMod = !e.ctrlKey && !e.metaKey && !e.altKey;
      if (e.key === "Tab") {
        e.preventDefault();
        this._navigateDFS(e.shiftKey);
        return;
      }
      if (e.key === "ArrowDown" || e.key === "j" && noMod) {
        e.preventDefault();
        this._navigateDFS(false);
        return;
      }
      if (e.key === "ArrowUp" || e.key === "k" && noMod) {
        e.preventDefault();
        this._navigateDFS(true);
        return;
      }
      if (e.key === "ArrowLeft" || e.key === "h" && noMod) {
        if (!blockId) return;
        var parentId = this._state.blocks[blockId].parentId;
        if (parentId) {
          e.preventDefault();
          this._selectBlock(parentId);
        }
        return;
      }
      if (e.key === "ArrowRight" || e.key === "l" && noMod) {
        if (!blockId) return;
        var childId = this._state.blocks[blockId].firstChildId;
        if (childId) {
          e.preventDefault();
          this._selectBlock(childId);
        }
        return;
      }
      if (e.key === ">" && noMod) {
        if (!blockId) return;
        e.preventDefault();
        this._doIndentBlock(blockId);
        return;
      }
      if (e.key === "<" && noMod) {
        if (!blockId) return;
        e.preventDefault();
        this._doOutdentBlock(blockId);
        return;
      }
      if (e.key === "d" || e.key === "D") {
        if ((e.ctrlKey || e.metaKey) && e.shiftKey) {
          if (!blockId) return;
          e.preventDefault();
          this._doDuplicateBlock(blockId);
          return;
        }
      }
      if (e.key === "Delete" || e.key === "Backspace") {
        if (!blockId) return;
        e.preventDefault();
        this._doDeleteBlock(blockId);
        return;
      }
      if (e.key === "Enter") {
        if (!blockId) return;
        e.preventDefault();
        this._openPicker(blockId, "after");
        return;
      }
    }
    /**
     * Navigate DFS order: select next (backward=false) or previous (backward=true) block.
     * Auto-selects first or last block if nothing is currently selected.
     */
    _navigateDFS(backward) {
      var order = getBlockTraversalOrder(this._state);
      if (order.length === 0) return;
      var blockId = this._state.selectedBlockId;
      if (!blockId) {
        this._selectBlock(backward ? order[order.length - 1] : order[0]);
        return;
      }
      var currentIndex = order.indexOf(blockId);
      if (currentIndex === -1) return;
      var nextIndex = backward ? currentIndex - 1 : currentIndex + 1;
      if (nextIndex < 0 || nextIndex >= order.length) return;
      this._selectBlock(order[nextIndex]);
    }
    _updateSaveButton() {
    }
    // ---- Collapse / Expand ----
    _toggleCollapse(blockId) {
      if (!blockId) return;
      var wrapper = this._wrapperRegistry.get(blockId);
      if (!wrapper) return;
      if (this._collapsedBlocks.has(blockId)) {
        this._collapsedBlocks.delete(blockId);
        wrapper.classList.remove("collapsed");
      } else {
        this._collapsedBlocks.add(blockId);
        wrapper.classList.add("collapsed");
      }
      var chevron = wrapper.querySelector(":scope > .block-item > .block-chevron");
      if (chevron) {
        chevron.textContent = this._collapsedBlocks.has(blockId) ? "\u25B8" : "\u25BE";
      }
    }
    _expandAll() {
      this._collapsedBlocks.clear();
      this._render();
    }
    _collapseAll() {
      for (var id in this._state.blocks) {
        this._collapsedBlocks.add(id);
      }
      this._render();
    }
    _updateContentPreview(blockId) {
      var el = this._elementRegistry.get(blockId);
      if (!el) return;
      var block = this._state.blocks[blockId];
      if (!block) return;
      var existingPreview = el.querySelector(".block-content-preview");
      var newPreview = this._renderContentPreview(block);
      if (existingPreview && newPreview) {
        existingPreview.replaceWith(newPreview);
      } else if (existingPreview && !newPreview) {
        existingPreview.remove();
      } else if (!existingPreview && newPreview) {
        el.appendChild(newPreview);
      }
    }
    // ---- Dev-mode validation ----
    _devValidate() {
      if (!this.dev) return;
      const errors = validateState(this._state);
      if (errors.length > 0) {
        console.warn("[block-editor] Validation errors after mutation:", errors);
      } else {
        console.log("[block-editor] State valid");
      }
    }
    // ---- beforeunload ----
    _onBeforeUnload(e) {
      if (this._state?.dirty) {
        e.preventDefault();
        e.returnValue = "You have unsaved changes.";
      }
    }
    // ---- Undo / Redo ----
    _undo() {
      if (!this._history.canUndo) return;
      var entry = this._history.popUndo(this._state);
      this._restoreSnapshot(entry);
    }
    _redo() {
      if (!this._history.canRedo) return;
      var entry = this._history.popRedo(this._state);
      this._restoreSnapshot(entry);
    }
    _restoreSnapshot(entry) {
      this._state.blocks = entry.snapshot.blocks;
      this._state.rootId = entry.snapshot.rootId;
      this._state.selectedBlockId = null;
      this._state.dirty = true;
      this._render();
      if (entry.selectedBlockId && this._state.blocks[entry.selectedBlockId]) {
        this._selectBlock(entry.selectedBlockId);
      }
      this.dispatchEvent(new CustomEvent("block-editor:change", {
        bubbles: true,
        composed: true,
        detail: { action: "undo-redo" }
      }));
    }
    remapIds(idMap) {
      var clientIds = Object.keys(idMap);
      for (var i = 0; i < clientIds.length; i++) {
        var clientId = clientIds[i];
        if (this._elementRegistry.has(clientId)) {
          var el = this._elementRegistry.get(clientId);
          this._elementRegistry.delete(clientId);
          this._elementRegistry.set(idMap[clientId], el);
          el.dataset.blockId = idMap[clientId];
        }
        if (this._wrapperRegistry.has(clientId)) {
          var wrapper = this._wrapperRegistry.get(clientId);
          this._wrapperRegistry.delete(clientId);
          this._wrapperRegistry.set(idMap[clientId], wrapper);
          wrapper.dataset.blockId = idMap[clientId];
        }
      }
      this._history.remapIds(idMap);
      if (this._state.selectedBlockId && idMap[this._state.selectedBlockId] !== void 0) {
        this._state.selectedBlockId = idMap[this._state.selectedBlockId];
      }
    }
  }
  Object.assign(BlockEditor.prototype, dragMethods, domPatchMethods, pickerMethods);
  customElements.define("block-editor", BlockEditor);
}
export {
  BLOCK_TYPE_CONFIG,
  MAX_DEPTH,
  addBlock,
  addBlockFromDatatype,
  collectDescendants,
  createState,
  duplicateBlock,
  findLastSibling,
  generateId,
  getBlockTraversalOrder,
  getChildren,
  getDepth,
  getDescendantCount,
  getRootList,
  getTypeConfig,
  indentBlock,
  insertAfter,
  insertAsFirstChild,
  insertAsLastChild,
  insertBefore,
  isDescendant,
  moveBlock,
  moveBlockDown,
  moveBlockUp,
  outdentBlock,
  removeBlock,
  unlink,
  validateState
};
