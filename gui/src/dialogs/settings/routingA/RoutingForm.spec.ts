// @vitest-environment happy-dom
import { afterEach, describe, expect, test } from "vitest";
import { enableAutoUnmount, flushPromises } from "@vue/test-utils";
import { mountWithApp } from "@/test/mount";
import { closeAllDialogs, dialogState } from "@/composables/useDialog";
import { useAppStore } from "@/stores/app";
import en from "@/locales/en";

import RoutingForm from "./RoutingForm.vue";

enableAutoUnmount(afterEach);
afterEach(closeAllDialogs);

const groups = ["proxy", "GEMINI", "TMP"];

function mountForm() {
  return mountWithApp(RoutingForm, {
    props: { modelValue: "default: proxy\n" },
  });
}

describe("the routing form's outbound options", () => {
  test("the default outbound select offers the proxy groups", async () => {
    const w = mountForm();
    useAppStore().setOutbounds(groups);
    await flushPromises();
    const select = w
      .findAllComponents({ name: "VSelect" })
      .find((c) => c.props("label") === en.routingA.form.default)!;
    // the store's "proxy" folds into the built-in, no duplicate
    expect(select.props("items")).toEqual([
      "proxy",
      "direct",
      "block",
      "GEMINI",
      "TMP",
    ]);
  });

  test("the rule dialog's outbound options include the proxy groups", async () => {
    const w = mountForm();
    useAppStore().setOutbounds(groups);
    await flushPromises();
    const add = w
      .findAll("button")
      .find((b) => b.text() === en.routingA.form.addRule)!;
    await add.trigger("click");
    await flushPromises();
    expect(dialogState.stack.at(-1)?.props.outbounds).toEqual([
      "proxy",
      "direct",
      "block",
      "GEMINI",
      "TMP",
    ]);
  });
});
