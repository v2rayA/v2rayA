// @vitest-environment happy-dom
import { expect, test, vi } from "vitest";
import { defineComponent, h } from "vue";
import { getSetting, putSetting } from "@/api";
import { mountWithApp } from "@/test/mount";
import { defaultForm, useSettings } from "./model";

vi.mock("@/api", () => ({
  getSetting: vi.fn(),
  putSetting: vi.fn(),
}));

test("applying settings preserves edits made during the request", async () => {
  vi.mocked(getSetting).mockResolvedValue({
    setting: defaultForm(),
    localGFWListVersion: "",
    localGeositeVersion: "",
  });
  let settings!: ReturnType<typeof useSettings>;
  const wrapper = mountWithApp(
    defineComponent({
      setup() {
        settings = useSettings();
        return () => h("div");
      },
    }),
  );
  try {
    await settings.load();
    settings.form.logLevel = "debug";
    expect(settings.dirty.value).toBe(true);
    let complete!: () => void;
    vi.mocked(putSetting).mockReturnValueOnce(
      new Promise<void>((resolve) => {
        complete = resolve;
      }),
    );
    const saving = settings.save();
    settings.form.logLevel = "warn";
    complete();
    await saving;
    expect(settings.form.logLevel).toBe("warn");
    expect(settings.dirty.value).toBe(true);
  } finally {
    wrapper.unmount();
  }
});
