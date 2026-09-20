# Routing rules

RoutingA is the rule language of v2rayA. **Settings → RoutingA** opens the editor; the rules apply to the rule port (`20172`) when its mode is RoutingA and to the transparent proxy when its policy is _Same as the rule port_. A custom inbound bound to RoutingA has its own rule text, see below.

## Syntax

One rule per line: conditions, joined with `&&`, then `->` and an outbound. Rules match from top to bottom and the first match wins; `default:` names the outbound for everything unmatched. A line starting with `#` is a comment. v2rayA puts a few rules of its own in front: the proxy servers' addresses and Apple's push service go direct, and an `inboundTag(...)` condition replaces the rule's usual inbound scope.

```
default: proxy
domain(geosite:category-ads-all) -> block
domain(geosite:cn) -> direct
ip(geoip:private, geoip:cn) -> direct
```

| Condition         | Matches                                                                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `domain(...)`     | `domain:` the domain and its subdomains, `full:` exact, `regexp:`, `geosite:<category>`, `ext:"file.dat:tag"`; a bare value such as `example.com` matches as a substring |
| `ip(...)`         | an address, a CIDR (quote IPv6: `ip("2001:db8::/32")`), `geoip:<code>`, `ext:"file.dat:tag"`                                                                             |
| `port(...)`       | destination ports and ranges, `port(80, 443, 1000-2000)`                                                                                                                 |
| `sourcePort(...)` | source ports                                                                                                                                                             |
| `network(...)`    | `tcp`, `udp`                                                                                                                                                             |
| `protocol(...)`   | sniffed protocol: `http`, `tls`, `quic`, `bittorrent` (needs **Sniffing** on)                                                                                            |
| `source(...)`     | source addresses and CIDRs                                                                                                                                               |
| `inboundTag(...)` | the inbound the connection arrived on                                                                                                                                    |

## Outbounds

`proxy`, `direct` and `block` are built in. Every other proxy group is an outbound under its own name once it has a connected member; the rules fail to apply otherwise. A SOCKS or HTTP server outside v2rayA is declared once and used like a group (an `ext:` file goes into the rule-data directory):

```
outbound: office = socks(address: 192.168.1.10, port: 1080)
outbound: gateway = http(address: 10.0.0.1, port: 8080, user: 'name', pass: 'secret')
domain(domain: corp.example) -> office
```

## Categories

`geosite:` categories come from the v2fly [domain-list-community](https://github.com/v2fly/domain-list-community) and `geoip:` codes from [v2fly/geoip](https://github.com/v2fly/geoip); v2rayA downloads whichever file is missing on first start, a package may ship its own build. Frequently used: `geosite:cn`, `geosite:geolocation-!cn`, `geosite:private`, `geosite:category-ads-all`, `geosite:greatfire`, `geosite:netflix`, `geosite:youtube`, `geosite:telegram`, `geosite:openai`; `geoip:cn`, `geoip:private` and the country codes. An attribute narrows a category: `geosite:apple@cn`, `geosite:category-games@cn`.

Names that exist only in Loyalsoldier's builds (`gfw`, `apple-cn`, `google-cn`, `geoip:telegram`) are not in these files; the core refuses them with `illegal domain rule` or `illegal ip rule`. **Update GFWList** downloads that build (from `v2rayA/dist-v2ray-rules-dat`) as `LoyalsoldierSite.dat`, and rules can then use it as `domain(ext:"LoyalsoldierSite.dat:gfw")`.

## The editor

The list mode edits each rule with a form; the text mode shows line numbers, colouring and a check per line. The reference beside the editor inserts examples at the cursor, and its **Templates** section holds complete rule sets for the common policies: insert one above your own rules, or replace the whole rule set with it. Rules are imported and exported as a text file.

Saving reloads a running core with the new rules; when the core rejects them, v2rayA puts the previous rules back and shows the core's error, which names the offending text. A stopped core takes the rules on its next start.

## Custom inbounds

A custom inbound bound to RoutingA has its own rule text that applies to that inbound only. It takes rules, not definitions: `default:` and `outbound:` lines are ignored, and traffic no rule matches goes to the group chosen in the form. End the text with a catch-all such as `network(tcp, udp) -> direct` to route everything else yourself.
