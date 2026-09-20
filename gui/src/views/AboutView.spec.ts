// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { mountWithApp } from "@/test/mount";
import About from "./AboutView.vue";

describe("the about dialog", () => {
  test("says what v2rayA is and links the discussions", () => {
    const w = mountWithApp(About);
    expect(w.text()).toContain("A web client for its own Xray-based core");
    expect(w.text()).toContain("Founded by @mzz2017");
    expect(w.find('a[href$="/discussions"]').exists()).toBe(true);
    w.unmount();
  });
});
