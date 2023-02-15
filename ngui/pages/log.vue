<script lang="ts" setup>
import { parseURL } from 'ufo'

definePageMeta({ middleware: ['auth'] })

const message = $ref<string[]>([])

const parsed = parseURL(system.value.api)
const socket = new WebSocket(`ws://${parsed.host}/api/message?Authorization=${encodeURIComponent(user.value.token)}`)

socket.onmessage = (msg) => { message.push(msg.data) }
</script>

<template>
  <pre class="bg-black text-white rounded-md">
    <code>
      {{ Array.isArray(message) && message.length === 0 ? 'Empty' : message }}
    </code>
  </pre>
</template>
