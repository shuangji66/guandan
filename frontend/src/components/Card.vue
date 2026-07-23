<template>
  <div
    :class="[
      'relative bg-white rounded shadow-md border border-gray-300 flex flex-col justify-between select-none cursor-pointer transition-transform',
      sizeClasses,
      selectClasses,
      colorClass,
      highlightClasses,
    ]"
    @click="onClick"
  >
    <!-- 王 -->
    <template v-if="isJoker">
      <div class="text-center w-full h-full flex items-center justify-center font-bold writing-vertical">
        {{ card.rank === Rank.SmallJoker ? '小王' : '大王' }}
      </div>
    </template>
    <!-- 普通牌 -->
    <template v-else>
      <div class="font-bold text-left leading-none">{{ rankLabel }}</div>
      <div class="absolute inset-0 flex items-center justify-center text-2xl opacity-20 pointer-events-none">
        {{ suitSymbol }}
      </div>
      <div class="text-right leading-none self-end">{{ suitSymbol }}</div>
      <!-- 级牌标记 -->
      <div v-if="card.isLevelCard" class="absolute top-0 right-0 w-2 h-2 bg-yellow-400 rounded-full"></div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Card as CardType, Suit, Rank } from '@/types'

const props = defineProps<{
  card: CardType
  selected?: boolean
  onClick?: () => void
  small?: boolean
  isHighlighted?: boolean
}>()

const emit = defineEmits<{
  (e: 'click'): void
}>()

const Rank = Rank
const Suit = Suit

const isRed = computed(() => 
  props.card.suit === Suit.Hearts || 
  props.card.suit === Suit.Diamonds || 
  props.card.rank === Rank.BigJoker
)

const isJoker = computed(() => props.card.suit === Suit.Joker)

const suitSymbol = computed(() => {
  switch (props.card.suit) {
    case Suit.Spades: return '♠'
    case Suit.Hearts: return '♥'
    case Suit.Clubs: return '♣'
    case Suit.Diamonds: return '♦'
    case Suit.Joker: return 'J'
    default: return ''
  }
})

const rankLabel = computed(() => {
  switch (props.card.rank) {
    case Rank.Two: return '2'
    case Rank.Three: return '3'
    case Rank.Four: return '4'
    case Rank.Five: return '5'
    case Rank.Six: return '6'
    case Rank.Seven: return '7'
    case Rank.Eight: return '8'
    case Rank.Nine: return '9'
    case Rank.Ten: return '10'
    case Rank.Jack: return 'J'
    case Rank.Queen: return 'Q'
    case Rank.King: return 'K'
    case Rank.Ace: return 'A'
    default: return ''
  }
})

const sizeClasses = computed(() => 
  props.small ? 'w-8 h-12 text-xs p-1' : 'w-16 h-24 text-base p-2 hover:-translate-y-2'
)

const selectClasses = computed(() => 
  props.selected ? 'ring-2 ring-blue-500 -translate-y-4' : ''
)

const colorClass = computed(() => 
  isRed.value ? 'text-red-600' : 'text-black'
)

const highlightClasses = computed(() => 
  props.isHighlighted 
    ? 'ring-4 ring-green-400 shadow-[0_0_15px_rgba(74,222,128,0.7)] animate-pulse' 
    : ''
)

const onClick = () => {
  emit('click')
}
</script>