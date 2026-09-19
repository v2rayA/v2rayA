// @vitest-environment happy-dom
import { expect, test } from "vitest";
import { mountWithApp } from "@/test/mount";
import type { Touch } from "@/api/types";
import GroupMembers from "./GroupMembers.vue";

const node = (id: number, name: string) => ({
  id,
  _type: "subscriptionServer" as const,
  name,
  address: `10.0.0.${id}:443`,
  net: "vless",
  pingLatency: "",
});
const touch: Touch = {
  servers: [],
  subscriptions: [
    {
      id: 1,
      _type: "subscription",
      host: "a",
      address: "https://a",
      status: "",
      info: "",
      autoSelect: false,
      servers: [node(1, "a1"), node(2, "a2")],
    },
    {
      id: 2,
      _type: "subscription",
      host: "b",
      address: "https://b",
      status: "",
      info: "",
      autoSelect: false,
      servers: [node(1, "b1"), node(2, "b2")],
    },
  ],
  connectedServer: [],
};

test("a member of the second subscription is shown checked and saved with its subscription", async () => {
  const wrapper = mountWithApp(GroupMembers, {
    props: {
      outbound: "proxy",
      touch,
      members: [
        { _type: "subscriptionServer", id: 2, sub: 1, outbound: "proxy" },
      ],
    },
  });
  const checked = wrapper
    .findAll(".v-list-item")
    .filter(
      (item) =>
        (item.find("input[type=checkbox]").element as HTMLInputElement).checked,
    )
    .map((item) => item.text());
  expect(checked).toHaveLength(1);
  expect(checked[0]).toContain("b2");
  await wrapper.findAll(".v-list-item")[0].trigger("click");
  await wrapper
    .find(".v-btn[color=primary], .v-btn--variant-flat")
    .trigger("click");
  const emitted = wrapper.emitted("close")?.[0]?.[0];
  expect(emitted).toEqual([
    { _type: "subscriptionServer", id: 1, sub: 0, outbound: "proxy" },
    { _type: "subscriptionServer", id: 2, sub: 1, outbound: "proxy" },
  ]);
  wrapper.unmount();
});
