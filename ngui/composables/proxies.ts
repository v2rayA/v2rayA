export const proxies
  = useLocalStorage<{
    currentOutbound: string
    outbounds: any[]
    subs: any[]
    servers: any[]
  }>('proxies', {
    currentOutbound: 'proxy',
    outbounds: [],
    subs: [],
    servers: []
  })
