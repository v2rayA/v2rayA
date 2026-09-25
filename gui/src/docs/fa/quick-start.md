# شروع سریع

این صفحه در مرورگر برای پیکربندی v2rayA و نمایش وضعیت آن است؛ سرویس با کاربر root اجرا می‌شود (در Windows با دسترسی مدیر؛ با `--lite` بدون دسترسی ویژه) و خودش `v2raya_core` را راه‌اندازی می‌کند.

## ورود

اولین حسابی که ثبت می‌شود، حساب مدیر است؛ حساب دیگری وجود ندارد. ثبت‌نام به نام کاربری و گذرواژه‌ای با 6 تا 32 نویسه نیاز دارد و به اطلاعات ورود قبلی نیازی ندارد؛ بنابراین در شبکهٔ مشترک، پیش از ثبت‌نام دسترسی به پورت 2017 را محدود کنید، یا سرویس را با `--address 127.0.0.1:2017` اجرا کنید تا فقط از همین دستگاه در دسترس باشد.

برای بازنشانی گذرواژهٔ فراموش‌شده، سرویس را متوقف کنید و فرمان زیر را اجرا کنید:

```sh
v2raya --reset-password
```

فرمان را با همان حساب کاربری و پوشهٔ `--config` سرویس اجرا کنید (در حالت lite، با `--lite` نیز)؛ در Windows سرویس با حساب SYSTEM اجرا می‌شود و پوشهٔ آن با پوشهٔ کاربری که دسترسی مدیر گرفته متفاوت است. سرویس را دوباره راه‌اندازی کنید و حساب جدیدی بسازید.

## وارد کردن گره‌ها

در صفحهٔ **پروکسی‌ها**، گزینهٔ **وارد کردن** پیوندهای اشتراک‌گذاری یا اشتراک را می‌پذیرد: برای پیوندهای `vmess://`، `vless://`، `ss://`، `trojan://`، `hysteria2://`، `tuic://`، `juicity://`، `anytls://`، `wireguard://`، `socks5://`، `http://` و `https://`، هرکدام در یک خط، یا تصویر کد QR، گزینهٔ **پیوند سرور** را انتخاب کنید؛ برای اشتراک، **نشانی اشتراک** را انتخاب کنید. پیوندهای ShadowsocksR پذیرفته نمی‌شوند؛ پیوندهای Shadowsocks با رمزنگاری جریانی (`rc4-md5`، `aes-*-cfb`، `chacha20-ietf`) یا `none` نیز پذیرفته نمی‌شوند، زیرا هسته فقط رمزنگاری‌های AEAD و روش‌های 2022-blake3 را می‌پذیرد. `allow_insecure` در پیوند نادیده گرفته می‌شود: v2rayA هرگز از بررسی گواهی صرف‌نظر نمی‌کند؛ برای سرور خودامضا، SHA-256 گواهی را در فرم گره وارد کنید.

اشتراک‌ها صفحهٔ خود را دارند: **اشتراک‌ها** (در تلفن، مستندات برای جا دادن آن به منوی نوار بالا منتقل می‌شود). گره‌های هر اشتراک کنار هم نگه داشته می‌شوند و می‌توان آن‌ها را دستی یا طبق زمان‌بندی به‌روز کرد (**تنظیمات → به‌روزرسانی خودکار اشتراک‌ها**). تنظیم حالت در کنار این گزینه مشخص می‌کند که به‌روزرسانی از پروکسی بگذرد یا نه.

## گروه‌ها

استفاده از گره‌ها از طریق گروه‌های پروکسی انجام می‌شود. گروه `proxy` همیشه وجود دارد؛ گروه‌های دیگر از دکمهٔ گروه در نوار بالا ساخته، تنظیم و حذف می‌شوند و این تنها جای مدیریت گروه‌هاست. گره از منوی خود به گروه می‌پیوندد، یا چند گره را در صفحهٔ پروکسی‌ها انتخاب کنید و گروه را زیر **افزودن به گروه پروکسی** برگزینید. گروهی با چند عضو متصل از گره‌ای استفاده می‌کند که کمترین تأخیر اندازه‌گیری‌شده را دارد (**خودکار (کمترین تأخیر)**)؛ منوی گره در داشبورد یک عضو را ثابت می‌کند تا گروه فقط از همان استفاده کند، **افزودن یا حذف گره‌ها** اعضا را تغییر می‌دهد و تنظیمات گروه، نشانی و فاصلهٔ زمانی آزمون را تعیین می‌کند.

قوانین مسیریابی از نام گروه‌ها به‌عنوان خروجی استفاده می‌کنند: به‌طور پیش‌فرض `proxy`، و برای هر گروه دیگر، نام آن پس از اتصال دست‌کم یک عضو.

## راه‌اندازی هسته

گزینهٔ **شروع** در داشبورد، یا دکمهٔ وضعیت در بالای صفحه، `v2raya_core` را با پیکربندی فعلی راه‌اندازی می‌کند. تغییرات صفحهٔ تنظیمات با **ذخیره و اعمال** اعمال می‌شوند: هستهٔ در حال اجرا با این تغییرات دوباره راه‌اندازی می‌شود؛ هستهٔ متوقف‌شده آن‌ها را در راه‌اندازی بعدی اعمال می‌کند.

پس از راه‌اندازی هسته، برنامه‌ها از طریق ورودی‌های محلی به آن وصل می‌شوند:

| ورودی          | نشانی             | مسیریابی                                     |
| -------------- | ----------------- | -------------------------------------------- |
| SOCKS5         | `127.0.0.1:20170` | همهٔ ترافیک از `proxy`                       |
| HTTP           | `127.0.0.1:20171` | همهٔ ترافیک از `proxy`                       |
| HTTP با قوانین | `127.0.0.1:20172` | تنظیم **حالت تفکیک ترافیک برای پورت قوانین** |

پورت‌ها در **تنظیمات → نشانی و پورت‌ها** تغییر می‌کنند؛ 0 ورودی را می‌بندد. برای پوشش برنامه‌ها بدون پیکربندی تک‌تک آن‌ها، پروکسی شفاف را روشن کنید.

## مسیریابی

در **تنظیمات → تفکیک ترافیک** مشخص می‌کنید که پورت قوانین و پروکسی شفاف چگونه ترافیک را تفکیک کنند: عبور همه‌چیز به‌جز سایت‌های چینی از پروکسی، عبور فقط GFWList از پروکسی، یا قوانین RoutingA خودتان. RoutingA در بخش جداگانه‌ای توضیح داده شده است.

## صفحه‌کلید

- صفحهٔ پروکسی‌ها در نمای فهرست: `Ctrl`+`A` همهٔ گره‌های فهرست‌شده را انتخاب می‌کند و `Esc` انتخاب را برمی‌دارد؛ در صفحه‌های پروکسی‌ها و گزارش‌ها، `/` به کادر جستجو می‌رود.
- تنظیمات: `Ctrl`+`S` ذخیره می‌کند. ویرایشگر RoutingA: `Ctrl`+`S` ذخیره و `Tab` تورفتگی.
- گفتگوهای دارای کادر متن چندخطی (وارد کردن، اشتراک، اسکریپت‌های مسیر، فهرست‌ها): `Ctrl`+`Enter` ارسال می‌کند؛ `Esc` هر گفتگویی را می‌بندد.
- گزارش‌ها: پس از تمرکز، `Home`، `End`، `PgUp` و `PgDn` در گزارش حرکت می‌کنند.

در macOS، `Cmd` جای `Ctrl` را می‌گیرد.

## Automatic groups and subscription updates

Enable **Automatically add available servers** in a proxy group's settings to
check every server in Proxies, including standalone nodes and all subscriptions.
The group keeps only servers that complete an HTTP(S) request through their VPN
connection. A listening TCP port or an ICMP response is not enough.

The first scan runs when enabled, then after catalog changes and at the group's
probe interval. The default for new groups is **300s**; saved intervals are
preserved. Turning the switch off keeps the current members and returns control
to manual editing. The group's existing Auto/least-ping mode chooses the server.
Subscription updates do not select a server or populate manual groups.

Enable **Automatically update subscription** for each subscription that should
refresh itself. Both minute fields are required:

- **Regular update interval:** 0 disables regular downloads; a positive whole
  number schedules unconditional updates.
- **Retry interval when all servers are unavailable:** at least 1 minute. If no
  saved candidate is reachable, refresh at this interval until one works.
  Failed or empty downloads retain the saved list. A failed subscription retries
  independently of healthy subscriptions in the same group.

Candidate health is checked after subscription/catalog changes and every 300
seconds for enabled subscription updates. Each automatic group uses its own
probe URL and interval. Timers start when enabled or the service starts.
Retries do not postpone unconditional updates. No background action starts a
manually stopped main core; temporary candidate probes can still run.

An empty automatic group has a blocking outbound. Traffic assigned to that group
is rejected instead of falling back to another group or direct access. Explicit
direct routes and other groups retain their own policy. On OpenWrt TPROXY and
Redirect, membership reloads retain existing interception when network settings
are unchanged and no custom hooks are installed. This is not a system-wide kill
switch for power loss, service termination, custom hooks, or TUN teardown.

Upgrading from the earlier Resilient recovery policy preserves accounts,
subscriptions, nodes and saved group members. Old first-server, auto-connect,
recovery and global subscription schedule flags no longer run. The new switches
default off: enable the desired group and subscription policies explicitly.
