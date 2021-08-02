# How to Install v2rayA on OpenWRT

## Before start

Some utils is needed:

```bash
opkg update
opkg install ca-certificates wget unzip tar curl
opkg install kmod-ipt-nat6 iptables-mod-tproxy iptables-mod-filter
```

## Install v2ray-core

Download and put v2ray-core files to `/usr/bin`.

The latest link can be found at https://github.com/v2fly/v2ray-core/releases.

```bash
# For example:
wget https://github.com/v2fly/v2ray-core/releases/download/v4.40.1/v2ray-linux-64.zip
unzip -d v2ray-core v2ray-linux-64.zip
cp v2ray-core/v2ray v2ray-core/v2ctl /usr/bin
```

## Install v2rayA

1. Install GUI:

   ```bash
   cd /tmp
   latest_version=$(curl -s https://apt.v2raya.mzz.pub/dists/v2raya/main/binary-amd64/Packages|grep Version|cut -d' ' -f2)
   wget https://apt.v2raya.mzz.pub/pool/main/v/v2raya/web_v${latest_version}.tar.gz
   mkdir /etc/v2raya
   tar xzvf web_v1.4.1.tar.gz --directory /etc/v2raya
   ```

2. Install Binary:

   ```bash
   wget -O /usr/bin/v2raya https://apt.v2raya.mzz.pub/pool/main/v/v2raya/v2raya_linux_x64_v${lasest_veresion}
   chmod +x /usr/bin/v2raya
   ```

3. Install Daemon File:

   ```bash
   cat <<EOF >/etc/init.d/v2raya
   #!/bin/sh /etc/rc.common
   command=/usr/bin/v2raya
   PIDFILE=/var/run/v2raya.pid
   depend() {
   	need net
   	after firewall
   	use dns logger
   }
   start() {
   	start-stop-daemon -b -S -m -p "${PIDFILE}" -x $command
   }
   stop() {
   	start-stop-daemon -K -p "${PIDFILE}"
   }
   EOF
   
   chmod +x /etc/init.d/v2raya
   ```

4. Start and auto-start

   ```bash
   /etc/init.d/v2raya start
   /etc/init.d/v2raya enable
   ```

   v2rayA will listen port 2017 by default.

