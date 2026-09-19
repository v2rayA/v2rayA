import { describe, expect, test } from "vitest";
import { tokenize } from "./highlight";

describe("RoutingA highlighting", () => {
  test("classifies syntax without losing any input", () => {
    const text = 'domain(regexp: "a#(b)") && ip(10.0.0.0/8) -> custom-2 # end';
    const tokens = tokenize(text);
    expect(tokens.map((token) => token.text).join("")).toBe(text);
    expect(tokens.filter((token) => token.type !== "text")).toEqual([
      { type: "function", text: "domain" },
      { type: "argument", text: "regexp:" },
      { type: "string", text: '"a#(b)"' },
      { type: "operator", text: "&&" },
      { type: "function", text: "ip" },
      { type: "number", text: "10.0.0.0/8" },
      { type: "arrow", text: "->" },
      { type: "outbound", text: "custom-2" },
      { type: "comment", text: "# end" },
    ]);
  });
  test("does not interpret syntax in comments or escaped strings", () => {
    expect(tokenize('# default: -> "x"')).toEqual([
      { type: "comment", text: '# default: -> "x"' },
    ]);
    expect(tokenize(String.raw`"a\"#b"`)).toEqual([
      { type: "string", text: String.raw`"a\"#b"` },
    ]);
    expect(tokenize("default: proxy")[0]).toEqual({
      type: "keyword",
      text: "default:",
    });
    expect(tokenize("outbound: x = socks(port: 80)")[0].type).toBe("keyword");
  });
});
