// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { ssrModel, type SsrModel } from "../models";
import SsrForm from "./SsrForm.vue";

const labels = (w: { findAll(s: string): { text(): string }[] }) =>
  w.findAll(".v-label").map((l) => l.text());

describe("the ssr form", () => {
  test("shows link parameters, preserves the round trip and edits in place", async () => {
    const model = reactive(parseShareLink(links.ssr) as SsrModel);
    const w = mountWithApp(SsrForm, { props: { modelValue: model } });
    expect(labels(w)).toEqual(
      expect.arrayContaining([
        "Protocol",
        "Method",
        "Protocol Param",
        "Obfs",
        "Obfs Param",
      ]),
    );
    const fields = w.findAllComponents(VTextField);
    expect(
      fields.find((c) => c.props("label") === "Obfs Param")!.get("input")
        .element.value,
    ).toBe("obfs.example.com");
    expect(generateShareLink(model)).toBe(fixtures.ssr.back);
    await fields
      .find((c) => c.props("label") === "Protocol Param")!
      .get("input")
      .setValue("2:def");
    expect(model.protoParam).toBe("2:def");
    w.unmount();
  });

  test("toggles protocol and obfuscation parameters independently", async () => {
    const model = reactive(ssrModel());
    const w = mountWithApp(SsrForm, { props: { modelValue: model } });
    expect(labels(w)).not.toContain("Protocol Param");
    expect(labels(w)).not.toContain("Obfs Param");
    model.proto = "auth_aes128_md5";
    await nextTick();
    expect(labels(w)).toContain("Protocol Param");
    expect(labels(w)).not.toContain("Obfs Param");
    model.obfs = "http_simple";
    await nextTick();
    expect(labels(w)).toContain("Obfs Param");
    model.proto = "origin";
    await nextTick();
    expect(labels(w)).not.toContain("Protocol Param");
    expect(labels(w)).toContain("Obfs Param");
    model.obfs = "plain";
    await nextTick();
    expect(labels(w)).not.toContain("Obfs Param");
    w.unmount();
  });

  test("requires host, port, password, cipher, protocol and obfuscation", async () => {
    const model = reactive({ ...ssrModel(), method: "", proto: "", obfs: "" });
    const Wrapped = defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(SsrForm, { modelValue: model })),
    });
    const w = mountWithApp(Wrapped);
    const form = w.findComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.find(".v-label").text()),
    ).toEqual(["Host", "Port", "Password", "Method", "Protocol", "Obfs"]);
    Object.assign(model, {
      server: "1.2.3.4",
      port: "8388",
      password: "secret",
      method: "aes-128-cfb",
      proto: "origin",
      obfs: "plain",
    });
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("shows subscription parameters without editable inputs", async () => {
    const model = reactive(parseShareLink(links.ssr) as SsrModel);
    const w = mountWithApp(SsrForm, {
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
    expect(generateShareLink(model)).toBe(fixtures.ssr.back);
    w.unmount();
  });
});
