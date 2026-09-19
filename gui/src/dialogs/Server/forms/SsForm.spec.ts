// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { ssModel, type SsModel } from "../models";
import SsForm from "./SsForm.vue";

const labels = (w: { findAll(s: string): { text(): string }[] }) =>
  w.findAll(".v-label").map((l) => l.text());

describe("the ss form", () => {
  test("shows the obfs link, preserves its round trip and edits in place", async () => {
    const model = reactive(parseShareLink(links.ss) as SsModel);
    const w = mountWithApp(SsForm, { props: { modelValue: model } });
    expect(labels(w)).toEqual(
      expect.arrayContaining(["Plugin", "Implementation", "Obfs", "Path"]),
    );
    expect(labels(w)).not.toContain("TLS");
    const fields = w.findAllComponents(VTextField);
    expect(
      fields
        .filter((c) => c.props("label") === "Host")
        .map((c) => c.get("input").element.value),
    ).toEqual(["1.2.3.4", "example.com"]);
    expect(generateShareLink(model)).toBe(fixtures.ss.back);
    await fields
      .find((c) => c.props("label") === "Password")!
      .get("input")
      .setValue("new-password");
    expect(model.password).toBe("new-password");
    w.unmount();
  });

  test("matches plugin and obfuscation field conditions without losing values", async () => {
    const model = reactive(ssModel());
    const w = mountWithApp(SsForm, { props: { modelValue: model } });
    expect(labels(w)).not.toContain("Implementation");
    expect(labels(w)).not.toContain("Obfs");
    expect(labels(w)).not.toContain("Path");
    model.plugin = "simple-obfs";
    model.path = "/obfs";
    await nextTick();
    expect(labels(w)).toContain("Implementation");
    expect(labels(w)).toContain("Obfs");
    expect(labels(w)).toContain("Path");
    model.obfs = "tls";
    await nextTick();
    expect(labels(w)).not.toContain("Path");
    expect(
      w
        .findAllComponents(VTextField)
        .filter((c) => c.props("label") === "Host"),
    ).toHaveLength(2);
    model.plugin = "v2ray-plugin";
    await nextTick();
    expect(labels(w)).not.toContain("Obfs");
    expect(labels(w)).toContain("Mode");
    expect(labels(w)).toContain("TLS");
    expect(labels(w)).toContain("Path");
    expect(
      w
        .findAllComponents(VTextField)
        .find((c) => c.props("label") === "Path")!
        .get("input").element.value,
    ).toBe("/obfs");
    model.plugin = "";
    await nextTick();
    expect(labels(w)).not.toContain("Implementation");
    expect(labels(w)).not.toContain("Mode");
    expect(labels(w)).not.toContain("TLS");
    expect(labels(w)).not.toContain("Path");
    expect(
      w
        .findAllComponents(VTextField)
        .filter((c) => c.props("label") === "Host"),
    ).toHaveLength(1);
    w.unmount();
  });

  test("requires host, port, password and cipher", async () => {
    const model = reactive({ ...ssModel(), method: "" });
    const Wrapped = defineComponent({
      setup: () => () => h(VForm, null, () => h(SsForm, { modelValue: model })),
    });
    const w = mountWithApp(Wrapped);
    const form = w.findComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.find(".v-label").text()),
    ).toEqual(["Host", "Port", "Password", "Method"]);
    Object.assign(model, {
      server: "1.2.3.4",
      port: "8388",
      password: "secret",
      method: "aes-128-gcm",
    });
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("shows subscription plugin values without editable inputs", async () => {
    const model = reactive(parseShareLink(links.ss) as SsModel);
    const w = mountWithApp(SsForm, {
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
    expect(generateShareLink(model)).toBe(fixtures.ss.back);
    w.unmount();
  });
});
