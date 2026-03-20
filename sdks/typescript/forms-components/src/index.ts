// Re-export types
export type {
  FormDefinition,
  FormFieldDefinition,
  FormEntry,
  FormWebhook,
  FieldType,
  ValidationRules,
  CaptchaConfig,
  PaginatedResponse,
  ExportResponse,
  SubmitResponse,
  QueueInfo,
  FieldValidationError,
  ApiError,
} from "./types.js";

// Re-export API client
export { FormsApiClient } from "./api-client.js";

// Import components for registration and re-export
import { ModulaCMSForm } from "./components/modulacms-form.js";
import { ModulaCMSEntries } from "./components/modulacms-entries.js";
import { ModulaCMSFormBuilder } from "./components/modulacms-form-builder.js";

export { ModulaCMSForm, ModulaCMSEntries, ModulaCMSFormBuilder };

// Re-export utilities
export { validateField, validateForm } from "./utils/validation.js";
export { initDragAndDrop } from "./utils/drag.js";

// Auto-register custom elements
if (typeof customElements !== "undefined") {
  const define = (name: string, ctor: CustomElementConstructor) => {
    if (!customElements.get(name)) {
      customElements.define(name, ctor);
    }
  };
  define("modulacms-form", ModulaCMSForm);
  define("modulacms-entries", ModulaCMSEntries);
  define("modulacms-form-builder", ModulaCMSFormBuilder);
}
