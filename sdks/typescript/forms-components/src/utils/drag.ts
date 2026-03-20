/**
 * Simplified drag-and-drop module for flat list reorder.
 * Uses native Pointer Events with setPointerCapture for reliable tracking.
 * Ported from block editor patterns — no tree nesting, flat list only.
 */

export interface DragOptions {
  onReorder: (fromIndex: number, toIndex: number) => void;
  handleSelector: string;
}

/** Pixels the pointer must travel before a drag is initiated. */
const DRAG_THRESHOLD = 8;

/** Pixels from container edge that trigger auto-scroll. */
const SCROLL_EDGE_ZONE = 40;

/** Pixels per frame to auto-scroll when near container edges. */
const SCROLL_SPEED = 8;

/** Height of the visual drop indicator line in pixels. */
const INDICATOR_HEIGHT = 2;

interface DragState {
  /** Index of the item being dragged. */
  fromIndex: number;
  /** The original item element. */
  sourceElement: HTMLElement;
  /** Fixed-position clone that follows the cursor. */
  overlay: HTMLElement;
  /** Visual indicator showing the drop position. */
  indicator: HTMLElement;
  /** Current computed drop index, or -1 if none. */
  currentDropIndex: number;
  /** Pointer coordinates at drag start for threshold check. */
  startX: number;
  startY: number;
  /** Whether the threshold has been exceeded and drag is active. */
  active: boolean;
  /** ID from requestAnimationFrame for auto-scroll. */
  scrollRafId: number;
  /** Pointer ID captured via setPointerCapture. */
  pointerId: number;
  /** The handle element that received the initial pointerdown. */
  handleElement: HTMLElement;
}

/**
 * Initialize drag-and-drop on a container element.
 * Items are the direct children of the container.
 * Returns a cleanup function that removes all listeners.
 */
export function initDragAndDrop(
  container: HTMLElement,
  options: DragOptions,
): () => void {
  let state: DragState | null = null;

  function getItems(): HTMLElement[] {
    return Array.from(container.children) as HTMLElement[];
  }

  function getItemIndex(el: HTMLElement): number {
    const items = getItems();
    return items.indexOf(el);
  }

  /**
   * Walk up from a target element to find the direct child of container
   * that contains it.
   */
  function findItemFromTarget(target: HTMLElement): HTMLElement | null {
    let current: HTMLElement | null = target;
    while (current && current !== container) {
      if (current.parentElement === container) {
        return current;
      }
      current = current.parentElement;
    }
    return null;
  }

  /**
   * Create the fixed-position overlay clone that tracks the cursor.
   */
  function createOverlay(source: HTMLElement, x: number, y: number): HTMLElement {
    const rect = source.getBoundingClientRect();
    const overlay = source.cloneNode(true) as HTMLElement;
    overlay.classList.add("drag-overlay");
    overlay.style.position = "fixed";
    overlay.style.left = "0";
    overlay.style.top = "0";
    overlay.style.width = `${rect.width}px`;
    overlay.style.height = `${rect.height}px`;
    overlay.style.pointerEvents = "none";
    overlay.style.zIndex = "10000";
    overlay.style.willChange = "transform";
    overlay.style.transform = `translate(${x - rect.width / 2}px, ${y - rect.height / 2}px)`;
    document.body.appendChild(overlay);
    return overlay;
  }

  /**
   * Create the drop indicator line element.
   */
  function createIndicator(): HTMLElement {
    const indicator = document.createElement("div");
    indicator.classList.add("drop-indicator");
    indicator.style.position = "absolute";
    indicator.style.left = "0";
    indicator.style.right = "0";
    indicator.style.height = `${INDICATOR_HEIGHT}px`;
    indicator.style.pointerEvents = "none";
    indicator.style.zIndex = "9999";
    indicator.style.display = "none";

    // Ensure container can anchor absolutely-positioned children
    const currentPosition = getComputedStyle(container).position;
    if (currentPosition === "static") {
      container.style.position = "relative";
    }

    container.appendChild(indicator);
    return indicator;
  }

  /**
   * Determine drop index based on pointer Y relative to sibling rects.
   * Uses two-zone detection: top 50% = before, bottom 50% = after.
   */
  function computeDropIndex(clientY: number): number {
    const items = getItems();
    const containerRect = container.getBoundingClientRect();

    // Above all items: drop at index 0
    if (clientY < containerRect.top) {
      return 0;
    }

    for (let i = 0; i < items.length; i++) {
      const item = items[i] as HTMLElement | undefined;
      if (!item) {
        continue;
      }

      // Skip the source element and the indicator
      if (state && item === state.sourceElement) {
        continue;
      }
      if (item.classList.contains("drop-indicator")) {
        continue;
      }

      const rect = item.getBoundingClientRect();
      if (clientY < rect.top || clientY > rect.bottom) {
        continue;
      }

      const midY = rect.top + rect.height / 2;
      if (clientY < midY) {
        return i;
      }
      return i + 1;
    }

    // Below all items: drop at end
    return items.length;
  }

  /**
   * Position the drop indicator line at the given drop index.
   */
  function positionIndicator(dropIndex: number): void {
    if (!state) {
      return;
    }

    const items = getItems();
    const indicator = state.indicator;
    const containerRect = container.getBoundingClientRect();

    // Filter out the indicator itself from the item list for positioning
    const realItems = items.filter(
      (el) => el !== indicator,
    );

    if (dropIndex < 0 || realItems.length === 0) {
      indicator.style.display = "none";
      return;
    }

    let topOffset: number;

    if (dropIndex >= realItems.length) {
      // After the last item
      const lastItem = realItems[realItems.length - 1];
      if (!lastItem) {
        indicator.style.display = "none";
        return;
      }
      const lastRect = lastItem.getBoundingClientRect();
      topOffset = lastRect.bottom - containerRect.top;
    } else {
      // Before the item at dropIndex
      const targetItem = realItems[dropIndex];
      if (!targetItem) {
        indicator.style.display = "none";
        return;
      }
      const targetRect = targetItem.getBoundingClientRect();
      topOffset = targetRect.top - containerRect.top;
    }

    indicator.style.display = "block";
    indicator.style.top = `${topOffset - INDICATOR_HEIGHT / 2}px`;
  }

  /**
   * Auto-scroll the container when the cursor is near its edges.
   * Runs via requestAnimationFrame for smooth scrolling.
   */
  function autoScroll(clientY: number): void {
    if (!state) {
      return;
    }

    const rect = container.getBoundingClientRect();
    const distFromTop = clientY - rect.top;
    const distFromBottom = rect.bottom - clientY;

    let scrollDelta = 0;
    if (distFromTop >= 0 && distFromTop < SCROLL_EDGE_ZONE) {
      scrollDelta = -SCROLL_SPEED;
    } else if (distFromBottom >= 0 && distFromBottom < SCROLL_EDGE_ZONE) {
      scrollDelta = SCROLL_SPEED;
    }

    if (scrollDelta !== 0) {
      container.scrollTop += scrollDelta;
      state.scrollRafId = requestAnimationFrame(() => autoScroll(clientY));
    }
  }

  /**
   * Cancel the current drag operation, cleaning up all state and DOM changes.
   */
  function cancelDrag(): void {
    if (!state) {
      return;
    }

    if (state.scrollRafId) {
      cancelAnimationFrame(state.scrollRafId);
    }

    state.sourceElement.classList.remove("dragging");

    if (state.overlay.parentElement) {
      state.overlay.remove();
    }
    if (state.indicator.parentElement) {
      state.indicator.remove();
    }

    try {
      state.handleElement.releasePointerCapture(state.pointerId);
    } catch {
      // Pointer capture may already be released if the pointer was lost
    }

    state = null;
  }

  /**
   * Complete the drag: call onReorder if the position changed, then clean up.
   */
  function completeDrag(): void {
    if (!state) {
      return;
    }

    const fromIndex = state.fromIndex;
    let toIndex = state.currentDropIndex;

    // Adjust toIndex: when dropping after the source position, subtract 1
    // because the source element will be removed from its original slot first.
    if (toIndex > fromIndex) {
      toIndex--;
    }

    cancelDrag();

    if (toIndex >= 0 && toIndex !== fromIndex) {
      options.onReorder(fromIndex, toIndex);
    }
  }

  // ---------------------------------------------------------------------------
  // Event handlers
  // ---------------------------------------------------------------------------

  function onPointerDown(event: PointerEvent): void {
    // Only primary button
    if (event.button !== 0) {
      return;
    }

    // Already dragging
    if (state) {
      return;
    }

    const target = event.target as HTMLElement;

    // Check if the target or an ancestor matches the handle selector
    const handle = target.closest(options.handleSelector) as HTMLElement | null;
    if (!handle || !container.contains(handle)) {
      return;
    }

    const item = findItemFromTarget(handle);
    if (!item) {
      return;
    }

    const index = getItemIndex(item);
    if (index < 0) {
      return;
    }

    event.preventDefault();

    // Capture the pointer on the handle for reliable cross-element tracking
    handle.setPointerCapture(event.pointerId);

    state = {
      fromIndex: index,
      sourceElement: item,
      overlay: createOverlay(item, event.clientX, event.clientY),
      indicator: createIndicator(),
      currentDropIndex: -1,
      startX: event.clientX,
      startY: event.clientY,
      active: false,
      scrollRafId: 0,
      pointerId: event.pointerId,
      handleElement: handle,
    };

    // Hide overlay until threshold is met
    state.overlay.style.display = "none";
  }

  function onPointerMove(event: PointerEvent): void {
    if (!state) {
      return;
    }

    const dx = event.clientX - state.startX;
    const dy = event.clientY - state.startY;

    // Check threshold before starting the visual drag
    if (!state.active) {
      const distance = Math.sqrt(dx * dx + dy * dy);
      if (distance < DRAG_THRESHOLD) {
        return;
      }
      state.active = true;
      state.sourceElement.classList.add("dragging");
      state.overlay.style.display = "";
    }

    // Move overlay to follow the cursor
    const overlayWidth = state.overlay.offsetWidth;
    const overlayHeight = state.overlay.offsetHeight;
    state.overlay.style.transform =
      `translate(${event.clientX - overlayWidth / 2}px, ${event.clientY - overlayHeight / 2}px)`;

    // Compute drop position
    const dropIndex = computeDropIndex(event.clientY);
    state.currentDropIndex = dropIndex;
    positionIndicator(dropIndex);

    // Auto-scroll near container edges
    if (state.scrollRafId) {
      cancelAnimationFrame(state.scrollRafId);
    }
    state.scrollRafId = requestAnimationFrame(() => autoScroll(event.clientY));
  }

  function onPointerUp(_event: PointerEvent): void {
    if (!state) {
      return;
    }

    if (!state.active) {
      // Threshold was never met — this was a click, not a drag
      cancelDrag();
      return;
    }

    completeDrag();
  }

  function onKeyDown(event: KeyboardEvent): void {
    if (!state) {
      return;
    }

    if (event.key === "Escape") {
      event.preventDefault();
      cancelDrag();
    }
  }

  function onPointerCancel(): void {
    cancelDrag();
  }

  // ---------------------------------------------------------------------------
  // Bind listeners
  // ---------------------------------------------------------------------------

  container.addEventListener("pointerdown", onPointerDown);
  document.addEventListener("pointermove", onPointerMove);
  document.addEventListener("pointerup", onPointerUp);
  document.addEventListener("pointercancel", onPointerCancel);
  document.addEventListener("keydown", onKeyDown);

  /**
   * Cleanup: remove all listeners and abort any in-progress drag.
   */
  return function destroy(): void {
    cancelDrag();
    container.removeEventListener("pointerdown", onPointerDown);
    document.removeEventListener("pointermove", onPointerMove);
    document.removeEventListener("pointerup", onPointerUp);
    document.removeEventListener("pointercancel", onPointerCancel);
    document.removeEventListener("keydown", onKeyDown);
  };
}
