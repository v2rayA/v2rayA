// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VSelect, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { hysteria2Model, type Hysteria2Model } from "../models";
import Hysteria2Form from "./Hysteria2Form.vue";

const labels = (w: { findAll(s: string): { text(): string }[] }) =>
  w.findAll(".v-label").map((l) => l.text());

describe("the hysteria2 form", () => {
  test("shows the obfuscated link, preserves its round trip and edits in place", async () => {
    const model = reactive(parseShareLink(links.hysteria2) as Hysteria2Model);
    const w = mountWithApp(Hysteria2Form, { props: { modelValue: model } });
    await nextTick();
    expect(generateShareLink(model)).toBe(fixtures.hysteria2.back);
    const fields = w.findAllComponents(VTextField);
    expect(
      fields.find((c) => c.props("label") === "Obfs Password")!.get("input")
        .element.value,
    ).toBe("obf");
    expect(
      fields.find((c) => c.props("label") === "SNI")!.get("input").element
        .value,
    ).toBe("example.com");
    await w.get('input[dir="ltr"]').setValue("5.6.7.8");
    expect(model.server).toBe("5.6.7.8");
    w.unmount();
  });

  test("obfuscation reveals its password and preserves it while hidden", async () => {
    const model = reactive(parseShareLink(links.hysteria2) as Hysteria2Model);
    const w = mountWithApp(Hysteria2Form, { props: { modelValue: model } });
    const select = w.getComponent(VSelect);
    expect(labels(w)).toContain("Obfs Password");
    select.vm.$emit("update:modelValue", "none");
    await nextTick();
    expect(labels(w)).not.toContain("Obfs Password");
    select.vm.$emit("update:modelValue", "salamander");
    await nextTick();
    expect(labels(w)).toContain("Obfs Password");
    expect(
      w
        .findAllComponents(VTextField)
        .find((c) => c.props("label") === "Obfs Password")!
        .get("input").element.value,
    ).toBe("obf");
    w.unmount();
  });

  test("rejects empty required fields without requiring the obfuscation password", async () => {
    const model = reactive({ ...hysteria2Model(), obfs: "" });
    const w = mountWithApp(
      defineComponent({
        setup: () => () =>
          h(VForm, null, () => h(Hysteria2Form, { modelValue: model })),
      }),
    );
    const form = w.getComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.get(".v-label").text()),
    ).toEqual(["Host", "Port", "Password", "Obfs"]);
    Object.assign(model, parseShareLink(links.hysteria2), { obfsPassword: "" });
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("subscription fields cannot be edited or open the obfuscation menu", async () => {
    const model = reactive(parseShareLink(links.hysteria2) as Hysteria2Model);
    const w = mountWithApp(Hysteria2Form, {
      props: { modelValue: model, readonly: true },
    });
    for (const input of w.findAll<HTMLInputElement>(
      'input:not([type="hidden"])',
    ))
      expect(input.element.readOnly).toBe(true);
    const select = w.getComponent(VSelect);
    await select.get(".v-field").trigger("mousedown");
    expect(select.get("[aria-expanded]").attributes("aria-expanded")).toBe(
      "false",
    );
    expect(generateShareLink(model)).toBe(fixtures.hysteria2.back);
    w.unmount();
  });
});
