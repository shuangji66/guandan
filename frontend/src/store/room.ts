import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RoomState, Player, RoomInfo, ChatMessage } from '@/types'

export const useRoomStore = defineStore('room', () => {
  const inRoom = ref(false)
  const roomState = ref<RoomState | null>(null)
  const mySeat = ref(-1)
  const chatMessages = ref<ChatMessage[]>([])
  const roomList = ref<RoomInfo[]>([])
  const currentPlayerName = ref('')

  function setCurrentPlayerName(name: string) {
    currentPlayerName.value = name
  }

  function updateRoomState(state: RoomState) {
    console.log('[RoomStore] updateRoomState:', state)
    roomState.value = state
    inRoom.value = true
    const me = state.players.find(
      (p): p is Player => p !== null && p.name === currentPlayerName.value
    )
    console.log('[RoomStore] currentPlayerName:', currentPlayerName.value, 'me:', me)
    if (me) mySeat.value = me.seatIndex
    else mySeat.value = -1
  }

  function addChatMessage(msg: ChatMessage) {
    chatMessages.value.push(msg)
  }

  function setRoomList(list: RoomInfo[]) {
    roomList.value = list
  }

  function resetRoom() {
    inRoom.value = false
    roomState.value = null
    mySeat.value = -1
    chatMessages.value = []
  }

  return {
    inRoom,
    roomState,
    mySeat,
    chatMessages,
    roomList,
    currentPlayerName,
    setCurrentPlayerName,
    updateRoomState,
    addChatMessage,
    setRoomList,
    resetRoom,
  }
})