// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import type { VueWrapper } from "@vue/test-utils";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import en from "@/locales/en";
import { socks5Model, type Socks5Model } from "../models";
import Socks5Form from "./Socks5Form.vue";

const field = (w: VueWrapper, label: string) =>
  w.findAllComponents(VTextField).find((c) => c.props("label") === label)!;

const labels = en.configureServer;

describe("the socks5 form", () => {
  test("shows the fixture, preserves its link and edits the model in place", async () => {
    const model = reactive(parseShareLink(links.socks5) as Socks5Model);
    const w = mountWithApp(Socks5Form, { props: { modelValue: model } });
    await nextTick();
    for (const [label, value] of [
      [labels.servername, "socks node"],
      [labels.host, "1.2.3.4"],
      [labels.port, "1080"],
      [labels.username, "user"],
      [labels.password, "passw0rd"],
    ]) {
      expect(field(w, label).get("input").element.value).toBe(value);
    }
    expect(generateShareLink(model)).toBe(fixtures.socks5.back);
    await field(w, labels.host).get("input").setValue("proxy.example.com");
    expect(model.host).toBe("proxy.example.com");
    w.unmount();
  });

  test("requires host and port but permits anonymous authentication", async () => {
    const model = reactive(socks5Model());
    const Wrapped = defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(Socks5Form, { modelValue: model })),
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
    await field(w, labels.port).get("input").setValue("1080");
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("keeps subscription values visible and read-only", () => {
    const model = reactive(parseShareLink(links.socks5) as Socks5Model);
    const w = mountWithApp(Socks5Form, {
      props: { modelValue: model, readonly: true },
    });
    for (const input of w.findAll("input"))
      expect(input.element.readOnly).toBe(true);
    expect(field(w, labels.password).get("input").element.value).toBe(
      "passw0rd",
    );
    expect(generateShareLink(model)).toBe(fixtures.socks5.back);
    w.unmount();
  });
});
