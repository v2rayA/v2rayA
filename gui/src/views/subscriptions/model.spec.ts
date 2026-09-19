// @vitest-environment happy-dom
import { afterEach, expect, test } from "vitest";
import { client } from "@/api/client";
import type { TouchResponse, TouchSubscription } from "@/api/types";
import { useSubscriptions } from "./model";

const subscription: TouchSubscription = {
  id: 3,
  _type: "subscription",
  remarks: "Travel",
  host: "example.test",
  address: "https://example.test/subscription",
  status: "2026-09-19T12:00:00Z",
  info: "Used 1 GiB / 10 GiB · Expires 2026-10-01",
  servers: [],
  autoSelect: false,
};
const adapter = client.defaults.adapter;
afterEach(() => {
  client.defaults.adapter = adapter;
});

test("update and delete address the subscription and reload the collection", async () => {
  const sent: { method?: string; path: string; body: unknown }[] = [];
  let rows = [subscription];
  client.defaults.adapter = async (config) => {
    sent.push({
      method: config.method,
      path: new URL(config.url!, "http://backend.test").pathname,
      body: config.data ? JSON.parse(config.data) : undefined,
    });
    if (config.method === "put")
      rows = [{ ...subscription, info: "Used 2 GiB" }];
    if (config.method === "delete") rows = [];
    const data: TouchResponse = {
      running: false,
      networkPaused: false,
      touch: { subscriptions: rows, servers: [], connectedServer: [] },
    };
    return {
      config,
      status: 200,
      statusText: "OK",
      headers: {},
      data: { code: "SUCCESS", data },
    };
  };
  const model = useSubscriptions();
  await model.sync();
  await model.update(subscription);
  expect(model.subscriptions.value[0].info).toBe("Used 2 GiB");
  await model.remove(subscription);
  expect(model.subscriptions.value).toEqual([]);
  expect(sent).toEqual([
    { method: "get", path: "/api/touch", body: undefined },
    {
      method: "put",
      path: "/api/subscription",
      body: { id: 3, _type: "subscription" },
    },
    { method: "get", path: "/api/touch", body: undefined },
    {
      method: "delete",
      path: "/api/touch",
      body: { touches: [{ id: 3, _type: "subscription" }] },
    },
    { method: "get", path: "/api/touch", body: undefined },
  ]);
});
