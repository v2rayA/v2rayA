"use strict";

import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import "dayjs/locale/zh-cn";
import "dayjs/locale/en";
import "dayjs/locale/fa";
import "dayjs/locale/ru";
import "dayjs/locale/pt-br";
import "dayjs/locale/ko";

dayjs.extend(relativeTime);
dayjs.extend(utc);
dayjs.extend(timezone);
