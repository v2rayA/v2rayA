// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { generateURL, parseURL } from "./url";
import { links } from "./__fixtures__/links";

describe("share-link URL parts", () => {
  test("preserves the custom scheme and encoded identity and fragment", () => {
    expect(parseURL(links.wireguard)).toMatchObject({
      protocol: "wireguard",
      username: "PRIVATEKEY%3D",
      host: "1.2.3.4",
      port: 51820,
      hash: "wg%20node",
      params: {
        publicKey: "PUBKEY=",
        address: "10.0.0.2/32",
        preSharedKey: "",
        allowedIPs: "0.0.0.0/0",
      },
    });
  });

  test("decodes structured query values once", () => {
    expect(parseURL(links.ss).params.plugin).toBe(
      "obfs-local;obfs=http;obfs-host=example.com",
    );
    expect(parseURL(links.vlessXhttp).params.xhttpHeaders).toBe('{"X-A":"b"}');
    expect(parseURL(links.trojan).params.path).toBe("/tr");
  });

  test("keeps credentials separate from host and port", () => {
    expect(parseURL(links.juicity)).toMatchObject({
      username: "b831381d-6324-4d53-ad4f-8cda48b30811",
      password: "passw0rd",
      protocol: "juicity",
      host: "1.2.3.4",
      port: 443,
    });
    expect(parseURL(links.http)).toMatchObject({
      username: "user",
      password: "passw0rd",
      port: 8080,
    });
  });

  test("uses the last repeated query value and retains literal plus signs", () => {
    expect(
      parseURL("tuic://user@host:443?alpn=h3&alpn=h2&key=a+b").params,
    ).toEqual({
      alpn: "h2",
      key: "a+b",
    });
  });

  test("generates custom URLs with query escaping and an unescaped name", () => {
    expect(
      generateURL({
        protocol: "wireguard",
        username: "PRIVATEKEY=",
        host: "1.2.3.4",
        port: 51820,
        params: { publicKey: "PUBKEY=", address: "10.0.0.2/32" },
        hash: "wg node",
      }),
    ).toBe(
      "wireguard://PRIVATEKEY%3D@1.2.3.4:51820/?publicKey=PUBKEY%3D&address=10.0.0.2%2F32#wg node",
    );
  });
});
