// @vitest-environment happy-dom
import { afterEach, describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { mountWithApp } from "@/test/mount";
import TrafficCard from "./TrafficCard.vue";

const props = {
  up: 1536,
  down: 1.25 * 1024 ** 2,
  upTotal: 2 * 1024 ** 3,
  downTotal: 3.5 * 1024 ** 3,
  upSeries: [0, 1024, 1536],
  downSeries: [1024, 2 * 1024 ** 2, 1.25 * 1024 ** 2],
};
const originalWidth = window.innerWidth;
let wrapper: VueWrapper | undefined;

afterEach(() => {
  wrapper?.unmount();
  window.innerWidth = originalWidth;
  window.dispatchEvent(new Event("resize"));
});

describe("traffic card", () => {
  test("shows formatted rates and totals and updates with incoming props", async () => {
    wrapper = mountWithApp(TrafficCard, { props });
    expect(wrapper.text()).toContain("1.5 KiB/s");
    expect(wrapper.text()).toContain("1.3 MiB/s");
    expect(wrapper.text()).toContain("2.0 GiB");
    expect(wrapper.text()).toContain("3.5 GiB");
    const chart = wrapper.get("svg.traffic-chart");
    const path = chart.get(".traffic-chart__down").attributes("d");
    await wrapper.setProps({
      down: 512,
      downTotal: 4 * 1024 ** 3,
      downSeries: [0, 256, 512],
    });
    expect(wrapper.text()).toContain("512.0 B/s");
    expect(wrapper.text()).toContain("4.0 GiB");
    expect(chart.get(".traffic-chart__down").attributes("d")).not.toBe(path);
  });
});
