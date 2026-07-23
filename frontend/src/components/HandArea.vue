<template>
  <div :class="['px-8 flex items-end justify-center pointer-events-auto transition-all duration-300', viewMode === 'normal' ? 'h-32 -space-x-8' : 'h-64 gap-1']">
    <!-- 普通视图 -->
    <template v-if="viewMode === 'normal'">
      <Card
        v-for="card in cards"
        :key="card.id"
        :card="card"
        :selected="selectedIds.includes(card.id)"
        :is-highlighted="highlightedIds.has(card.id)"
        @click="onToggleSelect(card.id)"
      />
    </template>

    <!-- 堆叠视图 -->
    <template v-else>
      <div
        v-for="(col, idx) in stackedMatrix"
        :key="idx"
        class="relative w-16 h-64 flex-shrink-0"
      >
        <template v-for="(suit, sIdx) in suitOrder" :key="suit">
          <template v-for="(card, cIdx) in col.slots[suit]" :key="card.id">
            <div
              class="absolute transition-transform"
              :class="straightFlushIds.has(card.id) ? 'ring-2 ring-yellow-400 shadow-[0_0_10px_rgba(250,204,21,0.5)] rounded' : ''"
              :style="{
                bottom: `${(4 - sIdx) * 30 + (cIdx * 5)}px`,
                zIndex: sIdx * 10 + cIdx,
              }"
            >
              <Card
                :card="card"
                :selected="selectedIds.includes(card.id)"
                :is-highlighted="highlightedIds.has(card.id)"
                small
                @click="onToggleSelect(card.id)"
              />
            </div>
          </template>
        </template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Card as CardType, Suit, Rank } from '@/types'
import Card from './Card.vue'
import { getLogicValue } from '@/utils/rules'   // 假设有该函数

const props = defineProps<{
  cards: CardType[]
  level: number
  viewMode: 'normal' | 'stacked'
  selectedIds: string[]
  highlightedIds: Set<string>
  straightFlushIds: Set<string>
}>()

const emit = defineEmits<{
  (e: 'toggleSelect', id: string): void
}>()

const suitOrder = [Suit.Joker, Suit.Spades, Suit.Hearts, Suit.Clubs, Suit.Diamonds]

const stackedMatrix = computed(() => {
  if (props.viewMode !== 'stacked') return []
  
  const presentValues = new Set<number>()
  props.cards.forEach(c => {
    presentValues.add(getLogicValue(c.rank, props.level))
  })
  const sortedVals = Array.from(presentValues).sort((a, b) => b - a)

  return sortedVals.map(val => {
    const cardsOfRank = props.cards.filter(c => getLogicValue(c.rank, props.level) === val)
    const slots: Record<number, CardType[]> = {
      [Suit.Joker]: [],
      [Suit.Spades]: [],
      [Suit.Hearts]: [],
      [Suit.Clubs]: [],
      [Suit.Diamonds]: [],
    }
    cardsOfRank.forEach(c => {
      if (c.rank === Rank.SmallJoker || c.rank === Rank.BigJoker) {
        slots[Suit.Joker].push(c)
      } else {
        slots[c.suit].push(c)
      }
    })
    return { val, slots }
  })
})

const onToggleSelect = (id: string) => {
  emit('toggleSelect', id)
}
</script>