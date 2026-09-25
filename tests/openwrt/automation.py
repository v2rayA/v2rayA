#!/usr/bin/env python3
"""Exercise catalog/group automation on a disposable OpenWrt 24.10 VM.

Requires a fresh application database. Never use against a physical router.
The independent DNAT trap makes a direct escape observable, including during
membership reloads: DIRECT can only come from bypassing the VLESS fixtures.
"""
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

from fixtures import Node, Proxy, Subscription, free_port


class CountedSubscription(Subscription):
    def do_GET(self):
        self.server.fetches += 1
        super().do_GET()


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('--disposable-vm', action='store_true', required=True)
    p.add_argument('--xray', type=Path, required=True)
    p.add_argument('--ssh-key', type=Path, required=True)
    p.add_argument('--quick', action='store_true')
    p.add_argument('--output', type=Path, required=True)
    p.add_argument('--host-address', default='10.0.2.2')
    args = p.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    token, results = None, []
    fast, backup, standalone, blocked = [Node(n, d, args, b) for n, d, b in
        [('fast', .01, False), ('backup', .2, False), ('standalone', .4, False), ('blackhole', 0, True)]]
    nodes = [fast, backup, standalone, blocked]
    direct = Proxy('DIRECT')
    dead = f'vless://{uuid.uuid4()}@{args.host_address}:{free_port()}?encryption=none&security=none&type=tcp#dead-first'
    fast_link = fast.link.replace(args.host_address, '198.18.0.3')
    wrong = fast_link.replace(fast.identity, str(uuid.uuid4())).replace('#fast', '#wrong-key')
    subs = []
    for links in [[dead, wrong, blocked.link, fast_link], [backup.link]]:
        sub = http.server.ThreadingHTTPServer(('127.0.0.1', 0), CountedSubscription)
        sub.links, sub.status, sub.fetches = links, 200, 0
        threading.Thread(target=sub.serve_forever, daemon=True).start()
        subs.append(sub)

    def ssh(command, timeout=45):
        return subprocess.check_output(['ssh', '-F', '/dev/null', '-i', str(args.ssh_key.resolve()),
            '-o', 'IdentitiesOnly=yes', '-o', 'StrictHostKeyChecking=accept-new', '-o', 'LogLevel=ERROR',
            '-o', 'UserKnownHostsFile='+str((args.output/'known_hosts').resolve()),
            '-p', '22022', 'root@127.0.0.1', command], timeout=timeout, stderr=subprocess.STDOUT).decode()

    def api(path, data=None, method=None, fail=False):
        headers = {'Content-Type':'application/json'}
        if token: headers['Authorization'] = 'Bearer '+token
        req = urllib.request.Request('http://127.0.0.1:22017/api/'+path, headers=headers,
            data=None if data is None else json.dumps(data).encode(), method=method)
        with urllib.request.urlopen(req, timeout=90) as response: reply=json.load(response)
        if not fail: assert reply['code']=='SUCCESS', reply
        return reply if fail else reply['data']

    def wait(check, seconds=60, message='condition not reached'):
        deadline=time.monotonic()+seconds
        while time.monotonic()<deadline:
            try:
                if check(): return
            except (OSError, AssertionError, KeyError, http.client.HTTPException, subprocess.SubprocessError): pass
            time.sleep(.5)
        raise AssertionError(message)

    def record(name, **evidence):
        results.append({'test':name, 'result':'PASS', **evidence})
        (args.output/'results.json').write_text(json.dumps(results,indent=2)+'\n')
        print('PASS:',name,evidence or '',flush=True)

    def touch(): return api('touch')['touch']
    def members(): return [w for w in (touch()['connectedServer'] or []) if w['outbound']=='proxy']
    def traffic():
        c=http.client.HTTPConnection('127.0.0.1',22171,timeout=3)
        try:
            c.request('GET','http://198.18.0.1/traffic');r=c.getresponse()
            return r.read().decode() if r.status==200 else ''
        finally:c.close()
    def group(enabled):
        api('outbound',{'outbound':'proxy','setting':{'autoAdd':enabled,'probeURL':'http://198.18.0.1/check','probeInterval':'10s','type':'leastping'}},'PUT')
    def refresh(index=0):api('subscription',{'_type':'subscription','id':index+1},'PUT')
    stop_traffic=threading.Event()
    traffic_errors=[]
    def router_traffic():
        while not stop_traffic.is_set():
            try:
                output=ssh('curl -s --max-time 1 http://198.18.0.1/traffic || true',5)
                if 'DIRECT' in output:traffic_errors.append(output)
            except subprocess.SubprocessError: pass
            stop_traffic.wait(.05)
    thread=None
    try:
        wait(lambda: api('version'), 15, 'service did not start')
        token=api('account',{'username':'vmtest','password':'disposable-vm-only'})['token']
        version=api('version');assert version['coreVersionValid'],version
        api('ports',{'socks5':20170,'http':20171,'socks5WithPac':0,'httpWithPac':0,'vmess':0},'PUT')
        for sub in subs:api('import',{'url':f'http://{args.host_address}:{sub.server_port}/subscription'})
        api('import',{'url':standalone.link})
        settings=api('setting')['setting'];settings.update(portSharing=True,transparent='proxy',transparentType='tproxy',subscriptionAutoUpdateMode='none')
        api('setting',settings,'PUT')
        trap=f'table ip resilient_test {{ chain output {{ type nat hook output priority dstnat; policy accept; meta mark & 0xc0 != 0x40 ip daddr 198.18.0.1 tcp dport 80 dnat to 10.0.2.2:{direct.server_address[1]}; meta mark & 0xc0 != 0x40 ip daddr 198.18.0.2 tcp dport 80 dnat to 10.0.2.2:{subs[0].server_port}; meta mark & 0xc0 != 0x40 ip daddr 198.18.0.3 tcp dport {fast.port} dnat to 10.0.2.2:{fast.port}; }}; }}'
        ssh("printf '%s' '"+trap+"' | nft -f -")
        assert ssh('curl -fsS --max-time 5 http://198.18.0.1/traffic').strip()=='DIRECT'
        record('independent direct-escape trap is reachable before interception')
        raw=touch()['subscriptions'][0];raw['address']='http://198.18.0.2/subscription'
        api('subscription',{'subscription':raw},'PATCH')
        group(True)
        wait(lambda:len(members())==3, message='did not collect all healthy nodes across two subscriptions and standalone')
        assert not api('touch').get('running', False)
        record('automatic group adds only three healthy VPN nodes from the entire catalog while stopped')
        api('v2ray',{},'POST')
        wait(lambda:traffic()=='fast')
        assert ssh('curl -fsS --max-time 10 http://198.18.0.1/traffic').strip()=='fast'
        base_direct=direct.requests
        thread=threading.Thread(target=router_traffic,daemon=True);thread.start()
        fast.stop();wait(lambda:len(members())==2);wait(lambda:traffic()=='backup')
        record('later healthy subscription replaces fastest failed member')
        backup.stop();standalone.stop();wait(lambda:len(members())==0)
        time.sleep(4)
        config=json.loads(ssh('cat /etc/v2raya/config.json'))
        assert any(o['tag']=='proxy' and o['protocol']=='blackhole' for o in config['outbounds'])
        assert 'v2raya' in ssh('nft list table inet v2raya')
        assert not traffic_errors and direct.requests==base_direct,(traffic_errors,direct.requests,base_direct)
        record('empty group uses blackhole, retains TPROXY and never reaches direct trap through reloads')
        fast.start();wait(lambda:len(members())==1);wait(lambda:traffic()=='fast')
        assert not traffic_errors and direct.requests==base_direct
        record('empty group recovers automatically without starting or changing a manual group')
        group(False);before=members();backup.start();time.sleep(13)
        assert members()==before
        record('disabling auto membership preserves members and suppresses scans')
        group(True);wait(lambda:len(members())==2)
        assert not traffic_errors and direct.requests==base_direct,(traffic_errors,direct.requests,base_direct)
        if args.quick: return
        raw=touch()['subscriptions'][0];raw.update(autoUpdate=True,updateIntervalMinutes=0,failureIntervalMinutes=1)
        api('subscription',{'subscription':raw},'PATCH')
        assert all(k not in touch()['subscriptions'][0] for k in ('autoSelect','monitor','preferFirst'))
        invalid=dict(raw);invalid['failureIntervalMinutes']=0
        assert api('subscription',{'subscription':invalid},'PATCH',True)['code']!='SUCCESS'
        record('new timers round-trip; old switches absent; zero failure interval rejected')
        fast.stop();subs[0].links=[dead];refresh();wait(lambda:len(members())==1)
        count=subs[0].fetches
        started=time.monotonic()
        wait(lambda:subs[0].fetches>count,85,'one-minute retry did not refresh failed subscription')
        assert time.monotonic()-started>=45
        record('regular interval zero still permits bounded one-minute failure retry')
        fast.start();subs[0].links=[dead,fast_link]
        wait(lambda:any(s['name']=='fast' for s in touch()['subscriptions'][0]['servers']),85,'retry did not discover new working node')
        wait(lambda:len(members())==2);wait(lambda:traffic()=='fast')
        count=subs[0].fetches;time.sleep(65)
        assert subs[0].fetches==count,(subs[0].fetches,count)
        assert not traffic_errors and direct.requests==base_direct,(traffic_errors,direct.requests,base_direct)
        record('retry discovers new server; stops after recovery; regular zero does not poll provider')
        stop_traffic.set();thread.join(7)
        api('v2ray',method='DELETE')
        before=ssh('pgrep -x v2raya_core || true')
        time.sleep(12)
        assert not api('touch')['running'], before
        record('manual stop remains stopped while automation is enabled')
        for node in nodes:node.stop()
        refresh();wait(lambda:len(members())==0)
        api('v2ray',method='POST')
        assert api('touch')['running']
        count=direct.requests
        try:response=traffic()
        except (OSError,http.client.HTTPException):response=''
        assert response!='DIRECT' and direct.requests==count
        api('v2ray',method='DELETE')
        record('explicit start with an empty automatic group succeeds and blocks traffic')
    finally:
        stop_traffic.set()
        if thread:thread.join(7)
        try:
            (args.output/'guest.log').write_text(ssh('cat /var/log/v2raya/v2raya.log'))
            ssh('nft delete table ip resilient_test 2>/dev/null || true')
        except (OSError,subprocess.SubprocessError):pass
        for sub in subs:sub.shutdown();sub.server_close()
        for node in nodes:node.close()
        direct.shutdown();direct.server_close()


if __name__=='__main__':main()
