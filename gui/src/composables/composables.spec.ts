// @vitest-environment happy-dom
import { describe, expect, test, vi } from "vitest";
import { defineComponent, h } from "vue";
import { mount } from "@vue/test-utils";
import { useUnsavedGuard } from "./useUnsavedGuard";
import {
  closeAllNotices,
  dismissNotice,
  noticeState,
  useNotify,
} from "./useNotify";
import { closeAllDialogs, dialogState, openDialog } from "./useDialog";
import { closeAllLoadings, loadingState, openLoading } from "./useLoading";

describe("useNotify shows one notice at a time", () => {
  test("queues in order and advances on dismiss", () => {
    closeAllNotices();
    const notify = useNotify();
    const a = notify.info("a");
    const b = notify.error("b");
    expect(noticeState.current?.id).toBe(a);
    expect(noticeState.queue.map((n) => n.id)).toEqual([b]);
    dismissNotice(a);
    expect(noticeState.current?.id).toBe(b);
    expect(noticeState.current?.timeout).toBe(8000);
    dismissNotice(b);
    expect(noticeState.current).toBeNull();
  });
  test("a queued notice can be removed before it shows", () => {
    closeAllNotices();
    const notify = useNotify();
    notify.info("a");
    const b = notify.info("b");
    dismissNotice(b);
    expect(noticeState.queue).toHaveLength(0);
  });
});

describe("useDialog resolves through the handle", () => {
  test("close resolves the result and removes the entry", async () => {
    closeAllDialogs();
    const h = openDialog<string>({ render: () => null }, { x: 1 });
    expect(dialogState.stack).toHaveLength(1);
    h.close("done");
    expect(dialogState.stack).toHaveLength(0);
    await expect(h.result).resolves.toBe("done");
  });
  test("closeAll closes the child first", async () => {
    const order: number[] = [];
    const a = openDialog({ render: () => null });
    const b = openDialog({ render: () => null });
    a.result.then(() => order.push(a.id));
    b.result.then(() => order.push(b.id));
    closeAllDialogs();
    await Promise.all([a.result, b.result]);
    expect(order).toEqual([b.id, a.id]);
  });
});

describe("useLoading keeps the overlay until the last handle closes", () => {
  test("two handles, two closes", () => {
    closeAllLoadings();
    const a = openLoading();
    const b = openLoading();
    expect(loadingState.open.size).toBe(2);
    a.close();
    expect(loadingState.open.size).toBe(1);
    b.close();
    expect(loadingState.open.size).toBe(0);
  });
  test("closeAll drops every handle", () => {
    openLoading();
    closeAllLoadings();
    expect(loadingState.open.size).toBe(0);
    vi.restoreAllMocks();
  });
});

describe("useUnsavedGuard asks before a reload drops edits", () => {
  test("cancels beforeunload only while dirty, and not after unmount", () => {
    let dirty = false;
    const Guarded = defineComponent({
      setup() {
        useUnsavedGuard(() => dirty);
        return () => h("div");
      },
    });
    const wrapper = mount(Guarded);
    const fire = () => {
      const event = new Event("beforeunload", { cancelable: true });
      window.dispatchEvent(event);
      return event.defaultPrevented;
    };
    expect(fire()).toBe(false);
    dirty = true;
    expect(fire()).toBe(true);
    wrapper.unmount();
    expect(fire()).toBe(false);
  });
});
