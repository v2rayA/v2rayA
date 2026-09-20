# Routing rules

RoutingA is the rule language of v2rayA. **Settings → RoutingA** opens the editor; the rules apply to the rule port (`20172`) when its mode is RoutingA, to the transparent proxy when its policy is *Same as the rule port*, and to custom inbounds bound to RoutingA.

## Syntax

One rule per line: conditions, joined with `&&`, then `->` and an outbound. Rules match from top to bottom and the first match wins; `default:` names the outbound for everything unmatched. `#` starts a comment.

```
default: proxy
domain(geosite:category-ads-all) -> block
domain(geosite:cn) -> direct
ip(geoip:private, geoip:cn) -> direct
```

| Condition | Matches |
| --- | --- |
| `domain(...)` | `full:` exact, `domain:` the domain and its subdomains (the default), `contains:`, `regexp:`, `geosite:<category>`, `ext:"file.dat:tag"` |
| `ip(...)` | an address, a CIDR, `geoip:<code>`, `ext:"file.dat:tag"` |
| `port(...)` | destination ports and ranges, `port(80, 443, 1000-2000)` |
| `sourcePort(...)` | source ports |
| `network(...)` | `tcp`, `udp` |
| `protocol(...)` | sniffed protocol: `http`, `tls`, `bittorrent` (needs **Sniffing** on) |
| `source(...)` | source addresses and CIDRs |
| `inboundTag(...)` | the inbound the connection arrived on |

## Outbounds

`proxy`, `direct` and `block` are built in. Every other proxy group is an outbound under its own name once it has a connected member; the rules fail to apply otherwise. A SOCKS or HTTP server outside v2rayA is declared once and used like a group:

```
outbound: office = socks(address: 192.168.1.10, port: 1080)
outbound: gateway = http(address: 10.0.0.1, port: 8080, user: 'name', pass: 'secret')
domain(domain: corp.example) -> office
```

## Categories

`geosite:` categories come from the v2fly [domain-list-community](https://github.com/v2fly/domain-list-community) and `geoip:` codes from [v2fly/geoip](https://github.com/v2fly/geoip); v2rayA downloads both on first start. Frequently used: `geosite:cn`, `geosite:geolocation-!cn`, `geosite:private`, `geosite:category-ads-all`, `geosite:netflix`, `geosite:youtube`, `geosite:telegram`, `geosite:openai`; `geoip:cn`, `geoip:private`, and any country code. An attribute narrows a category: `geosite:apple@cn`, `geosite:category-games@cn`.

Names that exist only in Loyalsoldier's builds (`gfw`, `greatfire`, `apple-cn`, `geoip:telegram`) are not in these files; the core refuses them with `illegal domain rule`. The GFWList mode downloads Loyalsoldier's file as `LoyalsoldierSite.dat`, and rules can then use it as `domain(ext:"LoyalsoldierSite.dat:gfw")`.

## The editor

The list mode edits each rule with a form; the text mode shows line numbers, colouring and a check per line. The reference beside the editor inserts examples at the cursor, and its **Templates** section holds complete rule sets for the common policies: insert one above your own rules, or replace the whole rule set with it. Rules are imported and exported as a text file.

Saving restarts the core with the new rules; when the core rejects them, the previous rules stay in place and the error names the line.
