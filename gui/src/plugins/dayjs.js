import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import Vue from "vue";
import "dayjs/locale/zh-cn";
import "dayjs/locale/en";

dayjs.extend(relativeTime);
Vue.prototype.$dayjs = dayjs;
