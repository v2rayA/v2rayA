// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { onSessionTeardown, resetSession, setSessionStarter } from "./session";
import { openDialog, dialogState } from "./composables/useDialog";
import { useNotify, noticeState } from "./composables/useNotify";
import { openLoading, loadingState } from "./composables/useLoading";
import { useAppStore } from "./stores/app";

describe("resetSession runs the teardown in order", () => {
  test("teardown, then what is on screen, then the stores, then the starter", async () => {
    const pinia = createPinia();
    setActivePinia(pinia);
    const order: string[] = [];
    onSessionTeardown(() => order.push("teardown"));
    openDialog({ render: () => null }).result.then(() =>
      order.push("dialog closed"),
    );
    useNotify().info("x");
    openLoading();
    const app = useAppStore(pinia);
    app.setRunning("running");
    setSessionStarter(async () => {
      order.push(`start:${app.token}@${app.backendAddress}:${app.running}`);
    });
    await resetSession({ token: "tok-2", backendAddress: "http://b" }, pinia);
    expect(order).toEqual([
      "teardown",
      "dialog closed",
      "start:tok-2@http://b:checking",
    ]);
    expect(dialogState.stack).toHaveLength(0);
    expect(noticeState.current).toBeNull();
    expect(loadingState.open.size).toBe(0);
    expect(localStorage.getItem("token")).toBe("tok-2");
  });
});
