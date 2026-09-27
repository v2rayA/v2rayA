// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { defineComponent, h } from "vue";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import type * as Api from "@/api";
import { getSetting } from "@/api";
import ConfirmDialog from "@/components/hosts/ConfirmDialog.vue";
import { mountWithApp } from "@/test/mount";
import { useAppStore, type Running } from "@/stores/app";
import { closeAllDialogs, closeDialog, dialogState } from "./useDialog";
import { useSubscriptionBypass } from "./useSubscriptionBypass";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getSetting: vi.fn(),
}));

let wrapper: VueWrapper;
let chooseBypass: () => Promise<boolean | null>;

const Probe = defineComponent({
  setup() {
    chooseBypass = useSubscriptionBypass();
    return () => h("div");
  },
});

function setting(proxyModeWhenSubscribe?: string) {
  return {
    setting: {
      transparent: "close",
      transparentType: "tproxy",
      pacMode: "whitelist",
      logLevel: "info",
      proxyModeWhenSubscribe,
    },
    localGFWListVersion: "",
    localGeositeVersion: "",
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  closeAllDialogs();
  wrapper = mountWithApp(Probe);
});

afterEach(() => {
  wrapper.unmount();
  closeAllDialogs();
});

describe("useSubscriptionBypass", () => {
  test.each<Running>(["checking", "running", "paused"])(
    "uses the configured route without loading settings while the core is %s",
    async (running) => {
      useAppStore().setRunning(running);

      await expect(chooseBypass()).resolves.toBe(false);

      expect(getSetting).not.toHaveBeenCalled();
      expect(dialogState.stack).toHaveLength(0);
    },
  );

  test("uses direct mode without asking when the core is stopped", async () => {
    useAppStore().setRunning("stopped");
    vi.mocked(getSetting).mockResolvedValue(setting("direct"));

    await expect(chooseBypass()).resolves.toBe(false);

    expect(getSetting).toHaveBeenCalledOnce();
    expect(dialogState.stack).toHaveLength(0);
  });

  test.each(["proxy", "pac"])(
    "returns true after confirmation for stopped-core %s mode",
    async (mode) => {
      useAppStore().setRunning("stopped");
      vi.mocked(getSetting).mockResolvedValue(setting(mode));

      const result = chooseBypass();
      await flushPromises();
      const dialog = dialogState.stack.at(-1)!;
      expect(dialog.component).toBe(ConfirmDialog);
      expect(dialog.props).toMatchObject({
        message: "The core is stopped. Update without the proxy?",
        confirmText: "Yes",
        cancelText: "No",
      });

      closeDialog(dialog.id, true);
      await expect(result).resolves.toBe(true);
    },
  );

  test("returns null when direct access is declined", async () => {
    useAppStore().setRunning("stopped");
    vi.mocked(getSetting).mockResolvedValue(setting("proxy"));

    const result = chooseBypass();
    await flushPromises();
    closeDialog(dialogState.stack.at(-1)!.id, false);

    await expect(result).resolves.toBeNull();
  });
});
