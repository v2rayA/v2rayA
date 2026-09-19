// @vitest-environment happy-dom
import { afterEach, describe, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { mountWithApp } from "@/test/mount";
import { call } from "@/api/client";
import LoadingHost from "./LoadingHost.vue";

let wrapper: VueWrapper;
afterEach(() => {
  wrapper?.unmount();
  vi.useRealTimers();
});

describe("loading host", () => {
  test("shows the top bar only when a request outlasts a moment", async () => {
    vi.useFakeTimers();
    wrapper = mountWithApp(LoadingHost);
    let release!: () => void;
    const pending = new Promise<void>((r) => (release = r));
    // a request that hangs until released; the adapter never reaches the network
    const request = call({
      url: "version",
      adapter: () =>
        pending.then(() => ({
          data: { code: "SUCCESS", data: 1 },
          status: 200,
          statusText: "OK",
          headers: {},
          config: {} as never,
        })),
    }).catch(() => {});
    await flushPromises();
    expect(wrapper.find(".loading-bar").exists()).toBe(false);
    vi.advanceTimersByTime(250);
    await flushPromises();
    expect(wrapper.find(".loading-bar").exists()).toBe(true);
    release();
    await request;
    await flushPromises();
    expect(wrapper.find(".loading-bar").exists()).toBe(false);
  });
});
