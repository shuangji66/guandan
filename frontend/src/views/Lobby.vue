<template>
  <div class="flex flex-col items-center justify-center min-h-screen bg-[#1e1e1e] text-gray-300 p-4">
    <h1 class="text-5xl font-bold mb-8 text-[#519aba] font-mono">VS Code - GuanDan</h1>

    <div class="flex gap-6 items-start">
      <!-- 加入表单 -->
      <form @submit.prevent="handleJoin"
        class="bg-[#252526] p-8 rounded-lg shadow-xl border border-[#333333] text-gray-300 flex flex-col gap-4 w-80">
        <div>
          <label class="block text-sm font-bold mb-2 text-[#9cdcfe]">Username</label>
          <input
            type="text"
            v-model="name"
            class="w-full bg-[#3c3c3c] border border-[#3c3c3c] p-2 rounded text-white focus:outline-none focus:border-[#007acc]"
            placeholder="Enter name..."
            maxlength="10"
            required
          />
        </div>
        <div>
          <label class="block text-sm font-bold mb-2 text-[#9cdcfe]">Room ID</label>
          <input
            type="text"
            v-model="roomId"
            class="w-full bg-[#3c3c3c] border border-[#3c3c3c] p-2 rounded text-white focus:outline-none focus:border-[#007acc]"
            placeholder="Default: default"
          />
        </div>
        <button type="submit" class="bg-[#0e639c] text-white py-2 rounded hover:bg-[#1177bb] font-bold mt-2">
          Connect
        </button>
        <button
          type="button"
          @click="showRoomList = !showRoomList"
          class="bg-[#3c3c3c] text-gray-300 py-2 rounded hover:bg-[#4a4a4a] font-bold border border-[#555555]"
        >
          {{ showRoomList ? '隐藏房间列表' : '查看房间列表' }}
        </button>
      </form>

      <!-- 房间列表 -->
      <div v-if="showRoomList"
        class="bg-[#252526] p-6 rounded-lg shadow-xl border border-[#333333] w-96 max-h-96 overflow-y-auto">
        <h2 class="text-xl font-bold mb-4 text-[#569cd6]">活跃房间</h2>
        <div v-if="roomList.length === 0" class="text-gray-500 text-center py-8">暂无活跃房间</div>
        <div v-else class="flex flex-col gap-2">
          <div v-for="room in roomList" :key="room.id"
            class="bg-[#1e1e1e] p-4 rounded border border-[#3c3c3c] hover:border-[#007acc] transition-colors cursor-pointer"
            @click="quickJoin(room.id)">
            <div class="flex justify-between items-start mb-2">
              <div class="font-bold text-[#9cdcfe]">房间: {{ room.id }}</div>
              <div :class="room.inGame ? 'bg-red-900/50 text-red-300' : 'bg-green-900/50 text-green-300'"
                class="text-xs px-2 py-1 rounded">
                {{ room.inGame ? '游戏中' : '等待中' }}
              </div>
            </div>
            <div class="text-sm text-gray-400 flex justify-between">
              <span>房主: {{ room.hostName }}</span>
              <span>{{ room.playerCount }}/{{ room.maxPlayers }} 人</span>
            </div>
            <div class="text-xs text-gray-500 mt-1">
              模式: {{ room.gameMode === GameMode.Normal ? '普通' : '技能' }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '@/store/room'
import { useWebSocket } from '@/composables/useWebSocket'
import { useGameEvents } from '@/composables/useGameEvents'
import { GameMode } from '@/types'

const router = useRouter()
const roomStore = useRoomStore()
const { sendMessage, onMessage, offMessage } = useWebSocket()
const { handleMessage } = useGameEvents()

const name = ref('')
const roomId = ref('default')
const showRoomList = ref(false)
const roomList = ref(roomStore.roomList)

// 监听进入房间
watch(() => roomStore.inRoom, (val) => {
  if (val) router.push('/game')
})

onMounted(() => {
  // 注册所有需要处理的消息类型，交给 handleMessage 统一分发
  const types = ['roomState', 'error', 'roomList']
  types.forEach(type => {
    onMessage(type, (payload) => {
      handleMessage({ type, payload })
    })
  })

  // 获取房间列表
  sendMessage('getRoomList', {})

  // 定期刷新房间列表
  const interval = setInterval(() => {
    if (showRoomList.value) {
      sendMessage('getRoomList', {})
    }
  }, 3000)

  onUnmounted(() => {
    clearInterval(interval)
    types.forEach(type => offMessage(type))
  })
})

function handleJoin() {
  if (!name.value.trim()) return
  console.log('[Lobby] handleJoin, name:', name.value, 'roomId:', roomId.value)
  roomStore.setCurrentPlayerName(name.value)
  sendMessage('joinRoom', { playerName: name.value, roomId: roomId.value || 'default' })
}

function quickJoin(roomId: string) {
  if (!name.value.trim()) {
    alert('请先输入用户名')
    return
  }
  roomStore.setCurrentPlayerName(name.value)  // 新增
  sendMessage('joinRoom', { playerName: name.value, roomId })
}
</script>