export const template = `default: proxy
# write your own rules below
domain(domain:mail.qq.com)->direct

domain(geosite:google-scholar)->proxy
domain(geosite:category-scholar-!cn, geosite:category-scholar-cn)->direct
domain(geosite:geolocation-!cn, geosite:google)->proxy
domain(geosite:cn)->direct
ip(geoip:hk,geoip:mo)->proxy
ip(geoip:private, geoip:cn)->direct`;
