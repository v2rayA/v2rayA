import { describe, expect, test } from "vitest";
import { check } from "./check";
import { presets } from "./presets";

describe("the rule templates", () => {
  test("every template passes the editor's own check", () => {
    for (const preset of presets) {
      const text = preset.code.startsWith("default:")
        ? preset.code
        : `default: proxy\n${preset.code}`;
      expect(check(text), preset.key).toEqual([]);
    }
  });
  test("only the full rule sets carry a default outbound", () => {
    const full = presets.filter((p) => p.code.startsWith("default:"));
    expect(full.map((p) => p.key)).toEqual([
      "whitelist",
      "blacklist",
      "global",
      "minimal",
    ]);
  });
});
