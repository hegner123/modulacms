import { describe, it, expect } from "vitest";
import { ModulaError } from "../src/errors.js";

describe("ModulaError", () => {
  it("extracts message from body.error", () => {
    const err = new ModulaError(404, { error: "Not found" });

    expect(err.message).toBe("Not found");
    expect(err.status).toBe(404);
    expect(err.name).toBe("ModulaError");
    expect(err.body).toEqual({ error: "Not found" });
  });

  it("falls back to status-based message when body has no error property", () => {
    const err = new ModulaError(500, "Internal server error");

    expect(err.message).toBe("Request failed with status 500");
    expect(err.body).toBe("Internal server error");
  });

  it("falls back to status-based message for null body", () => {
    const err = new ModulaError(502, null);

    expect(err.message).toBe("Request failed with status 502");
    expect(err.body).toBeNull();
  });

  it("is an instance of Error", () => {
    const err = new ModulaError(400, {});

    expect(err).toBeInstanceOf(Error);
    expect(err).toBeInstanceOf(ModulaError);
  });

  it("has a stack trace", () => {
    const err = new ModulaError(400, {});

    expect(err.stack).toBeDefined();
  });

  describe("errorMessage getter", () => {
    it("returns the error string from body", () => {
      const err = new ModulaError(404, { error: "Page not found" });

      expect(err.errorMessage).toBe("Page not found");
    });

    it("returns undefined when body is a string", () => {
      const err = new ModulaError(500, "plain text");

      expect(err.errorMessage).toBeUndefined();
    });

    it("returns undefined when body is null", () => {
      const err = new ModulaError(500, null);

      expect(err.errorMessage).toBeUndefined();
    });

    it("returns undefined when body has no error property", () => {
      const err = new ModulaError(500, { message: "something" });

      expect(err.errorMessage).toBeUndefined();
    });

    it("coerces non-string error property to string", () => {
      const err = new ModulaError(500, { error: 42 });

      expect(err.errorMessage).toBe("42");
    });
  });

  describe("readonly fields", () => {
    it("status and body are readonly", () => {
      const err = new ModulaError(404, { error: "test" });

      // TypeScript enforces readonly at compile time; verify values are stable
      expect(err.status).toBe(404);
      expect(err.body).toEqual({ error: "test" });
    });
  });
});
