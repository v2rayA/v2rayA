// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import {
  enableAutoUnmount,
  flushPromises,
  type VueWrapper,
} from "@vue/test-utils";
import { mountWithApp } from "@/test/mount";
import {
  closeAllDialogs,
  closeDialog,
  dialogState,
} from "@/composables/useDialog";
import { closeAllNotices, noticeState } from "@/composables/useNotify";
import DialogHost from "@/components/hosts/DialogHost.vue";
import en from "@/locales/en";
import RoutingA from "./RoutingA.vue";
import RuleDialog from "./routingA/RuleDialog.vue";
import { template } from "./routingA/template";

const api = vi.hoisted(() => ({ getRoutingA: vi.fn(), putRoutingA: vi.fn() }));
vi.mock("@/api", () => api);
const rules = "default: direct\n# inbound(http, 8080)\n";
const button = (wrapper: Pick<VueWrapper, "findAll">, label: string) =>
  wrapper.findAll("button").find((item) => item.text() === label)!;
const textView = async (wrapper: VueWrapper) => {
  await wrapper
    .get(`button[aria-label="${en.routingA.form.text}"]`)
    .trigger("click");
  return wrapper.get("textarea");
};
const respond = async (answer: boolean) => {
  closeDialog(dialogState.stack.at(-1)!.id, answer);
  await flushPromises();
};

enableAutoUnmount(afterEach);
beforeEach(() => {
  localStorage.clear();
  api.getRoutingA.mockReset().mockResolvedValue({ routingA: rules });
  api.putRoutingA.mockReset().mockResolvedValue(null);
});
afterEach(() => {
  closeAllDialogs();
  closeAllNotices();
  vi.restoreAllMocks();
  vi.useRealTimers();
});

describe("RoutingA dialog", () => {
  test("loads, highlights and saves rules verbatim without confirming comments", async () => {
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    const editor = await textView(wrapper);
    expect(editor.element.value).toBe(rules);
    expect(wrapper.get(".routing-editor .tok-keyword").text()).toBe("default:");
    await button(wrapper, en.operations.save).trigger("click");
    await flushPromises();
    expect(api.putRoutingA).toHaveBeenCalledExactlyOnceWith({
      routingA: rules,
    });
    expect(dialogState.stack).toHaveLength(0);
    expect(wrapper.emitted("close")).toEqual([[true]]);
  });

  test("marks bad lines after debounce but lets the backend judge on save", async () => {
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    const editor = await textView(wrapper);
    vi.useFakeTimers();
    await editor.setValue("default: proxy\nbad rule");
    await vi.advanceTimersByTimeAsync(150);
    expect(
      wrapper.get(".routing-editor__number--error").attributes("title"),
    ).toBe(en.routingA.errors.noArrow);
    expect(wrapper.get(".routing-editor__number--error").text()).toContain("2");
    await button(wrapper, en.operations.save).trigger("click");
    await flushPromises();
    expect(api.putRoutingA).toHaveBeenCalledWith({
      routingA: "default: proxy\nbad rule",
    });
  });

  test("shows raw backend errors safely and clears them on the next edit", async () => {
    const error = "invalid RoutingA rules: [error] table[1] <bad>";
    api.putRoutingA.mockRejectedValueOnce(new Error(error));
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    await button(wrapper, en.operations.save).trigger("click");
    await flushPromises();
    expect(wrapper.get('[data-testid="routing-error"]').text()).toBe(error);
    expect(wrapper.find("bad").exists()).toBe(false);
    expect(wrapper.emitted("close")).toBeUndefined();
    await (await textView(wrapper)).setValue("default: block");
    expect(wrapper.find('[data-testid="routing-error"]').exists()).toBe(false);
  });

  test("confirms reset and unsaved close without changing text on cancellation", async () => {
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    const editor = await textView(wrapper);
    await button(wrapper, en.routingA.resetDefault).trigger("click");
    await respond(false);
    expect(editor.element.value).toBe(rules);
    await button(wrapper, en.routingA.resetDefault).trigger("click");
    await respond(true);
    expect(editor.element.value).toBe(template);
    await button(wrapper, en.operations.cancel).trigger("click");
    await respond(false);
    expect(wrapper.emitted("close")).toBeUndefined();
    await wrapper
      .get(`button[aria-label="${en.operations.close}"]`)
      .trigger("click");
    await respond(true);
    expect(wrapper.emitted("close")).toEqual([[]]);
  });

  test("inserts a reference example at the caret and remembers panel visibility", async () => {
    localStorage.setItem("routingA.reference", "true");
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    const editor = await textView(wrapper);
    editor.element.setSelectionRange(rules.length, rules.length);
    await wrapper
      .get(`button[aria-label="${en.routingA.insert}"]`)
      .trigger("click");
    expect(editor.element.value).toBe(
      rules + "# HTTPS\nport(443) && network(tcp) -> proxy",
    );
    await editor.setValue("default: proxy");
    editor.element.setSelectionRange(0, 0);
    await wrapper
      .get(`button[aria-label="${en.routingA.insert}"]`)
      .trigger("click");
    expect(editor.element.value).toBe(
      "# HTTPS\nport(443) && network(tcp) -> proxy\ndefault: proxy",
    );
    await button(wrapper, en.routingA.reference.title).trigger("click");
    expect(localStorage.getItem("routingA.reference")).toBe("false");
  });

  test("supports indentation, tabs, current line and keyboard save", async () => {
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    const editor = await textView(wrapper);
    await editor.setValue("  default: proxy");
    editor.element.setSelectionRange(
      editor.element.value.length,
      editor.element.value.length,
    );
    await editor.trigger("keydown", { key: "Enter" });
    expect(editor.element.value).toBe("  default: proxy\n  ");
    await editor.trigger("keydown", { key: "Tab" });
    expect(editor.element.value).toBe("  default: proxy\n    ");
    expect(wrapper.get(".routing-editor__number--current").text()).toBe("2");
    await editor.trigger("keydown", { key: "s", ctrlKey: true });
    await flushPromises();
    expect(api.putRoutingA).toHaveBeenCalledWith({
      routingA: editor.element.value,
    });
  });

  test("still confirms deprecated inbounds after dismissing their warning", async () => {
    api.getRoutingA.mockResolvedValue({ routingA: "inbound (socks, 1080)" });
    api.putRoutingA.mockResolvedValue({ warning: "Deprecated inbound" });
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    await wrapper.get(".v-alert__close button").trigger("click");
    await button(wrapper, en.operations.save).trigger("click");
    expect(api.putRoutingA).not.toHaveBeenCalled();
    await respond(false);
    expect(api.putRoutingA).not.toHaveBeenCalled();
    await button(wrapper, en.operations.save).trigger("click");
    await respond(true);
    expect(api.putRoutingA).toHaveBeenCalledOnce();
    expect(noticeState.current?.kind).toBe("warning");
  });

  test("adds a rule through the form and exposes it in the text view", async () => {
    const wrapper = mountWithApp(RoutingA);
    const host = mountWithApp(DialogHost);
    await flushPromises();
    await button(wrapper, en.routingA.form.addRule).trigger("click");
    await flushPromises();
    const dialog = host.getComponent(RuleDialog);
    await dialog
      .get(".rule-condition__arguments input")
      .setValue("full: example.com");
    await button(dialog, en.operations.save).trigger("click");
    await flushPromises();
    expect((await textView(wrapper)).element.value).toContain(
      "domain(full: example.com) -> proxy",
    );
    expect(localStorage.getItem("routingA.view")).toBe("text");
  });

  test("exports exact text and imports only after confirmation", async () => {
    const create = vi
      .spyOn(URL, "createObjectURL")
      .mockReturnValue("blob:routing");
    vi.spyOn(URL, "revokeObjectURL").mockImplementation(() => {});
    const click = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => {});
    const wrapper = mountWithApp(RoutingA);
    await flushPromises();
    await button(wrapper, en.routingA.export).trigger("click");
    const blob = create.mock.calls[0][0];
    const anchor = click.mock.instances[0];
    if (!(blob instanceof Blob) || !(anchor instanceof HTMLAnchorElement))
      throw new Error("Expected a downloadable text blob");
    expect(await blob.text()).toBe(rules);
    expect(anchor.download).toBe("routingA.txt");
    const input = wrapper.get('input[type="file"]');
    const imported = "default: block\n# imported\n";
    Object.defineProperty(input.element, "files", {
      configurable: true,
      value: [new File([imported], "rules.txt")],
    });
    await input.trigger("change");
    await flushPromises();
    await respond(false);
    expect((await textView(wrapper)).element.value).toBe(rules);
    await input.trigger("change");
    await flushPromises();
    await respond(true);
    expect(wrapper.get("textarea").element.value).toBe(imported);
  });

  test("closes on load failure without writing an empty script", async () => {
    api.getRoutingA.mockRejectedValue(new Error("Offline"));
    const wrapper = mountWithApp(RoutingA);
    expect(
      button(wrapper, en.operations.save).attributes("disabled"),
    ).toBeDefined();
    await flushPromises();
    expect(wrapper.emitted("close")).toEqual([[]]);
    expect(api.putRoutingA).not.toHaveBeenCalled();
    expect(noticeState.current?.text).toBe("Offline");
  });
});
