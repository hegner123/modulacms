import { describe, it, expect } from "vitest";
import { CONTENT_FORMATS } from "@modulacms/types";

describe("CONTENT_FORMATS", () => {
  it("contains all expected formats", () => {
    expect(CONTENT_FORMATS).toEqual([
      "contentful",
      "sanity",
      "strapi",
      "wordpress",
      "clean",
      "raw",
    ]);
  });

  it("has exactly 6 entries", () => {
    expect(CONTENT_FORMATS).toHaveLength(6);
  });
});
