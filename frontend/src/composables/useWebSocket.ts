import { ref, onMounted, onUnmounted } from 'vue'
import type { ServerMessage } from '@/types'

export function useWebSocket(url: string) {
  const isConnected = ref(false)
  const lastError = ref<string | null>(null)
  let ws: WebSocket | null = null
  const messageHandlers: Record<string, ((data: any) => void)[]> = {}

  function connect() {
    ws = new WebSocket(url)
    ws.onopen = () => {
      isConnected.value = true
      lastError.value = null
      console.log('WebSocket connected')
    }
    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as ServerMessage
        const handlers = messageHandlers[msg.type] || []
        handlers.forEach((fn) => fn(msg.payload))
      } catch (e) {
        console.error('Invalid message', e)
      }
    }
    ws.onclose = () => {
      isConnected.value = false
      console.log('WebSocket disconnected')
      // 自动重连
      setTimeout(() => connect(), 3000)
    }
    ws.onerror = (e) => {
      lastError.value = 'WebSocket error'
      console.error(e)
    }
  }

  function sendMessage(type: string, payload: any) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not open')
      return
    }
    ws.send(JSON.stringify({ type, payload }))
  }

  function onMessage(type: string, callback: (data: any) => void) {
    if (!messageHandlers[type]) {
      messageHandlers[type] = []
    }
    messageHandlers[type].push(callback)
  }

  function offMessage(type: string, callback?: (data: any) => void) {
    if (!callback) {
      delete messageHandlers[type]
    } else {
      const handlers = messageHandlers[type]
      if (handlers) {
        const idx = handlers.indexOf(callback)
        if (idx !== -1) handlers.splice(idx, 1)
      }
    }
  }

  onMounted(() => connect())
  onUnmounted(() => {
    if (ws) ws.close()
  })

  return { isConnected, lastError, sendMessage, onMessage, offMessage }
}