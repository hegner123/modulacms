import { describe, test, expect } from "vitest";
import { isApiError } from "./common.js";
import type { ApiError } from "./common.js";

// ---------------------------------------------------------------------------
// isApiError — runtime type guard
//
// The basic cases (null, undefined, Error instance, missing _tag, wrong _tag,
// valid ApiError, valid ApiError with body) are already covered in
// src/index.test.ts via the re-export. These tests focus on edge cases and
// gaps not covered there.
// ---------------------------------------------------------------------------

describe("isApiError", () => {
  // -------------------------------------------------------------------------
  // Primitive inputs — none of these are objects, so all must return false.
  // The existing tests cover null and undefined; these cover the rest.
  // -------------------------------------------------------------------------

  describe("rejects non-object primitives", () => {
    test.each([
      { name: "number zero", value: 0 },
      { name: "positive number", value: 42 },
      { name: "negative number", value: -1 },
      { name: "NaN", value: NaN },
      { name: "Infinity", value: Infinity },
      { name: "empty string", value: "" },
      { name: "non-empty string", value: "ApiError" },
      // The string "ApiError" is interesting — it's the tag value itself
      { name: "boolean true", value: true },
      { name: "boolean false", value: false },
      { name: "bigint", value: BigInt(0) },
      { name: "symbol", value: Symbol("ApiError") },
    ])("returns false for $name", ({ value }) => {
      expect(isApiError(value)).toBe(false);
    });
  });

  // -------------------------------------------------------------------------
  // Callable/function — typeof function !== "object"
  // -------------------------------------------------------------------------

  test("returns false for a function", () => {
    const fn = () => "ApiError";
    expect(isApiError(fn)).toBe(false);
  });

  test("returns false for a function with _tag property", () => {
    const fn = () => {};
    // Functions can have arbitrary properties attached
    (fn as unknown as Record<string, unknown>)._tag = "ApiError";
    expect(isApiError(fn)).toBe(false);
  });

  // -------------------------------------------------------------------------
  // Array inputs — typeof [] === "object" and [] !== null, so the guard
  // must reject them via the _tag check
  // -------------------------------------------------------------------------

  test("returns false for an empty array", () => {
    expect(isApiError([])).toBe(false);
  });

  test("returns false for an array with _tag element", () => {
    // Arrays are objects, but "_tag" in [...] checks for a property named "_tag"
    const arr = ["ApiError"];
    expect(isApiError(arr)).toBe(false);
  });

  // -------------------------------------------------------------------------
  // Objects with _tag set to non-string types — the guard checks
  // (err as ApiError)._tag === "ApiError" which is strict equality, so
  // non-string _tag values must fail
  // -------------------------------------------------------------------------

  describe("rejects objects where _tag is not the string 'ApiError'", () => {
    test.each([
      { name: "_tag is null", obj: { _tag: null, status: 500, message: "fail" } },
      { name: "_tag is undefined", obj: { _tag: undefined, status: 500, message: "fail" } },
      { name: "_tag is a number", obj: { _tag: 123, status: 500, message: "fail" } },
      { name: "_tag is a boolean", obj: { _tag: true, status: 500, message: "fail" } },
      { name: "_tag is an empty string", obj: { _tag: "", status: 500, message: "fail" } },
      { name: "_tag is an object", obj: { _tag: { value: "ApiError" }, status: 500, message: "fail" } },
      { name: "_tag is an array", obj: { _tag: ["ApiError"], status: 500, message: "fail" } },
      // Close but not exact — case sensitivity
      { name: "_tag is 'apierror' (lowercase)", obj: { _tag: "apierror", status: 500, message: "fail" } },
      { name: "_tag is 'APIERROR' (uppercase)", obj: { _tag: "APIERROR", status: 500, message: "fail" } },
      { name: "_tag is 'ApiError ' (trailing space)", obj: { _tag: "ApiError ", status: 500, message: "fail" } },
      { name: "_tag is ' ApiError' (leading space)", obj: { _tag: " ApiError", status: 500, message: "fail" } },
    ])("returns false when $name", ({ obj }) => {
      expect(isApiError(obj)).toBe(false);
    });
  });

  // -------------------------------------------------------------------------
  // Minimal valid object — the guard only checks _tag, not status or message.
  // An object with just { _tag: "ApiError" } should pass the guard even
  // though it doesn't satisfy the full ApiError type at the TS level.
  // This matters because the guard is used on `unknown` values at runtime.
  // -------------------------------------------------------------------------

  test("returns true for an object with only _tag: 'ApiError' (no other fields)", () => {
    const minimal = { _tag: "ApiError" as const };
    expect(isApiError(minimal)).toBe(true);
  });

  // -------------------------------------------------------------------------
  // Prototype-inherited _tag — the `in` operator checks the prototype chain.
  // If _tag is inherited, the guard should still detect it.
  // -------------------------------------------------------------------------

  test("returns true when _tag is inherited from prototype", () => {
    const proto = { _tag: "ApiError" as const };
    const obj = Object.create(proto);
    obj.status = 500;
    obj.message = "inherited tag";
    expect(isApiError(obj)).toBe(true);
  });

  test("returns false when inherited _tag has wrong value", () => {
    const proto = { _tag: "OtherError" };
    const obj = Object.create(proto);
    obj.status = 500;
    obj.message = "wrong inherited tag";
    expect(isApiError(obj)).toBe(false);
  });

  // -------------------------------------------------------------------------
  // Object.create(null) — objects with no prototype. The `in` operator
  // still works on these, but they lack Object.prototype methods.
  // -------------------------------------------------------------------------

  test("returns true for Object.create(null) with _tag: 'ApiError'", () => {
    const obj = Object.create(null);
    obj._tag = "ApiError";
    obj.status = 500;
    obj.message = "no prototype";
    expect(isApiError(obj)).toBe(true);
  });

  test("returns false for Object.create(null) without _tag", () => {
    const obj = Object.create(null);
    obj.status = 500;
    obj.message = "no prototype";
    expect(isApiError(obj)).toBe(false);
  });

  // -------------------------------------------------------------------------
  // Frozen and sealed objects — property access should still work
  // -------------------------------------------------------------------------

  test("returns true for a frozen ApiError object", () => {
    const err = Object.freeze({
      _tag: "ApiError" as const,
      status: 403,
      message: "Forbidden",
    });
    expect(isApiError(err)).toBe(true);
  });

  test("returns true for a sealed ApiError object", () => {
    const err = Object.seal({
      _tag: "ApiError" as const,
      status: 403,
      message: "Forbidden",
    });
    expect(isApiError(err)).toBe(true);
  });

  // -------------------------------------------------------------------------
  // Extra/unexpected properties — the guard should not reject objects that
  // have additional properties beyond what ApiError defines
  // -------------------------------------------------------------------------

  test("returns true for ApiError with extra properties", () => {
    const err = {
      _tag: "ApiError" as const,
      status: 422,
      message: "Unprocessable Entity",
      body: { errors: ["field required"] },
      requestId: "abc-123",
      timestamp: Date.now(),
    };
    expect(isApiError(err)).toBe(true);
  });

  // -------------------------------------------------------------------------
  // Class instances — the guard should work with class instances that
  // have the right shape, not just plain objects
  // -------------------------------------------------------------------------

  test("returns true for a class instance with _tag: 'ApiError'", () => {
    class CustomError {
      readonly _tag = "ApiError" as const;
      status = 500;
      message = "Internal Server Error";
    }
    const err = new CustomError();
    expect(isApiError(err)).toBe(true);
  });

  test("returns false for a class instance with wrong _tag", () => {
    class CustomError {
      readonly _tag = "CustomError" as const;
      status = 500;
      message = "Internal Server Error";
    }
    const err = new CustomError();
    expect(isApiError(err)).toBe(false);
  });

  // -------------------------------------------------------------------------
  // Date, RegExp, Map, Set — built-in objects that are typeof "object"
  // and not null. None should pass.
  // -------------------------------------------------------------------------

  describe("rejects built-in object types", () => {
    test.each([
      { name: "Date", value: new Date() },
      { name: "Map", value: new Map() },
      { name: "Set", value: new Set() },
      { name: "WeakMap", value: new WeakMap() },
      { name: "WeakSet", value: new WeakSet() },
      { name: "ArrayBuffer", value: new ArrayBuffer(8) },
      { name: "Uint8Array", value: new Uint8Array(8) },
      { name: "Promise", value: Promise.resolve() },
    ])("returns false for $name", ({ value }) => {
      expect(isApiError(value)).toBe(false);
    });
  });

  // -------------------------------------------------------------------------
  // Type narrowing verification — after isApiError returns true, the
  // narrowed type should allow accessing ApiError fields without assertion
  // -------------------------------------------------------------------------

  test("narrows type to ApiError so fields are accessible", () => {
    const unknown: unknown = {
      _tag: "ApiError",
      status: 404,
      message: "Not Found",
      body: { detail: "missing resource" },
    };

    if (isApiError(unknown)) {
      // These accesses compile without type assertions because the guard
      // narrows `unknown` to `ApiError`
      expect(unknown._tag).toBe("ApiError");
      expect(unknown.status).toBe(404);
      expect(unknown.message).toBe("Not Found");
      expect(unknown.body).toEqual({ detail: "missing resource" });
    } else {
      // Force failure if the guard incorrectly returns false
      expect(true).toBe(false);
    }
  });
});
