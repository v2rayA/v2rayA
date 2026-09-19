// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { defineComponent, h, nextTick, reactive } from "vue";
import { VForm } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { noticeState } from "@/composables/useNotify";
import { parseShareLink } from "@/lib/serverCodec";
import { links } from "@/lib/__fixtures__/links";
import { v2rayModel, type V2rayModel } from "../models";
import VmessForm from "./VmessForm.vue";

const labels = (w: { findAll(s: string): { text(): string }[] }) =>
  w.findAll(".v-label").map((l) => l.text());

describe("the vmess form", () => {
  test("shows a link's fields and edits the model in place", async () => {
    const model = reactive(parseShareLink(links.vmess) as V2rayModel);
    const w = mountWithApp(VmessForm, { props: { modelValue: model } });
    const host = w.find('input[dir="ltr"]');
    await host.setValue("5.6.7.8");
    expect(model.add).toBe("5.6.7.8");
    // the link is a WebSocket node: host and path are up, gRPC's are not
    expect(labels(w)).toContain("Path");
    expect(labels(w)).not.toContain("Service Name");
    w.unmount();
  });

  test("the network decides the fields; gRPC without TLS turns TLS on", async () => {
    const model = reactive(v2rayModel());
    const w = mountWithApp(VmessForm, { props: { modelValue: model } });
    expect(labels(w)).toContain("Type");
    expect(labels(w)).not.toContain("Path");
    model.net = "grpc";
    // the change handler runs on the user's pick, not on a model write
    const select = w
      .findAllComponents({ name: "VSelect" })
      .find((c) => c.props("label") === "Network");
    select!.vm.$emit("update:modelValue", "grpc");
    await nextTick();
    expect(model.tls).toBe("tls");
    expect(model.type).toBe("none");
    expect(noticeState.current?.text).toMatch(/gRPC/);
    expect(labels(w)).toContain("Service Name");
    expect(labels(w)).toContain("SNI");
    w.unmount();
  });

  test("host, port and id are required", async () => {
    const model = reactive(v2rayModel());
    const Wrapped = defineComponent({
      setup: () => () =>
        h(VForm, null, () => h(VmessForm, { modelValue: model })),
    });
    const w = mountWithApp(Wrapped);
    const form = w.findComponent(VForm);
    const { valid } = await (
      form.vm as unknown as { validate(): Promise<{ valid: boolean }> }
    ).validate();
    expect(valid).toBe(false);
    expect(w.text()).toContain("Required");
    w.unmount();
  });
});
