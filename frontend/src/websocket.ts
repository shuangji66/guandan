import { ref, onMounted, onUnmounted } from 'vue'
import type { ServerMessage } from '@/types'

class WebSocketService {
  private ws: WebSocket | null = null
  private url: string
  public isConnected = ref(false)
  public lastError = ref<string | null>(null)
  private messageHandlers: Record<string, ((data: any) => void)[]> = {}
  private reconnectTimer: number | null = null

  constructor(url: string) {
    this.url = url
    this.connect()
  }

  private connect() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return
    this.ws = new WebSocket(this.url)
    this.ws.onopen = () => {
      this.isConnected.value = true
      this.lastError.value = null
      console.log('[WebSocket] connected')
    }
    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as ServerMessage
        console.log('[WebSocket] received:', msg.type, msg.payload)
        const handlers = this.messageHandlers[msg.type] || []
        handlers.forEach((fn) => fn(msg.payload))
      } catch (e) {
        console.error('[WebSocket] parse error', e)
      }
    }
    this.ws.onclose = () => {
      this.isConnected.value = false
      console.log('[WebSocket] disconnected')
      // 自动重连
      if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
      this.reconnectTimer = setTimeout(() => this.connect(), 3000)
    }
    this.ws.onerror = (e) => {
      this.lastError.value = 'WebSocket error'
      console.error('[WebSocket] error', e)
    }
  }

  sendMessage(type: string, payload: any) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('[WebSocket] not open, message not sent:', type)
      return
    }
    const msg = JSON.stringify({ type, payload })
    console.log('[WebSocket] sending:', type, payload)
    this.ws.send(msg)
  }

  onMessage(type: string, callback: (data: any) => void) {
    if (!this.messageHandlers[type]) this.messageHandlers[type] = []
    this.messageHandlers[type].push(callback)
  }

  offMessage(type: string, callback?: (data: any) => void) {
    if (!callback) delete this.messageHandlers[type]
    else {
      const handlers = this.messageHandlers[type]
      if (handlers) {
        const idx = handlers.indexOf(callback)
        if (idx !== -1) handlers.splice(idx, 1)
      }
    }
  }

  disconnect() {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
  }
}

// 导出单例
export const wsService = new WebSocketService('/ws')