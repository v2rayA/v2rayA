// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { generateShareLink, parseShareLink } from "./serverCodec";
import { links } from "./__fixtures__/links";
import expected from "./__fixtures__/serverCodec.json";

// The expected values are what the editor produced before the codec
// moved out of modalServer.vue: this suite pins the move, not a spec of
// the share-link formats (the http branch, for one, generates
// http-proxy:// and parses http://, and that stays as it is).
describe("serverCodec keeps the editor's behaviour", () => {
  for (const [name, link] of Object.entries(links)) {
    test(`${name}: link → form`, () => {
      expect(parseShareLink(link)).toEqual(
        expected[name as keyof typeof expected].form,
      );
    });
    test(`${name}: form → link`, () => {
      const form = parseShareLink(link);
      expect(form && generateShareLink(form)).toEqual(
        expected[name as keyof typeof expected].back,
      );
    });
  }

  test("an unknown scheme is null both ways", () => {
    expect(parseShareLink("gopher://x")).toBeNull();
    expect(generateShareLink({ protocol: "gopher" })).toBeNull();
  });
});

test("a wireguard link keeps workers, reserved and kernelMode through parse and generate", () => {
  const link =
    "wireguard://cHJpdmF0ZQ@1.2.3.4:51820?publicKey=cHVi&address=10.0.0.2%2F32&reserved=1%2C2%2C3&workers=4&kernelMode=true#wg";
  const model = parseShareLink(link) as Record<string, unknown>;
  expect(model.reserved).toBe("1,2,3");
  expect(model.workers).toBe("4");
  expect(model.kernelMode).toBe(true);
  const out = generateShareLink(model);
  expect(out).toContain("reserved=1%2C2%2C3");
  expect(out).toContain("workers=4");
  expect(out).toContain("kernelMode=true");
});
