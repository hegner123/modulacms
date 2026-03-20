import { describe, it, expect } from "vitest";
import { validateField, validateForm } from "../utils/validation.js";
import type { FormFieldDefinition, FieldValidationError } from "../types.js";

/** Helper to build a minimal field definition for testing. */
function makeField(overrides: Partial<FormFieldDefinition>): FormFieldDefinition {
  return {
    id: "field-1",
    form_id: "form-1",
    name: "test_field",
    label: "Test Field",
    field_type: "text",
    placeholder: null,
    default_value: null,
    help_text: null,
    required: false,
    validation_rules: null,
    options: null,
    position: 0,
    config: null,
    ...overrides,
  };
}

// ---------------------------------------------------------------------------
// Required check
// ---------------------------------------------------------------------------

describe("required field validation", () => {
  it("fails when required text field is empty string", () => {
    const field = makeField({ required: true, label: "Name" });
    const result = validateField(field, "");
    expect(result).toBe("Name is required");
  });

  it("fails when required text field is whitespace-only", () => {
    const field = makeField({ required: true, label: "Name" });
    const result = validateField(field, "   ");
    expect(result).toBe("Name is required");
  });

  it("passes when required text field has a value", () => {
    const field = makeField({ required: true });
    const result = validateField(field, "hello");
    expect(result).toBeNull();
  });

  it("passes when field is not required and empty", () => {
    const field = makeField({ required: false });
    const result = validateField(field, "");
    expect(result).toBeNull();
  });
});

// ---------------------------------------------------------------------------
// Email validation
// ---------------------------------------------------------------------------

describe("email validation", () => {
  it("passes for valid email a@b.com", () => {
    const field = makeField({ field_type: "email", required: true });
    const result = validateField(field, "a@b.com");
    expect(result).toBeNull();
  });

  it("passes for email with subdomain", () => {
    const field = makeField({ field_type: "email", required: true });
    const result = validateField(field, "user@mail.example.com");
    expect(result).toBeNull();
  });

  it("fails for value without @ sign", () => {
    const field = makeField({ field_type: "email", required: true });
    const result = validateField(field, "no-at");
    expect(result).toBe("Please enter a valid email address");
  });

  it("fails for value with @ but no domain part", () => {
    const field = makeField({ field_type: "email", required: true });
    const result = validateField(field, "user@");
    expect(result).toBe("Please enter a valid email address");
  });
});

// ---------------------------------------------------------------------------
// URL validation
// ---------------------------------------------------------------------------

describe("URL validation", () => {
  it("passes for http:// URL", () => {
    const field = makeField({ field_type: "url", required: true });
    const result = validateField(field, "http://example.com");
    expect(result).toBeNull();
  });

  it("passes for https:// URL", () => {
    const field = makeField({ field_type: "url", required: true });
    const result = validateField(field, "https://example.com/page");
    expect(result).toBeNull();
  });

  it("fails for ftp:// URL", () => {
    const field = makeField({ field_type: "url", required: true });
    const result = validateField(field, "ftp://files.example.com");
    expect(result).toBe("URL must start with http:// or https://");
  });

  it("fails for URL without protocol", () => {
    const field = makeField({ field_type: "url", required: true });
    const result = validateField(field, "example.com");
    expect(result).toBe("URL must start with http:// or https://");
  });
});

// ---------------------------------------------------------------------------
// Number validation
// ---------------------------------------------------------------------------

describe("number validation", () => {
  it("passes for integer string '42'", () => {
    const field = makeField({ field_type: "number", required: true });
    const result = validateField(field, "42");
    expect(result).toBeNull();
  });

  it("passes for decimal string '3.14'", () => {
    const field = makeField({ field_type: "number", required: true });
    const result = validateField(field, "3.14");
    expect(result).toBeNull();
  });

  it("passes for negative number", () => {
    const field = makeField({ field_type: "number", required: true });
    const result = validateField(field, "-7");
    expect(result).toBeNull();
  });

  it("fails for non-numeric string 'abc'", () => {
    const field = makeField({ field_type: "number", required: true });
    const result = validateField(field, "abc");
    expect(result).toBe("Please enter a valid number");
  });
});

// ---------------------------------------------------------------------------
// Select validation
// ---------------------------------------------------------------------------

describe("select validation", () => {
  it("passes when value is in options", () => {
    const field = makeField({
      field_type: "select",
      required: true,
      options: ["red", "green", "blue"],
    });
    const result = validateField(field, "green");
    expect(result).toBeNull();
  });

  it("fails when value is not in options", () => {
    const field = makeField({
      field_type: "select",
      required: true,
      options: ["red", "green", "blue"],
    });
    const result = validateField(field, "yellow");
    expect(result).toBe("Please select a valid option");
  });

  it("passes when options array is empty (no constraint)", () => {
    const field = makeField({
      field_type: "select",
      required: true,
      options: [],
    });
    const result = validateField(field, "anything");
    expect(result).toBeNull();
  });
});

// ---------------------------------------------------------------------------
// Checkbox validation
// ---------------------------------------------------------------------------

describe("checkbox validation", () => {
  it("passes when required checkbox is true", () => {
    const field = makeField({ field_type: "checkbox", required: true, label: "Agree" });
    const result = validateField(field, true);
    expect(result).toBeNull();
  });

  it("fails when required checkbox is false", () => {
    const field = makeField({ field_type: "checkbox", required: true, label: "Agree" });
    const result = validateField(field, false);
    expect(result).toBe("Agree is required");
  });

  it("fails when required checkbox is empty string", () => {
    const field = makeField({ field_type: "checkbox", required: true, label: "Agree" });
    const result = validateField(field, "");
    expect(result).toBe("Agree is required");
  });
});

// ---------------------------------------------------------------------------
// File validation
// ---------------------------------------------------------------------------

describe("file validation", () => {
  it("passes when base64 is within size limit", () => {
    const field = makeField({
      field_type: "file",
      required: true,
      validation_rules: { max_file_size: 1024 },
    });
    // 100 characters of base64 => ~75 bytes, well within 1024
    const result = validateField(field, "a".repeat(100));
    expect(result).toBeNull();
  });

  it("fails when base64 exceeds size limit", () => {
    const field = makeField({
      field_type: "file",
      required: true,
      validation_rules: { max_file_size: 100 },
    });
    // 200 characters of base64 => ~150 bytes, over 100 limit
    const result = validateField(field, "a".repeat(200));
    expect(result).toContain("File exceeds maximum size");
  });
});

// ---------------------------------------------------------------------------
// Length constraints
// ---------------------------------------------------------------------------

describe("min_length / max_length", () => {
  it("fails when value is shorter than min_length", () => {
    const field = makeField({
      required: true,
      validation_rules: { min_length: 5 },
    });
    const result = validateField(field, "abc");
    expect(result).toBe("Must be at least 5 characters");
  });

  it("passes at exactly min_length boundary", () => {
    const field = makeField({
      required: true,
      validation_rules: { min_length: 3 },
    });
    const result = validateField(field, "abc");
    expect(result).toBeNull();
  });

  it("fails when value exceeds max_length", () => {
    const field = makeField({
      required: true,
      validation_rules: { max_length: 5 },
    });
    const result = validateField(field, "abcdef");
    expect(result).toBe("Must be no more than 5 characters");
  });

  it("passes at exactly max_length boundary", () => {
    const field = makeField({
      required: true,
      validation_rules: { max_length: 5 },
    });
    const result = validateField(field, "abcde");
    expect(result).toBeNull();
  });
});

// ---------------------------------------------------------------------------
// validateForm (multi-field)
// ---------------------------------------------------------------------------

describe("validateForm", () => {
  it("returns empty array when all fields are valid", () => {
    const fields = [
      makeField({ name: "name", required: true }),
      makeField({ name: "email", field_type: "email", required: true }),
    ];
    const data = { name: "Alice", email: "alice@example.com" };

    const errors = validateForm(fields, data);
    expect(errors).toEqual([]);
  });

  it("returns errors for multiple invalid fields", () => {
    const fields = [
      makeField({ name: "name", label: "Name", required: true }),
      makeField({ name: "email", label: "Email", field_type: "email", required: true }),
      makeField({ name: "website", label: "Website", field_type: "url", required: true }),
    ];
    const data = { name: "", email: "bad", website: "no-protocol" };

    const errors = validateForm(fields, data);
    expect(errors).toHaveLength(3);
    expect(errors[0]).toEqual({ field: "name", message: "Name is required" });
    expect(errors[1]).toEqual({ field: "email", message: "Please enter a valid email address" });
    expect(errors[2]).toEqual({ field: "website", message: "URL must start with http:// or https://" });
  });

  it("validates only present fields and skips optional empty ones", () => {
    const fields = [
      makeField({ name: "name", required: true }),
      makeField({ name: "bio", required: false }),
    ];
    const data = { name: "Bob" };

    const errors = validateForm(fields, data);
    expect(errors).toEqual([]);
  });
});
