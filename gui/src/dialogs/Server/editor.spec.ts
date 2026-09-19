// @vitest-environment happy-dom
import { beforeEach, describe, expect, test, vi } from "vitest";
import { links } from "@/lib/__fixtures__/links";
import expected from "@/lib/__fixtures__/serverCodec.json";

const api = vi.hoisted(() => ({
  getSharingAddress: vi.fn(),
  postImport: vi.fn(),
}));
vi.mock("@/api", () => api);

import { createEditor, modelKey, protocolOf, protocols } from "./editor";
import { timeouts } from "@/api/client";

const which = { _type: "server" as const, id: 3 };

describe("the editor's tab and model", () => {
  beforeEach(() => {
    api.getSharingAddress.mockReset();
    api.postImport.mockReset().mockResolvedValue({});
  });

  test("every tab has a model, vmess and vless share one", () => {
    const editor = createEditor(null);
    for (const p of protocols) expect(editor.models[modelKey(p)]).toBeTruthy();
    expect(modelKey("vmess")).toBe("v2ray");
    expect(modelKey("vless")).toBe("v2ray");
  });

  test("a scheme picks its tab; hy2 and trojan-go are aliases", () => {
    expect(protocolOf("HY2://x")).toBe("hysteria2");
    expect(protocolOf("trojan-go://x")).toBe("trojan");
    expect(protocolOf("https://x")).toBe("http");
    expect(protocolOf("gopher://x")).toBeNull();
  });

  for (const [name, link] of Object.entries(links)) {
    test(`${name}: load, then save unchanged sends the same link the old editor did`, async () => {
      api.getSharingAddress.mockResolvedValue({ sharingAddress: link });
      const editor = createEditor(which);
      expect(await editor.load()).toBe(true);
      expect(api.getSharingAddress).toHaveBeenCalledWith(which);
      const p = protocolOf(link)!;
      expect(editor.protocol.value).toBe(p);
      expect(editor.models[modelKey(p)]).toEqual(
        expected[name as keyof typeof expected].form,
      );
      await editor.save();
      expect(api.postImport).toHaveBeenCalledWith(
        {
          url: expected[name as keyof typeof expected].back,
          kind: "server",
          which,
        },
        timeouts.none,
      );
    });
  }

  test("a new node saves with which: null and no time limit", async () => {
    const editor = createEditor(null);
    expect(await editor.load()).toBe(true);
    expect(api.getSharingAddress).not.toHaveBeenCalled();
    editor.protocol.value = "socks5";
    Object.assign(editor.models.socks5, { host: "1.2.3.4", port: "1080" });
    await editor.save();
    const [body, timeout] = api.postImport.mock.calls[0];
    expect(body.which).toBeNull();
    expect(body.url.startsWith("socks5://")).toBe(true);
    expect(timeout).toBe(timeouts.none);
  });

  test("switching between vmess and vless keeps the fields and sets the protocol", async () => {
    api.getSharingAddress.mockResolvedValue({ sharingAddress: links.vmess });
    const editor = createEditor(which);
    await editor.load();
    editor.protocol.value = "vless";
    expect(editor.link()!.startsWith("vless://")).toBe(true);
    expect(editor.models.v2ray.protocol).toBe("vless");
    expect(editor.models.v2ray.add).toBe("1.2.3.4");
  });

  test("a link the editor does not know loads as false", async () => {
    api.getSharingAddress.mockResolvedValue({ sharingAddress: "gopher://x" });
    const editor = createEditor(which);
    expect(await editor.load()).toBe(false);
    expect(editor.protocol.value).toBe("vmess");
  });
});
