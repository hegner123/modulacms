export { configure, getConfig } from "./config.js";
export type { McmsUIConfig, McmsThemeTokens } from "./config.js";

export {
  registerFieldRenderer,
  registerFieldRenderers,
  getFieldRenderer,
  hasFieldRenderer,
  unregisterFieldRenderer,
  registeredFieldTypes,
} from "./registry.js";
export type {
  FieldRenderContext,
  McmsFieldRenderer,
  McmsFieldRendererConstructor,
  FieldRendererEntry,
} from "./registry.js";
