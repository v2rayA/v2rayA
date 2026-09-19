// mountWithApp mounts a component the way the page does: inside a Vuetify
// layout (app bars, FABs and drawers need it), with Vuetify, Pinia and
// vue-i18n (English). For specs under happy-dom, which lacks the layout
// observers Vuetify's components register.
import {
  mount,
  type ComponentMountingOptions,
  type VueWrapper,
} from "@vue/test-utils";
import { createPinia, setActivePinia } from "pinia";
import { createI18n } from "vue-i18n";
import { defineComponent, h, nextTick, reactive, type Component } from "vue";
import { VLayout } from "vuetify/components";
import messages from "@/locales";
import { vuetify } from "@/theme";

class Observer {
  observe() {}
  unobserve() {}
  disconnect() {}
}
globalThis.ResizeObserver ??= Observer as unknown as typeof ResizeObserver;
globalThis.IntersectionObserver ??=
  Observer as unknown as typeof IntersectionObserver;
window.matchMedia ??= ((query: string) =>
  ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener() {},
    removeEventListener() {},
    addListener() {},
    removeListener() {},
    dispatchEvent: () => false,
  }) as MediaQueryList) as typeof window.matchMedia;
// menus and tooltips read the visual viewport to place themselves
(window as { visualViewport?: unknown }).visualViewport ??= {
  width: 1280,
  height: 900,
  offsetLeft: 0,
  offsetTop: 0,
  scale: 1,
  addEventListener() {},
  removeEventListener() {},
};

/** mountWithApp returns the wrapper of `component` itself; unmount() tears the whole tree down. */
export function mountWithApp<C extends Component>(
  component: C,
  options: ComponentMountingOptions<C> = {},
): VueWrapper {
  const pinia = createPinia();
  setActivePinia(pinia);
  const i18n = createI18n({ legacy: false, locale: "en", messages });
  // the props stay reactive so a spec can change them with setProps
  const props = reactive({
    ...(options.props as Record<string, unknown> | undefined),
    ...(options.attrs as Record<string, unknown> | undefined),
  });
  const Host = defineComponent({
    setup: () => () =>
      h(VLayout, null, () => h(component as Component, { ...props })),
  });
  const root = mount(Host, {
    global: {
      ...options.global,
      plugins: [vuetify, pinia, i18n, ...(options.global?.plugins ?? [])],
    },
    attachTo: document.body,
  });
  const inner = root.findComponent(component as Component) as VueWrapper;
  Object.defineProperty(inner, "unmount", { value: () => root.unmount() });
  Object.defineProperty(inner, "setProps", {
    value: async (next: Record<string, unknown>) => {
      Object.assign(props, next);
      await nextTick();
    },
  });
  return inner;
}
