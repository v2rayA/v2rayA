// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VProgressLinear } from "vuetify/components";
import type * as Api from "@/api";
import { getTouch } from "@/api";
import { mountWithApp } from "@/test/mount";
import { dialogState, closeAllDialogs } from "@/composables/useDialog";
import ImportDialog from "@/dialogs/Import.vue";
import SubscriptionSettings from "@/dialogs/SubscriptionSettings.vue";
import SubscriptionsView from "./SubscriptionsView.vue";
import SubscriptionCard from "./proxies/SubscriptionCard.vue";
import { fixture } from "./proxies/fixture";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getTouch: vi.fn(),
}));
dayjs.extend(utc);
dayjs.extend(timezone);
let wrapper: VueWrapper;
const button = (text: string) =>
  wrapper.findAll("button").find((b) => b.text() === text)!;
beforeEach(async () => {
  vi.clearAllMocks();
  vi.mocked(getTouch).mockImplementation(async () => fixture());
  wrapper = mountWithApp(SubscriptionsView);
  await flushPromises();
});
afterEach(() => {
  wrapper.unmount();
  closeAllDialogs();
});

describe("subscriptions page", () => {
  test("shows each subscription with its quota", () => {
    expect(
      wrapper.findAllComponents(SubscriptionCard).map((c) => c.text()),
    ).toEqual(
      expect.arrayContaining([
        expect.stringContaining("feed-0.example"),
        expect.stringContaining("feed-1.example"),
      ]),
    );
    expect(
      wrapper
        .findAllComponents(VProgressLinear)
        .some((p) => p.props("modelValue") === 25),
    ).toBe(true);
  });
  test("its buttons open the settings and the import dialogs", async () => {
    await button("Auto-update").trigger("click");
    expect(dialogState.stack.at(-1)?.component).toBe(SubscriptionSettings);
    closeAllDialogs();
    await flushPromises();
    await button("Import a subscription").trigger("click");
    expect(dialogState.stack.at(-1)?.component).toBe(ImportDialog);
    expect(dialogState.stack.at(-1)?.props).toMatchObject({
      kind: "subscription",
    });
  });
  test("with no subscriptions it offers to import one", async () => {
    wrapper.unmount();
    vi.mocked(getTouch).mockImplementation(async () => ({
      ...fixture(),
      touch: { ...fixture().touch, subscriptions: [] },
    }));
    wrapper = mountWithApp(SubscriptionsView);
    await flushPromises();
    expect(wrapper.findComponent(SubscriptionCard).exists()).toBe(false);
    expect(wrapper.text()).toContain("No subscriptions yet");
  });
});
