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
      "russia",
      "iran",
    ]);
  });
  test("Russian proxy exceptions precede the direct category", () => {
    const rules = presets.find((p) => p.key === "russia")!.code;
    expect(rules.indexOf("domain: abook-club.ru")).toBeGreaterThan(0);
    expect(rules.indexOf("domain: abook-club.ru")).toBeLessThan(
      rules.indexOf("geosite:category-ru"),
    );
    expect(rules).toContain("default: proxy");
  });
  test("the Iran rule set follows Russia in the menu and keeps the LAN direct", () => {
    const keys = presets.map((p) => p.key);
    expect(keys.indexOf("iran")).toBe(keys.indexOf("russia") + 1);
    const rules = presets.find((p) => p.key === "iran")!.code;
    expect(rules.startsWith("default: proxy\n")).toBe(true);
    expect(rules).toContain("domain(geosite:private) -> direct");
    expect(rules).toContain("ip(geoip:private) -> direct");
    expect(rules).toContain("domain(geosite:category-ir) -> direct");
    expect(rules).toContain("ip(geoip:ir) -> direct");
  });
});
