// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import type * as Api from "@/api";
import { getSetting, postImport } from "@/api";
import {
  closeAllDialogs,
  closeDialog,
  dialogState,
} from "@/composables/useDialog";
import { closeAllNotices } from "@/composables/useNotify";
import { useAppStore } from "@/stores/app";
import { mountWithApp } from "@/test/mount";
import ImportDialog from "./Import.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getSetting: vi.fn(),
  postImport: vi.fn(),
}));

let wrapper: VueWrapper;
const address = "https://subscription.example/list";

beforeEach(() => {
  vi.clearAllMocks();
  closeAllDialogs();
  closeAllNotices();
  vi.mocked(getSetting).mockResolvedValue({
    setting: {
      transparent: "close",
      transparentType: "tproxy",
      pacMode: "whitelist",
      logLevel: "info",
      proxyModeWhenSubscribe: "proxy",
    },
    localGFWListVersion: "",
    localGeositeVersion: "",
  });
  vi.mocked(postImport).mockResolvedValue({} as never);
  wrapper = mountWithApp(ImportDialog, {
    props: { kind: "subscription" },
  });
  useAppStore().setRunning("stopped");
});

afterEach(() => {
  wrapper.unmount();
  closeAllDialogs();
  closeAllNotices();
});

async function submit() {
  await wrapper.get("textarea").setValue(address);
  await wrapper
    .findAll("button")
    .find((button) => button.text() === "Import")!
    .trigger("click");
  await flushPromises();
}

describe("subscription import with a stopped core", () => {
  test("sends one explicit proxy bypass after Yes", async () => {
    await submit();
    expect(postImport).not.toHaveBeenCalled();

    closeDialog(dialogState.stack.at(-1)!.id, true);
    await flushPromises();

    expect(postImport).toHaveBeenCalledExactlyOnceWith({
      url: address,
      kind: "subscription",
      bypassProxy: true,
    });
    expect(wrapper.emitted("close")).toEqual([[true]]);
  });

  test("keeps the import open and sends nothing after No", async () => {
    await submit();

    closeDialog(dialogState.stack.at(-1)!.id, false);
    await flushPromises();

    expect(postImport).not.toHaveBeenCalled();
    expect(wrapper.emitted("close")).toBeUndefined();
    expect(wrapper.get<HTMLTextAreaElement>("textarea").element.value).toBe(
      address,
    );
  });
});
