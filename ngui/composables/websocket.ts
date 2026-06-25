import { ref, onUnmounted } from 'vue'

export type WsMessage = {
  type: string
  body: any
}

/**
 * WebSocket composable for real-time message communication with v2rayA backend.
 * Features:
 * - Automatic reconnection with exponential backoff (1s, 2s, 4s... max 30s)
 * - Message type filtering
 * - Lifecycle management (auto-disconnect on unmount)
 *
 * Usage:
 * ```ts
 * const ws = useWebSocket()
 * ws.onMessage('running_state', (msg) => { ... })
 * ws.connect()
 * ```
 */
export function useWebSocket() {
  const isConnected = ref(false)
  const lastMessage = ref<WsMessage | null>(null)

  let ws: WebSocket | null = null
  let retries = 0
  let messageHandlers: Map<string, Set<(msg: WsMessage) => void>> = new Map()
  let genericHandlers: Set<(msg: WsMessage) => void> = new Set()

  const getBackendUrl = (): string => {
    let url = localStorage.getItem('backendAddress') || 'http://127.0.0.1:2017'
    if (!url.trim() || url.startsWith('/')) {
      url = location.protocol + '//' + location.host + url
    }
    const u = new URL(url)
    const protocol = u.protocol === 'https:' ? 'wss:' : 'ws:'
    const token = localStorage.getItem('token') || ''
    return `${protocol}//${u.host}/api/message?Authorization=${encodeURIComponent(token)}`
  }

  const connect = () => {
    // Close existing connection
    if (ws) {
      ws.onclose = null
      ws.onmessage = null
      ws.onopen = null
      ws.close()
      ws = null
    }

    try {
      const url = getBackendUrl()
      ws = new WebSocket(url)

      ws.onopen = () => {
        isConnected.value = true
        retries = 0 // Reset retry count on successful connection
      }

      ws.onmessage = (event: MessageEvent) => {
        try {
          const msg: WsMessage = JSON.parse(event.data)
          lastMessage.value = msg

          // Notify generic handlers
          genericHandlers.forEach(handler => handler(msg))

          // Notify type-specific handlers
          const handlers = messageHandlers.get(msg.type)
          if (handlers) {
            handlers.forEach(handler => handler(msg))
          }
        } catch (e) {
          console.error('[WebSocket] Failed to parse message:', e)
        }
      }

      ws.onclose = () => {
        isConnected.value = false
        ws = null
        // Exponential backoff reconnection: 1s, 2s, 4s, 8s... max 30s
        const delay = Math.min(1000 * Math.pow(2, retries), 30000)
        retries++
        setTimeout(() => {
          if (ws === null) {
            connect()
          }
        }, delay)
      }

      ws.onerror = () => {
        // onerror will be followed by onclose, so we let onclose handle reconnection
        console.warn('[WebSocket] Connection error, will retry...')
      }
    } catch (e) {
      console.error('[WebSocket] Failed to create connection:', e)
      // Retry after delay
      const delay = Math.min(1000 * Math.pow(2, retries), 30000)
      retries++
      setTimeout(() => connect(), delay)
    }
  }

  /**
   * Register a handler for a specific message type
   */
  const onMessage = (type: string, handler: (msg: WsMessage) => void) => {
    if (!messageHandlers.has(type)) {
      messageHandlers.set(type, new Set())
    }
    messageHandlers.get(type)!.add(handler)
  }

  /**
   * Remove a handler for a specific message type
   */
  const offMessage = (type: string, handler: (msg: WsMessage) => void) => {
    const handlers = messageHandlers.get(type)
    if (handlers) {
      handlers.delete(handler)
      if (handlers.size === 0) {
        messageHandlers.delete(type)
      }
    }
  }

  /**
   * Register a handler for all message types
   */
  const onAnyMessage = (handler: (msg: WsMessage) => void) => {
    genericHandlers.add(handler)
  }

  /**
   * Remove a generic handler
   */
  const offAnyMessage = (handler: (msg: WsMessage) => void) => {
    genericHandlers.delete(handler)
  }

  /**
   * Disconnect WebSocket
   */
  const disconnect = () => {
    if (ws) {
      ws.onclose = null // Prevent auto-reconnect
      ws.close()
      ws = null
    }
    isConnected.value = false
    retries = 0
  }

  // Auto-disconnect on component unmount
  onUnmounted(() => {
    disconnect()
  })

  return {
    isConnected,
    lastMessage,
    connect,
    disconnect,
    onMessage,
    offMessage,
    onAnyMessage,
    offAnyMessage,
  }
}
