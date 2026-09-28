// Adapted from https://github.com/wywywywycloud/routinga-russia/blob/48a94f5734c525b11a1c6092eae9745b2a13017e/routinga-ru.txt
// Keep proxy exceptions above direct rules: the first match wins.
// Copyright (c) 2026 wywywywycloud. MIT license: see LICENSE-russia.txt.
export const russia = `default: proxy
# Russian domains that still need the proxy.
domain(domain: abook-club.ru, domain: amdm.ru, domain: anime-portal.su, domain: animespirit.ru, domain: anistars.ru) -> proxy
domain(domain: baikal-journal.ru, domain: bestchange.ru, domain: colta.ru, domain: coomer.su, domain: doramalive.ru) -> proxy
domain(domain: ej.ru, domain: fn-volga.ru, domain: grani.ru, domain: itsmycity.ru, domain: jut.su) -> proxy
domain(domain: kara.su, domain: kasparov.ru, domain: kemono.su, domain: mangahub.ru, domain: megapeer.ru) -> proxy
domain(domain: moscowtimes.ru, domain: newtimes.ru, domain: novayagazeta.ru, domain: ohmyswift.ru, domain: paperpaper.ru) -> proxy
domain(domain: pimpletv.ru, domain: polit.ru, domain: republic.ru, domain: scryde.ru, domain: seasonvar.ru) -> proxy
domain(domain: sklatchiki.ru, domain: the-village.ru, domain: theins.ru, domain: tvrain.ru, domain: ua) -> proxy
domain(domain: zahav.ru) -> proxy

# v2fly geosite.dat category-ru covers Russian services and domains.
domain(geosite: category-ru) -> direct

ip(geoip:ru) -> direct

# Russian services outside category-ru.
domain(domain: 1018213540.rsc.cdn77.org, domain: bitrix.info) -> direct

# Local hostnames and IP addresses.
domain(geosite:private) -> direct
ip(geoip:private) -> direct`;
