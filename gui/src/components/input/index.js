import Input from "./Input";

import { use, registerComponent } from "buefy/src/utils/plugins";

const Plugin = {
  install(Vue) {
    registerComponent(Vue, Input);
  },
};

use(Plugin);

export default Plugin;
