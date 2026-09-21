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
import { anytlsModel, type AnytlsModel } from "../models";
import AnytlsForm from "./AnytlsForm.vue";

const field = (w: VueWrapper, label: string) =>
  w.findAllComponents(VTextField).find((c) => c.props("label") === label)!;

const labels = en.configureServer;

describe("the anytls form", () => {
  test("shows the fixture, preserves its link and edits the model in place", async () => {
    const model = reactive(parseShareLink(links.anytls) as AnytlsModel);
    const w = mountWithApp(AnytlsForm, { props: { modelValue: model } });
    await nextTick();
    for (const [label, value] of [
      [labels.servername, "anytls node"],
      [labels.host, "1.2.3.4"],
      [labels.port, "443"],
      [labels.auth, "passw0rd"],
      ["SNI(Peer)", "example.com"],
      [en.pinnedPeerCertSha256, ""],
      [en.verifyPeerCertByName, ""],
      [labels.minIdleSession, ""],
    ]) {
      expect(field(w, label).get("input").element.value).toBe(value);
    }
    expect(generateShareLink(model)).toBe(fixtures.anytls.back);
    await field(w, labels.auth).get("input").setValue("new-secret");
    expect(model.auth).toBe("new-secret");
    w.unmount();
  });

  test("requires host, port and authentication but leaves TLS tuning optional", async () => {
    const model = reactive(anytlsModel());
    const Wrapped = defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(AnytlsForm, { modelValue: model })),
    });
    const w = mountWithApp(Wrapped);
    const form = w.findComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w
        .findAllComponents(VTextField)
        .filter((c) => c.classes("v-input--error"))
        .map((c) => c.props("label")),
    ).toEqual([labels.host, labels.port, labels.auth]);
    await field(w, labels.host).get("input").setValue("proxy.example.com");
    await field(w, labels.port).get("input").setValue("443");
    await field(w, labels.auth).get("input").setValue("secret");
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("keeps subscription values read-only and prevents security changes", async () => {
    const model = reactive(parseShareLink(links.anytls) as AnytlsModel);
    const w = mountWithApp(AnytlsForm, {
      props: { modelValue: model, readonly: true },
    });
    for (const input of w.findAll('input:not([type="checkbox"])'))
      expect((input.element as HTMLInputElement).readOnly).toBe(true);
    expect(generateShareLink(model)).toBe(fixtures.anytls.back);
    w.unmount();
  });
});
