<script lang="ts" setup>
import { parseURL } from 'ufo'
const { data: row, subID } = defineProps<{ data: any, subID: number }>()

let isVisible = $ref(false)
let serverInfo = $ref<any>()

const viewServer = async() => {
  isVisible = true
  const params = JSON.stringify({
    id: row.id,
    _type: row._type,
    sub: subID! - 1
  })

  const { data } = await useV2Fetch(`sharingAddress?touch=${params}`).get().json()

  /* ss://BASE64(method:password)@server:port#name */
  /* ssr://server:port:proto:method:obfs:URLBASE64(password)/?remarks=URLBASE64(remarks)&protoparam=URLBASE64(protoparam)&obfsparam=URLBASE64(obfsparam)) */
  /* trojan://password@server:port?allowInsecure=1&sni=sni#URIESCAPE(name) */

  serverInfo = parseURL(data.value.data.sharingAddress)

  serverInfo = {
    ...serverInfo,
    protocol: serverInfo.protocol.slice(0, -1),
    name: decodeURIComponent(serverInfo.hash).slice(1)
  }

  switch (serverInfo.protocol) {
    case 'ss': {
      const auth = atob(serverInfo.auth).split(':')
      const address = serverInfo.host.split(':')

      serverInfo = {
        ...serverInfo,
        host: address[0],
        port: address[1],
        method: auth[0],
        password: auth[1]
      }

      delete serverInfo.auth
      break
    }
    case 'trojan': {
      const address = serverInfo.host.split(':')

      serverInfo = {
        ...serverInfo,
        host: address[0],
        port: address[1],
        password: serverInfo.auth,
        name: decodeURIComponent(serverInfo.hash).slice(1)
      }

      delete serverInfo.auth
      break
    }
    case 'ssr': {
      const auth = atob(serverInfo.auth).split(':')
      const address = serverInfo.host.split(':')

      serverInfo = {
        ...serverInfo,
        host: address[0],
        port: address[1],
        method: auth[0],
        password: auth[1],
        protocol: serverInfo.protocol,
        obfs: serverInfo.obfs,
        name: decodeURIComponent(serverInfo.hash).slice(1)
      }

      delete serverInfo.auth
      break
    }
    case 'vless': {
      const auth = atob(serverInfo.auth).split(':')
      const address = serverInfo.host.split(':')

      serverInfo = {
        ...serverInfo,
        host: address[0],
        port: address[1],
        method: auth[0],
        password: auth[1],
        name: decodeURIComponent(serverInfo.hash).slice(1)
      }

      delete serverInfo.auth
      break
    }
    case 'vmess': {
      const parsed: {
        ps: string
        add: string
        port: string
        id: string
        aid: string
        scy: string
        net: string
        type: string
        host: string
        sni: string
        path: string
        tls: string
        allowInsecure: boolean
        v: boolean
        protocol: string
      } = JSON.parse(atob(serverInfo.host))
      serverInfo = {
        ...serverInfo,
        name: parsed.ps,
        ...parsed
      }

      delete serverInfo.host
      delete serverInfo.ps
      break
    }
    default: break
  }

  delete serverInfo.hash
  Object.keys(serverInfo).forEach((key) => { if (serverInfo[key] === '') delete serverInfo[key] })
}
</script>

<template>
  <ElButton size="small" class="mr-3" @click="viewServer">
    <UnoIcon class="ri:file-code-line mr-1" />{{ $t("operations.view") }}
  </ElButton>

  <ElDialog v-model="isVisible">
    <ElForm>
      <ElFormItem
        v-for="(v, k) in serverInfo"
        :key="k"
        :label="k.toString()"
      >
        <ElInput :value="v" disabled />
      </ElFormItem>
    </ElForm>
  </ElDialog>
</template>
