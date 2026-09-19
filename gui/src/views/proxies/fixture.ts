import type { TouchResponse, TouchServer } from "@/api/types";
const server = (id: number, name: string, net = "vmess"): TouchServer => ({
  id,
  name,
  net,
  _type: "server",
  address: `${name.toLowerCase()}.example:443`,
  pingLatency: "",
});
export function fixture(): TouchResponse {
  return {
    running: true,
    networkPaused: false,
    touch: {
      servers: [server(1, "North"), server(2, "South", "trojan")],
      subscriptions: [0, 1].map((id) => ({
        id: id + 1,
        _type: "subscription",
        host: `feed-${id}.example`,
        address: `https://feed-${id}.example`,
        status: "2026-09-19T00:00:00Z",
        info: "Used 1.00 GiB / 4.00 GiB · Expires 2026-12-31",
        autoSelect: false,
        servers: [
          {
            ...server(1, id === 0 ? "West" : "East", "vless"),
            _type: "subscriptionServer",
            sub: id,
          },
        ],
      })),
      connectedServer: [
        { id: 1, _type: "server", outbound: "media" },
        { id: 1, _type: "subscriptionServer", sub: 0, outbound: "media" },
        { id: 2, _type: "server", outbound: "other" },
        { id: 1, _type: "subscriptionServer", sub: 1 },
      ],
    },
  };
}
