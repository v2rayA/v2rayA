<script>
import { h } from "vue";
import hljs from "highlight.js/lib/core";
import "highlight.js/styles/github.css";
import accesslog from "highlight.js/lib/languages/accesslog";
hljs.registerLanguage("accesslog", accesslog);
export default {
  name: "HightlightLog",
  props: {
    text: {
      type: String,
      default: "",
    },
  },
  render() {
    return h("div", {
      class: "log language-accesslog",
      innerHTML: hljs.highlight(this.text, {
        language: "accesslog",
        ignoreIllegals: true,
      }).value,
    });
  },
};
</script>

<style scoped>
.log {
  font-family: Consolas, monospace, Monaco, Menlo, Courier New;
}
</style>
