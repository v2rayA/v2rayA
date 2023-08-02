export const system = useLocalStorage('system', {
  api: 'http://127.0.0.1:2017',
  running: false,
  connect: '',
  docker: false,
  version: '',
  lite: 'false',
  gfwlist: ''
})
