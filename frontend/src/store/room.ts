import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { RoomState, Player, RoomInfo, ChatMessage } from '@/types'

export const useRoomStore = defineStore('room', () => {
  const inRoom = ref(false)
  const roomState = ref<RoomState | null>(null)
  const mySeat = ref(-1)
  const chatMessages = ref<ChatMessage[]>([])
  const roomList = ref<RoomInfo[]>([])

  function updateRoomState(state: RoomState) {
    roomState.value = state
    inRoom.value = true
    const me = state.players.find((p) => p && p.id === '')
    if (me) mySeat.value = me.seatIndex
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
    updateRoomState,
    addChatMessage,
    setRoomList,
    resetRoom,
  }
})