"use strict";

import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import Vue from "vue";
import "dayjs/locale/zh-cn";
import "dayjs/locale/en";

dayjs.extend(relativeTime);
dayjs.extend(utc);
dayjs.extend(timezone);
Vue.prototype.$dayjs = dayjs;
