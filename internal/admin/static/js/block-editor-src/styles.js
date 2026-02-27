// styles.js — CSS template literal for adoptedStyleSheets

export const EDITOR_CSS = `
:host {
  display: block;
  --font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  font-family: var(--font-family);
  --bg: #1e1e2e;
  --border: #383850;
  --border-hover: #4a4a6a;
  --text: #cdd6f4;
  --text-muted: #6c7086;
  --accent: #89b4fa;
  --danger: #f38ba8;
  --danger-hover: #e06688;
  --toolbar-bg: #181825;
  --block-bg: #1e1e2e;
  --error-bg: #302028;
  --error-border: #6e3040;
  --error-text: #f38ba8;
  --btn-hover-bg: #252536;
  --save-active-text: #1e1e2e;
  --save-hover-bg: #7aa8ec;
  --save-disabled-bg: #3a4a6a;
  --save-disabled-text: #6c7086;
  --badge-bg: #1e2a45;
  --badge-text: #89b4fa;
  --sp-1: 0.25rem;
  --sp-2: 0.5rem;
  --sp-3: 0.75rem;
  --sp-4: 1rem;
  --sp-6: 1.5rem;
  --sp-8: 2rem;
  --radius-sm: 3px;
  --radius-md: 4px;
  --radius-lg: 6px;
  --radius-xl: 8px;
  --fs-xs: 0.6875rem;
  --fs-sm: 0.75rem;
  --fs-md: 0.8125rem;
  --fs-base: 0.875rem;
}

.editor-container {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  outline: none;
}

.editor-container:focus-visible {
  border-color: var(--accent);
}

.save-btn {
  padding: var(--sp-2) var(--sp-3);
  border: 1px solid var(--accent);
  border-radius: var(--radius-md);
  background: var(--accent);
  color: var(--save-active-text);
  font-size: var(--fs-md);
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.save-btn:hover:not(:disabled) {
  background: var(--save-hover-bg);
}

.save-btn:disabled {
  background: var(--save-disabled-bg);
  border-color: var(--save-disabled-bg);
  color: var(--save-disabled-text);
  opacity: 0.4;
  cursor: not-allowed;
}

.block-list {
  padding: var(--sp-3);
  min-height: 4rem;
}

.block-list:empty::after {
  content: "No blocks yet.";
  display: block;
  padding: var(--sp-8);
  text-align: center;
  color: var(--text-muted);
  font-size: var(--fs-base);
}

.block-item {
  display: flex;
  align-items: center;
  gap: var(--sp-2);
  padding: var(--sp-3) var(--sp-3);
  margin-bottom: var(--sp-2);
  background: var(--block-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  transition: border-color 0.15s;
}

.block-item:hover {
  border-color: var(--border-hover);
}

.block-item.selected {
  border-color: var(--accent);
  box-shadow: 0 0 0 1px var(--accent);
}

.block-type-badge {
  display: inline-block;
  padding: var(--sp-1) var(--sp-2);
  border-radius: var(--radius-sm);
  font-size: var(--fs-xs);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  background: var(--badge-bg);
  color: var(--badge-text);
  flex-shrink: 0;
}

.block-label {
  flex: 1;
  font-size: var(--fs-base);
  color: var(--text);
}

.block-delete-btn {
  padding: var(--sp-1) var(--sp-2);
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  cursor: pointer;
  transition: color 0.15s, background 0.15s;
  flex-shrink: 0;
}

.block-delete-btn:hover {
  color: var(--danger);
  background: var(--error-bg);
}

.error-container {
  padding: var(--sp-6);
  margin: var(--sp-3);
  background: var(--error-bg);
  border: 1px solid var(--error-border);
  border-radius: var(--radius-lg);
}

.error-message {
  color: var(--error-text);
  font-size: var(--fs-base);
  font-weight: 500;
  margin-bottom: var(--sp-3);
}

.error-detail {
  color: var(--error-text);
  font-size: var(--fs-sm);
  opacity: 0.7;
  margin-bottom: var(--sp-3);
  white-space: pre-wrap;
  font-family: monospace;
}

.error-container button {
  padding: var(--sp-2) var(--sp-3);
  border: 1px solid var(--error-border);
  border-radius: var(--radius-md);
  background: var(--bg);
  color: var(--error-text);
  font-size: var(--fs-md);
  cursor: pointer;
}

.children-container {
  padding-inline-start: 0;
  padding-top: var(--sp-1);
}

.block-wrapper {
  margin-bottom: var(--sp-2);
}

/* ---- Phase 3: Drag and Drop with Nesting ---- */

.block-item {
  cursor: grab;
  user-select: none;
}

.block-item.dragging {
  opacity: 0.3;
}

.block-item.drop-inside {
  border-color: var(--accent);
  box-shadow: 0 0 0 1px var(--accent), inset 0 0 0 1px var(--accent);
}

.drag-overlay {
  position: fixed;
  pointer-events: none;
  opacity: 0.7;
  z-index: 9999;
  will-change: transform;
}

.drop-indicator {
  position: absolute;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--accent);
  pointer-events: none;
  z-index: 100;
  border-radius: 1px;
}

.block-list {
  position: relative;
}

.child-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.25rem;
  height: 1.25rem;
  padding: 0 var(--sp-1);
  border-radius: var(--radius-sm);
  font-size: var(--fs-xs);
  font-weight: 600;
  background: var(--badge-bg);
  color: var(--badge-text);
  flex-shrink: 0;
}

/* ---- Phase 5: Block Type Rendering ---- */

.block-item {
  flex-wrap: wrap;
}

.block-type-content {
  width: 100%;
  margin-top: var(--sp-2);
  order: 99;
}

/* Text block: paragraph placeholder lines */
.block-type-content--text {
  display: flex;
  flex-direction: column;
  gap: var(--sp-1);
}

.block-type-content--text .text-line {
  height: 0.5rem;
  background: var(--border);
  border-radius: 2px;
  opacity: 0.4;
}

.block-type-content--text .text-line:last-child {
  width: 60%;
}

/* Heading block: large bold label */
.block-type-content--heading {
  font-size: 1.125rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: -0.01em;
}

/* Image block: dashed rect placeholder */
.block-type-content--image {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 5rem;
  border: 2px dashed var(--border);
  border-radius: var(--radius-md);
  color: var(--text-muted);
  font-size: var(--fs-sm);
}

/* Container block: labeled box with visible children area */
.block-item--container {
  border-color: var(--accent);
  border-style: dashed;
  background: color-mix(in srgb, var(--accent) 4%, var(--block-bg));
}

.block-type-content--container {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* Type-specific badge colors */
.block-type-badge--heading {
  background: #2a1e45;
  color: #cba6f7;
}

.block-type-badge--image {
  background: #1e3a2a;
  color: #a6e3a1;
}

.block-type-badge--container {
  background: #1e2a45;
  color: #89b4fa;
}

.block-type-badge--text {
  background: #2a2a1e;
  color: #f9e2af;
}

/* ---- Phase 6: Hover Toolbar ---- */

.block-item {
  position: relative;
}

.block-hover-toolbar {
  display: none;
  position: absolute;
  top: var(--sp-1);
  right: var(--sp-1);
  z-index: 50;
  gap: 2px;
  padding: 2px;
  background: var(--toolbar-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.block-item:hover > .block-hover-toolbar,
.block-hover-toolbar:hover {
  display: flex;
}

.block-item.dragging > .block-hover-toolbar {
  display: none;
}

.block-hover-toolbar button {
  padding: 2px var(--sp-2);
  border: none;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  cursor: pointer;
  white-space: nowrap;
  transition: color 0.15s, background 0.15s;
}

.block-hover-toolbar button:hover {
  color: var(--text);
  background: var(--btn-hover-bg);
}

.block-hover-toolbar button[data-action="toolbar-delete"]:hover {
  color: var(--danger);
  background: var(--error-bg);
}

/* ---- Editor Header (save button) ---- */

.editor-header {
  display: flex;
  justify-content: flex-end;
  padding: var(--sp-2) var(--sp-3);
  border-bottom: 1px solid var(--border);
  background: var(--toolbar-bg);
}

/* ---- Insert Buttons ---- */

.insert-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 1.5rem;
  border: 1px dashed var(--border);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-muted);
  font-size: var(--fs-base);
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.15s, border-color 0.15s, color 0.15s;
  margin: var(--sp-1) 0;
}

.block-list:hover .insert-btn,
.insert-btn:focus-visible {
  opacity: 1;
}

.insert-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}

/* Empty state insert button — always visible, large */
.insert-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--sp-8);
}

.insert-btn--empty {
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  border: 2px dashed var(--border);
  font-size: 1.5rem;
  opacity: 1;
}

.insert-btn--empty:hover {
  border-color: var(--accent);
  color: var(--accent);
}

/* Inside-container insert button */
.insert-btn--inside {
  margin: var(--sp-2);
}

/* ---- Insert Dialog ---- */

.insert-dialog-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.insert-dialog-panel {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  padding: var(--sp-4);
  min-width: 280px;
  max-width: 400px;
  max-height: 60vh;
  overflow-y: auto;
}

.insert-dialog-title {
  font-size: var(--fs-base);
  font-weight: 600;
  color: var(--text);
  margin-bottom: var(--sp-3);
}

.insert-dialog-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--sp-2);
  padding: var(--sp-3);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  margin-bottom: var(--sp-2);
  background: transparent;
  color: var(--text);
  cursor: pointer;
  width: 100%;
  text-align: left;
  font-size: var(--fs-base);
  transition: border-color 0.15s, background 0.15s;
}

.insert-dialog-option:hover {
  border-color: var(--accent);
  background: var(--btn-hover-bg);
}

.insert-dialog-option-type {
  font-size: var(--fs-xs);
  color: var(--text-muted);
}

.insert-dialog-cancel {
  display: block;
  width: 100%;
  padding: var(--sp-2);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  font-size: var(--fs-sm);
  margin-top: var(--sp-2);
}

.insert-dialog-cancel:hover {
  color: var(--text);
}

.insert-dialog-empty {
  padding: var(--sp-4);
  text-align: center;
  color: var(--text-muted);
  font-size: var(--fs-sm);
}
`;
