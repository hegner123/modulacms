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
      if (block.parentId !== null) {
        errors.push(`Root chain block "${currentId}" has parentId "${block.parentId}", expected null`);
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

// internal/admin/static/js/block-editor-src/cache.js
var _dtCache = {
  data: null,
  // Array of {id, label, type} from API
  fetchedAt: 0,
  // timestamp ms
  ttl: 5 * 60 * 1e3,
  // 5 minutes
  pending: null
  // in-flight promise to deduplicate concurrent fetches
};
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
      return { id: dt.datatype_id, label: dt.label, type: dt.type };
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
    const result = indentBlock(this._state, blockId);
    if (!result) return;
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
    const result = outdentBlock(this._state, blockId);
    if (!result) return;
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
    const result = moveBlockUp(this._state, blockId);
    if (!result) return;
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
    const result = moveBlockDown(this._state, blockId);
    if (!result) return;
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
    const block = this._state?.blocks[blockId];
    if (!block) return;
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
    const editorContainer = this.querySelector(".editor-container");
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
    this._drag = null;
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
      this._beforeUnloadHandler = this._onBeforeUnload.bind(this);
      this._drag = null;
      this._dropIndicator = null;
      this._escapeHandler = this._onEscapeKey.bind(this);
      this._autoScrollRaf = null;
      this._lastPointerY = 0;
      this._keydownHandler = this._onKeyDown.bind(this);
    }
    get dev() {
      return this.hasAttribute("data-dev");
    }
    get state() {
      return this._state;
    }
    set state(newState) {
      this._state = newState;
      this._state.dirty = false;
      this._elementRegistry.clear();
      this._wrapperRegistry.clear();
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
        field.value = value;
        this._state.dirty = true;
        this._updateSaveButton();
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
      this._elementRegistry.clear();
      this._wrapperRegistry.clear();
      this._render();
    }
    // ---- Rendering ----
    _render() {
      this.innerHTML = "";
      this._elementRegistry.clear();
      this._wrapperRegistry.clear();
      const container = document.createElement("div");
      container.className = "editor-container";
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
      this.appendChild(container);
    }
    _renderHeader() {
      const header = document.createElement("div");
      header.className = "editor-header";
      const saveBtn = document.createElement("button");
      saveBtn.textContent = "Save";
      saveBtn.className = "save-btn";
      saveBtn.dataset.action = "save";
      saveBtn.disabled = !this._state?.dirty;
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
      const deleteBtn = document.createElement("button");
      deleteBtn.className = "block-delete-btn";
      deleteBtn.textContent = "Delete";
      deleteBtn.dataset.action = "delete";
      deleteBtn.dataset.blockId = block.id;
      el.appendChild(deleteBtn);
      const hoverToolbar = this._renderHoverToolbar(block.id);
      el.appendChild(hoverToolbar);
      const content = this._renderTypeContent(block);
      if (content) {
        el.appendChild(content);
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
        { label: "\u2192", action: "toolbar-indent", title: "Indent (Tab)" },
        { label: "\u2190", action: "toolbar-outdent", title: "Outdent (Shift+Tab)" },
        { label: "Dup", action: "toolbar-duplicate", title: "Duplicate (Ctrl+Shift+D)" },
        { label: "Del", action: "toolbar-delete", title: "Delete" }
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
     * Render type-specific content for a block.
     * text = paragraph placeholder lines
     * heading = large bold label
     * image = dashed rect placeholder
     * container = labeled children area indicator
     */
    _renderTypeContent(block) {
      const wrapper = document.createElement("div");
      wrapper.className = "block-type-content block-type-content--" + block.type;
      if (block.type === "text") {
        for (let i = 0; i < 3; i++) {
          const line = document.createElement("div");
          line.className = "text-line";
          wrapper.appendChild(line);
        }
        return wrapper;
      }
      if (block.type === "heading") {
        wrapper.textContent = block.label;
        return wrapper;
      }
      if (block.type === "image") {
        wrapper.textContent = "Image placeholder";
        return wrapper;
      }
      if (block.type === "container") {
        const childCount = getDescendantCount(this._state, block.id);
        if (childCount > 0) {
          wrapper.textContent = childCount + " block" + (childCount === 1 ? "" : "s") + " inside";
        } else {
          wrapper.textContent = "Drop blocks here";
        }
        return wrapper;
      }
      return null;
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
    _openInsertDialog(position, targetId) {
      var self = this;
      fetchDatatypes().then(function(datatypes) {
        self._showDialog(datatypes, position, targetId);
      }).catch(function(err) {
        console.error("[block-editor] Failed to load datatypes:", err);
      });
    }
    _showDialog(datatypes, position, targetId) {
      var self = this;
      var backdrop = document.createElement("div");
      backdrop.className = "insert-dialog-backdrop";
      var panel = document.createElement("div");
      panel.className = "insert-dialog-panel";
      var title = document.createElement("div");
      title.className = "insert-dialog-title";
      title.textContent = "Select Content Type";
      panel.appendChild(title);
      if (datatypes.length === 0) {
        var emptyMsg = document.createElement("div");
        emptyMsg.className = "insert-dialog-empty";
        emptyMsg.textContent = "No content types available";
        panel.appendChild(emptyMsg);
      } else {
        for (var i = 0; i < datatypes.length; i++) {
          var dt = datatypes[i];
          var option = document.createElement("button");
          option.className = "insert-dialog-option";
          option.dataset.datatypeIdx = String(i);
          var labelSpan = document.createElement("span");
          labelSpan.className = "insert-dialog-option-label";
          labelSpan.textContent = dt.label;
          option.appendChild(labelSpan);
          var typeSpan = document.createElement("span");
          typeSpan.className = "insert-dialog-option-type";
          typeSpan.textContent = dt.type;
          option.appendChild(typeSpan);
          option.addEventListener("click", /* @__PURE__ */ (function(datatype) {
            return function() {
              self._closeDialog();
              self._onDatatypeSelected(datatype, position, targetId);
            };
          })(dt));
          panel.appendChild(option);
        }
      }
      var cancelBtn = document.createElement("button");
      cancelBtn.className = "insert-dialog-cancel";
      cancelBtn.textContent = "Cancel";
      cancelBtn.addEventListener("click", function() {
        self._closeDialog();
      });
      panel.appendChild(cancelBtn);
      backdrop.appendChild(panel);
      backdrop.addEventListener("click", function(e) {
        if (e.target === backdrop) {
          self._closeDialog();
        }
      });
      this._dialogEscHandler = function(e) {
        if (e.key === "Escape") {
          self._closeDialog();
        }
      };
      window.addEventListener("keydown", this._dialogEscHandler);
      this._dialogBackdrop = backdrop;
      this.appendChild(backdrop);
    }
    _closeDialog() {
      if (this._dialogBackdrop) {
        this._dialogBackdrop.remove();
        this._dialogBackdrop = null;
      }
      if (this._dialogEscHandler) {
        window.removeEventListener("keydown", this._dialogEscHandler);
        this._dialogEscHandler = null;
      }
    }
    _onDatatypeSelected(datatype, position, targetId) {
      if (!this._state) return;
      var id = addBlockFromDatatype(this._state, datatype, position, targetId);
      this._devValidate();
      this._render();
      this._selectBlock(id);
      this.dispatchEvent(new CustomEvent("block-editor:change", {
        bubbles: true,
        composed: true,
        detail: { action: "add", blockId: id }
      }));
    }
    _renderError(message, detail) {
      this.innerHTML = "";
      const container = document.createElement("div");
      container.className = "editor-container";
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
        this._openInsertDialog(position, targetId);
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
      }
    }
    _doAddBlock(type) {
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
        const confirmed = confirm(`Delete "${block.label}" and ${descendantCount} children?`);
        if (!confirmed) return;
      }
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
      }
      this.dispatchEvent(new CustomEvent("block-editor:select", {
        bubbles: true,
        composed: true,
        detail: { blockId }
      }));
    }
    // ---- Keyboard Shortcuts ----
    _onKeyDown(e) {
      if (!this._state) return;
      if (e.key === "Tab") {
        const blockId = this._state.selectedBlockId;
        if (!blockId) return;
        e.preventDefault();
        if (e.shiftKey) {
          this._doOutdentBlock(blockId);
        } else {
          this._doIndentBlock(blockId);
        }
        return;
      }
      if (e.key === "ArrowUp" || e.key === "ArrowDown") {
        const blockId = this._state.selectedBlockId;
        if (!blockId) return;
        const order = getBlockTraversalOrder(this._state);
        if (order.length === 0) return;
        const currentIndex = order.indexOf(blockId);
        if (currentIndex === -1) return;
        let nextIndex;
        if (e.key === "ArrowUp") {
          nextIndex = currentIndex - 1;
        } else {
          nextIndex = currentIndex + 1;
        }
        if (nextIndex < 0 || nextIndex >= order.length) return;
        e.preventDefault();
        this._selectBlock(order[nextIndex]);
        return;
      }
      if (e.key === "d" || e.key === "D") {
        if ((e.ctrlKey || e.metaKey) && e.shiftKey) {
          const blockId = this._state.selectedBlockId;
          if (!blockId) return;
          e.preventDefault();
          this._doDuplicateBlock(blockId);
          return;
        }
      }
      if (e.key === "Delete" || e.key === "Backspace") {
        const blockId = this._state.selectedBlockId;
        if (!blockId) return;
        e.preventDefault();
        this._doDeleteBlock(blockId);
        return;
      }
      if (e.key === "Enter") {
        const blockId = this._state.selectedBlockId;
        if (!blockId) return;
        e.preventDefault();
        this._doAddBlockAfter(blockId, "text");
        return;
      }
    }
    _updateSaveButton() {
      const saveBtn = this.querySelector('[data-action="save"]');
      if (saveBtn) {
        saveBtn.disabled = !this._state?.dirty;
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
  }
  Object.assign(BlockEditor.prototype, dragMethods, domPatchMethods);
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
