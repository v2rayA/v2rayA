// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { mountWithApp } from "@/test/mount";
import About from "./AboutView.vue";

describe("the about dialog", () => {
  test("says what v2rayA is and links the discussions", () => {
    const w = mountWithApp(About);
    expect(w.text()).toContain("v2rayA is a web GUI");
    expect(w.find('a[href$="/discussions"]').exists()).toBe(true);
    w.unmount();
  });
});
