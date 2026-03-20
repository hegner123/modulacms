export const builderStyles = `
  /* --- Builder layout (two-column: palette + canvas) --- */
  .mcms-builder {
    display: grid;
    grid-template-columns: 14rem 1fr;
    grid-template-rows: auto 1fr;
    gap: 1rem;
    min-height: 24rem;
  }

  /* --- Toolbar (spans full width) --- */
  .mcms-builder-toolbar {
    grid-column: 1 / -1;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: #f9fafb;
  }

  .mcms-builder-toolbar-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--modulacms-text-color, #111827);
  }

  .mcms-builder-toolbar-actions {
    display: flex;
    gap: 0.5rem;
  }

  .mcms-builder-btn {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    color: var(--modulacms-text-color, #111827);
    font-family: inherit;
    font-size: 0.8125rem;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .mcms-builder-btn:hover {
    background: #f3f4f6;
  }

  .mcms-builder-btn:focus-visible {
    outline: 2px solid var(--modulacms-primary-color, #2563eb);
    outline-offset: 2px;
  }

  .mcms-builder-btn--primary {
    background: var(--modulacms-button-bg, var(--modulacms-primary-color, #2563eb));
    color: var(--modulacms-button-color, #ffffff);
    border-color: transparent;
  }

  .mcms-builder-btn--primary:hover {
    opacity: 0.9;
    background: var(--modulacms-button-bg, var(--modulacms-primary-color, #2563eb));
  }

  /* --- Field palette (left sidebar) --- */
  .mcms-palette {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: #f9fafb;
    overflow-y: auto;
  }

  .mcms-palette-title {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #6b7280;
    margin: 0 0 0.25rem;
  }

  .mcms-palette-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.375rem;
  }

  .mcms-palette-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    padding: 0.5rem 0.25rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    font-family: inherit;
    font-size: 0.6875rem;
    color: var(--modulacms-text-color, #111827);
    cursor: grab;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
    user-select: none;
  }

  .mcms-palette-item:hover {
    border-color: var(--modulacms-primary-color, #2563eb);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--modulacms-primary-color, #2563eb) 25%, transparent);
  }

  .mcms-palette-item:active {
    cursor: grabbing;
  }

  .mcms-palette-icon {
    font-size: 1.125rem;
    line-height: 1;
  }

  /* --- Canvas (drop zone) --- */
  .mcms-canvas {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 1rem;
    border: 2px dashed var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    min-height: 12rem;
    transition: border-color 0.15s ease, background 0.15s ease;
    overflow-y: auto;
  }

  .mcms-canvas--empty {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #9ca3af;
    font-size: 0.875rem;
  }

  .mcms-canvas--dragover {
    border-color: var(--modulacms-primary-color, #2563eb);
    background: color-mix(in srgb, var(--modulacms-primary-color, #2563eb) 4%, transparent);
  }

  /* --- Field items (draggable cards on canvas) --- */
  .mcms-field-card {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    padding: 0.625rem 0.75rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    cursor: default;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
  }

  .mcms-field-card:hover {
    border-color: #9ca3af;
  }

  .mcms-field-card.mcms-field-card--selected {
    border-color: var(--modulacms-primary-color, #2563eb);
    box-shadow: 0 0 0 1px var(--modulacms-primary-color, #2563eb);
  }

  /* --- Drag handle --- */
  .mcms-drag-handle {
    display: flex;
    flex-direction: column;
    gap: 2px;
    cursor: grab;
    padding: 0.25rem;
    color: #9ca3af;
    flex-shrink: 0;
  }

  .mcms-drag-handle:active {
    cursor: grabbing;
  }

  .mcms-drag-handle-dot {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: currentColor;
  }

  .mcms-drag-handle-row {
    display: flex;
    gap: 2px;
  }

  /* --- Field card content --- */
  .mcms-field-card-body {
    flex: 1;
    min-width: 0;
  }

  .mcms-field-card-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--modulacms-text-color, #111827);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .mcms-field-card-type {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .mcms-field-card-actions {
    display: flex;
    gap: 0.25rem;
    flex-shrink: 0;
  }

  .mcms-field-card-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.75rem;
    height: 1.75rem;
    border: none;
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: transparent;
    color: #6b7280;
    font-size: 0.8125rem;
    cursor: pointer;
    transition: background 0.1s ease, color 0.1s ease;
  }

  .mcms-field-card-btn:hover {
    background: #f3f4f6;
    color: var(--modulacms-text-color, #111827);
  }

  .mcms-field-card-btn--delete:hover {
    background: color-mix(in srgb, var(--modulacms-error-color, #dc2626) 10%, transparent);
    color: var(--modulacms-error-color, #dc2626);
  }

  /* --- Field config panel (right side or overlay) --- */
  .mcms-field-config {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 1rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: #f9fafb;
  }

  .mcms-field-config-title {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--modulacms-text-color, #111827);
    margin: 0;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .mcms-field-config-row {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .mcms-field-config-label {
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--modulacms-text-color, #111827);
  }

  .mcms-field-config-input {
    padding: 0.375rem 0.625rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    color: var(--modulacms-text-color, #111827);
    font-family: inherit;
    font-size: 0.8125rem;
  }

  .mcms-field-config-input:focus {
    outline: none;
    border-color: var(--modulacms-primary-color, #2563eb);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--modulacms-primary-color, #2563eb) 25%, transparent);
  }

  .mcms-field-config-checkbox {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8125rem;
  }

  .mcms-field-config-checkbox input {
    accent-color: var(--modulacms-primary-color, #2563eb);
  }

  /* --- Drag states --- */
  .mcms-field-card.dragging {
    opacity: 0.4;
  }

  .mcms-drop-indicator {
    height: 2px;
    background: var(--modulacms-primary-color, #2563eb);
    border-radius: 1px;
    margin: -0.25rem 0;
    pointer-events: none;
  }

  .mcms-drag-overlay {
    position: fixed;
    top: 0;
    left: 0;
    pointer-events: none;
    z-index: 9999;
    opacity: 0.85;
  }

  .mcms-drag-overlay .mcms-field-card {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    transform: rotate(1.5deg);
  }
`;
