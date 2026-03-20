import type { FormFieldDefinition, FieldValidationError } from "../types.js";

/** Default max file size in bytes when not specified (750 KB base64-estimated). */
const DEFAULT_MAX_FILE_SIZE = 768000;

/**
 * Validate a single field value against its definition.
 * Returns an error message string, or null when valid.
 */
export function validateField(
  field: FormFieldDefinition,
  value: unknown,
): string | null {
  const strValue = value === null || value === undefined ? "" : String(value);
  const isEmpty = strValue.trim() === "";

  // Required check
  if (field.required) {
    if (field.field_type === "checkbox") {
      if (!value) {
        return `${field.label} is required`;
      }
    } else if (isEmpty) {
      return `${field.label} is required`;
    }
  }

  // Skip further validation when the value is empty and not required
  if (isEmpty && field.field_type !== "checkbox") {
    return null;
  }

  // Type-specific checks
  switch (field.field_type) {
    case "email": {
      const emailPattern = /^[^@]+@[^@]+\.[^@]+$/;
      if (!emailPattern.test(strValue)) {
        return "Please enter a valid email address";
      }
      break;
    }

    case "url": {
      if (
        !strValue.startsWith("http://") &&
        !strValue.startsWith("https://")
      ) {
        return "URL must start with http:// or https://";
      }
      break;
    }

    case "number": {
      const num = Number(strValue);
      if (!Number.isFinite(num)) {
        return "Please enter a valid number";
      }
      break;
    }

    case "select":
    case "radio": {
      if (field.options && field.options.length > 0) {
        if (!field.options.includes(strValue)) {
          return "Please select a valid option";
        }
      }
      break;
    }

    case "file": {
      // Base64 strings encode 3 bytes per 4 characters.
      // Estimate original file size from the base64-encoded length.
      const maxSize =
        field.validation_rules?.max_file_size ?? DEFAULT_MAX_FILE_SIZE;
      const estimatedSize = Math.ceil((strValue.length * 3) / 4);
      if (estimatedSize > maxSize) {
        const maxKB = Math.round(maxSize / 1024);
        return `File exceeds maximum size of ${maxKB} KB`;
      }
      break;
    }
  }

  // Length constraints from validation_rules
  if (field.validation_rules) {
    const { min_length, max_length } = field.validation_rules;

    if (min_length !== undefined && strValue.length < min_length) {
      return `Must be at least ${min_length} characters`;
    }
    if (max_length !== undefined && strValue.length > max_length) {
      return `Must be no more than ${max_length} characters`;
    }
  }

  return null;
}

/**
 * Validate all fields in a form. Returns an array of errors (empty if valid).
 */
export function validateForm(
  fields: FormFieldDefinition[],
  data: Record<string, unknown>,
): FieldValidationError[] {
  const errors: FieldValidationError[] = [];

  for (const field of fields) {
    const value = data[field.name];
    const message = validateField(field, value);
    if (message !== null) {
      errors.push({ field: field.name, message });
    }
  }

  return errors;
}
