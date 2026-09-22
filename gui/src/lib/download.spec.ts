// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { exportName } from "./download";

describe("exportName", () => {
  test("names an export after the moment it was taken", () => {
    expect(exportName(new Date(2026, 8, 22, 8, 49, 9))).toBe(
      "export-2026-09-22_08-49-09.txt",
    );
  });
  test("keeps the name to characters every file system accepts", () => {
    expect(exportName()).toMatch(
      /^export-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.txt$/,
    );
  });
});
