// Iranian sites and addresses go direct, everything else through the proxy.
// Only categories in the v2fly geosite.dat and geoip.dat: category-ir and ir.
export const iran = `default: proxy
# Route local hostnames and private IP addresses directly
domain(geosite:private) -> direct
ip(geoip:private) -> direct

# Route Iranian domains directly
domain(geosite:category-ir) -> direct

# Route Iranian IP addresses directly
ip(geoip:ir) -> direct`;
