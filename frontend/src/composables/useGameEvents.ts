import { useRoomStore } from '@/store/room'
import { useGameStore } from '@/store/game'
import { useUIStore } from '@/store/ui'
import type { ServerMessage } from '@/types'

export function useGameEvents() {
  const roomStore = useRoomStore()
  const gameStore = useGameStore()
  const uiStore = useUIStore()

  const handlers: Record<string, (payload: any) => void> = {
    roomState: (payload) => {
      roomStore.updateRoomState(payload)
    },
    gameState: (payload) => {
      gameStore.updateGameState(payload)
    },
    chatMessage: (payload) => {
      roomStore.addChatMessage(payload)
    },
    error: (payload) => {
      uiStore.showError(payload)
    },
    gameOver: (payload) => {
      gameStore.setGameOver(payload.winners)
    },
    matchOver: (payload) => {
      uiStore.showMatchOver(payload)
    },
    gameTerminated: () => {
      gameStore.resetGame()
      uiStore.showError('游戏已被房主强制结束')
    },
    roomList: (payload) => {
      roomStore.setRoomList(payload)
    },
    historyUpdate: (payload) => {
      gameStore.addHistoryEntry(payload)
    },
  }

  function handleMessage(msg: ServerMessage) {
    const handler = handlers[msg.type]
    if (handler) {
      handler(msg.payload)
    } else {
      console.warn('Unknown message type:', msg.type)
    }
  }

  return { handleMessage }
}