import type { ModulaCMSAdminClient } from "@modulacms/admin-sdk";

/**
 * Theme token contract. Consumers provide CSS custom property values
 * to style all admin-ui components. Every property has a sensible
 * default so a bare `{}` works out of the box.
 */
export interface McmsThemeTokens {
  "--mcms-color-primary"?: string;
  "--mcms-color-primary-hover"?: string;
  "--mcms-color-danger"?: string;
  "--mcms-color-success"?: string;
  "--mcms-color-warning"?: string;
  "--mcms-color-border"?: string;
  "--mcms-color-bg"?: string;
  "--mcms-color-surface"?: string;
  "--mcms-color-text"?: string;
  "--mcms-color-text-dim"?: string;
}

export interface McmsUIConfig {
  /** Admin SDK client instance — components use this for all API calls. */
  client: ModulaCMSAdminClient;

  /** Optional theme tokens applied as CSS custom properties on :root. */
  theme?: McmsThemeTokens;
}

let globalConfig: McmsUIConfig | null = null;

/**
 * Initialize admin-ui components with a configured SDK client.
 * Call once at app startup before rendering any mcms-* elements.
 */
export function configure(config: McmsUIConfig): void {
  globalConfig = config;

  if (config.theme) {
    const root = document.documentElement;
    for (const [prop, value] of Object.entries(config.theme)) {
      if (value !== undefined) {
        root.style.setProperty(prop, value);
      }
    }
  }
}

/**
 * Retrieve the current config. Throws if configure() has not been called.
 */
export function getConfig(): McmsUIConfig {
  if (!globalConfig) {
    throw new Error(
      "@modulacms/admin-ui: configure() must be called before using components."
    );
  }
  return globalConfig;
}
