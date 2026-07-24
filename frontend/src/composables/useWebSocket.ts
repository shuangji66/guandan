import { wsService } from '@/websocket'
import type { ServerMessage } from '@/types'

export function useWebSocket() {
  return {
    isConnected: wsService.isConnected,
    lastError: wsService.lastError,
    sendMessage: wsService.sendMessage.bind(wsService),
    onMessage: wsService.onMessage.bind(wsService),
    offMessage: wsService.offMessage.bind(wsService),
  }
}