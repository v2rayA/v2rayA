# moona

A Fast and Convenient Cross-Protocols latency tester built with love in Go.

## Usage

```bash
$ docker run --rm mzz2017/moona --help

Usage of moona:
  -f, --file string    input file where share-links are split by lines
  -l, --link string    subscription link or share-link
  -t, --timeout int    test timeout(ms) (default 10000)
  -p, --parallel int   the max number of parallel tests (default 5)
  -h, --help           print this help menu
```

### single test
```bash
$ docker run --rm mzz2017/moona --link ss://***@***:***
Importing ss://***@***:***
Test done[1]248ms: ss://***@***:***
```

### batch test
**from subscription link**
```bash
$ docker run --rm mzz2017/moona --link https://**********
Importing https://**********
Test done[1]308: vmess://**********
Test done[3]896ms: vmess://**********
Test done[2]1115ms: vmess://**********
```

**from file**
```bash
$ cat f.txt
vmess://**********
ssr://**********
ss://***@***:***
trojan://***@***:***?allowInsecure=false

$ docker run --rm -v $(pwd)/f.txt:/f.txt mzz2017/moona --file /f.txt
Importing https://**********
Test done[1]338ms: vmess://**********
Test done[4]288ms: trojan://***@***:***?allowInsecure=false
Test done[3]266ms: ss://***@***:***
Test done[2]188ms: ssr://**********
```