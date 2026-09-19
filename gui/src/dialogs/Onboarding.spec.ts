// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import DialogHost from "@/components/hosts/DialogHost.vue";
import { closeAllDialogs, openDialog } from "@/composables/useDialog";
import ImportDialog from "@/dialogs/Import.vue";
import ServerDialog from "@/dialogs/Server/index.vue";
import en from "@/locales/en";
import { useAppStore } from "@/stores/app";
import { mountWithApp } from "@/test/mount";
import AboutDialog from "@/views/settings/AboutDialog.vue";
import OnboardingDialog, { shouldShowOnboarding } from "./Onboarding.vue";

let host: VueWrapper;
let tutorial: Omit<VueWrapper, "exists">;

function button(wrapper: Pick<VueWrapper, "findAll">, text: string) {
  const result = wrapper.findAll("button").find((item) => item.text() === text);
  if (!result) throw new Error(`Button not found: ${text}`);
  return result;
}

async function openTutorial() {
  host = mountWithApp(DialogHost);
  const handle = openDialog<boolean>(OnboardingDialog, {}, { width: 560 });
  await flushPromises();
  tutorial = host.getComponent(OnboardingDialog);
  return handle;
}

beforeEach(() => localStorage.removeItem("onboardingSeen"));
afterEach(() => {
  closeAllDialogs();
  host?.unmount();
  localStorage.removeItem("onboardingSeen");
});

describe("onboarding", () => {
  test("offers the tutorial only when the seen key is absent", () => {
    expect(shouldShowOnboarding()).toBe(true);
    localStorage.setItem("onboardingSeen", "1");
    expect(shouldShowOnboarding()).toBe(false);
    localStorage.setItem("onboardingSeen", "0");
    expect(shouldShowOnboarding()).toBe(false);
  });

  test("starts at import and moves the page and indicator together", async () => {
    await openTutorial();
    const expectStep = (index: number, title: string) => {
      expect(tutorial.get(".v-window-item--active h2").text()).toBe(title);
      expect(
        tutorial
          .findAll(".onboarding-dot")
          .map((dot) => dot.classes().includes("onboarding-dot--active")),
      ).toEqual([index === 0, index === 1, index === 2, index === 3]);
      expect(tutorial.get('[role="status"]').attributes("aria-label")).toBe(
        en.onboarding.progress
          .replace("{current}", String(index + 1))
          .replace("{total}", "4"),
      );
    };
    expectStep(0, en.onboarding.importTitle);
    expect(
      button(tutorial, en.onboarding.back).attributes("disabled"),
    ).toBeDefined();
    await button(tutorial, en.onboarding.next).trigger("click");
    expectStep(1, en.onboarding.groupTitle);
    await button(tutorial, en.onboarding.back).trigger("click");
    expectStep(0, en.onboarding.importTitle);
    await button(tutorial, en.onboarding.next).trigger("click");
    await button(tutorial, en.onboarding.next).trigger("click");
    expectStep(2, en.onboarding.rulesTitle);
    await button(tutorial, en.onboarding.next).trigger("click");
    expectStep(3, en.onboarding.startTitle);
    expect(localStorage.getItem("onboardingSeen")).toBeNull();
  });

  test("opens working import and new-node dialogs without dismissing the tutorial", async () => {
    await openTutorial();
    await button(tutorial, en.operations.import).trigger("click");
    await flushPromises();
    const importer = host.getComponent(ImportDialog);
    expect(importer.find("textarea").exists()).toBe(true);
    await button(importer, en.operations.cancel).trigger("click");
    await flushPromises();
    expect(host.findComponent(ImportDialog).exists()).toBe(false);
    expect(tutorial.get(".v-window-item--active h2").text()).toBe(
      en.onboarding.importTitle,
    );
    await button(tutorial, en.onboarding.newNode).trigger("click");
    await flushPromises();
    const editor = host.getComponent(ServerDialog);
    expect(editor.find("form").exists()).toBe(true);
    await button(editor, en.operations.cancel).trigger("click");
    await flushPromises();
    expect(host.findComponent(ServerDialog).exists()).toBe(false);
    expect(host.findComponent(OnboardingDialog).exists()).toBe(true);
    expect(localStorage.getItem("onboardingSeen")).toBeNull();
  });

  test("changes the page behind the tutorial, then finishes on the dashboard", async () => {
    const handle = await openTutorial();
    await button(tutorial, en.onboarding.next).trigger("click");
    await button(tutorial, en.onboarding.goToProxies).trigger("click");
    expect(useAppStore().view).toBe("proxies");
    expect(host.findComponent(OnboardingDialog).exists()).toBe(true);
    expect(localStorage.getItem("onboardingSeen")).toBeNull();
    await button(tutorial, en.onboarding.next).trigger("click");
    await button(tutorial, en.onboarding.next).trigger("click");
    await button(tutorial, en.onboarding.finish).trigger("click");
    expect(await handle.result).toBe(true);
    expect(useAppStore().view).toBe("dashboard");
    expect(host.findComponent(OnboardingDialog).exists()).toBe(false);
    expect(localStorage.getItem("onboardingSeen")).toBe("1");
    expect(shouldShowOnboarding()).toBe(false);
  });

  test("reopens from About after the tutorial has already been seen", async () => {
    localStorage.setItem("onboardingSeen", "1");
    host = mountWithApp(DialogHost);
    openDialog(AboutDialog);
    await flushPromises();
    await button(
      host.getComponent(AboutDialog),
      en.onboarding.viewTutorial,
    ).trigger("click");
    await flushPromises();
    expect(
      host
        .getComponent(OnboardingDialog)
        .get(".v-window-item--active h2")
        .text(),
    ).toBe(en.onboarding.importTitle);
    expect(shouldShowOnboarding()).toBe(false);
  });

  test.each(["close", "escape"])(
    "counts %s dismissal as seen without changing the page",
    async (action) => {
      const handle = await openTutorial();
      useAppStore().view = "proxies";
      if (action === "close") {
        await tutorial
          .get(`button[aria-label="${en.operations.close}"]`)
          .trigger("click");
      } else {
        // Vuetify registers the top overlay after its opening frame.
        await vi.waitFor(() => {
          window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
          expect(host.findComponent(OnboardingDialog).exists()).toBe(false);
        });
      }
      await handle.result;
      expect(useAppStore().view).toBe("proxies");
      expect(localStorage.getItem("onboardingSeen")).toBe("1");
      expect(shouldShowOnboarding()).toBe(false);
    },
  );
});
