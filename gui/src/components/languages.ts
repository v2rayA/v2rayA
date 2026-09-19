// The six languages the interface ships: the locale code (localStorage
// _lang and the locales/ files), the label in its own language, and the
// dayjs locale.
export const languages = [
  { code: "zh_CN", label: "中文-中国", flag: "zh", dayjs: "zh-cn" },
  { code: "en_US", label: "English-US", flag: "en", dayjs: "en" },
  { code: "fa_IR", label: "فارسی", flag: "fa", dayjs: "fa" },
  { code: "ru_RU", label: "Русский", flag: "ru", dayjs: "ru" },
  { code: "pt_BR", label: "Português-Brasil", flag: "pt", dayjs: "pt-br" },
  { code: "ko_KR", label: "한국어-대한민국", flag: "ko", dayjs: "ko" },
] as const;
