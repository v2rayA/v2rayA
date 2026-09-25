#!/usr/bin/env python3
"""Exercise production monitoring timers in a disposable OpenWrt VM."""
import argparse
import http.client
import http.server
import json
from pathlib import Path
import subprocess
import threading
import time
import urllib.request
import uuid

from openwrt_vm_e2e import Node
from subscription_e2e import Subscription, free_port


class CountedSubscription(Subscription):
    def do_GET(self):
        self.server.fetches += 1
        if self.server.block:
            self.server.entered.set()
            self.server.release.wait(30)
        super().do_GET()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--disposable-vm', action='store_true', required=True)
    p.add_argument('--xray', type=Path, required=True)
    p.add_argument('--ssh-key', type=Path, required=True)
    p.add_argument('--output', type=Path, required=True)
    p.add_argument('--host-address', default='10.0.2.2')
    args = p.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    fast, backup = Node('fast', .01, args), Node('backup', .2, args)
    dead = f'vless://{uuid.uuid4()}@{args.host_address}:{free_port()}?encryption=none&security=none&type=tcp#dead-first'
    sub = http.server.ThreadingHTTPServer(('127.0.0.1', 0), CountedSubscription)
    sub.links, sub.status, sub.fetches = [dead, fast.link, backup.link], 200, 0
    sub.block, sub.entered, sub.release = False, threading.Event(), threading.Event()
    threading.Thread(target=sub.serve_forever, daemon=True).start()
    token, results = None, []

    def ssh(command):
        return subprocess.check_output(['ssh', '-F', '/dev/null', '-i', str(args.ssh_key.resolve()),
            '-o', 'IdentitiesOnly=yes', '-o', 'StrictHostKeyChecking=accept-new', '-o', 'LogLevel=ERROR',
            '-o', 'UserKnownHostsFile='+str((args.output/'known_hosts').resolve()),
            '-p', '22022', 'root@127.0.0.1', command], timeout=45, stderr=subprocess.STDOUT).decode()

    def api(path, data=None, method=None, fail=False):
        headers = {'Content-Type':'application/json'}
        if token: headers['Authorization'] = 'Bearer '+token
        request = urllib.request.Request('http://127.0.0.1:22017/api/'+path, headers=headers,
            data=None if data is None else json.dumps(data).encode(), method=method)
        with urllib.request.urlopen(request, timeout=90) as r: response = json.load(r)
        if not fail: assert response['code']=='SUCCESS', response
        return response if fail else response['data']

    def selected():
        touch = api('touch')['touch']
        which = next(w for w in touch['connectedServer'] if w['outbound']=='proxy')
        return touch['subscriptions'][which['sub']]['servers'][which['id']-1]['name']

    def traffic():
        c = http.client.HTTPConnection('127.0.0.1', 22171, timeout=3)
        try:
            c.request('GET', 'http://198.18.0.1/traffic')
            r = c.getresponse()
            assert r.status==200, r.status
            return r.read().decode()
        finally: c.close()

    def toggle(enabled):
        raw = api('touch')['touch']['subscriptions'][0]
        raw['monitor'] = enabled
        api('subscription', {'subscription':raw}, 'PATCH')
        assert api('touch')['touch']['subscriptions'][0]['monitor']==enabled

    def await_condition(check, seconds, message):
        deadline = time.monotonic()+seconds
        while time.monotonic()<deadline:
            try:
                if check(): return
            except (OSError, AssertionError, KeyError, http.client.HTTPException): pass
            time.sleep(1)
        raise AssertionError(message)

    def record(name, **evidence):
        results.append({'test':name, 'result':'PASS', **evidence})
        (args.output/'results.json').write_text(json.dumps(results,indent=2)+'\n')
        print('PASS:', name, evidence or '', flush=True)

    try:
        assert api('version')['version']=='2.2.7.3-failover.3'
        token = api('account', {'username':'monitortest','password':'disposable-monitor-only'})['token']
        api('ports', {'socks5':20170,'http':20171,'socks5WithPac':0,'httpWithPac':0,'vmess':0}, 'PUT')
        api('outbound', {'outbound':'proxy','setting':{'probeURL':'http://198.18.0.1/check','probeInterval':'10s','type':'leastping'}}, 'PUT')
        api('import', {'url':f'http://{args.host_address}:{sub.server_port}/subscription'})
        raw = api('touch')['touch']['subscriptions'][0]
        assert not raw['monitor'], raw
        settings = api('setting')['setting']
        settings.update(portSharing=True, transparent='proxy', transparentType='tproxy', subscriptionAutoUpdateMode='none')
        api('setting', settings, 'PUT')
        api('connection', {'_type':'subscriptionServer','sub':0,'id':2,'outbound':'proxy'})
        api('v2ray', {}, 'POST')
        assert traffic()=='fast'
        toggle(True)
        original_fetches, checks = sub.fetches, fast.fixture.requests
        time.sleep(22)
        assert sub.fetches==original_fetches
        assert fast.fixture.requests>checks, 'active tunnel not checked'
        assert int(ssh("pgrep -f '^/usr/bin/xray run ' | wc -l").strip())==1
        config = json.loads(ssh('cat /etc/v2raya/config.json'))
        inbound = next(i for i in config['inbounds'] if i['tag']=='subscription-monitor')
        assert inbound['listen']=='127.0.0.1'
        record('healthy monitoring reuses one core, sends real traffic, does not refresh subscription')

        # Short failures must not accumulate into the one-minute window.
        fast.stop()
        time.sleep(20)
        fast.start()
        time.sleep(20)
        assert sub.fetches==original_fetches
        assert selected()=='fast' and traffic()=='fast'
        record('short outage recovers without subscription refresh or switch')

        fast.stop()
        failed_at = time.monotonic()
        time.sleep(50)
        assert sub.fetches==original_fetches, 'recovery started before a continuous minute'
        await_condition(lambda: selected()=='backup' and traffic()=='backup', 55,
                        'monitor did not recover a sustained failure')
        elapsed = time.monotonic()-failed_at
        assert elapsed>=60, elapsed
        assert len(api('touch')['touch']['subscriptions'][0]['servers'])==3
        assert ssh("curl -fsS --noproxy '*' --max-time 5 http://198.18.0.1/transparent")=='backup'
        record('real one-minute outage triggers automatic refresh and VLESS/TPROXY failover', elapsed_seconds=round(elapsed,1))

        # A failed manual refresh must wake recovery without another minute/hour.
        backup.stop()
        fetches = sub.fetches
        result = api('subscription', {'_type':'subscription','id':1}, 'PUT', True)
        assert result['code']=='FAIL', result
        await_condition(lambda: sub.fetches>=fetches+3, 18, 'no immediate recovery retry after all-dead update')
        record('all-dead update wakes recovery and immediately retries the subscription', requests=sub.fetches-fetches)
        fast.start()
        await_condition(lambda: selected()=='fast' and traffic()=='fast', 50, 'recovery loop did not find restored node')
        record('recovery continues independently of the normal update schedule')

        # Disabling the switch must cancel future autonomous attempts.
        toggle(False)
        fetches, checks = sub.fetches, fast.fixture.requests
        fast.stop()
        time.sleep(22)
        assert sub.fetches==fetches and fast.fixture.requests==checks
        assert selected()=='fast'
        record('monitor switch off stops health checks and recovery')
        fast.start()
        toggle(True)
        fast.stop()
        assert api('subscription', {'_type':'subscription','id':1}, 'PUT', True)['code']=='FAIL'
        sub.block = True
        assert sub.entered.wait(25), 'background recovery did not enter blocked download'
        started = time.monotonic()
        toggle(False)
        assert time.monotonic()-started<4, 'background download blocked settings'
        sub.block = False
        sub.release.set()
        time.sleep(2)
        assert not api('touch')['touch']['subscriptions'][0]['monitor']
        record('switch off cancels blocked recovery promptly without overwriting settings')
        fast.start()
        toggle(True)
        api('v2ray', {}, 'DELETE')
        fetches = sub.fetches
        time.sleep(22)
        assert not api('touch')['running']
        assert sub.fetches==fetches
        assert int(ssh("pgrep -f '^/usr/bin/xray run ' | wc -l").strip())==0
        record('manual service stop is respected even while monitoring remains enabled')
        api('v2ray', {}, 'POST')
        ssh('/etc/init.d/v2raya restart')
        def logged_in():
            nonlocal token
            token = api('login', {'username':'monitortest','password':'disposable-monitor-only'})['token']
            return api('touch')['touch']['subscriptions'][0]['monitor'] and traffic()=='fast'
        await_condition(logged_in, 20, 'monitor setting did not survive restart')
        record('monitor setting and active traffic survive OpenWrt service restart')
        record('final memory and cleanup', details=ssh("free; pgrep -f '^/usr/bin/xray run '; find /tmp -maxdepth 1 -name 'v2raya-subscription-*'"))
    finally:
        sub.release.set()
        try: (args.output/'guest.log').write_text(ssh('cat /var/log/v2raya/v2raya.log'))
        except (OSError, subprocess.SubprocessError): pass
        sub.shutdown()
        fast.close()
        backup.close()


if __name__=='__main__': main()
