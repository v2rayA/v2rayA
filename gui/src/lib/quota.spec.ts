import { describe, expect, test } from "vitest";
import { parseQuota } from "./quota";

describe("parseQuota", () => {
  test("reads the three shapes the backend writes", () => {
    expect(
      parseQuota("Used 54.30 GiB / 100.00 GiB · Expires 2026-10-04"),
    ).toEqual({
      used: "54.30 GiB",
      total: "100.00 GiB",
      expires: "2026-10-04",
      percent: 54.3,
    });
    expect(parseQuota("Used 1.50 GiB")).toEqual({
      used: "1.50 GiB",
      total: "",
      expires: "",
    });
    expect(parseQuota("Total 20.00 GiB · Expires 2027-01-01")).toEqual({
      used: "",
      total: "20.00 GiB",
      expires: "2027-01-01",
    });
    expect(parseQuota("Expires 2027-01-01")).toEqual({
      used: "",
      total: "",
      expires: "2027-01-01",
    });
    expect(parseQuota("")).toBeNull();
    expect(parseQuota("anything else")).toBeNull();
  });

  test("normalises units and caps the bar", () => {
    expect(parseQuota("Used 512.00 MiB / 1.00 GiB")?.percent).toBe(50);
    expect(parseQuota("Used 3.00 GiB / 1.00 GiB")?.percent).toBe(100);
    expect(parseQuota("Used 1.00 GiB / 0.00 GiB")?.percent).toBeUndefined();
  });
});
