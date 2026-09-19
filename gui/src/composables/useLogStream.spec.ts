// @vitest-environment happy-dom
import { describe, expect, test, vi } from "vitest";

vi.mock("@/api", () => ({ getLogger: vi.fn() }));
import { detectLevel, detectSource, useLogStream } from "./useLogStream";

describe("the log stream", () => {
  test("a chunk ending mid-line is completed by the next chunk", () => {
    const s = useLogStream();
    s.append("2026/09/19 [I] [main.go:24] Start");
    expect(s.lines.value.map((l) => l.text)).toEqual([
      "2026/09/19 [I] [main.go:24] Start",
    ]);
    s.append("ing...\n2026/09/19 [W] [tun.go:9] warn\n");
    expect(s.lines.value.map((l) => l.text)).toEqual([
      "2026/09/19 [I] [main.go:24] Starting...",
      "2026/09/19 [W] [tun.go:9] warn",
    ]);
    expect(s.lines.value[1].level).toBe("warn");
    expect(s.sources.value).toEqual(["main.go", "tun.go"]);
  });

  test("multi-byte text advances the offset by bytes", async () => {
    const api = await import("@/api");
    const s = useLogStream();
    s.append("中文\n");
    await s.fetch();
    expect(api.getLogger).toHaveBeenCalledWith({ skip: 7 });
  });

  test("level and source detection", () => {
    expect(detectLevel("x [E] boom")).toBe("error");
    expect(detectLevel("plain")).toBe("other");
    expect(detectSource("[A] [index.go:224] listening")).toBe("index.go");
    expect(detectSource("[LoggerService] up")).toBe("LoggerService");
  });
});
