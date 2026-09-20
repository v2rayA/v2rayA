// @vitest-environment happy-dom
import { beforeEach, describe, expect, test, vi } from "vitest";
import { flushPromises } from "@vue/test-utils";
import { mountWithApp } from "@/test/mount";
import type * as Api from "@/api";
import { getParams } from "@/api";
import { useAppStore } from "@/stores/app";
import DocsView from "./DocsView.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getParams: vi.fn(),
}));

// the Markdown is a dynamic import; wait until the article is in
const settle = (w: { find: (s: string) => { exists(): boolean } }) =>
  vi.waitFor(async () => {
    await flushPromises();
    expect(w.find(".docs__body h1").exists()).toBe(true);
  });

describe("the documentation page", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.mocked(getParams).mockResolvedValue({
      params: [
        {
          flag: "--address",
          short: "a",
          env: "V2RAYA_ADDRESS",
          default: "0.0.0.0:2017",
          desc: "Listening address",
        },
      ],
    });
  });

  test("renders the Markdown of the first section with heading ids", async () => {
    const w = mountWithApp(DocsView);
    await settle(w);
    expect(w.find(".docs__body h1").text()).toBe("Quick start");
    expect(w.find(".docs__list .v-list-item--active").text()).toBe(
      "Quick start",
    );
    expect(w.find(".docs__body h2#sign-in").exists()).toBe(true);
    expect(w.find(".docs__body pre code").text()).toContain("--reset-password");
    w.unmount();
  });

  test("switches to the parameters section and lists the flags", async () => {
    const w = mountWithApp(DocsView);
    await settle(w);
    await w
      .findAll(".docs__list .v-list-item")
      .find((item) => item.text() === "Flags and environment")!
      .trigger("click");
    await vi.waitFor(async () => {
      await flushPromises();
      expect(w.find(".docs__body h1").text()).toBe(
        "Flags and environment variables",
      );
      expect(w.find(".docs__params tbody tr").exists()).toBe(true);
    });
    const row = w.find(".docs__params tbody tr");
    expect(row.text()).toContain("--address");
    expect(row.text()).toContain("V2RAYA_ADDRESS");
    expect(getParams).toHaveBeenCalledTimes(1);
    expect(useAppStore().docsSection).toBe("parameters");
    w.unmount();
  });
});
