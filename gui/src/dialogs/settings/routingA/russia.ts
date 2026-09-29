// Adapted from https://github.com/wywywywycloud/routinga-russia/blob/48a94f5734c525b11a1c6092eae9745b2a13017e/routinga-ru.txt
// Keep proxy exceptions above direct rules: the first match wins.
export const russia = `default: proxy
# Google and YouTube: explicit exceptions for .ru domains
domain(domain: google.ru, domain: youtube.ru) -> proxy

# Service domains across all regions
domain(geosite:google, geosite:youtube) -> proxy
domain(geosite:telegram) -> proxy
domain(geosite:facebook, geosite:instagram) -> proxy
domain(geosite:twitter, geosite:discord) -> proxy
domain(geosite:openai) -> proxy
domain(geosite:spotify, geosite:netflix) -> proxy

# WhatsApp domains and their subdomains
domain(domain: whatsapp.com, domain: whatsapp.net, domain: wa.me) -> proxy

# Telegram IPv4 ranges, regardless of geolocation
ip("91.108.56.0/22", "91.108.4.0/22", "91.108.8.0/22") -> proxy
ip("91.108.16.0/22", "91.108.12.0/22", "149.154.160.0/20") -> proxy
ip("91.105.192.0/23", "91.108.20.0/22", "185.76.151.0/24") -> proxy

# Telegram IPv6 ranges, regardless of geolocation
ip("2001:b28:f23d::/48", "2001:b28:f23f::/48", "2001:67c:4e8::/48") -> proxy
ip("2001:b28:f23c::/48", "2a0a:f280::/32") -> proxy

# Russian domains that still need the proxy
domain(domain: abook-club.ru, domain: amdm.ru, domain: anime-portal.su) -> proxy
domain(domain: animespirit.ru, domain: anistars.ru) -> proxy
domain(domain: baikal-journal.ru, domain: bestchange.ru, domain: colta.ru) -> proxy
domain(domain: coomer.su, domain: doramalive.ru) -> proxy
domain(domain: ej.ru, domain: fn-volga.ru, domain: grani.ru) -> proxy
domain(domain: itsmycity.ru, domain: jut.su) -> proxy
domain(domain: kara.su, domain: kasparov.ru, domain: kemono.su) -> proxy
domain(domain: mangahub.ru, domain: megapeer.ru) -> proxy
domain(domain: moscowtimes.ru, domain: newtimes.ru, domain: novayagazeta.ru) -> proxy
domain(domain: ohmyswift.ru, domain: paperpaper.ru) -> proxy
domain(domain: pimpletv.ru, domain: polit.ru, domain: republic.ru) -> proxy
domain(domain: scryde.ru, domain: seasonvar.ru) -> proxy
domain(domain: sklatchiki.ru, domain: the-village.ru, domain: theins.ru) -> proxy
domain(domain: tvrain.ru, domain: zahav.ru) -> proxy

# Route the Ukrainian domain zone through the proxy
domain(domain: ua) -> proxy

# Route Russian services and domains directly
domain(geosite:category-ru) -> direct

# Route Russian IP addresses directly unless an earlier exception matches
ip(geoip:ru) -> direct

# Additional Russian service domains
domain(domain: 1018213540.rsc.cdn77.org, domain: bitrix.info) -> direct

# Route local hostnames and private IP addresses directly
domain(geosite:private) -> direct
ip(geoip:private) -> direct`;
