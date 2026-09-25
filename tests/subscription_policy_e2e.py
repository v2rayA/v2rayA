#!/usr/bin/env python3
"""Exercise selection policy and recovery edge cases in a disposable OpenWrt VM."""
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
            except (OSError, AssertionError, KeyError, http.client.HTTPException, subprocess.SubprocessError): pass
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
        assert not api('touch')['touch']['subscriptions'][0]['preferFirst']
        settings = api('setting')['setting']
        settings.update(portSharing=True, transparent='proxy', transparentType='tproxy', subscriptionAutoUpdateMode='none')
        api('setting', settings, 'PUT')
        api('connection', {'_type':'subscriptionServer','sub':0,'id':2,'outbound':'proxy'})
        api('v2ray', {}, 'POST')
        toggle(True)
        def policy(first):
            raw = api('touch')['touch']['subscriptions'][0]
            raw['preferFirst'] = first
            api('subscription', {'subscription':raw}, 'PATCH')
            assert api('touch')['touch']['subscriptions'][0]['preferFirst']==first
        policy(True)
        assert selected()=='dead-first'
        assert api('touch')['touch']['subscriptions'][0]['servers'][1].get('latency','')==''
        assert api('connection', {'_type':'subscriptionServer','sub':0,'id':2,'outbound':'proxy'}, fail=True)['code']=='FAIL'
        record('first-entry policy applies immediately and blocks manual fallback')
        fetches = sub.fetches
        await_condition(lambda:sub.fetches>fetches, 95, 'first-entry monitoring did not refresh')
        assert selected()=='dead-first'
        record('monitoring never falls back to available later entries')
        # Change to another dead first entry while recovery is active.
        other_dead = f'vless://{uuid.uuid4()}@{args.host_address}:{free_port()}?encryption=none&security=none&type=tcp#replacement-dead'
        sub.links = [other_dead,fast.link,backup.link]
        await_condition(lambda:selected()=='replacement-dead',35,'new first entry not applied')
        fetches = sub.fetches
        await_condition(lambda:sub.fetches>fetches,40,'dead replacement reset recovery to the minute window')
        record('dead replacement first entry continues recovery without another minute')
        sub.links = [backup.link,fast.link,dead]
        await_condition(lambda:selected()=='backup' and traffic()=='backup',45,'recovery did not adopt working first entry')
        assert ssh("curl -fsS --noproxy '*' --max-time 5 http://198.18.0.1/transparent")=='backup'
        record('first entry wins over faster later node, with real TPROXY traffic')
        sub.links = []
        assert api('subscription', {'_type':'subscription','id':1}, 'PUT', True)['code']=='FAIL'
        assert selected()=='backup' and traffic()=='backup'
        record('empty subscription preserves the first-entry connection')
        sub.links = [fast.link,backup.link,dead]
        api('subscription', {'_type':'subscription','id':1}, 'PUT')
        assert selected()=='fast' and traffic()=='fast'
        record('subscription reorder follows new first position')
        sub.links = [dead,backup.link,fast.link]
        api('subscription', {'_type':'subscription','id':1}, 'PUT')
        assert selected()=='dead-first'
        fast.stop()
        backup.stop()
        policy(False)
        assert not api('touch')['touch']['subscriptions'][0]['preferFirst']
        record('working-server policy can be saved while every node is dead')
        fast.start()
        backup.start()
        await_condition(lambda:selected()=='fast' and traffic()=='fast',40,'saved policy did not recover after nodes returned')
        assert selected()=='fast'  and traffic()=='fast'
        record('working-server switch immediately restores a reachable later node')
        toggle(False)
        fast.stop()
        api('subscription', {'_type':'subscription','id':1}, 'PUT')
        assert selected()=='backup' and traffic()=='backup'
        record('active policy applies with monitoring and automatic selection disabled')
        fast.start()
        api('subscription', {'_type':'subscription','id':1}, 'PUT')
        toggle(True)
        sub.status=503
        fast.stop()
        await_condition(lambda:selected()=='backup' and traffic()=='backup',100,'cached candidates did not recover unavailable subscription host')
        assert 'refresh unavailable; checking saved candidates' in ssh('cat /var/log/v2raya/v2raya.log')
        record('HTTP 503 subscription outage recovers through a saved candidate')
        sub.status=200
        # The private monitor must work with every public proxy port disabled.
        api('ports', {'socks5':0,'http':0,'socks5WithPac':0,'httpWithPac':0,'vmess':0}, 'PUT')
        checks,fetches=backup.fixture.requests,sub.fetches
        time.sleep(22)
        assert backup.fixture.requests>checks and sub.fetches==fetches
        record('monitoring works with public proxy listeners disabled')
        ssh("kill -9 $(pgrep -f '^/usr/bin/xray run --config=/etc/v2raya/config.json$')")
        def transparent_restored():
            return ssh("curl -fsS --noproxy '*' --max-time 3 http://198.18.0.1/transparent")=='backup'
        # Let the production failure window elapse before checking traffic;
        # repeated SSH process launches are unnecessary during that minute.
        time.sleep(65)
        await_condition(transparent_restored,40,'dead local Xray core was not restarted')
        record('unexpected local core death recovers without public proxy ports')
        api('v2ray', {}, 'DELETE')
        policy(True)
        assert not api('touch')['running'] and selected()=='dead-first'
        policy(False)
        assert not api('touch')['running'] and selected()=='backup'
        record('policy changes while manually stopped never start the core')
        api('v2ray', {}, 'POST')
        ssh('/etc/init.d/v2raya restart')
        def logged_in():
            nonlocal token
            token = api('login', {'username':'monitortest','password':'disposable-monitor-only'})['token']
            raw=api('touch')['touch']['subscriptions'][0]
            return not raw['preferFirst'] and raw['monitor']
        await_condition(logged_in,20,'policy persistence failed')
        record('selection policy and monitor settings persist across service restart')
        record('final memory and cleanup',details=ssh("free; pgrep -f '^/usr/bin/xray run '; find /tmp -maxdepth 1 -name 'v2raya-subscription-*'"))
    finally:
        sub.release.set()
        try: (args.output/'guest.log').write_text(ssh('cat /var/log/v2raya/v2raya.log'))
        except (OSError, subprocess.SubprocessError): pass
        sub.shutdown()
        fast.close()
        backup.close()


if __name__=='__main__': main()
