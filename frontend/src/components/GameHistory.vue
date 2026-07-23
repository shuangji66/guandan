<template>
  <div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
    <div class="bg-gray-900 rounded-lg shadow-2xl w-11/12 max-w-4xl h-5/6 flex flex-col border-2 border-gray-700">
      <!-- 头部 -->
      <div class="flex items-center justify-between p-4 border-b border-gray-700">
        <div>
          <h2 class="text-2xl font-bold text-white">游戏历史记录</h2>
          <p class="text-sm text-gray-400">第 {{ currentRound }} 局 · 共 {{ filteredHistory.length }} 条记录</p>
        </div>
        <button @click="close" class="text-gray-400 hover:text-white text-3xl leading-none px-3 py-1">×</button>
      </div>

      <!-- 过滤 -->
      <div class="p-4 border-b border-gray-700 space-y-3">
        <input
          type="text"
          v-model="searchTerm"
          placeholder="搜索玩家名或事件..."
          class="w-full px-4 py-2 bg-gray-800 text-white rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none"
        />
        <div class="flex flex-wrap gap-2">
          <button
            @click="filter = 'all'"
            :class="[
              'px-3 py-1 rounded-full text-sm font-medium transition',
              filter === 'all' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700',
            ]"
          >
            全部
          </button>
          <button
            v-for="type in eventTypes"
            :key="type"
            @click="filter = type"
            :class="[
              'px-3 py-1 rounded-full text-sm font-medium transition',
              filter === type ? `${getEventTypeInfo(type).bgColor} ${getEventTypeInfo(type).color} border border-current` : 'bg-gray-800 text-gray-400 hover:bg-gray-700',
            ]"
          >
            {{ getEventTypeInfo(type).name }}
          </button>
        </div>
      </div>

      <!-- 列表 -->
      <div ref="historyContainer" class="flex-1 overflow-y-auto p-4 space-y-2" @scroll="handleScroll">
        <div v-if="filteredHistory.length === 0" class="text-center text-gray-500 py-8">
          {{ searchTerm || filter !== 'all' ? '没有匹配的记录' : '暂无历史记录' }}
        </div>
        <div
          v-for="entry in filteredHistory"
          :key="entry.id"
          :class="[getEventTypeInfo(entry.type).bgColor, 'rounded-lg p-3 border border-gray-700/50 hover:border-gray-600 transition']"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="flex-1">
              <div class="flex items-center gap-2 mb-1">
                <span :class="['text-xs font-bold', getEventTypeInfo(entry.type).color, 'px-2 py-0.5 rounded']">
                  {{ getEventTypeInfo(entry.type).name }}
                </span>
                <span v-if="entry.playerName" class="text-xs text-gray-400">{{ entry.playerName }}</span>
                <span class="text-xs text-gray-500">{{ formatTime(entry.timestamp) }}</span>
              </div>
              <p class="text-white text-sm leading-relaxed">{{ entry.message }}</p>
            </div>
          </div>
        </div>
        <div ref="historyEnd" />
      </div>

      <!-- 底部 -->
      <div class="p-4 border-t border-gray-700 flex items-center justify-between">
        <label class="flex items-center gap-2 text-sm text-gray-400">
          <input type="checkbox" v-model="autoScroll" class="rounded" />
          自动滚动到最新
        </label>
        <button @click="close" class="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition">
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import type { HistoryEntry } from '@/types'

const props = defineProps<{
  history: HistoryEntry[]
  currentRound: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const filter = ref<'all' | string>('all')
const searchTerm = ref('')
const autoScroll = ref(true)
const historyContainer = ref<HTMLElement | null>(null)
const historyEnd = ref<HTMLElement | null>(null)

const eventTypes = [
  'GameStart', 'PhaseChange', 'Play', 'Pass', 'Tribute', 'ReturnTribute',
  'SkillUse', 'RoundEnd', 'PlayerFinish', 'GameEnd', 'LevelUp',
]

const getEventTypeInfo = (type: string) => {
  const map: Record<string, { name: string; color: string; bgColor: string }> = {
    GameStart: { name: '游戏开始', color: 'text-green-400', bgColor: 'bg-green-900/30' },
    PhaseChange: { name: '阶段变化', color: 'text-blue-400', bgColor: 'bg-blue-900/30' },
    Play: { name: '出牌', color: 'text-yellow-400', bgColor: 'bg-yellow-900/30' },
    Pass: { name: '过牌', color: 'text-gray-400', bgColor: 'bg-gray-900/30' },
    Tribute: { name: '进贡', color: 'text-purple-400', bgColor: 'bg-purple-900/30' },
    ReturnTribute: { name: '还贡', color: 'text-pink-400', bgColor: 'bg-pink-900/30' },
    SkillUse: { name: '技能', color: 'text-cyan-400', bgColor: 'bg-cyan-900/30' },
    RoundEnd: { name: '回合结束', color: 'text-orange-400', bgColor: 'bg-orange-900/30' },
    PlayerFinish: { name: '出完', color: 'text-red-400', bgColor: 'bg-red-900/30' },
    GameEnd: { name: '游戏结束', color: 'text-red-500', bgColor: 'bg-red-900/50' },
    LevelUp: { name: '升级', color: 'text-green-500', bgColor: 'bg-green-900/50' },
  }
  return map[type] || { name: type, color: 'text-gray-400', bgColor: 'bg-gray-900/30' }
}

const formatTime = (timestamp: number) => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

const filteredHistory = computed(() => {
  return props.history.filter(entry => {
    const matchesFilter = filter.value === 'all' || entry.type === filter.value
    const matchesSearch = searchTerm.value === '' || 
      entry.message.toLowerCase().includes(searchTerm.value.toLowerCase()) ||
      (entry.playerName && entry.playerName.toLowerCase().includes(searchTerm.value.toLowerCase()))
    return matchesFilter && matchesSearch
  })
})

const handleScroll = (e: Event) => {
  const el = e.target as HTMLElement
  const isAtBottom = el.scrollHeight - el.scrollTop <= el.clientHeight + 50
  autoScroll.value = isAtBottom
}

// 自动滚动到底部
watch(filteredHistory, () => {
  if (autoScroll.value) {
    nextTick(() => {
      historyEnd.value?.scrollIntoView({ behavior: 'smooth' })
    })
  }
}, { deep: true })

const close = () => {
  emit('close')
}
</script>