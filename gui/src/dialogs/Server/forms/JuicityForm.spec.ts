// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { juicityModel, type JuicityModel } from "../models";
import JuicityForm from "./JuicityForm.vue";

describe("the juicity form", () => {
  test("shows the link's identity and security fields and edits in place", async () => {
    const model = reactive(parseShareLink(links.juicity) as JuicityModel);
    const w = mountWithApp(JuicityForm, { props: { modelValue: model } });
    await nextTick();
    expect(generateShareLink(model)).toBe(fixtures.juicity.back);
    const fields = w.findAllComponents(VTextField);
    expect(
      fields.find((c) => c.props("label") === "UUID")!.get("input").element
        .value,
    ).toBe(model.uuid);
    expect(
      fields.find((c) => c.props("label") === "SNI")!.get("input").element
        .value,
    ).toBe("example.com");
    await w.get('input[dir="ltr"]').setValue("5.6.7.8");
    expect(model.server).toBe("5.6.7.8");
    w.unmount();
  });

  test("rejects every empty required field and accepts a populated link", async () => {
    const model = reactive({ ...juicityModel(), cc: "" });
    const w = mountWithApp(
      defineComponent({
        setup: () => () =>
          h(VForm, null, () => h(JuicityForm, { modelValue: model })),
      }),
    );
    const form = w.getComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.get(".v-label").text()),
    ).toEqual(["Host", "Port", "UUID", "Password", "Congestion Control"]);
    Object.assign(model, parseShareLink(links.juicity));
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("subscription credentials and the insecure switch are read only", async () => {
    const model = reactive(parseShareLink(links.juicity) as JuicityModel);
    const w = mountWithApp(JuicityForm, {
      props: { modelValue: model, readonly: true },
    });
    for (const input of w.findAll<HTMLInputElement>(
      'input:not([type="checkbox"]):not([type="hidden"])',
    ))
      expect(input.element.readOnly).toBe(true);
    expect(generateShareLink(model)).toBe(fixtures.juicity.back);
    w.unmount();
  });
});
