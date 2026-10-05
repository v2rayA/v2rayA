// @vitest-environment happy-dom
import { expect, test, vi } from "vitest";
import { defineComponent, h } from "vue";
import { getSetting, getNodeDnsOptions, putSetting } from "@/api";
import { mountWithApp } from "@/test/mount";
import { defaultForm, useSettings } from "./model";

vi.mock("@/api", () => ({
  getSetting: vi.fn(),
  getNodeDnsOptions: vi.fn(),
  putSetting: vi.fn(),
}));

test("refreshing DNS options preserves edits made during an apply", async () => {
  vi.mocked(getSetting).mockResolvedValue({
    setting: defaultForm(),
    localGFWListVersion: "",
    localGeositeVersion: "",
  });
  vi.mocked(getNodeDnsOptions).mockResolvedValue({ options: [] });
  let settings!: ReturnType<typeof useSettings>;
  const wrapper = mountWithApp(
    defineComponent({
      setup() {
        settings = useSettings({ nodeDns: true });
        return () => h("div");
      },
    }),
  );
  try {
    await settings.load();
    settings.form.nodeDns = "udp://223.5.5.5:53";
    expect(settings.dirty.value).toBe(true);
    let complete!: () => void;
    vi.mocked(putSetting).mockReturnValueOnce(
      new Promise<void>((resolve) => {
        complete = resolve;
      }),
    );
    const saving = settings.save();
    settings.form.nodeDns = "auto";
    complete();
    await saving;
    expect(settings.form.nodeDns).toBe("auto");
    expect(settings.dirty.value).toBe(true);
    expect(getNodeDnsOptions).toHaveBeenCalledTimes(2);
  } finally {
    wrapper.unmount();
  }
});
