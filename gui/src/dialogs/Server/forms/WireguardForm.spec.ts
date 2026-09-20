// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm, VTextField } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import fixtures from "@/lib/__fixtures__/serverCodec.json";
import { wireguardModel, type WireguardModel } from "../models";
import WireguardForm from "./WireguardForm.vue";

describe("the wireguard form", () => {
  test("shows every peer field, preserves the round trip and edits in place", async () => {
    const model = reactive(parseShareLink(links.wireguard) as WireguardModel);
    const w = mountWithApp(WireguardForm, { props: { modelValue: model } });
    const fields = w.findAllComponents(VTextField);
    expect(fields.map((c) => c.props("label"))).toEqual([
      "Name",
      "Address",
      "Port",
      "Public Key",
      "Private Key",
      "Address (Local)",
      "MTU",
      "Allowed IPs",
      "Persistent Keepalive",
      "Pre-shared Key",
      "Reserved bytes",
      "Workers",
    ]);
    expect(
      fields.find((c) => c.props("label") === "Address (Local)")!.get("input")
        .element.value,
    ).toBe("10.0.0.2/32");
    expect(
      fields.find((c) => c.props("label") === "Private Key")!.get("input")
        .element.value,
    ).toBe("PRIVATEKEY=");
    expect(generateShareLink(model)).toBe(fixtures.wireguard.back);
    await fields
      .find((c) => c.props("label") === "Address")!
      .get("input")
      .setValue("5.6.7.8");
    expect(model.address).toBe("5.6.7.8");
    w.unmount();
  });

  test("requires the server address, port and both keys", async () => {
    const model = reactive(wireguardModel());
    const Wrapped = defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(WireguardForm, { modelValue: model })),
    });
    const w = mountWithApp(Wrapped);
    const form = w.findComponent(VForm);
    expect((await form.vm.validate()).valid).toBe(false);
    expect(
      w.findAll(".v-input--error").map((c) => c.find(".v-label").text()),
    ).toEqual(["Address", "Port", "Public Key", "Private Key"]);
    Object.assign(model, {
      address: "1.2.3.4",
      port: "51820",
      publicKey: "PUBLICKEY=",
      privateKey: "PRIVATEKEY=",
    });
    await nextTick();
    expect((await form.vm.validate()).valid).toBe(true);
    w.unmount();
  });

  test("shows subscription values without editable inputs", () => {
    const model = reactive(parseShareLink(links.wireguard) as WireguardModel);
    const w = mountWithApp(WireguardForm, {
      props: { modelValue: model, readonly: true },
    });
    expect(
      w
        .findAll("input[type=text], input[type=number]")
        .every((c) => c.attributes("readonly") !== undefined),
    ).toBe(true);
    expect(generateShareLink(model)).toBe(fixtures.wireguard.back);
    w.unmount();
  });
});
