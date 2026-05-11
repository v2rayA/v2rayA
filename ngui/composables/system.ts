// 优先从 localStorage 读取自定义后端地址（与 /gui 的 backendAddress 保持一致）
const defaultApi = import.meta.client
  ? (localStorage.getItem('backendAddress') || 'http://127.0.0.1:2017')
  : 'http://127.0.0.1:2017'

export const system = useLocalStorage('system', {
  api: defaultApi,
  running: false,
  networkPaused: false,
  connect: '',
  docker: false,
  version: '',
  lite: 'false',
  gfwlist: '',
  os: '',
  isRoot: false
})
