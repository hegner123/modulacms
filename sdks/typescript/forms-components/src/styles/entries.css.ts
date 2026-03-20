export const entriesStyles = `
  /* --- Container --- */
  .mcms-entries {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  /* --- Toolbar (filter + export) --- */
  .mcms-entries-toolbar {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .mcms-filter-input {
    flex: 1 1 12rem;
    min-width: 10rem;
    padding: var(--modulacms-input-padding, 0.5rem 0.75rem);
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    color: var(--modulacms-text-color, #111827);
    font-family: inherit;
    font-size: 0.875rem;
  }

  .mcms-filter-input:focus {
    outline: none;
    border-color: var(--modulacms-primary-color, #2563eb);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--modulacms-primary-color, #2563eb) 25%, transparent);
  }

  .mcms-export-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.4375rem 0.875rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    color: var(--modulacms-text-color, #111827);
    font-family: inherit;
    font-size: 0.8125rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .mcms-export-btn:hover {
    background: #f3f4f6;
  }

  .mcms-export-btn:focus-visible {
    outline: 2px solid var(--modulacms-primary-color, #2563eb);
    outline-offset: 2px;
  }

  /* --- Table --- */
  .mcms-entries-table-wrap {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .mcms-entries-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }

  .mcms-entries-table thead {
    border-bottom: 2px solid var(--modulacms-border-color, #d1d5db);
  }

  .mcms-entries-table th {
    padding: 0.625rem 0.75rem;
    text-align: left;
    font-weight: 600;
    color: var(--modulacms-text-color, #111827);
    white-space: nowrap;
    user-select: none;
  }

  .mcms-entries-table th[data-sortable] {
    cursor: pointer;
  }

  .mcms-entries-table th[data-sortable]:hover {
    color: var(--modulacms-primary-color, #2563eb);
  }

  .mcms-entries-table th[data-sortable]:focus-visible {
    outline: 2px solid var(--modulacms-primary-color, #2563eb);
    outline-offset: -2px;
  }

  .mcms-entries-table tbody tr {
    border-bottom: 1px solid #e5e7eb;
    transition: background 0.1s ease;
  }

  .mcms-entries-table tbody tr:hover {
    background: #f9fafb;
  }

  .mcms-entries-table tbody tr:focus-visible {
    outline: 2px solid var(--modulacms-primary-color, #2563eb);
    outline-offset: -2px;
    background: #eff6ff;
  }

  .mcms-entries-table tbody tr.selected {
    background: #eff6ff;
  }

  .mcms-entries-table tbody tr[tabindex] {
    cursor: pointer;
  }

  .mcms-entries-table td {
    padding: 0.5rem 0.75rem;
    color: var(--modulacms-text-color, #111827);
    vertical-align: top;
    max-width: 20rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* --- Pagination --- */
  .mcms-pagination {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    font-size: 0.8125rem;
    color: #6b7280;
  }

  .mcms-pagination-info {
    white-space: nowrap;
  }

  .mcms-pagination-controls {
    display: flex;
    align-items: center;
    gap: 0.375rem;
  }

  .mcms-page-btn {
    padding: 0.3125rem 0.625rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    color: var(--modulacms-text-color, #111827);
    font-family: inherit;
    font-size: 0.8125rem;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .mcms-page-btn:hover:not([disabled]) {
    background: #f3f4f6;
  }

  .mcms-page-btn[disabled] {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .mcms-page-btn:focus-visible {
    outline: 2px solid var(--modulacms-primary-color, #2563eb);
    outline-offset: 2px;
  }

  .mcms-page-btn[aria-current="page"] {
    background: var(--modulacms-primary-color, #2563eb);
    color: var(--modulacms-button-color, #ffffff);
    border-color: var(--modulacms-primary-color, #2563eb);
  }

  .mcms-pagination-ellipsis {
    padding: 0.3125rem 0.25rem;
    color: #9ca3af;
    font-size: 0.8125rem;
  }

  /* --- Status badges --- */
  .mcms-status-badge {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: 9999px;
    font-size: 0.75rem;
    font-weight: 500;
    line-height: 1.25rem;
  }

  .mcms-status-pending {
    background: #fef3c7;
    color: #92400e;
  }

  .mcms-status-reviewed {
    background: #dbeafe;
    color: #1e40af;
  }

  .mcms-status-spam {
    background: #fee2e2;
    color: #991b1b;
  }

  .mcms-status-default {
    background: #f3f4f6;
    color: #374151;
  }

  /* --- Empty state --- */
  .mcms-entries-empty {
    padding: 2.5rem 1rem;
    text-align: center;
    color: #9ca3af;
    font-size: 0.9375rem;
  }

  /* --- Loading state --- */
  .mcms-entries-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2.5rem 1rem;
    color: #9ca3af;
    gap: 0.5rem;
    font-size: 0.875rem;
  }

  .mcms-entries-loading .mcms-spinner {
    display: inline-block;
    width: 1.25rem;
    height: 1.25rem;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: mcms-entries-spin 0.6s linear infinite;
  }

  @keyframes mcms-entries-spin {
    to { transform: rotate(360deg); }
  }
`;
