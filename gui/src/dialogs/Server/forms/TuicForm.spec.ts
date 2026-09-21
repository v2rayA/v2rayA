// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VSelect, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { tuicModel, type TuicModel } from "../models";
import TuicForm from "./TuicForm.vue";

const labels = (w: { findAll(s: string): { text(): string }[] }) =>
  w.findAll(".v-label").map((l) => l.text());

describe("the tuic form", () => {
  test("shows the link's ALPN and SNI, preserves its round trip and edits in place", async () => {
    const model = reactive(parseShareLink(links.tuic) as TuicModel);
    const w = mountWithApp(TuicForm, { props: { modelValue: model } });
    await nextTick();
    expect(generateShareLink(model)).toBe(fixtures.tuic.back);
    const fields = w.findAllComponents(VTextField);
    expect(
      fields.find((c) => c.props("label") === "ALPN")!.get("input").element
        .value,
    ).toBe("h3");
    expect(
      fields.find((c) => c.props("label") === "SNI")!.get("input").element
        .value,
    ).toBe("example.com");
    expect(
      w
        .findAllComponents(VSelect)
        .find((c) => c.props("label") === "UDP Relay Mode")!
        .text(),
    ).toContain("native");
    await w.get('input[dir="ltr"]').setValue("5.6.7.8");
    expect(model.server).toBe("5.6.7.8");
    w.unmount();
  });

  test("disabling SNI hides the field without losing its value", async () => {
    const model = reactive(parseShareLink(links.tuic) as TuicModel);
    const w = mountWithApp(TuicForm, { props: { modelValue: model } });
    const select = w
      .findAllComponents(VSelect)
      .find((c) => c.props("label") === "Disable SNI")!;
    expect(labels(w)).toContain("SNI");
    select.vm.$emit("update:modelValue", true);
    await nextTick();
    expect(labels(w)).not.toContain("SNI");
    select.vm.$emit("update:modelValue", false);
    await nextTick();
    expect(labels(w)).toContain("SNI");
    expect(
      w
        .findAllComponents(VTextField)
        .find((c) => c.props("label") === "SNI")!
        .get("input").element.value,
    ).toBe("example.com");
    w.unmount();
  });

  test("rejects empty required fields but accepts false for Disable SNI", async () => {
    const model = reactive({ ...tuicModel(), cc: "", udpRelayMode: "" });
    const w = mountWithApp(
      defineComponent({
        setup: () => () =>
          h(VForm, null, () => h(TuicForm, { modelValue: model })),
      }),
    );
    const form = w.getComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.get(".v-label").text()),
    ).toEqual([
      "Host",
      "Port",
      "UUID",
      "Password",
      "Congestion Control",
      "UDP Relay Mode",
    ]);
    Object.assign(model, parseShareLink(links.tuic));
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    const select = w
      .findAllComponents(VSelect)
      .find((c) => c.props("label") === "Disable SNI")!;
    select.vm.$emit("update:modelValue", null);
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(false);
    expect(w.get(".v-input--error .v-label").text()).toBe("Disable SNI");
    w.unmount();
  });

  test("subscription fields and the insecure switch are read only", async () => {
    const model = reactive(parseShareLink(links.tuic) as TuicModel);
    const w = mountWithApp(TuicForm, {
      props: { modelValue: model, readonly: true },
    });
    for (const input of w.findAll<HTMLInputElement>(
      'input:not([type="checkbox"]):not([type="hidden"])',
    ))
      expect(input.element.readOnly).toBe(true);
    expect(generateShareLink(model)).toBe(fixtures.tuic.back);
    w.unmount();
  });
});
