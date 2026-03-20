export const baseStyles = `
  *,
  *::before,
  *::after {
    box-sizing: border-box;
  }

  :host {
    display: block;
    font-family: var(--modulacms-font-family, system-ui, -apple-system, sans-serif);
    color: var(--modulacms-text-color, #111827);
    line-height: 1.5;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  :host([hidden]) {
    display: none;
  }
`;
