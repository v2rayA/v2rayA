import { describe, expect, test } from "vitest";
import type { TrafficMessage } from "@/api/types";
import { formatBytes, formatRate } from "@/lib/format";
import { createTraffic } from "./useTraffic";

function message(n: number): TrafficMessage {
  return {
    type: "traffic",
    body: { up: n, down: n * 2, upTotal: n * 100, downTotal: n * 200 },
  };
}

describe("traffic history", () => {
  test("keeps the newest 30 rates in order across repeated wraps, padded to the window", () => {
    const traffic = createTraffic();
    expect(traffic.upSeries.value).toEqual(new Array(30).fill(0));
    for (let n = 1; n <= 125; n++) {
      traffic.feed(message(n));
      const expected = [
        ...new Array(Math.max(0, 30 - n)).fill(0),
        ...Array.from(
          { length: Math.min(n, 30) },
          (_, i) => Math.max(1, n - 29) + i,
        ),
      ];
      expect(traffic.upSeries.value).toEqual(expected);
      expect(traffic.downSeries.value).toEqual(expected.map((v) => v * 2));
    }
    expect(traffic.up.value).toBe(125);
    expect(traffic.down.value).toBe(250);
    expect(traffic.upTotal.value).toBe(12500);
    expect(traffic.downTotal.value).toBe(25000);
  });

  test("resets a wrapped history and starts a new session without stale samples", () => {
    const traffic = createTraffic();
    for (let n = 1; n <= 65; n++) traffic.feed(message(n));
    traffic.reset();
    expect(traffic.upSeries.value).toEqual(new Array(30).fill(0));
    expect(traffic.downSeries.value).toEqual(new Array(30).fill(0));
    expect([
      traffic.up.value,
      traffic.down.value,
      traffic.upTotal.value,
      traffic.downTotal.value,
    ]).toEqual([0, 0, 0, 0]);
    traffic.feed(message(2));
    expect(traffic.upSeries.value.slice(-2)).toEqual([0, 2]);
    expect(traffic.downSeries.value.slice(-2)).toEqual([0, 4]);
    expect(traffic.upTotal.value).toBe(200);
    expect(traffic.downTotal.value).toBe(400);
  });
});

describe("traffic formatting", () => {
  test.each([
    [0, "0.0 B"],
    [1023, "1023.0 B"],
    [1024, "1.0 KiB"],
    [1536, "1.5 KiB"],
    [1024 ** 2, "1.0 MiB"],
    [1.25 * 1024 ** 2, "1.3 MiB"],
    [1024 ** 3, "1.0 GiB"],
    [1024 ** 4, "1024.0 GiB"],
  ])("formats %s bytes with binary units and one decimal", (n, expected) => {
    expect(formatBytes(n)).toBe(expected);
    expect(formatRate(n)).toBe(`${expected}/s`);
  });
});
