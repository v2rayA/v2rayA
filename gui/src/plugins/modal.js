import Vue from "vue";
import { ModalProgrammatic } from "buefy";
import ModalCertFix from "@/components/modalCertFix";

const registry = {
  certfix: ModalCertFix,
};

const modal = {
  /**
   * Open a registered modal by name.
   *
   * @param {string} name - registered modal name, e.g. "certfix"
   * @param {object} props - props passed to the modal component
   * @param {object} options - extra Buefy modal options (e.g. { parent })
   * @returns {object|undefined} the Buefy modal instance
   */
  show(name, props = {}, options = {}) {
    const component = registry[name];
    if (!component) {
      console.warn(`[modal] "${name}" is not registered`);
      return;
    }
    return ModalProgrammatic.open({
      component,
      hasModalCard: true,
      canCancel: false,
      ...options,
      props: {
        ...props,
        ...(options.props || {}),
      },
    });
  },
};

export default {
  install() {
    Vue.prototype.$modal = modal;
  },
};
