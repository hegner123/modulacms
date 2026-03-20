export const formStyles = `
  /* --- Layout --- */
  .mcms-form {
    display: flex;
    flex-direction: column;
    gap: var(--modulacms-field-gap, 1rem);
  }

  /* --- Field container --- */
  .mcms-field {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  /* --- Labels --- */
  .mcms-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--modulacms-text-color, #111827);
  }

  .mcms-label .mcms-required {
    color: var(--modulacms-error-color, #dc2626);
    margin-left: 0.125rem;
  }

  /* --- Inputs, selects, textareas --- */
  .mcms-input,
  .mcms-select,
  .mcms-textarea {
    width: 100%;
    padding: var(--modulacms-input-padding, 0.5rem 0.75rem);
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    color: var(--modulacms-text-color, #111827);
    font-family: inherit;
    font-size: 0.9375rem;
    line-height: 1.5;
    transition: border-color 0.15s ease, box-shadow 0.15s ease;
  }

  .mcms-input:focus,
  .mcms-select:focus,
  .mcms-textarea:focus {
    outline: none;
    border-color: var(--modulacms-primary-color, #2563eb);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--modulacms-primary-color, #2563eb) 25%, transparent);
  }

  .mcms-input[aria-invalid="true"],
  .mcms-select[aria-invalid="true"],
  .mcms-textarea[aria-invalid="true"] {
    border-color: var(--modulacms-error-color, #dc2626);
  }

  .mcms-textarea {
    min-height: 5rem;
    resize: vertical;
  }

  .mcms-select {
    appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%236b7280' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
    background-repeat: no-repeat;
    background-position: right 0.75rem center;
    padding-right: 2rem;
  }

  /* --- Checkbox / Radio rows --- */
  .mcms-check-group {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
    padding-top: 0.125rem;
  }

  .mcms-check-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
  }

  .mcms-check-item input {
    margin: 0;
    accent-color: var(--modulacms-primary-color, #2563eb);
  }

  /* --- Help text --- */
  .mcms-help {
    font-size: 0.8125rem;
    color: #6b7280;
    margin: 0;
  }

  /* --- Error messages --- */
  .mcms-error {
    font-size: 0.8125rem;
    color: var(--modulacms-error-color, #dc2626);
    margin: 0;
    min-height: 1.25rem;
  }

  .mcms-form-error {
    padding: 0.75rem 1rem;
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: color-mix(in srgb, var(--modulacms-error-color, #dc2626) 8%, transparent);
    color: var(--modulacms-error-color, #dc2626);
    font-size: 0.875rem;
  }

  /* --- Submit button --- */
  .mcms-submit {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: var(--modulacms-button-padding, 0.625rem 1.25rem);
    border: none;
    border-radius: var(--modulacms-button-radius, 0.375rem);
    background: var(--modulacms-button-bg, var(--modulacms-primary-color, #2563eb));
    color: var(--modulacms-button-color, #ffffff);
    font-family: inherit;
    font-size: 0.9375rem;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.15s ease;
    align-self: flex-start;
  }

  .mcms-submit:hover {
    opacity: 0.9;
  }

  .mcms-submit:active {
    opacity: 0.8;
  }

  .mcms-submit:focus-visible {
    outline: 2px solid var(--modulacms-primary-color, #2563eb);
    outline-offset: 2px;
  }

  /* --- Loading state --- */
  .mcms-submit[disabled] {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .mcms-spinner {
    display: inline-block;
    width: 1em;
    height: 1em;
    border: 2px solid currentColor;
    border-right-color: transparent;
    border-radius: 50%;
    animation: mcms-spin 0.6s linear infinite;
  }

  @keyframes mcms-spin {
    to { transform: rotate(360deg); }
  }

  /* --- Success state --- */
  .mcms-success {
    padding: 1rem;
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: color-mix(in srgb, var(--modulacms-success-color, #16a34a) 8%, transparent);
    color: var(--modulacms-success-color, #16a34a);
    font-size: 0.9375rem;
    text-align: center;
  }

  /* --- Honeypot (anti-spam hidden field) --- */
  .mcms-honeypot {
    position: absolute;
    left: -9999px;
    top: -9999px;
    width: 1px;
    height: 1px;
    overflow: hidden;
    opacity: 0;
    pointer-events: none;
    tab-index: -1;
  }

  /* --- File input --- */
  .mcms-file-input {
    font-size: 0.875rem;
  }

  .mcms-file-input::file-selector-button {
    padding: 0.375rem 0.75rem;
    border: 1px solid var(--modulacms-border-color, #d1d5db);
    border-radius: var(--modulacms-border-radius, 0.375rem);
    background: var(--modulacms-bg-color, #ffffff);
    color: var(--modulacms-text-color, #111827);
    font-family: inherit;
    font-size: 0.8125rem;
    cursor: pointer;
    margin-right: 0.5rem;
  }
`;
