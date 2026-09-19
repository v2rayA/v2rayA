// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VSelect, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import en from "@/locales/en";
import { httpModel, type HttpModel } from "../models";
import HttpForm from "./HttpForm.vue";

const field = (w: VueWrapper, label: string) =>
  w.findAllComponents(VTextField).find((c) => c.props("label") === label)!;

const labels = en.configureServer;

describe("the http form", () => {
  beforeEach(() => {
    // Happy DOM has no visual viewport; Vuetify also supports its absence.
    vi.stubGlobal("visualViewport", undefined);
  });
  afterEach(() => vi.unstubAllGlobals());

  test("shows the fixture, preserves its link and edits the model in place", async () => {
    const model = reactive(parseShareLink(links.http) as HttpModel);
    const w = mountWithApp(HttpForm, { props: { modelValue: model } });
    await nextTick();
    for (const [label, value] of [
      [labels.servername, "http node"],
      [labels.host, "1.2.3.4"],
      [labels.port, "8080"],
      [labels.username, "user"],
      [labels.password, "passw0rd"],
    ]) {
      expect(field(w, label).get("input").element.value).toBe(value);
    }
    expect(w.findComponent(VSelect).text()).toContain("HTTP");
    expect(generateShareLink(model)).toBe(fixtures.http.back);
    await field(w, labels.host).get("input").setValue("proxy.example.com");
    expect(model.host).toBe("proxy.example.com");
    w.unmount();
  });

  test("offers HTTP and HTTPS and applies the selected protocol", async () => {
    const model = reactive(parseShareLink(links.http) as HttpModel);
    const w = mountWithApp(HttpForm, { props: { modelValue: model } });
    await w
      .findComponent(VSelect)
      .get("input")
      .trigger("keydown", { key: "Enter" });
    const options = Array.from(
      document.querySelectorAll<HTMLElement>('[role="option"]'),
    );
    expect(options.map((option) => option.textContent)).toEqual([
      "HTTP",
      "HTTPS",
    ]);
    options[1].click();
    await nextTick();
    expect(model.protocol).toBe("https");
    expect(generateShareLink(model)).toMatch(/^https-proxy:\/\//);
    w.unmount();
  });

  test("requires host and port but permits anonymous authentication", async () => {
    const model = reactive(httpModel());
    const Wrapped = defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(HttpForm, { modelValue: model })),
    });
    const w = mountWithApp(Wrapped);
    const form = w.findComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w
        .findAllComponents(VTextField)
        .filter((c) => c.classes("v-input--error"))
        .map((c) => c.props("label")),
    ).toEqual([labels.host, labels.port]);
    await field(w, labels.host).get("input").setValue("proxy.example.com");
    await field(w, labels.port).get("input").setValue("8080");
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("keeps subscription values read-only and prevents protocol selection", async () => {
    const model = reactive(parseShareLink(links.http) as HttpModel);
    const w = mountWithApp(HttpForm, {
      props: { modelValue: model, readonly: true },
    });
    for (const input of w.findAll('input:not([type="hidden"])'))
      expect((input.element as HTMLInputElement).readOnly).toBe(true);
    await w
      .findComponent(VSelect)
      .get("input")
      .trigger("keydown", { key: "Enter" });
    expect(document.querySelector('[role="listbox"]')).toBeNull();
    expect(generateShareLink(model)).toBe(fixtures.http.back);
    w.unmount();
  });
});
