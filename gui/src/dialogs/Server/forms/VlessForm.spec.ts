// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { useAppStore } from "@/stores/app";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { v2rayModel, type V2rayModel } from "../models";
import VlessForm from "./VlessForm.vue";

const labels = (w: { findAll(s: string): { text(): string }[] }) =>
  w.findAll(".v-label").map((l) => l.text());

describe("the vless form", () => {
  test("shows the reality link, preserves its round trip and edits in place", async () => {
    const model = reactive(parseShareLink(links.vless) as V2rayModel);
    const w = mountWithApp(VlessForm, { props: { modelValue: model } });
    useAppStore().variant = "xray";
    await nextTick();
    expect(labels(w)).toEqual(
      expect.arrayContaining([
        "ID",
        "Network",
        "Flow",
        "SNI",
        "Public Key (pbk)",
        "Short ID (sid)",
        "Spider X (spx)",
      ]),
    );
    const fields = w.findAllComponents(VTextField);
    expect(
      fields.find((c) => c.props("label") === "Public Key (pbk)")!.get("input")
        .element.value,
    ).toBe("PUBLICKEY");
    expect(generateShareLink(model)).toBe(fixtures.vless.back);
    await fields
      .find((c) => c.props("label") === "ID")!
      .get("input")
      .setValue("new-user-id");
    expect(model.id).toBe("new-user-id");
    w.unmount();
  });

  test("shows reality keys only for reality and flow only with security", async () => {
    const model = reactive({ ...v2rayModel(), protocol: "vless" });
    const w = mountWithApp(VlessForm, { props: { modelValue: model } });
    expect(labels(w)).not.toContain("Flow");
    expect(labels(w)).not.toContain("Public Key (pbk)");
    model.tls = "reality";
    await nextTick();
    expect(labels(w)).toContain("Flow");
    expect(labels(w)).toContain("Public Key (pbk)");
    expect(labels(w)).not.toContain("ALPN");
    model.tls = "tls";
    await nextTick();
    expect(labels(w)).not.toContain("Public Key (pbk)");
    expect(labels(w)).toContain("ALPN");
    expect(labels(w)).toContain("Flow");
    model.tls = "none";
    await nextTick();
    expect(labels(w)).not.toContain("Flow");
    expect(labels(w)).not.toContain("SNI");
    w.unmount();
  });

  test("rejects each missing required identity field", async () => {
    const model = reactive({ ...v2rayModel(), protocol: "vless" });
    const Wrapped = defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(VlessForm, { modelValue: model })),
    });
    const w = mountWithApp(Wrapped);
    const form = w.findComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.find(".v-label").text()),
    ).toEqual(["Host", "Port", "ID"]);
    model.add = "example.com";
    model.port = "443";
    model.id = "b831381d-6324-4d53-ad4f-8cda48b30811";
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("keeps subscription identity and shared fields readonly", async () => {
    const model = reactive(parseShareLink(links.vless) as V2rayModel);
    const w = mountWithApp(VlessForm, {
      props: { modelValue: model, readonly: true },
    });
    expect(
      w
        .findAll('input:not([type="hidden"])')
        .every((c) => c.attributes("readonly") !== undefined),
    ).toBe(true);
    for (const select of w.findAll(".v-select")) {
      const input = select.get('input:not([type="hidden"])');
      await select.get(".v-field").trigger("mousedown");
      await input.trigger("keydown", { key: "ArrowDown" });
      expect(input.attributes("aria-expanded")).toBe("false");
    }
    expect(generateShareLink(model)).toBe(fixtures.vless.back);
    w.unmount();
  });
});
