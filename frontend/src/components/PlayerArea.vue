<template>
  <div
    :class="[
      'flex flex-col items-center p-4 rounded-lg transition-colors',
      isTeammate ? 'bg-blue-900/40 border-2 border-blue-400' : 'bg-black/20',
      !gameExists && !player ? 'cursor-pointer hover:bg-white/10' : '',
    ]"
    :style="positionStyle"
    @click="onClickSeat"
  >
    <!-- 聊天气泡 -->
    <div v-if="chatBubble" class="absolute -top-16 left-1/2 -translate-x-1/2 z-50 animate-bounce-in">
      <div class="relative bg-white text-gray-800 px-4 py-2 rounded-xl shadow-lg max-w-48 text-sm font-medium whitespace-pre-wrap">
        {{ chatBubble }}
        <div class="absolute -bottom-2 left-1/2 -translate-x-1/2 w-0 h-0 border-l-8 border-r-8 border-t-8 border-l-transparent border-r-transparent border-t-white"></div>
      </div>
    </div>

    <!-- 头像 -->
    <div class="w-12 h-12 bg-gray-300 rounded-full flex items-center justify-center mb-2 relative">
      <span>{{ avatarText }}</span>
      <div v-if="isTeammate" class="absolute -top-1 -right-1 bg-blue-500 text-xs text-white px-1 rounded">友</div>
      <div v-if="isOpponent" class="absolute -top-1 -right-1 bg-red-500 text-xs text-white px-1 rounded">敌</div>
      <div v-if="isHost" class="absolute -bottom-1 -right-1 text-xs bg-yellow-500 text-black px-1 rounded font-bold border border-white">Host</div>
      <div
        v-if="winnerLabel"
        :class="[
          'absolute -top-2 left-1/2 -translate-x-1/2 text-white text-xs px-2 py-0.5 rounded-full font-bold shadow-lg border-2 border-white animate-pulse',
          winnerColor,
        ]"
      >
        {{ winnerLabel }}
      </div>
    </div>

    <!-- 名称 -->
    <div class="text-white font-bold flex items-center gap-2">
      <span>{{ playerName }}</span>
      <span v-if="isDisconnected" class="text-red-500 text-xs font-bold bg-white px-1 rounded animate-pulse">OFF</span>
    </div>

    <!-- 牌数 -->
    <div v-if="gameExists" class="text-yellow-400">Cards: {{ handCount }}</div>
    <div v-if="player && player.isReady && !gameExists" class="text-green-400 text-sm">Ready</div>

    <!-- 动作 -->
    <div v-if="gameExists && action" class="mt-2 flex flex-col items-center">
      <template v-if="action.type === 'pass'">
        <div class="text-gray-400 font-bold text-sm bg-gray-700/50 px-3 py-1 rounded">过</div>
      </template>
      <template v-else>
        <div class="text-green-400 text-xs mb-1">{{ action.hand?.type || '出牌' }}</div>
        <div v-if="action.cards" class="flex gap-0.5 mt-1">
          <div
            v-for="(card, idx) in action.cards.slice(0, 6)"
            :key="idx"
            class="w-6 h-8 bg-white rounded text-xs flex items-center justify-center font-bold border border-gray-300"
            :style="{ color: (card.suit === Suit.Hearts || card.suit === Suit.Diamonds) ? 'red' : 'black' }"
          >
            {{ getCardShortLabel(card) }}
          </div>
          <span v-if="action.cards.length > 6" class="text-white text-xs">+{{ action.cards.length - 6 }}</span>
        </div>
      </template>
    </div>

    <!-- 轮到谁 -->
    <div v-if="gameExists && isMyTurn && !action" class="animate-bounce text-red-500 font-bold mt-2">Thinking...</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Player, Card, RoundAction } from '@/types'
import { Suit, Rank } from '@/types'

const props = defineProps<{
  player: Player | null
  seat: number
  handCount: number
  isTeammate: boolean
  isOpponent: boolean
  action?: RoundAction
  winnerLabel?: string | null
  chatBubble?: string | null
  isMyTurn?: boolean
  gameExists?: boolean
  // 定位属性（可选）
  top?: string
  left?: string
  right?: string
  transform?: string
}>()

const emit = defineEmits<{
  (e: 'clickSeat', seat: number): void
}>()

// 构建内联样式对象，强制 position: absolute
const positionStyle = computed(() => {
  const style: Record<string, string> = {
    position: 'absolute',
  }
  if (props.top) style.top = props.top
  if (props.left) style.left = props.left
  if (props.right) style.right = props.right
  if (props.transform) style.transform = props.transform
  return style
})

const avatarText = computed(() => {
  if (props.player) return props.player.name[0].toUpperCase()
  return props.gameExists ? '?' : '+'
})

const playerName = computed(() => {
  if (props.player) return props.player.name
  return props.gameExists ? 'Waiting...' : '点击入座'
})

const isHost = computed(() => props.player?.seatIndex === 0)
const isDisconnected = computed(() => props.player?.isDisconnected)

const winnerColor = computed(() => {
  const map: Record<string, string> = {
    '头游': 'bg-yellow-500',
    '二游': 'bg-orange-500',
    '三游': 'bg-purple-500',
    '末游': 'bg-gray-500',
  }
  return props.winnerLabel ? map[props.winnerLabel] || 'bg-gray-500' : ''
})

const getCardShortLabel = (card: Card) => {
  if (card.rank === Rank.SmallJoker) return '🃏'
  if (card.rank === Rank.BigJoker) return '🃟'
  const rankMap = ['', '', '2', '3', '4', '5', '6', '7', '8', '9', '10', 'J', 'Q', 'K', 'A']
  return rankMap[card.rank] || '?'
}

const onClickSeat = () => {
  if (!props.gameExists && !props.player) {
    emit('clickSeat', props.seat)
  }
}
</script>