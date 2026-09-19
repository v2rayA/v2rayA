import { describe, expect, test } from "vitest";
import { check } from "./check";
import { template } from "./template";

describe("RoutingA advisory checker", () => {
  test("accepts the template, definitions, quotes and comments", () => {
    expect(check(template)).toEqual([]);
    expect(
      check(
        `\n# ( ->\noutbound: x = http(address: 'host#name')\ndomain(regexp: '[()]#->') -> x # )`,
      ),
    ).toEqual([]);
  });
  test("reports line numbers and the three advisory errors", () => {
    expect(
      check(
        "# rules\ndomain(full: example.com)\nip(10.0.0.0/8 -> direct\nport(80) -> # no outbound\nport([80)] -> direct",
      ),
    ).toEqual([
      { line: 2, message: "routingA.errors.noArrow" },
      { line: 3, message: "routingA.errors.brackets" },
      { line: 4, message: "routingA.errors.noOutbound" },
      { line: 5, message: "routingA.errors.brackets" },
    ]);
  });
  test("does not treat a quoted arrow as a rule separator", () => {
    expect(check('domain(contains: "->")')).toEqual([
      { line: 1, message: "routingA.errors.noArrow" },
    ]);
  });
});
