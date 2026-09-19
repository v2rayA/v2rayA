// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VSelect, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { trojanModel, type TrojanModel } from "../models";
import TrojanForm from "./TrojanForm.vue";

const labels = (w: { findAll(s: string): { text(): string }[] }) =>
  w.findAll(".v-label").map((l) => l.text());

const mountForm = (model: TrojanModel) =>
  mountWithApp(
    defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(TrojanForm, { modelValue: model })),
    }),
  );

describe("the trojan form", () => {
  test("shows the WebSocket link, preserves its round trip and edits in place", async () => {
    const model = reactive(parseShareLink(links.trojan) as TrojanModel);
    const w = mountWithApp(TrojanForm, { props: { modelValue: model } });
    await nextTick();
    expect(generateShareLink(model)).toBe(fixtures.trojan.back);
    const path = w
      .findAllComponents(VTextField)
      .find((c) => c.props("label") === "Path")!;
    expect(path.get("input").element.value).toBe("/tr");
    expect(labels(w)).toContain("SNI(Peer)");
    expect(labels(w)).not.toContain("Service Name");
    await w.get('input[dir="ltr"]').setValue("5.6.7.8");
    expect(model.server).toBe("5.6.7.8");
    w.unmount();
  });

  test("only validates Shadowsocks credentials while the method is selected", async () => {
    const model = reactive({
      ...trojanModel(),
      ...parseShareLink(links.trojan),
    });
    const w = mountForm(model);
    const form = w.getComponent(VForm);
    expect(labels(w)).not.toContain("Shadowsocks Password");
    model.method = "shadowsocks";
    model.ssCipher = "";
    await nextTick();
    expect(labels(w)).toContain("Shadowsocks Cipher");
    expect(labels(w)).toContain("Shadowsocks Password");
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.get(".v-label").text()),
    ).toEqual(["Shadowsocks Cipher", "Shadowsocks Password"]);
    model.ssCipher = "aes-128-gcm";
    model.ssPassword = "secret";
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    model.method = "origin";
    model.ssCipher = "";
    model.ssPassword = "";
    await nextTick();
    expect(labels(w)).not.toContain("Shadowsocks Cipher");
    expect(labels(w)).not.toContain("Shadowsocks Password");
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("network and obfuscation reveal their own host and path fields", async () => {
    const model = reactive(trojanModel());
    const w = mountWithApp(TrojanForm, { props: { modelValue: model } });
    expect(labels(w)).not.toContain("WebSocket Host");
    model.obfs = "websocket";
    await nextTick();
    expect(labels(w)).toContain("WebSocket Host");
    expect(labels(w)).toContain("WebSocket Path");
    model.obfs = "none";
    model.net = "kcp";
    await nextTick();
    expect(labels(w)).not.toContain("WebSocket Host");
    expect(labels(w)).not.toContain("WebSocket Path");
    expect(labels(w)).toContain("Seed");
    model.net = "grpc";
    await nextTick();
    expect(labels(w)).not.toContain("Seed");
    expect(labels(w)).toContain("Service Name");
    model.net = "h2";
    await nextTick();
    expect(labels(w)).not.toContain("Service Name");
    expect(labels(w)).toContain("Path");
    model.net = "tcp";
    await nextTick();
    expect(labels(w)).not.toContain("Path");
    w.unmount();
  });

  test("rejects every empty required field and accepts a populated link", async () => {
    const model = reactive({ ...trojanModel(), method: "", net: "", obfs: "" });
    const w = mountForm(model);
    const form = w.getComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.get(".v-label").text()),
    ).toEqual(["Host", "Port", "Password", "Protocol", "Network", "Obfs"]);
    Object.assign(model, parseShareLink(links.trojan));
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("subscription fields cannot be edited or open a select menu", async () => {
    const model = reactive(parseShareLink(links.trojan) as TrojanModel);
    const w = mountWithApp(TrojanForm, {
      props: { modelValue: model, readonly: true },
    });
    for (const input of w.findAll<HTMLInputElement>(
      'input:not([type="hidden"])',
    ))
      expect(input.element.readOnly).toBe(true);
    const select = w.findComponent(VSelect);
    await select.get(".v-field").trigger("mousedown");
    expect(select.get("[aria-expanded]").attributes("aria-expanded")).toBe(
      "false",
    );
    expect(generateShareLink(model)).toBe(fixtures.trojan.back);
    w.unmount();
  });
});
