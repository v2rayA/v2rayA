import { describe, expect, test } from "vitest";
import { parse, serialize } from "./rules";
import { template } from "./template";

describe("visual RoutingA model", () => {
  test.each([
    template,
    '  # keep spacing\r\n\r\noutbound: mine = http(address: localhost, port: 80, pass: "a,b#c")\r\ndomain(regexp: "a,b(->)&&")&&port(80,443)->mine\r\nunknown syntax ->\r\n',
  ])("round-trips original bytes", (text) => {
    expect(serialize(parse(text))).toBe(text);
  });
  test("splits only unquoted delimiters and formats only edited lines", () => {
    const text =
      '# untouched\ndomain(regexp: "a,b(->)&&", full: x) && port(80,443)->mine\n??\n';
    const entries = parse(text);
    expect(entries[1]).toMatchObject({
      kind: "rule",
      conditions: [
        { fn: "domain", args: ['regexp: "a,b(->)&&"', "full: x"] },
        { fn: "port", args: ["80", "443"] },
      ],
      outbound: "mine",
    });
    if (entries[1].kind !== "rule") throw new Error("Expected a rule");
    entries[1].outbound = "block";
    expect(serialize(entries)).toBe(
      '# untouched\ndomain(regexp: "a,b(->)&&", full: x) && port(80, 443) -> block\n??\n',
    );
    expect(entries[2].kind).toBe("raw");
  });
  test("retains malformed and unsupported lines rather than dropping them", () => {
    const text =
      'domain("unterminated) -> proxy\ndefault: proxy # comment\noutbound: a = socks(address)\nport(80)) -> direct';
    expect(parse(text).map((entry) => entry.kind)).toEqual([
      "raw",
      "raw",
      "raw",
      "raw",
    ]);
    expect(serialize(parse(text))).toBe(text);
  });
});
